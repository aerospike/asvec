//go:build integration

package main

import (
	"asvec/tests"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"
	"testing"
	"time"

	avs "github.com/aerospike/avs-client-go"
	"github.com/aerospike/avs-client-go/protos"
	"github.com/aerospike/tools-common-go/client"
	"github.com/stretchr/testify/suite"
)

var wd, _ = os.Getwd()
var logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))

var (
	testNamespace = "test"
	testSet       = "testset"
	barNamespace  = "bar"
)

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

type CmdTestSuite struct {
	suite.Suite
	app          string
	composeFile  string
	suiteName    string
	suiteFlags   []string
	coverFile    string
	avsIP        string
	avsPort      int
	avsHostPort  *avs.HostPort
	avsTLSConfig *tls.Config
	avsCreds     *avs.UserPassCredentials
	avsClient    *avs.AdminClient
}

func TestCmdSuite(t *testing.T) {
	logger = logger.With(slog.Bool("test-logger", true)) // makes it easy to see which logger is which
	rootCA, err := GetCACert("docker/tls/config/tls/ca.aerospike.com.crt")
	if err != nil {
		t.Fatalf("unable to read root ca %v", err)
		t.FailNow()
		logger.Error("Failed to read cert")
	}

	// certificates, err := GetCertificates("docker/mtls/config/tls/localhost.crt", "docker/mtls/config/tls/localhost.key")
	// if err != nil {
	// 	t.Fatalf("unable to read certificates %v", err)
	// 	t.FailNow()
	// 	logger.Error("Failed to read cert")
	// }

	logger.Info("%v", slog.Any("cert", rootCA))
	suite.Run(t, &CmdTestSuite{
		composeFile: "docker/vanilla/docker-compose.yml", // vanilla
		suiteFlags:  []string{"--log-level debug", "--timeout 10s"},
		avsIP:       "localhost",
	})
	// suite.Run(t, &CmdTestSuite{
	// 	composeFile: "docker/tls/docker-compose.yml", // tls
	// 	suiteFlags: []string{
	// 		"--log-level debug",
	// 		"--timeout 10s",
	// 		tests.CreateFlagStr(flags.TLSCaFile, "docker/tls/config/tls/ca.aerospike.com.crt"),
	// 	},
	// 	avsTLSConfig: &tls.Config{
	// 		Certificates: nil,
	// 		RootCAs:      rootCA,
	// 	},
	// 	avsIP: "localhost",
	// })
	// suite.Run(t, &CmdTestSuite{
	// 	composeFile: "docker/mtls/docker-compose.yml", // mutual tls
	// 	suiteFlags: []string{
	// 		"--log-level debug",
	// 		"--timeout 10s",
	// 		tests.CreateFlagStr(flags.TLSCaFile, "docker/mtls/config/tls/ca.aerospike.com.crt"),
	// 		tests.CreateFlagStr(flags.TLSCertFile, "docker/mtls/config/tls/localhost.crt"),
	// 		tests.CreateFlagStr(flags.TLSKeyFile, "docker/mtls/config/tls/localhost.key"),
	// 	},
	// 	avsTLSConfig: &tls.Config{
	// 		Certificates: certificates,
	// 		RootCAs:      rootCA,
	// 	},
	// 	avsIP: "localhost",
	// })
	// suite.Run(t, &CmdTestSuite{
	// 	composeFile: "docker/auth/docker-compose.yml", // tls + auth (auth requires tls)
	// 	suiteFlags: []string{
	// 		"--log-level debug",
	// 		"--timeout 10s",
	// 		tests.CreateFlagStr(flags.TLSCaFile, "docker/auth/config/tls/ca.aerospike.com.crt"),
	// 		tests.CreateFlagStr(flags.AuthUser, "admin"),
	// 		tests.CreateFlagStr(flags.AuthPassword, "admin"),
	// 	},
	// 	avsCreds: avs.NewCredntialsFromUserPass("admin", "admin"),
	// 	avsTLSConfig: &tls.Config{
	// 		Certificates: nil,
	// 		RootCAs:      rootCA,
	// 	},
	// 	avsIP: "localhost",
	// })
}

