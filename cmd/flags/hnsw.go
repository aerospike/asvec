package flags

import (
	"log/slog"

	commonFlags "github.com/aerospike/tools-common-go/flags"
	"github.com/spf13/pflag"
)

type BatchingFlags struct {
	MaxRecords Uint32OptionalFlag
	Interval   Uint32OptionalFlag
}

func NewBatchingFlags() *BatchingFlags {
	return &BatchingFlags{
		MaxRecords: Uint32OptionalFlag{},
		Interval:   Uint32OptionalFlag{},
	}
}

func (cf *BatchingFlags) NewFlagSet() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}
	flagSet.Var(&cf.MaxRecords, BatchMaxRecords, commonFlags.DefaultWrapHelpString("Maximum number of records to fit in a batch. The default value is 10000."))                              //nolint:lll // For readability
	flagSet.Var(&cf.Interval, BatchInterval, commonFlags.DefaultWrapHelpString("The maximum amount of time in milliseconds to wait before finalizing a batch. The default value is 10000.")) //nolint:lll // For readability

	return flagSet
}

func (cf *BatchingFlags) NewSLogAttr() []any {
	return []any{
		slog.Any(BatchMaxRecords, cf.MaxRecords.Val),
		slog.Any(BatchInterval, cf.Interval.Val),
	}
}

type CachingFlags struct {
	maxEntries Uint64OptionalFlag
	expiry     Uint64OptionalFlag
}

func NewCachingFlags() *CachingFlags {
	return &CachingFlags{
		maxEntries: Uint64OptionalFlag{},
		expiry:     Uint64OptionalFlag{},
	}
}

func (cf *CachingFlags) NewFlagSet() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}                                                                 //nolint:lll // For readability
	flagSet.Var(&cf.maxEntries, HnswCacheMaxEntries, commonFlags.DefaultWrapHelpString("TODO")) //nolint:lll // For readability
	flagSet.Var(&cf.expiry, HnswCacheExpiry, commonFlags.DefaultWrapHelpString("TODO"))         //nolint:lll // For readability

	return flagSet
}

func (cf *CachingFlags) NewSLogAttr() []any {
	return []any{
		slog.Any(HnswCacheMaxEntries, cf.maxEntries.Val),
		slog.Any(HnswCacheExpiry, cf.expiry.Val),
	}
}

type HealerFlags struct {
	maxScanRatePerNode Uint32OptionalFlag
	maxScanPageSize    Uint32OptionalFlag
	reindexPercent     Float32OptionalFlag
	scheduleDelay      Uint64OptionalFlag
	parallelism        Uint32OptionalFlag
}

func NewHealerFlags() *HealerFlags {
	return &HealerFlags{
		maxScanRatePerNode: Uint32OptionalFlag{},
		maxScanPageSize:    Uint32OptionalFlag{},
		reindexPercent:     Float32OptionalFlag{},
		scheduleDelay:      Uint64OptionalFlag{},
		parallelism:        Uint32OptionalFlag{},
	}
}

func (cf *HealerFlags) NewFlagSet() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}                                                                                  //nolint:lll // For readability
	flagSet.Var(&cf.maxScanRatePerNode, HnswHealerMaxScanRatePerNode, commonFlags.DefaultWrapHelpString("TODO")) //nolint:lll // For readability
	flagSet.Var(&cf.maxScanPageSize, HnswHealerMaxScanPageSize, commonFlags.DefaultWrapHelpString("TODO"))       //nolint:lll // For readability
	flagSet.Var(&cf.reindexPercent, HnswHealerReindexPercent, commonFlags.DefaultWrapHelpString("TODO"))         //nolint:lll // For readability
	flagSet.Var(&cf.scheduleDelay, HnswHealerScheduleDelay, commonFlags.DefaultWrapHelpString("TODO"))           //nolint:lll // For readability
	flagSet.Var(&cf.parallelism, HnswHealerParallelism, commonFlags.DefaultWrapHelpString("TODO"))               //nolint:lll // For readability

	return flagSet
}

func (cf *HealerFlags) NewSLogAttr() []any {
	return []any{
		slog.Any(HnswHealerMaxScanRatePerNode, cf.maxScanRatePerNode.Val),
		slog.Any(HnswHealerMaxScanPageSize, cf.maxScanPageSize.Val),
		slog.Any(HnswHealerReindexPercent, cf.reindexPercent.Val),
		slog.Any(HnswHealerScheduleDelay, cf.scheduleDelay.Val),
		slog.Any(HnswHealerParallelism, cf.parallelism.Val),
	}
}

type MergeFlags struct {
	parallelism Uint32OptionalFlag
}

func NewMergeFlags() *MergeFlags {
	return &MergeFlags{
		parallelism: Uint32OptionalFlag{},
	}
}

func (cf *MergeFlags) NewFlagSet() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}                                                                   //nolint:lll // For readability
	flagSet.Var(&cf.parallelism, HnswMergeParallelism, commonFlags.DefaultWrapHelpString("TODO")) //nolint:lll // For readability

	return flagSet
}

func (cf *MergeFlags) NewSLogAttr() []any {
	return []any{
		slog.Any(HnswMergeParallelism, cf.parallelism.Val),
	}
}
