//go:build integration || integration_large

package main

import (
	"asvec/cmd/flags"
	"asvec/tests"
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	avs "github.com/aerospike/avs-client-go"
	"github.com/aerospike/avs-client-go/protos"
	"github.com/stretchr/testify/suite"
)

var logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))

var (
	testNamespace = "test"
	testSet       = "testset"
	barNamespace  = "bar"
)

type CmdTestSuite struct {
	tests.CmdBaseTestSuite
	configFileClusterName string
}

func TestCmdSuite(t *testing.T) {
	logger = logger.With(slog.Bool("test-logger", true)) // makes it easy to see which logger is which
	rootCA, err := tests.GetCACert("docker/tls/config/tls/ca.aerospike.com.crt")
	if err != nil {
		t.Fatalf("unable to read root ca %v", err)
		t.FailNow()
		logger.Error("Failed to read cert")
	}

	certificates, err := tests.GetCertificates("docker/mtls/config/tls/localhost.crt", "docker/mtls/config/tls/localhost.key")
	if err != nil {
		t.Fatalf("unable to read certificates %v", err)
		t.FailNow()
		logger.Error("Failed to read cert")
	}

	avsSeed := "localhost"
	avsPort := 10000
	avsHostPort := avs.NewHostPort(avsSeed, avsPort)

	logger.Info("%v", slog.Any("cert", rootCA))
	suites := []*CmdTestSuite{
		{
			configFileClusterName: "vanilla",
			CmdBaseTestSuite: tests.CmdBaseTestSuite{

				ComposeFile: "docker/vanilla/docker-compose.yml", // vanilla
				SuiteFlags: []string{
					"--log-level debug",
					"--timeout 10s",
					tests.CreateFlagStr(flags.Seeds, avsHostPort.String()),
				},
				AvsHostPort: avsHostPort,
			},
		},
		{
			configFileClusterName: "tls",
			CmdBaseTestSuite: tests.CmdBaseTestSuite{
				ComposeFile: "docker/tls/docker-compose.yml", // tls
				SuiteFlags: []string{
					"--log-level debug",
					"--timeout 10s",
					tests.CreateFlagStr(flags.Seeds, avsHostPort.String()),
					tests.CreateFlagStr(flags.TLSCaFile, "docker/tls/config/tls/ca.aerospike.com.crt"),
				},
				AvsTLSConfig: &tls.Config{
					Certificates: nil,
					RootCAs:      rootCA,
				},
				AvsHostPort: avsHostPort,
			},
		},
		{
			configFileClusterName: "mtls",
			CmdBaseTestSuite: tests.CmdBaseTestSuite{
				ComposeFile: "docker/mtls/docker-compose.yml", // mutual tls
				SuiteFlags: []string{
					"--log-level debug",
					"--timeout 10s",
					tests.CreateFlagStr(flags.Host, avsHostPort.String()),
					tests.CreateFlagStr(flags.TLSCaFile, "docker/mtls/config/tls/ca.aerospike.com.crt"),
					tests.CreateFlagStr(flags.TLSCertFile, "docker/mtls/config/tls/localhost.crt"),
					tests.CreateFlagStr(flags.TLSKeyFile, "docker/mtls/config/tls/localhost.key"),
				},
				AvsTLSConfig: &tls.Config{
					Certificates: certificates,
					RootCAs:      rootCA,
				},
				AvsHostPort: avsHostPort,
			},
		},
		{
			configFileClusterName: "auth",
			CmdBaseTestSuite: tests.CmdBaseTestSuite{
				ComposeFile: "docker/auth/docker-compose.yml", // tls + auth (auth requires tls)
				SuiteFlags: []string{
					"--log-level debug",
					"--timeout 10s",
					tests.CreateFlagStr(flags.Host, avsHostPort.String()),
					tests.CreateFlagStr(flags.TLSCaFile, "docker/auth/config/tls/ca.aerospike.com.crt"),
					tests.CreateFlagStr(flags.AuthUser, "admin"),
					tests.CreateFlagStr(flags.AuthPassword, "admin"),
				},
				AvsCreds: avs.NewCredentialsFromUserPass("admin", "admin"),
				AvsTLSConfig: &tls.Config{
					Certificates: nil,
					RootCAs:      rootCA,
				},
				AvsHostPort: avsHostPort,
			},
		},
	}

	testSuiteEnv := os.Getenv("ASVEC_TEST_SUITES")
	picked_suites := map[int]struct{}{}

	if testSuiteEnv != "" {
		testSuites := strings.Split(testSuiteEnv, ",")

		for _, s := range testSuites {
			i, err := strconv.Atoi(s)
			if err != nil {
				t.Fatalf("unable to convert %s to int: %v", s, err)
			}

			picked_suites[i] = struct{}{}
		}
	}

	logger.Info("Running test suites", slog.Any("suites", picked_suites))

	for i, s := range suites {
		if len(picked_suites) != 0 {
			if _, ok := picked_suites[i]; !ok {
				continue
			}
		}
		suite.Run(t, s)
	}
}

func (suite *CmdTestSuite) SkipIfUserPassAuthDisabled() {
	if suite.AvsCreds == nil {
		suite.T().Skip("authentication is disabled. skipping test")
	}
}

