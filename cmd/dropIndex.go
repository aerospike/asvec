/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"asvec/cmd/flags"
	"context"
	"fmt"
	"log/slog"
	"time"

	avs "github.com/aerospike/aerospike-proximus-client-go"
	commonFlags "github.com/aerospike/tools-common-go/flags"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

//nolint:govet // Padding not a concern for a CLI
var dropIndexFlags = &struct {
	host         *flags.HostPortFlag
	seeds        *flags.SeedsSliceFlag
	listenerName flags.StringOptionalFlag
	namespace    string
	sets         []string
	indexName    string
	timeout      time.Duration
	tls          *flags.TLSFlags
}{
	host:  flags.NewDefaultHostPortFlag(),
	seeds: &flags.SeedsSliceFlag{},
	tls:   &flags.TLSFlags{},
}

func newDropIndexFlagSet() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}
	flagSet.VarP(dropIndexFlags.host, flagNameHost, "h", commonFlags.DefaultWrapHelpString(fmt.Sprintf("The AVS host to connect to. If cluster discovery is needed use --%s", flagNameSeeds)))                                         //nolint:lll // For readability
	flagSet.Var(dropIndexFlags.seeds, flagNameSeeds, commonFlags.DefaultWrapHelpString(fmt.Sprintf("The AVS seeds to use for cluster discovery. If no cluster discovery is needed (i.e. load-balancer) then use --%s", flagNameHost))) //nolint:lll // For readability
	flagSet.VarP(&dropIndexFlags.listenerName, flagNameListenerName, "l", commonFlags.DefaultWrapHelpString("The listener to ask the AVS server for as configured in the AVS server. Likely required for cloud deployments."))         //nolint:lll // For readability
	flagSet.StringVarP(&dropIndexFlags.namespace, flagNameNamespace, "n", "", commonFlags.DefaultWrapHelpString("The namespace for the index."))                                                                                       //nolint:lll // For readability
	flagSet.StringArrayVarP(&dropIndexFlags.sets, flagNameSets, "s", nil, commonFlags.DefaultWrapHelpString("The sets for the index."))                                                                                                //nolint:lll // For readability
	flagSet.StringVarP(&dropIndexFlags.indexName, flagNameIndexName, "i", "", commonFlags.DefaultWrapHelpString("The name of the index."))                                                                                             //nolint:lll // For readability
	flagSet.DurationVar(&dropIndexFlags.timeout, flagNameTimeout, time.Second*5, commonFlags.DefaultWrapHelpString("The distance metric for the index."))                                                                              //nolint:lll // For readability
	flagSet.AddFlagSet(dropIndexFlags.tls.NewFlagSet(commonFlags.DefaultWrapHelpString))

	return flagSet
}

var dropIndexRequiredFlags = []string{
	flagNameNamespace,
	flagNameIndexName,
}

// dropIndexCmd represents the dropIndex command
func newDropIndexCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "index",
		Short: "A command for dropping indexes",
		Long: `A command for dropping indexes. Deleting an index will free up 
		storage but will also disable vector search on your data.

		For example:
			export ASVEC_HOST=<avs-ip>:5000
			asvec drop index -i myindex -n test
			`,
		PreRunE: func(_ *cobra.Command, _ []string) error {
			if viper.IsSet(flagNameSeeds) && viper.IsSet(flagNameHost) {
				return fmt.Errorf("only --%s or --%s allowed", flagNameSeeds, flagNameHost)
			}

			return nil
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			logger.Debug("parsed flags",
				slog.String(flagNameHost, dropIndexFlags.host.String()),
				slog.String(flagNameSeeds, dropIndexFlags.seeds.String()),
				slog.String(flagNameListenerName, dropIndexFlags.listenerName.String()),
				slog.String(flagNameNamespace, dropIndexFlags.namespace),
				slog.Any(flagNameSets, dropIndexFlags.sets),
				slog.String(flagNameIndexName, dropIndexFlags.indexName),
				slog.Duration(flagNameTimeout, dropIndexFlags.timeout),
			)

			hosts, isLoadBalancer := parseBothHostSeedsFlag(dropIndexFlags.seeds, dropIndexFlags.host)

			ctx, cancel := context.WithTimeout(context.Background(), dropIndexFlags.timeout)
			defer cancel()

			adminClient, err := avs.NewAdminClient(
				ctx, hosts, nil, isLoadBalancer, nil, logger,
			)
			if err != nil {
				logger.Error("failed to create AVS client", slog.Any("error", err))
				return err
			}

			cancel()
			defer adminClient.Close()

			ctx, cancel = context.WithTimeout(context.Background(), dropIndexFlags.timeout)
			defer cancel()

			err = adminClient.IndexDrop(ctx, dropIndexFlags.namespace, dropIndexFlags.indexName)
			if err != nil {
				logger.Error("unable to drop index", slog.Any("error", err))
				return err
			}

			view.Printf("Successfully dropped index %s.%s", dropIndexFlags.namespace, dropIndexFlags.indexName)
			return nil
		},
	}
}

func init() {
	dropIndexCmd := newDropIndexCommand()
	dropCmd.AddCommand(dropIndexCmd)
	dropIndexCmd.Flags().AddFlagSet(newDropIndexFlagSet())

	for _, flag := range dropIndexRequiredFlags {
		err := dropIndexCmd.MarkFlagRequired(flag)
		if err != nil {
			panic(err)
		}
	}
}
