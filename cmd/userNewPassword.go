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
var userNewPassFlags = &struct {
	clientFlags flags.ClientFlags
	username    string
	password    string
	roles       []string
}{
	clientFlags: *flags.NewClientFlags(),
}

func newUserNewPassFlagSet() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{} //nolint:lll // For readability                                                                                                                                                                                                //nolint:lll // For readability
	flagSet.AddFlagSet(userNewPassFlags.clientFlags.NewClientFlagSet())
	flagSet.StringVar(&userNewPassFlags.username, flags.Username, "", "TODO")
	flagSet.StringVar(&userNewPassFlags.password, flags.NewPassword, "", "TODO")

	return flagSet
}

var userNewPassRequiredFlags = []string{
	flags.Username,
}

// createUserCmd represents the createIndex command
func newUserNewPasswordCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "new-password",
		Aliases: []string{"new-pass"},
		Short:   "A command for creating users",
		Long: `A command for creating users. TODO

		For example:
			export ASVEC_HOST=127.0.0.1:5000 ASVEC_USER=admin
			asvec user new-password --name
			`,

		RunE: func(_ *cobra.Command, _ []string) error {
			logger.Debug("parsed flags",
				append(
					userNewPassFlags.clientFlags.NewSLogAttr(),
					slog.String(flags.Username, userNewPassFlags.username),
					slog.Any(flags.Roles, userNewPassFlags.roles),
				)...,
			)

			adminClient, err := createClientFromFlags(&userNewPassFlags.clientFlags)
			if err != nil {
				return err
			}
			defer adminClient.Close()

			if userNewPassFlags.password == "" {
				userNewPassFlags.password, err = passwordPrompt("New Password: ")
				if err != nil {
					logger.Error("failed to read new password", slog.Any("error", err))
					return err
				}
			}

			ctx, cancel := context.WithTimeout(context.Background(), userNewPassFlags.clientFlags.Timeout)
			defer cancel()

			err = adminClient.UpdateCredentials(
				ctx,
				userNewPassFlags.username,
				userNewPassFlags.password,
			)
			if err != nil {
				logger.Error("unable to update user credentials", slog.String("user", userNewPassFlags.username), slog.Any("error", err))
				return err
			}

			view.Printf("Successfully updated user %s's credentials", userNewPassFlags.username)
			return nil
		},
	}
}

func init() {
	userNewPassCmd := newUserNewPasswordCmd()
	userCmd.AddCommand(userNewPassCmd)

	// TODO: Add custom template for usage to take into account terminal width
	// Ex: https://github.com/sigstore/cosign/pull/3011/files

	flagSet := newUserNewPassFlagSet()
	userNewPassCmd.Flags().AddFlagSet(flagSet)

	for _, flag := range userNewPassRequiredFlags {
		err := userNewPassCmd.MarkFlagRequired(flag)
		if err != nil {
			panic(err)
		}
	}
}
