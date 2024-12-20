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

var userGrantFlags = &struct {
	clientFlags *flags.ClientFlags
	grantUser   string
	roles       []string
}{
	clientFlags: rootFlags.clientFlags,
}

func newUserGrantFlagSet() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}
	flagSet.StringVar(&userGrantFlags.grantUser, flags.Name, "", "The existing user to grant new roles")                                                           //nolint:lll // For readability
	flagSet.StringSliceVar(&userGrantFlags.roles, flags.Roles, []string{}, "The roles to grant the existing user. New roles are added to a users existing roles.") //nolint:lll // For readability

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
For more information on managing users, refer to: 
https://aerospike.com/docs/vector/operate/user-management

For example:

%s
asvec user grant --%s foo --%s admin
			`, HelpTxtSetupEnv, flags.Name, flags.Roles),
		PreRunE: func(_ *cobra.Command, _ []string) error {
			return checkSeedsAndHost()
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			logger.Debug("parsed flags",
				append(
					userGrantFlags.clientFlags.NewSLogAttr(),
					slog.String(flags.Name, userGrantFlags.grantUser),
					slog.Any(flags.Roles, userGrantFlags.roles),
				)...,
			)

			client, err := createClientFromFlags(userGrantFlags.clientFlags)
			if err != nil {
				return err
			}
			defer client.Close()

			ctx, cancel := context.WithTimeout(context.Background(), userGrantFlags.clientFlags.Timeout)
			defer cancel()

			err = client.GrantRoles(
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
