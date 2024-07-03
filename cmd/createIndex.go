/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"asvec/cmd/flags"
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/aerospike/avs-client-go/protos"
	commonFlags "github.com/aerospike/tools-common-go/flags"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

//nolint:govet // Padding not a concern for a CLI
var createIndexFlags = &struct {
	clientFlags         flags.ClientFlags
	namespace           string
	sets                []string
	indexName           string
	vectorField         string
	dimensions          uint32
	distanceMetric      flags.DistanceMetricFlag
	indexMeta           map[string]string
	storageNamespace    flags.StringOptionalFlag
	storageSet          flags.StringOptionalFlag
	hnswMaxEdges        flags.Uint32OptionalFlag
	hnswEf              flags.Uint32OptionalFlag
	hnswConstructionEf  flags.Uint32OptionalFlag
	hnswBatchMaxRecords flags.Uint32OptionalFlag
	hnswBatchInterval   flags.Uint32OptionalFlag
	hnswBatchEnabled    flags.BoolOptionalFlag
	timeout             time.Duration
}{
	clientFlags:         *flags.NewClientFlags(),
	storageNamespace:    flags.StringOptionalFlag{},
	storageSet:          flags.StringOptionalFlag{},
	hnswMaxEdges:        flags.Uint32OptionalFlag{},
	hnswEf:              flags.Uint32OptionalFlag{},
	hnswConstructionEf:  flags.Uint32OptionalFlag{},
	hnswBatchMaxRecords: flags.Uint32OptionalFlag{},
	hnswBatchInterval:   flags.Uint32OptionalFlag{},
	hnswBatchEnabled:    flags.BoolOptionalFlag{},
}

func newCreateIndexFlagSet() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}                                                                                                                                                                                                                                                                                                                                                                  //nolint:lll // For readability
	flagSet.StringVarP(&createIndexFlags.namespace, flags.Namespace, "n", "", commonFlags.DefaultWrapHelpString("The namespace for the index."))                                                                                                                                                                                                                                                 //nolint:lll // For readability
	flagSet.StringSliceVarP(&createIndexFlags.sets, flags.Sets, "s", nil, commonFlags.DefaultWrapHelpString("The sets for the index."))                                                                                                                                                                                                                                                          //nolint:lll // For readability
	flagSet.StringVarP(&createIndexFlags.indexName, flags.IndexName, "i", "", commonFlags.DefaultWrapHelpString("The name of the index."))                                                                                                                                                                                                                                                       //nolint:lll // For readability
	flagSet.StringVarP(&createIndexFlags.vectorField, flags.VectorField, "f", "", commonFlags.DefaultWrapHelpString("The name of the vector field."))                                                                                                                                                                                                                                            //nolint:lll // For readability
	flagSet.Uint32VarP(&createIndexFlags.dimensions, flags.Dimension, "d", 0, commonFlags.DefaultWrapHelpString("The dimension of the vector field."))                                                                                                                                                                                                                                           //nolint:lll // For readability
	flagSet.VarP(&createIndexFlags.distanceMetric, flags.DistanceMetric, "m", commonFlags.DefaultWrapHelpString(fmt.Sprintf("The distance metric for the index. Valid values: %s", strings.Join(flags.DistanceMetricEnum(), ", "))))                                                                                                                                                             //nolint:lll // For readability
	flagSet.StringToStringVar(&createIndexFlags.indexMeta, flags.IndexMeta, nil, commonFlags.DefaultWrapHelpString("The distance metric for the index."))                                                                                                                                                                                                                                        //nolint:lll // For readability
	flagSet.DurationVar(&createIndexFlags.timeout, flags.Timeout, time.Second*5, commonFlags.DefaultWrapHelpString("The distance metric for the index."))                                                                                                                                                                                                                                        //nolint:lll // For readability
	flagSet.Var(&createIndexFlags.storageNamespace, flags.StorageNamespace, commonFlags.DefaultWrapHelpString("Optional storage namespace where the index is stored. Defaults to the index namespace."))                                                                                                                                                                                         //nolint:lll // For readability                                                                                                                                                                                                                  //nolint:lll // For readability
	flagSet.Var(&createIndexFlags.storageSet, flags.StorageSet, commonFlags.DefaultWrapHelpString("Optional storage set where the index is stored. Defaults to the index name."))                                                                                                                                                                                                                //nolint:lll // For readability                                                                                                                                                                                                                  //nolint:lll // For readability
	flagSet.Var(&createIndexFlags.hnswMaxEdges, flags.MaxEdges, commonFlags.DefaultWrapHelpString("Maximum number bi-directional links per HNSW vertex. Greater values of 'm' in general provide better recall for data with high dimensionality, while lower values work well for data with lower dimensionality. The storage space required for the index increases proportionally with 'm'")) //nolint:lll // For readability
	flagSet.Var(&createIndexFlags.hnswConstructionEf, flags.ConstructionEf, commonFlags.DefaultWrapHelpString("The number of candidate nearest neighbors shortlisted during index creation. Larger values provide better recall at the cost of longer index update times. The default is 100."))                                                                                                 //nolint:lll // For readability
	flagSet.Var(&createIndexFlags.hnswEf, flags.Ef, commonFlags.DefaultWrapHelpString("The default number of candidate nearest neighbors shortlisted during search. Larger values provide better recall at the cost of longer search times. The default is 100."))                                                                                                                               //nolint:lll // For readability
	flagSet.Var(&createIndexFlags.hnswBatchMaxRecords, flags.BatchMaxRecords, commonFlags.DefaultWrapHelpString("Maximum number of records to fit in a batch. The default value is 10000."))                                                                                                                                                                                                     //nolint:lll // For readability
	flagSet.Var(&createIndexFlags.hnswBatchInterval, flags.BatchInterval, commonFlags.DefaultWrapHelpString("The maximum amount of time in milliseconds to wait before finalizing a batch. The default value is 10000."))                                                                                                                                                                        //nolint:lll // For readability
	flagSet.Var(&createIndexFlags.hnswBatchEnabled, flags.BatchEnabled, commonFlags.DefaultWrapHelpString("Enables batching for index updates. Default is true meaning batching is enabled."))                                                                                                                                                                                                   //nolint:lll // For readability
	flagSet.AddFlagSet(createIndexFlags.clientFlags.NewClientFlagSet())

	return flagSet
}

