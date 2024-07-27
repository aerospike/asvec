package writers

import (
	"io"
	"log/slog"

	"github.com/aerospike/avs-client-go/protos"
	"github.com/jedib0t/go-pretty/v6/table"
)

type NodeClusterInfo struct {
	NodeId    *protos.NodeId
	Endpoints *protos.ClusterNodeEndpoints
	State     *protos.ClusteringState
	About     *protos.AboutResponse
}

//nolint:govet // Padding not a concern for a CLI
type ClusterTableWriter struct {
	table.Writer
	logger *slog.Logger
}

func NewClusterTableWriter(writer io.Writer, logger *slog.Logger) *ClusterTableWriter {
	t := ClusterTableWriter{NewDefaultWriter(writer), logger}

	t.SetTitle("Nodes")
	t.AppendHeader(table.Row{"Node ID", "Cluster ID", "Version", "In Cluster?", "Visible Nodes"}, rowConfigAutoMerge)
	t.SetAutoIndex(true)
	t.SortBy([]table.SortBy{
		{Name: "Node ID", Mode: table.Asc},
	})
	// t.SetColumnConfigs([]table.ColumnConfig{
	// 	{},
	// })

	t.Style().Options.SeparateRows = true

	return &t
}

func (itw *ClusterTableWriter) AppendNodeRow(node *NodeClusterInfo) {
	var id = node.NodeId.GetId()

	row := table.Row{id}

	if node.State != nil {
		row = append(row, node.State.ClusterId.GetId())
	} else {
		row = append(row, "N/A")
	}

	if node.About != nil {
		row = append(row, node.About.GetVersion())
	} else {
		row = append(row, "N/A")
	}

	if node.State != nil {
		row = append(row, node.State.IsInCluster)
	} else {
		row = append(row, "N/A")
	}

	if node.Endpoints != nil {
		row = append(row, formatEndpoints(id, node.Endpoints.Endpoints))
	} else {
		row = append(row, "N/A")
	}

	itw.AppendRow(row)
}