func (suite *CmdTestSuite) SetupSuite() {
	suite.app = path.Join(wd, "app.test")
	suite.coverFile = path.Join(wd, "../coverage/cmd-coverage.cov")
	suite.avsPort = 10000
	suite.avsHostPort = avs.NewHostPort(suite.avsIP, suite.avsPort)

	err := docker_compose_up(suite.composeFile)

	time.Sleep(time.Second * 10)

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
		suite.avsClient, err = avs.NewAdminClient(
			ctx,
			avs.HostPortSlice{suite.avsHostPort},
			nil,
			true,
			suite.avsCreds,
			suite.avsTLSConfig,
			logger,
		)

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

	err = docker_compose_down(suite.composeFile)
	if err != nil {
		fmt.Println("unable to stop docker compose down")
	}
}

func (suite *CmdTestSuite) SkipIfUserPassAuthDisabled() {
	if suite.avsCreds == nil {
		suite.T().Skip("authentication is disabled. skipping test")
	}
}

// All this does is append the suite flags to args because certain runs (e.g.
// flag parse error tests) should not append this flags
func (suite *CmdTestSuite) runSuiteCmd(asvecCmd ...string) ([]string, error) {
	suiteFlags := strings.Split(strings.Join(suite.suiteFlags, " "), " ")
	asvecCmd = append(suiteFlags, asvecCmd...)
	return suite.runCmd(asvecCmd...)
}

func (suite *CmdTestSuite) runCmd(asvecCmd ...string) ([]string, error) {
	logger.Info("running command", slog.String("cmd", strings.Join(asvecCmd, " ")))
	cmd := exec.Command(suite.app, asvecCmd...)
	cmd.Env = []string{"GOCOVERDIR=" + os.Getenv("COVERAGE_DIR")}
	stdout, err := cmd.Output()

	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return []string{string(ee.Stderr)}, err
		}
		return []string{string(stdout)}, err
	}

	lines := strings.Split(string(stdout), "\n")

	return lines, nil
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
			fmt.Sprintf("index create -y --host %s -n test -i index0 -d 256 -m SQUARED_EUCLIDEAN --vector-field vector0 --index-labels model=all-MiniLM-L6-v2,foo=bar", suite.avsHostPort.String()),
			tests.NewIndexDefinitionBuilder("index0", "test", 256, protos.VectorDistanceMetric_SQUARED_EUCLIDEAN, "vector0").
				WithLabels(map[string]string{"model": "all-MiniLM-L6-v2", "foo": "bar"}).
				Build(),
		},
		{
			"test with storage config",
			"index1",
			"test",
			fmt.Sprintf("index create -y --host %s -n test -i index1 -d 256 -m SQUARED_EUCLIDEAN --vector-field vector1 --storage-namespace bar --storage-set testbar s", suite.avsHostPort.String()),
			tests.NewIndexDefinitionBuilder("index1", "test", 256, protos.VectorDistanceMetric_SQUARED_EUCLIDEAN, "vector1").
				WithStorageNamespace("bar").
				WithStorageSet("testbar").
				Build(),
		},
		{
			"test with hnsw params and seeds",
			"index2",
			"test",
			fmt.Sprintf("index create -y --seeds %s -n test -i index2 -d 256 -m HAMMING --vector-field vector2 --hnsw-max-edges 10 --hnsw-ef 11 --hnsw-ef-construction 12 --hnsw-max-mem-queue-size 10", suite.avsHostPort.String()),
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
			fmt.Sprintf("index create -y --host %s -n test -i index3 -d 256 -m COSINE --vector-field vector3 --hnsw-batch-interval 50s --hnsw-batch-max-records 100", suite.avsHostPort.String()),
			tests.NewIndexDefinitionBuilder("index3", "test", 256, protos.VectorDistanceMetric_COSINE, "vector3").
				WithHnswBatchingMaxRecord(100).
				WithHnswBatchingInterval(50000).
				Build(),
		},
		{
			"test with hnsw cache params",
			"index4",
			"test",
			fmt.Sprintf("index create -y --host %s -n test -i index4 -d 256 -m COSINE --vector-field vector4 --hnsw-cache-max-entries 1000 --hnsw-cache-expiry 10s", suite.avsHostPort.String()),
			tests.NewIndexDefinitionBuilder("index4", "test", 256, protos.VectorDistanceMetric_COSINE, "vector4").
				WithHnswCacheExpiry(10000).
				WithHnswCacheMaxEntries(1000).
				Build(),
		},
		{
			"test with hnsw healer params",
			"index5",
			"test",
			fmt.Sprintf("index create -y --host %s -n test -i index5 -d 256 -m COSINE --vector-field vector5 --hnsw-healer-max-scan-rate-per-node 1000 --hnsw-healer-max-scan-page-size 1000 --hnsw-healer-reindex-percent 10.10 --hnsw-healer-schedule-delay 10s --hnsw-healer-parallelism 10", suite.avsHostPort.String()),
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
			fmt.Sprintf("index create -y --host %s -n test -i index6 -d 256 -m COSINE --vector-field vector6 --hnsw-merge-parallelism 10", suite.avsHostPort.String()),
			tests.NewIndexDefinitionBuilder("index6", "test", 256, protos.VectorDistanceMetric_COSINE, "vector6").
				WithHnswMergeParallelism(10).
				Build(),
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			lines, err := suite.runSuiteCmd(strings.Split(tc.cmd, " ")...)

			if err != nil {
				suite.Assert().NoError(err, "error: %s, stdout/err: %s", err, lines)
				suite.FailNow("unable to index create")
			}

			actual, err := suite.avsClient.IndexGet(context.Background(), tc.indexNamespace, tc.indexName)

			if err != nil {
				suite.FailNowf("unable to get index", "%v", err)
			}

			suite.EqualExportedValues(tc.expected_index, actual)
		})
	}
}