var createIndexRequiredFlags = []string{
	flags.Namespace,
	flags.IndexName,
	flags.VectorField,
	flags.Dimension,
	flags.DistanceMetric,
}

// createIndexCmd represents the createIndex command
func newCreateIndexCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "index",
		Short: "A command for creating indexes",
		Long: `A command for creating indexes. An index is required to enable vector 
		search on your data. The index tells AVS where your data is located, 
		what your vectors look like, and how vectors should be compared to each other. 
		Optionally, you can tweak where your index is stored and how the HNSW algorithm 
		behaves. For more information see: https://aerospike.com/docs/vector

		For example:
			export ASVEC_HOST=<avs-ip>:5000
			asvec create index -i myindex -n test -s testset -d 256 -m COSINE --vector-field vector \
				--storage-namespace test --hnsw-batch-enabled false
			`,
		PreRunE: func(_ *cobra.Command, _ []string) error {
			if viper.IsSet(flags.Seeds) && viper.IsSet(flags.Host) {
				return fmt.Errorf("only --%s or --%s allowed", flags.Seeds, flags.Host)
			}

			return nil
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			logger.Debug("parsed flags",
				append(createIndexFlags.clientFlags.NewSLogAttr(),
					slog.String(flags.Namespace, createIndexFlags.namespace),
					slog.Any(flags.Sets, createIndexFlags.sets),
					slog.String(flags.IndexName, createIndexFlags.indexName),
					slog.String(flags.VectorField, createIndexFlags.vectorField),
					slog.Uint64(flags.Dimension, uint64(createIndexFlags.dimensions)),
					slog.Any(flags.IndexMeta, createIndexFlags.indexMeta),
					slog.String(flags.DistanceMetric, createIndexFlags.distanceMetric.String()),
					slog.Duration(flags.Timeout, createIndexFlags.timeout),
					slog.Any(flags.StorageNamespace, createIndexFlags.storageNamespace.String()),
					slog.Any(flags.StorageSet, createIndexFlags.storageSet.String()),
					slog.Any(flags.MaxEdges, createIndexFlags.hnswMaxEdges.String()),
					slog.Any(flags.Ef, createIndexFlags.hnswEf),
					slog.Any(flags.ConstructionEf, createIndexFlags.hnswConstructionEf.String()),
					slog.Any(flags.BatchMaxRecords, createIndexFlags.hnswBatchMaxRecords.String()),
					slog.Any(flags.BatchInterval, createIndexFlags.hnswBatchInterval.String()),
					slog.Any(flags.BatchEnabled, createIndexFlags.hnswBatchEnabled.String()),
				)...,
			)

			adminClient, err := createClientFromFlags(&createIndexFlags.clientFlags, createIndexFlags.timeout)
			if err != nil {
				return err
			}
			defer adminClient.Close()

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

			if !confirm(fmt.Sprintf(
				"Are you sure you want to create the index %s field %s?",
				nsAndSetString(
					createIndexFlags.namespace,
					createIndexFlags.sets,
				),
				createIndexFlags.vectorField,
			)) {
				return nil
			}

			ctx, cancel := context.WithTimeout(context.Background(), createIndexFlags.timeout)
			defer cancel()

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
		err := createIndexCmd.MarkFlagRequired(flag)
		if err != nil {
			panic(err)
		}
	}
}
