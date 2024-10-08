package cmd

import (
	"asvec/cmd/flags"
	"context"
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

//nolint:govet // Padding not a concern for a CLI
var indexDropFlags = &struct {
	clientFlags *flags.ClientFlags
	yes         bool
	namespace   string
	indexName   string
}{
	clientFlags: rootFlags.clientFlags,
}

func newIndexDropFlagSet() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}
	flagSet.BoolVarP(&indexDropFlags.yes, flags.Yes, "y", false, "When true do not prompt for confirmation.")                //nolint:lll // For readability
	flagSet.StringVarP(&indexDropFlags.namespace, flags.Namespace, flags.NamespaceShort, "", "The namespace for the index.") //nolint:lll // For readability
	flagSet.StringVarP(&indexDropFlags.indexName, flags.IndexName, flags.IndexNameShort, "", "The name of the index.")       //nolint:lll // For readability

	return flagSet
}

var indexDropRequiredFlags = []string{
	flags.Namespace,
	flags.IndexName,
}

// dropIndexCmd represents the dropIndex command
func newIndexDropCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "drop",
		Short: "A command for dropping indexes",
		Long: fmt.Sprintf(`A command for dropping indexes. Deleting an index will free up 
storage but will also disable vector search on your data. For guidance on 
managing indexes, refer to: https://aerospike.com/docs/vector/operate/index-management

For example:

%s
asvec index drop -i myindex -n test
			`, HelpTxtSetupEnv),
		PreRunE: func(_ *cobra.Command, _ []string) error {
			return checkSeedsAndHost()
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			logger.Debug("parsed flags",
				append(indexDropFlags.clientFlags.NewSLogAttr(),
					slog.Bool(flags.Yes, indexDropFlags.yes),
					slog.String(flags.Namespace, indexDropFlags.namespace),
					slog.String(flags.IndexName, indexDropFlags.indexName),
				)...,
			)

			client, err := createClientFromFlags(indexDropFlags.clientFlags)
			if err != nil {
				return err
			}
			defer client.Close()

			if !indexDropFlags.yes && !confirm(fmt.Sprintf(
				"Are you sure you want to drop the %s index on field %s?",
				indexCreateFlags.namespace,
				indexCreateFlags.vectorField,
			)) {
				return nil
			}

			ctx, cancel := context.WithTimeout(context.Background(), indexDropFlags.clientFlags.Timeout)
			defer cancel()

			err = client.IndexDrop(ctx, indexDropFlags.namespace, indexDropFlags.indexName)
			if err != nil {
				logger.Error("unable to drop index", slog.Any("error", err))
				return err
			}

			view.Printf("Successfully dropped index %s.%s", indexDropFlags.namespace, indexDropFlags.indexName)
			return nil
		},
	}
}

func init() {
	indexDropCmd := newIndexDropCommand()
	indexCmd.AddCommand(indexDropCmd)
	indexDropCmd.Flags().AddFlagSet(newIndexDropFlagSet())

	for _, flag := range indexDropRequiredFlags {
		err := indexDropCmd.MarkFlagRequired(flag)
		if err != nil {
			panic(err)
		}
	}
}