func (suite *CmdTestSuite) TestSuccessfulCreateIndexCmd() {
	testCases := []struct {
		name           string
		indexName      string // index names must be unique for the suite
		indexNamespace string
		cmd            string
		expectedIndex  *protos.IndexDefinition
	}{
		{
			name:           "test with labels",
			indexName:      "index0",
			indexNamespace: "test",
			cmd:            "index create -y -n test -i index0 -d 256 -m SQUARED_EUCLIDEAN --vector-field vector0 --index-labels model=all-MiniLM-L6-v2,foo=bar",
			expectedIndex: tests.NewIndexDefinitionBuilder(false, "index0", "test", 256, protos.VectorDistanceMetric_SQUARED_EUCLIDEAN, "vector0").
				WithLabels(map[string]string{"model": "all-MiniLM-L6-v2", "foo": "bar"}).
				Build(),
		},
		{
			name:           "test with storage config",
			indexName:      "index1",
			indexNamespace: "test",
			cmd:            "index create -y -n test -i index1 -d 256 -m SQUARED_EUCLIDEAN --vector-field vector1 --storage-namespace bar --storage-set testbar",
			expectedIndex: tests.NewIndexDefinitionBuilder(false, "index1", "test", 256, protos.VectorDistanceMetric_SQUARED_EUCLIDEAN, "vector1").
				WithStorageNamespace("bar").
				WithStorageSet("testbar").
				Build(),
		},
		{
			name:           "test with hnsw params and seeds",
			indexName:      "index2",
			indexNamespace: "test",
			cmd:            "index create -y -n test -i index2 -d 256 -m HAMMING --vector-field vector2 --hnsw-m 10 --hnsw-ef 11 --hnsw-ef-construction 12 --hnsw-max-mem-queue-size 10",
			expectedIndex: tests.NewIndexDefinitionBuilder(false, "index2", "test", 256, protos.VectorDistanceMetric_HAMMING, "vector2").
				WithHnswM(10).
				WithHnswEf(11).
				WithHnswEfConstruction(12).
				WithHnswMaxMemQueueSize(10).
				Build(),
		},
		{
			name:           "test with hnsw batch params",
			indexName:      "index3",
			indexNamespace: "test",
			cmd:            "index create -y -n test -i index3 -d 256 -m COSINE --vector-field vector3 --hnsw-batch-index-interval 50s --hnsw-batch-max-index-records 10001 --hnsw-batch-reindex-interval 50s --hnsw-batch-max-reindex-records 10001",
			expectedIndex: tests.NewIndexDefinitionBuilder(false, "index3", "test", 256, protos.VectorDistanceMetric_COSINE, "vector3").
				WithHnswBatchingMaxIndexRecord(10001).
				WithHnswBatchingIndexInterval(50000).
				WithHnswBatchingMaxReindexRecord(10001).
				WithHnswBatchingReindexInterval(50000).
				Build(),
		},
		{
			name:           "test with hnsw cache params",
			indexName:      "index4",
			indexNamespace: "test",
			cmd:            "index create -y -n test -i index4 -d 256 -m COSINE --vector-field vector4 --hnsw-index-cache-max-entries 1000 --hnsw-index-cache-expiry 10s",
			expectedIndex: tests.NewIndexDefinitionBuilder(false, "index4", "test", 256, protos.VectorDistanceMetric_COSINE, "vector4").
				WithHnswIndexCacheExpiry(10000).
				WithHnswIndexCacheMaxEntries(1000).
				Build(),
		},
		{
			name:           "test with hnsw healer params",
			indexName:      "index5",
			indexNamespace: "test",
			cmd:            "index create -y -n test -i index5 -d 256 -m COSINE --vector-field vector5 --hnsw-healer-max-scan-rate-per-node 1000 --hnsw-healer-max-scan-page-size 1000 --hnsw-healer-reindex-percent 10.10 --hnsw-healer-schedule \"0 0 0 ? * *\" --hnsw-healer-parallelism 10",
			expectedIndex: tests.NewIndexDefinitionBuilder(false, "index5", "test", 256, protos.VectorDistanceMetric_COSINE, "vector5").
				WithHnswHealerMaxScanRatePerNode(1000).
				WithHnswHealerMaxScanPageSize(1000).
				WithHnswHealerReindexPercent(10.10).
				WithHnswHealerSchedule("0 0 0 ? * *").
				WithHnswHealerParallelism(10).
				Build(),
		},
		{
			name:           "test with hnsw merge params",
			indexName:      "index6",
			indexNamespace: "test",
			cmd:            "index create -y -n test -i index6 -d 256 -m COSINE --vector-field vector6 --hnsw-merge-index-parallelism 10 --hnsw-merge-reindex-parallelism 11",
			expectedIndex: tests.NewIndexDefinitionBuilder(false, "index6", "test", 256, protos.VectorDistanceMetric_COSINE, "vector6").
				WithHnswMergeIndexParallelism(10).
				WithHnswMergeReIndexParallelism(11).
				Build(),
		},
		{
			name:           "test with yaml file",
			indexName:      "yaml-file-index",
			indexNamespace: "test",
			cmd:            fmt.Sprintf("index create -y --file tests/indexDef.yaml"),
			expectedIndex: tests.NewIndexDefinitionBuilder(false, "yaml-file-index", "test", 10, protos.VectorDistanceMetric_COSINE, "vector").
				WithSet("testset").
				WithHnswEf(101).
				WithHnswEfConstruction(102).
				WithHnswM(103).
				WithHnswMaxMemQueueSize(10004).
				WithHnswBatchingIndexInterval(30001).
				WithHnswBatchingMaxIndexRecord(100001).
				WithHnswBatchingReindexInterval(30002).
				WithHnswBatchingMaxReindexRecord(100002).
				WithHnswIndexCacheMaxEntries(1001).
				WithHnswIndexCacheExpiry(1002).
				WithHnswRecordCacheMaxEntries(1006).
				WithHnswRecordCacheExpiry(1007).
				WithHnswHealerParallelism(7).
				WithHnswHealerMaxScanRatePerNode(1).
				WithHnswHealerMaxScanPageSize(2).
				WithHnswHealerReindexPercent(3).
				WithHnswHealerSchedule("0 15 10 ? * 6L 2022-2025").
				WithHnswMergeIndexParallelism(7).
				WithHnswMergeReIndexParallelism(5).
				WithStorageNamespace("test").
				WithStorageSet("name").
				Build(),
		},
		{
			name:           "test with enable vector integrity check",
			indexName:      "integidx",
			indexNamespace: "test",
			cmd:            "index create -y -n test -i integidx -d 256 -m COSINE --vector-field vector --hnsw-healer-max-scan-rate-per-node 1000 --hnsw-healer-max-scan-page-size 1000 --hnsw-healer-reindex-percent 10.10 --hnsw-healer-schedule \"0 0 0 ? * *\" --hnsw-healer-parallelism 10 --hnsw-vector-integrity-check false",
			expectedIndex: tests.NewIndexDefinitionBuilder(false, "integidx", "test", 256, protos.VectorDistanceMetric_COSINE, "vector").
				WithHnswHealerMaxScanRatePerNode(1000).
				WithHnswHealerMaxScanPageSize(1000).
				WithHnswHealerReindexPercent(10.10).
				WithHnswHealerSchedule("0 0 0 ? * *").
				WithHnswHealerParallelism(10).
				WithHnswVectorIntegrityCheck(false).
				Build(),
		},
		{
			name:           "test with record caching",
			indexName:      "recidx",
			indexNamespace: "test",
			cmd:            "index create -y -n test -i recidx -d 256 -m SQUARED_EUCLIDEAN --vector-field vector0 --hnsw-record-cache-max-entries 1001 --hnsw-record-cache-expiry 20s",
			expectedIndex: tests.NewIndexDefinitionBuilder(false, "recidx", "test", 256, protos.VectorDistanceMetric_SQUARED_EUCLIDEAN, "vector0").
				WithHnswRecordCacheMaxEntries(1001).
				WithHnswRecordCacheExpiry(20000).
				Build(),
		},
		{
			name:           "test with infinite record cache expiry",
			indexName:      "recinfidx",
			indexNamespace: "test",
			cmd:            "index create -y -n test -i recinfidx -d 256 -m SQUARED_EUCLIDEAN --vector-field vector0 --hnsw-record-cache-expiry -1",
			expectedIndex: tests.NewIndexDefinitionBuilder(false, "recinfidx", "test", 256, protos.VectorDistanceMetric_SQUARED_EUCLIDEAN, "vector0").
				WithHnswRecordCacheExpiry(-1).
				Build(),
		},
		{
			name:           "test with infinite index cache expiry",
			indexName:      "idxinfidx",
			indexNamespace: "test",
			cmd:            "index create -y -n test -i idxinfidx -d 256 -m SQUARED_EUCLIDEAN --vector-field vector0 --hnsw-index-cache-expiry -1",
			expectedIndex: tests.NewIndexDefinitionBuilder(false, "idxinfidx", "test", 256, protos.VectorDistanceMetric_SQUARED_EUCLIDEAN, "vector0").
				WithHnswIndexCacheExpiry(-1).
				Build(),
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			lines, stderr, err := suite.RunSuiteCmd(strings.FieldsFunc(tc.cmd, tests.SplitQuotedString)...)

			if err != nil {
				suite.Assert().NoError(err, "error: %s, stdout: %s, stderr: %s", err, lines, stderr)
				suite.FailNow("unable to index create")
			}

			actual, err := suite.AvsClient.IndexGet(context.Background(), tc.indexNamespace, tc.indexName, false)

			if err != nil {
				suite.FailNowf("unable to get index", "%v", err)
			}

			suite.EqualExportedValues(tc.expectedIndex, actual)
		})
	}
}

func (suite *CmdTestSuite) TestPipeFromListIndexToCreateIndex() {
	suite.CleanUpIndexes(context.Background())

	testCases := []struct {
		name          string
		indexDefs     []*protos.IndexDefinition
		sedReplaceStr string
		createFail    bool
		checkContains []string
	}{
		{
			name: "test with all indexes succeed",
			indexDefs: []*protos.IndexDefinition{
				tests.NewIndexDefinitionBuilder(false,
					"exists1", "test", 256, protos.VectorDistanceMetric_COSINE, "vector",
				).Build(),
				tests.NewIndexDefinitionBuilder(false,
					"exists2", "bar", 256, protos.VectorDistanceMetric_HAMMING, "vector",
				).WithSet("barset").Build(),
			},
			sedReplaceStr: "s/exists/does-not-exist-yet/g",
			createFail:    false,
			checkContains: []string{
				"Successfully created index test.*.does-not-exist-yet1",
				"Successfully created index bar.barset.does-not-exist-yet2",
				"Successfully created all indexes from yaml",
			},
		},
		{
			name: "test with one index that fails",
			indexDefs: []*protos.IndexDefinition{
				tests.NewIndexDefinitionBuilder(false,
					"exists3", "test", 256, protos.VectorDistanceMetric_COSINE, "vector",
				).Build(),
				tests.NewIndexDefinitionBuilder(false,
					"exists4", "bar", 256, protos.VectorDistanceMetric_HAMMING, "vector",
				).WithSet("barset").Build(),
			},
			sedReplaceStr: "s/exists3/does-not-exist-yet2/g",
			createFail:    true,
			checkContains: []string{
				"Successfully created index test.*.does-not-exist-yet2",
				"Failed to create index bar.barset.exists4",
				"Some indexes failed to be created",
			},
		},
		{
			name: "test with no index successfully created",
			indexDefs: []*protos.IndexDefinition{
				tests.NewIndexDefinitionBuilder(false,
					"exists1", "test", 256, protos.VectorDistanceMetric_COSINE, "vector",
				).Build(),
				tests.NewIndexDefinitionBuilder(false,
					"exists2", "bar", 256, protos.VectorDistanceMetric_HAMMING, "vector",
				).WithSet("barset").Build(),
			},
			sedReplaceStr: "s/COSINE/HAMMING/g", // Doing this rather than removing sed for this test case
			createFail:    true,
			checkContains: []string{
				"Failed to create index test.*.exists1",
				"Failed to create index bar.barset.exists2",
				"Unable to create any new indexes",
			},
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {

			for _, index := range tc.indexDefs {
				err := suite.AvsClient.IndexCreateFromIndexDef(context.Background(), index)
				if err != nil {
					suite.FailNowf("unable to index create", "%v", err)
				}

				defer suite.AvsClient.IndexDrop(context.Background(), index.Id.Namespace, index.Id.Name)
			}

			// Test "asvec index list --yaml | sed 's/exists1/does-not-exist-yet/g' | asvec index create"

			listArgs := []string{"index", "list", "--yaml"}
			listArgs = suite.AddSuiteArgs(listArgs...)
			listCmd := suite.GetCmd(listArgs...)
			listPipe, err := listCmd.StdoutPipe()

			if err != nil {
				suite.FailNowf("unable to create list pipe", "%v", err)
			}

			// We need to change the name to something that does not exist yet
			sedCmd := exec.Command("sed", tc.sedReplaceStr)
			sedPipe, err := sedCmd.StdoutPipe()

			if err != nil {
				suite.FailNowf("unable to create sed pipe", "%v", err)
			}

			listPipeStdout := &bytes.Buffer{}
			sedPipeStdout := &bytes.Buffer{}
			sedCmd.Stdin = io.TeeReader(listPipe, listPipeStdout)
			// sedCmd.Stdin = listPipe

			createArgs := []string{"index", "create", "--log-level", "debug"}
			createArgs = suite.AddSuiteArgs(createArgs...)
			createCmd := suite.GetCmd(createArgs...)
			createCmd.Stdin = io.TeeReader(sedPipe, sedPipeStdout)
			// createCmd.Stdin = sedPipe

			// Start list and sed commands so data can flow through the pipes
			if err := listCmd.Start(); err != nil {
				suite.FailNowf("unable to start list cmd", "%v", err)
			}

			logger.Debug("started list command", slog.String("cmd", listCmd.String()))

			// Need to pause a bit while listCmd has some output
			time.Sleep(time.Second * 1)

			if err := sedCmd.Start(); err != nil {
				suite.FailNowf("unable to start sed cmd", "%v", err)
			}

			logger.Debug("started sed command", slog.String("cmd", sedCmd.String()))

			// Need to pause a bit while listCmd has some output
			time.Sleep(time.Second * 1)

			// Run create Cmd to completion
			logger.Debug("running create command", slog.String("cmd", createCmd.String()))
			output, err := createCmd.CombinedOutput()
			logger.Debug(string(output))

			// Cleanup list and sed commands
			if err := listCmd.Wait(); err != nil {
				logger.Debug("asvec index ls output", slog.String("output", listPipeStdout.String()))
				logger.Debug("sed output", slog.String("output", sedPipeStdout.String()))
				suite.FailNowf("unable to wait for list cmd", "%v", err)
			}

			if err := sedCmd.Wait(); err != nil {
				logger.Debug("asvec index ls output", slog.String("output", listPipeStdout.String()))
				logger.Debug("sed output", slog.String("output", sedPipeStdout.String()))
				suite.FailNowf("unable to wait for sed cmd", "%v", err)
			}

			if tc.createFail && err == nil {
				logger.Debug("asvec index ls output", slog.String("output", listPipeStdout.String()))
				logger.Debug("sed output", slog.String("output", sedPipeStdout.String()))
				suite.Failf("expected create cmd to fail because at least one index failed to be created", "%v", err)
			} else if !tc.createFail && err != nil {
				logger.Debug("asvec index ls output", slog.String("output", listPipeStdout.String()))
				logger.Debug("sed output", slog.String("output", sedPipeStdout.String()))
				suite.Failf("expected create cmd to succeed because all indexes were created", "%v : %s", err.Error(), output)
			}

			for _, str := range tc.checkContains {
				suite.Assert().Contains(string(output), str)
			}
		})
	}
}

