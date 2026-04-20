package report_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Este teste atua como um gerador unificado de Documentação em Tempo Real.
// Ele sobe o Backend/Frontend temporariamente, e então atira uma instância
// Headless Chromium para fotografar cada tela em produção,
// compilando o result num Relatório Markdown final.

func TestGenerateReportFull(t *testing.T) {
	// 1. Setup Test Env & Directories
	reportDir := filepath.Join("..", "..", "documentacao", "relatorios")
	os.MkdirAll(reportDir, 0755)
	
	cliHome := filepath.Join(t.TempDir(), "cli_home")
	os.MkdirAll(cliHome, 0755)
	
	dbPath := filepath.Join(t.TempDir(), "crolab_test_report.db")

	crolabEnv := append(os.Environ(), "HOME="+cliHome, "CROLAB_HOME="+cliHome)

	// 2. Verify the exact Crolab Binary exists
	t.Log("Verifying target binary...")
	binPath, err := filepath.Abs("../../crolab")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(binPath); os.IsNotExist(err) {
		t.Fatalf("O binário Crolab não existe! Rode 'go build -o crolab ./cmd/crolab/' na raiz antes de rodar o teste.")
	}

	// 3. Start Crolab Provider (Frontend Admin/Client)
	t.Log("Starting Provider server for screenshots...")
	cmdServer := exec.Command(binPath, "provider", "--admin-port", ":18844", "--client-port", ":18855", "--db", dbPath)
	cmdServer.Dir = "../../"
	cmdServer.Env = crolabEnv
	if err := cmdServer.Start(); err != nil {
		t.Fatalf("falha ao iniciar server: %v", err)
	}
	defer func() {
		cmdServer.Process.Kill()
	}()

	// Wait for server boot
	time.Sleep(3 * time.Second)

	// 3.5. Configure CLI & Seed Admin via CLI (testando Local SSO Server)
	t.Log("🌱 Configurando CLI para servidor local e Registrando Admin...")
	
	cmdCfgLocal := exec.Command(binPath, "config", "add", "local", "http://localhost:18844")
	cmdCfgLocal.Env = crolabEnv
	cmdCfgLocal.Run()
	
	cmdCfgDefault := exec.Command(binPath, "config", "set-default", "local")
	cmdCfgDefault.Env = crolabEnv
	cmdCfgDefault.Run()
	
	cmdReg := exec.Command(binPath, "auth", "register", "admin@crolab.com", "admin123")
	cmdReg.Env = crolabEnv
	if err := cmdReg.Run(); err != nil {
		t.Logf("Aviso: Falha ao registrar admin (SaaS): %v", err)
	}

	// 4. Run Python Playwright Scrapes
	t.Log("📸 Acionando Scraper via Python Playwright...")
	scraperCmd := exec.Command("/tmp/crolab_screens/venv/bin/python", "screenshot.py", reportDir)
	// Not passing crolabEnv to scraperCmd, so it uses pure OS env to find Chromium
	outScrape, err := scraperCmd.CombinedOutput()
	if err != nil {
		t.Logf("Aviso: Falha ao rodar scraper (requer playwright instalado em /tmp/crolab_screens/venv). Output: %s", string(outScrape))
	} else {
		t.Logf("✅ Scraping finalizado: \n%s", string(outScrape))
	}

	// 5. Capture CLI logs
	t.Log("📝 Capturando outputs completos da CLI...")
	
	runCmd := func(args ...string) string {
		cmdObj := exec.Command(binPath, args...)
		cmdObj.Env = crolabEnv
		out, _ := cmdObj.CombinedOutput()
		return "```text\n$ crolab " + strings.Join(args, " ") + "\n" + string(out) + "\n```\n"
	}
	
	cliLogs := ""
	cliLogs += "### Core & Autenticação\n"
	cliLogs += runCmd("--help")
	cliLogs += runCmd("auth", "--help")
	cliLogs += runCmd("auth", "login", "--help")
	cliLogs += runCmd("auth", "register", "--help")
	
	cliLogs += "### Configuração de Provedores\n"
	cliLogs += runCmd("config", "--help")
	cliLogs += runCmd("config", "add", "--help")
	cliLogs += runCmd("config", "ls", "--help")
	
	cliLogs += "### Tenant & Consumo Client\n"
	cliLogs += runCmd("subscribe", "--help")
	cliLogs += runCmd("my-machines", "--help")
	cliLogs += runCmd("run", "--help")
	cliLogs += runCmd("lab", "--help")
	cliLogs += runCmd("monitor", "--help")

	cliLogs += "### Operador Core P2P\n"
	cliLogs += runCmd("provider", "--help")
	cliLogs += runCmd("admin", "--help")
	cliLogs += runCmd("admin", "plan", "--help")
	cliLogs += runCmd("admin", "pool", "--help")
	cliLogs += runCmd("admin", "users", "--help")
	cliLogs += runCmd("admin", "machines", "--help")

	// 6. Compose Markdown
	mdContent := fmt.Sprintf(`# Relatório Forense CLI & Web Automatizado

Este relatório foi gerado algoritmicamente executando a suíte E2E nativa do Crolab.
O cliente Web agora conta com SSO sincronizado via Local Token "~/.crolab/config.json".

## 1. Módulos Core P2P (CLI)

%s

## 2. Testes de Renderização
As imagens renderizadas capturam o dashboard renderizado com precisão Single-Binary:

### Visão Headless: Mode Client
- **Home:** ![Client Home](client_home.png)
- **Planos Contratados:** ![Client Plans](client_plans.png)
- **Máquinas Associadas:** ![Client Machines](client_machines.png)
- **Jobs Executados:** ![Client Jobs](client_jobs.png)
- **Faturamento/Créditos:** ![Client Billing](client_billing.png)

### Visão Headless: Mode Admin
- **Dashboard Overview:** ![Admin Dashboard](admin_dashboard.png)
- **Planos Globais:** ![Admin Plans](admin_plans.png)
- **Máquinas Conectadas:** ![Admin Machines](admin_machines.png)
- **Usuários Tenants:** ![Admin Users](admin_users.png)
- **Sys Logs:** ![Admin Logs](admin_logs.png)
`, cliLogs)
	
	errWrite := os.WriteFile(filepath.Join(reportDir, "relatorio_telas_comandos.md"), []byte(mdContent), 0644)
	if errWrite != nil {
		t.Fatalf("Falha logando md: %v", errWrite)
	}
	
	t.Logf("✅ Relatório de auditoria e Screenshots compilados com Sucesso em %s", reportDir)
}
