// Copyright (c) 2026 Crolab Contributors. All rights reserved.
// Licensed under the Crolab Sustainable License (CSL).
// Contact: mrj.crom@gmail.com
package main

import (
	"fmt"
	"runtime"

	"github.com/crolab/core/internal/cli"
	"github.com/crolab/core/internal/node"
	"github.com/spf13/cobra"
)

var statusCmd_ = &cobra.Command{
	Use:   "status",
	Short: "Mostra o estado local do Crolab (config, GPUs, versão)",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println()
		fmt.Println("  Crolab Status")
		fmt.Println("  ─────────────────────────")
		fmt.Printf("  OS:       %s/%s\n", runtime.GOOS, runtime.GOARCH)
		fmt.Printf("  Go:       %s\n", runtime.Version())

		// Config
		servers, defaultName, _ := cli.ListServers()
		fmt.Printf("  Servers:  %d configurados\n", len(servers))
		if defaultName != "" {
			fmt.Printf("  Default:  %s\n", defaultName)
		}

		// Cloud auth
		cfg, _ := cli.LoadConfig()
		if cfg.CloudToken != "" {
			fmt.Println("  Cloud:    ✓ logado")
		} else {
			fmt.Println("  Cloud:    ✗ não logado")
		}

		// GPUs
		gpus := node.DetectGPUs()
		if len(gpus) == 0 {
			fmt.Println("  GPUs:     nenhuma detectada (nvidia-smi)")
		} else {
			fmt.Printf("  GPUs:     %d encontrada(s)\n", len(gpus))
			for _, g := range gpus {
				fmt.Printf("            [%s] %s (%s, driver %s)\n", g.Index, g.Name, g.Memory, g.Driver)
			}
		}
		fmt.Println()
	},
}

func init() {
	// registered in main()
}
