/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// dropCmd represents the drop command
var dropCmd = &cobra.Command{
	Use:   "drop",
	Short: "A parent command for dropping resources",
	Long: `A parent command for dropping resources. It currently only supports dropping indexes. 
	For example:
		export ASVEC_HOST=<avs-ip>:5000
		asvec drop index -i myindex -n test
		`,
}

func init() {
	rootCmd.AddCommand(dropCmd)
}
