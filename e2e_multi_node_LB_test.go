//go:build integration

package main

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"testing"

	avs "github.com/aerospike/avs-client-go"
	"github.com/stretchr/testify/suite"
)

type MultiNodeLBCmdTestSuite struct {
	CmdBaseTestSuite
}

func TestMultiNodeLBCmdSuite(t *testing.T) {
	logger = logger.With(slog.Bool("test-logger", true)) // makes it easy to see which logger is which

	avsSeed := "localhost"
	avsPort := 10000
	avsHostPort := avs.NewHostPort(avsSeed, avsPort)
	composeFile := "docker/multi-node-LB/docker-compose.yml"

	suites := []*MultiNodeLBCmdTestSuite{
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

func (suite *MultiNodeLBCmdTestSuite) TestNodeListCmd() {

	testCases := []struct {
		name           string
		cmd            string
		expectErrCoded bool
		expectedTable  string
	}{
		{
			"node ls with LB and seeds",
			fmt.Sprintf("node ls --format 1 --no-color --seeds %s", suite.avsHostPort.String()),
			true,
			`Nodes
,Node,Endpoint,Cluster ID,Version,Visible Nodes
1,Seed,localhost:10000,<cluster-id>,0.9.0,"{
    18446651800632365960: [172.20.0.3:5000]
    18446651800632431496: [172.20.0.4:5000]
    18446651800632497032: [172.20.0.5:5000]
}"
`,
		},
		{
			"node ls with LB and host",
			fmt.Sprintf("node ls --format 1 --no-color --host %s", suite.avsHostPort.String()),
			false,
			`Nodes
,Node,Endpoint,Cluster ID,Version,Visible Nodes
1,LB,localhost:10000,<cluster-id>,0.9.0,"{
    18446651800632365960: [172.20.0.3:5000]
    18446651800632431496: [172.20.0.4:5000]
    18446651800632497032: [172.20.0.5:5000]
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

			if tc.expectErrCoded {
				suite.Assert().Error(err)
			} else {
				suite.Assert().NoError(err)
			}

			expectedTableLines := strings.Split(tc.expectedTable, "\n")

			for i, expectedLine := range expectedTableLines {
				suite.Assert().Equal(expectedLine, outLines[i])
			}

		})
	}
}
