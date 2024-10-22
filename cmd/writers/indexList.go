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

	// type any because the table.Row type is a slice of interface{}
	headings := table.Row{
		"Name",
		"Namespace",
		"Set",
		"Field",
		"Dimensions",
		"Distance Metric",
		"Unmerged",
	}
	verboseHeadings := append(table.Row{}, headings...)
	verboseHeadings = append(
		verboseHeadings,
		"Vector Records",
		"Vertices",
		"Labels*",
		"Storage",
		"Index Parameters",
	)

	if verbose {
		t.table.AppendHeader(verboseHeadings, rowConfigAutoMerge)
	} else {
		t.table.AppendHeader(headings)
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
	row := table.Row{
		index.Id.Name,
		index.Id.Namespace,
		index.SetFilter,
		index.Field,
		index.Dimensions,
		index.VectorDistanceMetric,
		status.GetUnmergedRecordCount(),
	}

	if itw.verbose {
		row = append(row,
			status.GetIndexHealerVectorRecordsIndexed(),
			status.GetIndexHealerVerticesValid(),
			index.Labels,
		)

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
				{"Batch Max Index Records*", v.HnswParams.BatchingParams.GetMaxIndexRecords()},
				{"Batch Index Interval*", convertMillisecondToDuration(uint64(v.HnswParams.BatchingParams.GetIndexInterval()))},
				{"Batch Max Reindex Records*", v.HnswParams.BatchingParams.GetMaxReindexRecords()},
				{"Batch Reindex Interval*", convertMillisecondToDuration(uint64(v.HnswParams.BatchingParams.GetReindexInterval()))},
				{"Cache Max Entires*", v.HnswParams.IndexCachingParams.GetMaxEntries()},
				{"Cache Expiry*", convertMillisecondToDuration(v.HnswParams.IndexCachingParams.GetExpiry())},
				{"Healer Max Scan Rate / Node*", v.HnswParams.HealerParams.GetMaxScanRatePerNode()},
				{"Healer Max Page Size*", v.HnswParams.HealerParams.GetMaxScanPageSize()},
				{"Healer Re-index % *", convertFloatToPercentStr(v.HnswParams.HealerParams.GetReindexPercent())},
				{"Healer Schedule*", v.HnswParams.HealerParams.GetSchedule()},
				{"Healer Parallelism*", v.HnswParams.HealerParams.GetParallelism()},
				{"Merge Index Parallelism*", v.HnswParams.MergeParams.GetIndexParallelism()},
				{"Merge Re-Index Parallelism*", v.HnswParams.MergeParams.GetReIndexParallelism()},
				{"Enable Vector Integrity Check", v.HnswParams.GetEnableVectorIntegrityCheck()},
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

func convertMillisecondToDuration[T int64 | uint64 | uint32](m T) time.Duration {
	return time.Millisecond * time.Duration(m)
}

func convertFloatToPercentStr(f float32) string {
	return fmt.Sprintf("%.2f%%", f)
}
