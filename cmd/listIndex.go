/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"log/slog"

	avs "github.com/aerospike/aerospike-proximus-client-go"
	"github.com/aerospike/aerospike-proximus-client-go/protos"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// type indexInfo struct {
// 	Definition *protos.IndexDefinition
// 	Status     *protos.IndexStatusResponse
// }

// listIndexCmd represents the listIndex command
var listIndexCmd = &cobra.Command{
	Use:   "index",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		seed := viper.GetString(flagNameSeeds)
		// port := viper.GetInt(flagNamePort)
		hostPort := avs.NewHostPort(seed, 5002, false)
		timeout := viper.GetDuration(flagNameTimeout)
		verbose := viper.GetBool(flagNameVerbose)

		// logger.Debug("parsed flags",
		// 	slog.String("seeds", seed), slog.Int("port", port), slog.Duration("timeout", timeout),
		// )

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

		indexList, err := adminClient.IndexList(ctx)
		if err != nil {
			logger.Error("failed to list indexes", slog.Any("error", err))
			return err
		}

		indexStatusList := make([]*protos.IndexStatusResponse, len(indexList.GetIndices()))

		if verbose {
			cancel()

			ctx, cancel = context.WithTimeout(context.Background(), timeout)
			defer cancel()

			for i, index := range indexList.GetIndices() {
				indexStatus, err := adminClient.IndexGetStatus(ctx, index.Id.Namespace, index.Id.Name)
				if err != nil {
					logger.ErrorContext(ctx, "failed to get index status", slog.Any("error", err), slog.String("index", index.Id.String()))
					continue
				}

				indexStatusList[i] = indexStatus
			}

		}

		logger.Debug(indexList.String())
		view.PrintIndexes(indexList, indexStatusList, verbose)

		return nil
	},
}

func init() {
	listCmd.AddCommand(listIndexCmd)

	flags := NewFlagSetBuilder(listIndexCmd.Flags())
	flags.AddSeedFlag()
	// flags.AddPortFlag()
	flags.AddTimeoutFlag()
	flags.AddVerbose()
}