func (suite *CmdTestSuite) TestSuccessfulUpdateIndexCmd() {
	ns := "test"
	index := "successful-update"

	newBuilder := func() *tests.IndexDefinitionBuilder {
		return tests.NewIndexDefinitionBuilder(true, index, ns, 256, protos.VectorDistanceMetric_COSINE, "field")
	}

	testCases := []struct {
		name           string
		indexName      string // index names must be unique for the suite
		indexNamespace string
		cmd            string
		expectedIndex  *protos.IndexDefinition
	}{
		{
			name:           "test with hnsw params and seeds",
			indexName:      "successful-update",
			indexNamespace: ns,
			cmd:            "index update -y -n test -i successful-update --index-labels new-label=foo --hnsw-max-mem-queue-size 10",
			expectedIndex: newBuilder().
				WithLabels(map[string]string{"new-label": "foo"}).
				WithHnswMaxMemQueueSize(10).
				Build(),
		},
		{
			name:           "test with hnsw batch params",
			indexName:      "successful-update",
			indexNamespace: "test",
			cmd:            "index update -y -n test -i successful-update --hnsw-batch-index-interval 50s --hnsw-batch-max-index-records 10001 --hnsw-batch-reindex-interval 50s --hnsw-batch-max-reindex-records 10001",
			expectedIndex: newBuilder().
				WithHnswBatchingMaxIndexRecord(10001).
				WithHnswBatchingIndexInterval(50000).
				WithHnswBatchingMaxReindexRecord(10001).
				WithHnswBatchingReindexInterval(50000).
				Build(),
		},
		{
			name:           "test with hnsw cache params",
			indexName:      "successful-update",
			indexNamespace: "test",
			cmd:            "index update -y -n test -i successful-update --hnsw-index-cache-max-entries 1000 --hnsw-index-cache-expiry 10s",
			expectedIndex: newBuilder().
				WithHnswIndexCacheExpiry(10000).
				WithHnswIndexCacheMaxEntries(1000).
				Build(),
		},
		{
			name:           "test with hnsw healer params",
			indexName:      "successful-update",
			indexNamespace: "test",
			cmd:            "index update -y -n test -i successful-update --hnsw-healer-max-scan-rate-per-node 1000 --hnsw-healer-max-scan-page-size 1000 --hnsw-healer-reindex-percent 10.10 --hnsw-healer-schedule \"0 30 11 ? * 6#2\" --hnsw-healer-parallelism 10",
			expectedIndex: newBuilder().
				WithHnswHealerMaxScanRatePerNode(1000).
				WithHnswHealerMaxScanPageSize(1000).
				WithHnswHealerReindexPercent(10.10).
				WithHnswHealerSchedule("0 30 11 ? * 6#2").
				WithHnswHealerParallelism(10).
				Build(),
		},
		{
			name:           "test with hnsw merge params",
			indexName:      "successful-update",
			indexNamespace: "test",
			cmd:            "index update -y -n test -i successful-update --hnsw-merge-index-parallelism 10 --hnsw-merge-reindex-parallelism 11",
			expectedIndex: newBuilder().
				WithHnswMergeIndexParallelism(10).
				WithHnswMergeReIndexParallelism(11).
				Build(),
		},
		{
			name:           "test with enable vector integrity check",
			indexName:      "successful-update",
			indexNamespace: "test",
			cmd:            "index update -y -n test -i successful-update --hnsw-vector-integrity-check false",
			expectedIndex: newBuilder().
				WithHnswVectorIntegrityCheck(false).
				Build(),
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			err := suite.AvsClient.IndexCreate(context.Background(), ns, "successful-update", "field", uint32(256), protos.VectorDistanceMetric_COSINE, nil)
			if err != nil {
				suite.FailNowf("unable to index create", "%v", err)
			}

			defer suite.AvsClient.IndexDrop(context.Background(), ns, "successful-update")

			lines, stderr, err := suite.RunSuiteCmd(strings.FieldsFunc(tc.cmd, tests.SplitQuotedString)...)

			if err != nil {
				suite.Assert().NoError(err, "error: %s, stdout: %s, stderr: %s", err, lines, stderr)
				suite.FailNow("unable to index update")
			}

			time.Sleep(5 * time.Second)

			actual, err := suite.AvsClient.IndexGet(context.Background(), tc.indexNamespace, tc.indexName, false)

			if err != nil {
				suite.FailNowf("unable to get index", "%v", err)
			}

			suite.EqualExportedValues(tc.expectedIndex, actual)
		})
	}

}

func (suite *CmdTestSuite) TestUpdateIndexDoesNotExist() {
	_, lines, err := suite.RunSuiteCmd(strings.Split("index update -y -n test -i DNE --hnsw-merge-index-parallelism 10", " ")...)
	suite.Assert().Error(err, "index should have NOT existed. stdout/err: %s", lines)
	suite.Assert().Contains(lines, "server error")
}

func (suite *CmdTestSuite) TestSuccessfulGCIndexCmd() {
	index := "successful-gc"
	ns := "test"
	suite.AvsClient.IndexCreate(context.Background(), ns, index, "field", uint32(256), protos.VectorDistanceMetric_COSINE, nil)
	testCases := []struct {
		name           string
		indexName      string // index names must be unique for the suite
		indexNamespace string
		cmd            string
	}{
		{
			name:           "test with hnsw params and seeds",
			indexName:      "successful-gc",
			indexNamespace: ns,
			cmd:            "index gc -n test -i successful-gc -c 10",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			lines, _, err := suite.RunSuiteCmd(strings.Split(tc.cmd, " ")...)

			if err != nil {
				suite.Assert().NoError(err, "error: %s, stdout/err: %s", err, lines)
				suite.FailNow("unable to index gc")
			}
		})
	}
}

func (suite *CmdTestSuite) TestGCIndexDoesNotExist() {
	_, lines, err := suite.RunSuiteCmd(strings.Split("index gc -n test -i DNE -c 10", " ")...)
	suite.Assert().Error(err, "index should have NOT existed. stdout/err: %s", lines)
	suite.Assert().Contains(lines, "server error")
}

func (suite *CmdTestSuite) TestCreateIndexFailsAlreadyExistsCmd() {
	_, lines, err := suite.RunSuiteCmd(strings.Split("index create -y -n test -i exists -d 256 -m SQUARED_EUCLIDEAN --vector-field vector1 --storage-namespace bar --storage-set testbar", " ")...)
	suite.Assert().NoError(err, "index should have NOT existed on first call. error: %s, stdout/err: %s", err, lines)

	_, lines, err = suite.RunSuiteCmd(strings.Split("index create -y -n test -i exists -d 256 -m SQUARED_EUCLIDEAN --vector-field vector1 --storage-namespace bar --storage-set testbar", " ")...)
	suite.Assert().Error(err, "index should HAVE existed on first call. error: %s, stdout/err: %s", err, lines)

	suite.Assert().Contains(lines, "AlreadyExists")
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
			name:           "test with just namespace and seeds",
			indexName:      "indexdrop1",
			indexNamespace: "test",
			indexSet:       nil,
			cmd:            "index drop -y -n test -i indexdrop1",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			err := suite.AvsClient.IndexCreate(
				context.Background(),
				tc.indexNamespace,
				tc.indexName,
				"vector",
				1,
				protos.VectorDistanceMetric_COSINE,
				&avs.IndexCreateOpts{
					Sets: tc.indexSet,
				})
			if err != nil {
				suite.FailNowf("unable to index create", "%v", err)
			}

			time.Sleep(time.Second * 3)

			lines, _, err := suite.RunSuiteCmd(strings.Split(tc.cmd, " ")...)
			suite.Assert().NoError(err, "error: %s, stdout/err: %s", err, lines)

			if err != nil {
				suite.FailNow("unable to index drop")
			}

			_, err = suite.AvsClient.IndexGet(context.Background(), tc.indexNamespace, tc.indexName, false)

			time.Sleep(time.Second * 3)

			if err == nil {
				suite.FailNow("err is nil, that means the index still exists")
			}
		})
	}
}