func (suite *CmdTestSuite) TestSuccessfulUpdateIndexCmd() {
	suite.avsClient.IndexCreate(context.Background(), "test", "successful-update", "field", uint32(256), protos.VectorDistanceMetric_COSINE, nil)
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
			fmt.Sprintf("index update -y --seeds %s -n test -i successful-update --index-labels new-label=foo --hnsw-max-mem-queue-size 10 --hnsw-batch-max-records 100  --hnsw-batch-interval 10s", suite.avsHostPort.String()),
			builder.
				WithLabels(map[string]string{"new-label": "foo"}).
				WithHnswBatchingInterval(10000).
				WithHnswBatchingMaxRecord(100).
				WithHnswMaxMemQueueSize(10).
				Build(),
		},
		{
			"test with hnsw batch params",
			"successful-update",
			"test",
			fmt.Sprintf("index update -y --host %s -n test -i successful-update --hnsw-batch-interval 50s --hnsw-batch-max-records 100", suite.avsHostPort.String()),
			builder.
				WithHnswBatchingMaxRecord(100).
				WithHnswBatchingInterval(50000).
				Build(),
		},
		{
			"test with hnsw cache params",
			"successful-update",
			"test",
			fmt.Sprintf("index update -y --host %s -n test -i successful-update --hnsw-cache-max-entries 1000 --hnsw-cache-expiry 10s --hnsw-batch-interval 50s --hnsw-batch-max-records 100", suite.avsHostPort.String()),
			builder.
				WithHnswCacheExpiry(10000).
				WithHnswCacheMaxEntries(1000).
				Build(),
		},
		{
			"test with hnsw healer params",
			"successful-update",
			"test",
			fmt.Sprintf("index update -y s --host %s -n test -i successful-update --hnsw-healer-max-scan-rate-per-node 1000 --hnsw-healer-max-scan-page-size 1000 --hnsw-healer-reindex-percent 10.10 --hnsw-healer-schedule-delay 10s --hnsw-healer-parallelism 10 --hnsw-batch-interval 50s --hnsw-batch-max-records 100", suite.avsHostPort.String()),
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
			fmt.Sprintf("index update -y s --host %s -n test -i successful-update --hnsw-merge-parallelism 10  --hnsw-batch-interval 50s --hnsw-batch-max-records 100", suite.avsHostPort.String()),
			builder.
				WithHnswMergeParallelism(10).
				Build(),
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			lines, err := suite.runSuiteCmd(strings.Split(tc.cmd, " ")...)

			if err != nil {
				suite.Assert().NoError(err, "error: %s, stdout/err: %s", err, lines)
				suite.FailNow("unable to index update")
			}

			time.Sleep(5 * time.Second)

			actual, err := suite.avsClient.IndexGet(context.Background(), tc.indexNamespace, tc.indexName)

			if err != nil {
				suite.FailNowf("unable to get index", "%v", err)
			}

			suite.EqualExportedValues(tc.expected_index, actual)
		})
	}

}

