package cmd

import (
	"asvec/cmd/flags"
	"context"
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var userNewPassFlags = &struct {
	clientFlags *flags.ClientFlags
	username    string
	password    string
}{
	clientFlags: rootFlags.clientFlags,
}

func newUserNewPassFlagSet() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}
	flagSet.StringVar(&userNewPassFlags.username, flags.Name, "", "The name of the user.")                                                                                                     //nolint:lll // For readability
	flagSet.StringVar(&userNewPassFlags.password, flags.NewPassword, "", "The new password for the user. If a new password is not provided you you will be prompted to enter a new password.") //nolint:lll // For readability

	return flagSet
}

var userNewPassRequiredFlags = []string{
	flags.Name,
}

// createUserCmd represents the createIndex command
func newUserNewPasswordCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "new-password",
		Aliases: []string{"new-pass"},
		Short:   "Change the password for a user",
		Long: fmt.Sprintf(`A command for changing the password for an existing user.
For more information on managing users, refer to: 
https://aerospike.com/docs/vector/operate/user-management

For example:

%s
asvec user new-password --%s foo
			`, HelpTxtSetupEnv, flags.Name),
		PreRunE: func(_ *cobra.Command, _ []string) error {
			return checkSeedsAndHost()
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			logger.Debug("parsed flags",
				append(
					userNewPassFlags.clientFlags.NewSLogAttr(),
					slog.String(flags.Name, userNewPassFlags.username),
				)...,
			)

			client, err := createClientFromFlags(userNewPassFlags.clientFlags)
			if err != nil {
				return err
			}
			defer client.Close()

			if userNewPassFlags.password == "" {
				userNewPassFlags.password, err = passwordPrompt("New Password: ")
				if err != nil {
					logger.Error("failed to read new password", slog.Any("error", err))
					return err
				}
			}

			ctx, cancel := context.WithTimeout(context.Background(), userNewPassFlags.clientFlags.Timeout)
			defer cancel()

			err = client.UpdateCredentials(
				ctx,
				userNewPassFlags.username,
				userNewPassFlags.password,
			)
			if err != nil {
				logger.Error(
					"unable to update user credentials",
					slog.String("user", userNewPassFlags.username),
					slog.Any("error", err),
				)
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
