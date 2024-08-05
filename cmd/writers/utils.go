package writers

import (
	"fmt"
	"sort"
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

func formatEndpoint(nodeEndpoint *protos.ServerEndpoint) string {
	if nodeEndpoint == nil {
		return "N/A"
	}

	return fmt.Sprintf("%s:%d", nodeEndpoint.GetAddress(), nodeEndpoint.GetPort())
}

func formatEndpoints(nodeId uint64, nodeEndpoints map[uint64]*protos.ServerEndpointList) string {
	if len(nodeEndpoints) == 0 {
		return "N/A"
	}

	nodeToEndpointsStr := make([]string, 0, len(nodeEndpoints)+2)
	nodeToEndpointsStr = append(nodeToEndpointsStr, "{")
	ids := make([]uint64, 0, len(nodeEndpoints))

	for id := range nodeEndpoints {
		ids = append(ids, id)
	}

	sort.Slice(ids, func(i, j int) bool {
		return ids[i] < ids[j]
	})

	for _, id := range ids {
		if nodeId == id {
			// Node endpoint include themselves. Remove it from output
			continue
		}

		endpoints := nodeEndpoints[id]

		endpointStrs := make([]string, 0, len(endpoints.GetEndpoints()))
		for _, endpoint := range endpoints.GetEndpoints() {
			endpointStrs = append(endpointStrs, formatEndpoint(endpoint))
		}

		nodeToEndpointsStr = append(nodeToEndpointsStr, fmt.Sprintf("	%d: [%s]", id, strings.Join(endpointStrs, ",")))
	}

	nodeToEndpointsStr = append(nodeToEndpointsStr, "}")

	return strings.Join(nodeToEndpointsStr, "\n")
}