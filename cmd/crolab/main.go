package main

import (
	"fmt"
	"os"

	"github.com/crolab/core/internal/cli"
	"github.com/crolab/core/internal/node"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "crolab",
	Short: "Crolab Engine - The local-first Docker Orchestrator",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		cli.InitConfig()
	},
}

var servePort string
var serveToken string

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Starts the Crolab Node local agent",
	Run: func(cmd *cobra.Command, args []string) {
		if err := node.Start(servePort, serveToken); err != nil {
			fmt.Printf("Failed to start node: %v\n", err)
		}
	},
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Gerencia a malha P2P SRE (Hosts Salvos)",
}

var configAddCmd = &cobra.Command{
	Use:   "add [NOME] [IP:PORTA] [TOKEN_OU_NULO]",
	Short: "Aponta o cli para um hospedeiro. (Use 'none' para sem token)",
	Args:  cobra.RangeArgs(2, 3),
	Run: func(cmd *cobra.Command, args []string) {
		token := ""
		if len(args) == 3 && args[2] != "none" {
			token = args[2]
		}
		if err := cli.AddServer(args[0], args[1], token); err != nil {
			fmt.Printf("Falha ao salvar: %v\n", err)
		} else {
			fmt.Printf("✅ Hospedeiro [%s] Atrelado ao Registro.\n", args[0])
		}
	},
}

var runCmd = &cobra.Command{
	Use:   "run [script/dir]",
	Short: "Pushes job to the Crolab Node",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		targetDir := args[0]
		
		server, err := cli.GetServer("")
		if err != nil {
			fmt.Printf("🚫 Erro CLI: %v\n", err)
			return
		}

		fmt.Printf("🚀 Roteando Job para Provider: [%s] (%s)\n", server.Name, server.Address)
		
		image := "alpine:latest"
		command := "echo 'Script Executado com Sucesso'"
		
		if err := cli.SubmitJob(server.Address, server.Token, image, command, targetDir); err != nil {
			fmt.Printf("\n❌ CLI Abortou: %v\n", err)
		}
	},
}

func init() {
	serveCmd.Flags().StringVarP(&servePort, "port", "p", ":4422", "Port to bind the server")
	serveCmd.Flags().StringVarP(&serveToken, "token", "t", "T4NK_SECRET", "Security token SRE for gRPC connections") // Default security on the MVP tank

	configCmd.AddCommand(configAddCmd)
}

func main() {
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(configCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
