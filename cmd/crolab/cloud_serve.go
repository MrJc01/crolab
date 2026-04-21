package main

import (
	"fmt"
	"os"

	"github.com/crolab/core/internal/cli"
	"github.com/crolab/core/internal/cloud"
	"github.com/spf13/cobra"
)

var cloudServeCmd = &cobra.Command{
	Use:   "cloud-serve",
	Short: "Gerencia o servidor Central Cloud API",
}

var cloudServeStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Inicia o servidor principal",
	Run: func(cmd *cobra.Command, args []string) {
		if cli.Daemonize("cloud-serve") {
			return
		}
		if err := cloud.StartCloudServer(":8844", cloudWebDir, cloudTlsCert, cloudTlsKey); err != nil {
			fmt.Fprintf(os.Stderr, "Erro: %v\n", err)
			os.Exit(1)
		}
	},
}

var cloudServeStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Desliga o servidor rodando em background",
	Run: func(cmd *cobra.Command, args []string) {
		cli.DaemonStop("cloud-serve")
	},
}

func init() {
	cloudServeStartCmd.Flags().StringVar(&cloudWebDir, "web", "", "Diretório do frontend")
	cloudServeStartCmd.Flags().StringVar(&cloudTlsCert, "tls-cert", "", "Certificado TLS")
	cloudServeStartCmd.Flags().StringVar(&cloudTlsKey, "tls-key", "", "Chave privada TLS")
	cloudServeStartCmd.Flags().BoolP("daemon", "d", false, "Rodar em background via Daemon")

	cloudServeCmd.AddCommand(cloudServeStartCmd)
	cloudServeCmd.AddCommand(cloudServeStopCmd)
}
