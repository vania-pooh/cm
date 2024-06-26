// Package rewriter original code was copied from https://github.com/vbauerster/mpb
package rewriter

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

// ESC is the ASCII code for escape character
const ESC = 27

var (
	cursorUp           = fmt.Sprintf("%c[%dA", ESC, 1)
	clearLine          = fmt.Sprintf("%c[2K\r", ESC)
	clearCursorAndLine = cursorUp + clearLine
)

// Rewriter is a buffered writer that updates the terminal.
// The contents of writer will be flushed when Flush is called.
type Rewriter struct {
	out io.Writer

	buf       bytes.Buffer
	lineCount int
}

// New returns a new Rewriter with defaults
func New(w io.Writer) *Rewriter {
	return &Rewriter{
		out: w,
	}
}

// Flush flushes the underlying buffer
func (w *Rewriter) Flush() error {
	// Do nothing if buffer is empty
	if w.buf.Len() == 0 {
		return nil
	}
	w.clearLines()
	w.lineCount = bytes.Count(w.buf.Bytes(), []byte("\n"))
	_, err := w.out.Write(w.buf.Bytes())
	w.buf.Reset()
	return err
}

// Write save the contents of b to its buffers. The only errors returned are ones encountered while writing to the underlying buffer.
func (w *Rewriter) Write(b []byte) (n int, err error) {
	return w.buf.Write(b)
}

func (w *Rewriter) clearLines() {
	_, _ = fmt.Fprint(w.out, strings.Repeat(clearCursorAndLine, w.lineCount))
}
