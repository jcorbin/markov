package extractor

import (
	"bufio"
	"bytes"
	"errors"
	"strings"
)

var errEmptyBody = errors.New("empty body")

// BodyResultor is the interface implement to receive body text extraction
// events.
type BodyResultor interface {
	OnToken([]byte) error
	EndParagraph() error
	Close() error
}

type bodyExtractor struct {
	title   string
	blanks  int
	began   bool
	buf     [][]byte
	procBuf bytes.Buffer
	res     BodyResultor
}

func (be *bodyExtractor) close() error {
	if !be.began {
		return errEmptyBody
	}
	return be.res.Close()
}

func (be *bodyExtractor) data(buf []byte) error {
	if len(buf) == 0 {
		be.blanks++
		return nil
	}

	if be.blanks > 1 {
		if err := be.flushSection(); err != nil {
			return err
		}
	} else if be.blanks > 0 {
		if err := be.flushPara(); err != nil {
			return err
		}
	}
	be.blanks = 0

	be.buf = append(be.buf, buf)
	return nil
}

func (be *bodyExtractor) flushSection() error {
	switch len(be.buf) {
	case 0:
		return nil
	case 1:
		if header := strings.TrimSpace(string(be.buf[0])); len(header) > 0 {
			// TODO: expose
			// fmt.Printf("SECTION: %q\n", header)

			if strings.ToLower(header) == be.title {
				// fmt.Printf("BEGIN\n")
				be.began = true
			}

			be.buf = be.buf[:0]
			return nil
		}
	}
	return be.flushPara()
}

func (be *bodyExtractor) flushPara() error {
	if len(be.buf) == 0 {
		return nil
	}
	if !be.began {
		return be.mayBegin()
	}
	return be.emitParagraph()
}

func (be *bodyExtractor) mayBegin() error {
	if be.isBegin() {
		be.began = true
		return be.emitParagraph()
	}

	// fmt.Printf("???: %q\n", be.buf)
	be.buf = be.buf[:0]
	return nil
}

func (be *bodyExtractor) isBegin() bool {
	// // only begin with a paragraph that's at least 2 lines
	// if len(be.buf) < 2 {
	// 	return false
	// }

	// // heuristic: if more than the initial line are indented, than take it as a
	// // sign of formatted pre-amble (and skip / don't begin yet)
	// for _, bline := range be.buf[1:] {
	// 	line := string(bline)
	// 	if strings.TrimLeft(line, " \t") != line {
	// 		return false
	// 	}
	// }
	// fmt.Printf("BEGIN: %q\n", be.buf)
	// return true

	return false
}

func (be *bodyExtractor) emitParagraph() error {
	defer func() {
		be.buf = be.buf[:0]
	}()

	// skip if the all lines are "formatted"
	formatted := true
	for _, bline := range be.buf {
		if !isFormatted(bline) {
			formatted = false
			break
		}
	}
	if formatted {
		// fmt.Printf("SKIP %q\n", be.buf)
		return nil
	}

	// skip any paragraph that says "project gutenberg"
	for _, bline := range be.buf {
		if bytes.Contains(bytes.ToLower(bline), []byte("project gutenberg")) {
			return nil
		}
	}

	// TODO: may still recognize section header if len == 1 , formatted, all
	// caps, title caps, etc

	be.procBuf.Reset()
	for _, p := range be.buf {
		be.procBuf.Write(p)
		be.procBuf.WriteRune('\n')
	}
	sc := bufio.NewScanner(&be.procBuf)
	sc.Split(ScanTokens)
	for sc.Scan() {
		if err := be.res.OnToken(sc.Bytes()); err != nil {
			return err
		}
	}
	if err := sc.Err(); err != nil {
		return err
	}

	return be.res.EndParagraph()
}
