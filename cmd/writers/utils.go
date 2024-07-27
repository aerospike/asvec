package writers

import (
	"fmt"
	"strings"

	"github.com/aerospike/avs-client-go/protos"
)

func formatNodeIdList(nodeIds []*protos.NodeId) string {
	ids := make([]string, 0, len(nodeIds))

	for _, nodeId := range nodeIds {
		ids = append(ids, fmt.Sprintf("%d", nodeId.GetId()))
	}

	return strings.Join(ids, "\n")
}

func formatEndpoints(nodeId uint64, nodeEndpoints map[uint64]*protos.ServerEndpointList) string {
	nodeToEndpointsStr := make([]string, 0, len(nodeEndpoints)+2)
	nodeToEndpointsStr = append(nodeToEndpointsStr, "{")

	for id, endpoints := range nodeEndpoints {
		if nodeId == id {
			// Node endpoint include themselves. Remove it from output
			continue
		}

		endpointStrs := make([]string, 0, len(endpoints.GetEndpoints()))
		for _, endpoint := range endpoints.GetEndpoints() {
			endpointStrs = append(endpointStrs, fmt.Sprintf("%s:%d", endpoint.GetAddress(), endpoint.GetPort()))
		}

		nodeToEndpointsStr = append(nodeToEndpointsStr, fmt.Sprintf("	%d: [%s]", id, strings.Join(endpointStrs, ",")))
	}

	nodeToEndpointsStr = append(nodeToEndpointsStr, "}")

	return strings.Join(nodeToEndpointsStr, "\n")
}
