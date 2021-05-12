package printer

import (
	"fmt"
	"io"
)

type Printer struct {
	W io.Writer
	E io.Writer
}

func (p Printer) Print(a ...interface{}) {
	_, _ = fmt.Fprint(p.W, a...)
}

func (p Printer) Println(a ...interface{}) {
	_, _ = fmt.Fprintln(p.W, a...)
}

func (p Printer) Printf(format string, a ...interface{}) {
	_, _ = fmt.Fprintf(p.W, format, a...)
}

func (p Printer) ErrPrint(a ...interface{}) {
	_, _ = fmt.Fprint(p.W, a...)
}

func (p Printer) ErrPrintln(a ...interface{}) {
	_, _ = fmt.Fprintln(p.W, a...)
}

func (p Printer) ErrPrintf(format string, a ...interface{}) {
	_, _ = fmt.Fprintf(p.W, format, a...)
}
