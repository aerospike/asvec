/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "A parent command for listing resources",
	Long: `A parent command for listings resources. It currently only supports listing indexes. 
	For example:
		export ASVEC_HOST=<avs-ip>:5000
		asvec list index
		`,
}

func init() {
	rootCmd.AddCommand(listCmd)
}
