/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "A parent command for creating resources",
	Long: `A parent command for creating resources. It currently only supports creating indexes. 
	For example:
		export ASVEC_HOST=<avs-ip>:5000
		asvec create index -i myindex -n test -s testset -f vector-field -d 256 -m COSINE
		`,
}

func init() {
	rootCmd.AddCommand(createCmd)
}
