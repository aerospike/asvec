package writers

import (
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/aerospike/avs-client-go/protos"
	"github.com/jedib0t/go-pretty/v6/table"
)

var rowConfigAutoMerge = table.RowConfig{AutoMerge: true}

//nolint:govet // Padding not a concern for a CLI
type IndexTableWriter struct {
	table   table.Writer
	verbose bool
	logger  *slog.Logger
}

func NewIndexTableWriter(writer io.Writer, verbose bool, logger *slog.Logger) *IndexTableWriter {
	t := IndexTableWriter{NewDefaultWriter(writer), verbose, logger}

	if verbose {
		t.table.AppendHeader(table.Row{"Name", "Namespace", "Set", "Field", "Dimensions",
			"Distance Metric", "Unmerged", "Labels*", "Storage", "Index Parameters"}, rowConfigAutoMerge)
	} else {
		t.table.AppendHeader(table.Row{"Name", "Namespace", "Set", "Field", "Dimensions", "Distance Metric", "Unmerged"})
	}

	t.table.SetTitle("Indexes")
	t.table.SetAutoIndex(true)
	t.table.SortBy([]table.SortBy{
		{Name: "Namespace", Mode: table.Asc},
		{Name: "Set", Mode: table.Asc},
		{Name: "Name", Mode: table.Asc},
	})
	t.table.SetColumnConfigs([]table.ColumnConfig{
		{

			Number:      3,
			Transformer: removeNil,
		},
	})

	t.table.Style().Options.SeparateRows = true

	return &t
}

func (itw *IndexTableWriter) AppendIndexRow(
	index *protos.IndexDefinition,
	status *protos.IndexStatusResponse,
	format int,
) {
	row := table.Row{index.Id.Name, index.Id.Namespace, index.SetFilter, index.Field,
		index.Dimensions, index.VectorDistanceMetric, status.GetUnmergedRecordCount()}

	if itw.verbose {
		row = append(row, index.Labels)

		tStorage := NewDefaultWriter(nil)
		tStorage.AppendRow(table.Row{"Namespace", index.Storage.GetNamespace()})
		tStorage.AppendRow(table.Row{"Set", index.Storage.GetSet()})

		row = append(row, renderTable(tStorage, format))

		switch v := index.Params.(type) {
		case *protos.IndexDefinition_HnswParams:
			tHNSW := NewDefaultWriter(nil)
			tHNSW.SetTitle("HNSW")
			tHNSW.AppendRows([]table.Row{
				{"Max Edges", v.HnswParams.GetM()},
				{"Ef", v.HnswParams.GetEf()},
				{"Construction Ef", v.HnswParams.GetEfConstruction()},
				{"MaxMemQueueSize*", v.HnswParams.GetMaxMemQueueSize()},
				{"Batch Max Records*", v.HnswParams.BatchingParams.GetMaxRecords()},
				{"Batch Interval*", convertMillisecondToDuration(uint64(v.HnswParams.BatchingParams.GetInterval()))},
				{"Cache Max Entires*", v.HnswParams.CachingParams.GetMaxEntries()},
				{"Cache Expiry*", convertMillisecondToDuration(v.HnswParams.CachingParams.GetExpiry())},
				{"Healer Max Scan Rate / Node*", v.HnswParams.HealerParams.GetMaxScanRatePerNode()},
				{"Healer Max Page Size*", v.HnswParams.HealerParams.GetMaxScanPageSize()},
				{"Healer Re-index % *", convertFloatToPercentStr(v.HnswParams.HealerParams.GetReindexPercent())},
				{"Healer Schedule*", v.HnswParams.HealerParams.GetSchedule()},
				{"Healer Parallelism*", v.HnswParams.HealerParams.GetParallelism()},
				{"Merge Index Parallelism*", v.HnswParams.MergeParams.GetIndexParallelism()},
				{"Merge Re-Index Parallelism*", v.HnswParams.MergeParams.GetReIndexParallelism()},
			})

			row = append(row, renderTable(tHNSW, format))
		default:
			itw.logger.Warn("the server returned unrecognized index type params. recognized index param types are: HNSW")
		}
	}

	itw.table.AppendRow(row)
}

func (itw *IndexTableWriter) Render(renderFormat int) {
	if renderFormat == RenderFormatCSV {
		itw.table.RenderCSV()
	} else {
		itw.table.Render()
	}
}

func convertMillisecondToDuration(m uint64) time.Duration {
	return time.Millisecond * time.Duration(m)
}

func convertFloatToPercentStr(f float32) string {
	return fmt.Sprintf("%.2f%%", f)
}
