package writers

import (
	"fmt"
	"slices"
	"strings"

	"github.com/aerospike/avs-client-go/protos"
	"github.com/jedib0t/go-pretty/v6/table"
)

func formatEndpoint(nodeEndpoint *protos.ServerEndpoint) string {
	if nodeEndpoint == nil {
		return "N/A"
	}

	return fmt.Sprintf("%s:%d", nodeEndpoint.GetAddress(), nodeEndpoint.GetPort())
}

func formatEndpoints(nodeID uint64, nodeEndpoints map[uint64]*protos.ServerEndpointList) string {
	if len(nodeEndpoints) == 0 {
		return "N/A"
	}

	nodeToEndpointsStr := make([]string, 0, len(nodeEndpoints)+2)
	nodeToEndpointsStr = append(nodeToEndpointsStr, "{")
	ids := make([]uint64, 0, len(nodeEndpoints))

	for id := range nodeEndpoints {
		ids = append(ids, id)
	}

	slices.Sort(ids)

	for _, id := range ids {
		if nodeID == id {
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

func renderTable(t table.Writer, format int) string {
	if format == 0 {
		return t.Render()
	}

	return t.RenderCSV()
}
