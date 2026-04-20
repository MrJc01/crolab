package cli_integration_test

import (
	"os/exec"
	"strings"
	"testing"
)

func crolab(args ...string) (string, error) {
	cmd := exec.Command("../../crolab", args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func TestCLIHelp(t *testing.T) {
	out, err := crolab("--help")
	if err != nil {
		t.Fatalf("help falhou: %v", err)
	}
	cmds := []string{"serve", "run", "monitor", "lab", "config", "auth", "billing", "status", "cloud-serve"}
	for _, c := range cmds {
		if !strings.Contains(out, c) {
			t.Errorf("comando '%s' não listado no help", c)
		}
	}
}

func TestCLIStatusRuns(t *testing.T) {
	out, err := crolab("status")
	if err != nil {
		t.Fatalf("status falhou: %v\n%s", err, out)
	}
	if !strings.Contains(out, "Crolab Status") {
		t.Error("output não contém header esperado")
	}
	if !strings.Contains(out, "OS:") {
		t.Error("output não contém OS info")
	}
}

func TestCLIConfigAddAndList(t *testing.T) {
	// Add
	out, err := crolab("config", "add", "ci-test-node", "127.0.0.1:9999", "tok", "--provider", "ci", "--priority", "99")
	if err != nil {
		t.Fatalf("config add falhou: %v\n%s", err, out)
	}
	if !strings.Contains(out, "ci-test-node") {
		t.Error("output não confirma adição")
	}

	// List
	out, err = crolab("config", "ls")
	if err != nil {
		t.Fatalf("config ls falhou: %v\n%s", err, out)
	}
	if !strings.Contains(out, "ci-test-node") {
		t.Error("node não aparece no ls")
	}
	if !strings.Contains(out, "ci") {
		t.Error("provider não aparece")
	}

	// Cleanup
	crolab("config", "rm", "ci-test-node")
}

func TestCLIConfigSetDefault(t *testing.T) {
	crolab("config", "add", "def-a", "1.1.1.1:4422", "", "--priority", "1")
	crolab("config", "add", "def-b", "2.2.2.2:4422", "", "--priority", "2")

	out, err := crolab("config", "set-default", "def-b")
	if err != nil {
		t.Fatalf("set-default falhou: %v\n%s", err, out)
	}

	out, _ = crolab("config", "ls")
	// def-b should have the star
	lines := strings.Split(out, "\n")
	for _, l := range lines {
		if strings.Contains(l, "def-b") && !strings.Contains(l, "★") {
			t.Error("def-b deveria estar marcado como default (★)")
		}
	}

	// Cleanup
	crolab("config", "rm", "def-a")
	crolab("config", "rm", "def-b")
}

func TestCLIConfigRemove(t *testing.T) {
	crolab("config", "add", "to-remove", "1.1.1.1:4422", "")

	out, err := crolab("config", "rm", "to-remove")
	if err != nil {
		t.Fatalf("rm falhou: %v\n%s", err, out)
	}

	out, _ = crolab("config", "ls")
	if strings.Contains(out, "to-remove") {
		t.Error("node ainda aparece após remoção")
	}
}

func TestCLIConfigRemoveNonExistent(t *testing.T) {
	_, err := crolab("config", "rm", "nao-existe-xyz")
	if err == nil {
		t.Error("deveria falhar ao remover inexistente")
	}
}

// Test removido: a lógica agora depende da Cloud API em vez de target local.

func TestCLIServeHelp(t *testing.T) {
	out, _ := crolab("serve", "start", "--help")
	flags := []string{"--port", "--token", "--gen", "--slots"}
	for _, f := range flags {
		if !strings.Contains(out, f) {
			t.Errorf("flag '%s' não aparece no help do serve start", f)
		}
	}
}

func TestCLIRunHelp(t *testing.T) {
	out, _ := crolab("run", "--help")
	flags := []string{"--image", "--cmd", "--plan"}
	for _, f := range flags {
		if !strings.Contains(out, f) {
			t.Errorf("flag '%s' não aparece no help do run", f)
		}
	}
}

func TestCLILabHelp(t *testing.T) {
	out, _ := crolab("lab", "--help")
	if !strings.Contains(out, "--port") {
		t.Error("lab help não mostra --port")
	}
}
