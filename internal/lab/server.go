// Copyright (c) 2026 Crolab Contributors. All rights reserved.
// Licensed under the Crolab Sustainable License (CSL).
// Contact: mrj.crom@gmail.com
package lab

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/net/websocket"
)

// FileInfo represents a file or directory entry.
type FileInfo struct {
	Name  string `json:"name"`
	Path  string `json:"path"`
	IsDir bool   `json:"is_dir"`
	Size  int64  `json:"size"`
}

var (
	basePath   string
	basePathMu sync.RWMutex
)

func safePath(requested string) (string, error) {
	basePathMu.RLock()
	base := basePath
	basePathMu.RUnlock()

	clean := filepath.Clean(filepath.Join(base, requested))
	if !strings.HasPrefix(clean, filepath.Clean(base)) {
		return "", fmt.Errorf("acesso negado: %s", requested)
	}
	return clean, nil
}

// --- File API ---

func handleListFiles(w http.ResponseWriter, r *http.Request) {
	dir := r.URL.Query().Get("dir")
	if dir == "" {
		dir = "."
	}

	fullPath, err := safePath(dir)
	if err != nil {
		http.Error(w, err.Error(), 403)
		return
	}

	entries, err := os.ReadDir(fullPath)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	var files []FileInfo
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".") {
			continue // Hide dotfiles
		}
		info, _ := e.Info()
		size := int64(0)
		if info != nil {
			size = info.Size()
		}
		files = append(files, FileInfo{
			Name:  e.Name(),
			Path:  filepath.Join(dir, e.Name()),
			IsDir: e.IsDir(),
			Size:  size,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}

func handleReadFile(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	fullPath, err := safePath(path)
	if err != nil {
		http.Error(w, err.Error(), 403)
		return
	}

	data, err := os.ReadFile(fullPath)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write(data)
}

func handleSaveFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", 405)
		return
	}

	var body struct {
		Path    string `json:"path"`
		Content string `json:"content"`
	}
	json.NewDecoder(r.Body).Decode(&body)

	fullPath, err := safePath(body.Path)
	if err != nil {
		http.Error(w, err.Error(), 403)
		return
	}

	os.MkdirAll(filepath.Dir(fullPath), 0755)
	if err := os.WriteFile(fullPath, []byte(body.Content), 0644); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "saved"})
}

func handleSetDir(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", 405)
		return
	}
	var body struct {
		Path string `json:"path"`
	}
	json.NewDecoder(r.Body).Decode(&body)

	abs, err := filepath.Abs(body.Path)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	info, err := os.Stat(abs)
	if err != nil || !info.IsDir() {
		http.Error(w, "diretório não encontrado: "+abs, 404)
		return
	}

	basePathMu.Lock()
	basePath = abs
	basePathMu.Unlock()

	log.Printf("📂 Lab workspace → %s", abs)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"path": abs})
}

func handleGetDir(w http.ResponseWriter, r *http.Request) {
	basePathMu.RLock()
	p := basePath
	basePathMu.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"path": p})
}

// --- WebSocket Exec ---

func handleExecWS(ws *websocket.Conn) {
	var msg struct {
		Command string `json:"command"`
	}
	if err := websocket.JSON.Receive(ws, &msg); err != nil {
		websocket.JSON.Send(ws, map[string]string{"error": err.Error()})
		return
	}

	basePathMu.RLock()
	cwd := basePath
	basePathMu.RUnlock()

	log.Printf("▶ Executando: %s (em %s)", msg.Command, cwd)

	cmd := exec.Command("bash", "-c", msg.Command)
	cmd.Dir = cwd

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		websocket.JSON.Send(ws, map[string]interface{}{"type": "error", "data": err.Error()})
		return
	}

	var wg sync.WaitGroup
	stream := func(r io.Reader, streamType string) {
		defer wg.Done()
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			websocket.JSON.Send(ws, map[string]interface{}{
				"type": streamType,
				"data": scanner.Text(),
			})
		}
	}

	wg.Add(2)
	go stream(stdout, "stdout")
	go stream(stderr, "stderr")
	wg.Wait()

	err := cmd.Wait()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	}

	websocket.JSON.Send(ws, map[string]interface{}{
		"type":      "exit",
		"exit_code": exitCode,
	})
}

// StartLabServer starts the Lab web interface.
func StartLabServer(port string, workDir string, webDir string) error {
	abs, err := filepath.Abs(workDir)
	if err != nil {
		return err
	}
	basePath = abs

	mux := http.NewServeMux()

	// File API
	mux.HandleFunc("/api/files", handleListFiles)
	mux.HandleFunc("/api/file", handleReadFile)
	mux.HandleFunc("/api/save", handleSaveFile)
	mux.HandleFunc("/api/dir", handleGetDir)
	mux.HandleFunc("/api/setdir", handleSetDir)

	// WebSocket for execution
	mux.Handle("/api/exec", websocket.Handler(handleExecWS))

	// Serve lab frontend
	labDir := filepath.Join(webDir, "lab")
	mux.Handle("/", http.FileServer(http.Dir(labDir)))

	log.Printf("🧪 Crolab Lab em http://localhost%s", port)
	log.Printf("📂 Workspace: %s", abs)

	return http.ListenAndServe(port, mux)
}
