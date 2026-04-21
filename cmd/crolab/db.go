// Copyright (c) 2026 Crolab Contributors. All rights reserved.
// Licensed under the Crolab Sustainable License (CSL).
package main

import (
	"fmt"
	"log"

	"github.com/crolab/core/internal/cloud"
	"github.com/spf13/cobra"
)

var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Gerencia persistência de dados (Backups, Snapshots SQLite)",
}

var dbBackupCmd = &cobra.Command{
	Use:   "backup [destino.db]",
	Short: "Exporta banco WAL em snapshot real-time (Ex: crolab db backup s3-sync.db)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		destPath := args[0]
		dbPath, _ := cmd.Flags().GetString("db")

		fmt.Printf("📦 Inicializando Backup WAL Runtime do banco %s para %s...\n", dbPath, destPath)

		if err := cloud.InitDB(dbPath); err != nil {
			log.Fatalf("❌ Falha DB: %v", err)
		}

		if err := cloud.DBBackup(destPath); err != nil {
			log.Fatalf("❌ Falha ao realizar Snapshot WAL P2P: %v", err)
		}

		fmt.Printf("✅ DB Snapshot Snapshot concluído com sucesso e gravado em %s\n", destPath)
		fmt.Println("👉 Dica: Execute 'aws s3 cp' se quiser sincronizar este arquivo na Cloud. O WAL original permanece intacto!")
	},
}

func init() {
	dbBackupCmd.Flags().String("db", "crolab.db", "Caminho do bash SQLite local")
	dbCmd.AddCommand(dbBackupCmd)
	rootCmd.AddCommand(dbCmd)
}
