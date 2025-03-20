package writers

import (
	"fmt"
	"io"
	"log/slog"
	"math"
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
		"Unmerged %",
		"Mode*",
		"Status",
	}
	verboseHeadings := append(table.Row{}, headings...)
	verboseHeadings = append(
		verboseHeadings,
		"Vertices",
		"Labels*",
		"Storage",
		"Index Parameters",
		"Standalone Index Metrics",
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
		index.Mode,
		status.Status,
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

		if *index.Mode == protos.IndexMode_STANDALONE {
			tStandaloneIndexMetrics := NewDefaultWriter(nil)
			tStandaloneIndexMetrics.SetTitle("Standalone Index Metrics")
			tStandaloneIndexMetrics.AppendRows([]table.Row{
				{"State", status.StandaloneIndexMetrics.GetState()},
				{"Scanned Vector Records", status.StandaloneIndexMetrics.GetScannedVectorRecordCount()},
				{"Indexed Vector Records", status.StandaloneIndexMetrics.GetIndexedVectorRecordCount()},
			})

			row = append(row, renderTable(tStandaloneIndexMetrics, format))
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

// calculateIndexSize approximates the size of the index in bytes
func calculateIndexSize(index *protos.IndexDefinition, status *protos.IndexStatusResponse) int64 {
	// the "m" parameter in the Hnsw index.
	var m uint32
	switch v := index.Params.(type) {
	case *protos.IndexDefinition_HnswParams:
		m = v.HnswParams.GetM()
	default:
		panic(fmt.Sprintf("unrecognized index params type: %T", index.Params))
	}

	// TODO make sure this is the correct stat to use here
	validVertices := status.GetIndexHealerVerticesValid()
	// The total number of graph nodes in the Hnsw index.
	numGraphNodes := calculateTotalGraphNodes(int64(m), validVertices)

	var (
		// The unique id/digest of the graph node in the Hnsw index graph.
		graphNodeIDBytes int64 = 20
		// The unique id/digest of the Aerospike record corresponding to this graph node in the Hnsw index graph.
		vectorIDBytes int64 = 20
		// The graph layer of this node in the Hnsw index.
		graphLayerBytes int64 = 20
		// 20 bytes per neighbor
		neighborBytes int64 = 20
	)

	// Each dimension is a float32
	vectorBytes := int64(int(index.Dimensions) * 4)
	// Approximate number of neighbors per graph node.
	numNeighbors := 1.5 * float64(m) // Multiplying by 1.5 is as per experiments.

	totalNeighborBytes := int64(math.Round(numNeighbors * float64(neighborBytes)))
	graphNodeBytes := graphNodeIDBytes + vectorIDBytes + graphLayerBytes + totalNeighborBytes + vectorBytes
	indexBytes := numGraphNodes * graphNodeBytes

	return indexBytes
}

func calculateTotalGraphNodes(m, numValidVertices int64) int64 {
	// If m is 1, then int(math.Pow(float64(m), float64(pow)) will always be 1.
	// So, we can return the number of valid vertices as the total graph nodes.
	if m == 1 {
		return numValidVertices
	}

	var (
		totalGraphNodes int64 // Total graph nodes.
		mPow            int64 = 1
		nodes                 = numValidVertices
	)

	for nodes > 0 {
		nodes = numValidVertices / mPow
		totalGraphNodes += nodes
		mPow *= m
	}

	return totalGraphNodes
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
