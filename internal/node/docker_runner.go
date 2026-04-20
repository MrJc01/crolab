// Copyright (c) 2026 Crolab Contributors. All rights reserved.
// Licensed under the Crolab Sustainable License (CSL).
// Contact: mrj.crom@gmail.com
package node

import (
	"archive/zip"
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// UnzipDir extracts a zip payload into destDir.
// Includes ZipSlip protection.
func UnzipDir(payload []byte, destDir string) error {
	zipReader, err := zip.NewReader(bytes.NewReader(payload), int64(len(payload)))
	if err != nil {
		return err
	}

	for _, f := range zipReader.File {
		fpath := filepath.Join(destDir, f.Name)

		// ZipSlip guard
		if !strings.HasPrefix(filepath.Clean(fpath), filepath.Clean(destDir)+string(os.PathSeparator)) {
			return fmt.Errorf("caminho suspeito no zip (zipslip): %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}
	return nil
}

// RunDockerJob creates a workspace, unzips the payload, runs Docker, and
// streams real stdout/stderr into logChan for the gRPC StreamLogs consumer.
func RunDockerJob(jobID, imageRef, cmdStr string, payload []byte, logChan chan<- string) error {
	workspace := filepath.Join(os.TempDir(), "crolab_jobs", jobID)

	if err := os.MkdirAll(workspace, 0755); err != nil {
		return fmt.Errorf("não conseguiu criar workspace: %v", err)
	}

	if err := UnzipDir(payload, workspace); err != nil {
		return fmt.Errorf("falha ao descompactar payload: %v", err)
	}

	log.Printf("⚙️  Docker run: image=%s cmd=%s workspace=%s", imageRef, cmdStr, workspace)

	cmdConfig := []string{
		"run", "--rm",
		"--cpus", "2.0",
		"--memory", "4g",
		"-v", fmt.Sprintf("%s:/workspace", workspace),
		"-w", "/workspace",
		imageRef,
		"sh", "-c", cmdStr,
	}

	cmd := exec.Command("docker", cmdConfig...)

	// Capture stdout and stderr via pipes → logChan
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("falha ao criar pipe stdout: %v", err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("falha ao criar pipe stderr: %v", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("falha ao iniciar container: %v", err)
	}

	// Stream both stdout and stderr into logChan
	go func() {
		scanner := bufio.NewScanner(stdoutPipe)
		for scanner.Scan() {
			logChan <- scanner.Text() + "\n"
		}
	}()
	go func() {
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			logChan <- "[stderr] " + scanner.Text() + "\n"
		}
	}()

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("container terminou com erro: %v", err)
	}

	// Cleanup workspace
	_ = os.RemoveAll(workspace)
	return nil
}
