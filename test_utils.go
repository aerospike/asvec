//go:build integration

package main

import (
	"fmt"

	"github.com/aerospike/avs-client-go/protos"
)

func getStrPtr(str string) *string {
	ptr := str
	return &ptr
}

func getUint32Ptr(i int) *uint32 {
	ptr := uint32(i)
	return &ptr
}

func getBoolPtr(b bool) *bool {
	ptr := b
	return &ptr
}

func createFlagStr(name, value string) string {
	return fmt.Sprintf("--%s %s", name, value)
}

type IndexDefinitionBuilder struct {
	indexName             string
	namespace             string
	set                   *string
	dimension             int
	vectorDistanceMetric  protos.VectorDistanceMetric
	vectorField           string
	storageNamespace      *string
	storageSet            *string
	hnsfM                 *uint32
	hnsfEfC               *uint32
	hnsfEf                *uint32
	hnsfBatchingMaxRecord *uint32
	hnsfBatchingInterval  *uint32
	hnsfBatchingDisabled  *bool
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

func (idb *IndexDefinitionBuilder) WithHnswBatchingMaxRecord(maxRecord uint32) *IndexDefinitionBuilder {
	idb.hnsfBatchingMaxRecord = &maxRecord
	return idb
}

func (idb *IndexDefinitionBuilder) WithHnswBatchingInterval(interval uint32) *IndexDefinitionBuilder {
	idb.hnsfBatchingInterval = &interval
	return idb
}

func (idb *IndexDefinitionBuilder) WithHnswBatchingDisabled(disabled bool) *IndexDefinitionBuilder {
	idb.hnsfBatchingDisabled = &disabled
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
				M:              getUint32Ptr(16),
				EfConstruction: getUint32Ptr(100),
				Ef:             getUint32Ptr(100),
				BatchingParams: &protos.HnswBatchingParams{
					MaxRecords: getUint32Ptr(100000),
					Interval:   getUint32Ptr(30000),
					// Disabled:   getBoolPtr(false),
				},
			},
		},
	}

	if idb.set != nil {
		indexDef.SetFilter = idb.set
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
	// if idb.hnsfBatchingDisabled != nil {
	// 	indexDef.Params.(*protos.IndexDefinition_HnswParams).HnswParams.BatchingParams.Disabled = idb.hnsfBatchingDisabled
	// }

	return indexDef
}
