package cmd

import (
	"github.com/spf13/cobra"
)

// clusterCmd represents the create command
var clusterCmd = &cobra.Command{
	Use:     "node",
	Aliases: []string{"nodes"},
	Short:   "A parent command viewing information about your cluster.",
	Long: `A parent command viewing information about your cluster..
	
For example:

	asvec node --help
		`,
}

func init() {
	rootCmd.AddCommand(clusterCmd)
}
