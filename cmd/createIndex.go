/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	avs "github.com/aerospike/aerospike-proximus-client-go"
	"github.com/aerospike/aerospike-proximus-client-go/protos"
	commonFlags "github.com/aerospike/tools-common-go/flags"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const ()

type ciFlags struct {
	host                *HostPortFlag
	seeds               *SeedsSliceFlag
	listenerName        StringOptionalFlag
	namespace           string
	sets                []string
	indexName           string
	vectorField         string
	dimensions          uint32
	distanceMetric      DistanceMetricFlag
	indexMeta           map[string]string
	storageNamespace    StringOptionalFlag
	storageSet          StringOptionalFlag
	hnswMaxEdges        Uint32OptionalFlag
	hnswEf              Uint32OptionalFlag
	hnswConstructionEf  Uint32OptionalFlag
	hnswBatchMaxRecords Uint32OptionalFlag
	hnswBatchInterval   Uint32OptionalFlag
	hnswBatchEnabled    BoolOptionalFlag
	timeout             time.Duration
}

var createIndexFlags = &ciFlags{
	host:                NewDefaultHostPortFlag(),
	seeds:               &SeedsSliceFlag{},
	storageNamespace:    StringOptionalFlag{},
	storageSet:          StringOptionalFlag{},
	hnswMaxEdges:        Uint32OptionalFlag{},
	hnswEf:              Uint32OptionalFlag{},
	hnswConstructionEf:  Uint32OptionalFlag{},
	hnswBatchMaxRecords: Uint32OptionalFlag{},
	hnswBatchInterval:   Uint32OptionalFlag{},
	hnswBatchEnabled:    BoolOptionalFlag{},
}

func newCreateIndexFlagSet() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}
	flagSet.VarP(createIndexFlags.host, flagNameHost, "h", commonFlags.DefaultWrapHelpString(fmt.Sprintf("The AVS host to connect to. If cluster discovery is needed use --%s", flagNameSeeds)))
	flagSet.Var(createIndexFlags.seeds, flagNameSeeds, commonFlags.DefaultWrapHelpString(fmt.Sprintf("The AVS seeds to use for cluster discovery. If no cluster discovery is needed (i.e. load-balancer) then use --%s", flagNameHost)))
	flagSet.VarP(&createIndexFlags.listenerName, flagNameListenerName, "l", commonFlags.DefaultWrapHelpString("The listener to ask the AVS server for as configured in the AVS server. Likely required for cloud deployments."))
	flagSet.StringVarP(&createIndexFlags.namespace, flagNameNamespace, "n", "", commonFlags.DefaultWrapHelpString("The namespace for the index."))
	flagSet.StringArrayVarP(&createIndexFlags.sets, flagNameSets, "s", nil, commonFlags.DefaultWrapHelpString("The sets for the index."))
	flagSet.StringVarP(&createIndexFlags.indexName, flagNameIndexName, "i", "", commonFlags.DefaultWrapHelpString("The name of the index."))
	flagSet.StringVarP(&createIndexFlags.vectorField, flagNameVectorField, "f", "", commonFlags.DefaultWrapHelpString("The name of the vector field."))
	flagSet.Uint32VarP(&createIndexFlags.dimensions, flagNameDimension, "d", 0, commonFlags.DefaultWrapHelpString("The dimension of the vector field."))
	flagSet.VarP(&createIndexFlags.distanceMetric, flagNameDistanceMetric, "m", commonFlags.DefaultWrapHelpString("The distance metric for the index."))
	flagSet.StringToStringVar(&createIndexFlags.indexMeta, flagNameIndexMeta, nil, commonFlags.DefaultWrapHelpString("The distance metric for the index."))
	flagSet.DurationVar(&createIndexFlags.timeout, flagNameTimeout, time.Second*5, commonFlags.DefaultWrapHelpString("The distance metric for the index."))
	flagSet.Var(&createIndexFlags.storageNamespace, flagNameStorageNamespace, commonFlags.DefaultWrapHelpString("Optional storage namespace where the index is stored. Defaults to the index namespace."))
	flagSet.Var(&createIndexFlags.storageSet, flagNameStorageSet, commonFlags.DefaultWrapHelpString("Optional storage set where the index is stored. Defaults to the index name."))
	flagSet.Var(&createIndexFlags.hnswMaxEdges, flagNameMaxEdges, commonFlags.DefaultWrapHelpString("Maximum number bi-directional links per HNSW vertex. Greater values of 'm' in general provide better recall for data with high dimensionality, while lower values work well for data with lower dimensionality. The storage space required for the index increases proportionally with 'm'. The default value is 16."))
	flagSet.Var(&createIndexFlags.hnswConstructionEf, flagNameConstructionEf, commonFlags.DefaultWrapHelpString("The number of candidate nearest neighbors shortlisted during index creation. Larger values provide better recall at the cost of longer index update times. The default is 100."))
	flagSet.Var(&createIndexFlags.hnswEf, flagNameEf, commonFlags.DefaultWrapHelpString("The default number of candidate nearest neighbors shortlisted during search. Larger values provide better recall at the cost of longer search times. The default is 100."))
	flagSet.Var(&createIndexFlags.hnswBatchMaxRecords, flagNameBatchMaxRecords, commonFlags.DefaultWrapHelpString("Maximum number of records to fit in a batch. The default value is 10000."))
	flagSet.Var(&createIndexFlags.hnswBatchInterval, flagNameBatchInterval, commonFlags.DefaultWrapHelpString("The maximum amount of time in milliseconds to wait before finalizing a batch. The default value is 10000."))
	flagSet.Var(&createIndexFlags.hnswBatchEnabled, flagNameBatchEnabled, commonFlags.DefaultWrapHelpString("Enables batching for index updates. Default is true meaning batching is enabled."))

	return flagSet
}

