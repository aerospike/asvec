package cmd

import (
	"asvec/cmd/flags"
	"context"
	"fmt"
	"log/slog"

	commonFlags "github.com/aerospike/tools-common-go/flags"
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
	flagSet := &pflag.FlagSet{}
	flagSet.AddFlagSet(userDropFlags.clientFlags.NewClientFlagSet())
	flagSet.StringVar(&userDropFlags.dropUser, flags.Name, "", commonFlags.DefaultWrapHelpString("The name of the user to drop.")) //nolint:lll // For readability

	return flagSet
}

var userDropRequiredFlags = []string{
	flags.Name,
}

func newUserDropCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "drop",
		Short: "A command for dropping users",
		Long: fmt.Sprintf(`A command for dropping users.

For example:

%s
asvec user drop --%s foo
			`, HelpTxtSetupEnv, flags.Name),
		RunE: func(_ *cobra.Command, _ []string) error {
			logger.Debug("parsed flags",
				append(
					userDropFlags.clientFlags.NewSLogAttr(),
					slog.String(flags.Name, userDropFlags.dropUser),
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
