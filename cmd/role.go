package cmd

import (
	"github.com/spf13/cobra"
)

// userCmd represents the create command
var roleCmd = &cobra.Command{
	Use:     "role",
	Aliases: []string{"roles"},
	Short:   "A parent command for listing roles.",
	Long: `A parent command for listing roles. Other sub-commands will be added
in the future. 

For example:

asvec role --help
	`,
}

func init() {
	rootCmd.AddCommand(roleCmd)
}
