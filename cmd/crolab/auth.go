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
	"strings"

	"github.com/crolab/core/internal/cli"
	"github.com/spf13/cobra"
)

func GetCloudAPI() string {
	cfg, err := cli.LoadConfig()
	if err == nil && cfg.DefaultServer != "" {
		srv, err := cli.GetServer(cfg.DefaultServer)
		if err == nil && srv.Address != "" {
			addr := srv.Address
			if !strings.HasPrefix(addr, "http://") && !strings.HasPrefix(addr, "https://") {
				addr = "http://" + addr
			}
			return addr
		}
	}
	return "https://api.crom.cloud"
}

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Autenticação com a Crom Cloud",
}

var loginCmd = &cobra.Command{
	Use:   "login <email> <password>",
	Short: "Login na Crom Cloud",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		body, _ := json.Marshal(map[string]string{"email": args[0], "password": args[1]})
		resp, err := http.Post(GetCloudAPI()+"/auth/login", "application/json", bytes.NewReader(body))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Erro conectando à Crom Cloud: %v\n", err)
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

		// Salvar token no config
		cfg, _ := cli.LoadConfig()
		cfg.CloudToken = result["token"].(string)
		cli.SaveConfig(cfg)

		fmt.Printf("✓ Logado como %s\n", args[0])
		fmt.Printf("  Créditos: %.2f\n", result["credits"].(float64))
	},
}

var registerCmd = &cobra.Command{
	Use:   "register <email> <password>",
	Short: "Criar conta na Crom Cloud",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		body, _ := json.Marshal(map[string]string{"email": args[0], "password": args[1]})
		resp, err := http.Post(GetCloudAPI()+"/auth/register", "application/json", bytes.NewReader(body))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Erro conectando à Crom Cloud: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		data, _ := io.ReadAll(resp.Body)
		var result map[string]interface{}
		json.Unmarshal(data, &result)

		if resp.StatusCode != 201 {
			fmt.Fprintf(os.Stderr, "✗ %s\n", result["error"])
			os.Exit(1)
		}

		cfg, _ := cli.LoadConfig()
		cfg.CloudToken = result["token"].(string)
		cli.SaveConfig(cfg)

		fmt.Printf("✓ Conta criada: %s\n", args[0])
		fmt.Printf("  %s\n", result["message"])
	},
}

var (
	cloudWebDir string
	cloudTlsCert string
	cloudTlsKey  string
)

func init() {
	authCmd.AddCommand(loginCmd)
	authCmd.AddCommand(registerCmd)
}