func (suite *CmdTestSuite) TestUpdateIndexDoesNotExist() {
	lines, err := suite.runSuiteCmd(strings.Split(fmt.Sprintf("index update -y --host %s -n test -i DNE --hnsw-merge-parallelism 10", suite.avsHostPort.String()), " ")...)
	suite.Assert().Error(err, "index should have NOT existed. stdout/err: %s", lines)
	suite.Assert().Contains(lines[0], "server error")
}

func (suite *CmdTestSuite) TestSuccessfulGCIndexCmd() {
	index := "successful-gc"
	ns := "test"
	suite.avsClient.IndexCreate(context.Background(), ns, index, "field", uint32(256), protos.VectorDistanceMetric_COSINE, nil)
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
			fmt.Sprintf("index gc --host %s -n test -i successful-gc -c 10", suite.avsHostPort.String()),
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			lines, err := suite.runSuiteCmd(strings.Split(tc.cmd, " ")...)

			if err != nil {
				suite.Assert().NoError(err, "error: %s, stdout/err: %s", err, lines)
				suite.FailNow("unable to index gc")
			}
		})
	}
}

func (suite *CmdTestSuite) TestGCIndexDoesNotExist() {
	lines, err := suite.runSuiteCmd(strings.Split(fmt.Sprintf("index gc --host %s -n test -i DNE -c 10", suite.avsHostPort.String()), " ")...)
	suite.Assert().Error(err, "index should have NOT existed. stdout/err: %s", lines)
	suite.Assert().Contains(lines[0], "server error")
}

func (suite *CmdTestSuite) TestCreateIndexFailsAlreadyExistsCmd() {
	lines, err := suite.runSuiteCmd(strings.Split(fmt.Sprintf("index create -y --host %s -n test -i exists -d 256 -m SQUARED_EUCLIDEAN --vector-field vector1 --storage-namespace bar --storage-set testbar s", suite.avsHostPort.String()), " ")...)
	suite.Assert().NoError(err, "index should have NOT existed on first call. error: %s, stdout/err: %s", err, lines)

	lines, err = suite.runSuiteCmd(strings.Split(fmt.Sprintf("index create -y --host %s -n test -i exists -d 256 -m SQUARED_EUCLIDEAN --vector-field vector1 --storage-namespace bar --storage-set testbar s", suite.avsHostPort.String()), " ")...)
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
			"test with just namespace and seeds",
			"indexdrop1",
			"test",
			nil,
			fmt.Sprintf("index drop -y --seeds %s -n test -i indexdrop1 s", suite.avsHostPort.String()),
		},
		{
			"test with set",
			"indexdrop2",
			"test",
			[]string{
				"testset",
			},
			fmt.Sprintf("index drop -y --host %s -n test -s testset -i indexdrop2 s", suite.avsHostPort.String()),
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			err := suite.avsClient.IndexCreate(
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

			lines, err := suite.runSuiteCmd(strings.Split(tc.cmd, " ")...)
			suite.Assert().NoError(err, "error: %s, stdout/err: %s", err, lines)

			if err != nil {
				suite.FailNow("unable to index drop")
			}

			_, err = suite.avsClient.IndexGet(context.Background(), tc.indexNamespace, tc.indexName)

			time.Sleep(time.Second * 3)

			if err == nil {
				suite.FailNow("err is nil, that means the index still exists")
			}
		})
	}
}

func (suite *CmdTestSuite) TestDropIndexFailsDoesNotExistCmd() {
	lines, err := suite.runSuiteCmd(strings.Split(fmt.Sprintf("index drop -y --seeds %s -n test -i DNE s", suite.avsHostPort.String()), " ")...)

	suite.Assert().Error(err, "index should have NOT existed. stdout/err: %s", lines)
	suite.Assert().Contains(lines[0], "server error")
}

func removeANSICodes(input string) string {
	re := regexp.MustCompile(`\x1b[^m]*m`)
	return re.ReplaceAllString(input, "")
}