func (suite *CmdTestSuite) TestDropIndexFailsDoesNotExistCmd() {
	_, lines, err := suite.RunSuiteCmd(strings.Split("index drop -y -n test -i DNE", " ")...)

	suite.Assert().Error(err, "index should have NOT existed. stdout/err: %s", lines)
	suite.Assert().Contains(lines, "server error")
}

func removeANSICodes(input string) string {
	re := regexp.MustCompile(`\x1b[^m]*m`)
	return re.ReplaceAllString(input, "")
}

func (suite *CmdTestSuite) TestSuccessfulListIndexCmd() {
	suite.CleanUpIndexes(context.Background())

	testCases := []struct {
		name          string
		indexes       []*protos.IndexDefinition
		cmd           string
		expectedTable string
	}{
		{
			name: "single index",
			indexes: []*protos.IndexDefinition{
				tests.NewIndexDefinitionBuilder(false,
					"list", "test", 256, protos.VectorDistanceMetric_COSINE, "vector",
				).Build(),
			},
			cmd: "index list --no-color --format 1",
			expectedTable: `Indexes
,Name,Namespace,Field,Dimensions,Distance Metric,Unmerged,Vector Records,Size,Unmerged %
1,list,test,vector,256,COSINE,0,0,0 B,0%
`,
		},
		{
			name: "double index with set",
			indexes: []*protos.IndexDefinition{
				tests.NewIndexDefinitionBuilder(false,
					"list1", "test", 256, protos.VectorDistanceMetric_COSINE, "vector",
				).Build(),
				tests.NewIndexDefinitionBuilder(false,
					"list2", "bar", 256, protos.VectorDistanceMetric_HAMMING, "vector",
				).WithSet("barset").Build(),
			},
			cmd: "index list --no-color --format 1",
			expectedTable: `Indexes
,Name,Namespace,Set,Field,Dimensions,Distance Metric,Unmerged,Vector Records,Size,Unmerged %
1,list2,bar,barset,vector,256,HAMMING,0,0,0 B,0%
2,list1,test,,vector,256,COSINE,0,0,0 B,0%
`,
		},
		{
			name: "double index with set, and verbose",
			indexes: []*protos.IndexDefinition{
				tests.NewIndexDefinitionBuilder(false,
					"list1", "test", 256, protos.VectorDistanceMetric_COSINE, "vector",
				).WithLabels(map[string]string{"foo": "bar"}).
					WithHnswMergeIndexParallelism(80).
					WithHnswMergeReIndexParallelism(26).
					WithHnswRecordCacheExpiry(20000).
					WithHnswRecordCacheMaxEntries(1003).
					Build(),
				tests.NewIndexDefinitionBuilder(false,
					"list2", "bar", 256, protos.VectorDistanceMetric_HAMMING, "vector",
				).WithSet("barset").
					WithHnswMergeIndexParallelism(80).
					WithHnswMergeReIndexParallelism(26).
					// -1 means never expire
					WithHnswRecordCacheExpiry(-1).
					WithHnswRecordCacheMaxEntries(1002).
					Build(),
			},
			cmd: "index list --verbose --no-color --format 1",
			expectedTable: `Indexes
,Name,Namespace,Set,Field,Dimensions,Distance Metric,Unmerged,Vector Records,Size,Unmerged %,Vertices,Labels*,Storage,Index Parameters
1,list2,bar,barset,vector,256,HAMMING,0,0,0 B,0%,0,map[],"Namespace\,bar
Set\,list2","HNSW
Max Edges\,16
Ef\,100
Construction Ef\,100
MaxMemQueueSize*\,1000000
Batch Max Index Records*\,100000
Batch Index Interval*\,30s
Batch Max Reindex Records*\,10000
Batch Reindex Interval*\,30s
Index Cache Max Entries*\,2000000
Index Cache Expiry*\,1h0m0s
Record Cache Max Entries*\,1002
Record Cache Expiry*\,-1ms
Healer Max Scan Rate / Node*\,1000
Healer Max Page Size*\,10000
Healer Re-index % *\,10.00%
Healer Schedule*\,0 0/15 * ? * * *
Healer Parallelism*\,1
Merge Index Parallelism*\,80
Merge Re-Index Parallelism*\,26
Enable Vector Integrity Check\,true"
2,list1,test,,vector,256,COSINE,0,0,0 B,0%,0,map[foo:bar],"Namespace\,test
Set\,list1","HNSW
Max Edges\,16
Ef\,100
Construction Ef\,100
MaxMemQueueSize*\,1000000
Batch Max Index Records*\,100000
Batch Index Interval*\,30s
Batch Max Reindex Records*\,10000
Batch Reindex Interval*\,30s
Index Cache Max Entries*\,2000000
Index Cache Expiry*\,1h0m0s
Record Cache Max Entries*\,1003
Record Cache Expiry*\,20s
Healer Max Scan Rate / Node*\,1000
Healer Max Page Size*\,10000
Healer Re-index % *\,10.00%
Healer Schedule*\,0 0/15 * ? * * *
Healer Parallelism*\,1
Merge Index Parallelism*\,80
Merge Re-Index Parallelism*\,26
Enable Vector Integrity Check\,true"
Values ending with * can be dynamically configured using the 'asvec index update' command.
`,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			for _, index := range tc.indexes {
				setFilter := []string{}
				if index.SetFilter != nil {
					setFilter = append(setFilter, *index.SetFilter)
				}

				err := suite.AvsClient.IndexCreate(
					context.Background(),
					index.Id.Namespace,
					index.Id.Name,
					index.GetField(),
					index.GetDimensions(),
					index.GetVectorDistanceMetric(),
					&avs.IndexCreateOpts{
						Sets:       setFilter,
						HnswParams: index.GetHnswParams(),
						Labels:     index.GetLabels(),
						Storage:    index.GetStorage(),
					},
				)
				if err != nil {
					suite.FailNowf("unable to create index", "%v", err)
				}

				defer suite.AvsClient.IndexDrop(
					context.Background(),
					index.Id.Namespace,
					index.Id.Name,
				)
			}

			actualTable, _, err := suite.RunSuiteCmd(strings.Split(tc.cmd, " ")...)
			suite.Assert().NoError(err, "error: %s, stdout/err: %s", err, actualTable)

			suite.Assert().Equal(tc.expectedTable, actualTable)

		})
	}
}

func (suite *CmdTestSuite) TestSuccessfulUserCreateCmd() {
	suite.SkipIfUserPassAuthDisabled()

	testCases := []struct {
		name         string       // name of the test case
		cmd          string       // command to be executed
		expectedUser *protos.User // expected user details after command execution
	}{
		{
			name: "create user with comma sep roles",
			cmd:  "users create --name foo1 --new-password foo --roles admin,read-write",
			expectedUser: &protos.User{
				Username: "foo1",
				Roles: []string{
					"admin",
					"read-write",
				},
			},
		},
		{
			name: "create user with comma multiple roles",
			cmd:  "users create --name foo2 --new-password foo --roles admin --roles read-write",
			expectedUser: &protos.User{
				Username: "foo2",
				Roles: []string{
					"admin",
					"read-write",
				},
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			lines, _, err := suite.RunSuiteCmd(strings.Split(tc.cmd, " ")...)
			suite.Assert().NoError(err, "error: %s, stdout/err: %s", err, lines)

			if err != nil {
				suite.FailNow("failed")
			}

			time.Sleep(time.Second * 1)

			actualUser, err := suite.AvsClient.GetUser(context.Background(), tc.expectedUser.Username)
			suite.Assert().NoError(err, "error: %s", err)

			suite.Assert().EqualExportedValues(tc.expectedUser, actualUser)
		})

	}
}

func (suite *CmdTestSuite) TestFailUserCreateCmd() {
	suite.SkipIfUserPassAuthDisabled()

	testCases := []struct {
		name        string // name of the test case
		cmd         string // command to be executed
		expectedErr string // expected error message after command execution
	}{
		{
			name:        "fail to create user with invalid role",
			cmd:         "users create --name foo1 --new-password foo --roles invalid",
			expectedErr: "unknown roles [invalid]",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			_, lines, err := suite.RunSuiteCmd(strings.Split(tc.cmd, " ")...)
			suite.Assert().Error(err, "error: %s, stdout/err: %s", err, lines)
			suite.Assert().Contains(lines, tc.expectedErr)
		})

	}
}

func (suite *CmdTestSuite) TestSuccessfulUserDropCmd() {
	suite.SkipIfUserPassAuthDisabled()

	testCases := []struct {
		name string
		user string
		cmd  string
	}{
		{
			"drop user",
			"drop0",
			"users drop --name drop0",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			err := suite.AvsClient.CreateUser(context.Background(), tc.user, tc.user, []string{"admin"})
			suite.Assert().NoError(err, "we were not able to create the user before we try to drop it", err)

			lines, stderr, err := suite.RunSuiteCmd(strings.Split(tc.cmd, " ")...)
			suite.Assert().NoError(err, "error: %s, stdout: %s stderr:%s", err, lines, stderr)

			if err != nil {
				suite.FailNow("failed")
			}

			_, err = suite.AvsClient.GetUser(context.Background(), tc.user)
			suite.Assert().Error(err, "we should not have retrieved the dropped user")
		})
	}
}

