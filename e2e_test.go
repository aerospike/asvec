//go:build integration

package main_test

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"
	"time"

	avs "github.com/aerospike/aerospike-proximus-client-go"
	"github.com/aerospike/aerospike-proximus-client-go/protos"
	"github.com/stretchr/testify/suite"
)

var wd, _ = os.Getwd()
var logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))

var (
	testNamespace = "test"
	testSet       = "testset"
	barNamespace  = "bar"
)

type CmdTestSuite struct {
	suite.Suite
	app              string
	coverFile        string
	coverFileCounter int
	avsIP            string
	avsPort          int
	avsHostPort      *avs.HostPort
	avsClient        *avs.AdminClient
}

func TestDistanceMetricFlagSuite(t *testing.T) {
	suite.Run(t, new(CmdTestSuite))
}

func (suite *CmdTestSuite) SetupSuite() {
	suite.app = path.Join(wd, "app.test")
	suite.coverFile = path.Join(wd, "../coverage/cmd-coverage.cov")
	suite.coverFileCounter = 0
	suite.avsIP = "127.0.0.1"
	suite.avsPort = 10000
	suite.avsHostPort = avs.NewHostPort(suite.avsIP, suite.avsPort, false)
	// var err error

	err := docker_compose_up()
	if err != nil {
		suite.FailNowf("unable to start docker compose up", "%v", err)
	}

	goArgs := []string{"build", "-cover", "-coverpkg", "./...", "-o", suite.app}

	// Compile test binary
	compileCmd := exec.Command("go", goArgs...)
	out, err := compileCmd.CombinedOutput()

	if err != nil {
		fmt.Printf("Couldn't compile test bin stdout/err: %v\n", string(out))
	}

	suite.Assert().NoError(err)

	// Connect avs client
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	for {
		suite.avsClient, err = avs.NewAdminClient(ctx, avs.HostPortSlice{suite.avsHostPort}, nil, true, logger)

		if err != nil {
			fmt.Printf("unable to create avs client %v", err)

			if ctx.Err() != nil {
				suite.FailNowf("unable to create avs client", "%v", err)
			}

			time.Sleep(time.Second)
		} else {
			break
		}
	}

}

func (suite *CmdTestSuite) TearDownSuite() {
	err := os.Remove(suite.app)
	suite.Assert().NoError(err)
	time.Sleep(time.Second * 5)
	suite.Assert().NoError(err)
	suite.avsClient.Close()

	err = docker_compose_down()
	if err != nil {
		fmt.Println("unable to stop docker compose down")
	}
}

func (suite *CmdTestSuite) runCmd(asvecCmd ...string) ([]string, error) {
	cmd := exec.Command(suite.app, asvecCmd...)
	cmd.Env = []string{"GOCOVERDIR=" + os.Getenv("COVERAGE_DIR")}
	stdout, err := cmd.Output()
	// fmt.Printf("stdout: %v", string(stdout))

	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return []string{string(ee.Stderr)}, err
		}
		return []string{string(stdout)}, err
	}

	lines := strings.Split(string(stdout), "\n")

	return lines, nil
}

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

type IndexDefinitionBuilder struct {
	indexName             string
	namespace             string
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
					Disabled:   getBoolPtr(false),
				},
			},
		},
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
	if idb.hnsfBatchingDisabled != nil {
		indexDef.Params.(*protos.IndexDefinition_HnswParams).HnswParams.BatchingParams.Disabled = idb.hnsfBatchingDisabled
	}

	return indexDef
}

