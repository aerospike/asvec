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

	// Default values
	DefaultLogLevel                     = "DISABLED"
	DefaultNoColor                      = false
	DefaultClusterName                  = ""
	DefaultConfigFile                   = "asvec.yml"
	DefaultSeeds                        = ""
	DefaultHost                         = ""
	DefaultListenerName                 = ""
	DefaultAuthUser                     = ""
	DefaultAuthPassword                 = ""
	DefaultAuthCredentials              = ""
	DefaultName                         = ""
	DefaultNewPassword                  = ""
	DefaultRoles                        = ""
	DefaultNamespace                    = ""
	DefaultSet                          = ""
	DefaultYes                          = false
	DefaultIndexName                    = ""
	DefaultVectorField                  = ""
	DefaultVector                       = ""
	DefaultKeyString                    = ""
	DefaultKeyInt                       = 0
	DefaultFields                       = ""
	DefaultMaxResults                   = 10
	DefaultMaxDataKeys                  = 100
	DefaultMaxDataColWidth              = 80
	DefaultDimension                    = 128
	DefaultDistanceMetric               = "cosine"
	DefaultIndexLabels                  = false
	DefaultTimeout                      = 30
	DefaultVerbose                      = false
	DefaultFormat                       = "json"
	DefaultYaml                         = false
	DefaultInputFile                    = ""
	DefaultStorageNamespace             = ""
	DefaultStorageSet                   = ""
	DefaultCutoffTime                   = ""
	DefaultHnswMaxEdges                 = 16
	DefaultHnswConstructionEf           = 200
	DefaultHnswEf                       = 10
	DefaultHnswMaxMemQueueSize          = 1000
	DefaultBatchMaxIndexRecords         = 10000
	DefaultBatchIndexInterval           = 60
	DefaultBatchMaxReindexRecords       = 10000
	DefaultBatchReindexInterval         = 60
	DefaultHnswIndexCacheMaxEntries     = 100000
	DefaultHnswIndexCacheExpiry         = 3600
	DefaultHnswRecordCacheMaxEntries    = 100000
	DefaultHnswRecordCacheExpiry        = 3600
	DefaultHnswHealerMaxScanRatePerNode = 1000
	DefaultHnswHealerMaxScanPageSize    = 100
	DefaultHnswHealerReindexPercent     = 10
	DefaultHnswHealerSchedule           = "0 0 * * *"
	DefaultHnswHealerParallelism        = 4
	DefaultHnswMergeParallelism         = 4
	DefaultHnswMergeReIndexParallelism  = 4
	DefaultHnswVectorIntegrityCheck     = false
	DefaultIndexMode                    = "sync"
	DefaultTLSProtocols                 = "TLSv1.2,TLSv1.3"
	DefaultTLSCaFile                    = ""
	DefaultTLSCaPath                    = ""
	DefaultTLSCertFile                  = ""
	DefaultTLSKeyFile                   = ""
	DefaultTLSKeyFilePass               = ""
	DefaultTLSHostnameOverride          = ""

	DefaultIPv4 = "127.0.0.1"
	DefaultPort = 5000

	Infinity = -1
)

func AddFormatTestFlag(flagSet *pflag.FlagSet, val *int) error {
	flagSet.IntVar(val, Format, 0, "For testing only")
	return flagSet.MarkHidden(Format)
}
