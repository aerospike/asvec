/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"log/slog"

	avs "github.com/aerospike/aerospike-proximus-client-go"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var dropIndexRequiredFlags = []string{
	flagNameNamespace,
	flagNameIndexName,
	flagNameDimension,
	flagNameDistance,
}

// dropIndexCmd represents the dropIndex command
var dropIndexCmd = &cobra.Command{
	Use:   "index",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: likely add to prerun step
		seed := viper.GetString(flagNameSeeds)
		port := viper.GetInt(flagNamePort)
		hostPort := avs.NewHostPort(seed, port, false)
		namespace := viper.GetString(flagNameNamespace)
		indexName := viper.GetString(flagNameIndexName)
		timeout := viper.GetDuration(flagNameTimeout)

		logger.Debug("parsed flags",
			slog.String("seeds", seed), slog.Int("port", port), slog.String("namespace", namespace),
			slog.String("index-name", indexName), slog.Duration("timeout", timeout),
		)

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		adminClient, err := avs.NewAdminClient(ctx, avs.HostPortSlice{hostPort}, nil, false, logger)
		if err != nil {
			logger.Error("failed to create AVS client", slog.Any("error", err))
			return err
		}

		cancel()
		defer adminClient.Close()

		ctx, cancel = context.WithTimeout(context.Background(), timeout)
		defer cancel()

		err = adminClient.IndexDrop(ctx, namespace, indexName)
		if err != nil {
			logger.Error("unable to drop index", slog.Any("error", err))
			return err
		}

		view.Printf("Successfully dropped index %s.%s", namespace, indexName)
		return nil
	},
}

func init() {
	dropCmd.AddCommand(dropIndexCmd)

	flags := NewFlagSetBuilder(dropIndexCmd.Flags())
	flags.AddSeedFlag()
	flags.AddPortFlag()
	flags.AddNamespaceFlag()
	flags.AddIndexNameFlag()
	flags.AddTimeoutFlag()

	var requiredFlags = []string{
		flagNameNamespace,
		flagNameIndexName,
	}

	for _, flag := range requiredFlags {
		dropIndexCmd.MarkFlagRequired(flag)
	}
}
