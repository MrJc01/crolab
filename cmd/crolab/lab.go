// Copyright (c) 2026 Crolab Contributors. All rights reserved.
// Licensed under the Crolab Sustainable License (CSL).
// Contact: mrj.crom@gmail.com
package main

import (
	"fmt"
	"os/exec"
	"runtime"
	"time"

	"github.com/spf13/cobra"
)

var labPort string

var labCmd = &cobra.Command{
	Use:   "lab [diretório]",
	Short: "Abre o Crolab Lab — editor web com execução de scripts",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🚀 Crolab Lab")
		fmt.Println("O ambiente Lab agora é 100% Nativo na Cloud (Jupyter/Colab Clone).")
		fmt.Println("Injetando motor Web unificado...")

		// Executa o provider start implicitamente
		url := "http://localhost:8855/#lab"
		go func() {
			// Dá tempo pro server subir
			time.Sleep(2 * time.Second)
			openBrowser(url)
		}()
        
		providerStartCmd.Run(providerStartCmd, []string{})
	},
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		return
	}
	cmd.Start()
}

func init() {
	labCmd.Flags().StringVarP(&labPort, "port", "p", ":8855", "Porta do Lab")
}
