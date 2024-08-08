package writers

import (
	"io"
	"log/slog"

	"github.com/aerospike/avs-client-go/protos"
	"github.com/jedib0t/go-pretty/v6/table"
)

type RoleTableWriter struct {
	table  table.Writer
	logger *slog.Logger
}

func NewRoleTableWriter(writer io.Writer, logger *slog.Logger) *RoleTableWriter {
	t := RoleTableWriter{NewDefaultWriter(writer), logger}

	t.table.AppendHeader(table.Row{"Roles"}, rowConfigAutoMerge)
	t.table.SetAutoIndex(true)
	t.table.SortBy([]table.SortBy{
		{Name: "Roles", Mode: table.Asc},
		{Name: "User", Mode: table.Asc},
	})

	return &t
}

func (itw *RoleTableWriter) AppendRoleRow(role *protos.Role) {
	itw.table.AppendRow(table.Row{role.GetId()})
}

func (itw *RoleTableWriter) Render(renderFormat int) {
	if renderFormat == RenderFormatCSV {
		itw.table.RenderCSV()
	} else {
		itw.table.Render()
	}
}
