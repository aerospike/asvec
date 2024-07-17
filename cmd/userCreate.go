package cmd

import (
	"asvec/cmd/flags"
	"context"
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var userCreateFlags = &struct {
	clientFlags flags.ClientFlags
	newUsername string
	newPassword string
	roles       []string
}{
	clientFlags: *flags.NewClientFlags(),
}

func newUserCreateFlagSet() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}
	flagSet.AddFlagSet(userCreateFlags.clientFlags.NewClientFlagSet())
	flagSet.StringVar(&userCreateFlags.newUsername, flags.Name, "", "The name of the new user.")                                                                                                 //nolint:lll // For readability
	flagSet.StringVar(&userCreateFlags.newPassword, flags.NewPassword, "", "The password for the new user. If a new password is not provided you you will be prompted to enter a new password.") //nolint:lll // For readability
	flagSet.StringSliceVar(&userCreateFlags.roles, flags.Roles, []string{}, "The roles to assign to the new user. To see valid roles run 'asvec role ls'.")                                      //nolint:lll // For readability

	return flagSet
}

var userCreateRequiredFlags = []string{
	flags.Name,
	flags.Roles,
}

// createUserCmd represents the createIndex command
func newUserCreateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create",
		Short: "A command for creating new users",
		Long: fmt.Sprintf(`A command for creating new users. Users are assigned 
roles which have certain privileges. Users should have the minimum number of
roles necessary to perform their tasks.

For example:

%s
asvec user create --%s foo --%s read-write
			`, HelpTxtSetupEnv, flags.Name, flags.Roles),
		PreRunE: func(_ *cobra.Command, _ []string) error {
			return checkSeedsAndHost()
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			logger.Debug("parsed flags",
				append(
					userCreateFlags.clientFlags.NewSLogAttr(),
					slog.String(flags.Name, userCreateFlags.newUsername),
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
