package cmd

import (
	"github.com/spf13/cobra"
)

// userCmd represents the create command
var roleCmd = &cobra.Command{
	Use:     "role",
	Aliases: []string{"roles"},
	Short:   "A parent command for viewing roles.",
	Long: `A parent command for listing, creating, dropping, and granting roles to users. 
	For example:
		export ASVEC_HOST=<avs-ip>:5000
		asvec user list
		`,
}

func init() {
	rootCmd.AddCommand(roleCmd)
}
