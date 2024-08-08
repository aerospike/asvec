package writers

import (
	"io"
	"log/slog"
	"strings"

	"github.com/aerospike/avs-client-go/protos"
	"github.com/jedib0t/go-pretty/v6/table"
)

type UserTableWriter struct {
	table  table.Writer
	logger *slog.Logger
}

func NewUserTableWriter(writer io.Writer, logger *slog.Logger) *UserTableWriter {
	t := UserTableWriter{NewDefaultWriter(writer), logger}

	t.table.AppendHeader(table.Row{"User", "Roles"}, rowConfigAutoMerge)

	t.table.SetTitle("Users")
	t.table.SetAutoIndex(true)
	t.table.SortBy([]table.SortBy{
		{Name: "Roles", Mode: table.Asc},
		{Name: "User", Mode: table.Asc},
	})

	t.table.Style().Options.SeparateRows = true

	return &t
}

func (itw *UserTableWriter) AppendUserRow(user *protos.User) {
	itw.table.AppendRow(table.Row{user.GetUsername(), strings.Join(user.GetRoles(), ", ")})
}

func (itw *UserTableWriter) Render(renderFormat int) {
	if renderFormat == RenderFormatCSV {
		itw.table.RenderCSV()
	} else {
		itw.table.Render()
	}
}
