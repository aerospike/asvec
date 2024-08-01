//go:build integration

package main

import (
	"asvec/tests"
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"
	"time"

	avs "github.com/aerospike/avs-client-go"
	"github.com/stretchr/testify/suite"
)

type CmdBaseTestSuite struct {
	suite.Suite
	app          string
	composeFile  string
	coverFile    string
	avsHostPort  *avs.HostPort
	avsTLSConfig *tls.Config
	avsCreds     *avs.UserPassCredentials
	avsClient    *avs.AdminClient
	suiteFlags   []string
}

func (suite *CmdBaseTestSuite) SetupSuite() {
	suite.app = path.Join(wd, "app.test")
	suite.coverFile = path.Join(wd, "../coverage/cmd-coverage.cov")

	err := tests.DockerComposeUp(suite.composeFile)

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

	suite.avsClient, err = tests.GetAdminClient(suite.avsHostPort, suite.avsCreds, suite.avsTLSConfig, logger)
	if err != nil {
		suite.FailNowf("unable to create admin client", "%v", err)
	}
}

func (suite *CmdBaseTestSuite) TearDownSuite() {
	err := os.Remove(suite.app)
	suite.Assert().NoError(err)
	time.Sleep(time.Second * 5)
	suite.Assert().NoError(err)
	suite.avsClient.Close()

	err = tests.DockerComposeDown(suite.composeFile)
	if err != nil {
		fmt.Println("unable to stop docker compose down")
	}
}

// All this does is append the suite flags to args because certain runs (e.g.
// flag parse error tests) should not append this flags
func (suite *CmdBaseTestSuite) runSuiteCmd(asvecCmd ...string) ([]string, []string, error) {
	suiteFlags := strings.Split(strings.Join(suite.suiteFlags, " "), " ")
	asvecCmd = append(suiteFlags, asvecCmd...)
	return suite.runCmd(asvecCmd...)
}

func (suite *CmdBaseTestSuite) runCmd(asvecCmd ...string) ([]string, []string, error) {
	logger.Info("running command", slog.String("cmd", strings.Join(asvecCmd, " ")))
	cmd := exec.Command(suite.app, asvecCmd...)
	cmd.Env = []string{"GOCOVERDIR=" + os.Getenv("COVERAGE_DIR")}
	// stdout := &bytes.Buffer{}
	// stderr := &bytes.Buffer{}
	// cmd.Stdout = stdout
	// cmd.Stderr = stderr
	// err := cmd.Run()

	// outLines := strings.Split(stdout.String(), "\n")
	// errLines := strings.Split(stderr.String(), "\n")

	output, err := cmd.Output()

	return strings.Split(string(output), "\n"), nil, err
}

type MultiNodeCmdTestSuite struct {
	CmdBaseTestSuite
}

func TestMultiNodeCmdSuite(t *testing.T) {
	logger = logger.With(slog.Bool("test-logger", true)) // makes it easy to see which logger is which

	avsSeed := "localhost"
	avsPort := 10000
	avsHostPort := avs.NewHostPort(avsSeed, avsPort)
	composeFile := "docker/multi-node/docker-compose.yml"

	suites := []*MultiNodeCmdTestSuite{
		{
			CmdBaseTestSuite: CmdBaseTestSuite{
				suiteFlags: []string{
					// "--log-level Error",
					"--timeout 10s",
				},
				avsHostPort: avsHostPort,
				composeFile: composeFile,
			},
		},
	}

	for _, s := range suites {
		suite.Run(t, s)
	}
}

func (suite *MultiNodeCmdTestSuite) TestNodeListCmd() {

	testCases := []struct {
		name          string
		cmd           string
		expectedTable string
	}{
		{
			"node ls with multiple nodes and seeds",
			fmt.Sprintf("node ls --format 1 --no-color --seeds %s", suite.avsHostPort.String()),
			`Nodes
,Node,Endpoint,Cluster ID,Version,Visible Nodes
1,139637976803088,127.0.0.1:10000,<cluster-id>,0.9.0,"{
    139637976803089: [127.0.0.1:10001]
    139637976803090: [127.0.0.1:10002]
}"
2,139637976803089,127.0.0.1:10001,<cluster-id>,0.9.0,"{
    139637976803088: [127.0.0.1:10000]
    139637976803090: [127.0.0.1:10002]
}"
3,139637976803090,127.0.0.1:10002,<cluster-id>,0.9.0,"{
    139637976803088: [127.0.0.1:10000]
    139637976803089: [127.0.0.1:10001]
}"
`,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			state, err := suite.avsClient.ClusteringState(context.Background(), nil)
			suite.Assert().NoError(err)

			clusterIDStr := fmt.Sprintf("%d", state.ClusterId.GetId())
			tc.expectedTable = strings.ReplaceAll(tc.expectedTable, "<cluster-id>", clusterIDStr)
			outLines, _, err := suite.runSuiteCmd(strings.Split(tc.cmd, " ")...)
			// suite.Assert().NoError(err, "error: %s, stdout/err: %s", err, lines)

			expectedTableLines := strings.Split(tc.expectedTable, "\n")

			for i, expectedLine := range expectedTableLines {
				suite.Assert().Equal(expectedLine, outLines[i])
			}

		})
	}
}
