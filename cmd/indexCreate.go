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

	"github.com/aerospike/avs-client-go/protos"
	commonFlags "github.com/aerospike/tools-common-go/flags"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

//nolint:govet // Padding not a concern for a CLI
var indexCreateFlags = &struct {
	clientFlags         flags.ClientFlags
	yes                 bool
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

func newIndexCreateFlagSet() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}
	flagSet.BoolVarP(&indexCreateFlags.yes, flags.Yes, "y", false, commonFlags.DefaultWrapHelpString("When true do not prompt for confirmation."))                                                                                                                                                                                                                                                //nolint:lll // For readability
	flagSet.StringVarP(&indexCreateFlags.namespace, flags.Namespace, "n", "", commonFlags.DefaultWrapHelpString("The namespace for the index."))                                                                                                                                                                                                                                                  //nolint:lll // For readability
	flagSet.StringSliceVarP(&indexCreateFlags.sets, flags.Sets, "s", nil, commonFlags.DefaultWrapHelpString("The sets for the index."))                                                                                                                                                                                                                                                           //nolint:lll // For readability
	flagSet.StringVarP(&indexCreateFlags.indexName, flags.IndexName, "i", "", commonFlags.DefaultWrapHelpString("The name of the index."))                                                                                                                                                                                                                                                        //nolint:lll // For readability
	flagSet.StringVarP(&indexCreateFlags.vectorField, flags.VectorField, "f", "", commonFlags.DefaultWrapHelpString("The name of the vector field."))                                                                                                                                                                                                                                             //nolint:lll // For readability
	flagSet.Uint32VarP(&indexCreateFlags.dimensions, flags.Dimension, "d", 0, commonFlags.DefaultWrapHelpString("The dimension of the vector field."))                                                                                                                                                                                                                                            //nolint:lll // For readability
	flagSet.VarP(&indexCreateFlags.distanceMetric, flags.DistanceMetric, "m", commonFlags.DefaultWrapHelpString(fmt.Sprintf("The distance metric for the index. Valid values: %s", strings.Join(flags.DistanceMetricEnum(), ", "))))                                                                                                                                                              //nolint:lll // For readability
	flagSet.StringToStringVar(&indexCreateFlags.indexMeta, flags.IndexMeta, nil, commonFlags.DefaultWrapHelpString("The distance metric for the index."))                                                                                                                                                                                                                                         //nolint:lll // For readability                                                                                                                                                                                                                                //nolint:lll // For readability
	flagSet.Var(&indexCreateFlags.storageNamespace, flags.StorageNamespace, commonFlags.DefaultWrapHelpString("Optional storage namespace where the index is stored. Defaults to the index namespace."))                                                                                                                                                                                          //nolint:lll // For readability                                                                                                                                                                                                                  //nolint:lll // For readability
	flagSet.Var(&indexCreateFlags.storageSet, flags.StorageSet, commonFlags.DefaultWrapHelpString("Optional storage set where the index is stored. Defaults to the index name."))                                                                                                                                                                                                                 //nolint:lll // For readability                                                                                                                                                                                                                  //nolint:lll // For readability
	flagSet.Var(&indexCreateFlags.hnswMaxEdges, flags.MaxEdges, commonFlags.DefaultWrapHelpString("Maximum number bi-directional links per HNSW vertex. Greater values of 'm' in general provide better recall for data with high dimensionality, while lower values work well for data with lower dimensionality. The storage space required for the index increases proportionally with 'm'.")) //nolint:lll // For readability
	flagSet.Var(&indexCreateFlags.hnswConstructionEf, flags.ConstructionEf, commonFlags.DefaultWrapHelpString("The number of candidate nearest neighbors shortlisted during index creation. Larger values provide better recall at the cost of longer index update times. The default is 100."))                                                                                                  //nolint:lll // For readability
	flagSet.Var(&indexCreateFlags.hnswEf, flags.Ef, commonFlags.DefaultWrapHelpString("The default number of candidate nearest neighbors shortlisted during search. Larger values provide better recall at the cost of longer search times. The default is 100."))                                                                                                                                //nolint:lll // For readability
	flagSet.Var(&indexCreateFlags.hnswBatchMaxRecords, flags.BatchMaxRecords, commonFlags.DefaultWrapHelpString("Maximum number of records to fit in a batch. The default value is 10000."))                                                                                                                                                                                                      //nolint:lll // For readability
	flagSet.Var(&indexCreateFlags.hnswBatchInterval, flags.BatchInterval, commonFlags.DefaultWrapHelpString("The maximum amount of time in milliseconds to wait before finalizing a batch. The default value is 10000."))                                                                                                                                                                         //nolint:lll // For readability
	flagSet.Var(&indexCreateFlags.hnswBatchEnabled, flags.BatchEnabled, commonFlags.DefaultWrapHelpString("Enables batching for index updates. Default is true meaning batching is enabled."))                                                                                                                                                                                                    //nolint:lll // For readability
	flagSet.AddFlagSet(indexCreateFlags.clientFlags.NewClientFlagSet())

	return flagSet
}

var indexCreateRequiredFlags = []string{
	flags.Namespace,
	flags.IndexName,
	flags.VectorField,
	flags.Dimension,
	flags.DistanceMetric,
}

