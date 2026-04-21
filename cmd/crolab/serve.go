package main

import (
	"fmt"
	"os"

	"github.com/crolab/core/internal/cli"
	"github.com/crolab/core/internal/node"
	"github.com/spf13/cobra"
)

var (
	servePort    string
	serveToken   string
	serveGenAuth bool
	serveSlots   int
	serveTlsCert string
	serveTlsKey  string
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Gerencia o Crolab Node (daemon gRPC que executa jobs)",
}

var serveStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Inicia o Crolab Node",
	Run: func(cmd *cobra.Command, args []string) {
		if cli.Daemonize("serve") {
			return
		}

		activeToken := serveToken

		if serveGenAuth {
			hash, err := cli.GenerateCrolabHash()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Erro gerando token: %v\n", err)
				os.Exit(1)
			}
			activeToken = hash
			fmt.Println("════════════════════════════════════════")
			fmt.Println("  Crolab Node — Token P2P gerado")
			fmt.Printf("  Token: %s\n", activeToken)
			fmt.Printf("  Conecte: crolab config add meu-node <IP>%s %s\n", servePort, activeToken)
			fmt.Println("════════════════════════════════════════")
		}

		if activeToken == "" {
			fmt.Printf("2026/04/19 20:58:57 ⚠️  Crolab Node em %s (SEM auth, %d slots)\n", servePort, serveSlots)
		} else {
			fmt.Printf("2026/04/19 20:58:57 🔒 Crolab Node em %s (AUTH ATIVA, %d slots)\n", servePort, serveSlots)
		}

		node.MaxConcurrentJobs = serveSlots
		if err := node.Start(servePort, activeToken, serveTlsCert, serveTlsKey); err != nil {
			fmt.Fprintf(os.Stderr, "Erro: %v\n", err)
			os.Exit(1)
		}
	},
}

var serveStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Desliga o Crolab Node",
	Run: func(cmd *cobra.Command, args []string) {
		cli.DaemonStop("serve")
	},
}

func init() {
	serveStartCmd.Flags().StringVar(&servePort, "port", ":4422", "Porta para ouvir as requisições gRPC")
	serveStartCmd.Flags().StringVar(&serveToken, "token", "", "Token de autenticação P2P esperado")
	serveStartCmd.Flags().BoolVar(&serveGenAuth, "gen", false, "Gerar token dinâmico seguro aleatório no boot")
	serveStartCmd.Flags().IntVar(&serveSlots, "slots", 2, "Quantidade de jobs simultâneos suportada")
	serveStartCmd.Flags().StringVar(&serveTlsCert, "tls-cert", "", "Certificado SSL")
	serveStartCmd.Flags().StringVar(&serveTlsKey, "tls-key", "", "Chave Privada SSL")
	serveStartCmd.Flags().BoolP("daemon", "d", false, "Rodar em background via Daemon")

	serveCmd.AddCommand(serveStartCmd)
	serveCmd.AddCommand(serveStopCmd)
}
