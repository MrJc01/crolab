package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func getPidFile(cmdName string) string {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".crolab")
	os.MkdirAll(dir, 0755)
	return filepath.Join(dir, cmdName+".pid")
}

// Daemonize checks if we are running in detached mode. If so, starts the process and exits the parent.
func Daemonize(cmdName string) bool {
	// Se tivermos a flag magica de background no exec
	if len(os.Args) > 1 && os.Args[len(os.Args)-1] == "__DAEMON__" {
		return false // Já somos o filho! rodar normalmente.
	}

	// Filtra args removendo -d ou --daemon se existir pra montar a call final
	var cleanArgs []string
	isDaemon := false
	for _, arg := range os.Args[1:] {
		if arg == "-d" || arg == "--daemon" || arg == "--detach" {
			isDaemon = true
			continue
		}
		cleanArgs = append(cleanArgs, arg)
	}

	if !isDaemon {
		return false // Roda em foreground.
	}

	cleanArgs = append(cleanArgs, "__DAEMON__")

	cmd := exec.Command(os.Args[0], cleanArgs...) // Repassa tudo, incluindo shell env
	home, _ := os.UserHomeDir()
	logPath := filepath.Join(home, ".crolab", cmdName+".log")
	logFile, _ := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	// Start o processo de fundo (detached no SO)
	if err := cmd.Start(); err != nil {
		fmt.Printf("❌ Falha subindo daemon %s: %v\n", cmdName, err)
		os.Exit(1)
	}

	pidFile := getPidFile(cmdName)
	os.WriteFile(pidFile, []byte(strconv.Itoa(cmd.Process.Pid)), 0644)
	fmt.Printf("🌟 Crolab Daemon [%s] lançado em Background (PID: %d)\n", cmdName, cmd.Process.Pid)
	fmt.Printf("→ Use 'crolab %s stop' para desligar.\n", cmdName)
	os.Exit(0)
	return true
}

func DaemonStop(cmdName string) {
	pidFile := getPidFile(cmdName)
	data, err := os.ReadFile(pidFile)
	if err != nil {
		fmt.Printf("⚠️  Serviço %s não parece estar rodando no background.\n", cmdName)
		return
	}

	pid, _ := strconv.Atoi(strings.TrimSpace(string(data)))
	
	process, err := os.FindProcess(pid)
	if err == nil {
		// Envia signal generico de kill (funciona mac/unix e windows fallbacks dependendo da runtime)
		err = process.Kill()
		if err == nil {
			fmt.Printf("🛑 Daemon [%s] ID %d interceptado e desligado com sucesso.\n", cmdName, pid)
		} else {
			fmt.Printf("⚠️  Falha matando PID %d: %v\n", pid, err)
		}
	}
	
	os.Remove(pidFile)
}
