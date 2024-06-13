package writers

import (
	"io"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

func NewDefaultWriter(writer io.Writer) table.Writer {
	t := table.NewWriter()

	t.SetOutputMirror(writer)
	t.AppendSeparator()
	t.SuppressEmptyColumns()
	t.SetStyle(table.StyleRounded)

	t.Style().Title.Colors = append(t.Style().Title.Colors, text.Bold)

	t.Style().Title.Align = text.AlignCenter

	return t
}