// Server treats non-existing users as a no-op in drop cmd
//
// func (suite *CmdTestSuite) TestFailedUserDropCmd() {

// 	if suite.AvsUser == nil {
// 		suite.T().Skip("authentication is disabled. skipping test")
// 	}

// 	lines, err := suite.RunCmd(strings.Split("users drop --name DNE", " ")...)
// 	suite.Assert().Error(err, "error: %s, stdout/err: %s", err, lines)
// 	suite.Assert().Contains(lines[0], "server error")
// }

func (suite *CmdTestSuite) TestSuccessfulUserGrantCmd() {
	suite.SkipIfUserPassAuthDisabled()

	testCases := []struct {
		name         string       // name of the test case
		user         string       // username of the user
		cmd          string       // command to be executed
		expectedUser *protos.User // expected user details after command execution
	}{
		{
			name: "grant user",
			user: "grant0",
			cmd:  "users grant --name grant0 --roles read-write",
			expectedUser: &protos.User{
				Username: "grant0",
				Roles:    []string{"read-write", "admin"},
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			err := suite.AvsClient.CreateUser(context.Background(), tc.user, "foo", []string{"admin"})
			suite.Assert().NoError(err, "we were not able to create the user before we try to grant it", err)

			lines, _, err := suite.RunSuiteCmd(strings.Split(tc.cmd, " ")...)
			suite.Assert().NoError(err, "error: %s, stdout/err: %s", err, lines)

			if err != nil {
				suite.FailNow("failed")
			}

			actualUser, err := suite.AvsClient.GetUser(context.Background(), tc.user)
			suite.Assert().NoError(err, "error: %s", err)

			suite.Assert().EqualExportedValues(tc.expectedUser, actualUser)
		})
	}
}

func (suite *CmdTestSuite) TestSuccessfulUserRevokeCmd() {
	suite.SkipIfUserPassAuthDisabled()

	testCases := []struct {
		name         string       // name of the test case
		user         string       // username of the user
		cmd          string       // command to be executed
		expectedUser *protos.User // expected user details after command execution
	}{
		{
			name: "revoke user",
			user: "revoke0",
			cmd:  "users revoke --name revoke0 --roles read-write",
			expectedUser: &protos.User{
				Username: "revoke0",
				Roles:    []string{"admin"},
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			err := suite.AvsClient.CreateUser(context.Background(), tc.user, "foo", []string{"admin", "read-write"})
			suite.Assert().NoError(err, "we were not able to create the user before we try to revoke it", err)

			lines, _, err := suite.RunSuiteCmd(strings.Split(tc.cmd, " ")...)
			suite.Assert().NoError(err, "error: %s, stdout/err: %s", err, lines)

			if err != nil {
				suite.FailNow("failed")
			}

			actualUser, err := suite.AvsClient.GetUser(context.Background(), tc.user)
			suite.Assert().NoError(err, "error: %s", err)

			suite.Assert().EqualExportedValues(tc.expectedUser, actualUser)
		})
	}
}

func (suite *CmdTestSuite) TestSuccessfulUsersNewPasswordCmd() {
	suite.SkipIfUserPassAuthDisabled()

	testCases := []struct {
		name        string
		user        string
		newPassword string
		cmd         string
	}{
		{
			name:        "change password",
			user:        "password0",
			newPassword: "foo",
			cmd:         "users new-password --name password0 --new-password foo",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			err := suite.AvsClient.CreateUser(context.Background(), tc.user, "oldpass", []string{"admin"})
			suite.Assert().NoError(err, "we were not able to create the user before we try to change password", err)

			lines, _, err := suite.RunSuiteCmd(strings.Split(tc.cmd, " ")...)
			suite.Assert().NoError(err, "error: %s, stdout/err: %s", err, lines)

			if err != nil {
				suite.FailNow("failed")
			}

			ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
			defer cancel()

			creds := avs.NewCredentialsFromUserPass(tc.user, tc.newPassword)
			_, err = avs.NewClient(
				ctx,
				avs.HostPortSlice{suite.AvsHostPort},
				nil,
				true,
				creds,
				suite.AvsTLSConfig,
				logger,
			)
			suite.Assert().NoError(err, "error: %s", err)
		})
	}
}

func (suite *CmdTestSuite) TestFailUsersNewPasswordCmd() {
	// If the user DNE it will not fail. Only fails if auth is disabled.
	if suite.AvsCreds != nil {
		suite.T().Skip("authentication is enabled. skipping test")
	}

	testCases := []struct {
		name        string
		user        string
		newPassword string
		cmd         string
	}{
		{
			name:        "change password with invalid user",
			user:        "DNE",
			newPassword: "foo",
			cmd:         "users new-password --name DNE --new-password foo",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			_, stderr, err := suite.RunSuiteCmd(strings.Split(tc.cmd, " ")...)
			suite.Assert().Error(err, "error: %s, stdout: %s stderr:%s", err, stderr, stderr)
			suite.Assert().Contains(stderr, "server error")
		})
	}
}

func (suite *CmdTestSuite) TestSuccessfulListUsersCmd() {
	suite.SkipIfUserPassAuthDisabled()

	testCases := []struct {
		name          string
		cmd           string
		expectedTable string
	}{
		{
			name: "users list",
			cmd:  "users list --no-color --format 1",
			expectedTable: `Users
,User,Roles
1,admin,"admin\, read-write"
Use 'role list' to view available roles
`,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			actualTable, _, err := suite.RunSuiteCmd(strings.Split(tc.cmd, " ")...)
			suite.Assert().NoError(err, "error: %s, stdout/err: %s", err, actualTable)

			suite.Assert().Equal(tc.expectedTable, actualTable)
		})
	}
}

func getVectorFloat32(length int, last float32) []float32 {
	vector := make([]float32, length)
	for i := 0; i < length-1; i++ {
		vector[i] = 0.0
	}

	vector[length-1] = last

	return vector
}

func getVectorBool(length int, last int) []bool {
	vector := make([]bool, length)
	for i := 0; i < last; i++ {
		vector[i] = true
	}

	return vector
}

