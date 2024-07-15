//go:build unit && integration

package tests

import (
	"fmt"

	"github.com/aerospike/avs-client-go/protos"
)

func GetStrPtr(str string) *string {
	ptr := str
	return &ptr
}

func GetUint32Ptr(i int) *uint32 {
	ptr := uint32(i)
	return &ptr
}

func GetBoolPtr(b bool) *bool {
	ptr := b
	return &ptr
}

func CreateFlagStr(name, value string) string {
	return fmt.Sprintf("--%s %s", name, value)
}

type IndexDefinitionBuilder struct {
	indexName                      string
	namespace                      string
	set                            *string
	dimension                      int
	vectorDistanceMetric           protos.VectorDistanceMetric
	vectorField                    string
	storageNamespace               *string
	storageSet                     *string
	labels                         map[string]string
	hnsfM                          *uint32
	hnsfEfC                        *uint32
	hnsfEf                         *uint32
	hnswMemQueueSize               *uint32
	hnsfBatchingMaxRecord          *uint32
	hnsfBatchingInterval           *uint32
	hnswCacheExpiry                *uint64
	hnswCacheMaxEntries            *uint64
	hnswHealerMaxScanPageSize      *uint32
	hnswHealerMaxScanRatePerSecond *uint32
	hnswHealerParallelism          *uint32
	HnswHealerReindexPercent       *float32
	HnswHealerScheduleDelay        *uint64
	hnswMergeParallelism           *uint32
}

func NewIndexDefinitionBuilder(
	indexName,
	namespace string,
	dimension int,
	distanceMetric protos.VectorDistanceMetric,
	vectorField string,
) *IndexDefinitionBuilder {
	return &IndexDefinitionBuilder{
		indexName:            indexName,
		namespace:            namespace,
		dimension:            dimension,
		vectorDistanceMetric: distanceMetric,
		vectorField:          vectorField,
	}
}

func (idb *IndexDefinitionBuilder) WithSet(set string) *IndexDefinitionBuilder {
	idb.set = &set
	return idb
}

func (idb *IndexDefinitionBuilder) WithLabels(labels map[string]string) *IndexDefinitionBuilder {
	idb.labels = labels
	return idb
}

func (idb *IndexDefinitionBuilder) WithStorageNamespace(storageNamespace string) *IndexDefinitionBuilder {
	idb.storageNamespace = &storageNamespace
	return idb
}

func (idb *IndexDefinitionBuilder) WithStorageSet(storageSet string) *IndexDefinitionBuilder {
	idb.storageSet = &storageSet
	return idb
}

func (idb *IndexDefinitionBuilder) WithHnswM(m uint32) *IndexDefinitionBuilder {
	idb.hnsfM = &m
	return idb
}

func (idb *IndexDefinitionBuilder) WithHnswEf(ef uint32) *IndexDefinitionBuilder {
	idb.hnsfEf = &ef
	return idb
}

func (idb *IndexDefinitionBuilder) WithHnswEfConstruction(efConstruction uint32) *IndexDefinitionBuilder {
	idb.hnsfEfC = &efConstruction
	return idb
}

func (idb *IndexDefinitionBuilder) WithHnswMaxMemQueueSize(maxMemQueueSize uint32) *IndexDefinitionBuilder {
	idb.hnswMemQueueSize = &maxMemQueueSize
	return idb
}

func (idb *IndexDefinitionBuilder) WithHnswBatchingMaxRecord(maxRecord uint32) *IndexDefinitionBuilder {
	idb.hnsfBatchingMaxRecord = &maxRecord
	return idb
}

func (idb *IndexDefinitionBuilder) WithHnswBatchingInterval(interval uint32) *IndexDefinitionBuilder {
	idb.hnsfBatchingInterval = &interval
	return idb
}

func (idb *IndexDefinitionBuilder) WithHnswCacheExpiry(expiry uint64) *IndexDefinitionBuilder {
	idb.hnswCacheExpiry = &expiry
	return idb
}

func (idb *IndexDefinitionBuilder) WithHnswCacheMaxEntries(maxEntries uint64) *IndexDefinitionBuilder {
	idb.hnswCacheMaxEntries = &maxEntries
	return idb
}

func (idb *IndexDefinitionBuilder) WithHnswHealerMaxScanPageSize(maxScanPageSize uint32) *IndexDefinitionBuilder {
	idb.hnswHealerMaxScanPageSize = &maxScanPageSize
	return idb
}

func (idb *IndexDefinitionBuilder) WithHnswHealerMaxScanRatePerNode(maxScanRatePerSecond uint32) *IndexDefinitionBuilder {
	idb.hnswHealerMaxScanRatePerSecond = &maxScanRatePerSecond
	return idb
}

func (idb *IndexDefinitionBuilder) WithHnswHealerParallelism(parallelism uint32) *IndexDefinitionBuilder {
	idb.hnswHealerParallelism = &parallelism
	return idb
}

func (idb *IndexDefinitionBuilder) WithHnswHealerReindexPercent(reindexPercent float32) *IndexDefinitionBuilder {
	idb.HnswHealerReindexPercent = &reindexPercent
	return idb
}

