package writers

import (
	"io"
	"log/slog"

	"github.com/aerospike/avs-client-go/protos"
	"github.com/jedib0t/go-pretty/v6/table"
)

type NodeInfo struct {
	NodeID            *protos.NodeId
	ConnectedEndpoint *protos.ServerEndpoint
	Endpoints         *protos.ClusterNodeEndpoints
	State             *protos.ClusteringState
	About             *protos.AboutResponse
}

//nolint:govet // Padding not a concern for a CLI
type NodeTableWriter struct {
	table  table.Writer
	isLB   bool
	logger *slog.Logger
}

func NewNodeTableWriter(writer io.Writer, isLB bool, logger *slog.Logger) *NodeTableWriter {
	t := NodeTableWriter{NewDefaultWriter(writer), isLB, logger}

	t.table.SetTitle("Nodes")
	t.table.AppendHeader(
		table.Row{
			"Node",
			"Roles",
			"Endpoint",
			"Cluster ID",
			"Version",
			"Visible Nodes",
		},
		rowConfigAutoMerge,
	)
	t.table.SetAutoIndex(true)
	t.table.SortBy([]table.SortBy{
		{Name: "Node", Mode: table.Asc},
	})
	t.table.SetColumnConfigs([]table.ColumnConfig{
		{
			Name:      "Cluster ID",
			AutoMerge: true,
		},
	})

	t.table.Style().Options.SeparateRows = true

	return &t
}

func (itw *NodeTableWriter) AppendNodeRow(node *NodeInfo) {
	var id = node.NodeID.GetId()

	row := table.Row{}

	if id == 0 {
		if itw.isLB {
			row = append(row, "LB")
		} else {
			row = append(row, "Seed")
		}
	} else {
		row = append(row, id)
	}

	// If the node is a load balancer, it does not have roles.
	if !itw.isLB {
		row = append(row, formatRoles(node.About.GetRoles()))
	} else {
		row = append(row, "N/A")
	}

	row = append(row, formatEndpoint(node.ConnectedEndpoint))

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

	if node.Endpoints != nil {
		row = append(row, formatEndpoints(id, node.Endpoints.Endpoints))
	} else {
		row = append(row, "N/A")
	}

	itw.table.AppendRow(row)
}

func (itw *NodeTableWriter) Render(renderFormat int) {
	if renderFormat == RenderFormatCSV {
		itw.table.RenderCSV()
	} else {
		itw.table.Render()
	}
}
