/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"asvec/cmd/flags"
	"context"
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var userListFlags = &struct {
	clientFlags flags.ClientFlags
}{
	clientFlags: *flags.NewClientFlags(),
}

func newUserListFlagSet() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}

	flagSet.AddFlagSet(userListFlags.clientFlags.NewClientFlagSet())

	return flagSet
}

var userListRequiredFlags = []string{}

// listIndexCmd represents the listIndex command
func newUserListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "A command for listing users",
		Long: `A command for displaying useful information about AVS users.
	For example:
		export ASVEC_HOST=<avs-ip>:5000
		asvec user list
		`,
		RunE: func(_ *cobra.Command, _ []string) error {
			logger.Debug("parsed flags",
				userListFlags.clientFlags.NewSLogAttr()...,
			)

			adminClient, err := createClientFromFlags(&userListFlags.clientFlags)
			if err != nil {
				return err
			}
			defer adminClient.Close()

			ctx, cancel := context.WithTimeout(context.Background(), userListFlags.clientFlags.Timeout)
			defer cancel()

			userList, err := adminClient.ListUsers(ctx)
			if err != nil {
				logger.Error("failed to list users", slog.Any("error", err))
				return err
			}

			cancel()

			logger.Debug("server user list", slog.String("response", userList.String()))

			view.PrintUsers(userList)
			view.Print("Use 'role list' to view available roles")

			return nil
		},
	}
}

func init() {
	userListCmd := newUserListCmd()

	userCmd.AddCommand(userListCmd)
	userListCmd.Flags().AddFlagSet(newUserListFlagSet())

	for _, flag := range userListRequiredFlags {
		err := userListCmd.MarkFlagRequired(flag)
		if err != nil {
			panic(err)
		}
	}
}
