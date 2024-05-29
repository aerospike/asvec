package writers

import (
	"fmt"
	"io"

	"github.com/aerospike/aerospike-proximus-client-go/protos"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

type IndexTableWriter struct {
	table.Writer
	verbose bool
}

func NewIndexTableWriter(writer io.Writer, verbose bool) *IndexTableWriter {
	removeNil := text.Transformer(func(val interface{}) string {
		switch v := val.(type) {
		case *string:
			if v == nil {
				return ""
			}

			return *v
		default:
			return fmt.Sprintf("%v", v)
		}
	})

	t := IndexTableWriter{*NewDefaultWriter(writer), verbose}
	row := table.Row{"Name", "Namespace", "Set", "Field", "Dimensions", "Distance Metric"}

	if verbose {
		row = append(row, "Unmerged")
	}

	t.AppendHeader(row)

	t.SetTitle("Indexes")
	t.SetAutoIndex(true)
	t.SortBy([]table.SortBy{
		{Name: "Namespace", Mode: table.Asc},
		{Name: "Set", Mode: table.Asc},
		{Name: "Name", Mode: table.Asc},
	})
	t.SetColumnConfigs([]table.ColumnConfig{
		{
			Name:        "Set",
			Transformer: removeNil,
		},
	})

	return &t
}

func (itw *IndexTableWriter) AppendIndexRow(index *protos.IndexDefinition, status *protos.IndexStatusResponse) {
	row := table.Row{index.Id.Name, index.Id.Namespace, index.SetFilter, index.Field, index.Dimensions, index.VectorDistanceMetric}

	if itw.verbose {
		row = append(row, status.GetUnmergedRecordCount())
	}

	itw.AppendRow(row)
}
