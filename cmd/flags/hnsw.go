package flags

import (
	"log/slog"

	commonFlags "github.com/aerospike/tools-common-go/flags"
	"github.com/spf13/pflag"
)

type BatchingFlags struct {
	MaxRecords Uint32OptionalFlag
	Interval   DurationOptionalFlag
}

func NewHnswBatchingFlags() *BatchingFlags {
	return &BatchingFlags{
		MaxRecords: Uint32OptionalFlag{},
		Interval:   DurationOptionalFlag{},
	}
}

func (cf *BatchingFlags) NewFlagSet() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}
	flagSet.Var(&cf.MaxRecords, BatchMaxRecords, commonFlags.DefaultWrapHelpString("Maximum number of records to fit in a batch."))              //nolint:lll // For readability
	flagSet.Var(&cf.Interval, BatchInterval, commonFlags.DefaultWrapHelpString("The maximum amount of time to wait before finalizing a batch.")) //nolint:lll // For readability

	return flagSet
}

func (cf *BatchingFlags) NewSLogAttr() []any {
	return []any{
		slog.Any(BatchMaxRecords, cf.MaxRecords.Val),
		slog.Any(BatchInterval, cf.Interval.Val),
	}
}

type CachingFlags struct {
	MaxEntries Uint64OptionalFlag
	Expiry     DurationOptionalFlag
}

func NewHnswCachingFlags() *CachingFlags {
	return &CachingFlags{
		MaxEntries: Uint64OptionalFlag{},
		Expiry:     DurationOptionalFlag{},
	}
}

func (cf *CachingFlags) NewFlagSet() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}
	flagSet.Var(&cf.MaxEntries, HnswCacheMaxEntries, commonFlags.DefaultWrapHelpString("Maximum number of entries to cache."))                                                        //nolint:lll // For readability
	flagSet.Var(&cf.Expiry, HnswCacheExpiry, commonFlags.DefaultWrapHelpString("A cache entry will expire after this amount of time has expired after the entry was added to cache")) //nolint:lll // For readability

	return flagSet
}

func (cf *CachingFlags) NewSLogAttr() []any {
	return []any{
		slog.Any(HnswCacheMaxEntries, cf.MaxEntries.Val),
		slog.Any(HnswCacheExpiry, cf.Expiry.Val),
	}
}

type HealerFlags struct {
	MaxScanRatePerNode Uint32OptionalFlag
	MaxScanPageSize    Uint32OptionalFlag
	ReindexPercent     Float32OptionalFlag
	ScheduleDelay      DurationOptionalFlag
	Parallelism        Uint32OptionalFlag
}

func NewHnswHealerFlags() *HealerFlags {
	return &HealerFlags{
		MaxScanRatePerNode: Uint32OptionalFlag{},
		MaxScanPageSize:    Uint32OptionalFlag{},
		ReindexPercent:     Float32OptionalFlag{},
		ScheduleDelay:      DurationOptionalFlag{},
		Parallelism:        Uint32OptionalFlag{},
	}
}

func (cf *HealerFlags) NewFlagSet() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}
	flagSet.Var(&cf.MaxScanRatePerNode, HnswHealerMaxScanRatePerNode, commonFlags.DefaultWrapHelpString("Maximum allowed record scan rate per AVS node."))                                                  //nolint:lll // For readability
	flagSet.Var(&cf.MaxScanPageSize, HnswHealerMaxScanPageSize, commonFlags.DefaultWrapHelpString("Maximum number of records in a single scanned page."))                                                   //nolint:lll // For readability
	flagSet.Var(&cf.ReindexPercent, HnswHealerReindexPercent, commonFlags.DefaultWrapHelpString("Percentage of good records randomly selected for reindexing in a healer cycle."))                          //nolint:lll // For readability
	flagSet.Var(&cf.ScheduleDelay, HnswHealerScheduleDelay, commonFlags.DefaultWrapHelpString("The time delay between the termination of a healer run and the commencement of the next one for an index.")) //nolint:lll // For readability
	flagSet.Var(&cf.Parallelism, HnswHealerParallelism, commonFlags.DefaultWrapHelpString("Maximum number of records to heal in parallel."))                                                                //nolint:lll // For readability

	return flagSet
}

func (cf *HealerFlags) NewSLogAttr() []any {
	return []any{
		slog.Any(HnswHealerMaxScanRatePerNode, cf.MaxScanRatePerNode.Val),
		slog.Any(HnswHealerMaxScanPageSize, cf.MaxScanPageSize.Val),
		slog.Any(HnswHealerReindexPercent, cf.ReindexPercent.Val),
		slog.Any(HnswHealerScheduleDelay, cf.ScheduleDelay.Val),
		slog.Any(HnswHealerParallelism, cf.Parallelism.Val),
	}
}

type MergeFlags struct {
	Parallelism Uint32OptionalFlag
}

func NewHnswMergeFlags() *MergeFlags {
	return &MergeFlags{
		Parallelism: Uint32OptionalFlag{},
	}
}

func (cf *MergeFlags) NewFlagSet() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}
	flagSet.Var(&cf.Parallelism, HnswMergeParallelism, commonFlags.DefaultWrapHelpString("The number of vectors merged in parallel from a batch index to main index.")) //nolint:lll // For readability

	return flagSet
}

func (cf *MergeFlags) NewSLogAttr() []any {
	return []any{
		slog.Any(HnswMergeParallelism, cf.Parallelism.Val),
	}
}
