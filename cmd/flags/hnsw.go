package flags

import (
	"fmt"
	"log/slog"

	"github.com/spf13/pflag"
)

type BatchingFlags struct {
	MaxIndexRecords   Uint32OptionalFlag
	IndexInterval     DurationOptionalFlag
	MaxReindexRecords Uint32OptionalFlag
	ReindexInterval   DurationOptionalFlag
}

func NewHnswBatchingFlags() *BatchingFlags {
	return &BatchingFlags{
		MaxIndexRecords:   Uint32OptionalFlag{},
		IndexInterval:     DurationOptionalFlag{},
		MaxReindexRecords: Uint32OptionalFlag{},
		ReindexInterval:   DurationOptionalFlag{},
	}
}

func (cf *BatchingFlags) NewFlagSet() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}
	flagSet.Var(&cf.MaxIndexRecords, BatchMaxIndexRecords, "Maximum number of records to fit in a batch. Defaults to 100_000")                                                                          //nolint:lll // For readability
	flagSet.Var(&cf.IndexInterval, BatchIndexInterval, "The maximum amount of time to wait before finalizing a batch. Defaults to 30_000")                                                              //nolint:lll // For readability
	flagSet.Var(&cf.MaxReindexRecords, BatchMaxReindexRecords, fmt.Sprintf("Maximum number of re-index records to fit in a batch. Defaults to the maximum of %s / 10 and 1000.", BatchMaxIndexRecords)) //nolint:lll // For readability
	flagSet.Var(&cf.ReindexInterval, BatchReindexInterval, fmt.Sprintf("The maximum amount of time to wait before finalizing a re-index batch. Defaults to %s", BatchIndexInterval))                    //nolint:lll // For readability

	return flagSet
}

func (cf *BatchingFlags) NewSLogAttr() []any {
	return []any{
		slog.Any(BatchMaxIndexRecords, cf.MaxIndexRecords.Val),
		slog.Any(BatchIndexInterval, cf.IndexInterval.Val),
		slog.Any(BatchMaxReindexRecords, cf.MaxIndexRecords.Val),
		slog.Any(BatchReindexInterval, cf.IndexInterval.Val),
	}
}

type IndexCachingFlags struct {
	MaxEntries Uint64OptionalFlag
	Expiry     InfDurationOptionalFlag
}

func NewHnswIndexCachingFlags() *IndexCachingFlags {
	return &IndexCachingFlags{
		MaxEntries: Uint64OptionalFlag{},
		Expiry:     InfDurationOptionalFlag{},
	}
}

//nolint:lll // For readability
func (cf *IndexCachingFlags) NewFlagSet() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}
	flagSet.Var(&cf.MaxEntries, HnswIndexCacheMaxEntries, "Maximum number of entries to cache. Defaults to 2_000_000.")
	flagSet.Var(&cf.Expiry, HnswIndexCacheExpiry, "A cache entry will expire after this amount of time has passed since the entry was added to cache, or -1 to never expire. Defaults to 3_600_000 (1 hour).")

	return flagSet
}

func (cf *IndexCachingFlags) NewSLogAttr() []any {
	return []any{
		slog.Any(HnswIndexCacheMaxEntries, cf.MaxEntries.Val),
		slog.String(HnswIndexCacheExpiry, cf.Expiry.String()),
	}
}

type RecordCachingFlags struct {
	MaxEntries Uint64OptionalFlag
	Expiry     InfDurationOptionalFlag
}

func NewHnswRecordCachingFlags() *RecordCachingFlags {
	return &RecordCachingFlags{
		MaxEntries: Uint64OptionalFlag{},
		Expiry:     InfDurationOptionalFlag{},
	}
}

//nolint:lll // For readability
func (cf *RecordCachingFlags) NewFlagSet() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}
	flagSet.Var(&cf.MaxEntries, HnswRecordCacheMaxEntries, "Maximum number of entries to cache. Defaults to 2_000_000.")
	flagSet.Var(&cf.Expiry, HnswRecordCacheExpiry, "A cache entry will expire after this amount of time has passed since the entry was added to cache, or -1 to never expire. Defaults to 3_600_000 (1 hour).")

	return flagSet
}

