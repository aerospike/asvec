package cmd

import (
	"asvec/cmd/flags"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	avs "github.com/aerospike/avs-client-go"
	"github.com/aerospike/avs-client-go/protos"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"google.golang.org/protobuf/encoding/protojson"
	"gopkg.in/yaml.v3"
)

//nolint:govet // Padding not a concern for a CLI
var indexCreateFlags = &struct {
	clientFlags                *flags.ClientFlags
	yes                        bool
	inputFile                  string
	namespace                  string
	set                        flags.StringOptionalFlag
	indexName                  string
	vectorField                string
	dimensions                 uint32
	distanceMetric             flags.DistanceMetricFlag
	indexLabels                map[string]string
	storageNamespace           flags.StringOptionalFlag
	storageSet                 flags.StringOptionalFlag
	hnswMaxEdges               flags.Uint32OptionalFlag
	hnswEf                     flags.Uint32OptionalFlag
	hnswConstructionEf         flags.Uint32OptionalFlag
	hnswMaxMemQueueSize        flags.Uint32OptionalFlag
	hnswBatch                  flags.BatchingFlags
	hnswCache                  flags.CachingFlags
	hnswHealer                 flags.HealerFlags
	hnswMerge                  flags.MergeFlags
	enableVectorIntegrityCheck flags.BoolOptionalFlag
}{
	clientFlags:                rootFlags.clientFlags,
	set:                        flags.StringOptionalFlag{},
	storageNamespace:           flags.StringOptionalFlag{},
	storageSet:                 flags.StringOptionalFlag{},
	hnswMaxEdges:               flags.Uint32OptionalFlag{},
	hnswEf:                     flags.Uint32OptionalFlag{},
	hnswConstructionEf:         flags.Uint32OptionalFlag{},
	hnswMaxMemQueueSize:        flags.Uint32OptionalFlag{},
	hnswBatch:                  *flags.NewHnswBatchingFlags(),
	hnswCache:                  *flags.NewHnswCachingFlags(),
	hnswHealer:                 *flags.NewHnswHealerFlags(),
	hnswMerge:                  *flags.NewHnswMergeFlags(),
	enableVectorIntegrityCheck: flags.BoolOptionalFlag{},
}

