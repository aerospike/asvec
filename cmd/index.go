/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// indexCmd represents the create command
var indexCmd = &cobra.Command{
	Use:   "index",
	Short: "A parent command for creating resources",
	Long: `A parent command for creating resources. It currently only supports creating indexes. 
	For example:
		export ASVEC_HOST=<avs-ip>:5000
		asvec index ---help
		`,
}

func init() {
	rootCmd.AddCommand(indexCmd)
}
