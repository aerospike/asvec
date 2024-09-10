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

type MultiNodeLBCmdTestSuite struct {
	tests.CmdBaseTestSuite
}

func TestMultiNodeLBCmdSuite(t *testing.T) {
	avsSeed := "localhost"
	avsPort := 10000
	avsHostPort := avs.NewHostPort(avsSeed, avsPort)
	composeFile := "docker/multi-node-LB/docker-compose.yml"

	suites := []*MultiNodeLBCmdTestSuite{
		{
			CmdBaseTestSuite: tests.CmdBaseTestSuite{

				SuiteFlags: []string{
					// "--log-level Error",
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

func (suite *MultiNodeLBCmdTestSuite) TestNodeListCmd() {
	about, err := suite.AvsClient.About(context.Background(), nil)
	if err != nil {
		suite.T().Fatal(err)
	}

	testCases := []struct {
		name            string
		cmd             string
		expectErrCoded  bool
		expectedTable   string
		expectedWarning string
	}{
		{
			"node ls with LB and seeds",
			fmt.Sprintf("node ls --format 1 --no-color --seeds %s", suite.AvsHostPort.String()),
			true,
			`Nodes
,Node,Endpoint,Cluster ID,Version,Visible Nodes
1,Seed,localhost:10000,<cluster-id>,<version>,"{
    1103823447824: [1.1.1.1:10000]
    2207646885648: [2.2.2.2:10000]
    3311470323472: [3.3.3.3:10000]
}"`,
			`Warning: Not all nodes are visible to asvec. 
Asvec can't reach: 1103823447824, 2207646885648, 3311470323472
Possible scenarios:
1. You should use --host instead of --seeds to indicate you are connection through a load balancer.
2. Asvec was able to connect to your seeds but the server(s) are returning unreachable endpoints.
   Did you forget --listener-name?
`,
		},
		{
			"node ls with LB and host",
			fmt.Sprintf("node ls --format 1 --no-color --host %s", suite.AvsHostPort.String()),
			false,
			`Nodes
,Node,Endpoint,Cluster ID,Version,Visible Nodes
1,LB,localhost:10000,<cluster-id>,<version>,"{
    1103823447824: [1.1.1.1:10000]
    2207646885648: [2.2.2.2:10000]
    3311470323472: [3.3.3.3:10000]
}"`,
			"",
		},
	}

	for _, tc := range testCases {

		suite.Run(tc.name, func() {
			state, err := suite.AvsClient.ClusteringState(context.Background(), nil)
			suite.Assert().NoError(err)

			clusterIDStr := fmt.Sprintf("%d", state.ClusterId.GetId())
			tc.expectedTable = strings.ReplaceAll(tc.expectedTable, "<cluster-id>", clusterIDStr)
			tc.expectedTable = strings.ReplaceAll(tc.expectedTable, "<version>", about.Version)
			outLines, errLines, err := suite.RunSuiteCmd(strings.Split(tc.cmd, " ")...)

			if tc.expectErrCoded {
				suite.Assert().Error(err)
			} else {
				suite.Assert().NoError(err)
			}

			// expectedTableLines := strings.Split(tc.expectedTable, "\n")

			suite.Assert().Contains(outLines, tc.expectedTable)

			if tc.expectedWarning != "" {
				suite.Assert().Contains(errLines, tc.expectedWarning)
			}
		})
	}
}
