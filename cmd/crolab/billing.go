// Copyright (c) 2026 Crolab Contributors. All rights reserved.
// Licensed under the Crolab Sustainable License (CSL).
// Contact: mrj.crom@gmail.com
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/crolab/core/internal/cli"
	"github.com/spf13/cobra"
)

var billingCmd = &cobra.Command{
	Use:   "billing",
	Short: "Faturamento e créditos Crom Cloud",
}

var billingStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Mostra saldo de créditos",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, _ := cli.LoadConfig()
		if cfg.CloudToken == "" {
			fmt.Println("Não logado. Use: crolab auth login")
			return
		}

		req, _ := http.NewRequest("GET", GetCloudAPI()+"/billing/status", nil)
		req.Header.Set("Authorization", cfg.CloudToken)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Erro: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		data, _ := io.ReadAll(resp.Body)
		var result map[string]interface{}
		json.Unmarshal(data, &result)

		if resp.StatusCode != 200 {
			fmt.Fprintf(os.Stderr, "✗ %s\n", result["error"])
			os.Exit(1)
		}

		fmt.Printf("  Email:    %s\n", result["email"])
		fmt.Printf("  Créditos: %.2f\n", result["credits"].(float64))
	},
}

var machinesCmd = &cobra.Command{
	Use:   "machines",
	Short: "Lista máquinas disponíveis na Crom Cloud",
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := http.Get(GetCloudAPI() + "/machines")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Erro: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		data, _ := io.ReadAll(resp.Body)
		var machines []map[string]interface{}
		json.Unmarshal(data, &machines)

		fmt.Println()
		fmt.Printf("  %-18s %-10s %-6s %-10s %s\n", "NOME", "GPU", "VRAM", "PREÇO/H", "STATUS")
		fmt.Printf("  %-18s %-10s %-6s %-10s %s\n", "──────────────────", "──────────", "──────", "──────────", "──────────")
		for _, m := range machines {
			fmt.Printf("  %-18s %-10s %-6s $%-9.2f %s\n",
				m["name"], m["gpu"], m["vram"], m["price_hr"], m["status"])
		}
		fmt.Println()
	},
}

func init() {
	billingCmd.AddCommand(billingStatusCmd)
	billingCmd.AddCommand(machinesCmd)
}
