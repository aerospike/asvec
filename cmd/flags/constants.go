package flags

import "github.com/spf13/pflag"

const (
	LogLevel                     = "log-level"
	NoColor                      = "no-color"
	ClusterName                  = "cluster-name"
	ConfigFile                   = "config-file"
	Seeds                        = "seeds"
	Host                         = "host"
	ListenerName                 = "listener-name"
	AuthUser                     = "user"
	AuthPassword                 = "password"
	AuthCredentials              = "credentials"
	Name                         = "name"
	NewPassword                  = "new-password"
	Roles                        = "roles"
	Namespace                    = "namespace"
	Set                          = "set"
	Yes                          = "yes"
	IndexName                    = "index-name"
	VectorField                  = "vector-field"
	Vector                       = "vector"
	KeyString                    = "key-str"
	KeyInt                       = "key-int"
	Fields                       = "fields"
	MaxResults                   = "max-results"
	MaxDataKeys                  = "max-keys"
	MaxDataColWidth              = "max-width"
	Dimension                    = "dimension"
	DistanceMetric               = "distance-metric"
	IndexLabels                  = "index-labels"
	Timeout                      = "timeout"
	Verbose                      = "verbose"
	Format                       = "format"
	Yaml                         = "yaml"
	InputFile                    = "file"
	StorageNamespace             = "storage-namespace"
	StorageSet                   = "storage-set"
	CutoffTime                   = "cutoff-time"
	HnswMaxEdges                 = "hnsw-m"
	HnswConstructionEf           = "hnsw-ef-construction"
	HnswEf                       = "hnsw-ef"
	HnswMaxMemQueueSize          = "hnsw-max-mem-queue-size"
	BatchMaxIndexRecords         = "hnsw-batch-max-index-records"
	BatchIndexInterval           = "hnsw-batch-index-interval"
	BatchMaxReindexRecords       = "hnsw-batch-max-reindex-records"
	BatchReindexInterval         = "hnsw-batch-reindex-interval"
	HnswIndexCacheMaxEntries     = "hnsw-index-cache-max-entries"
	HnswIndexCacheExpiry         = "hnsw-index-cache-expiry"
	HnswRecordCacheMaxEntries    = "hnsw-record-cache-max-entries"
	HnswRecordCacheExpiry        = "hnsw-record-cache-expiry"
	HnswHealerMaxScanRatePerNode = "hnsw-healer-max-scan-rate-per-node"
	HnswHealerMaxScanPageSize    = "hnsw-healer-max-scan-page-size"
	HnswHealerReindexPercent     = "hnsw-healer-reindex-percent"
	HnswHealerSchedule           = "hnsw-healer-schedule"
	HnswHealerParallelism        = "hnsw-healer-parallelism"
	HnswMergeParallelism         = "hnsw-merge-index-parallelism"
	HnswMergeReIndexParallelism  = "hnsw-merge-reindex-parallelism"
	HnswVectorIntegrityCheck     = "hnsw-vector-integrity-check"
	IndexMode                    = "index-mode"
	TLSProtocols                 = "tls-protocols"
	TLSCaFile                    = "tls-cafile"
	TLSCaPath                    = "tls-capath"
	TLSCertFile                  = "tls-certfile"
	TLSKeyFile                   = "tls-keyfile"
	TLSKeyFilePass               = "tls-keyfile-password" //nolint:gosec // Not a credential
	TLSHostnameOverride          = "tls-hostname-override"
	Watch                        = "watch"
	WatchInterval                = "watch-interval"

	// TODO  Replace short flag constants with variables
	DimensionShort       = "d"
	VectorFieldShort     = "f"
	DistanceMetricShort  = "m"
	NamespaceShort       = "n"
	SetShort             = "s"
	IndexNameShort       = "i"
	VectorShort          = "v"
	KeyStrShort          = "k"
	KeyIntShort          = "t"
	MaxDataColWidthShort = "w"
	YesShort             = "y"

	// Flag types
	FlagTypeEnum = "enum"

	DefaultIPv4          = "127.0.0.1"
	DefaultPort          = 5000
	DefaultWatchInterval = 2 // Default watch interval in seconds

	Infinity = -1
)

func AddFormatTestFlag(flagSet *pflag.FlagSet, val *int) error {
	flagSet.IntVar(val, Format, 0, "For testing only")
	return flagSet.MarkHidden(Format)
}
