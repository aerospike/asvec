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

	commonFlags "github.com/aerospike/tools-common-go/flags"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var userGrantFlags = &struct {
	clientFlags flags.ClientFlags
	grantUser   string
	roles       []string
}{
	clientFlags: *flags.NewClientFlags(),
}

func newUserGrantFlagSet() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}
	flagSet.AddFlagSet(userGrantFlags.clientFlags.NewClientFlagSet())
	flagSet.StringVar(&userGrantFlags.grantUser, flags.Name, "", commonFlags.DefaultWrapHelpString("The existing user to grant new roles"))                                                           //nolint:lll // For readability
	flagSet.StringSliceVar(&userGrantFlags.roles, flags.Roles, []string{}, commonFlags.DefaultWrapHelpString("The roles to grant the existing user. New roles are added to a users existing roles.")) //nolint:lll // For readability

	return flagSet
}

var userGrantRequiredFlags = []string{
	flags.Name,
	flags.Roles,
}

func newUserGrantCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "grant",
		Short: "A command for granting roles to an existing users.",
		Long: fmt.Sprintf(`A command for granting roles to an existing users.

For example:

%s
asvec user grant --%s foo --%s admin
			`, HelpTxtSetupEnv, flags.Name, flags.Roles),
		//nolint:dupl // Ignore code duplication
		RunE: func(_ *cobra.Command, _ []string) error {
			logger.Debug("parsed flags",
				append(
					userGrantFlags.clientFlags.NewSLogAttr(),
					slog.String(flags.Name, userGrantFlags.grantUser),
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
				logger.Error(
					"unable to grant user roles",
					slog.String("user", userGrantFlags.grantUser),
					slog.Any("roles", userGrantFlags.roles),
					slog.Any("error", err),
				)
				return err
			}

			view.Printf(
				"Successfully granted user %s roles %s",
				userGrantFlags.grantUser,
				strings.Join(userGrantFlags.roles, ", "),
			)
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
