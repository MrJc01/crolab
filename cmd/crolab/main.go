// Copyright (c) 2026 Crolab Contributors. All rights reserved.
// Licensed under the Crolab Sustainable License (CSL).
// Contact: mrj.crom@gmail.com
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/crolab/core/internal/cli"
	"github.com/crolab/core/internal/tui"
	"github.com/spf13/cobra"
)

// --- Root ---

var Version = "0.2.0"

var rootCmd = &cobra.Command{
	Use:     "crolab",
	Short:   "Crolab — Orquestrador P2P de GPU para IA",
	Version: Version,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		cli.InitConfig()
	},
}

// --- Run ---

var (
	runImage   string
	runCmd_    string
	runPlan    string
	runMachine string
	runWatch   bool
	runJson    bool
)

var runCmd = &cobra.Command{
	Use:   "run [diretório]",
	Short: "Envia código para execução via Crolab Cloud",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		targetDir := args[0]
		cfg, _ := cli.LoadConfig()

		if runPlan == "" && runMachine == "" {
			err := cli.RunLocalProject(targetDir, runWatch, runJson)
			if err != nil {
				if runJson {
					fmt.Fprintf(os.Stderr, "{\"status\": \"error\", \"error\": \"%s\"}\n", err.Error())
				} else {
					fmt.Fprintf(os.Stderr, "❌ Erro local: %v\n", err)
				}
				os.Exit(1)
			}
			return
		}

		if cfg.CloudToken == "" {
			fmt.Fprintln(os.Stderr, "Erro: não autenticado. Rode: crolab auth login")
			os.Exit(1)
		}

		fmt.Printf("☁️  Requisitando escalonamento na Nuvem...\n")
		
		bodyReq := map[string]string{
			"plan_id":    runPlan,
			"machine_id": runMachine,
		}
		bodyBytes, _ := json.Marshal(bodyReq)
		
		targetCloud := cfg.CloudAPI
		if targetCloud == "" {
			targetCloud = "http://localhost:8844"
		}
		
		req, _ := http.NewRequest("POST", targetCloud+"/client/run", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", cfg.CloudToken)
		
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Erro conectando à API: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()
		
		data, _ := io.ReadAll(resp.Body)
		var result map[string]interface{}
		json.Unmarshal(data, &result)

		if resp.StatusCode != 201 {
			fmt.Fprintf(os.Stderr, "❌ Falha: %v\n", result["error"])
			os.Exit(1)
		}

		nodesInterfaces := result["nodes"].([]interface{})
		jobID := result["job_id"].(string)

		fmt.Printf("✅ %s (Job ID: %s)\n", result["message"], jobID)
		
		success := false
		for i, n := range nodesInterfaces {
			nodeMap := n.(map[string]interface{})
			targetIP := nodeMap["address"].(string)
			targetTok := nodeMap["token"].(string)

			fmt.Printf("🚀 Ligando Node %d [%s] via gRPC...\n", i+1, targetIP)

			err := cli.SubmitJob(targetIP, targetTok, runImage, runCmd_, targetDir, runTls)
			if err == nil {
				success = true
				break
			}
			fmt.Fprintf(os.Stderr, "⚠️ Falha no Node %d: %v. Tentando próximo...\n", i+1, err)
		}

		if !success {
			fmt.Fprintf(os.Stderr, "❌ Colapso de Roteamento SRE: Nenhum Node do pool disponível para aceitar a carga!\n")
			os.Exit(1)
		}
	},
}

var runTls bool

// --- Monitor ---

var monitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Abre o dashboard interativo no terminal",
	Run: func(cmd *cobra.Command, args []string) {
		if err := tui.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Erro: %v\n", err)
			os.Exit(1)
		}
	},
}

// --- Config ---

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Gerencia servidores conectados",
}

var (
	cfgProvider string
	cfgPriority int
)

var configAddCmd = &cobra.Command{
	Use:   "add <nome> <ip:porta> [token]",
	Short: "Adiciona ou atualiza um servidor",
	Args:  cobra.RangeArgs(2, 3),
	Run: func(cmd *cobra.Command, args []string) {
		token := ""
		if len(args) == 3 {
			token = args[2]
		}
		if err := cli.AddServer(args[0], args[1], token, cfgProvider, cfgPriority); err != nil {
			fmt.Fprintf(os.Stderr, "Erro: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ Servidor [%s] salvo (provider: %s, prioridade: %d)\n", args[0], cfgProvider, cfgPriority)
	},
}

var configRmCmd = &cobra.Command{
	Use:   "rm <nome>",
	Short: "Remove um servidor",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := cli.RemoveServer(args[0]); err != nil {
			fmt.Fprintf(os.Stderr, "Erro: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ Servidor [%s] removido.\n", args[0])
	},
}

var configLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "Lista servidores em ordem de prioridade",
	Run: func(cmd *cobra.Command, args []string) {
		servers, defaultName, err := cli.ListServers()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Erro: %v\n", err)
			os.Exit(1)
		}
		if len(servers) == 0 {
			fmt.Println("Nenhum servidor configurado. Use: crolab config add")
			return
		}

		fmt.Println()
		fmt.Printf("  %-3s %-15s %-10s %-5s %s\n", "", "NOME", "PROVIDER", "PRIO", "ENDEREÇO")
		fmt.Printf("  %-3s %-15s %-10s %-5s %s\n", "---", "───────────────", "──────────", "─────", "────────────────────")
		for _, s := range servers {
			marker := "   "
			if s.Name == defaultName {
				marker = " ★ "
			}
			fmt.Printf("  %s %-15s %-10s %-5d %s\n", marker, s.Name, s.Provider, s.Priority, s.Address)
		}
		fmt.Println()
	},
}

var configDefaultCmd = &cobra.Command{
	Use:   "set-default <nome>",
	Short: "Define o servidor padrão",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := cli.SetDefault(args[0]); err != nil {
			fmt.Fprintf(os.Stderr, "Erro: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ Servidor padrão: [%s]\n", args[0])
	},
}

// --- Init & Main ---

func init() {

	runCmd.Flags().StringVar(&runImage, "image", "python:3.11-slim", "Imagem Docker")
	runCmd.Flags().StringVar(&runCmd_, "cmd", "ls /workspace", "Comando a executar")
	runCmd.Flags().StringVar(&runPlan, "plan", "", "ID do Plano para Cloud routing (ex: start, pro)")
	runCmd.Flags().StringVar(&runMachine, "machine", "", "ID (ou Nome) da máquina específica")
	runCmd.Flags().BoolVar(&runTls, "tls-rpc", false, "Usa TLS para conexão gRPC P2P")
	runCmd.Flags().BoolVar(&runWatch, "watch", false, "Hot-reload (reinicia reexecutando se o projeto/arquivo mudar localmente)")
	runCmd.Flags().BoolVar(&runJson, "json", false, "Cuspir o resultado estruturado em stdOut JSON puro")

	configAddCmd.Flags().StringVar(&cfgProvider, "provider", "local", "Provedor (local, vastai, runpod, aws...)")
	configAddCmd.Flags().IntVar(&cfgPriority, "priority", 1, "Prioridade (1 = mais alta)")

	configCmd.AddCommand(configAddCmd)
	configCmd.AddCommand(configRmCmd)
	configCmd.AddCommand(configLsCmd)
	configCmd.AddCommand(configDefaultCmd)
}

func main() {
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(monitorCmd)
	rootCmd.AddCommand(labCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(authCmd)
	rootCmd.AddCommand(billingCmd)
	rootCmd.AddCommand(cloudServeCmd)
	rootCmd.AddCommand(statusCmd_)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
