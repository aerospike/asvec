package cmd

import (
	"asvec/cmd/writers"
	"fmt"
	"io"
	"log/slog"

	"github.com/aerospike/aerospike-proximus-client-go/protos"
)

type View struct {
	writer io.Writer
	logger *slog.Logger
}

func NewView(writer io.Writer, logger *slog.Logger) *View {
	return &View{writer: writer, logger: logger}
}

func (v *View) Print(a ...any) {
	_, err := fmt.Fprint(v.writer, a...)
	if err != nil {
		panic(err)
	}

	v.Newline()
}

func (v *View) Printf(f string, a ...any) {
	s := fmt.Sprintf(f, a...)

	_, err := v.writer.Write([]byte(s))
	if err != nil {
		panic(err)
	}

	v.Newline()
}

func (v *View) Newline() {
	_, err := v.writer.Write([]byte("\n"))
	if err != nil {
		panic(err)
	}
}

func (v *View) getIndexListWriter(verbose bool) *writers.IndexTableWriter {
	return writers.NewIndexTableWriter(v.writer, verbose, v.logger)
}

func (v *View) PrintIndexes(
	indexList *protos.IndexDefinitionList,
	indexStatusList []*protos.IndexStatusResponse,
	verbose bool,
) {
	t := v.getIndexListWriter(verbose)

	for i, index := range indexList.Indices {
		if index.Id.Name == "" || index.Id.Namespace == "" {
			continue
		}

		t.AppendIndexRow(index, indexStatusList[i])
	}

	t.Render()
}
