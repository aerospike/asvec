//nolint:dupl // Ignore code duplication
package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/aerospike/asvec/cmd/flags"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var userRevokeFlags = &struct {
	clientFlags *flags.ClientFlags
	revokeUser  string
	roles       []string
}{
	clientFlags: rootFlags.clientFlags,
}

func newUserRevokeFlagSet() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}
	flagSet.StringVar(&userRevokeFlags.revokeUser, flags.Name, "", "The existing user to revoke new roles.")                                                      //nolint:lll // For readability
	flagSet.StringSliceVar(&userRevokeFlags.roles, flags.Roles, []string{}, "The roles to revoke from the user. Roles are removed from a user's existing roles.") //nolint:lll // For readability

	return flagSet
}

var userRevokeRequiredFlags = []string{
	flags.Name,
	flags.Roles,
}

func newUserRevokeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "revoke",
		Short: "A command for revoking roles from an existing user.",
		Long: fmt.Sprintf(`A command for revoking roles from an existing user.
For more information on managing users, refer to: 
https://aerospike.com/docs/vector/operate/user-management

For example:

%s
asvec user revoke --%s foo --%s admin
			`, HelpTxtSetupEnv, flags.Name, flags.Roles),
		PreRunE: func(_ *cobra.Command, _ []string) error {
			return checkSeedsAndHost()
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			logger.Debug("parsed flags",
				append(
					userRevokeFlags.clientFlags.NewSLogAttr(),
					slog.String(flags.Name, userRevokeFlags.revokeUser),
					slog.Any(flags.Roles, userRevokeFlags.roles),
				)...,
			)

			client, err := createClientFromFlags(userRevokeFlags.clientFlags)
			if err != nil {
				return err
			}
			defer client.Close()

			ctx, cancel := context.WithTimeout(context.Background(), userRevokeFlags.clientFlags.Timeout)
			defer cancel()

			err = client.RevokeRoles(
				ctx,
				userRevokeFlags.revokeUser,
				userRevokeFlags.roles,
			)
			if err != nil {
				logger.Error(
					"unable to revoke user roles",
					slog.String("user", userRevokeFlags.revokeUser),
					slog.Any("roles", userRevokeFlags.roles),
					slog.Any("error", err),
				)
				return err
			}

			view.Printf(
				"Successfully revoked user %s's roles %s",
				userRevokeFlags.revokeUser,
				strings.Join(userRevokeFlags.roles, ", "),
			)
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
