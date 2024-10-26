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

	headings := table.Row{
		"Name",
		"Namespace",
		"Set",
		"Field",
		"Dimensions",
		"Distance Metric",
		"Unmerged",
		"Vector Records",
		"Size",
		"Umerged %",
	}
	verboseHeadings := append(table.Row{}, headings...)
	verboseHeadings = append(
		verboseHeadings,
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
		status.GetIndexHealerVectorRecordsIndexed(),
		formatBytes(calculateIndexSize(index, status)),
		getPercentUnmerged(status),
	}

	if itw.verbose {
		row = append(row,
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
				{"Index Cache Max Entries*", v.HnswParams.IndexCachingParams.GetMaxEntries()},
				{"Index Cache Expiry*", convertMillisecondToDuration(v.HnswParams.IndexCachingParams.GetExpiry())},
				{"Record Cache Max Entries*", v.HnswParams.RecordCachingParams.GetMaxEntries()},
				{"Record Cache Expiry*", convertMillisecondToDuration(v.HnswParams.RecordCachingParams.GetExpiry())},
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

// calculateIndexSize calculates the size of the index in bytes
func calculateIndexSize(index *protos.IndexDefinition, status *protos.IndexStatusResponse) int64 {
	// Each dimension is a float32
	vectorSize := int64(index.Dimensions) * 4
	// Each index record has ~500 bytes of overhead + the vector size
	indexRecSize := 500 + vectorSize
	// The total size is the number of records times the size of each record
	indexSize := indexRecSize * status.GetIndexHealerVerticesValid()
	return indexSize
}

// formatBytes converts bytes to human readable string format
func formatBytes(bytes int64) string {
	const (
		B  = 1
		KB = 1024 * B
		MB = 1024 * KB
		GB = 1024 * MB
		TB = 1024 * GB
		PB = 1024 * TB
	)

	switch {
	case bytes >= PB:
		return fmt.Sprintf("%.2f PB", float64(bytes)/float64(PB))
	case bytes >= TB:
		return fmt.Sprintf("%.2f TB", float64(bytes)/float64(TB))
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

func getPercentUnmerged(status *protos.IndexStatusResponse) string {
	unmergedCount := status.GetUnmergedRecordCount()
	verticies := status.GetIndexHealerVerticesValid()
	if verticies == 0 {
		return "0%"
	}

	return fmt.Sprintf("%.2f%%", float64(unmergedCount)/float64(verticies)*100)
}
