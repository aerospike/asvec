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
		expected_index *protos.IndexDefinition
	}{
		{
			"test with labels",
			"index0",
			"test",
			"index create -y -n test -i index0 -d 256 -m SQUARED_EUCLIDEAN --vector-field vector0 --index-labels model=all-MiniLM-L6-v2,foo=bar",
			tests.NewIndexDefinitionBuilder("index0", "test", 256, protos.VectorDistanceMetric_SQUARED_EUCLIDEAN, "vector0").
				WithLabels(map[string]string{"model": "all-MiniLM-L6-v2", "foo": "bar"}).
				Build(),
		},
		{
			"test with storage config",
			"index1",
			"test",
			"index create -y -n test -i index1 -d 256 -m SQUARED_EUCLIDEAN --vector-field vector1 --storage-namespace bar --storage-set testbar",
			tests.NewIndexDefinitionBuilder("index1", "test", 256, protos.VectorDistanceMetric_SQUARED_EUCLIDEAN, "vector1").
				WithStorageNamespace("bar").
				WithStorageSet("testbar").
				Build(),
		},
		{
			"test with hnsw params and seeds",
			"index2",
			"test",
			"index create -y -n test -i index2 -d 256 -m HAMMING --vector-field vector2 --hnsw-max-edges 10 --hnsw-ef 11 --hnsw-ef-construction 12 --hnsw-max-mem-queue-size 10",
			tests.NewIndexDefinitionBuilder("index2", "test", 256, protos.VectorDistanceMetric_HAMMING, "vector2").
				WithHnswM(10).
				WithHnswEf(11).
				WithHnswEfConstruction(12).
				WithHnswMaxMemQueueSize(10).
				Build(),
		},
		{
			"test with hnsw batch params",
			"index3",
			"test",
			"index create -y -n test -i index3 -d 256 -m COSINE --vector-field vector3 --hnsw-batch-interval 50s --hnsw-batch-max-records 100",
			tests.NewIndexDefinitionBuilder("index3", "test", 256, protos.VectorDistanceMetric_COSINE, "vector3").
				WithHnswBatchingMaxRecord(100).
				WithHnswBatchingInterval(50000).
				Build(),
		},
		{
			"test with hnsw cache params",
			"index4",
			"test",
			"index create -y -n test -i index4 -d 256 -m COSINE --vector-field vector4 --hnsw-cache-max-entries 1000 --hnsw-cache-expiry 10s",
			tests.NewIndexDefinitionBuilder("index4", "test", 256, protos.VectorDistanceMetric_COSINE, "vector4").
				WithHnswCacheExpiry(10000).
				WithHnswCacheMaxEntries(1000).
				Build(),
		},
		{
			"test with hnsw healer params",
			"index5",
			"test",
			"index create -y -n test -i index5 -d 256 -m COSINE --vector-field vector5 --hnsw-healer-max-scan-rate-per-node 1000 --hnsw-healer-max-scan-page-size 1000 --hnsw-healer-reindex-percent 10.10 --hnsw-healer-schedule-delay 10s --hnsw-healer-parallelism 10",
			tests.NewIndexDefinitionBuilder("index5", "test", 256, protos.VectorDistanceMetric_COSINE, "vector5").
				WithHnswHealerMaxScanRatePerNode(1000).
				WithHnswHealerMaxScanPageSize(1000).
				WithHnswHealerReindexPercent(10.10).
				WithHnswHealerScheduleDelay(10000).
				WithHnswHealerParallelism(10).
				Build(),
		},
		{
			"test with hnsw merge params",
			"index6",
			"test",
			"index create -y -n test -i index6 -d 256 -m COSINE --vector-field vector6 --hnsw-merge-parallelism 10",
			tests.NewIndexDefinitionBuilder("index6", "test", 256, protos.VectorDistanceMetric_COSINE, "vector6").
				WithHnswMergeParallelism(10).
				Build(),
		},
		{
			"test with yaml file",
			"yaml-file-index",
			"test",
			fmt.Sprintf("index create -y --file tests/indexDef.yaml"),
			tests.NewIndexDefinitionBuilder("yaml-file-index", "test", 10, protos.VectorDistanceMetric_COSINE, "vector").
				WithSet("testset").
				WithHnswEf(101).
				WithHnswEfConstruction(102).
				WithHnswM(103).
				WithHnswMaxMemQueueSize(10004).
				WithHnswBatchingInterval(30001).
				WithHnswBatchingMaxRecord(100001).
				WithHnswCacheMaxEntries(1001).
				WithHnswCacheExpiry(1002).
				WithHnswHealerParallelism(7).
				WithHnswHealerMaxScanRatePerNode(1).
				WithHnswHealerMaxScanPageSize(2).
				WithHnswHealerReindexPercent(3).
				WithHnswHealerScheduleDelay(4).
				WithHnswMergeParallelism(7).
				WithStorageSet("name").
				Build(),
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			lines, _, err := suite.RunSuiteCmd(strings.Split(tc.cmd, " ")...)

			if err != nil {
				suite.Assert().NoError(err, "error: %s, stdout/err: %s", err, lines)
				suite.FailNow("unable to index create")
			}

			actual, err := suite.AvsClient.IndexGet(context.Background(), tc.indexNamespace, tc.indexName)

			if err != nil {
				suite.FailNowf("unable to get index", "%v", err)
			}

			suite.EqualExportedValues(tc.expected_index, actual)
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
			"test with all indexes succeed",
			[]*protos.IndexDefinition{
				tests.NewIndexDefinitionBuilder(
					"exists1", "test", 256, protos.VectorDistanceMetric_COSINE, "vector",
				).Build(),
				tests.NewIndexDefinitionBuilder(
					"exists2", "bar", 256, protos.VectorDistanceMetric_HAMMING, "vector",
				).WithSet("barset").Build(),
			},
			"s/exists/does-not-exist-yet/g",
			false,
			[]string{
				"Successfully created index test.*.does-not-exist-yet",
				"Successfully created index bar.barset.does-not-exist-yet",
				"Successfully created all indexes from yaml",
			},
		},
		{
			"test with one index that fails",
			[]*protos.IndexDefinition{
				tests.NewIndexDefinitionBuilder(
					"exists3", "test", 256, protos.VectorDistanceMetric_COSINE, "vector",
				).Build(),
				tests.NewIndexDefinitionBuilder(
					"exists4", "bar", 256, protos.VectorDistanceMetric_HAMMING, "vector",
				).WithSet("barset").Build(),
			},
			"s/exists3/does-not-exist-yet2/g",
			true,
			[]string{
				"Successfully created index test.*.does-not-exist-yet2",
				"Failed to create index bar.barset.exists4",
				"Some indexes failed to be created",
			},
		},
		{
			"test with no index successfully created",
			[]*protos.IndexDefinition{
				tests.NewIndexDefinitionBuilder(
					"exists1", "test", 256, protos.VectorDistanceMetric_COSINE, "vector",
				).Build(),
				tests.NewIndexDefinitionBuilder(
					"exists2", "bar", 256, protos.VectorDistanceMetric_HAMMING, "vector",
				).WithSet("barset").Build(),
			},
			"s/COSINE/HAMMING/g", // Doing this rather than removing sed for this test case
			true,
			[]string{
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
	suite.AvsClient.IndexCreate(context.Background(), "test", "successful-update", "field", uint32(256), protos.VectorDistanceMetric_COSINE, nil)
	ns := "test"
	index := "successful-update"
	builder := tests.NewIndexDefinitionBuilder(index, ns, 256, protos.VectorDistanceMetric_COSINE, "field")
	testCases := []struct {
		name           string
		indexName      string // index names must be unique for the suite
		indexNamespace string
		cmd            string
		expected_index *protos.IndexDefinition
	}{
		{
			"test with hnsw params and seeds",
			"successful-update",
			ns,
			"index update -y -n test -i successful-update --index-labels new-label=foo --hnsw-max-mem-queue-size 10",
			builder.
				WithLabels(map[string]string{"new-label": "foo"}).
				WithHnswMaxMemQueueSize(10).
				Build(),
		},
		{
			"test with hnsw batch params",
			"successful-update",
			"test",
			"index update -y -n test -i successful-update --hnsw-batch-interval 50s --hnsw-batch-max-records 100",
			builder.
				WithHnswBatchingMaxRecord(100).
				WithHnswBatchingInterval(50000).
				Build(),
		},
		{
			"test with hnsw cache params",
			"successful-update",
			"test",
			"index update -y -n test -i successful-update --hnsw-cache-max-entries 1000 --hnsw-cache-expiry 10s",
			builder.
				WithHnswCacheExpiry(10000).
				WithHnswCacheMaxEntries(1000).
				Build(),
		},
		{
			"test with hnsw healer params",
			"successful-update",
			"test",
			"index update -y -n test -i successful-update --hnsw-healer-max-scan-rate-per-node 1000 --hnsw-healer-max-scan-page-size 1000 --hnsw-healer-reindex-percent 10.10 --hnsw-healer-schedule-delay 10s --hnsw-healer-parallelism 10",
			builder.
				WithHnswHealerMaxScanRatePerNode(1000).
				WithHnswHealerMaxScanPageSize(1000).
				WithHnswHealerReindexPercent(10.10).
				WithHnswHealerScheduleDelay(10000).
				WithHnswHealerParallelism(10).
				Build(),
		},
		{
			"test with hnsw merge params",
			"successful-update",
			"test",
			"index update -y -n test -i successful-update --hnsw-merge-parallelism 10",
			builder.
				WithHnswMergeParallelism(10).
				Build(),
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			lines, _, err := suite.RunSuiteCmd(strings.Split(tc.cmd, " ")...)

			if err != nil {
				suite.Assert().NoError(err, "error: %s, stdout/err: %s", err, lines)
				suite.FailNow("unable to index update")
			}

			time.Sleep(5 * time.Second)

			actual, err := suite.AvsClient.IndexGet(context.Background(), tc.indexNamespace, tc.indexName)

			if err != nil {
				suite.FailNowf("unable to get index", "%v", err)
			}

			suite.EqualExportedValues(tc.expected_index, actual)
		})
	}

}

func (suite *CmdTestSuite) TestUpdateIndexDoesNotExist() {
	_, lines, err := suite.RunSuiteCmd(strings.Split("index update -y -n test -i DNE --hnsw-merge-parallelism 10", " ")...)
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
			"test with hnsw params and seeds",
			"successful-gc",
			ns,
			"index gc -n test -i successful-gc -c 10",
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
			"test with just namespace and seeds",
			"indexdrop1",
			"test",
			nil,
			"index drop -y -n test -i indexdrop1",
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

			_, err = suite.AvsClient.IndexGet(context.Background(), tc.indexNamespace, tc.indexName)

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
			"single index",
			[]*protos.IndexDefinition{
				tests.NewIndexDefinitionBuilder(
					"list", "test", 256, protos.VectorDistanceMetric_COSINE, "vector",
				).Build(),
			},
			"index list --no-color --format 1",
			`Indexes
,Name,Namespace,Field,Dimensions,Distance Metric,Unmerged
1,list,test,vector,256,COSINE,0
`,
		},
		{
			"double index with set",
			[]*protos.IndexDefinition{
				tests.NewIndexDefinitionBuilder(
					"list1", "test", 256, protos.VectorDistanceMetric_COSINE, "vector",
				).Build(),
				tests.NewIndexDefinitionBuilder(
					"list2", "bar", 256, protos.VectorDistanceMetric_HAMMING, "vector",
				).WithSet("barset").Build(),
			},
			"index list --no-color --format 1",
			`Indexes
,Name,Namespace,Set,Field,Dimensions,Distance Metric,Unmerged
1,list2,bar,barset,vector,256,HAMMING,0
2,list1,test,,vector,256,COSINE,0
`,
		},
		{
			"double index with set and verbose",
			[]*protos.IndexDefinition{
				tests.NewIndexDefinitionBuilder(
					"list1", "test", 256, protos.VectorDistanceMetric_COSINE, "vector",
				).WithLabels(map[string]string{"foo": "bar"}).Build(),
				tests.NewIndexDefinitionBuilder(
					"list2", "bar", 256, protos.VectorDistanceMetric_HAMMING, "vector",
				).WithSet("barset").Build(),
			},
			"index list --verbose --no-color --format 1",
			`Indexes
,Name,Namespace,Set,Field,Dimensions,Distance Metric,Unmerged,Labels*,Storage,Index Parameters
1,list2,bar,barset,vector,256,HAMMING,0,map[],"Namespace\,bar
Set\,list2","HNSW
Max Edges\,16
Ef\,100
Construction Ef\,100
MaxMemQueueSize*\,0
Batch Max Records*\,100000
Batch Interval*\,30s
Cache Max Entires*\,0
Cache Expiry*\,0s
Healer Max Scan Rate / Node*\,0
Healer Max Page Size*\,0
Healer Re-index % *\,0.00%
Healer Schedule Delay*\,0s
Healer Parallelism*\,0
Merge Parallelism*\,0"
2,list1,test,,vector,256,COSINE,0,map[foo:bar],"Namespace\,test
Set\,list1","HNSW
Max Edges\,16
Ef\,100
Construction Ef\,100
MaxMemQueueSize*\,0
Batch Max Records*\,100000
Batch Interval*\,30s
Cache Max Entires*\,0
Cache Expiry*\,0s
Healer Max Scan Rate / Node*\,0
Healer Max Page Size*\,0
Healer Re-index % *\,0.00%
Healer Schedule Delay*\,0s
Healer Parallelism*\,0
Merge Parallelism*\,0"
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
		name         string
		cmd          string
		expectedUser *protos.User
	}{
		{
			"create user with comma sep roles",
			"users create --name foo1 --new-password foo --roles admin,read-write",
			&protos.User{
				Username: "foo1",
				Roles: []string{
					"admin",
					"read-write",
				},
			},
		},
		{
			"create user with comma multiple roles",
			"users create --name foo2 --new-password foo --roles admin --roles read-write",
			&protos.User{
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
		name        string
		cmd         string
		expectedErr string
	}{
		{
			"fail to create user with invalid role",
			"users create --name foo1 --new-password foo --roles invalid",
			"unknown roles [invalid]",
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
		name         string
		user         string
		cmd          string
		expectedUser *protos.User
	}{
		{
			"grant user",
			"grant0",
			"users grant --name grant0 --roles read-write",
			&protos.User{
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
		name         string
		user         string
		cmd          string
		expectedUser *protos.User
	}{
		{
			"revoke user",
			"revoke0",
			"users revoke --name revoke0 --roles read-write",
			&protos.User{
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
			"change password",
			"password0",
			"foo",
			"users new-password --name password0 --new-password foo",
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

func (suite *CmdTestSuite) TestSuccessfulListUsersCmd() {
	suite.SkipIfUserPassAuthDisabled()

	testCases := []struct {
		name          string
		cmd           string
		expectedTable string
	}{
		{
			"users list",
			"users list --no-color --format 1",
			`Users
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

	type testRecord struct {
		key  any
		data map[string]any
	}

	key := "a"

	testRecords := []testRecord{
		{
			key: key,
			data: map[string]any{
				"str":     "a",
				"int":     1,
				"float":   3.14,
				"float32": getVectorFloat32(10, 0.0),
				"map": map[any]any{
					"foo": "bar",
				},
				"extra": "to not display",
			},
		},
		{
			key: "b",
			data: map[string]any{
				"float32": getVectorFloat32(10, 1.0),
			},
		},
		{
			key: "c",
			data: map[string]any{
				"float32": getVectorFloat32(10, 2.0),
			},
		},
		{
			key: "d",
			data: map[string]any{
				"float32": getVectorFloat32(10, 3.0),
			},
		},
		{
			key: "e",
			data: map[string]any{
				"float32": getVectorFloat32(10, 4.0),
			},
		},
		{
			key: "f",
			data: map[string]any{
				"float32": getVectorFloat32(10, 5.0),
			},
		},
		{
			key: "g",
			data: map[string]any{
				"float32": getVectorFloat32(10, 6.0),
			},
		},
		{
			key: "h",
			data: map[string]any{
				"float32": getVectorFloat32(10, 7.0),
			},
		},
		{
			key: "i",
			data: map[string]any{
				"float32": getVectorFloat32(10, 8.0),
			},
		},
		{
			key: "j",
			data: map[string]any{
				"float32": getVectorFloat32(10, 9.0),
			},
		},
	}

	testCases := []struct {
		name          string
		index         *protos.IndexDefinition
		records       []testRecord
		cmd           string
		expectedTable string
	}{
		{
			name: "run query with zero vector",
			index: tests.NewIndexDefinitionBuilder(
				"query-single-index-test", "test", 10, protos.VectorDistanceMetric_SQUARED_EUCLIDEAN, "float32",
			).Build(),
			records: testRecords,
			cmd:     "query -i query-single-index-test -n test --max-results 3 --fields str,int,float,float32,map --no-color --format 1",
			expectedTable: `Query Results
,Namespace,Key,Distance,Generation,Data
1,test,a,0,0,"Key\,Value
float\,3.14
float32\,\"[0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0]\"
int\,1
map\,map[foo:bar]
str\,a"
2,test,b,1,0,"Key\,Value
float32\,\"[0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,1.0]\""
3,test,c,4,0,"Key\,Value
float32\,\"[0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,2.0]\""
`,
		},
		{
			name: "run query with custom vector",
			index: tests.NewIndexDefinitionBuilder(
				"query-single-index-test", "test", 10, protos.VectorDistanceMetric_SQUARED_EUCLIDEAN, "float32",
			).Build(),
			records: testRecords,
			cmd:     "query -i query-single-index-test -n test --vector [0.0,0.0,0.0,0.0,0.0,0.0,0.0,0.0,0.0,1.0]  --no-color --format 1",
			expectedTable: `Query Results
,Namespace,Key,Distance,Generation,Data
1,test,b,0,0,"Key\,Value
float32\,\"[0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,1.0]\""
2,test,a,1,0,"Key\,Value
extra\,to not display
float\,3.14
float32\,\"[0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0]\"
int\,1
map\,map[foo:bar]
...\,..."
3,test,c,1,0,"Key\,Value
float32\,\"[0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,2.0]\""
4,test,d,4,0,"Key\,Value
float32\,\"[0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,3.0]\""
5,test,e,9,0,"Key\,Value
float32\,\"[0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,4.0]\""
Hint: To increase the number of records returned, use the --max-results flag.
Hint: To choose which record keys are displayed, use the --fields flag. By default only 5 are displayed.
`,
		},
		{
			name: "run query with using key",
			index: tests.NewIndexDefinitionBuilder(
				"query-single-index-test", "test", 10, protos.VectorDistanceMetric_SQUARED_EUCLIDEAN, "float32",
			).Build(),
			records: testRecords,
			cmd:     "query -i query-single-index-test -n test -k b --no-color --format 1",
			expectedTable: `Query Results
,Namespace,Key,Distance,Generation,Data
1,test,a,1,0,"Key\,Value
extra\,to not display
float\,3.14
float32\,\"[0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0]\"
int\,1
map\,map[foo:bar]
...\,..."
2,test,c,1,0,"Key\,Value
float32\,\"[0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,2.0]\""
3,test,d,4,0,"Key\,Value
float32\,\"[0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,3.0]\""
4,test,e,9,0,"Key\,Value
float32\,\"[0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,4.0]\""
5,test,f,16,0,"Key\,Value
float32\,\"[0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,0.0\\,5.0]\""
Hint: To increase the number of records returned, use the --max-results flag.
Hint: To choose which record keys are displayed, use the --fields flag. By default only 5 are displayed.
`,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			err := suite.AvsClient.IndexCreateFromIndexDef(
				context.Background(),
				tc.index,
			)
			if err != nil {
				suite.FailNowf("unable to create index", "%v", err)
			}

			defer suite.AvsClient.IndexDrop(
				context.Background(),
				tc.index.Id.Namespace,
				tc.index.Id.Name,
			)

			for _, record := range tc.records {
				suite.AvsClient.Upsert(
					context.Background(),
					tc.index.Id.Namespace,
					tc.index.SetFilter,
					record.key,
					record.data,
					false,
				)
			}

			suite.AvsClient.WaitForIndexCompletion(
				context.Background(),
				tc.index.Id.Namespace,
				tc.index.Id.Name,
				time.Second*12,
			)

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
			"use seeds and hosts together",
			"query --host 1.1.1.1:3001 --seeds 2.2.2.2:3000 -n test -i i",
			"Error: only --seeds or --host allowed",
		},
		{
			"use set without the key flag",
			"query --namespace test -i index --set testset",
			"Warning: The --set flag is only used when the --key flag is set.",
		},
		{
			"try to query an index that does not exist",
			"query --namespace test -i DNE",
			"Error: Failed to get index definition: failed to get index: server error: NotFound, msg=index test:DNE not found",
		},
		{
			"try to query a key that does not exist",
			fmt.Sprintf("query --namespace %s -i %s -k DNE", namespace, indexName),
			"Error: Failed to get vector using key: unable to get record: failed to get record: server error: NotFound",
		},
		{
			"query a key without a set and check for prompt",
			fmt.Sprintf("query --namespace %s -i %s -k DNE", namespace, indexName),
			"Warning: The requested record was not found. If the record is in a set, use may also need to provide the --set flag.",
		},

		//Warning: The requested record was not found. If the record is in a set, use may also need to provide the --set flag.
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
			"fail to revoke user to invalid user",
			"users revoke --name foo1 --roles admin",
			"failed to revoke user roles: server error: NotFound",
		},
		{
			"fail to grant user to invalid user",
			"users grant --name foo1 --roles admin",
			"failed to grant user roles: server error: NotFound",
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
			"fail to grant user with invalid role",
			"users grant --name foo1 --roles invalid",
			"unknown roles [invalid]",
		},
		{
			"fail to revoke user with invalid role",
			"users revoke --name foo1 --roles invalid",
			"unknown roles [invalid]",
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
			"use seeds and hosts together",
			"index create -y --seeds 2.2.2.2:3000 --host 1.1.1.1:3001 -n test -i index1 -d 256 -m SQUARED_EUCLIDEAN --vector-field vector1 --storage-namespace bar --storage-set testbar",
			"Error: only --seeds or --host allowed",
		},
		{
			"use seeds and hosts together",
			"index list --host 1.1.1.1:3001 --seeds 2.2.2.2:3000",
			"Error: only --seeds or --host allowed",
		},
		{
			"use seeds and hosts together",
			"index drop -y --host 1.1.1.1:3001 --seeds 2.2.2.2:3000 -n test -i index1",
			"Error: only --seeds or --host allowed",
		},
		{
			"use seeds and hosts together",
			"index gc --host 1.1.1.1:3001 --seeds 2.2.2.2:3000 -n test -i index1",
			"Error: only --seeds or --host allowed",
		},
		{ // To test `asvec index create` logic where it infers that the user is trying to pass via stdin or not
			"error because no create index required args are provided",
			fmt.Sprintf("index create --seeds %s", suite.AvsHostPort.String()),
			"Error: required flag(s) \"dimension\", \"distance-metric\", \"index-name\", \"namespace\", \"vector-field\" not set",
		},
		{ // To test `asvec index create` logic where it infers that the user is trying to pass via stdin or not
			"error because no create index required args are provided",
			fmt.Sprintf("index create"),
			"Error: required flag(s) \"dimension\", \"distance-metric\", \"index-name\", \"namespace\", \"vector-field\" not set",
		},
		{ // To test `asvec index create` logic where it infers that the user is trying to pass via stdin or not
			"error because some create index required args are not provided",
			fmt.Sprintf("index create --seeds %s --dimension 10", suite.AvsHostPort.String()),
			"Error: required flag(s) \"distance-metric\", \"index-name\", \"namespace\", \"vector-field\" not set",
		},
		{
			"use seeds and hosts together",
			"user create --host 1.1.1.1:3001 --seeds 2.2.2.2:3000 --name foo --roles admin",
			"Error: only --seeds or --host allowed",
		},
		{
			"use seeds and hosts together",
			"user drop --host 1.1.1.1:3001 --seeds 2.2.2.2:3000 --name foo",
			"Error: only --seeds or --host allowed",
		},
		{
			"use seeds and hosts together",
			"user ls --host 1.1.1.1:3001 --seeds 2.2.2.2:3000",
			"Error: only --seeds or --host allowed",
		},
		{
			"use seeds and hosts together",
			"roles ls --host 1.1.1.1:3001 --seeds 2.2.2.2:3000",
			"Error: only --seeds or --host allowed",
		},
		{
			"use seeds and hosts together",
			"nodes ls --host 1.1.1.1:3001 --seeds 2.2.2.2:3000",
			"Error: only --seeds or --host allowed",
		},
		{
			"test with bad dimension",
			"index create -y --host 1.1.1.1:3001  -n test -i index1 -d -1 -m SQUARED_EUCLIDEAN --vector-field vector1 --storage-namespace bar --storage-set testbar",
			"Error: invalid argument \"-1\" for \"-d, --dimension\"",
		},
		{
			"test with bad distance metric",
			"index create -y --host 1.1.1.1:3001  -n test -i index1 -d 10 -m BAD --vector-field vector1 --storage-namespace bar --storage-set testbar",
			"Error: invalid argument \"BAD\" for \"-m, --distance-metric\"",
		},
		{
			"test with bad timeout",
			"index create -y --host 1.1.1.1:3001  -n test -i index1 -d 10 -m SQUARED_EUCLIDEAN --vector-field vector1 --storage-namespace bar --storage-set testbar --timeout 10",
			"Error: invalid argument \"10\" for \"--timeout\"",
		},
		{
			"test with bad hnsw-ef",
			"index create -y --hnsw-ef foo --host 1.1.1.1:3001  -n test -i index1 -d 10 -m SQUARED_EUCLIDEAN --vector-field vector1 --storage-namespace bar --storage-set testbar ",
			"Error: invalid argument \"foo\" for \"--hnsw-ef\"",
		},
		{
			"test with bad hnsw-ef-construction",
			"index create -y --hnsw-ef-construction foo --host 1.1.1.1:3001  -n test -i index1 -d 10 -m SQUARED_EUCLIDEAN --vector-field vector1 --storage-namespace bar --storage-set testbar ",
			"Error: invalid argument \"foo\" for \"--hnsw-ef-construction\"",
		},
		{
			"test with bad hnsw-max-edges",
			"index create -y --hnsw-max-edges foo --host 1.1.1.1:3001  -n test -i index1 -d 10 -m SQUARED_EUCLIDEAN --vector-field vector1 --storage-namespace bar --storage-set testbar ",
			"Error: invalid argument \"foo\" for \"--hnsw-max-edges\"",
		},
		{
			"test with bad hnsw-batch-interval",
			"index create -y --hnsw-batch-interval foo --host 1.1.1.1:3001  -n test -i index1 -d 10 -m SQUARED_EUCLIDEAN --vector-field vector1 --storage-namespace bar --storage-set testbar ",
			"Error: invalid argument \"foo\" for \"--hnsw-batch-interval\"",
		},
		{
			"test with bad hnsw-batch-max-records",
			"index create -y --hnsw-batch-max-records foo --host 1.1.1.1:3001  -n test -i index1 -d 10 -m SQUARED_EUCLIDEAN --vector-field vector1 --storage-namespace bar --storage-set testbar ",
			"Error: invalid argument \"foo\" for \"--hnsw-batch-max-records\"",
		},
		{
			"test with bad hnsw-cache-max-entries",
			"index create -y --hnsw-cache-max-entries foo --host 1.1.1.1:3001  -n test -i index1 -d 10 -m SQUARED_EUCLIDEAN --vector-field vector1 --storage-namespace bar --storage-set testbar ",
			"Error: invalid argument \"foo\" for \"--hnsw-cache-max-entries\"",
		},
		{
			"test with bad hnsw-cache-expiry",
			"index create -y --hnsw-cache-expiry 10 --host 1.1.1.1:3001  -n test -i index1 -d 10 -m SQUARED_EUCLIDEAN --vector-field vector1 --storage-namespace bar --storage-set testbar ",
			"Error: invalid argument \"10\" for \"--hnsw-cache-expiry\"",
		},
		{
			"test with bad hnsw-healer-max-scan-rate-per-node",
			"index create -y --hnsw-healer-max-scan-rate-per-node foo --host 1.1.1.1:3001  -n test -i index1 -d 10 -m SQUARED_EUCLIDEAN --vector-field vector1 --storage-namespace bar --storage-set testbar ",
			"Error: invalid argument \"foo\" for \"--hnsw-healer-max-scan-rate-per-node\"",
		},
		{
			"test with bad hnsw-healer-max-scan-page-size",
			"index create -y --hnsw-healer-max-scan-page-size foo --host 1.1.1.1:3001  -n test -i index1 -d 10 -m SQUARED_EUCLIDEAN --vector-field vector1 --storage-namespace bar --storage-set testbar ",
			"Error: invalid argument \"foo\" for \"--hnsw-healer-max-scan-page-size\"",
		},
		{
			"test with bad hnsw-healer-reindex-percent",
			"index create -y --hnsw-healer-reindex-percent foo --host 1.1.1.1:3001  -n test -i index1 -d 10 -m SQUARED_EUCLIDEAN --vector-field vector1 --storage-namespace bar --storage-set testbar ",
			"Error: invalid argument \"foo\" for \"--hnsw-healer-reindex-percent\"",
		},
		{
			"test with bad hnsw-healer-schedule-delay",
			"index create -y --hnsw-healer-schedule-delay foo --host 1.1.1.1:3001  -n test -i index1 -d 10 -m SQUARED_EUCLIDEAN --vector-field vector1 --storage-namespace bar --storage-set testbar ",
			"Error: invalid argument \"foo\" for \"--hnsw-healer-schedule-delay\"",
		},
		{
			"test with bad hnsw-healer-parallelism",
			"index create -y --hnsw-healer-parallelism foo --host 1.1.1.1:3001  -n test -i index1 -d 10 -m SQUARED_EUCLIDEAN --vector-field vector1 --storage-namespace bar --storage-set testbar ",
			"Error: invalid argument \"foo\" for \"--hnsw-healer-parallelism\"",
		},
		{
			"test with bad hnsw-merge-parallelism",
			"index create -y --hnsw-merge-parallelism foo --host 1.1.1.1:3001  -n test -i index1 -d 10 -m SQUARED_EUCLIDEAN --vector-field vector1 --storage-namespace bar --storage-set testbar ",
			"Error: invalid argument \"foo\" for \"--hnsw-merge-parallelism\"",
		},

		// {
		// 	"test with bad password",
		// 	"user create --password file:blah --name foo --roles admin",
		// 	"blah: no such file or directory",
		// },
		{
			"test with bad tls-cafile",
			"user create --tls-cafile blah --name foo --roles admin",
			"blah: no such file or directory",
		},
		{
			"test with bad tls-capath",
			"user create --tls-capath blah --name foo --roles admin",
			"blah: no such file or directory",
		},
		{
			"test with bad tls-certfile",
			"user create --tls-certfile blah --name foo --roles admin",
			"blah: no such file or directory",
		},
		{
			"test with bad tls-keyfile",
			"user create --tls-keyfile blah --name foo --roles admin",
			"blah: no such file or directory",
		},
		{
			"test with bad tls-keyfile-password",
			"user create --tls-keyfile-password b64:bla65asdf54r345123!@#$h --name foo --roles admin",
			"Error: invalid argument \"b64:bla65asdf54r345123!@#$h\"",
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