func (suite *CmdTestSuite) TestSuccessfulCreateIndexCmd() {
	testCases := []struct {
		name           string
		indexName      string // index names must be unique for the suite
		indexNamespace string
		cmd            string
		expected_index *protos.IndexDefinition
	}{
		{
			"test with storage config",
			"index1",
			"test",
			fmt.Sprintf("create index --seeds %s -n test -i index1 -d 256 -m SQUARED_EUCLIDEAN --vector-field vector1 --storage-namespace bar --storage-set testbar --timeout 10s", suite.avsHostPort.String()),
			NewIndexDefinitionBuilder("index1", "test", 256, protos.VectorDistanceMetric_SQUARED_EUCLIDEAN, "vector1").
				WithStorageNamespace("bar").
				WithStorageSet("testbar").
				Build(),
		},
		{
			"test with hnsw params",
			"index2",
			"test",
			fmt.Sprintf("create index --timeout 10s --seeds %s -n test -i index2 -d 256 -m HAMMING --vector-field vector2 --hnsw-max-edges 10 --hnsw-ef 11 --hnsw-ef-construction 12", suite.avsHostPort.String()),
			NewIndexDefinitionBuilder("index2", "test", 256, protos.VectorDistanceMetric_HAMMING, "vector2").
				WithHnswM(10).
				WithHnswEf(11).
				WithHnswEfConstruction(12).
				Build(),
		},
		{
			"test with hnsw batch params",
			"index3",
			"test",
			fmt.Sprintf("create index --timeout 10s --seeds %s -n test -i index3 -d 256 -m COSINE --vector-field vector3 --hnsw-batch-enabled false --hnsw-batch-interval 50 --hnsw-batch-max-records 100", suite.avsHostPort.String()),
			NewIndexDefinitionBuilder("index3", "test", 256, protos.VectorDistanceMetric_COSINE, "vector3").
				WithHnswBatchingMaxRecord(100).
				WithHnswBatchingInterval(50).
				WithHnswBatchingDisabled(true).
				Build(),
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			lines, err := suite.runCmd(strings.Split(tc.cmd, " ")...)
			suite.Assert().NoError(err, "error: %s, stdout/err: %s", err, lines)

			actual, err := suite.avsClient.IndexGet(context.Background(), tc.indexNamespace, tc.indexName)

			time.Sleep(time.Second)

			if err != nil {
				suite.FailNowf("unable to get index", "%v", err)
			}

			suite.EqualExportedValues(tc.expected_index, actual)
		})
	}
}

func (suite *CmdTestSuite) TestCreateIndexFailsAlreadyExistsCmd() {
	lines, err := suite.runCmd(strings.Split(fmt.Sprintf("create index --seeds %s -n test -i exists -d 256 -m SQUARED_EUCLIDEAN --vector-field vector1 --storage-namespace bar --storage-set testbar --timeout 10s", suite.avsHostPort.String()), " ")...)
	suite.Assert().NoError(err, "index should have NOT existed on first call. error: %s, stdout/err: %s", err, lines)

	lines, err = suite.runCmd(strings.Split(fmt.Sprintf("create index --seeds %s -n test -i exists -d 256 -m SQUARED_EUCLIDEAN --vector-field vector1 --storage-namespace bar --storage-set testbar --timeout 10s", suite.avsHostPort.String()), " ")...)
	suite.Assert().Error(err, "index should HAVE existed on first call. error: %s, stdout/err: %s", err, lines)

	suite.Assert().Contains(lines[0], "AlreadyExists")
}

func (suite *CmdTestSuite) TestSuccessfulDropIndexCmd() {
	testCases := []struct {
		name           string
		indexName      string // index names must be unique for the suite
		indexNamespace string
		indexSet       []string
		cmd            string
	}{
		{
			"test with just namespace",
			"indexdrop1",
			"test",
			nil,
			fmt.Sprintf("drop index --seeds %s -n test -i indexdrop1 --timeout 10s", suite.avsHostPort.String()),
		},
		{
			"test with set",
			"indexdrop2",
			"test",
			[]string{
				"testset",
			},
			fmt.Sprintf("drop index --seeds %s -n test -s testset -i indexdrop2 --timeout 10s", suite.avsHostPort.String()),
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			err := suite.avsClient.IndexCreate(context.Background(), tc.indexNamespace, tc.indexSet, tc.indexName, "vector", 1, protos.VectorDistanceMetric_COSINE, nil, nil, nil)
			if err != nil {
				suite.FailNowf("unable to create index", "%v", err)
			}

			lines, err := suite.runCmd(strings.Split(tc.cmd, " ")...)
			suite.Assert().NoError(err, "error: %s, stdout/err: %s", err, lines)

			_, err = suite.avsClient.IndexGet(context.Background(), tc.indexNamespace, tc.indexName)

			time.Sleep(time.Second)

			if err == nil {
				suite.FailNow("err is nil, that means the index still exists")
			}
		})
	}
}

func (suite *CmdTestSuite) TestDropIndexFailsDoesNotExistCmd() {
	lines, err := suite.runCmd(strings.Split(fmt.Sprintf("drop index --seeds %s -n test -i DNE --timeout 10s", suite.avsHostPort.String()), " ")...)

	suite.Assert().Error(err, "index should have NOT existed. stdout/err: %s", lines)
	suite.Assert().Contains(lines[0], "server error")
}

