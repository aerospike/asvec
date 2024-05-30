/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	avs "github.com/aerospike/aerospike-proximus-client-go"
	"github.com/aerospike/aerospike-proximus-client-go/protos"
	commonFlags "github.com/aerospike/tools-common-go/flags"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	flagNameStorageNamespace = "storage-namespace"
	flagNameStorageSet       = "storage-set"
	flagNameMaxEdges         = "hnsw-max-edges"
	flagNameConstructionEf   = "hnsw-ef-construction"
	flagNameEf               = "hnsw-ef"
	flagNameBatchMaxRecords  = "hnsw-batch-max-records"
	flagNameBatchInterval    = "hnsw-batch-interval"
	flagNameBatchEnabled     = "hnsw-batch-disabled"
)

func parseHostPort(rawHost string) (*avs.HostPort, error) {
	split := strings.SplitN(rawHost, ":", 2)
	host := split[0]
	port := 5000

	if len(split) > 1 {
		var err error
		port, err = strconv.Atoi(split[1])

		if err != nil {
			return nil, fmt.Errorf("unparsable port: %w", err)
		}
	}
	return avs.NewHostPort(host, port, false), nil
}

// createIndexCmd represents the createIndex command
var createIndexCmd = &cobra.Command{
	Use:   "index",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if viper.IsSet(flagNameSeeds) && viper.IsSet(flagNameHost) {
			return fmt.Errorf(fmt.Sprintf("only %s or %s allowed", flagNameSeeds, flagNameHost))
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: likely add to prerun step
		isLoadBalancer := false
		hosts := avs.HostPortSlice{}

		if viper.IsSet(flagNameHost) {
			isLoadBalancer = true
			rawHost := viper.GetString(flagNameHost)
			hostPort, err := parseHostPort(rawHost)

			if err != nil {
				return err
			}

			hosts = append(hosts, hostPort)
		} else if viper.IsSet(flagNameSeeds) {
			rawSeed := viper.GetStringSlice(flagNameSeeds)

			for _, rawHost := range rawSeed {
				hostPort, err := parseHostPort(rawHost)

				if err != nil {
					return err
				}

				hosts = append(hosts, hostPort)
			}
		}

		namespace := viper.GetString(flagNameNamespace)
		sets := viper.GetStringSlice(flagNameSets)
		indexName := viper.GetString(flagNameIndexName)
		vectorField := viper.GetString(flagNameVectorField)
		dimension := viper.GetUint32(flagNameDimension)
		indexMeta := viper.GetStringMapString(flagNameIndexMeta)
		distanceMetric := viper.GetString(flagNameDistanceMetric)
		timeout := viper.GetDuration(flagNameTimeout)

		storageNamespace := viperGetIfSetString(flagNameStorageNamespace)
		storageSet := viperGetIfSetString(flagNameSets)
		hnswM := viperGetIfSetUint32(flagNameMaxEdges)
		hnswEf := viperGetIfSetUint32(flagNameEf)
		hnswConEf := viperGetIfSetUint32(flagNameConstructionEf)
		hnswBatchMaxConns := viperGetIfSetUint32(flagNameBatchMaxRecords)
		hnswBatchInterval := viperGetIfSetUint32(flagNameBatchInterval)
		hnswBatchEnabled := viperGetIfSetBool(flagNameBatchEnabled)

		// TODO, fix seeds
		logger.Debug("parsed flags",
			slog.String("namespace", namespace),
			slog.Any("sets", sets), slog.String("index-name", indexName), slog.String("vector-field", vectorField),
			slog.Uint64("dimension", uint64(dimension)), slog.Any("index-meta", indexMeta), slog.String("distance-metric", distanceMetric),
			slog.Duration("timeout", timeout), slog.Any("storage-namespace", storageNamespace), slog.Any("storage-set", storageSet),
			slog.Any("hnsw-max-edges", hnswM), slog.Any("hnsw-ef", hnswEf),
			slog.Any("hnsw-ef-construction", hnswConEf), slog.Any("hnsw-batch-max-records", hnswBatchMaxConns),
			slog.Any("hnsw-batch-interval", hnswBatchInterval), slog.Any("hnsw-batch-enabled", hnswBatchEnabled),
		)

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		// TODO listener name
		adminClient, err := avs.NewAdminClient(ctx, hosts, nil, isLoadBalancer, logger)
		if err != nil {
			logger.Error("failed to create AVS client", slog.Any("error", err))
			return err
		}

		cancel()
		defer adminClient.Close()

		ctx, cancel = context.WithTimeout(context.Background(), timeout)
		defer cancel()

		// Inverted to make it easier to understand
		var hnswBatchDisabled *bool
		if hnswBatchEnabled != nil {
			*hnswBatchDisabled = !*hnswBatchEnabled
		}

		indexStorage := &protos.IndexStorage{
			Namespace: storageNamespace,
			Set:       storageSet,
		}
		hnswParams := &protos.HnswParams{
			M:              hnswM,
			Ef:             hnswEf,
			EfConstruction: hnswConEf,
			BatchingParams: &protos.HnswBatchingParams{
				MaxRecords: hnswBatchMaxConns,
				Interval:   hnswBatchInterval,
				Disabled:   hnswBatchDisabled,
			},
		}

		err = adminClient.IndexCreate(
			ctx, namespace, sets, indexName, vectorField, dimension,
			protos.VectorDistanceMetric(protos.VectorDistanceMetric_value[distanceMetric]),
			hnswParams, indexMeta, indexStorage)
		if err != nil {
			logger.Error("unable to create index", slog.Any("error", err))
			return err
		}

		view.Printf("Successfully created index %s.%s", namespace, indexName)
		return nil
	},
}

