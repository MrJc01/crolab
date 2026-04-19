package node

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func RunDockerJob(jobID string, imageRef string, cmdStr string, payload []byte) error {
	// 1. Setup do Workspace Base
	workspace := filepath.Join(os.TempDir(), "crolab", jobID)
	os.MkdirAll(workspace, 0755)

	if len(payload) > 0 {
		zipReader, err := zip.NewReader(bytes.NewReader(payload), int64(len(payload)))
		if err == nil {
			for _, f := range zipReader.File {
				p := filepath.Join(workspace, f.Name)
				// Proteção SRE ZipSlip O(1)
				if !strings.HasPrefix(p, filepath.Clean(workspace)+string(os.PathSeparator)) {
					continue
				}
				if f.FileInfo().IsDir() {
					os.MkdirAll(p, f.Mode())
					continue
				}
				os.MkdirAll(filepath.Dir(p), 0755)
				dst, _ := os.OpenFile(p, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
				src, _ := f.Open()
				io.Copy(dst, src)
				dst.Close()
				src.Close()
			}
		}
	}

	// 2. Acionar CLI do Docker (Gerenciamento Padrão/Normal SRE)
	log.Printf("Iniciando Docker run via Raw CLI para a imagem: %s", imageRef)
	
	cmd := exec.Command("docker", "run", "--rm", "-v", fmt.Sprintf("%s:/workspace", workspace), "-w", "/workspace", imageRef, "sh", "-c", cmdStr)
	
	// Acoplando Saída Standard aos Sockets do Host para Streaming
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("O Container falhou ou obteve Exception TTY: %v", err)
	}

	log.Printf("Processo Crolab [%s] consolidado com sucesso natural.", jobID)
	return nil
}