func (suite *CmdTestSuite) TestFailInvalidArg() {
	testCases := []struct {
		name   string
		cmd    string
		errStr string
	}{
		{
			"use seeds and hosts together",
			fmt.Sprintf("create index --seeds %s --host 1.1.1.1:3001 -n test -i index1 -d 256 -m SQUARED_EUCLIDEAN --vector-field vector1 --storage-namespace bar --storage-set testbar --timeout 10s", suite.avsHostPort.String()),
			"Error: only --seeds or --host allowed",
		},
		{
			"use seeds and hosts together",
			fmt.Sprintf("list index --seeds %s --host 1.1.1.1:3001", suite.avsHostPort.String()),
			"Error: only --seeds or --host allowed",
		},
		{
			"use seeds and hosts together",
			fmt.Sprintf("drop index --seeds %s --host 1.1.1.1:3001 -n test -i index1", suite.avsHostPort.String()),
			"Error: only --seeds or --host allowed",
		},
		{
			"test with bad dimension",
			"create index --host 1.1.1.1:3001  -n test -i index1 -d -1 -m SQUARED_EUCLIDEAN --vector-field vector1 --storage-namespace bar --storage-set testbar --timeout 10s",
			"Error: invalid argument \"-1\" for \"-d, --dimension\"",
		},
		{
			"test with bad distance metric",
			"create index --host 1.1.1.1:3001  -n test -i index1 -d 10 -m BAD --vector-field vector1 --storage-namespace bar --storage-set testbar --timeout 10s",
			"Error: invalid argument \"BAD\" for \"-m, --distance-metric\"",
		},
		{
			"test with bad timeout",
			"create index --host 1.1.1.1:3001  -n test -i index1 -d 10 -m SQUARED_EUCLIDEAN --vector-field vector1 --storage-namespace bar --storage-set testbar --timeout 10",
			"Error: invalid argument \"10\" for \"--timeout\"",
		},
		{
			"test with bad hnsw-batch-enabled",
			"create index --hnsw-batch-enabled foo --host 1.1.1.1:3001  -n test -i index1 -d 10 -m SQUARED_EUCLIDEAN --vector-field vector1 --storage-namespace bar --storage-set testbar --timeout 10",
			"Error: invalid argument \"foo\" for \"--hnsw-batch-enabled\"",
		},
		{
			"test with bad hnsw-batch-interval",
			"create index --hnsw-batch-interval foo --host 1.1.1.1:3001  -n test -i index1 -d 10 -m SQUARED_EUCLIDEAN --vector-field vector1 --storage-namespace bar --storage-set testbar --timeout 10",
			"Error: invalid argument \"foo\" for \"--hnsw-batch-interval\"",
		},
		{
			"test with bad hnsw-batch-max-records",
			"create index --hnsw-batch-max-records foo --host 1.1.1.1:3001  -n test -i index1 -d 10 -m SQUARED_EUCLIDEAN --vector-field vector1 --storage-namespace bar --storage-set testbar --timeout 10",
			"Error: invalid argument \"foo\" for \"--hnsw-batch-max-records\"",
		},
		{
			"test with bad hnsw-ef",
			"create index --hnsw-ef foo --host 1.1.1.1:3001  -n test -i index1 -d 10 -m SQUARED_EUCLIDEAN --vector-field vector1 --storage-namespace bar --storage-set testbar --timeout 10",
			"Error: invalid argument \"foo\" for \"--hnsw-ef\"",
		},
		{
			"test with bad hnsw-ef-construction",
			"create index --hnsw-ef-construction foo --host 1.1.1.1:3001  -n test -i index1 -d 10 -m SQUARED_EUCLIDEAN --vector-field vector1 --storage-namespace bar --storage-set testbar --timeout 10",
			"Error: invalid argument \"foo\" for \"--hnsw-ef-construction\"",
		},
		{
			"test with bad hnsw-max-edges",
			"create index --hnsw-max-edges foo --host 1.1.1.1:3001  -n test -i index1 -d 10 -m SQUARED_EUCLIDEAN --vector-field vector1 --storage-namespace bar --storage-set testbar --timeout 10",
			"Error: invalid argument \"foo\" for \"--hnsw-max-edges\"",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			lines, err := suite.runCmd(strings.Split(tc.cmd, " ")...)

			suite.Assert().Error(err, "error: %s, stdout/err: %s", err, lines)
			suite.Assert().Contains(lines[0], tc.errStr)
		})
	}
}

func docker_compose_up() error {
	fmt.Println("Starting docker containers")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// docker/docker-compose.yml
	cmd := exec.CommandContext(ctx, "docker", "compose", fmt.Sprintf("-fdocker/docker-compose.yml"), "up", "-d")
	output, err := cmd.CombinedOutput()

	fmt.Printf("docker compose up output: %s\n", string(output))

	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			return err
		}
		return err
	}

	return nil
}

func docker_compose_down() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	cmd := exec.CommandContext(ctx, "docker", "compose", fmt.Sprintf("-fdocker/docker-compose.yml"), "down")
	_, err := cmd.Output()

	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			return err
		}
		return err
	}

	return nil
}

// func Index
