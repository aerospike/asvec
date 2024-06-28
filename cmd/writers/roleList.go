package writers

import (
	"io"
	"log/slog"

	"github.com/aerospike/avs-client-go/protos"
	"github.com/jedib0t/go-pretty/v6/table"
)

type RoleTableWriter struct {
	table.Writer
	logger *slog.Logger
}

func NewRoleTableWriter(writer io.Writer, logger *slog.Logger) *RoleTableWriter {
	t := RoleTableWriter{NewDefaultWriter(writer), logger}

	t.AppendHeader(table.Row{"Roles"}, rowConfigAutoMerge)

	// t.SetTitle("Roles")
	t.SetAutoIndex(true)
	t.SortBy([]table.SortBy{
		{Name: "Roles", Mode: table.Asc},
		{Name: "User", Mode: table.Asc},
	})

	// t.Style().Options.SeparateRows = true

	return &t
}

func (itw *RoleTableWriter) AppendRoleRow(role *protos.Role) {
	itw.AppendRow(table.Row{role.GetId()})
}
