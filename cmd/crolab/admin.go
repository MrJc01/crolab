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

	"github.com/spf13/cobra"
)

var adminServer string

func init() {
	adminCmd := &cobra.Command{
		Use:   "admin",
		Short: "Comandos administrativos (planos, pool, máquinas, usuários)",
	}
	adminCmd.PersistentFlags().StringVar(&adminServer, "server", GetCloudAPI(), "Endereço do servidor admin")
	adminCmd.PersistentFlags().String("token", "", "Token de autenticação admin")

	// --- Plans ---
	planCmd := &cobra.Command{Use: "plan", Short: "Gerenciar planos"}

	planCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "Listar planos",
		Run: func(cmd *cobra.Command, args []string) {
			token, _ := cmd.Flags().GetString("token")
			data := adminGet("/admin/plans", token)
			plans, ok := data.([]interface{})
			if !ok || len(plans) == 0 {
				fmt.Println("Nenhum plano encontrado.")
				return
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tNOME\tVRAM\tSTORAGE\t$/HORA\t$/MÊS")
			for _, p := range plans {
				m := p.(map[string]interface{})
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t$%.2f\t$%.2f\n",
					m["id"], m["name"], m["vram"], m["storage"],
					toFloat(m["price_hr"]), toFloat(m["price_month"]))
			}
			w.Flush()
		},
	})

	planCreateCmd := &cobra.Command{
		Use:   "create [id] [name]",
		Short: "Criar plano",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			token, _ := cmd.Flags().GetString("token")
			vram, _ := cmd.Flags().GetString("vram")
			storage, _ := cmd.Flags().GetString("storage")
			priceHr, _ := cmd.Flags().GetFloat64("price")
			priceMonth, _ := cmd.Flags().GetFloat64("monthly")

			body := map[string]interface{}{
				"id": args[0], "name": args[1],
				"vram": vram, "storage": storage,
				"price_hr": priceHr, "price_month": priceMonth,
			}
			resp := adminPost("/admin/plans", token, body)
			if resp != nil {
				fmt.Printf("✅ Plano '%s' criado\n", args[1])
			}
		},
	}
	planCreateCmd.Flags().String("vram", "", "VRAM (ex: 6GB)")
	planCreateCmd.Flags().String("storage", "", "Storage (ex: 100GB)")
	planCreateCmd.Flags().Float64("price", 0, "Preço/hora")
	planCreateCmd.Flags().Float64("monthly", 0, "Preço/mês")
	planCmd.AddCommand(planCreateCmd)

	planCmd.AddCommand(&cobra.Command{
		Use:   "delete [id]",
		Short: "Remover plano",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			token, _ := cmd.Flags().GetString("token")
			adminDelete("/admin/plans/"+args[0], token)
			fmt.Printf("🗑️  Plano '%s' removido\n", args[0])
		},
	})

	// --- Pool ---
	poolCmd := &cobra.Command{Use: "pool", Short: "Gerenciar pool de prioridade"}

	poolCmd.AddCommand(&cobra.Command{
		Use:   "list [planID]",
		Short: "Listar pool de um plano",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			token, _ := cmd.Flags().GetString("token")
			data := adminGet("/admin/pool/"+args[0], token)
			entries, ok := data.([]interface{})
			if !ok || len(entries) == 0 {
				fmt.Println("Pool vazio.")
				return
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "PRIORIDADE\tPROVIDER\tLABEL\tENDEREÇO")
			for _, e := range entries {
				m := e.(map[string]interface{})
				fmt.Fprintf(w, "%v\t%s\t%s\t%s\n",
					m["priority"], m["provider"], m["label"], m["address"])
			}
			w.Flush()
		},
	})

	poolAddCmd := &cobra.Command{
		Use:   "add [planID] [address]",
		Short: "Adicionar entrada ao pool",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			token, _ := cmd.Flags().GetString("token")
			provider, _ := cmd.Flags().GetString("provider")
			label, _ := cmd.Flags().GetString("label")
			priority, _ := cmd.Flags().GetInt("priority")

			body := map[string]interface{}{
				"priority": priority, "provider": provider,
				"label": label, "address": args[1],
			}
			adminPost("/admin/pool/"+args[0], token, body)
			fmt.Printf("✅ Pool entry adicionada ao plano '%s'\n", args[0])
		},
	}
	poolAddCmd.Flags().String("provider", "", "Provider (vastai, runpod...)")
	poolAddCmd.Flags().String("label", "", "Label descritivo")
	poolAddCmd.Flags().Int("priority", 1, "Prioridade (1=topo)")
	poolCmd.AddCommand(poolAddCmd)

	// --- Machines ---
	machinesCmd := &cobra.Command{
		Use:   "machines",
		Short: "Listar máquinas",
		Run: func(cmd *cobra.Command, args []string) {
			token, _ := cmd.Flags().GetString("token")
			data := adminGet("/admin/machines", token)
			machines, ok := data.([]interface{})
			if !ok || len(machines) == 0 {
				fmt.Println("Nenhuma máquina.")
				return
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tNOME\tGPU\tVRAM\t$/H\tSTATUS\tPROVIDER")
			for _, m := range machines {
				d := m.(map[string]interface{})
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t$%.2f\t%s\t%s\n",
					d["id"], d["name"], d["gpu"], d["vram"],
					toFloat(d["price_hr"]), d["status"], d["provider"])
			}
			w.Flush()
		},
	}

	// --- Users ---
	usersCmd := &cobra.Command{
		Use:   "users",
		Short: "Listar usuários",
		Run: func(cmd *cobra.Command, args []string) {
			token, _ := cmd.Flags().GetString("token")
			data := adminGet("/admin/users", token)
			users, ok := data.([]interface{})
			if !ok || len(users) == 0 {
				fmt.Println("Nenhum usuário.")
				return
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tEMAIL\tROLE\tCRÉDITOS")
			for _, u := range users {
				d := u.(map[string]interface{})
				fmt.Fprintf(w, "%.0f\t%s\t%s\t$%.2f\n",
					toFloat(d["ID"]), d["Email"], d["Role"], toFloat(d["Credits"]))
			}
			w.Flush()
		},
	}

	// --- Metrics ---
	metricsCmd := &cobra.Command{
		Use:   "metrics",
		Short: "Ver métricas do dashboard",
		Run: func(cmd *cobra.Command, args []string) {
			token, _ := cmd.Flags().GetString("token")
			data := adminGet("/admin/dashboard", token)
			if m, ok := data.(map[string]interface{}); ok {
				fmt.Printf("👥 Usuários:  %.0f\n", toFloat(m["users_total"]))
				fmt.Printf("📋 Planos:    %.0f\n", toFloat(m["plans_total"]))
				fmt.Printf("🖥️  Máquinas:  %.0f\n", toFloat(m["machines_total"]))
				fmt.Printf("🟢 Online:    %.0f\n", toFloat(m["machines_online"]))
			}
		},
	}

	planCmd.PersistentFlags().String("token", "", "Token admin")
	poolCmd.PersistentFlags().String("token", "", "Token admin")

	adminCmd.AddCommand(planCmd, poolCmd, machinesCmd, usersCmd, metricsCmd)
	rootCmd.AddCommand(adminCmd)
}

// --- HTTP Helpers ---

func adminGet(path, token string) interface{} {
	req, _ := http.NewRequest("GET", adminServer+path, nil)
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

	// Try array first, then object
	var arr []interface{}
	if json.Unmarshal(body, &arr) == nil {
		return arr
	}
	var obj map[string]interface{}
	json.Unmarshal(body, &obj)
	return obj
}

func adminPost(path, token string, data map[string]interface{}) map[string]interface{} {
	b, _ := json.Marshal(data)
	req, _ := http.NewRequest("POST", adminServer+path, bytes.NewReader(b))
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

func adminDelete(path, token string) {
	req, _ := http.NewRequest("DELETE", adminServer+path, nil)
	if token != "" {
		req.Header.Set("Authorization", token)
	}
	http.DefaultClient.Do(req)
}

func toFloat(v interface{}) float64 {
	if f, ok := v.(float64); ok {
		return f
	}
	return 0
}