func newIndexCreateFlagSet() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}
	flagSet.BoolVarP(&indexCreateFlags.yes, flags.Yes, flags.YesShort, false, "When true do not prompt for confirmation.")
	flagSet.StringVar(&indexCreateFlags.inputFile, flags.InputFile, StdIn, "A yaml file containing IndexDefinitions created using \"asvec index list --yaml\"")                                                                                                                                                                                                    //nolint:lll // For readability
	flagSet.StringVarP(&indexCreateFlags.namespace, flags.Namespace, flags.NamespaceShort, "", "The namespace for the index.")                                                                                                                                                                                                                                     //nolint:lll // For readability
	flagSet.VarP(&indexCreateFlags.set, flags.Set, flags.SetShort, "The sets for the index.")                                                                                                                                                                                                                                                                      //nolint:lll // For readability                                                                                                                                                                                                                                                                 //nolint:lll // For readability
	flagSet.StringVarP(&indexCreateFlags.indexName, flags.IndexName, flags.IndexNameShort, "", "The name of the index.")                                                                                                                                                                                                                                           //nolint:lll // For readability
	flagSet.StringVarP(&indexCreateFlags.vectorField, flags.VectorField, flags.VectorFieldShort, "", "The name of the vector field.")                                                                                                                                                                                                                              //nolint:lll // For readability
	flagSet.Uint32VarP(&indexCreateFlags.dimensions, flags.Dimension, flags.DimensionShort, 0, "The dimension of the vector field.")                                                                                                                                                                                                                               //nolint:lll // For readability
	flagSet.VarP(&indexCreateFlags.distanceMetric, flags.DistanceMetric, flags.DistanceMetricShort, fmt.Sprintf("The distance metric for the index. Valid values: %s", strings.Join(flags.DistanceMetricEnum(), ", ")))                                                                                                                                            //nolint:lll // For readability
	flagSet.StringToStringVar(&indexCreateFlags.indexLabels, flags.IndexLabels, nil, "Optional labels to assign to the index. Example: \"model=all-MiniLM-L6-v2,foo=bar\"")                                                                                                                                                                                        //nolint:lll // For readability
	flagSet.Var(&indexCreateFlags.storageNamespace, flags.StorageNamespace, "Optional storage namespace where the index is stored. Defaults to the index namespace.")                                                                                                                                                                                              //nolint:lll // For readability
	flagSet.Var(&indexCreateFlags.storageSet, flags.StorageSet, "Optional storage set where the index is stored. Defaults to the index name.")                                                                                                                                                                                                                     //nolint:lll // For readability
	flagSet.Var(&indexCreateFlags.hnswMaxEdges, flags.HnswMaxEdges, "Maximum number bi-directional links per HNSW vertex. Greater values of 'm' in general provide better recall for data with high dimensionality, while lower values work well for data with lower dimensionality. The storage space required for the index increases proportionally with 'm'.") //nolint:lll // For readability
	flagSet.Var(&indexCreateFlags.hnswConstructionEf, flags.HnswConstructionEf, "The number of candidate nearest neighbors shortlisted during index creation. Larger values provide better recall at the cost of longer index update times.")                                                                                                                      //nolint:lll // For readability
	flagSet.Var(&indexCreateFlags.hnswEf, flags.HnswEf, "The default number of candidate nearest neighbors shortlisted during search. Larger values provide better recall at the cost of longer search times.")                                                                                                                                                    //nolint:lll // For readability
	flagSet.Var(&indexCreateFlags.hnswMaxMemQueueSize, flags.HnswMaxMemQueueSize, "Maximum size of in-memory queue for inserted/updated vector records.")                                                                                                                                                                                                          //nolint:lll // For readability                                                                                                                                                                                                                      //nolint:lll // For readability
	flagSet.Var(&indexCreateFlags.enableVectorIntegrityCheck, flags.EnableVectorIntegrityCheck, "Enable/disable vector integrity check. Defaults to enabled.")                                                                                                                                                                                                     //nolint:lll // For readability
	flagSet.AddFlagSet(indexCreateFlags.hnswBatch.NewFlagSet())
	flagSet.AddFlagSet(indexCreateFlags.hnswCache.NewFlagSet())
	flagSet.AddFlagSet(indexCreateFlags.hnswHealer.NewFlagSet())
	flagSet.AddFlagSet(indexCreateFlags.hnswMerge.NewFlagSet())

	// For backwards compatibility
	flagSet.Var(&indexCreateFlags.set, "sets", "The sets for the index.")
	err := flagSet.MarkHidden("sets")

	if err != nil {
		panic(err)
	}

	return flagSet
}

var indexCreateRequiredFlags = []string{
	flags.Namespace,
	flags.IndexName,
	flags.VectorField,
	flags.Dimension,
	flags.DistanceMetric,
}
var stdinIndexDefinitions *protos.IndexDefinitionList

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
https://aerospike.com/docs/vector/operate/index-management

For example:

%s
asvec index create -i myindex -n test -s testset -d 256 -m COSINE --%s vector \
	--%s test
			`, HelpTxtSetupEnv, flags.VectorField, flags.StorageNamespace),
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			err := checkSeedsAndHost()
			if err != nil {
				return err
			}

			oneRequiredFlagsSet := false
			configureRequiredFlags := true

			for _, name := range indexCreateRequiredFlags {
				if viper.IsSet(name) {
					oneRequiredFlagsSet = true
					break
				}
			}

			if !oneRequiredFlagsSet {
				var data []byte

				if indexCreateFlags.inputFile != StdIn {
					logger.Info("reading input file")
					r, err := os.Open(indexCreateFlags.inputFile)
					if err != nil {
						logger.Error("failed to open index definitions file", slog.Any("error", err))
						return err
					}

					defer r.Close()

					data, err = io.ReadAll(r)
					if err != nil {
						logger.Error("failed to read index definitions", slog.Any("error", err))
						return err
					}
				} else {
					logger.Debug("checking if index definitions are being piped from stdin")

					stat, _ := os.Stdin.Stat()
					if (stat.Mode() & os.ModeCharDevice) == 0 {
						logger.Debug("stdin is from a pipe")

						data, err = io.ReadAll(os.Stdin)
						if err != nil {
							logger.Error("failed to read index definitions from stdin", slog.Any("error", err))
							return err
						}
					} else {
						logger.Debug("no data is being piped from stdin")
					}
				}

				if data != nil {
					logger.Debug("read index definitions", slog.Any("data", data))

					// Unmarshal YAML data
					intermediate := map[string]interface{}{}
					err = yaml.Unmarshal(data, &intermediate)

					if err != nil {
						logger.Error("failed to unmarshal index definitions to untyped map", slog.Any("error", err))
						return err
					}

					jsonBytes, err := json.Marshal(intermediate)
					if err != nil {
						logger.Error("failed to marshal index definitions to json", slog.Any("error", err))
						return err
					}

					logger.Debug("marshalled index definitions", slog.Any("data", string(jsonBytes)))

					stdinIndexDefinitions = &protos.IndexDefinitionList{}

					err = protojson.Unmarshal(jsonBytes, stdinIndexDefinitions)
					if err != nil {
						logger.Error("failed to unmarshal index definitions IndexDefinitionList", slog.Any("error", err))
						return err
					}

					logger.Debug("parsed index definitions from stdin", slog.Any("indexes", stdinIndexDefinitions))
					configureRequiredFlags = false
				}
			}

			if configureRequiredFlags {
				markFlagsRequired(cmd, indexCreateRequiredFlags)
			}

			return nil
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
					slog.Any(flags.Set, indexCreateFlags.set),
					slog.String(flags.IndexName, indexCreateFlags.indexName),
					slog.String(flags.VectorField, indexCreateFlags.vectorField),
					slog.Uint64(flags.Dimension, uint64(indexCreateFlags.dimensions)),
					slog.Any(flags.IndexLabels, indexCreateFlags.indexLabels),
					slog.Any(flags.DistanceMetric, indexCreateFlags.distanceMetric),
					slog.Any(flags.StorageNamespace, indexCreateFlags.storageNamespace.Val),
					slog.Any(flags.StorageSet, indexCreateFlags.storageSet.Val),
					slog.Any(flags.HnswMaxEdges, indexCreateFlags.hnswMaxEdges.Val),
					slog.Any(flags.HnswEf, indexCreateFlags.hnswEf.Val),
					slog.Any(flags.HnswConstructionEf, indexCreateFlags.hnswConstructionEf.Val),
					slog.Any(flags.HnswMaxMemQueueSize, indexCreateFlags.hnswMaxMemQueueSize.Val),
					slog.Any(flags.EnableVectorIntegrityCheck, indexCreateFlags.enableVectorIntegrityCheck),
				)...,
			)

			client, err := createClientFromFlags(indexCreateFlags.clientFlags)
			if err != nil {
				return err
			}
			defer client.Close()

			if stdinIndexDefinitions != nil {
				return runCreateIndexFromDef(client)
			}

			return runCreateIndexFromFlags(client)
		},
	}
}

func runCreateIndexFromDef(client *avs.Client) error {
	if len(stdinIndexDefinitions.GetIndices()) == 0 {
		view.Print("No indexes to create")
		return nil
	}

	successful := 0

	for _, indexDef := range stdinIndexDefinitions.GetIndices() {
		ctx, cancel := context.WithTimeout(context.Background(), indexCreateFlags.clientFlags.Timeout)

		err := client.IndexCreateFromIndexDef(ctx, indexDef)

		cancel()

		if err != nil {
			logger.Warn("failed to create index from yaml", slog.Any("error", err))
			view.Printf("Failed to create index %s.%s from yaml: %s",
				nsAndSetString(
					indexDef.Id.Namespace,
					indexDef.SetFilter,
				),
				indexDef.Id.Name, err)
		} else {
			view.Printf("Successfully created index %s.%s", nsAndSetString(
				indexDef.Id.Namespace,
				indexDef.SetFilter,
			), indexDef.Id.Name)

			successful++
		}
	}

	if successful == 0 {
		err := fmt.Errorf("unable to create any new indexes")
		logger.Error(err.Error())
		view.Print("Unable to create any new indexes")

		return err
	} else if successful < len(stdinIndexDefinitions.GetIndices()) {
		err := fmt.Errorf("some indexes failed to create")
		logger.Warn(err.Error())
		view.Print("Some indexes failed to be created")

		return err
	}

	view.Print("Successfully created all indexes from yaml")

	return nil
}

func runCreateIndexFromFlags(client *avs.Client) error {
	if !indexCreateFlags.yes && !confirm(fmt.Sprintf(
		"Are you sure you want to create the index %s.%s on field %s?",
		nsAndSetString(
			indexCreateFlags.namespace,
			indexCreateFlags.set.Val,
		),
		indexCreateFlags.indexName,
		indexCreateFlags.vectorField,
	)) {
		return nil
	}

	sets := []string{}

	if indexCreateFlags.set.Val != nil {
		sets = append(sets, *indexCreateFlags.set.Val)
	}

	indexOpts := &avs.IndexCreateOpts{
		Sets:   sets,
		Labels: indexCreateFlags.indexLabels,
		Storage: &protos.IndexStorage{
			Namespace: indexCreateFlags.storageNamespace.Val,
			Set:       indexCreateFlags.storageSet.Val,
		},
		HnswParams: &protos.HnswParams{
			M:               indexCreateFlags.hnswMaxEdges.Val,
			Ef:              indexCreateFlags.hnswEf.Val,
			EfConstruction:  indexCreateFlags.hnswConstructionEf.Val,
			MaxMemQueueSize: indexCreateFlags.hnswMaxMemQueueSize.Val,
			BatchingParams: &protos.HnswBatchingParams{
				MaxIndexRecords:   indexCreateFlags.hnswBatch.MaxIndexRecords.Val,
				IndexInterval:     indexCreateFlags.hnswBatch.IndexInterval.Uint32(),
				MaxReindexRecords: indexCreateFlags.hnswBatch.MaxReindexRecords.Val,
				ReindexInterval:   indexCreateFlags.hnswBatch.ReindexInterval.Uint32(),
			},
			IndexCachingParams: &protos.HnswCachingParams{
				MaxEntries: indexCreateFlags.hnswCache.MaxEntries.Val,
				Expiry:     indexCreateFlags.hnswCache.Expiry.Int64(),
			},
			HealerParams: &protos.HnswHealerParams{
				MaxScanRatePerNode: indexCreateFlags.hnswHealer.MaxScanRatePerNode.Val,
				MaxScanPageSize:    indexCreateFlags.hnswHealer.MaxScanPageSize.Val,
				ReindexPercent:     indexCreateFlags.hnswHealer.ReindexPercent.Val,
				Schedule:           indexCreateFlags.hnswHealer.Schedule.Val,
				Parallelism:        indexCreateFlags.hnswHealer.Parallelism.Val,
			},
			MergeParams: &protos.HnswIndexMergeParams{
				IndexParallelism:   indexCreateFlags.hnswMerge.IndexParallelism.Val,
				ReIndexParallelism: indexCreateFlags.hnswMerge.ReIndexParallelism.Val,
			},
			EnableVectorIntegrityCheck: indexCreateFlags.enableVectorIntegrityCheck.Val,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), indexCreateFlags.clientFlags.Timeout)
	defer cancel()

	err := client.IndexCreate(
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
}

func init() {
	createIndexCmd := newIndexCreateCmd()
	indexCmd.AddCommand(createIndexCmd)

	// TODO: Add custom template for usage to take into account terminal width
	// Ex: https://github.com/sigstore/cosign/pull/3011/files
	flagSet := newIndexCreateFlagSet()
	createIndexCmd.Flags().AddFlagSet(flagSet)
}

func markFlagsRequired(cmd *cobra.Command, flagNames []string) {
	for _, flag := range flagNames {
		err := cmd.MarkFlagRequired(flag)
		if err != nil {
			panic(err)
		}
	}
}