func init() {
	createCmd.AddCommand(createIndexCmd)

	persistentFlags := NewFlagSetBuilder(createIndexCmd.PersistentFlags())
	flags := NewFlagSetBuilder(createIndexCmd.Flags())

	persistentFlags.AddSeedFlag()
	persistentFlags.AddHostFlag()

	flags.AddNamespaceFlag()
	flags.AddSetsFlag()
	flags.AddIndexNameFlag()
	flags.AddVectorFieldFlag()
	flags.AddDimensionFlag()
	flags.AddDistanceMetricFlag()
	flags.AddIndexMetaFlag()
	flags.AddTimeoutFlag()

	var requiredFlags = []string{
		flagNameNamespace,
		flagNameIndexName,
		flagNameVectorField,
		flagNameDimension,
		flagNameDistanceMetric,
	}

	// TODO: Add custom template for usage to take into account terminal width
	// Ex: https://github.com/sigstore/cosign/pull/3011/files
	flags.String(flagNameStorageNamespace, "", commonFlags.DefaultWrapHelpString("Optional storage namespace where the index is stored. Defaults to the index namespace."))
	flags.String(flagNameStorageSet, "", commonFlags.DefaultWrapHelpString("Optional storage set where the index is stored. Defaults to the index name."))
	flags.Uint32(flagNameMaxEdges, 0, commonFlags.DefaultWrapHelpString("Maximum number bi-directional links per HNSW vertex. Greater values of 'm' in general provide better recall for data with high dimensionality, while lower values work well for data with lower dimensionality. The storage space required for the index increases proportionally with 'm'. The default value is 16."))
	flags.Uint32(flagNameConstructionEf, 0, commonFlags.DefaultWrapHelpString("The number of candidate nearest neighbors shortlisted during index creation. Larger values provide better recall at the cost of longer index update times. The default is 100."))
	flags.Uint32(flagNameEf, 0, commonFlags.DefaultWrapHelpString("The default number of candidate nearest neighbors shortlisted during search. Larger values provide better recall at the cost of longer search times. The default is 100."))
	flags.Uint32(flagNameBatchMaxRecords, 0, commonFlags.DefaultWrapHelpString("Maximum number of records to fit in a batch. The default value is 10000."))
	flags.Uint32(flagNameBatchInterval, 0, commonFlags.DefaultWrapHelpString("The maximum amount of time in milliseconds to wait before finalizing a batch. The default value is 10000."))
	flags.Bool(flagNameBatchEnabled, true, commonFlags.DefaultWrapHelpString("Enables batching for index updates. Default is true meaning batching is enabled."))

	for _, flag := range requiredFlags {
		createIndexCmd.MarkFlagRequired(flag)
	}

	createIndexCmd.MarkFlagsMutuallyExclusive(flagNameHost, flagNameSeeds)
}
