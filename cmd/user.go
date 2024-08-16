package cmd

import (
	"github.com/spf13/cobra"
)

// userCmd represents the create command
var userCmd = &cobra.Command{
	Use:     "user",
	Aliases: []string{"users"},
	Short:   "A parent command for viewing and configuring users.",
	Long: `A parent command for listing, creating, dropping, and granting roles to users. 
For more information on managing users, refer to: 
https://aerospike.com/docs/vector/operate/user-management

For example:

asvec user --help
		`,
}

func init() {
	rootCmd.AddCommand(userCmd)
}