// createIndexCmd represents the createIndex command
func newIndexCreateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create",
		Short: "A command for creating indexes",
		Long: fmt.Sprintf(`A command for creating indexes. An index is required to enable vector 
search on your data. The index tells AVS where your data is located, 
what your vectors look like, and how vectors should be compared to each other. 
Optionally, you can tweak where your index is stored and how the HNSW algorithm 
behaves. For more information see: https://aerospike.com/docs/vector

For example:

%s
asvec index create -i myindex -n test -s testset -d 256 -m COSINE --%s vector \
	--%s test --%s false
			`, HelpTxtSetupEnv, flags.VectorField, flags.StorageNamespace, flags.BatchEnabled),
		PreRunE: func(_ *cobra.Command, _ []string) error {
			if viper.IsSet(flags.Seeds) && viper.IsSet(flags.Host) {
				return fmt.Errorf("only --%s or --%s allowed", flags.Seeds, flags.Host)
			}

			return nil
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			logger.Debug("parsed flags",
				append(indexCreateFlags.clientFlags.NewSLogAttr(),
					slog.Bool(flags.Yes, indexCreateFlags.yes),
					slog.String(flags.Namespace, indexCreateFlags.namespace),
					slog.Any(flags.Sets, indexCreateFlags.sets),
					slog.String(flags.IndexName, indexCreateFlags.indexName),
					slog.String(flags.VectorField, indexCreateFlags.vectorField),
					slog.Uint64(flags.Dimension, uint64(indexCreateFlags.dimensions)),
					slog.Any(flags.IndexMeta, indexCreateFlags.indexMeta),
					slog.String(flags.DistanceMetric, indexCreateFlags.distanceMetric.String()),
					slog.Any(flags.StorageNamespace, indexCreateFlags.storageNamespace.String()),
					slog.Any(flags.StorageSet, indexCreateFlags.storageSet.String()),
					slog.Any(flags.MaxEdges, indexCreateFlags.hnswMaxEdges.String()),
					slog.Any(flags.Ef, indexCreateFlags.hnswEf),
					slog.Any(flags.ConstructionEf, indexCreateFlags.hnswConstructionEf.String()),
					slog.Any(flags.BatchMaxRecords, indexCreateFlags.hnswBatchMaxRecords.String()),
					slog.Any(flags.BatchInterval, indexCreateFlags.hnswBatchInterval.String()),
					slog.Any(flags.BatchEnabled, indexCreateFlags.hnswBatchEnabled.String()),
				)...,
			)

			adminClient, err := createClientFromFlags(&indexCreateFlags.clientFlags)
			if err != nil {
				return err
			}
			defer adminClient.Close()

			// Inverted to make it easier to understand
			var hnswBatchDisabled *bool
			if indexCreateFlags.hnswBatchEnabled.Val != nil {
				bd := !(*indexCreateFlags.hnswBatchEnabled.Val)
				hnswBatchDisabled = &bd
			}

			indexStorage := &protos.IndexStorage{
				Namespace: indexCreateFlags.storageNamespace.Val,
				Set:       indexCreateFlags.storageSet.Val,
			}

			hnswParams := &protos.HnswParams{
				M:              indexCreateFlags.hnswMaxEdges.Val,
				Ef:             indexCreateFlags.hnswEf.Val,
				EfConstruction: indexCreateFlags.hnswConstructionEf.Val,
				BatchingParams: &protos.HnswBatchingParams{
					MaxRecords: indexCreateFlags.hnswBatchMaxRecords.Val,
					Interval:   indexCreateFlags.hnswBatchInterval.Val,
					Disabled:   hnswBatchDisabled,
				},
			}

			if !indexCreateFlags.yes && !confirm(fmt.Sprintf(
				"Are you sure you want to create the index %s field %s?",
				nsAndSetString(
					indexCreateFlags.namespace,
					indexCreateFlags.sets,
				),
				indexCreateFlags.vectorField,
			)) {
				return nil
			}

			ctx, cancel := context.WithTimeout(context.Background(), indexCreateFlags.clientFlags.Timeout)
			defer cancel()

			err = adminClient.IndexCreate(
				ctx,
				indexCreateFlags.namespace,
				indexCreateFlags.sets,
				indexCreateFlags.indexName,
				indexCreateFlags.vectorField,
				indexCreateFlags.dimensions,
				protos.VectorDistanceMetric(protos.VectorDistanceMetric_value[indexCreateFlags.distanceMetric.String()]),
				hnswParams,
				indexCreateFlags.indexMeta,
				indexStorage,
			)
			if err != nil {
				logger.Error("unable to create index", slog.Any("error", err))
				return err
			}

			view.Printf("Successfully created index %s.%s", indexCreateFlags.namespace, indexCreateFlags.indexName)
			return nil
		},
	}
}

func init() {
	createIndexCmd := newIndexCreateCmd()
	indexCmd.AddCommand(createIndexCmd)

	// TODO: Add custom template for usage to take into account terminal width
	// Ex: https://github.com/sigstore/cosign/pull/3011/files

	flagSet := newIndexCreateFlagSet()
	createIndexCmd.Flags().AddFlagSet(flagSet)

	for _, flag := range indexCreateRequiredFlags {
		err := createIndexCmd.MarkFlagRequired(flag)
		if err != nil {
			panic(err)
		}
	}
}