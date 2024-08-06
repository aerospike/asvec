package flags

import "github.com/spf13/pflag"

const (
	LogLevel                     = "log-level"
	NoColor                      = "no-color"
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
	Sets                         = "sets"
	Yes                          = "yes"
	IndexName                    = "index-name"
	VectorField                  = "vector-field"
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
	MaxEdges                     = "hnsw-max-edges"
	ConstructionEf               = "hnsw-ef-construction"
	Ef                           = "hnsw-ef"
	HnswMaxMemQueueSize          = "hnsw-max-mem-queue-size"
	BatchMaxRecords              = "hnsw-batch-max-records"
	BatchInterval                = "hnsw-batch-interval"
	HnswCacheMaxEntries          = "hnsw-cache-max-entries"
	HnswCacheExpiry              = "hnsw-cache-expiry"
	HnswHealerMaxScanRatePerNode = "hnsw-healer-max-scan-rate-per-node"
	HnswHealerMaxScanPageSize    = "hnsw-healer-max-scan-page-size"
	HnswHealerReindexPercent     = "hnsw-healer-reindex-percent"
	HnswHealerScheduleDelay      = "hnsw-healer-schedule-delay"
	HnswHealerParallelism        = "hnsw-healer-parallelism"
	HnswMergeParallelism         = "hnsw-merge-parallelism"
	TLSProtocols                 = "tls-protocols"
	TLSCaFile                    = "tls-cafile"
	TLSCaPath                    = "tls-capath"
	TLSCertFile                  = "tls-certfile"
	TLSKeyFile                   = "tls-keyfile"
	TLSKeyFilePass               = "tls-keyfile-password" //nolint:gosec // Not a credential
)

func AddFormatTestFlag(flagSet *pflag.FlagSet, val *int) {
	flagSet.IntVar(val, Format, 0, "For testing only")
	flagSet.MarkHidden(Format)
}
