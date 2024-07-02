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
	Short: "A parent command viewing, creating, and removing indexes.",
	Long: `A parent command viewing, creating, and removing indexes.
	
For example:

	asvec index --help
		`,
}

func init() {
	rootCmd.AddCommand(indexCmd)
}
