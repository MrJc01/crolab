// Copyright (c) 2026 Crolab Contributors. All rights reserved.
// Licensed under the Crolab Sustainable License (CSL).
package main

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/crolab/core/internal/cli"
	"github.com/crolab/core/internal/cloud"
	"github.com/spf13/cobra"
)

var providerCmd = &cobra.Command{
	Use:   "provider",
	Short: "Inicia modo provedor (admin + client em portas separadas)",
}

var providerStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Sobe o Provider Node (use -d para rodar em background)",
	Run: runProvider,
}

var providerStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Desliga um Provider Node rodando em background",
	Run: func(cmd *cobra.Command, args []string) {
		cli.DaemonStop("provider")
	},
}

func init() {
	providerStartCmd.Flags().String("admin-port", ":8844", "Porta do painel admin")
	providerStartCmd.Flags().String("client-port", ":8855", "Porta do painel client")
	providerStartCmd.Flags().String("db", "crolab.db", "Caminho do bash SQLite")
	providerStartCmd.Flags().String("web", "./web", "Diretório raiz dos frontends")
	providerStartCmd.Flags().String("tls-cert", "", "Certificado SSL")
	providerStartCmd.Flags().String("tls-key", "", "Chave Privada SSL")
	providerStartCmd.Flags().Bool("no-prompt", false, "Desabilitar setup interativo (usa Env Vars CROLAB_ADMIN_*)")
	providerStartCmd.Flags().Bool("json-logs", false, "Habilita logs em formato JSON estruturado SRE")
	providerStartCmd.Flags().Float64("free-credits", 10.0, "Quantidade de créditos de boas-vindas atribuídos a novas contas")
	providerStartCmd.Flags().BoolP("daemon", "d", false, "Rodar em background via Daemon")

	providerCmd.AddCommand(providerStartCmd)
	providerCmd.AddCommand(providerStopCmd)
	rootCmd.AddCommand(providerCmd)
}

func genRootPassword() string {
	b := make([]byte, 6)
	rand.Read(b)
	return "cr0_" + hex.EncodeToString(b)
}

func runProvider(cmd *cobra.Command, args []string) {
	adminPort, _ := cmd.Flags().GetString("admin-port")
	clientPort, _ := cmd.Flags().GetString("client-port")
	dbPath, _ := cmd.Flags().GetString("db")
	webDir, _ := cmd.Flags().GetString("web")
	tlsCert, _ := cmd.Flags().GetString("tls-cert")
	tlsKey, _ := cmd.Flags().GetString("tls-key")
	useJsonLogs, _ := cmd.Flags().GetBool("json-logs")

	var handler slog.Handler
	if useJsonLogs {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	}
	logger := slog.New(handler)
	slog.SetDefault(logger)
	// Bind legacy log outputs to slog
	log.SetOutput(os.Stdout)
	log.SetFlags(0) // Remover datas legadas default


	isFirstBoot := false
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		isFirstBoot = true
	}

	isDaemonChild := (len(os.Args) > 1 && os.Args[len(os.Args)-1] == "__DAEMON__")

	noPrompt, _ := cmd.Flags().GetBool("no-prompt")
	if os.Getenv("CROLAB_NO_PROMPT") == "true" {
		noPrompt = true
	}

	if isFirstBoot && !isDaemonChild {
		fmt.Println("=======================================================")
		fmt.Println("🌌 BEM-VINDO À INICIALIZAÇÃO DO CROLAB PROVIDER!")
		
		var emailInput, passInput, vastKey string

		if noPrompt {
			fmt.Println("⚡ Modo interativo desativado (no-prompt). Lendo variáveis de ambiente...")
			emailInput = os.Getenv("CROLAB_ADMIN_EMAIL")
			if emailInput == "" { emailInput = "root@crolab.local" }
			passInput = os.Getenv("CROLAB_ADMIN_PASS")
			if passInput == "" { passInput = genRootPassword() }
			vastKey = os.Getenv("CROLAB_VASTAI_KEY")
		} else {
			fmt.Println("Crie sua conta Administradora Suprema.")
			reader := bufio.NewReader(os.Stdin)

			fmt.Print("👉 Email Master [Padrão: root@crolab.local]: ")
			emailInput, _ = reader.ReadString('\n')
			emailInput = strings.TrimSpace(emailInput)
			if emailInput == "" {
				emailInput = "root@crolab.local"
			}
			
			fmt.Print("👉 Senha Master [Deixe vazio para Gerar Automático]: ")
			passInput, _ = reader.ReadString('\n')
			passInput = strings.TrimSpace(passInput)
			if passInput == "" {
				passInput = genRootPassword()
			}

			fmt.Print("\n👉 Token API da Vast.AI (Opcional - deixe vazio p/ pular): ")
			vastKey, _ = reader.ReadString('\n')
			vastKey = strings.TrimSpace(vastKey)
		}

		if err := cloud.InitDB(dbPath); err != nil {
			log.Fatalf("Erro DB: %v", err)
		}
		
		cloud.DBCreateUser(emailInput, passInput, "admin", "127.0.0.1")
		if vastKey != "" {
			cloud.DBSetSetting("vastai_api_key", vastKey)
		}
		
		fmt.Println("\n✅ Configuração Padrão Concluída!")
		fmt.Println("Sua Conta Mestra Admin foi Criada:")
		fmt.Printf("   Usuário: %s\n", emailInput)
		fmt.Printf("   Senha:   %s\n", passInput)
		fmt.Println("Guarde estes dados. Se for uma senha gerada, ela jamais será exibida novamente.")
		fmt.Println("=======================================================")
	}

	if cli.Daemonize("provider") {
		return
	}

	if err := cloud.InitDB(dbPath); err != nil {
		log.Fatalf("❌ Falha DB: %v", err)
	}

	freeCredits, _ := cmd.Flags().GetFloat64("free-credits")
	envCred := os.Getenv("CROLAB_FREE_CREDITS")
	if envCred != "" {
		fmt.Sscanf(envCred, "%f", &freeCredits)
	}
	// Configura o modelo econômico antes do boot
	cloud.DBSetSetting("free_credits_amount", fmt.Sprintf("%.2f", freeCredits))

	protocol := "http"
	if tlsCert != "" && tlsKey != "" {
		protocol = "https"
	}

	fmt.Println()
	fmt.Println("  ⚡ CROLAB PROVIDER MODE")
	fmt.Println("  ════════════════════════════════════")
	fmt.Printf("  Admin:  %s://localhost%s\n", protocol, adminPort)
	fmt.Printf("  Client: %s://localhost%s\n", protocol, clientPort)
	fmt.Printf("  DB:     %s\n", dbPath)
	fmt.Println("  ════════════════════════════════════")
	fmt.Println()

	adminWeb := webDir + "/admin"
	clientWeb := webDir + "/client"

	// Daemon CRON
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		for range ticker.C {
			log.Println("🔄 CRON: Sincronizando Vast.AI Marketplace (P2P)...")
			cloud.SyncVastAIOffers()
		}
	}()

	go func() {
		mux := cloud.BuildMux(adminWeb)
		if protocol == "https" {
			http.ListenAndServeTLS(adminPort, tlsCert, tlsKey, mux)
		} else {
			http.ListenAndServe(adminPort, mux)
		}
	}()

	mux := cloud.BuildMux(clientWeb)
	if protocol == "https" {
		http.ListenAndServeTLS(clientPort, tlsCert, tlsKey, mux)
	} else {
		http.ListenAndServe(clientPort, mux)
	}
}
