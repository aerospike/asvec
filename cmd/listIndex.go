/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	avs "github.com/aerospike/aerospike-proximus-client-go"
	"github.com/aerospike/aerospike-proximus-client-go/protos"
	commonFlags "github.com/aerospike/tools-common-go/flags"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type liFlags struct {
	host         *HostPortFlag
	seeds        *SeedsSliceFlag
	listenerName StringOptionalFlag
	verbose      bool
	timeout      time.Duration
}

var listIndexFlags = &liFlags{
	host:  NewDefaultHostPortFlag(),
	seeds: &SeedsSliceFlag{},
}

func newListIndexFlagSet() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}
	flagSet.VarP(listIndexFlags.host, flagNameHost, "h", commonFlags.DefaultWrapHelpString(fmt.Sprintf("The AVS host to connect to. If cluster discovery is needed use --%s", flagNameSeeds)))
	flagSet.Var(listIndexFlags.seeds, flagNameSeeds, commonFlags.DefaultWrapHelpString(fmt.Sprintf("The AVS seeds to use for cluster discovery. If no cluster discovery is needed (i.e. load-balancer) then use --%s", flagNameHost)))
	flagSet.VarP(&listIndexFlags.listenerName, flagNameListenerName, "l", commonFlags.DefaultWrapHelpString("The listener to ask the AVS server for as configured in the AVS server. Likely required for cloud deployments."))
	flagSet.BoolVarP(&listIndexFlags.verbose, flagNameVerbose, "v", false, commonFlags.DefaultWrapHelpString("Print detailed index information."))
	flagSet.DurationVar(&listIndexFlags.timeout, flagNameTimeout, time.Second*5, commonFlags.DefaultWrapHelpString("The distance metric for the index."))

	return flagSet
}

var listIndexRequiredFlags = []string{
	flagNameNamespace,
	flagNameIndexName,
}

// listIndexCmd represents the listIndex command
func newListIndexCmd() *cobra.Command {
	return &cobra.Command{Use: "index",
		Short: "A brief description of your command",
		Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if viper.IsSet(flagNameSeeds) && viper.IsSet(flagNameHost) {
				return fmt.Errorf(fmt.Sprintf("only --%s or --%s allowed", flagNameSeeds, flagNameHost))
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			logger.Debug("parsed flags",
				slog.String(flagNameHost, listIndexFlags.host.String()),
				slog.String(flagNameSeeds, listIndexFlags.seeds.String()),
				slog.String(flagNameListenerName, listIndexFlags.listenerName.String()),
				slog.Bool(flagNameVerbose, listIndexFlags.verbose),
				slog.Duration(flagNameTimeout, listIndexFlags.timeout),
			)

			hosts, isLoadBalancer := parseBothHostSeedsFlag(*listIndexFlags.seeds, *listIndexFlags.host)

			ctx, cancel := context.WithTimeout(context.Background(), listIndexFlags.timeout)
			defer cancel()

			adminClient, err := avs.NewAdminClient(ctx, hosts, listIndexFlags.listenerName.Val, isLoadBalancer, logger)
			if err != nil {
				logger.Error("failed to create AVS client", slog.Any("error", err))
				return err
			}

			cancel()
			defer adminClient.Close()

			ctx, cancel = context.WithTimeout(context.Background(), listIndexFlags.timeout)
			defer cancel()

			indexList, err := adminClient.IndexList(ctx)
			if err != nil {
				logger.Error("failed to list indexes", slog.Any("error", err))
				return err
			}

			indexStatusList := make([]*protos.IndexStatusResponse, len(indexList.GetIndices()))

			if listIndexFlags.verbose {
				cancel()

				ctx, cancel = context.WithTimeout(context.Background(), listIndexFlags.timeout)
				defer cancel()

				wg := sync.WaitGroup{}
				for i, index := range indexList.GetIndices() {
					wg.Add(1)
					go func(i int, index *protos.IndexDefinition) {
						defer wg.Done()
						indexStatus, err := adminClient.IndexGetStatus(ctx, index.Id.Namespace, index.Id.Name)
						if err != nil {
							logger.ErrorContext(ctx, "failed to get index status", slog.Any("error", err), slog.String("index", index.Id.String()))
							return
						}

						indexStatusList[i] = indexStatus
						logger.Debug("server index status", slog.Int("index", i), slog.Any("response", indexStatus))
					}(i, index)
				}

				wg.Wait()
			}

			logger.Debug("server index list", slog.String("response", indexList.String()))

			view.PrintIndexes(indexList, indexStatusList, listIndexFlags.verbose)

			return nil
		},
	}
}

func init() {
	listIndexCmd := newListIndexCmd()

	listCmd.AddCommand(listIndexCmd)
	listIndexCmd.Flags().AddFlagSet(newListIndexFlagSet())

	for _, flag := range listIndexRequiredFlags {
		listIndexCmd.MarkFlagRequired(flag)
	}
}
