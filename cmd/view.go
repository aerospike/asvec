package cmd

import (
	"asvec/cmd/writers"
	"fmt"
	"io"
	"log/slog"
	"sync/atomic"

	tableColor "github.com/jedib0t/go-pretty/v6/text"

	"github.com/aerospike/avs-client-go"
	"github.com/aerospike/avs-client-go/protos"
)

var errCode atomic.Uint32

type View struct {
	out    io.Writer
	err    io.Writer
	logger *slog.Logger
}

func NewView(out, err io.Writer, logger *slog.Logger) *View {
	return &View{out: out, err: err, logger: logger}
}

func (v *View) DisableColor() {
	tableColor.DisableColors()
}

func (v *View) Print(a ...any) {
	_, err := fmt.Fprint(v.out, a...)
	if err != nil {
		panic(err)
	}

	v.Newline()
}

func (v *View) Printf(f string, a ...any) {
	s := fmt.Sprintf(f, a...)

	v.Print(s)
}

func (v *View) PrintErr(a ...any) {
	_, err := fmt.Fprint(v.err, a...)
	if err != nil {
		panic(err)
	}

	v.Newline()
}

func (v *View) PrintfErr(f string, a ...any) {
	s := fmt.Sprintf(f, a...)

	v.PrintErr(s)
}

func (v *View) Newline() {
	_, err := v.out.Write([]byte("\n"))
	if err != nil {
		panic(err)
	}
}

func (v *View) Warning(f string) {
	errCode.Store(1)
	v.PrintErr(v.yellowString("Warning: %s", f))
}

func (v *View) Warningf(f string, a ...any) {
	errCode.Store(1)
	v.PrintfErr(v.yellowString("Warning: "+f, a...))
}

func (v *View) Error(f string) {
	errCode.Store(1)
	v.PrintfErr(v.redString("Error: %s", f))
}

func (v *View) Errorf(f string, a ...any) {
	errCode.Store(1)
	v.PrintfErr(v.redString("Error: "+f, a...))
}

func (v *View) getIndexListWriter(verbose bool) *writers.IndexTableWriter {
	return writers.NewIndexTableWriter(v.out, verbose, v.logger)
}

func (v *View) PrintIndexes(
	indexList *protos.IndexDefinitionList,
	indexStatusList []*protos.IndexStatusResponse,
	verbose bool,
	format int,
) {
	t := v.getIndexListWriter(verbose)

	for i, index := range indexList.Indices {
		if index.Id.Name == "" || index.Id.Namespace == "" {
			continue
		}

		t.AppendIndexRow(index, indexStatusList[i], format)
	}

	t.Render(format)
}

func (v *View) getUserListWriter() *writers.UserTableWriter {
	return writers.NewUserTableWriter(v.out, v.logger)
}

func (v *View) PrintUsers(usersList *protos.ListUsersResponse, format int) {
	t := v.getUserListWriter()

	for _, user := range usersList.GetUsers() {
		t.AppendUserRow(user)
	}

	t.Render(format)
}

func (v *View) getRoleListWriter() *writers.RoleTableWriter {
	return writers.NewRoleTableWriter(v.out, v.logger)
}

func (v *View) PrintRoles(usersList *protos.ListRolesResponse, format int) {
	t := v.getRoleListWriter()

	for _, role := range usersList.GetRoles() {
		t.AppendRoleRow(role)
	}

	t.Render(format)
}

func (v *View) getNodeInfoListWriter(isLB bool) *writers.NodeTableWriter {
	return writers.NewNodeTableWriter(v.out, isLB, v.logger)
}

func (v *View) PrintNodeInfoList(nodeInfos []*writers.NodeInfo, isLB bool, format int) {
	t := v.getNodeInfoListWriter(isLB)

	for _, node := range nodeInfos {
		t.AppendNodeRow(node)
	}

	t.Render(format)
}

func (v *View) getNeighborTableWriter() *writers.NeighborTableWriter {
	return writers.NewNeighborTableWriter(v.out, v.logger)
}

func (v *View) PrintQueryResults(neighbors []*avs.Neighbor, format int, maxDataKeys int) {
	t := v.getNeighborTableWriter()

	for _, n := range neighbors {
		t.AppendNeighborRow(n, maxDataKeys, format)
	}

	t.Render(format)
}

func (v *View) redString(f string, a ...any) string {
	return tableColor.FgRed.Sprint(fmt.Sprintf(f, a...))
}

func (v *View) yellowString(f string, a ...any) string {
	return tableColor.FgYellow.Sprint(fmt.Sprintf(f, a...))
}
