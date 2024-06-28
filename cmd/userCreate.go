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

//nolint:govet // Padding not a concern for a CLI
var userCreateFlags = &struct {
	clientFlags flags.ClientFlags
	newUsername string
	newPassword string
	roles       []string
}{
	clientFlags: *flags.NewClientFlags(),
}

func newUserCreateFlagSet() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{} //nolint:lll // For readability                                                                                                                                                                                                //nolint:lll // For readability
	flagSet.AddFlagSet(userCreateFlags.clientFlags.NewClientFlagSet())
	flagSet.StringVar(&userCreateFlags.newUsername, flags.NewUser, "", "TODO")
	flagSet.StringVar(&userCreateFlags.newPassword, flags.NewPassword, "", "TODO")
	flagSet.StringSliceVar(&userCreateFlags.roles, flags.Roles, []string{}, "TODO")

	return flagSet
}

var userCreateRequiredFlags = []string{
	flags.NewUser,
	flags.Roles,
}

// createUserCmd represents the createIndex command
func newUserCreateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create",
		Short: "A command for creating users",
		Long: `A command for creating users. TODO

		For example:
			export ASVEC_HOST=127.0.0.1:5000 ASVEC_USER=admin
			asvec user create --new-user foo --roles read-write
			`,
		RunE: func(_ *cobra.Command, _ []string) error {
			logger.Debug("parsed flags",
				append(
					userCreateFlags.clientFlags.NewSLogAttr(),
					slog.String(flags.NewUser, userCreateFlags.newUsername),
					slog.Any(flags.Roles, userCreateFlags.roles),
				)...,
			)

			adminClient, err := createClientFromFlags(&userCreateFlags.clientFlags)
			if err != nil {
				return err
			}
			defer adminClient.Close()

			if userCreateFlags.newPassword == "" {
				userCreateFlags.newPassword, err = passwordPrompt("New User Password: ")
				if err != nil {
					logger.Error("failed to read new password", slog.Any("error", err))
					return err
				}
			}

			ctx, cancel := context.WithTimeout(context.Background(), userCreateFlags.clientFlags.Timeout)
			defer cancel()

			err = adminClient.CreateUser(
				ctx,
				userCreateFlags.newUsername,
				userCreateFlags.newPassword,
				userCreateFlags.roles,
			)
			if err != nil {
				logger.Error("unable to create user", slog.String("user", userCreateFlags.newUsername), slog.Any("error", err))
				return err
			}

			view.Printf("Successfully created user %s", userCreateFlags.newUsername)
			return nil
		},
	}
}

func init() {
	userCreateCmd := newUserCreateCmd()
	userCmd.AddCommand(userCreateCmd)

	// TODO: Add custom template for usage to take into account terminal width
	// Ex: https://github.com/sigstore/cosign/pull/3011/files

	flagSet := newUserCreateFlagSet()
	userCreateCmd.Flags().AddFlagSet(flagSet)

	for _, flag := range userCreateRequiredFlags {
		err := userCreateCmd.MarkFlagRequired(flag)
		if err != nil {
			panic(err)
		}
	}
}
