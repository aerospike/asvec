package writers

import (
	"io"
	"log/slog"
	"strings"

	"github.com/aerospike/avs-client-go/protos"
	"github.com/jedib0t/go-pretty/v6/table"
)

type UserTableWriter struct {
	table.Writer
	logger *slog.Logger
}

func NewUserTableWriter(writer io.Writer, logger *slog.Logger) *UserTableWriter {
	t := UserTableWriter{NewDefaultWriter(writer), logger}

	t.AppendHeader(table.Row{"User", "Roles"}, rowConfigAutoMerge)

	t.SetTitle("Users")
	t.SetAutoIndex(true)
	t.SortBy([]table.SortBy{
		{Name: "Roles", Mode: table.Asc},
		{Name: "User", Mode: table.Asc},
	})

	t.Style().Options.SeparateRows = true

	return &t
}

func (itw *UserTableWriter) AppendUserRow(user *protos.User) {
	itw.AppendRow(table.Row{user.GetUsername(), strings.Join(user.GetRoles(), ", ")})
}