func (suite *CmdTestSuite) TestSuccessfulListIndexCmd() {
	indexes, err := suite.avsClient.IndexList(context.Background())
	if err != nil {
		suite.FailNow(err.Error())
	}

	for _, index := range indexes.GetIndices() {
		err := suite.avsClient.IndexDrop(context.Background(), index.Id.Namespace, index.Id.Name)
		if err != nil {
			suite.FailNow(err.Error())
		}
	}

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
			fmt.Sprintf("index list -h %s", suite.avsHostPort.String()),
			`╭─────────────────────────────────────────────────────────────────────────╮
│                                 Indexes                                 │
├───┬──────┬───────────┬────────┬────────────┬─────────────────┬──────────┤
│   │ NAME │ NAMESPACE │ FIELD  │ DIMENSIONS │ DISTANCE METRIC │ UNMERGED │
├───┼──────┼───────────┼────────┼────────────┼─────────────────┼──────────┤
│ 1 │ list │ test      │ vector │        256 │          COSINE │        0 │
╰───┴──────┴───────────┴────────┴────────────┴─────────────────┴──────────╯
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
			fmt.Sprintf("index list -h %s", suite.avsHostPort.String()),
			`╭───────────────────────────────────────────────────────────────────────────────────╮
│                                      Indexes                                      │
├───┬───────┬───────────┬────────┬────────┬────────────┬─────────────────┬──────────┤
│   │ NAME  │ NAMESPACE │ SET    │ FIELD  │ DIMENSIONS │ DISTANCE METRIC │ UNMERGED │
├───┼───────┼───────────┼────────┼────────┼────────────┼─────────────────┼──────────┤
│ 1 │ list2 │ bar       │ barset │ vector │        256 │         HAMMING │        0 │
├───┼───────┼───────────┼────────┼────────┼────────────┼─────────────────┼──────────┤
│ 2 │ list1 │ test      │        │ vector │        256 │          COSINE │        0 │
╰───┴───────┴───────────┴────────┴────────┴────────────┴─────────────────┴──────────╯
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
			fmt.Sprintf("index list -h %s --verbose", suite.avsHostPort.String()),
			`╭──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╮
│                                                                                Indexes                                                                               │
├───┬───────┬───────────┬────────┬────────┬────────────┬─────────────────┬──────────┬──────────────┬───────────────────────┬───────────────────────────────────────────┤
│   │ NAME  │ NAMESPACE │ SET    │ FIELD  │ DIMENSIONS │ DISTANCE METRIC │ UNMERGED │ LABELS*      │ STORAGE               │ INDEX PARAMETERS                          │
├───┼───────┼───────────┼────────┼────────┼────────────┼─────────────────┼──────────┼──────────────┼───────────────────────┼───────────────────────────────────────────┤
│ 1 │ list2 │ bar       │ barset │ vector │        256 │         HAMMING │        0 │ map[]        │ ╭───────────┬───────╮ │ ╭───────────────────────────────────────╮ │
│   │       │           │        │        │            │                 │          │              │ │ Namespace │ bar   │ │ │                  HNSW                 │ │
│   │       │           │        │        │            │                 │          │              │ │ Set       │ list2 │ │ ├──────────────────────────────┬────────┤ │
│   │       │           │        │        │            │                 │          │              │ ╰───────────┴───────╯ │ │ Max Edges                    │ 16     │ │
│   │       │           │        │        │            │                 │          │              │                       │ │ Ef                           │ 100    │ │
│   │       │           │        │        │            │                 │          │              │                       │ │ Construction Ef              │ 100    │ │
│   │       │           │        │        │            │                 │          │              │                       │ │ MaxMemQueueSize*             │ 0      │ │
│   │       │           │        │        │            │                 │          │              │                       │ │ Batch Max Records*           │ 100000 │ │
│   │       │           │        │        │            │                 │          │              │                       │ │ Batch Interval*              │ 30s    │ │
│   │       │           │        │        │            │                 │          │              │                       │ │ Cache Max Entires*           │ 0      │ │
│   │       │           │        │        │            │                 │          │              │                       │ │ Cache Expiry*                │ 0s     │ │
│   │       │           │        │        │            │                 │          │              │                       │ │ Healer Max Scan Rate / Node* │ 0      │ │
│   │       │           │        │        │            │                 │          │              │                       │ │ Healer Max Page Size*        │ 0      │ │
│   │       │           │        │        │            │                 │          │              │                       │ │ Healer Re-index % *          │ 0.00%  │ │
│   │       │           │        │        │            │                 │          │              │                       │ │ Healer Schedule Delay*       │ 0s     │ │
│   │       │           │        │        │            │                 │          │              │                       │ │ Healer Parallelism*          │ 0      │ │
│   │       │           │        │        │            │                 │          │              │                       │ │ Merge Parallelism*           │ 0      │ │
│   │       │           │        │        │            │                 │          │              │                       │ ╰──────────────────────────────┴────────╯ │
├───┼───────┼───────────┼────────┼────────┼────────────┼─────────────────┼──────────┼──────────────┼───────────────────────┼───────────────────────────────────────────┤
│ 2 │ list1 │ test      │        │ vector │        256 │          COSINE │        0 │ map[foo:bar] │ ╭───────────┬───────╮ │ ╭───────────────────────────────────────╮ │
│   │       │           │        │        │            │                 │          │              │ │ Namespace │ test  │ │ │                  HNSW                 │ │
│   │       │           │        │        │            │                 │          │              │ │ Set       │ list1 │ │ ├──────────────────────────────┬────────┤ │
│   │       │           │        │        │            │                 │          │              │ ╰───────────┴───────╯ │ │ Max Edges                    │ 16     │ │
│   │       │           │        │        │            │                 │          │              │                       │ │ Ef                           │ 100    │ │
│   │       │           │        │        │            │                 │          │              │                       │ │ Construction Ef              │ 100    │ │
│   │       │           │        │        │            │                 │          │              │                       │ │ MaxMemQueueSize*             │ 0      │ │
│   │       │           │        │        │            │                 │          │              │                       │ │ Batch Max Records*           │ 100000 │ │
│   │       │           │        │        │            │                 │          │              │                       │ │ Batch Interval*              │ 30s    │ │
│   │       │           │        │        │            │                 │          │              │                       │ │ Cache Max Entires*           │ 0      │ │
│   │       │           │        │        │            │                 │          │              │                       │ │ Cache Expiry*                │ 0s     │ │
│   │       │           │        │        │            │                 │          │              │                       │ │ Healer Max Scan Rate / Node* │ 0      │ │
│   │       │           │        │        │            │                 │          │              │                       │ │ Healer Max Page Size*        │ 0      │ │
│   │       │           │        │        │            │                 │          │              │                       │ │ Healer Re-index % *          │ 0.00%  │ │
│   │       │           │        │        │            │                 │          │              │                       │ │ Healer Schedule Delay*       │ 0s     │ │
│   │       │           │        │        │            │                 │          │              │                       │ │ Healer Parallelism*          │ 0      │ │
│   │       │           │        │        │            │                 │          │              │                       │ │ Merge Parallelism*           │ 0      │ │
│   │       │           │        │        │            │                 │          │              │                       │ ╰──────────────────────────────┴────────╯ │
╰───┴───────┴───────────┴────────┴────────┴────────────┴─────────────────┴──────────┴──────────────┴───────────────────────┴───────────────────────────────────────────╯
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

				err := suite.avsClient.IndexCreate(
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

				defer suite.avsClient.IndexDrop(
					context.Background(),
					index.Id.Namespace,
					index.Id.Name,
				)
			}

			lines, err := suite.runSuiteCmd(strings.Split(tc.cmd, " ")...)
			suite.Assert().NoError(err, "error: %s, stdout/err: %s", err, lines)

			actualTable := removeANSICodes(strings.Join(lines, "\n"))

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
			fmt.Sprintf("users create --host %s s --name foo1 --new-password foo --roles admin,read-write", suite.avsHostPort.String()),
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
			fmt.Sprintf("users create --host %s s --name foo2 --new-password foo --roles admin --roles read-write", suite.avsHostPort.String()),
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
			lines, err := suite.runSuiteCmd(strings.Split(tc.cmd, " ")...)
			suite.Assert().NoError(err, "error: %s, stdout/err: %s", err, lines)

			if err != nil {
				suite.FailNow("failed")
			}

			time.Sleep(time.Second * 1)

			actualUser, err := suite.avsClient.GetUser(context.Background(), tc.expectedUser.Username)
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
			fmt.Sprintf("users create --host %s s --name foo1 --new-password foo --roles invalid", suite.avsHostPort.String()),
			"unknown roles [invalid]",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			lines, err := suite.runSuiteCmd(strings.Split(tc.cmd, " ")...)
			suite.Assert().Error(err, "error: %s, stdout/err: %s", err, lines)
			suite.Assert().Contains(lines[0], tc.expectedErr)
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
			fmt.Sprintf("users drop --host %s s --name drop0", suite.avsHostPort.String()),
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			err := suite.avsClient.CreateUser(context.Background(), tc.user, tc.user, []string{"admin"})
			suite.Assert().NoError(err, "we were not able to create the user before we try to drop it", err)

			lines, err := suite.runSuiteCmd(strings.Split(tc.cmd, " ")...)
			suite.Assert().NoError(err, "error: %s, stdout/err: %s", err, lines)

			if err != nil {
				suite.FailNow("failed")
			}

			_, err = suite.avsClient.GetUser(context.Background(), tc.user)
			suite.Assert().Error(err, "we should not have retrieved the dropped user")
		})
	}
}

// Server treats non-existing users as a no-op in drop cmd
//
// func (suite *CmdTestSuite) TestFailedUserDropCmd() {

// 	if suite.avsUser == nil {
// 		suite.T().Skip("authentication is disabled. skipping test")
// 	}

// 	lines, err := suite.runCmd(strings.Split(fmt.Sprintf("users drop --host %s s --name DNE", suite.avsHostPort.String()), " ")...)
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
			fmt.Sprintf("users grant --host %s s --name grant0 --roles read-write", suite.avsHostPort.String()),
			&protos.User{
				Username: "grant0",
				Roles:    []string{"read-write", "admin"},
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			err := suite.avsClient.CreateUser(context.Background(), tc.user, "foo", []string{"admin"})
			suite.Assert().NoError(err, "we were not able to create the user before we try to grant it", err)

			lines, err := suite.runSuiteCmd(strings.Split(tc.cmd, " ")...)
			suite.Assert().NoError(err, "error: %s, stdout/err: %s", err, lines)

			if err != nil {
				suite.FailNow("failed")
			}

			actualUser, err := suite.avsClient.GetUser(context.Background(), tc.user)
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
			fmt.Sprintf("users revoke --host %s s --name revoke0 --roles read-write", suite.avsHostPort.String()),
			&protos.User{
				Username: "revoke0",
				Roles:    []string{"admin"},
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			err := suite.avsClient.CreateUser(context.Background(), tc.user, "foo", []string{"admin", "read-write"})
			suite.Assert().NoError(err, "we were not able to create the user before we try to revoke it", err)

			lines, err := suite.runSuiteCmd(strings.Split(tc.cmd, " ")...)
			suite.Assert().NoError(err, "error: %s, stdout/err: %s", err, lines)

			if err != nil {
				suite.FailNow("failed")
			}

			actualUser, err := suite.avsClient.GetUser(context.Background(), tc.user)
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
			fmt.Sprintf("users new-password --host %s s --name password0 --new-password foo", suite.avsHostPort.String()),
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			err := suite.avsClient.CreateUser(context.Background(), tc.user, "oldpass", []string{"admin"})
			suite.Assert().NoError(err, "we were not able to create the user before we try to change password", err)

			lines, err := suite.runSuiteCmd(strings.Split(tc.cmd, " ")...)
			suite.Assert().NoError(err, "error: %s, stdout/err: %s", err, lines)

			if err != nil {
				suite.FailNow("failed")
			}

			ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
			defer cancel()

			creds := avs.NewCredntialsFromUserPass(tc.user, tc.newPassword)
			_, err = avs.NewAdminClient(
				ctx,
				avs.HostPortSlice{suite.avsHostPort},
				nil,
				true,
				creds,
				suite.avsTLSConfig,
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
			fmt.Sprintf("users list --seeds %s s", suite.avsHostPort.String()),
			`╭───────────────────────────────╮
│             Users             │
├───┬───────┬───────────────────┤
│   │ USER  │ ROLES             │
├───┼───────┼───────────────────┤
│ 1 │ admin │ admin, read-write │
╰───┴───────┴───────────────────╯
Use 'role list' to view available roles
`,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			lines, err := suite.runSuiteCmd(strings.Split(tc.cmd, " ")...)
			suite.Assert().NoError(err, "error: %s, stdout/err: %s", err, lines)

			actualTable := removeANSICodes(strings.Join(lines, "\n"))

			suite.Assert().Equal(tc.expectedTable, actualTable)
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
			fmt.Sprintf("users revoke --host %s s --name foo1 --roles admin", suite.avsHostPort.String()),
			"failed to revoke user roles: server error: NotFound",
		},
		{
			"fail to grant user to invalid user",
			fmt.Sprintf("users grant --host %s s --name foo1 --roles admin", suite.avsHostPort.String()),
			"failed to grant user roles: server error: NotFound",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			lines, err := suite.runSuiteCmd(strings.Split(tc.cmd, " ")...)
			suite.Assert().Error(err, "error: %s, stdout/err: %s", err, lines)
			suite.Assert().Contains(lines[0], tc.expectedErr)
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
			fmt.Sprintf("users grant --host %s s --name foo1 --roles invalid", suite.avsHostPort.String()),
			"unknown roles [invalid]",
		},
		{
			"fail to revoke user with invalid role",
			fmt.Sprintf("users revoke --host %s s --name foo1 --roles invalid", suite.avsHostPort.String()),
			"unknown roles [invalid]",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			lines, err := suite.runSuiteCmd(strings.Split(tc.cmd, " ")...)
			suite.Assert().Error(err, "error: %s, stdout/err: %s", err, lines)
			suite.Assert().Contains(lines[0], tc.expectedErr)
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
			fmt.Sprintf("role list --seeds %s s", suite.avsHostPort.String()),
			`╭───┬────────────╮
│   │ ROLES      │
├───┼────────────┤
│ 1 │ admin      │
│ 2 │ read-write │
╰───┴────────────╯
`,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			lines, err := suite.runSuiteCmd(strings.Split(tc.cmd, " ")...)
			suite.Assert().NoError(err, "error: %s, stdout/err: %s", err, lines)

			actualTable := removeANSICodes(strings.Join(lines, "\n"))

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
			fmt.Sprintf("index create -y --seeds %s --host 1.1.1.1:3001 -n test -i index1 -d 256 -m SQUARED_EUCLIDEAN --vector-field vector1 --storage-namespace bar --storage-set testbar s", suite.avsHostPort.String()),
			"Error: only --seeds or --host allowed",
		},
		{
			"use seeds and hosts together",
			fmt.Sprintf("index list --seeds %s --host 1.1.1.1:3001", suite.avsHostPort.String()),
			"Error: only --seeds or --host allowed",
		},
		{
			"use seeds and hosts together",
			fmt.Sprintf("index drop -y --seeds %s --host 1.1.1.1:3001 -n test -i index1", suite.avsHostPort.String()),
			"Error: only --seeds or --host allowed",
		},
		{
			"test with bad dimension",
			"index create -y --host 1.1.1.1:3001  -n test -i index1 -d -1 -m SQUARED_EUCLIDEAN --vector-field vector1 --storage-namespace bar --storage-set testbar s",
			"Error: invalid argument \"-1\" for \"-d, --dimension\"",
		},
		{
			"test with bad distance metric",
			"index create -y --host 1.1.1.1:3001  -n test -i index1 -d 10 -m BAD --vector-field vector1 --storage-namespace bar --storage-set testbar s",
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
			lines, err := suite.runCmd(strings.Split(tc.cmd, " ")...)

			suite.Assert().Error(err, "error: %s, stdout/err: %s", err, lines)
			suite.Assert().Contains(lines[0], tc.expectedErrStr)
		})
	}
}

func docker_compose_up(composeFile string) error {
	fmt.Println("Starting docker containers")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	cmd := exec.CommandContext(ctx, "docker", "-lDEBUG", "compose", fmt.Sprintf("-f%s", composeFile), "up", "-d")
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

func docker_compose_down(composeFile string) error {
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
