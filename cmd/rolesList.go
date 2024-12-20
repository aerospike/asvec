package cmd

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/aerospike/asvec/module-rename/cmd/flags"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var rolesListFlags = &struct {
	clientFlags *flags.ClientFlags
	format      int // For testing. Hidden
}{
	clientFlags: rootFlags.clientFlags,
}

func newRoleListFlagSet() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}

	err := flags.AddFormatTestFlag(flagSet, &rolesListFlags.format)
	if err != nil {
		panic(err)
	}

	return flagSet
}

// listIndexCmd represents the listIndex command
func newRoleListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "ls",
		Aliases: []string{"list"},
		Short:   "A command for listing roles",
		Long: fmt.Sprintf(`A command for listing roles.

For example:

%s
asvec role ls
		`, HelpTxtSetupEnv),
		PreRunE: func(_ *cobra.Command, _ []string) error {
			return checkSeedsAndHost()
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			logger.Debug("parsed flags",
				rolesListFlags.clientFlags.NewSLogAttr()...,
			)

			client, err := createClientFromFlags(rolesListFlags.clientFlags)
			if err != nil {
				return err
			}
			defer client.Close()

			ctx, cancel := context.WithTimeout(context.Background(), rolesListFlags.clientFlags.Timeout)
			defer cancel()

			userList, err := client.ListRoles(ctx)
			if err != nil {
				logger.Error("failed to list roles", slog.Any("error", err))
				return err
			}

			logger.Debug("server role list", slog.String("response", userList.String()))

			view.PrintRoles(userList, rolesListFlags.format)

			return nil
		},
	}
}

func init() {
	roleListCmd := newRoleListCmd()

	roleCmd.AddCommand(roleListCmd)
	roleListCmd.Flags().AddFlagSet(newRoleListFlagSet())
}
