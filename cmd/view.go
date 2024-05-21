package cmd

import (
	"fmt"
	"io"
)

type View struct {
	writer io.Writer
}

func NewView(writer io.Writer) *View {
	return &View{writer: writer}
}

func (v *View) Print(a ...any) {
	s := fmt.Sprint(a...)
	fmt.Print(s)
	v.writer.Write([]byte(fmt.Sprint(a...)))
	v.Newline()
}

func (v *View) Printf(f string, a ...any) {
	s := fmt.Sprintf(f, a...)
	v.writer.Write([]byte(s))
	v.Newline()
}

func (v *View) Newline() {
	_, err := v.writer.Write([]byte("\n"))
	if err != nil {
		panic(err)
	}
}
