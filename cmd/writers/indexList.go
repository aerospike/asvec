package writers

import (
	"io"
	"log/slog"

	"github.com/aerospike/avs-client-go/protos"
	"github.com/jedib0t/go-pretty/v6/table"
)

var rowConfigAutoMerge = table.RowConfig{AutoMerge: true}

//nolint:govet // Padding not a concern for a CLI
type IndexTableWriter struct {
	table.Writer
	verbose bool
	logger  *slog.Logger
}

func NewIndexTableWriter(writer io.Writer, verbose bool, logger *slog.Logger) *IndexTableWriter {
	t := IndexTableWriter{NewDefaultWriter(writer), verbose, logger}

	if verbose {
		t.AppendHeader(table.Row{"Name", "Namespace", "Set", "Field", "Dimensions",
			"Distance Metric", "Unmerged", "Storage", "Index Parameters"}, rowConfigAutoMerge)
	} else {
		t.AppendHeader(table.Row{"Name", "Namespace", "Set", "Field", "Dimensions", "Distance Metric", "Unmerged"})
	}

	t.SetTitle("Indexes")
	t.SetAutoIndex(true)
	t.SortBy([]table.SortBy{
		{Name: "Namespace", Mode: table.Asc},
		{Name: "Set", Mode: table.Asc},
		{Name: "Name", Mode: table.Asc},
	})
	t.SetColumnConfigs([]table.ColumnConfig{
		{

			Number:      3,
			Transformer: removeNil,
		},
	})

	t.Style().Options.SeparateRows = true

	return &t
}

func (itw *IndexTableWriter) AppendIndexRow(index *protos.IndexDefinition, status *protos.IndexStatusResponse) {
	row := table.Row{index.Id.Name, index.Id.Namespace, index.SetFilter, index.Field,
		index.Dimensions, index.VectorDistanceMetric, status.GetUnmergedRecordCount()}

	if itw.verbose {
		tStorage := NewDefaultWriter(nil)

		tStorage.AppendRow(table.Row{"Namespace", index.Storage.GetNamespace()})
		tStorage.AppendRow(table.Row{"Set", index.Storage.GetSet()})

		row = append(row, tStorage.Render())

		switch v := index.Params.(type) {
		case *protos.IndexDefinition_HnswParams:
			tHNSW := NewDefaultWriter(nil)
			tHNSW.SetTitle("HNSW")
			tHNSW.AppendRows([]table.Row{
				{"Max Edges", v.HnswParams.GetM()},
				{"Ef", v.HnswParams.GetEf()},
				{"Construction Ef", v.HnswParams.GetEfConstruction()},
				{"Batch Max Records", v.HnswParams.BatchingParams.GetMaxRecords()},
				{"Batch Interval", v.HnswParams.BatchingParams.GetInterval()},
				{"Batch Enabled", !v.HnswParams.BatchingParams.GetDisabled()},
			})

			row = append(row, tHNSW.Render())
		default:
			itw.logger.Warn("the server returned unrecognized index type params. recognized index param types are: HNSW")
		}
	}

	itw.AppendRow(row)
}
