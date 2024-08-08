package cmd

import (
	"github.com/spf13/cobra"
)

// nodeCmd represents the create command
var nodeCmd = &cobra.Command{
	Use:     "node",
	Aliases: []string{"nodes"},
	Short:   "A parent command for viewing information about your nodes.",
	Long: `A parent command for viewing information about your nodes.
	
For example:

	asvec node --help
		`,
}

func init() {
	rootCmd.AddCommand(nodeCmd)
}