func (cf *RecordCachingFlags) NewSLogAttr() []any {
	return []any{
		slog.Any(HnswRecordCacheMaxEntries, cf.MaxEntries.Val),
		slog.String(HnswRecordCacheExpiry, cf.Expiry.String()),
	}
}

type HealerFlags struct {
	MaxScanRatePerNode Uint32OptionalFlag
	MaxScanPageSize    Uint32OptionalFlag
	ReindexPercent     Float32OptionalFlag
	Schedule           StringOptionalFlag
	Parallelism        Uint32OptionalFlag
}

func NewHnswHealerFlags() *HealerFlags {
	return &HealerFlags{
		MaxScanRatePerNode: Uint32OptionalFlag{},
		MaxScanPageSize:    Uint32OptionalFlag{},
		ReindexPercent:     Float32OptionalFlag{},
		Schedule:           StringOptionalFlag{},
		Parallelism:        Uint32OptionalFlag{},
	}
}

func (cf *HealerFlags) NewFlagSet() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}
	flagSet.Var(&cf.MaxScanRatePerNode, HnswHealerMaxScanRatePerNode, "Maximum allowed record scan rate per AVS node. Defaults to 1_000.")                                                            //nolint:lll // For readability
	flagSet.Var(&cf.MaxScanPageSize, HnswHealerMaxScanPageSize, "Maximum number of records in a single scanned page. Defaults to 10_000.")                                                            //nolint:lll // For readability
	flagSet.Var(&cf.ReindexPercent, HnswHealerReindexPercent, "Percentage of good records randomly selected for reindexing in a healer cycle. Defaults to 10.")                                       //nolint:lll // For readability
	flagSet.Var(&cf.Schedule, HnswHealerSchedule, "The quartz cron expression defining the schedule at which the index healer cycle is invoked.. Defaults to '0 0/15 * ? * * *' (every 15 minutes).") //nolint:lll // For readability
	flagSet.Var(&cf.Parallelism, HnswHealerParallelism, "Maximum number of records to heal in parallel. Defaults to 1.")                                                                              //nolint:lll // For readability

	return flagSet
}

func (cf *HealerFlags) NewSLogAttr() []any {
	return []any{
		slog.Any(HnswHealerMaxScanRatePerNode, cf.MaxScanRatePerNode.String()),
		slog.Any(HnswHealerMaxScanPageSize, cf.MaxScanPageSize.String()),
		slog.Any(HnswHealerReindexPercent, cf.ReindexPercent.String()),
		slog.Any(HnswHealerSchedule, cf.Schedule.String()),
		slog.Any(HnswHealerParallelism, cf.Parallelism.String()),
	}
}

type MergeFlags struct {
	IndexParallelism   Uint32OptionalFlag
	ReIndexParallelism Uint32OptionalFlag
}

func NewHnswMergeFlags() *MergeFlags {
	return &MergeFlags{}
}

func (cf *MergeFlags) NewFlagSet() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}
	flagSet.Var(&cf.IndexParallelism, HnswMergeParallelism, "The number of vectors merged in parallel from a batch index to main index. Defaults to 10 * times the number of available CPU cores.")                                                  //nolint:lll // For readability
	flagSet.Var(&cf.ReIndexParallelism, HnswMergeReIndexParallelism, fmt.Sprintf("The number of vectors merged in parallel from a re-indexing record batch-index to the main index. Defaults to the maximum of 1 or %s / 3.", HnswMergeParallelism)) //nolint:lll // For readability

	return flagSet
}

func (cf *MergeFlags) NewSLogAttr() []any {
	return []any{
		slog.Any(HnswMergeParallelism, cf.IndexParallelism.Val),
		slog.Any(HnswMergeReIndexParallelism, cf.ReIndexParallelism.Val),
	}
}