func (suite *CmdTestSuite) TestSuccessfulQueryCmd() {
	suite.CleanUpIndexes(context.Background())
	strIndexName := "query-str-index"
	intIndexName := "query-int-index"
	boolIndexName := "query-bool-index"
	indexes := []*protos.IndexDefinition{
		tests.NewIndexDefinitionBuilder(false,
			strIndexName, "test", 10, protos.VectorDistanceMetric_SQUARED_EUCLIDEAN, "float32-str",
		).Build(),
		tests.NewIndexDefinitionBuilder(false,
			intIndexName, "test", 10, protos.VectorDistanceMetric_SQUARED_EUCLIDEAN, "float32-int",
		).Build(),
		tests.NewIndexDefinitionBuilder(false,
			boolIndexName, "test", 10, protos.VectorDistanceMetric_SQUARED_EUCLIDEAN, "bool",
		).Build(),
	}

	for _, index := range indexes {
		err := suite.AvsClient.IndexCreateFromIndexDef(context.Background(), index)
		if err != nil {
			suite.FailNowf("unable to index create", "%v", err)
		}

		defer suite.AvsClient.IndexDrop(context.Background(), index.Id.Namespace, index.Id.Name)
	}

	type testRecord struct {
		key  any
		data map[string]any
	}

	records := []testRecord{
		{
			key: "a",
			data: map[string]any{
				"str":         "a",
				"int":         1,
				"float":       3.14,
				"float32-str": getVectorFloat32(10, 0.0),
				"map": map[any]any{
					"foo": "bar",
				},
				"extra": "to not display",
			},
		},
		{
			key: "b",
			data: map[string]any{
				"float32-str": getVectorFloat32(10, 1.0),
			},
		},
		{
			key: "c",
			data: map[string]any{
				"float32-str": getVectorFloat32(10, 2.0),
			},
		},
		{
			key: "d",
			data: map[string]any{
				"float32-str": getVectorFloat32(10, 3.0),
			},
		},
		{
			key: "e",
			data: map[string]any{
				"float32-str": getVectorFloat32(10, 4.0),
			},
		},
		{
			key: "f",
			data: map[string]any{
				"float32-str": getVectorFloat32(10, 5.0),
			},
		},
		{
			key: "g",
			data: map[string]any{
				"float32-str": getVectorFloat32(10, 6.0),
			},
		},
		{
			key: "h",
			data: map[string]any{
				"float32-str": getVectorFloat32(10, 7.0),
			},
		},
		{
			key: "i",
			data: map[string]any{
				"float32-str": getVectorFloat32(10, 8.0),
			},
		},
		{
			key: "j",
			data: map[string]any{
				"float32-str": getVectorFloat32(10, 9.0),
			},
		},
		{
			key: 0,
			data: map[string]any{
				"str":         "a",
				"int":         1,
				"float":       3.14,
				"float32-int": getVectorFloat32(10, 0.0),
				"map": map[any]any{
					"foo": "bar",
				},
				"extra": "to not display",
			},
		},
		{
			key: 1,
			data: map[string]any{
				"float32-int": getVectorFloat32(10, 1.0),
			},
		},
		{
			key: 2,
			data: map[string]any{
				"float32-int": getVectorFloat32(10, 2.0),
			},
		},
		{
			key: 3,
			data: map[string]any{
				"float32-int": getVectorFloat32(10, 3.0),
			},
		},
		{
			key: 4,
			data: map[string]any{
				"float32-int": getVectorFloat32(10, 4.0),
			},
		},
		{
			key: 5,
			data: map[string]any{
				"float32-int": getVectorFloat32(10, 5.0),
			},
		},
		{
			key: 6,
			data: map[string]any{
				"float32-int": getVectorFloat32(10, 6.0),
			},
		},
		{
			key: 7,
			data: map[string]any{
				"float32-int": getVectorFloat32(10, 7.0),
			},
		},
		{
			key: 8,
			data: map[string]any{
				"float32-int": getVectorFloat32(10, 8.0),
			},
		},
		{
			key: 9,
			data: map[string]any{
				"float32-int": getVectorFloat32(10, 9.0),
			},
		},
		{
			key: 10,
			data: map[string]any{
				"bool": getVectorBool(10, 7.0),
			},
		},
		{
			key: 11,
			data: map[string]any{
				"bool": getVectorBool(10, 8.0),
			},
		},
		{
			key: 12,
			data: map[string]any{
				"bool": getVectorBool(10, 9.0),
			},
		},
	}

	for _, record := range records {
		suite.AvsClient.Upsert(
			context.Background(),
			"test",
			nil,
			record.key,
			record.data,
			false,
		)
	}

	suite.AvsClient.WaitForIndexCompletion(context.Background(), "test", strIndexName, 12*time.Second)
	suite.AvsClient.WaitForIndexCompletion(context.Background(), "test", intIndexName, 12*time.Second)
	suite.AvsClient.WaitForIndexCompletion(context.Background(), "test", boolIndexName, 12*time.Second)

	testCases := []struct {
		name          string
		index         *protos.IndexDefinition
		records       []testRecord
		cmd           string
		expectedTable string
	}{
		{
			name:    "run query with zero vector",
			records: records,
			cmd:     fmt.Sprintf("query -i %s -n test --max-results 3 --fields str,int,float,float32-str,map --no-color --format 1", strIndexName),
			expectedTable: `Query Results
,Namespace,Key,Distance,Generation,Data
1,test,a,0,0,"Key\,Value
float\,3.14
float32-str\,\"[0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0]\"
int\,1
map\,map[foo:bar]
str\,a"
2,test,b,1,0,"Key\,Value
float32-str\,\"[0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,1.0]\""
3,test,c,4,0,"Key\,Value
float32-str\,\"[0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,2.0]\""
`,
		},
		{
			name:    "run query with custom float32 vector",
			records: records,
			cmd:     fmt.Sprintf("query -i %s -n test --vector [0.0,0.0,0.0,0.0,0.0,0.0,0.0,0.0,0.0,1.0]  --no-color --format 1", strIndexName),
			expectedTable: `Query Results
,Namespace,Key,Distance,Generation,Data
1,test,b,0,0,"Key\,Value
float32-str\,\"[0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,1.0]\""
2,test,a,1,0,"Key\,Value
extra\,to not display
float\,3.14
float32-str\,\"[0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0]\"
int\,1
map\,map[foo:bar]
...\,..."
3,test,c,1,0,"Key\,Value
float32-str\,\"[0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,2.0]\""
4,test,d,4,0,"Key\,Value
float32-str\,\"[0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,3.0]\""
5,test,e,9,0,"Key\,Value
float32-str\,\"[0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,4.0]\""
Hint: To increase the number of records returned, use the --max-results flag.
Hint: To choose which record keys are displayed, use the --fields flag. By default only 5 are displayed.
`,
		},
		{
			name:    "run query with custom bool vector",
			records: records,
			cmd:     fmt.Sprintf("query -i %s -n test --vector [0,0,0,0,0,0,0,0,0,1]  --no-color --format 1", boolIndexName),
			expectedTable: `Query Results
,Namespace,Key,Distance,Generation,Data
1,test,10,8,0,"Key\,Value
bool\,\"[1\\,1\\,1\\,1\\,1\\,1\\,1\\,0\\,0\\,0]\""
2,test,11,9,0,"Key\,Value
bool\,\"[1\\,1\\,1\\,1\\,1\\,1\\,1\\,1\\,0\\,0]\""
3,test,12,10,0,"Key\,Value
bool\,\"[1\\,1\\,1\\,1\\,1\\,1\\,1\\,1\\,1\\,0]\""
Hint: To increase the number of records returned, use the --max-results flag.
Hint: To choose which record keys are displayed, use the --fields flag. By default only 5 are displayed.
`,
		}, {
			name:    "run query with using int key with bool vector",
			records: records,
			cmd:     fmt.Sprintf("query -i %s -n test --key-int 10  --no-color --format 1", boolIndexName),
			expectedTable: `Query Results
,Namespace,Key,Distance,Generation,Data
1,test,11,1,0,"Key\,Value
bool\,\"[1\\,1\\,1\\,1\\,1\\,1\\,1\\,1\\,0\\,0]\""
2,test,12,2,0,"Key\,Value
bool\,\"[1\\,1\\,1\\,1\\,1\\,1\\,1\\,1\\,1\\,0]\""
Hint: To increase the number of records returned, use the --max-results flag.
Hint: To choose which record keys are displayed, use the --fields flag. By default only 5 are displayed.
`,
		},
		{
			name:    "run query with using str key",
			records: records,
			cmd:     fmt.Sprintf("query -i %s -n test -k b --no-color --format 1", strIndexName),
			expectedTable: `Query Results
,Namespace,Key,Distance,Generation,Data
1,test,a,1,0,"Key\,Value
extra\,to not display
float\,3.14
float32-str\,\"[0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0]\"
int\,1
map\,map[foo:bar]
...\,..."
2,test,c,1,0,"Key\,Value
float32-str\,\"[0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,2.0]\""
3,test,d,4,0,"Key\,Value
float32-str\,\"[0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,3.0]\""
4,test,e,9,0,"Key\,Value
float32-str\,\"[0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,4.0]\""
5,test,f,16,0,"Key\,Value
float32-str\,\"[0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,5.0]\""
Hint: To increase the number of records returned, use the --max-results flag.
Hint: To choose which record keys are displayed, use the --fields flag. By default only 5 are displayed.
`,
		},
		{
			name:    "run query with using int key",
			records: records,
			cmd:     fmt.Sprintf("query -i %s -n test -t 1 --no-color --format 1", intIndexName),
			expectedTable: `Query Results
,Namespace,Key,Distance,Generation,Data
1,test,0,1,0,"Key\,Value
extra\,to not display
float\,3.14
float32-int\,\"[0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0]\"
int\,1
map\,map[foo:bar]
...\,..."
2,test,2,1,0,"Key\,Value
float32-int\,\"[0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,2.0]\""
3,test,3,4,0,"Key\,Value
float32-int\,\"[0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,3.0]\""
4,test,4,9,0,"Key\,Value
float32-int\,\"[0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,4.0]\""
5,test,5,16,0,"Key\,Value
float32-int\,\"[0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,5.0]\""
Hint: To increase the number of records returned, use the --max-results flag.
Hint: To choose which record keys are displayed, use the --fields flag. By default only 5 are displayed.
`,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			actualTable, stderr, err := suite.RunSuiteCmd(strings.Split(tc.cmd, " ")...)
			suite.Assert().NoError(err, "error: %s, stdout: %s, stderr: %s", err, actualTable, stderr)

			suite.Assert().Equal(tc.expectedTable, actualTable)
		})
	}
}

func (suite *CmdTestSuite) TestFailedQueryCmd() {
	namespace := "test"
	indexName := "index"
	suite.AvsClient.IndexCreate(
		context.Background(),
		namespace,
		indexName,
		"field",
		10,
		protos.VectorDistanceMetric_COSINE,
		nil,
	)

	testCases := []struct {
		name           string
		cmd            string
		expectedErrStr string
	}{
		{
			name:           "use seeds and hosts together",
			cmd:            "query --host 1.1.1.1:3001 --seeds 2.2.2.2:3000 -n test -i i",
			expectedErrStr: "Error: only --seeds or --host allowed",
		},
		{
			name:           "use set without the key flag",
			cmd:            "query --namespace test -i index --set testset",
			expectedErrStr: "Warning: The --set flag is only used when the --key-str or --key-int flag is set.",
		},
		{
			name:           "try to query an index that does not exist",
			cmd:            "query --namespace test -i DNE",
			expectedErrStr: "Error: Failed to get index definition: failed to get index: server error: NotFound, msg=index test:DNE not found",
		},
		{
			name:           "try to query a key that does not exist",
			cmd:            fmt.Sprintf("query --namespace %s -i %s -k DNE", namespace, indexName),
			expectedErrStr: "Error: Failed to get vector using key: unable to get record: failed to get record: server error: NotFound",
		},
		{
			name:           "query a key without a set and check for prompt",
			cmd:            fmt.Sprintf("query --namespace %s -i %s -k DNE", namespace, indexName),
			expectedErrStr: "Warning: The requested record was not found. If the record is in a set, you may also need to provide the --set flag.",
		},
		{
			name:           "query a key without a set and check for prompt",
			cmd:            fmt.Sprintf("query --namespace %s -i %s -t 1234", namespace, indexName),
			expectedErrStr: "Warning: The requested record was not found. If the record is in a set, you may also need to provide the --set flag.",
		},
		{
			name:           "query using an invalid int key",
			cmd:            fmt.Sprintf("query --namespace %s -i %s -t DNE", namespace, indexName),
			expectedErrStr: "Error: invalid argument \"DNE\" for \"-t, --key-int\" flag: strconv.ParseInt: parsing \"DNE\": invalid syntax",
		},
		{
			name:           "query using key-int and key-str together",
			cmd:            fmt.Sprintf("query --namespace %s -i %s --key-str DNA --key-int 1", namespace, indexName),
			expectedErrStr: "Error: if any flags in the group [vector key-str key-int] are set none of the others can be; [key-int key-str] were all set",
		},
		{
			name:           "query using key-str and vector together",
			cmd:            fmt.Sprintf("query --namespace %s -i %s --key-str DNA --vector [0,1,1,1]", namespace, indexName),
			expectedErrStr: "Error: if any flags in the group [vector key-str key-int] are set none of the others can be; [key-str vector] were all set",
		},
		{
			name:           "query using key-int and vector together",
			cmd:            fmt.Sprintf("query --namespace %s -i %s --key-int 1 --vector [0,1,1,1]", namespace, indexName),
			expectedErrStr: "Error: if any flags in the group [vector key-str key-int] are set none of the others can be; [key-int vector] were all set",
		},
		{
			name:           "query using key-int and vector together",
			cmd:            fmt.Sprintf("query --namespace %s -i %s --vector [0,1,1,1]", namespace, indexName),
			expectedErrStr: "Error: Failed to get vector using vector: failed to receive all neighbors: rpc error: code = InvalidArgument desc = dimension mismatch, required 10, actual 4",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			_, lines, err := suite.RunSuiteCmd(strings.Split(tc.cmd, " ")...)

			suite.Assert().Error(err, "error: %s, stdout/err: %s", err, lines)
			suite.Assert().Contains(lines, tc.expectedErrStr)
		})
	}

}

