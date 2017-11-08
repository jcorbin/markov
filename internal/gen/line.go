package gen

import (
	"bytes"
	"io"
)

type lineGen struct {
	buf bytes.Buffer
	w   io.Writer
}

func (lg *lineGen) flush() error {
	if lg.buf.Len() == 0 {
		return nil
	}
	lg.buf.WriteRune('\n')
	_, err := lg.w.Write(lg.buf.Bytes())
	lg.buf.Reset()
	return err
}

func (lg *lineGen) flushIfExceeds(add, limit int) error {
	n := lg.buf.Len()
	if n == 0 {
		return nil
	}
	if n+add <= limit {
		return nil
	}
	return lg.flush()
}