func (idb *IndexDefinitionBuilder) WithHnswHealerScheduleDelay(scheduleDelay uint64) *IndexDefinitionBuilder {
	idb.HnswHealerScheduleDelay = &scheduleDelay
	return idb
}

func (idb *IndexDefinitionBuilder) WithHnswMergeParallelism(mergeParallelism uint32) *IndexDefinitionBuilder {
	idb.hnswMergeParallelism = &mergeParallelism
	return idb
}

func (idb *IndexDefinitionBuilder) Build() *protos.IndexDefinition {
	indexDef := &protos.IndexDefinition{
		Id: &protos.IndexId{
			Name:      idb.indexName,
			Namespace: idb.namespace,
		},
		Dimensions:           uint32(idb.dimension),
		VectorDistanceMetric: idb.vectorDistanceMetric,
		Field:                idb.vectorField,
		Type:                 protos.IndexType_HNSW,
		Storage: &protos.IndexStorage{
			Namespace: &idb.namespace,
			Set:       &idb.indexName,
		},
		Params: &protos.IndexDefinition_HnswParams{
			HnswParams: &protos.HnswParams{
				M:              GetUint32Ptr(16),
				EfConstruction: GetUint32Ptr(100),
				Ef:             GetUint32Ptr(100),
				BatchingParams: &protos.HnswBatchingParams{
					MaxRecords: GetUint32Ptr(100000),
					Interval:   GetUint32Ptr(30000),
				},
				CachingParams: &protos.HnswCachingParams{},
				HealerParams:  &protos.HnswHealerParams{},
				MergeParams:   &protos.HnswIndexMergeParams{},
			},
		},
	}

	indexDef.SetFilter = idb.set

	if idb.labels != nil {
		indexDef.Labels = idb.labels
	}

	if idb.storageNamespace != nil {
		indexDef.Storage.Namespace = idb.storageNamespace
	}

	if idb.storageSet != nil {
		indexDef.Storage.Set = idb.storageSet
	}

	if idb.hnsfM != nil {
		indexDef.Params.(*protos.IndexDefinition_HnswParams).HnswParams.M = idb.hnsfM
	}

	if idb.hnsfEf != nil {
		indexDef.Params.(*protos.IndexDefinition_HnswParams).HnswParams.Ef = idb.hnsfEf
	}

	if idb.hnsfEfC != nil {
		indexDef.Params.(*protos.IndexDefinition_HnswParams).HnswParams.EfConstruction = idb.hnsfEfC
	}

	if idb.hnsfBatchingMaxRecord != nil {
		indexDef.Params.(*protos.IndexDefinition_HnswParams).HnswParams.BatchingParams.MaxRecords = idb.hnsfBatchingMaxRecord
	}

	if idb.hnsfBatchingInterval != nil {
		indexDef.Params.(*protos.IndexDefinition_HnswParams).HnswParams.BatchingParams.Interval = idb.hnsfBatchingInterval
	}

	if idb.hnswMemQueueSize != nil {
		indexDef.Params.(*protos.IndexDefinition_HnswParams).HnswParams.MaxMemQueueSize = idb.hnswMemQueueSize
	}

	if idb.hnswCacheExpiry != nil {
		indexDef.Params.(*protos.IndexDefinition_HnswParams).HnswParams.CachingParams.Expiry = idb.hnswCacheExpiry
	}

	if idb.hnswCacheMaxEntries != nil {
		indexDef.Params.(*protos.IndexDefinition_HnswParams).HnswParams.CachingParams.MaxEntries = idb.hnswCacheMaxEntries
	}

	if idb.hnswHealerMaxScanPageSize != nil {
		indexDef.Params.(*protos.IndexDefinition_HnswParams).HnswParams.HealerParams.MaxScanPageSize = idb.hnswHealerMaxScanPageSize
	}

	if idb.hnswHealerMaxScanRatePerSecond != nil {
		indexDef.Params.(*protos.IndexDefinition_HnswParams).HnswParams.HealerParams.MaxScanRatePerNode = idb.hnswHealerMaxScanRatePerSecond
	}

	if idb.hnswHealerParallelism != nil {
		indexDef.Params.(*protos.IndexDefinition_HnswParams).HnswParams.HealerParams.Parallelism = idb.hnswHealerParallelism
	}

	if idb.HnswHealerReindexPercent != nil {
		indexDef.Params.(*protos.IndexDefinition_HnswParams).HnswParams.HealerParams.ReindexPercent = idb.HnswHealerReindexPercent
	}

	if idb.HnswHealerScheduleDelay != nil {
		indexDef.Params.(*protos.IndexDefinition_HnswParams).HnswParams.HealerParams.ScheduleDelay = idb.HnswHealerScheduleDelay
	}

	if idb.hnswMergeParallelism != nil {
		indexDef.Params.(*protos.IndexDefinition_HnswParams).HnswParams.MergeParams.Parallelism = idb.hnswMergeParallelism
	}

	return indexDef
}
