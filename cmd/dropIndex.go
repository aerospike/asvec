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

	avs "github.com/aerospike/avs-client-go"
	commonFlags "github.com/aerospike/tools-common-go/flags"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

//nolint:govet // Padding not a concern for a CLI
var dropIndexFlags = &struct {
	flags.ClientFlags
	namespace string
	sets      []string
	indexName string
	timeout   time.Duration
}{
	ClientFlags: *flags.NewClientFlags(),
}

func newDropIndexFlagSet() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}
	flagSet.StringVarP(&dropIndexFlags.namespace, flags.Namespace, "n", "", commonFlags.DefaultWrapHelpString("The namespace for the index."))          //nolint:lll // For readability
	flagSet.StringArrayVarP(&dropIndexFlags.sets, flags.Sets, "s", nil, commonFlags.DefaultWrapHelpString("The sets for the index."))                   //nolint:lll // For readability
	flagSet.StringVarP(&dropIndexFlags.indexName, flags.IndexName, "i", "", commonFlags.DefaultWrapHelpString("The name of the index."))                //nolint:lll // For readability
	flagSet.DurationVar(&dropIndexFlags.timeout, flags.Timeout, time.Second*5, commonFlags.DefaultWrapHelpString("The distance metric for the index.")) //nolint:lll // For readability
	flagSet.AddFlagSet(dropIndexFlags.NewClientFlagSet())

	return flagSet
}

var dropIndexRequiredFlags = []string{
	flags.Namespace,
	flags.IndexName,
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
			if viper.IsSet(flags.Seeds) && viper.IsSet(flags.Host) {
				return fmt.Errorf("only --%s or --%s allowed", flags.Seeds, flags.Host)
			}

			return nil
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			logger.Debug("parsed flags",
				slog.String(flags.Host, dropIndexFlags.Host.String()),
				slog.String(flags.Seeds, dropIndexFlags.Seeds.String()),
				slog.String(flags.ListenerName, dropIndexFlags.ListenerName.String()),
				slog.Bool(flags.TLSCaFile, createIndexFlags.TLSRootCAFile != nil),
				slog.Bool(flags.TLSCaPath, createIndexFlags.TLSRootCAPath != nil),
				slog.Bool(flags.TLSCertFile, createIndexFlags.TLSCertFile != nil),
				slog.Bool(flags.TLSKeyFile, createIndexFlags.TLSKeyFile != nil),
				slog.Bool(flags.TLSKeyFilePass, createIndexFlags.TLSKeyFilePass != nil),
				slog.String(flags.Namespace, dropIndexFlags.namespace),
				slog.Any(flags.Sets, dropIndexFlags.sets),
				slog.String(flags.IndexName, dropIndexFlags.indexName),
				slog.Duration(flags.Timeout, dropIndexFlags.timeout),
			)

			hosts, isLoadBalancer := parseBothHostSeedsFlag(dropIndexFlags.Seeds, dropIndexFlags.Host)

			ctx, cancel := context.WithTimeout(context.Background(), dropIndexFlags.timeout)
			defer cancel()

			tlsConfig, err := dropIndexFlags.NewTLSConfig()
			if err != nil {
				logger.Error("failed to create TLS config", slog.Any("error", err))
				return err
			}

			adminClient, err := avs.NewAdminClient(
				ctx, hosts, dropIndexFlags.ListenerName.Val, isLoadBalancer, tlsConfig, logger,
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
