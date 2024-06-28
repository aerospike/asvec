/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"asvec/cmd/flags"
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

//nolint:govet // Padding not a concern for a CLI
var userRevokeFlags = &struct {
	clientFlags flags.ClientFlags
	revokeUser  string
	roles       []string
}{
	clientFlags: *flags.NewClientFlags(),
}

func newUserRevokeFlagSet() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{} //nolint:lll // For readability                                                                                                                                                                                                //nolint:lll // For readability
	flagSet.AddFlagSet(userRevokeFlags.clientFlags.NewClientFlagSet())
	flagSet.StringVar(&userRevokeFlags.revokeUser, flags.RevokeUser, "", "TODO")
	flagSet.StringSliceVar(&userRevokeFlags.roles, flags.Roles, []string{}, "TODO")

	return flagSet
}

var userRevokeRequiredFlags = []string{
	flags.RevokeUser,
	flags.Roles,
}

func newUserRevokeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "revoke",
		Short: "A command for revoking users roles",
		Long: `A command for revoking users roles. TODO

		For example:
			export ASVEC_HOST=127.0.0.1:5000 ASVEC_USER=admin
			asvec user revoke --revoke-user foo --roles admin
			`,
		PreRunE: func(_ *cobra.Command, _ []string) error {
			if viper.IsSet(flags.Seeds) && viper.IsSet(flags.Host) {
				return fmt.Errorf("only --%s or --%s allowed", flags.Seeds, flags.Host)
			}

			return nil
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			logger.Debug("parsed flags",
				append(
					userRevokeFlags.clientFlags.NewSLogAttr(),
					slog.String(flags.NewUser, userRevokeFlags.revokeUser),
					slog.Any(flags.Roles, userRevokeFlags.roles),
				)...,
			)

			adminClient, err := createClientFromFlags(&userRevokeFlags.clientFlags)
			if err != nil {
				return err
			}
			defer adminClient.Close()

			ctx, cancel := context.WithTimeout(context.Background(), userRevokeFlags.clientFlags.Timeout)
			defer cancel()

			err = adminClient.RevokeRoles(
				ctx,
				userRevokeFlags.revokeUser,
				userRevokeFlags.roles,
			)
			if err != nil {
				logger.Error("unable to revoke user roles", slog.String("user", userRevokeFlags.revokeUser), slog.Any("roles", userRevokeFlags.roles), slog.Any("error", err))
				return err
			}

			view.Printf("Successfully revoked user %s's roles %s", userRevokeFlags.revokeUser, strings.Join(userRevokeFlags.roles, ", "))
			return nil
		},
	}
}

func init() {
	userRevokeCmd := newUserRevokeCmd()
	userCmd.AddCommand(userRevokeCmd)

	// TODO: Add custom template for usage to take into account terminal width
	// Ex: https://github.com/sigstore/cosign/pull/3011/files

	flagSet := newUserRevokeFlagSet()
	userRevokeCmd.Flags().AddFlagSet(flagSet)

	for _, flag := range userRevokeRequiredFlags {
		err := userRevokeCmd.MarkFlagRequired(flag)
		if err != nil {
			panic(err)
		}
	}
}
