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
var userDropFlags = &struct {
	clientFlags flags.ClientFlags
	dropUser    string
}{
	clientFlags: *flags.NewClientFlags(),
}

func newUserDropFlagSet() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{} //nolint:lll // For readability                                                                                                                                                                                                //nolint:lll // For readability
	flagSet.AddFlagSet(userDropFlags.clientFlags.NewClientFlagSet())
	flagSet.StringVar(&userDropFlags.dropUser, flags.DropUser, "", "TODO")

	return flagSet
}

var userDropRequiredFlags = []string{
	flags.DropUser,
}

func newUserDropCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "drop",
		Short: "A command for dropping users",
		Long: `A command for dropping users. TODO

		For example:
			export ASVEC_HOST=127.0.0.1:5000 ASVEC_USER=admin
			asvec user drop --new-user foo --roles read-write
			`,
		RunE: func(_ *cobra.Command, _ []string) error {
			logger.Debug("parsed flags",
				append(
					userDropFlags.clientFlags.NewSLogAttr(),
					slog.String(flags.NewUser, userDropFlags.dropUser),
				)...,
			)

			adminClient, err := createClientFromFlags(&userDropFlags.clientFlags)
			if err != nil {
				return err
			}
			defer adminClient.Close()

			ctx, cancel := context.WithTimeout(context.Background(), userDropFlags.clientFlags.Timeout)
			defer cancel()

			err = adminClient.DropUser(
				ctx,
				userDropFlags.dropUser,
			)
			if err != nil {
				logger.Error("unable to create user", slog.String("user", userDropFlags.dropUser), slog.Any("error", err))
				return err
			}

			view.Printf("Successfully dropped user %s", userDropFlags.dropUser)
			return nil
		},
	}
}

func init() {
	userDropCmd := newUserDropCmd()
	userCmd.AddCommand(userDropCmd)

	// TODO: Add custom template for usage to take into account terminal width
	// Ex: https://github.com/sigstore/cosign/pull/3011/files

	flagSet := newUserDropFlagSet()
	userDropCmd.Flags().AddFlagSet(flagSet)

	for _, flag := range userDropRequiredFlags {
		err := userDropCmd.MarkFlagRequired(flag)
		if err != nil {
			panic(err)
		}
	}
}
