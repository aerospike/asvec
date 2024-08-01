package cmd

import (
	"asvec/cmd/flags"
	"context"
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var userListFlags = &struct {
	format      int
	clientFlags flags.ClientFlags
}{
	clientFlags: *flags.NewClientFlags(),
}

func newUserListFlagSet() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}

	flags.AddFormatTestFlag(flagSet, &userListFlags.format)
	flagSet.AddFlagSet(userListFlags.clientFlags.NewClientFlagSet())

	return flagSet
}

var userListRequiredFlags = []string{}

// listIndexCmd represents the listIndex command
func newUserListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "ls",
		Aliases: []string{"list"},
		Short:   "A command for listing users",
		Long: fmt.Sprintf(`A command for listing useful information about AVS users.
For example:

%s
asvec user ls
		`, HelpTxtSetupEnv),
		PreRunE: func(_ *cobra.Command, _ []string) error {
			return checkSeedsAndHost()
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			logger.Debug("parsed flags",
				userListFlags.clientFlags.NewSLogAttr()...,
			)

			adminClient, err := createClientFromFlags(&userListFlags.clientFlags)
			if err != nil {
				return err
			}
			defer adminClient.Close()

			ctx, cancel := context.WithTimeout(context.Background(), userListFlags.clientFlags.Timeout)
			defer cancel()

			userList, err := adminClient.ListUsers(ctx)
			if err != nil {
				logger.Error("failed to list users", slog.Any("error", err))
				return err
			}

			logger.Debug("server user list", slog.String("response", userList.String()))

			view.PrintUsers(userList, userListFlags.format)
			view.Print("Use 'role list' to view available roles")

			return nil
		},
	}
}

func init() {
	userListCmd := newUserListCmd()

	userCmd.AddCommand(userListCmd)
	userListCmd.Flags().AddFlagSet(newUserListFlagSet())

	for _, flag := range userListRequiredFlags {
		err := userListCmd.MarkFlagRequired(flag)
		if err != nil {
			panic(err)
		}
	}
}
