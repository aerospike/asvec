//go:build integration_large

package main

import (
	"asvec/tests"
	"context"
	"fmt"
	"strings"
	"testing"

	avs "github.com/aerospike/avs-client-go"
	"github.com/stretchr/testify/suite"
)

type MultiNodeCmdTestSuite struct {
	tests.CmdBaseTestSuite
}

func TestMultiNodeCmdSuite(t *testing.T) {
	avsSeed := "localhost"
	avsPort := 10000
	avsHostPort := avs.NewHostPort(avsSeed, avsPort)
	composeFile := "docker/multi-node/docker-compose.yml"

	suites := []*MultiNodeCmdTestSuite{
		{
			CmdBaseTestSuite: tests.CmdBaseTestSuite{
				SuiteFlags: []string{
					"--timeout 10s",
				},
				AvsHostPort: avsHostPort,
				ComposeFile: composeFile,
			},
		},
	}

	for _, s := range suites {
		suite.Run(t, s)
	}
}

func (suite *MultiNodeCmdTestSuite) TestNodeListCmd() {
	about, err := suite.AvsClient.About(context.Background(), nil)
	if err != nil {
		suite.T().Fatal(err)
	}

	testCases := []struct {
		name          string
		cmd           string
		expectedTable string
	}{
		{
			"node ls with multiple nodes and seeds",
			fmt.Sprintf("node ls --format 1 --no-color --seeds %s", suite.AvsHostPort.String()),
			`Nodes
,Node,Endpoint,Cluster ID,Version,Visible Nodes
1,139637976803088,127.0.0.1:10000,<cluster-id>,<version>,"{
    139637976803089: [127.0.0.1:10001]
    139637976803090: [127.0.0.1:10002]
}"
2,139637976803089,127.0.0.1:10001,<cluster-id>,<version>,"{
    139637976803088: [127.0.0.1:10000]
    139637976803090: [127.0.0.1:10002]
}"
3,139637976803090,127.0.0.1:10002,<cluster-id>,<version>,"{
    139637976803088: [127.0.0.1:10000]
    139637976803089: [127.0.0.1:10001]
}"
`,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			state, err := suite.AvsClient.ClusteringState(context.Background(), nil)
			suite.Assert().NoError(err)

			clusterIDStr := fmt.Sprintf("%d", state.ClusterId.GetId())
			tc.expectedTable = strings.ReplaceAll(tc.expectedTable, "<cluster-id>", clusterIDStr)
			tc.expectedTable = strings.ReplaceAll(tc.expectedTable, "<version>", about.Version)
			outLines, _, err := suite.RunSuiteCmd(strings.Split(tc.cmd, " ")...)
			// suite.Assert().NoError(err, "error: %s, stdout/err: %s", err,
			// lines

			suite.Assert().Contains(outLines, tc.expectedTable)
		})
	}
}
