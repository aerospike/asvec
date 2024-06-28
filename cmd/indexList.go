/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"asvec/cmd/flags"
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/aerospike/avs-client-go/protos"
	commonFlags "github.com/aerospike/tools-common-go/flags"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var indexListFlags = &struct {
	clientFlags flags.ClientFlags
	verbose     bool
}{
	clientFlags: *flags.NewClientFlags(),
}

func newIndexListFlagSet() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}
	flagSet.BoolVarP(&indexListFlags.verbose, flags.Verbose, "v", false, commonFlags.DefaultWrapHelpString("Print detailed index information.")) //nolint:lll // For readability
	flagSet.AddFlagSet(indexListFlags.clientFlags.NewClientFlagSet())

	return flagSet
}

var indexListRequiredFlags = []string{}

// listIndexCmd represents the listIndex command
func newIndexListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "A command for listing indexes",
		Long: fmt.Sprintf(`A command for displaying useful information about AVS indexes. To display additional
		index information use the --%s flag.
	For example:
		export ASVEC_HOST=<avs-ip>:5000
		asvec index list
		`, flags.Verbose),
		PreRunE: func(_ *cobra.Command, _ []string) error {
			if viper.IsSet(flags.Seeds) && viper.IsSet(flags.Host) {
				return fmt.Errorf("only --%s or --%s allowed", flags.Seeds, flags.Host)
			}

			return nil
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			logger.Debug("parsed flags",
				append(indexListFlags.clientFlags.NewSLogAttr(),
					slog.Bool(flags.Verbose, indexListFlags.verbose),
				)...,
			)

			adminClient, err := createClientFromFlags(&indexListFlags.clientFlags)
			if err != nil {
				return err
			}
			defer adminClient.Close()

			ctx, cancel := context.WithTimeout(context.Background(), indexListFlags.clientFlags.Timeout)
			defer cancel()

			indexList, err := adminClient.IndexList(ctx)
			if err != nil {
				logger.Error("failed to list indexes", slog.Any("error", err))
				return err
			}

			indexStatusList := make([]*protos.IndexStatusResponse, len(indexList.GetIndices()))

			cancel()

			ctx, cancel = context.WithTimeout(context.Background(), indexListFlags.clientFlags.Timeout)
			defer cancel()

			wg := sync.WaitGroup{}
			for i, index := range indexList.GetIndices() {
				wg.Add(1)
				go func(i int, index *protos.IndexDefinition) {
					defer wg.Done()
					indexStatus, err := adminClient.IndexGetStatus(ctx, index.Id.Namespace, index.Id.Name)
					if err != nil {
						logger.ErrorContext(ctx,
							"failed to get index status",
							slog.Any("error", err),
							slog.String("index", index.Id.String()),
						)
						return
					}

					indexStatusList[i] = indexStatus
					logger.Debug("server index status", slog.Int("index", i), slog.Any("response", indexStatus))
				}(i, index)
			}

			wg.Wait()

			logger.Debug("server index list", slog.String("response", indexList.String()))

			view.PrintIndexes(indexList, indexStatusList, indexListFlags.verbose)

			return nil
		},
	}
}

func init() {
	indexListCmd := newIndexListCmd()

	indexCmd.AddCommand(indexListCmd)
	indexListCmd.Flags().AddFlagSet(newIndexListFlagSet())

	for _, flag := range indexListRequiredFlags {
		err := indexListCmd.MarkFlagRequired(flag)
		if err != nil {
			panic(err)
		}
	}
}
