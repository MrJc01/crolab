package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

type LocalRunResult struct {
	Status   string `json:"status"`
	Output   string `json:"output"`
	ExitCode int    `json:"exit_code"`
	Error    string `json:"error,omitempty"`
}

// RunLocalProject executa o código fonte especificado (arquivo ou diretório).
// Se watchMode for ativo, ele ficará travado aguardando eventos do fsnotify.
func RunLocalProject(targetPath string, watchMode bool, jsonFormat bool) error {
	info, err := os.Stat(targetPath)
	if err != nil {
		return fmt.Errorf("caminho não encontrado: %s", targetPath)
	}

	execPath := targetPath
	if info.IsDir() {
		// Se diretório, procura por entrypoints comuns
		if _, err := os.Stat(filepath.Join(targetPath, "main.py")); err == nil {
			execPath = filepath.Join(targetPath, "main.py")
		} else if _, err := os.Stat(filepath.Join(targetPath, "index.js")); err == nil {
			execPath = filepath.Join(targetPath, "index.js")
		} else {
			return fmt.Errorf("diretório informado não contém entrypoint suportado (main.py, index.js)")
		}
	}

	if watchMode {
		if !jsonFormat {
			fmt.Printf("👀 Modo Watch ativado em: %s\n", targetPath)
		}
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			return err
		}
		defer watcher.Close()

		if info.IsDir() {
			err = watcher.Add(targetPath)
		} else {
			err = watcher.Add(filepath.Dir(targetPath))
		}
		if err != nil {
			return err
		}

		// Primeira execução a frio
		executeLocalCode(execPath, jsonFormat)

		// Loop de eventos com debounce modesto
		var debounceTimer *time.Timer
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return nil
				}
				if event.Has(fsnotify.Write) {
					if info.IsDir() && !strings.HasPrefix(event.Name, targetPath) {
						continue // Ignore outside of tree, but native driver handle this
					}
					
					// Ignora arquivos indevidos se quiser, mas para simplificar, trigger no Write
					if debounceTimer != nil {
						debounceTimer.Stop()
					}
					debounceTimer = time.AfterFunc(500*time.Millisecond, func() {
						if !jsonFormat {
							fmt.Printf("\n↻ %s alterado. Reexecutando...\n\n", filepath.Base(event.Name))
						}
						executeLocalCode(execPath, jsonFormat)
					})
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return nil
				}
				if !jsonFormat {
					fmt.Println("Erro no watcher:", err)
				}
			}
		}
	}

	// Single execution
	return executeLocalCode(execPath, jsonFormat)
}

func executeLocalCode(execPath string, jsonFormat bool) error {
	var cmd *exec.Cmd

	ext := strings.ToLower(filepath.Ext(execPath))
	switch ext {
	case ".py":
		cmd = exec.Command("python3", execPath)
	case ".js":
		cmd = exec.Command("node", execPath)
	case ".sh":
		cmd = exec.Command("bash", execPath)
	case ".ipynb":
		// Na Fase de CLI, nós iremos parsear e compilar as cells `.ipynb` no `temp.py` e rodar
		return runJupyterNotebook(execPath, jsonFormat)
	default:
		return fmt.Errorf("extensão não suportada localmente nativo: %s", ext)
	}

	cmd.Dir = filepath.Dir(execPath)
	
	if jsonFormat {
		output, err := cmd.CombinedOutput()
		exitCode := 0
		if err != nil {
			exitCode = 1
			if exitErr, ok := err.(*exec.ExitError); ok {
				exitCode = exitErr.ExitCode()
			}
		}

		res := LocalRunResult{
			Status:   "success",
			Output:   string(output),
			ExitCode: exitCode,
		}
		if err != nil {
			res.Status = "error"
			res.Error = err.Error()
		}

		resBytes, _ := json.Marshal(res)
		fmt.Println(string(resBytes))
		return nil
	}

	// Execução tradicional, conectando TTY
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

type JupyterCell struct {
	CellType string      `json:"cell_type"` // "code" or "markdown"
	Source   interface{} `json:"source"`    // string ou []string
}
type JupyterNotebook struct {
	Cells []JupyterCell `json:"cells"`
}

func runJupyterNotebook(execPath string, jsonFormat bool) error {
	content, err := os.ReadFile(execPath)
	if err != nil {
		return err
	}

	var nb JupyterNotebook
	if err := json.Unmarshal(content, &nb); err != nil {
		return fmt.Errorf("falha ao parsear .ipynb: %v", err)
	}

	var codeBuilder strings.Builder
	for _, cell := range nb.Cells {
		if cell.CellType == "code" {
			switch src := cell.Source.(type) {
			case string:
				codeBuilder.WriteString(src)
			case []interface{}:
				for _, line := range src {
					if lineStr, ok := line.(string); ok {
						codeBuilder.WriteString(lineStr)
					}
				}
			}
			codeBuilder.WriteString("\n\n")
		}
	}

	tempPy := filepath.Join(filepath.Dir(execPath), ".crolab_temp_notebook.py")
	os.WriteFile(tempPy, []byte(codeBuilder.String()), 0600)
	defer os.Remove(tempPy)

	return executeLocalCode(tempPy, jsonFormat)
}
