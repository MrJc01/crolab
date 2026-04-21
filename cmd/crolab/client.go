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
	"text/tabwriter"

	"github.com/crolab/core/internal/cli"
	"github.com/spf13/cobra"
)

var clientServer string

func init() {
	// --- Plans (public) ---
	plansCmd := &cobra.Command{
		Use:   "plans",
		Short: "Ver planos disponíveis",
		Run: func(cmd *cobra.Command, args []string) {
			cfg, _ := cli.LoadConfig()
			server, _ := cmd.Flags().GetString("server")
			if server == "http://localhost:8855" && cfg.CloudAPI != "" {
				server = cfg.CloudAPI
			}
			
			data := clientGet(server, "/client/plans", "")
			plans, ok := data.([]interface{})
			if !ok || len(plans) == 0 {
				fmt.Println("Nenhum plano disponível.")
				return
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tNOME\tVRAM\t$/HORA\t$/MÊS")
			for _, p := range plans {
				m := p.(map[string]interface{})
				fmt.Fprintf(w, "%s\t%s\t%s\t$%.2f\t$%.2f\n",
					m["id"], m["name"], m["vram"],
					toFloat(m["price_hr"]), toFloat(m["price_month"]))
			}
			w.Flush()
		},
	}
	plansCmd.Flags().String("server", "http://localhost:8855", "Servidor client")

	// --- Subscribe ---
	subscribeCmd := &cobra.Command{
		Use:   "subscribe [planID]",
		Short: "Assinar um plano",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cfg, _ := cli.LoadConfig()
			server, _ := cmd.Flags().GetString("server")
			token, _ := cmd.Flags().GetString("token")
			
			if server == "http://localhost:8855" && cfg.CloudAPI != "" {
				server = cfg.CloudAPI
			}
			if token == "" && cfg.CloudToken != "" {
				token = cfg.CloudToken
			}

			resp := clientPost(server, "/client/subscribe", token, map[string]interface{}{
				"plan_id": args[0],
			})
			if resp != nil {
				fmt.Printf("✅ %s\n", resp["message"])
			}
		},
	}
	subscribeCmd.Flags().String("server", "http://localhost:8855", "Servidor")
	subscribeCmd.Flags().String("token", "", "Token de autenticação")

	// --- My Machines ---
	myMachinesCmd := &cobra.Command{
		Use:   "my-machines",
		Short: "Listar máquinas pessoais conectadas",
		Run: func(cmd *cobra.Command, args []string) {
			cfg, _ := cli.LoadConfig()
			server, _ := cmd.Flags().GetString("server")
			token, _ := cmd.Flags().GetString("token")
			if server == "http://localhost:8855" && cfg.CloudAPI != "" {
				server = cfg.CloudAPI
			}
			if token == "" && cfg.CloudToken != "" {
				token = cfg.CloudToken
			}
			data := clientGet(server, "/client/machines", token)
			machines, ok := data.([]interface{})
			if !ok || len(machines) == 0 {
				fmt.Println("Nenhuma máquina pessoal conectada.")
				return
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tNOME\tENDEREÇO\tPROVIDER\tPRIORIDADE")
			for _, m := range machines {
				d := m.(map[string]interface{})
				fmt.Fprintf(w, "%.0f\t%s\t%s\t%s\t%.0f\n",
					toFloat(d["id"]), d["name"], d["address"], d["provider"], toFloat(d["priority"]))
			}
			w.Flush()
		},
	}
	myMachinesCmd.Flags().String("server", "http://localhost:8855", "Servidor")
	myMachinesCmd.Flags().String("token", "", "Token")

	// --- Connect Machine ---
	connectCmd := &cobra.Command{
		Use:   "connect [address] [token]",
		Short: "Conectar máquina pessoal",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cfg, _ := cli.LoadConfig()
			server, _ := cmd.Flags().GetString("server")
			authToken, _ := cmd.Flags().GetString("token")
			if server == "http://localhost:8855" && cfg.CloudAPI != "" {
				server = cfg.CloudAPI
			}
			if authToken == "" && cfg.CloudToken != "" {
				authToken = cfg.CloudToken
			}
			name, _ := cmd.Flags().GetString("name")
			provider, _ := cmd.Flags().GetString("provider")

			machineToken := ""
			if len(args) > 1 {
				machineToken = args[1]
			}
			if name == "" {
				name = args[0]
			}

			resp := clientPost(server, "/client/machines", authToken, map[string]interface{}{
				"name": name, "address": args[0], "token": machineToken, "provider": provider,
			})
			if resp != nil {
				fmt.Printf("✅ Máquina '%s' conectada em %s\n", name, args[0])
			}
		},
	}
	connectCmd.Flags().String("server", "http://localhost:8855", "Servidor")
	connectCmd.Flags().String("token", "", "Token auth")
	connectCmd.Flags().String("name", "", "Nome da máquina")
	connectCmd.Flags().String("provider", "personal", "Provider")

	rootCmd.AddCommand(plansCmd, subscribeCmd, myMachinesCmd, connectCmd)
}

// --- HTTP Helpers ---

func clientGet(server, path, token string) interface{} {
	req, _ := http.NewRequest("GET", server+path, nil)
	if token != "" {
		req.Header.Set("Authorization", token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Erro: %v\n", err)
		return nil
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var arr []interface{}
	if json.Unmarshal(body, &arr) == nil {
		return arr
	}
	var obj map[string]interface{}
	json.Unmarshal(body, &obj)
	return obj
}

func clientPost(server, path, token string, data map[string]interface{}) map[string]interface{} {
	b, _ := json.Marshal(data)
	req, _ := http.NewRequest("POST", server+path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Erro: %v\n", err)
		return nil
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)
	if resp.StatusCode >= 400 {
		fmt.Fprintf(os.Stderr, "❌ %s\n", result["error"])
		return nil
	}
	return result
}
