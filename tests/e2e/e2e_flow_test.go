package e2e

import (
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// TestFullCycleIntegration dispara o binário "crolab serve" em background como um Daemon real
// TestFullCycleIntegration dispara o binário com Auth e executa comandos reais na CLI.
func TestFullCycleIntegration(t *testing.T) {
	// 1. Instanciar The Tank
	cmdServer := exec.Command("../../crolab", "serve", "--port", ":6969", "--token", "E2EAuth")
	
	err := cmdServer.Start()
	if err != nil {
		t.Fatalf("Falha crítica ao subir o processo base RPC (verifique compilação prévia): %v", err)
	}

	// Garante que o processo Node Agente receba o SIGKILL ao final do teste (E2E Cleanup SRE)
	defer func() {
		if cmdServer.Process != nil {
			cmdServer.Process.Kill()
		}
	}()

	// Wait port 4422 open
	time.Sleep(1 * time.Second)

	// 2. Preparar payload de desenvolvedor Python via CLI
	tmpDir := t.TempDir()
	envPath := tmpDir + "/test_job.py"
	os.WriteFile(envPath, []byte("print('Tudo funcional')"), 0644)

	// Adicionar o server no config
	cmdConfig := exec.Command("../../crolab", "config", "add", "TheTank", "127.0.0.1:6969", "E2EAuth")
	if err := cmdConfig.Run(); err != nil {
		t.Fatalf("Falha ao registrar node de teste no config.yaml: %v", err)
	}

	// 3. Atirar pelo Pipeline
	cmdClient := exec.Command("../../crolab", "run", tmpDir)
	out, err := cmdClient.CombinedOutput()

	if err != nil {
		t.Fatalf("O teste CLI Client abortou ou travou.\nSaída do Container:\n%s\nErro: %v", string(out), err)
	}

	outStr := string(out)
	
	// Validando as Strings provindas do terminal
	if !strings.Contains(outStr, "Deploying job") || !strings.Contains(outStr, "REMOTE LOGS") {
		t.Errorf("A saída E2E obteve código 0 mas não imprimiu o Log Real previsto na UI.\nOut: %s", outStr)
	} else {
		t.Logf("[Sucesso E2E]: O ciclo CLI -> Server -> CLI ocorreu nativamente.\n>> Payload Resultante:\n%s", outStr)
	}
}
