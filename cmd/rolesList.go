/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"asvec/cmd/flags"
	"context"
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var rolesListFlags = &struct {
	clientFlags flags.ClientFlags
}{
	clientFlags: *flags.NewClientFlags(),
}

func newRoleListFlagSet() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}

	flagSet.AddFlagSet(rolesListFlags.clientFlags.NewClientFlagSet())

	return flagSet
}

var roleListRequiredFlags = []string{}

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
			if viper.IsSet(flags.Seeds) && viper.IsSet(flags.Host) {
				return fmt.Errorf("only --%s or --%s allowed", flags.Seeds, flags.Host)
			}

			return nil
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			logger.Debug("parsed flags",
				rolesListFlags.clientFlags.NewSLogAttr()...,
			)

			adminClient, err := createClientFromFlags(&rolesListFlags.clientFlags)
			if err != nil {
				return err
			}
			defer adminClient.Close()

			ctx, cancel := context.WithTimeout(context.Background(), rolesListFlags.clientFlags.Timeout)
			defer cancel()

			userList, err := adminClient.ListRoles(ctx)
			if err != nil {
				logger.Error("failed to list roles", slog.Any("error", err))
				return err
			}

			cancel()

			logger.Debug("server role list", slog.String("response", userList.String()))

			view.PrintRoles(userList)

			return nil
		},
	}
}

func init() {
	roleListCmd := newRoleListCmd()

	roleCmd.AddCommand(roleListCmd)
	roleListCmd.Flags().AddFlagSet(newRoleListFlagSet())

	for _, flag := range roleListRequiredFlags {
		err := roleListCmd.MarkFlagRequired(flag)
		if err != nil {
			panic(err)
		}
	}
}
