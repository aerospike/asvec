/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"log/slog"

	avs "github.com/aerospike/aerospike-proximus-client-go"
	"github.com/aerospike/aerospike-proximus-client-go/protos"
	commonFlags "github.com/aerospike/tools-common-go/flags"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var requiredFlags = []string{
	flagNameNamespace,
	flagNameIndexName,
	flagNameDimension,
	flagNameDistance,
}

var persistentRequiredFlags = []string{}

const (
	flagNameMaxEdges        = "hnsw-max-edges"
	flagNameConstructionEf  = "hnsw-ef-construction"
	flagNameEf              = "hnsw-ef"
	flagNameBatchMaxRecords = "hnsw-batch-max-records"
	flagNameBatchInterval   = "hnsw-batch-interval"
	flagNameBatchDisabled   = "hnsw-batch-disabled"
)

// createIndexCmd represents the createIndex command
var createIndexCmd = &cobra.Command{
	Use:   "index",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		seed := viper.GetString(flagNameSeeds)
		port := viper.GetInt(flagNamePort)
		hostPort := avs.NewHostPort(seed, port, false)
		namespace := viper.GetString(flagNameNamespace)
		sets := viper.GetStringSlice(flagNameSets)
		indexName := viper.GetString(flagNameIndexName)
		vectorField := viper.GetString(flagNameVector)
		dimension := viper.GetUint32(flagNameDimension)
		indexMeta := viper.GetStringMapString(flagNameIndexMeta)
		distanceMetric := viper.GetString(flagNameDistance)

		logger.Debug("parsed flags", slog.String("seeds", seed), slog.Int("port", port), slog.String("namespace", namespace), slog.Any("sets", sets), slog.String("index-name", indexName), slog.String("vector-field", vectorField), slog.Uint64("dimension", uint64(dimension)), slog.Any("index-meta", indexMeta))

		ctx := context.TODO()

		adminClient, err := avs.NewAdminClient(ctx, []*avs.HostPort{hostPort}, nil, false, logger)
		if err != nil {
			logger.Error("failed to create AVS client", slog.Any("error", err))
			view.Printf("Failed to connect to AVS: %v", err)
			return
		}

		// TODO: parse cosine
		err = adminClient.IndexCreate(ctx, namespace, sets, indexName, vectorField, dimension, protos.VectorDistanceMetric(protos.VectorDistanceMetric_value[distanceMetric]), nil, indexMeta)
		if err != nil {
			logger.Error("unable to create index", slog.Any("error", err))
			view.Printf("Unable to create index: %v", err)
			return
		}

		view.Printf("Successfully created index %s.%s", namespace, indexName)
	},
}

func init() {
	createCmd.AddCommand(createIndexCmd)

	persistentFlags := NewFlagSetBuilder(createIndexCmd.PersistentFlags())
	flags := NewFlagSetBuilder(createIndexCmd.Flags())

	persistentFlags.AddSeedFlag()
	persistentFlags.AddPortFlag()

	flags.AddNamespaceFlag()
	flags.AddSetsFlag()
	flags.AddIndexNameFlag()
	flags.AddVectorFieldFlag()
	flags.AddDimensionFlag()
	flags.AddDistanceMetricFlag()
	flags.AddIndexMetaFlag()

	flags.Uint32(flagNameMaxEdges, 0, commonFlags.DefaultWrapHelpString("Maximum number bi-directional links per HNSW vertex. Greater values of 'm' in general provide better recall for data with high dimensionality, while lower values work well for data with lower dimensionality. The storage space required for the index increases proportionally with 'm'. The default value is 16."))
	flags.Uint32(flagNameConstructionEf, 0, commonFlags.DefaultWrapHelpString("The number of candidate nearest neighbors shortlisted during index creation. Larger values provide better recall at the cost of longer index update times. The default is 100."))
	flags.Uint32(flagNameEf, 0, commonFlags.DefaultWrapHelpString("The default number of candidate nearest neighbors shortlisted during search. Larger values provide better recall at the cost of longer search times. The default is 100."))
	flags.Uint32(flagNameBatchMaxRecords, 0, commonFlags.DefaultWrapHelpString("Maximum number of records to fit in a batch. The default value is 10000."))
	flags.Uint32(flagNameBatchInterval, 0, commonFlags.DefaultWrapHelpString("The maximum amount of time in milliseconds to wait before finalizing a batch. The default value is 10000."))
	flags.Bool(flagNameBatchDisabled, false, commonFlags.DefaultWrapHelpString("Disables batching for index updates. Default is false meaning batching is enabled."))

	for _, flag := range requiredFlags {
		createIndexCmd.MarkFlagRequired(flag)
	}

	for _, flag := range persistentRequiredFlags {
		createIndexCmd.MarkPersistentFlagRequired(flag)
	}

	// TODO hnsw metadata
	viper.BindPFlags(createIndexCmd.PersistentFlags())
	viper.BindPFlags(createIndexCmd.Flags())

}
