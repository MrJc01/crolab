// Copyright (c) 2026 Crolab Contributors. All rights reserved.
// Licensed under the Crolab Sustainable License (CSL).
// Contact: mrj.crom@gmail.com
package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Mostra a versão do Crolab",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Crolab CLI versão %s\n", Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
