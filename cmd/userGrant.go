/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"asvec/cmd/flags"
	"context"
	"log/slog"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

//nolint:govet // Padding not a concern for a CLI
var userGrantFlags = &struct {
	clientFlags flags.ClientFlags
	grantUser   string
	roles       []string
}{
	clientFlags: *flags.NewClientFlags(),
}

func newUserGrantFlagSet() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{} //nolint:lll // For readability                                                                                                                                                                                                //nolint:lll // For readability
	flagSet.AddFlagSet(userGrantFlags.clientFlags.NewClientFlagSet())
	flagSet.StringVar(&userGrantFlags.grantUser, flags.GrantUser, "", "TODO")
	flagSet.StringSliceVar(&userGrantFlags.roles, flags.Roles, []string{}, "TODO")

	return flagSet
}

var userGrantRequiredFlags = []string{
	flags.GrantUser,
	flags.Roles,
}

func newUserGrantCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "grant",
		Short: "A command for granting users roles",
		Long: `A command for creating users. TODO

		For example:
			export ASVEC_HOST=127.0.0.1:5000 ASVEC_USER=admin
			asvec user grant --grant-user foo --roles admin
			`,
		RunE: func(_ *cobra.Command, _ []string) error {
			logger.Debug("parsed flags",
				append(
					userGrantFlags.clientFlags.NewSLogAttr(),
					slog.String(flags.NewUser, userGrantFlags.grantUser),
					slog.Any(flags.Roles, userGrantFlags.roles),
				)...,
			)

			adminClient, err := createClientFromFlags(&userGrantFlags.clientFlags)
			if err != nil {
				return err
			}
			defer adminClient.Close()

			ctx, cancel := context.WithTimeout(context.Background(), userGrantFlags.clientFlags.Timeout)
			defer cancel()

			err = adminClient.GrantRoles(
				ctx,
				userGrantFlags.grantUser,
				userGrantFlags.roles,
			)
			if err != nil {
				logger.Error("unable to grant user roles", slog.String("user", userGrantFlags.grantUser), slog.Any("roles", userGrantFlags.roles), slog.Any("error", err))
				return err
			}

			view.Printf("Successfully granted user %s roles %s", userGrantFlags.grantUser, strings.Join(userGrantFlags.roles, ", "))
			return nil
		},
	}
}

func init() {
	userGrantCmd := newUserGrantCmd()
	userCmd.AddCommand(userGrantCmd)

	// TODO: Add custom template for usage to take into account terminal width
	// Ex: https://github.com/sigstore/cosign/pull/3011/files

	flagSet := newUserGrantFlagSet()
	userGrantCmd.Flags().AddFlagSet(flagSet)

	for _, flag := range userGrantRequiredFlags {
		err := userGrantCmd.MarkFlagRequired(flag)
		if err != nil {
			panic(err)
		}
	}
}