var createIndexRequiredFlags = []string{
	flagNameNamespace,
	flagNameIndexName,
	flagNameVectorField,
	flagNameDimension,
	flagNameDistanceMetric,
}

// createIndexCmd represents the createIndex command
func newCreateIndexCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "index",
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
			hosts, isLoadBalancer := parseBothHostSeedsFlag(*createIndexFlags.seeds, *createIndexFlags.host)

			logger.Debug("parsed flags",
				slog.String(flagNameHost, createIndexFlags.host.String()),
				slog.String(flagNameSeeds, createIndexFlags.seeds.String()),
				slog.String(flagNameListenerName, createIndexFlags.listenerName.String()),
				slog.String(flagNameNamespace, createIndexFlags.namespace),
				slog.Any(flagNameSets, createIndexFlags.sets),
				slog.String(flagNameIndexName, createIndexFlags.indexName),
				slog.String(flagNameVectorField, createIndexFlags.vectorField),
				slog.Uint64(flagNameDimension, uint64(createIndexFlags.dimensions)),
				slog.Any(flagNameIndexMeta, createIndexFlags.indexMeta),
				slog.String(flagNameDistanceMetric, createIndexFlags.distanceMetric.String()),
				slog.Duration(flagNameTimeout, createIndexFlags.timeout),
				slog.Any(flagNameStorageNamespace, createIndexFlags.storageNamespace.String()),
				slog.Any(flagNameStorageSet, createIndexFlags.storageSet.String()),
				slog.Any(flagNameMaxEdges, createIndexFlags.hnswMaxEdges.String()),
				slog.Any(flagNameEf, createIndexFlags.hnswEf),
				slog.Any(flagNameConstructionEf, createIndexFlags.hnswConstructionEf.String()),
				slog.Any(flagNameBatchMaxRecords, createIndexFlags.hnswBatchMaxRecords.String()),
				slog.Any(flagNameBatchInterval, createIndexFlags.hnswBatchInterval.String()),
				slog.Any(flagNameBatchEnabled, createIndexFlags.hnswBatchEnabled.String()),
			)

			ctx, cancel := context.WithTimeout(context.Background(), createIndexFlags.timeout)
			defer cancel()

			adminClient, err := avs.NewAdminClient(
				ctx, hosts, createIndexFlags.listenerName.Val, isLoadBalancer, logger,
			)
			if err != nil {
				logger.Error("failed to create AVS client", slog.Any("error", err))
				return err
			}

			cancel()
			defer adminClient.Close()

			ctx, cancel = context.WithTimeout(context.Background(), createIndexFlags.timeout)
			defer cancel()

			// Inverted to make it easier to understand
			var hnswBatchDisabled *bool
			if createIndexFlags.hnswBatchEnabled.Val != nil {
				bd := !(*createIndexFlags.hnswBatchEnabled.Val)
				hnswBatchDisabled = &bd
			}

			indexStorage := &protos.IndexStorage{
				Namespace: createIndexFlags.storageNamespace.Val,
				Set:       createIndexFlags.storageSet.Val,
			}

			hnswParams := &protos.HnswParams{
				M:              createIndexFlags.hnswMaxEdges.Val,
				Ef:             createIndexFlags.hnswEf.Val,
				EfConstruction: createIndexFlags.hnswConstructionEf.Val,
				BatchingParams: &protos.HnswBatchingParams{
					MaxRecords: createIndexFlags.hnswBatchMaxRecords.Val,
					Interval:   createIndexFlags.hnswBatchInterval.Val,
					Disabled:   hnswBatchDisabled,
				},
			}

			err = adminClient.IndexCreate(
				ctx,
				createIndexFlags.namespace,
				createIndexFlags.sets,
				createIndexFlags.indexName,
				createIndexFlags.vectorField,
				createIndexFlags.dimensions,
				protos.VectorDistanceMetric(protos.VectorDistanceMetric_value[createIndexFlags.distanceMetric.String()]),
				hnswParams,
				createIndexFlags.indexMeta,
				indexStorage,
			)
			if err != nil {
				logger.Error("unable to create index", slog.Any("error", err))
				return err
			}

			view.Printf("Successfully created index %s.%s", createIndexFlags.namespace, createIndexFlags.indexName)
			return nil
		},
	}
}

func init() {
	createIndexCmd := newCreateIndexCmd()
	createCmd.AddCommand(createIndexCmd)

	// TODO: Add custom template for usage to take into account terminal width
	// Ex: https://github.com/sigstore/cosign/pull/3011/files

	flagSet := newCreateIndexFlagSet()
	createIndexCmd.Flags().AddFlagSet(flagSet)

	for _, flag := range createIndexRequiredFlags {
		createIndexCmd.MarkFlagRequired(flag)
	}
}