func (suite *CmdTestSuite) TestFailUserCmdsWithInvalidUser() {
	suite.SkipIfUserPassAuthDisabled()

	testCases := []struct {
		name        string
		cmd         string
		expectedErr string
	}{
		{
			name:        "fail to revoke user to invalid user",
			cmd:         "users revoke --name foo1 --roles admin",
			expectedErr: "failed to revoke user roles: server error: NotFound",
		},
		{
			name:        "fail to grant user to invalid user",
			cmd:         "users grant --name foo1 --roles admin",
			expectedErr: "failed to grant user roles: server error: NotFound",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			_, lines, err := suite.RunSuiteCmd(strings.Split(tc.cmd, " ")...)
			suite.Assert().Error(err, "error: %s, stdout/err: %s", err, lines)
			suite.Assert().Contains(lines, tc.expectedErr)
		})

	}
}

func (suite *CmdTestSuite) TestFailUserCmdsWithInvalidRoles() {
	suite.SkipIfUserPassAuthDisabled()

	testCases := []struct {
		name        string
		cmd         string
		expectedErr string
	}{
		{
			name:        "fail to grant user with invalid role",
			cmd:         "users grant --name foo1 --roles invalid",
			expectedErr: "unknown roles [invalid]",
		},
		{
			name:        "fail to revoke user with invalid role",
			cmd:         "users revoke --name foo1 --roles invalid",
			expectedErr: "unknown roles [invalid]",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			_, lines, err := suite.RunSuiteCmd(strings.Split(tc.cmd, " ")...)
			suite.Assert().Error(err, "error: %s, stdout/err: %s", err, lines)
			suite.Assert().Contains(lines, tc.expectedErr)
		})

	}
}

func (suite *CmdTestSuite) TestSuccessfulListRolesCmd() {
	suite.SkipIfUserPassAuthDisabled()

	testCases := []struct {
		name          string
		cmd           string
		expectedTable string
	}{
		{
			"roles list",
			"role list --format 1",
			`,Roles
1,admin
2,read-write
`,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			actualTable, _, err := suite.RunSuiteCmd(strings.Split(tc.cmd, " ")...)
			suite.Assert().NoError(err, "error: %s, stdout/err: %s", err, actualTable)
			suite.Assert().Equal(tc.expectedTable, actualTable)
		})
	}
}

func (suite *CmdTestSuite) TestFailInvalidArg() {
	testCases := []struct {
		name           string
		cmd            string
		expectedErrStr string
	}{
		{
			name:           "use seeds and hosts together",
			cmd:            "index create -y --seeds 2.2.2.2:3000 --host 1.1.1.1:3001 -n test -i index1 -d 256 -m SQUARED_EUCLIDEAN --vector-field vector1 --storage-namespace bar --storage-set testbar",
			expectedErrStr: "Error: only --seeds or --host allowed",
		},
		{
			name:           "use seeds and hosts together",
			cmd:            "index list --host 1.1.1.1:3001 --seeds 2.2.2.2:3000",
			expectedErrStr: "Error: only --seeds or --host allowed",
		},
		{
			name:           "use seeds and hosts together",
			cmd:            "index drop -y --host 1.1.1.1:3001 --seeds 2.2.2.2:3000 -n test -i index1",
			expectedErrStr: "Error: only --seeds or --host allowed",
		},
		{
			name:           "use seeds and hosts together",
			cmd:            "index gc --host 1.1.1.1:3001 --seeds 2.2.2.2:3000 -n test -i index1",
			expectedErrStr: "Error: only --seeds or --host allowed",
		},
		{
			name:           "error because no create index required args are provided",
			cmd:            fmt.Sprintf("index create --seeds %s", suite.AvsHostPort.String()),
			expectedErrStr: "Error: required flag(s) \"dimension\", \"distance-metric\", \"index-name\", \"namespace\", \"vector-field\" not set",
		},
		{
			name:           "error because no create index required args are provided",
			cmd:            fmt.Sprintf("index create"),
			expectedErrStr: "Error: required flag(s) \"dimension\", \"distance-metric\", \"index-name\", \"namespace\", \"vector-field\" not set",
		},
		{
			name:           "error because some create index required args are not provided",
			cmd:            fmt.Sprintf("index create --seeds %s --dimension 10", suite.AvsHostPort.String()),
			expectedErrStr: "Error: required flag(s) \"distance-metric\", \"index-name\", \"namespace\", \"vector-field\" not set",
		},
		{
			name:           "use seeds and hosts together",
			cmd:            "user create --host 1.1.1.1:3001 --seeds 2.2.2.2:3000 --name foo --roles admin",
			expectedErrStr: "Error: only --seeds or --host allowed",
		},
		{
			name:           "use seeds and hosts together",
			cmd:            "user drop --host 1.1.1.1:3001 --seeds 2.2.2.2:3000 --name foo",
			expectedErrStr: "Error: only --seeds or --host allowed",
		},
		{
			name:           "use seeds and hosts together",
			cmd:            "user ls --host 1.1.1.1:3001 --seeds 2.2.2.2:3000",
			expectedErrStr: "Error: only --seeds or --host allowed",
		},
		{
			name:           "use seeds and hosts together",
			cmd:            "roles ls --host 1.1.1.1:3001 --seeds 2.2.2.2:3000",
			expectedErrStr: "Error: only --seeds or --host allowed",
		},
		{
			name:           "use seeds and hosts together",
			cmd:            "nodes ls --host 1.1.1.1:3001 --seeds 2.2.2.2:3000",
			expectedErrStr: "Error: only --seeds or --host allowed",
		},
		{
			name:           "test with bad dimension",
			cmd:            "index create -y --host 1.1.1.1:3001  -n test -i index1 -d -1 -m SQUARED_EUCLIDEAN --vector-field vector1 --storage-namespace bar --storage-set testbar",
			expectedErrStr: "Error: invalid argument \"-1\" for \"-d, --dimension\"",
		},
		{
			name:           "test with bad distance metric",
			cmd:            "index create -y --host 1.1.1.1:3001  -n test -i index1 -d 10 -m BAD --vector-field vector1 --storage-namespace bar --storage-set testbar",
			expectedErrStr: "Error: invalid argument \"BAD\" for \"-m, --distance-metric\"",
		},
		{
			name:           "test with bad timeout",
			cmd:            "index create -y --host 1.1.1.1:3001  -n test -i index1 -d 10 -m SQUARED_EUCLIDEAN --vector-field vector1 --storage-namespace bar --storage-set testbar --timeout 10",
			expectedErrStr: "Error: invalid argument \"10\" for \"--timeout\"",
		},
		{
			name:           "test with bad hnsw-ef",
			cmd:            "index create -y --hnsw-ef foo --host 1.1.1.1:3001  -n test -i index1 -d 10 -m SQUARED_EUCLIDEAN --vector-field vector1 --storage-namespace bar --storage-set testbar ",
			expectedErrStr: "Error: invalid argument \"foo\" for \"--hnsw-ef\"",
		},
		{
			name:           "test with bad hnsw-ef-construction",
			cmd:            "index create -y --hnsw-ef-construction foo --host 1.1.1.1:3001  -n test -i index1 -d 10 -m SQUARED_EUCLIDEAN --vector-field vector1 --storage-namespace bar --storage-set testbar ",
			expectedErrStr: "Error: invalid argument \"foo\" for \"--hnsw-ef-construction\"",
		},
		{
			name:           "test with bad hnsw-m",
			cmd:            "index create -y --hnsw-m foo --host 1.1.1.1:3001  -n test -i index1 -d 10 -m SQUARED_EUCLIDEAN --vector-field vector1 --storage-namespace bar --storage-set testbar ",
			expectedErrStr: "Error: invalid argument \"foo\" for \"--hnsw-m\"",
		},
		{
			name:           "test with bad hnsw-batch-interval",
			cmd:            "index create -y --hnsw-batch-index-interval foo --host 1.1.1.1:3001  -n test -i index1 -d 10 -m SQUARED_EUCLIDEAN --vector-field vector1 --storage-namespace bar --storage-set testbar ",
			expectedErrStr: "Error: invalid argument \"foo\" for \"--hnsw-batch-index-interval\"",
		},
		{
			name:           "test with bad hnsw-batch-max-records",
			cmd:            "index create -y --hnsw-batch-max-index-records foo --host 1.1.1.1:3001  -n test -i index1 -d 10 -m SQUARED_EUCLIDEAN --vector-field vector1 --storage-namespace bar --storage-set testbar ",
			expectedErrStr: "Error: invalid argument \"foo\" for \"--hnsw-batch-max-index-records\"",
		},
		{
			name:           "test with bad hnsw-index-cache-max-entries",
			cmd:            "index create -y --hnsw-index-cache-max-entries foo --host 1.1.1.1:3001  -n test -i index1 -d 10 -m SQUARED_EUCLIDEAN --vector-field vector1 --storage-namespace bar --storage-set testbar ",
			expectedErrStr: "Error: invalid argument \"foo\" for \"--hnsw-index-cache-max-entries\"",
		},
		{
			name:           "test with bad hnsw-index-cache-expiry",
			cmd:            "index create -y --hnsw-index-cache-expiry 10 --host 1.1.1.1:3001  -n test -i index1 -d 10 -m SQUARED_EUCLIDEAN --vector-field vector1 --storage-namespace bar --storage-set testbar ",
			expectedErrStr: "Error: invalid argument \"10\" for \"--hnsw-index-cache-expiry\"",
		},
		{
			name:           "test with bad hnsw-record-cache-max-entries",
			cmd:            "index create -y --hnsw-record-cache-max-entries foo --host 1.1.1.1:3001  -n test -i index1 -d 10 -m SQUARED_EUCLIDEAN --vector-field vector1 --storage-namespace bar --storage-set testbar ",
			expectedErrStr: "Error: invalid argument \"foo\" for \"--hnsw-record-cache-max-entries\"",
		},
		{
			name:           "test with bad hnsw-record-cache-expiry",
			cmd:            "index create -y --hnsw-record-cache-expiry 10 --host 1.1.1.1:3001  -n test -i index1 -d 10 -m SQUARED_EUCLIDEAN --vector-field vector1 --storage-namespace bar --storage-set testbar ",
			expectedErrStr: "Error: invalid argument \"10\" for \"--hnsw-record-cache-expiry\"",
		},
		{
			name:           "test with bad hnsw-healer-max-scan-rate-per-node",
			cmd:            "index create -y --hnsw-healer-max-scan-rate-per-node foo --host 1.1.1.1:3001  -n test -i index1 -d 10 -m SQUARED_EUCLIDEAN --vector-field vector1 --storage-namespace bar --storage-set testbar ",
			expectedErrStr: "Error: invalid argument \"foo\" for \"--hnsw-healer-max-scan-rate-per-node\"",
		},
		{
			name:           "test with bad hnsw-healer-max-scan-page-size",
			cmd:            "index create -y --hnsw-healer-max-scan-page-size foo --host 1.1.1.1:3001  -n test -i index1 -d 10 -m SQUARED_EUCLIDEAN --vector-field vector1 --storage-namespace bar --storage-set testbar ",
			expectedErrStr: "Error: invalid argument \"foo\" for \"--hnsw-healer-max-scan-page-size\"",
		},
		{
			name:           "test with bad hnsw-healer-reindex-percent",
			cmd:            "index create -y --hnsw-healer-reindex-percent foo --host 1.1.1.1:3001  -n test -i index1 -d 10 -m SQUARED_EUCLIDEAN --vector-field vector1 --storage-namespace bar --storage-set testbar ",
			expectedErrStr: "Error: invalid argument \"foo\" for \"--hnsw-healer-reindex-percent\"",
		},
		{
			name:           "test with bad hnsw-healer-parallelism",
			cmd:            "index create -y --hnsw-healer-parallelism foo --host 1.1.1.1:3001  -n test -i index1 -d 10 -m SQUARED_EUCLIDEAN --vector-field vector1 --storage-namespace bar --storage-set testbar ",
			expectedErrStr: "Error: invalid argument \"foo\" for \"--hnsw-healer-parallelism\"",
		},
		{
			name:           "test with bad hnsw-merge-index-parallelism",
			cmd:            "index create -y --hnsw-merge-index-parallelism foo --host 1.1.1.1:3001  -n test -i index1 -d 10 -m SQUARED_EUCLIDEAN --vector-field vector1 --storage-namespace bar --storage-set testbar ",
			expectedErrStr: "Error: invalid argument \"foo\" for \"--hnsw-merge-index-parallelism\"",
		},
		{
			name:           "test with bad tls-cafile",
			cmd:            "user create --tls-cafile blah --name foo --roles admin",
			expectedErrStr: "blah: no such file or directory",
		},
		{
			name:           "test with bad tls-capath",
			cmd:            "user create --tls-capath blah --name foo --roles admin",
			expectedErrStr: "blah: no such file or directory",
		},
		{
			name:           "test with bad tls-certfile",
			cmd:            "user create --tls-certfile blah --name foo --roles admin",
			expectedErrStr: "blah: no such file or directory",
		},
		{
			name:           "test with bad tls-keyfile",
			cmd:            "user create --tls-keyfile blah --name foo --roles admin",
			expectedErrStr: "blah: no such file or directory",
		},
		{
			name:           "test with bad tls-keyfile-password",
			cmd:            "user create --tls-keyfile-password b64:bla65asdf54r345123!@#$h --name foo --roles admin",
			expectedErrStr: "Error: invalid argument \"b64:bla65asdf54r345123!@#$h\"",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			_, lines, err := suite.RunCmd(strings.Split(tc.cmd, " ")...)

			suite.Assert().Error(err, "error: %s, stdout/err: %s", err, lines)
			suite.Assert().Contains(lines, tc.expectedErrStr)
		})
	}
}

