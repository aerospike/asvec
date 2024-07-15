package cmd

import (
	"asvec/cmd/flags"
	"context"
	"fmt"
	"log/slog"

	"github.com/aerospike/avs-client-go/protos"
	commonFlags "github.com/aerospike/tools-common-go/flags"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

//nolint:govet // Padding not a concern for a CLI
var indexUpdateFlags = &struct {
	clientFlags         flags.ClientFlags
	yes                 bool
	namespace           string
	indexName           string
	indexMeta           map[string]string
	hnswMaxMemQueueSize flags.Uint32OptionalFlag
	hnswBatch           flags.BatchingFlags
	hnswCache           flags.CachingFlags
	hnswHealer          flags.HealerFlags
	hnswMerge           flags.MergeFlags
}{
	clientFlags:         *flags.NewClientFlags(),
	hnswMaxMemQueueSize: flags.Uint32OptionalFlag{},
	hnswBatch:           *flags.NewHnswBatchingFlags(),
	hnswCache:           *flags.NewHnswCachingFlags(),
	hnswHealer:          *flags.NewHnswHealerFlags(),
	hnswMerge:           *flags.NewHnswMergeFlags(),
}

func newIndexUpdateFlagSet() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}
	flagSet.BoolVarP(&indexUpdateFlags.yes, flags.Yes, "y", false, commonFlags.DefaultWrapHelpString("When true do not prompt for confirmation."))                                           //nolint:lll // For readability
	flagSet.StringVarP(&indexUpdateFlags.namespace, flags.Namespace, "n", "", commonFlags.DefaultWrapHelpString("The namespace for the index."))                                             //nolint:lll // For readability
	flagSet.StringVarP(&indexUpdateFlags.indexName, flags.IndexName, "i", "", commonFlags.DefaultWrapHelpString("The name of the index."))                                                   //nolint:lll // For readability
	flagSet.StringToStringVar(&indexUpdateFlags.indexMeta, flags.IndexLabels, nil, commonFlags.DefaultWrapHelpString("The distance metric for the index."))                                  //nolint:lll // For readability
	flagSet.Var(&indexUpdateFlags.hnswMaxMemQueueSize, flags.HnswMaxMemQueueSize, commonFlags.DefaultWrapHelpString("Maximum size of in-memory queue for inserted/updated vector records.")) //nolint:lll // For readability
	flagSet.AddFlagSet(indexUpdateFlags.clientFlags.NewClientFlagSet())
	flagSet.AddFlagSet(indexUpdateFlags.hnswBatch.NewFlagSet())
	flagSet.AddFlagSet(indexUpdateFlags.hnswCache.NewFlagSet())
	flagSet.AddFlagSet(indexUpdateFlags.hnswHealer.NewFlagSet())
	flagSet.AddFlagSet(indexUpdateFlags.hnswMerge.NewFlagSet())

	return flagSet
}

var indexUpdateRequiredFlags = []string{
	flags.Namespace,
	flags.IndexName,
}

// updateIndexCmd represents the updateIndex command
func newIndexUpdateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "A command for updating your indexes",
		Long: fmt.Sprintf(`A command for updating the behavior of your hnsw indexes. 
Modify parameters such as batching, caching, index healing, and index merging. 
For guidance on updating your indexes and to view defaults, refer to: 
https://aerospike.com/docs/vector/operate/index-management"

For example:

%s
asvec index update -i myindex -n test --%s 10000 --%s 10000ms --%s 10s --%s 16 --%s 16
			`, HelpTxtSetupEnv, flags.BatchMaxRecords, flags.BatchInterval,
			flags.HnswCacheExpiry, flags.HnswHealerParallelism, flags.HnswMergeParallelism),
		PreRunE: func(_ *cobra.Command, _ []string) error {
			return checkSeedsAndHost()
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			debugFlags := indexUpdateFlags.clientFlags.NewSLogAttr()
			debugFlags = append(debugFlags, indexUpdateFlags.hnswBatch.NewSLogAttr()...)
			debugFlags = append(debugFlags, indexUpdateFlags.hnswCache.NewSLogAttr()...)
			debugFlags = append(debugFlags, indexUpdateFlags.hnswHealer.NewSLogAttr()...)
			debugFlags = append(debugFlags, indexUpdateFlags.hnswMerge.NewSLogAttr()...)
			logger.Debug("parsed flags",
				append(debugFlags,
					slog.Bool(flags.Yes, indexUpdateFlags.yes),
					slog.String(flags.Namespace, indexUpdateFlags.namespace),
					slog.String(flags.IndexName, indexUpdateFlags.indexName),
					slog.String(flags.HnswMaxMemQueueSize, indexUpdateFlags.hnswMaxMemQueueSize.String()),
				)...,
			)

			adminClient, err := createClientFromFlags(&indexUpdateFlags.clientFlags)
			if err != nil {
				return err
			}
			defer adminClient.Close()

			hnswParams := &protos.HnswIndexUpdate{
				MaxMemQueueSize: indexUpdateFlags.hnswMaxMemQueueSize.Val,
				BatchingParams: &protos.HnswBatchingParams{
					MaxRecords: indexUpdateFlags.hnswBatch.MaxRecords.Val,
					Interval:   indexUpdateFlags.hnswBatch.Interval.Uint32(),
				},
				CachingParams: &protos.HnswCachingParams{
					MaxEntries: indexUpdateFlags.hnswCache.MaxEntries.Val,
					Expiry:     indexUpdateFlags.hnswCache.Expiry.Uint64(),
				},
				HealerParams: &protos.HnswHealerParams{
					MaxScanRatePerNode: indexUpdateFlags.hnswHealer.MaxScanRatePerNode.Val,
					MaxScanPageSize:    indexUpdateFlags.hnswHealer.MaxScanPageSize.Val,
					ReindexPercent:     indexUpdateFlags.hnswHealer.ReindexPercent.Val,
					ScheduleDelay:      indexUpdateFlags.hnswHealer.ScheduleDelay.Uint64(),
					Parallelism:        indexUpdateFlags.hnswHealer.Parallelism.Val,
				},
				MergeParams: &protos.HnswIndexMergeParams{
					Parallelism: indexUpdateFlags.hnswMerge.Parallelism.Val,
				},
			}

			if !indexUpdateFlags.yes && !confirm(fmt.Sprintf(
				"Are you sure you want to update the index %s.%s?",
				indexUpdateFlags.namespace,
				indexUpdateFlags.indexName,
			)) {
				return nil
			}

			ctx, cancel := context.WithTimeout(context.Background(), indexUpdateFlags.clientFlags.Timeout)
			defer cancel()

			err = adminClient.IndexUpdate(
				ctx,
				indexUpdateFlags.namespace,
				indexUpdateFlags.indexName,
				indexUpdateFlags.indexMeta,
				hnswParams,
			)
			if err != nil {
				logger.Error("unable to update index", slog.Any("error", err))
				return err
			}

			view.Printf("Successfully updated index %s.%s", indexUpdateFlags.namespace, indexUpdateFlags.indexName)
			return nil
		},
	}
}

func init() {
	updateIndexCmd := newIndexUpdateCmd()
	indexCmd.AddCommand(updateIndexCmd)

	// TODO: Add custom template for usage to take into account terminal width
	// Ex: https://github.com/sigstore/cosign/pull/3011/files

	flagSet := newIndexUpdateFlagSet()
	updateIndexCmd.Flags().AddFlagSet(flagSet)

	for _, flag := range indexUpdateRequiredFlags {
		err := updateIndexCmd.MarkFlagRequired(flag)
		if err != nil {
			panic(err)
		}
	}
}
