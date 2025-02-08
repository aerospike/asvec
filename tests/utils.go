//go:build unit || integration || integration_large

package tests

import (
	"asvec/utils"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"time"

	avs "github.com/aerospike/avs-client-go"
	"github.com/aerospike/avs-client-go/protos"
	"github.com/aerospike/tools-common-go/client"
)

func CreateFlagStr(name, value string) string {
	return fmt.Sprintf("--%s %s", name, value)
}

type IndexDefinitionBuilder struct {
	updateCmd                      bool
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
	hnswBatchingMaxRecord          *uint32
	hnswBatchingInterval           *uint32
	hnswBatchingMaxReindexRecord   *uint32
	hnswBatchingReindexInterval    *uint32
	hnswIndexCacheExpiry           *int64
	hnswIndexCacheMaxEntries       *uint64
	hnswRecordCacheExpiry          *int64
	hnswRecordCacheMaxEntries      *uint64
	hnswHealerMaxScanPageSize      *uint32
	hnswHealerMaxScanRatePerSecond *uint32
	hnswHealerParallelism          *uint32
	HnswHealerReindexPercent       *float32
	HnswHealerSchedule             *string
	hnswMergeIndexParallelism      *uint32
	hnswMergeReIndexParallelism    *uint32
	hnswVectorIntegrityCheck       *bool
	mode                           *protos.IndexMode
}

