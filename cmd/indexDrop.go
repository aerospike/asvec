/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"asvec/cmd/flags"
	"context"
	"fmt"
	"log/slog"

	commonFlags "github.com/aerospike/tools-common-go/flags"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

//nolint:govet // Padding not a concern for a CLI
var indexDropFlags = &struct {
	clientFlags flags.ClientFlags
	yes         bool
	namespace   string
	sets        []string
	indexName   string
}{
	clientFlags: *flags.NewClientFlags(),
}

func newIndexDropFlagSet() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}
	flagSet.BoolVarP(&indexDropFlags.yes, flags.Yes, "y", false, commonFlags.DefaultWrapHelpString("When true do not prompt for confirmation."))
	flagSet.StringVarP(&indexDropFlags.namespace, flags.Namespace, "n", "", commonFlags.DefaultWrapHelpString("The namespace for the index.")) //nolint:lll // For readability
	flagSet.StringSliceVarP(&indexDropFlags.sets, flags.Sets, "s", nil, commonFlags.DefaultWrapHelpString("The sets for the index."))          //nolint:lll // For readability
	flagSet.StringVarP(&indexDropFlags.indexName, flags.IndexName, "i", "", commonFlags.DefaultWrapHelpString("The name of the index."))       //nolint:lll // For readability
	flagSet.AddFlagSet(indexDropFlags.clientFlags.NewClientFlagSet())

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
		Long: `A command for dropping indexes. Deleting an index will free up 
		storage but will also disable vector search on your data.

		For example:
			export ASVEC_HOST=<avs-ip>:5000
			asvec index drop -i myindex -n test
			`,
		PreRunE: func(_ *cobra.Command, _ []string) error {
			if viper.IsSet(flags.Seeds) && viper.IsSet(flags.Host) {
				return fmt.Errorf("only --%s or --%s allowed", flags.Seeds, flags.Host)
			}

			return nil
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			logger.Debug("parsed flags",
				append(indexDropFlags.clientFlags.NewSLogAttr(),
					slog.Bool(flags.Yes, indexDropFlags.yes),
					slog.String(flags.Namespace, indexDropFlags.namespace),
					slog.Any(flags.Sets, indexDropFlags.sets),
					slog.String(flags.IndexName, indexDropFlags.indexName),
				)...,
			)

			adminClient, err := createClientFromFlags(&indexDropFlags.clientFlags)
			if err != nil {
				return err
			}
			defer adminClient.Close()

			if !indexDropFlags.yes && !confirm(fmt.Sprintf(
				"Are you sure you want to drop the index %s on field %s?",
				nsAndSetString(
					indexCreateFlags.namespace,
					indexCreateFlags.sets,
				),
				indexCreateFlags.vectorField,
			)) {
				return nil
			}

			ctx, cancel := context.WithTimeout(context.Background(), indexDropFlags.clientFlags.Timeout)
			defer cancel()

			err = adminClient.IndexDrop(ctx, indexDropFlags.namespace, indexDropFlags.indexName)
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