func (suite *CmdTestSuite) TestConfigFile() {
	configFile := "tests/asvec_.yml"
	cmd := fmt.Sprintf("index ls --log-level debug --config-file %s --cluster-name %s", configFile, suite.configFileClusterName)

	stdout, stderr, err := suite.RunCmd(strings.Split(cmd, " ")...)

	suite.NoError(err, "err: %s, stdout: %s, stderr: %s", err, stdout, stderr)
}

func (suite *CmdTestSuite) TestEnvVars() {
	convertArgsToEnvs := func(args []string) []string {
		envs := make([]string, 0)
		for _, arg := range args {
			key_val := strings.Split(arg, " ")

			if len(key_val) != 2 {
				continue
			}

			key := key_val[0]
			val := key_val[1]

			key = strings.Replace(key, "--", "ASVEC_", 1)
			key = strings.ReplaceAll(key, "-", "_")
			key = strings.ToUpper(key)
			envs = append(envs, key+"="+val)
		}

		return envs
	}

	envs := convertArgsToEnvs(suite.SuiteFlags)
	suite.Logger.Debug("suite flags", slog.Any("env", envs))

	cmd := suite.GetCmd(strings.Split("index ls --log-level debug", " ")...)
	cmd.Env = append(cmd.Env, envs...)
	stdout, stderr, err := suite.GetCmdOutput(cmd)

	suite.NoError(err, "err: %s, stdout: %s, stderr: %s", err, stdout, stderr)
}

func (suite *CmdTestSuite) TestTLSHostnameOverride_Success() {
	if suite.AvsTLSConfig == nil {
		suite.T().Skip("Not a TLS suite")
	}

	newSuiteFlags := []string{}
	for _, flag := range suite.SuiteFlags {
		if strings.Contains(flag, tests.CreateFlagStr(flags.Host, "")) || strings.Contains(flag, tests.CreateFlagStr(flags.Seeds, "")) {
			flagSplit := strings.Split(flag, " ")

			flagSplit[1] = "127.0.0.1:10000" // For tls the certs only work with localhost not 127.0.0.1
			flag = strings.Join(flagSplit, " ")
		}

		newSuiteFlags = append(newSuiteFlags, flag)
	}

	newSuiteFlags = append(newSuiteFlags, tests.CreateFlagStr(flags.TLSHostnameOverride, "localhost"))

	suite.Logger.Debug("suite flags", slog.Any("flags", newSuiteFlags))
	asvecCmd := strings.Split("index ls --log-level debug --timeout 10s", " ")
	asvecCmd = append(asvecCmd, strings.Split(strings.Join(newSuiteFlags, " "), " ")...)

	stdout, stderr, err := suite.RunCmd(asvecCmd...)

	suite.NoError(err, "err: %s, stdout: %s, stderr: %s", err, stdout, stderr)
}

func (suite *CmdTestSuite) TestTLSHostnameOverride_Failure() {
	if suite.AvsTLSConfig == nil {
		suite.T().Skip("Not a TLS suite")
	}

	newSuiteFlags := []string{}
	for _, flag := range suite.SuiteFlags {
		if strings.Contains(flag, tests.CreateFlagStr(flags.Host, "")) || strings.Contains(flag, tests.CreateFlagStr(flags.Seeds, "")) {
			flagSplit := strings.Split(flag, " ")

			flagSplit[1] = "127.0.0.1:10000" // For tls the certs only work with localhost not 127.0.0.1
			flag = strings.Join(flagSplit, " ")
		}

		newSuiteFlags = append(newSuiteFlags, flag)
	}

	suite.Logger.Debug("suite flags", slog.Any("flags", newSuiteFlags))
	asvecCmd := strings.Split("index ls --log-level debug --timeout 10s", " ")
	asvecCmd = append(asvecCmd, strings.Split(strings.Join(newSuiteFlags, " "), " ")...)

	stdout, stderr, err := suite.RunCmd(asvecCmd...)
	suite.Error(err, "err: %s, stdout: %s, stderr: %s", err, stdout, stderr)
	suite.Contains(stdout, "Hint: Failed to verify because of certificate hostname mismatch.")
	suite.Contains(stdout, "Hint: Either correctly set your certificate SAN or use")
}
