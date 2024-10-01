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
	clientFlags *flags.ClientFlags
	format      int // For testing. Hidden
}{
	clientFlags: rootFlags.clientFlags,
}

func newUserListFlagSet() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}

	flagSet.AddFlagSet(userListFlags.clientFlags.NewClientFlagSet())

	err := flags.AddFormatTestFlag(flagSet, &userListFlags.format)
	if err != nil {
		panic(err)
	}

	return flagSet
}

// listIndexCmd represents the listIndex command
func newUserListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "ls",
		Aliases: []string{"list"},
		Short:   "A command for listing users",
		Long: fmt.Sprintf(`A command for listing useful information about AVS users.
For more information on managing users, refer to: 
https://aerospike.com/docs/vector/operate/user-management


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

			client, err := createClientFromFlags(userListFlags.clientFlags)
			if err != nil {
				return err
			}
			defer client.Close()

			ctx, cancel := context.WithTimeout(context.Background(), userListFlags.clientFlags.Timeout)
			defer cancel()

			userList, err := client.ListUsers(ctx)
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
}
