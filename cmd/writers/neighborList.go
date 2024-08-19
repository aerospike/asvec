package writers

import (
	"io"
	"log/slog"
	"slices"

	"github.com/aerospike/avs-client-go"
	"github.com/jedib0t/go-pretty/v6/table"
)

//nolint:govet // Padding not a concern for a CLI
type NeighborTableWriter struct {
	table  table.Writer
	logger *slog.Logger
}

func NewNeighborTableWriter(writer io.Writer, logger *slog.Logger) *NeighborTableWriter {
	t := NeighborTableWriter{NewDefaultWriter(writer), logger}

	t.table.AppendHeader(
		table.Row{
			"Namespace",
			"Set",
			"Key",
			"Distance",
			"Expiration",
			"Generation",
			"Data",
		},
	)

	t.table.SetTitle("Query Results")
	t.table.SetAutoIndex(true)
	// t.table.SortBy([]table.SortBy{
	// 	{Name: "Namespace", Mode: table.Asc},
	// 	{Name: "Set", Mode: table.Asc},
	// 	{Name: "Key", Mode: table.Asc},
	// })
	t.table.SetColumnConfigs([]table.ColumnConfig{
		{

			// Number:      4,
			Name:        "Expiration",
			Transformer: removeNil,
		},
		{

			// Number:      4,
			Name:        "Set",
			Transformer: removeNil,
		},
	})
	// })

	t.table.Style().Options.SeparateRows = true

	return &t
}

func (itw *NeighborTableWriter) AppendNeighborRow(neighbor *avs.Neighbor, maxDataKeys, renderFormat, maxDataValueColWidth int) {
	row := table.Row{
		neighbor.Namespace,
		neighbor.Set,
		neighbor.Key,
		neighbor.Distance,
		neighbor.Record.Expiration,
		neighbor.Record.Generation,
	}

	tData := NewDefaultWriter(nil)
	tData.AppendHeader(table.Row{"Key", "Value"})
	tData.SetColumnConfigs(
		[]table.ColumnConfig{
			{
				Name:        "Value",
				Transformer: vectorFormat,
				WidthMax:    maxDataValueColWidth,
			},
		},
	)
	keys := make([]string, 0, len(neighbor.Record.Data))

	for key := range neighbor.Record.Data {
		keys = append(keys, key)
	}

	slices.Sort(keys)

	for i, key := range keys {
		if maxDataKeys != 0 && i >= maxDataKeys {
			tData.AppendRow(table.Row{"...", "..."})

			break
		}
		tData.AppendRow(table.Row{key, neighbor.Record.Data[key]})
	}

	if renderFormat == RenderFormatCSV {
		row = append(row, tData.RenderCSV())
	} else {
		row = append(row, tData.Render())
	}

	itw.table.AppendRow(row)
}

func (itw *NeighborTableWriter) Render(renderFormat int) {
	if renderFormat == RenderFormatCSV {
		itw.table.RenderCSV()
	} else {
		itw.table.Render()
	}
}
