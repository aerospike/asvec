package writers

import (
	"io"
	"log/slog"

	"github.com/aerospike/aerospike-proximus-client-go/protos"
	"github.com/jedib0t/go-pretty/v6/table"
)

var rowConfigAutoMerge = table.RowConfig{AutoMerge: true}

type IndexTableWriter struct {
	table.Writer
	verbose bool
	logger  *slog.Logger
}

func NewIndexTableWriter(writer io.Writer, verbose bool, logger *slog.Logger) *IndexTableWriter {
	t := IndexTableWriter{NewDefaultWriter(writer), verbose, logger}

	if verbose {
		t.AppendHeader(table.Row{"Name", "Namespace", "Set", "Field", "Dimensions", "Distance Metric", "Unmerged", "Storage", "Index Parameters"}, rowConfigAutoMerge)
		// t.AppendHeader(table.Row{"Name", "Namespace", "Set", "Field", "Dimensions", "Distance Metric", "Unmerged", "Storage", "Storage", "HNSW", "HNSW", "HNSW", "HNSW", "HNSW", "HNSW"}, rowConfigAutoMerge)
		// t.AppendHeader(table.Row{"Name", "Namespace", "Set", "Field", "Dimensions", "Distance Metric", "Unmerged", "Namespace", "Set", "Max Edges", "Ef", "Construction Ef", "Batch", "Batch", "Batch"}, rowConfigAutoMerge)
		// t.AppendHeader(table.Row{"Name", "Namespace", "Set", "Field", "Dimensions", "Distance Metric", "Unmerged", "Namespace", "Set", "Max Edges", "Ef", "Construction Ef", "Max Records", "Interval", "Disabled"})
	} else {
		t.AppendHeader(table.Row{"Name", "Namespace", "Set", "Field", "Dimensions", "Distance Metric", "Unmerged"})
	}

	t.SetTitle("Indexes")
	t.Style().Options.SeparateRows = true
	t.SetAutoIndex(true)
	t.SortBy([]table.SortBy{
		{Name: "Namespace", Mode: table.Asc},
		{Name: "Set", Mode: table.Asc},
		{Name: "Name", Mode: table.Asc},
	})

	t.SetColumnConfigs([]table.ColumnConfig{
		{

			Number: 3,
			// Name:        "Set",
			Transformer: removeNil,
			// AutoMerge:   true,
		},
		// {
		// 	Number:    8,
		// 	Name:      "Namespace",
		// 	AutoMerge: true,
		// },
		// {
		// 	Number:    9,
		// 	Name:      "Set",
		// 	AutoMerge: true,
		// },

		// {Number: 1, AutoMerge: true},
		// {Number: 2, AutoMerge: true},
		// {Number: 3, AutoMerge: true},
		// {Number: 4, AutoMerge: true},
		// {Number: 5, AutoMerge: true},
		// {Number: 6, AutoMerge: true},
		// {Number: 7, AutoMerge: true},
		// {Number: 8, AutoMerge: true},
		// {Number: 9, AutoMerge: true},
		// {Number: 10, AutoMerge: true},
		// {Number: 11, AutoMerge: true},
		// {Number: 12, AutoMerge: true},
		// {Number: 13, AutoMerge: true},
		// {Number: 14, AutoMerge: true},
		// {Number: 15, AutoMerge: true},
	})

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
		// row = append(row, status.GetUnmergedRecordCount(), index.Storage.GetNamespace(), index.Storage.GetSet())

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
				{"Batch Disabled", v.HnswParams.BatchingParams.GetDisabled()},
			})
			// row = append(row, v.HnswParams.GetM(), v.HnswParams.GetEf(), v.HnswParams.GetEfConstruction(),
			// 	v.HnswParams.BatchingParams.GetMaxRecords(), v.HnswParams.BatchingParams.GetInterval(),
			// 	v.HnswParams.BatchingParams.GetDisabled())
			row = append(row, tHNSW.Render())
		default:
			itw.logger.Warn("the server returned unrecognized index type params. recognized index param types are: HNSW")
		}
	}

	itw.AppendRow(row)
}