func NewIndexDefinitionBuilder(
	updateCmd bool,
	indexName,
	namespace string,
	dimension int,
	distanceMetric protos.VectorDistanceMetric,
	vectorField string,
) *IndexDefinitionBuilder {
	return &IndexDefinitionBuilder{
		updateCmd:            updateCmd,
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

func (idb *IndexDefinitionBuilder) WithHnswBatchingMaxIndexRecord(maxRecord uint32) *IndexDefinitionBuilder {
	idb.hnswBatchingMaxRecord = &maxRecord
	return idb
}

func (idb *IndexDefinitionBuilder) WithHnswBatchingIndexInterval(interval uint32) *IndexDefinitionBuilder {
	idb.hnswBatchingInterval = &interval
	return idb
}

func (idb *IndexDefinitionBuilder) WithHnswBatchingMaxReindexRecord(maxRecord uint32) *IndexDefinitionBuilder {
	idb.hnswBatchingMaxReindexRecord = &maxRecord
	return idb
}

func (idb *IndexDefinitionBuilder) WithHnswBatchingReindexInterval(interval uint32) *IndexDefinitionBuilder {
	idb.hnswBatchingReindexInterval = &interval
	return idb
}

func (idb *IndexDefinitionBuilder) WithHnswIndexCacheExpiry(expiry int64) *IndexDefinitionBuilder {
	idb.hnswIndexCacheExpiry = &expiry
	return idb
}

func (idb *IndexDefinitionBuilder) WithHnswIndexCacheMaxEntries(maxEntries uint64) *IndexDefinitionBuilder {
	idb.hnswIndexCacheMaxEntries = &maxEntries
	return idb
}

func (idb *IndexDefinitionBuilder) WithHnswRecordCacheExpiry(expiry int64) *IndexDefinitionBuilder {
	idb.hnswRecordCacheExpiry = &expiry
	return idb
}

func (idb *IndexDefinitionBuilder) WithHnswRecordCacheMaxEntries(maxEntries uint64) *IndexDefinitionBuilder {
	idb.hnswRecordCacheMaxEntries = &maxEntries
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

func (idb *IndexDefinitionBuilder) WithHnswHealerSchedule(schedule string) *IndexDefinitionBuilder {
	idb.HnswHealerSchedule = &schedule
	return idb
}

func (idb *IndexDefinitionBuilder) WithHnswMergeIndexParallelism(mergeParallelism uint32) *IndexDefinitionBuilder {
	idb.hnswMergeIndexParallelism = &mergeParallelism
	return idb
}

func (idb *IndexDefinitionBuilder) WithHnswMergeReIndexParallelism(mergeParallelism uint32) *IndexDefinitionBuilder {
	idb.hnswMergeReIndexParallelism = &mergeParallelism
	return idb
}

func (idb *IndexDefinitionBuilder) WithHnswVectorIntegrityCheck(enableVectorIntegrityCheck bool) *IndexDefinitionBuilder {
	idb.hnswVectorIntegrityCheck = &enableVectorIntegrityCheck
	return idb
}

func (idb *IndexDefinitionBuilder) WithIndexMode(indexMode protos.IndexMode) *IndexDefinitionBuilder {
	idb.mode = &indexMode
	return idb
}

func (idb *IndexDefinitionBuilder) Build() *protos.IndexDefinition {
	var indexDef *protos.IndexDefinition

	if idb.updateCmd {
		indexDef = &protos.IndexDefinition{
			Id: &protos.IndexId{
				Name:      idb.indexName,
				Namespace: idb.namespace,
			},
			Dimensions:           uint32(idb.dimension),
			VectorDistanceMetric: utils.Ptr(idb.vectorDistanceMetric),
			Field:                idb.vectorField,
			// Storage:              ,
			Params: &protos.IndexDefinition_HnswParams{
				HnswParams: &protos.HnswParams{
					// BatchingParams: &protos.HnswBatchingParams{},
					IndexCachingParams:  &protos.HnswCachingParams{},
					RecordCachingParams: &protos.HnswCachingParams{},
					HealerParams:        &protos.HnswHealerParams{},
					MergeParams:         &protos.HnswIndexMergeParams{},
				},
			},
			Mode: idb.mode,
		}
	} else {
		indexDef = &protos.IndexDefinition{
			Id: &protos.IndexId{
				Name:      idb.indexName,
				Namespace: idb.namespace,
			},
			Dimensions:           uint32(idb.dimension),
			VectorDistanceMetric: utils.Ptr(idb.vectorDistanceMetric),
			Field:                idb.vectorField,
			Storage:              &protos.IndexStorage{},
			Params: &protos.IndexDefinition_HnswParams{
				HnswParams: &protos.HnswParams{
					BatchingParams:      &protos.HnswBatchingParams{},
					IndexCachingParams:  &protos.HnswCachingParams{},
					RecordCachingParams: &protos.HnswCachingParams{},
					HealerParams:        &protos.HnswHealerParams{},
					MergeParams:         &protos.HnswIndexMergeParams{},
				},
			},
			Mode: idb.mode,
		}
	}

	indexDef.SetFilter = idb.set

	if idb.labels != nil {
		indexDef.Labels = idb.labels
	}

	if idb.storageNamespace != nil || idb.storageSet != nil {
		indexDef.Storage = &protos.IndexStorage{
			Namespace: idb.storageNamespace,
			Set:       idb.storageSet,
		}
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

	if idb.hnswBatchingInterval != nil || idb.hnswBatchingMaxRecord != nil || idb.hnswBatchingMaxReindexRecord != nil || idb.hnswBatchingReindexInterval != nil {
		indexDef.Params.(*protos.IndexDefinition_HnswParams).HnswParams.BatchingParams = &protos.HnswBatchingParams{}
	}

	if idb.hnswBatchingMaxRecord != nil {
		indexDef.Params.(*protos.IndexDefinition_HnswParams).HnswParams.BatchingParams.MaxIndexRecords = idb.hnswBatchingMaxRecord
	}

	if idb.hnswBatchingInterval != nil {
		indexDef.Params.(*protos.IndexDefinition_HnswParams).HnswParams.BatchingParams.IndexInterval = idb.hnswBatchingInterval
	}

	if idb.hnswBatchingMaxReindexRecord != nil {
		indexDef.Params.(*protos.IndexDefinition_HnswParams).HnswParams.BatchingParams.MaxReindexRecords = idb.hnswBatchingMaxReindexRecord
	}

	if idb.hnswBatchingReindexInterval != nil {
		indexDef.Params.(*protos.IndexDefinition_HnswParams).HnswParams.BatchingParams.ReindexInterval = idb.hnswBatchingReindexInterval
	}

	if idb.hnswVectorIntegrityCheck != nil {
		indexDef.Params.(*protos.IndexDefinition_HnswParams).HnswParams.EnableVectorIntegrityCheck = idb.hnswVectorIntegrityCheck
	}

	if idb.hnswMemQueueSize != nil {
		indexDef.Params.(*protos.IndexDefinition_HnswParams).HnswParams.MaxMemQueueSize = idb.hnswMemQueueSize
	}

	if idb.hnswIndexCacheExpiry != nil {
		indexDef.Params.(*protos.IndexDefinition_HnswParams).HnswParams.IndexCachingParams.Expiry = idb.hnswIndexCacheExpiry
	}

	if idb.hnswIndexCacheMaxEntries != nil {
		indexDef.Params.(*protos.IndexDefinition_HnswParams).HnswParams.IndexCachingParams.MaxEntries = idb.hnswIndexCacheMaxEntries
	}

	if idb.hnswRecordCacheExpiry != nil {
		indexDef.Params.(*protos.IndexDefinition_HnswParams).HnswParams.RecordCachingParams.Expiry = idb.hnswRecordCacheExpiry
	}

	if idb.hnswRecordCacheMaxEntries != nil {
		indexDef.Params.(*protos.IndexDefinition_HnswParams).HnswParams.RecordCachingParams.MaxEntries = idb.hnswRecordCacheMaxEntries
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

	if idb.HnswHealerSchedule != nil {
		indexDef.Params.(*protos.IndexDefinition_HnswParams).HnswParams.HealerParams.Schedule = idb.HnswHealerSchedule
	}

	if idb.hnswMergeIndexParallelism != nil {
		indexDef.Params.(*protos.IndexDefinition_HnswParams).HnswParams.MergeParams.IndexParallelism = idb.hnswMergeIndexParallelism
	}

	if idb.hnswMergeReIndexParallelism != nil {
		indexDef.Params.(*protos.IndexDefinition_HnswParams).HnswParams.MergeParams.ReIndexParallelism = idb.hnswMergeReIndexParallelism
	}

	return indexDef
}

func DockerComposeUp(composeFile string) error {
	fmt.Println("Starting docker containers " + composeFile)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	cmd := exec.CommandContext(ctx, "docker", "-lDEBUG", "compose", fmt.Sprintf("-f%s", composeFile), "up", "-d")
	err := cmd.Run()
	cmd.Wait()

	// fmt.Printf("docker compose up output: %s\n", string())

	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			return err
		}
		return err
	}

	return nil
}

func DockerComposeDown(composeFile string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	cmd := exec.CommandContext(ctx, "docker", "compose", fmt.Sprintf("-f%s", composeFile), "down")
	_, err := cmd.Output()

	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			return err
		}
		return err
	}

	return nil
}

func GetClient(
	avsHostPort *avs.HostPort,
	avsCreds *avs.UserPassCredentials,
	avsTLSConfig *tls.Config,
	logger *slog.Logger,
) (*avs.Client, error) {
	// Connect avs client
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*50)
	defer cancel()

	var (
		avsClient *avs.Client
		err       error
	)

	for {
		avsClient, err = avs.NewClient(
			ctx,
			avs.HostPortSlice{avsHostPort},
			nil,
			true,
			avsCreds,
			avsTLSConfig,
			logger,
		)

		if err == nil {
			break
		}

		fmt.Printf("unable to create avs client %v", err)

		if ctx.Err() != nil {
			return nil, err
		}

		time.Sleep(time.Second)
	}

	// Wait for cluster to be ready
	for {
		_, err := avsClient.IndexList(ctx, false)
		if err == nil {
			break
		}

		fmt.Printf("waiting for the cluster to be ready %v", err)

		if ctx.Err() != nil {
			return nil, err
		}

		time.Sleep(time.Second)
	}

	return avsClient, nil
}

func GetCACert(cert string) (*x509.CertPool, error) {
	// read in file
	certBytes, err := os.ReadFile(cert)
	if err != nil {
		log.Fatalf("unable to read cert file %v", err)
		return nil, err
	}

	return client.LoadCACerts([][]byte{certBytes}), nil
}

func GetCertificates(certFile string, keyFile string) ([]tls.Certificate, error) {
	cert, err := os.ReadFile(certFile)
	if err != nil {
		log.Fatalf("unable to read cert file %v", err)
		return nil, err
	}

	key, err := os.ReadFile(keyFile)
	if err != nil {
		log.Fatalf("unable to read key file %v", err)
		return nil, err
	}

	return client.LoadServerCertAndKey([]byte(cert), []byte(key), nil)
}

var quoted = false

// Used with strings.FieldsFunc. Obviously not thread save.
func SplitQuotedString(r rune) bool {
	if r == '"' {
		quoted = !quoted
	}

	return (r == ' ' && !quoted) || r == '"'
}
