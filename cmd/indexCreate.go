package cmd

import (
	"asvec/cmd/flags"
	"context"
	"fmt"
	"log/slog"
	"strings"

	avs "github.com/aerospike/avs-client-go"
	"github.com/aerospike/avs-client-go/protos"
	commonFlags "github.com/aerospike/tools-common-go/flags"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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
	hnswMaxMemQueueSize flags.Uint32OptionalFlag
	hnswBatch           flags.BatchingFlags
	hnswCache           flags.CachingFlags
	hnswHealer          flags.HealerFlags
	hnswMerge           flags.MergeFlags
}{
	clientFlags:         *flags.NewClientFlags(),
	storageNamespace:    flags.StringOptionalFlag{},
	storageSet:          flags.StringOptionalFlag{},
	hnswMaxEdges:        flags.Uint32OptionalFlag{},
	hnswEf:              flags.Uint32OptionalFlag{},
	hnswConstructionEf:  flags.Uint32OptionalFlag{},
	hnswMaxMemQueueSize: flags.Uint32OptionalFlag{},
	hnswBatch:           *flags.NewHnswBatchingFlags(),
	hnswCache:           *flags.NewHnswCachingFlags(),
	hnswHealer:          *flags.NewHnswHealerFlags(),
	hnswMerge:           *flags.NewHnswMergeFlags(),
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
	flagSet.Var(&indexCreateFlags.hnswMaxMemQueueSize, flags.HnswMaxMemQueueSize, commonFlags.DefaultWrapHelpString("Maximum size of in-memory queue for inserted/updated vector records."))                                                                                                                                                                                                      //nolint:lll // For readability                                                                                                                                                                       //nolint:lll // For readability
	flagSet.AddFlagSet(indexCreateFlags.clientFlags.NewClientFlagSet())
	flagSet.AddFlagSet(indexCreateFlags.hnswBatch.NewFlagSet())
	flagSet.AddFlagSet(indexCreateFlags.hnswCache.NewFlagSet())
	flagSet.AddFlagSet(indexCreateFlags.hnswHealer.NewFlagSet())
	flagSet.AddFlagSet(indexCreateFlags.hnswMerge.NewFlagSet())

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
behaves. For guidance on creating indexes and for viewing defaults, refer to: 
https://aerospike.com/docs/vector/operate/index-management"

For example:

%s
asvec index create -i myindex -n test -s testset -d 256 -m COSINE --%s vector \
	--%s test
			`, HelpTxtSetupEnv, flags.VectorField, flags.StorageNamespace),
		PreRunE: func(_ *cobra.Command, _ []string) error {
			return checkSeedsAndHost()
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			debugFlags := indexCreateFlags.clientFlags.NewSLogAttr()
			debugFlags = append(debugFlags, indexCreateFlags.hnswBatch.NewSLogAttr()...)
			debugFlags = append(debugFlags, indexCreateFlags.hnswCache.NewSLogAttr()...)
			debugFlags = append(debugFlags, indexCreateFlags.hnswHealer.NewSLogAttr()...)
			debugFlags = append(debugFlags, indexCreateFlags.hnswMerge.NewSLogAttr()...)
			logger.Debug("parsed flags",
				append(debugFlags,
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
					slog.Any(flags.HnswMaxMemQueueSize, indexCreateFlags.hnswMaxMemQueueSize.String()),
				)...,
			)

			adminClient, err := createClientFromFlags(&indexCreateFlags.clientFlags)
			if err != nil {
				return err
			}
			defer adminClient.Close()

			indexOpts := &avs.IndexCreateOpts{
				Sets:     indexCreateFlags.sets,
				MetaData: indexCreateFlags.indexMeta,
				Storage: &protos.IndexStorage{
					Namespace: indexCreateFlags.storageNamespace.Val,
					Set:       indexCreateFlags.storageSet.Val,
				},
				HnswParams: &protos.HnswParams{
					M:              indexCreateFlags.hnswMaxEdges.Val,
					Ef:             indexCreateFlags.hnswEf.Val,
					EfConstruction: indexCreateFlags.hnswConstructionEf.Val,
					BatchingParams: &protos.HnswBatchingParams{
						MaxRecords: indexCreateFlags.hnswBatch.MaxRecords.Val,
						Interval:   indexCreateFlags.hnswBatch.Interval.Uint32(),
					},
					CachingParams: &protos.HnswCachingParams{
						MaxEntries: indexCreateFlags.hnswCache.MaxEntries.Val,
						Expiry:     indexCreateFlags.hnswCache.Expiry.Uint64(),
					},
					HealerParams: &protos.HnswHealerParams{
						MaxScanRatePerNode: indexCreateFlags.hnswHealer.MaxScanRatePerNode.Val,
						MaxScanPageSize:    indexCreateFlags.hnswHealer.MaxScanPageSize.Val,
						ReindexPercent:     indexCreateFlags.hnswHealer.ReindexPercent.Val,
						ScheduleDelay:      indexCreateFlags.hnswHealer.ScheduleDelay.Uint64(),
						Parallelism:        indexCreateFlags.hnswHealer.Parallelism.Val,
					},
					MergeParams: &protos.HnswIndexMergeParams{
						Parallelism: indexCreateFlags.hnswMerge.Parallelism.Val,
					},
				},
			}

			if !indexCreateFlags.yes && !confirm(fmt.Sprintf(
				"Are you sure you want to create the index %s.%s on field %s?",
				nsAndSetString(
					indexCreateFlags.namespace,
					indexCreateFlags.sets,
				),
				indexCreateFlags.indexName,
				indexCreateFlags.vectorField,
			)) {
				return nil
			}

			ctx, cancel := context.WithTimeout(context.Background(), indexCreateFlags.clientFlags.Timeout)
			defer cancel()

			err = adminClient.IndexCreate(
				ctx,
				indexCreateFlags.namespace,
				indexCreateFlags.indexName,
				indexCreateFlags.vectorField,
				indexCreateFlags.dimensions,
				protos.VectorDistanceMetric(protos.VectorDistanceMetric_value[indexCreateFlags.distanceMetric.String()]),
				indexOpts,
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
