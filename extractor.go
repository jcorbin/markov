package main

import (
	"bufio"
	"bytes"
	"errors"
	"strings"
	"unicode"
	"unicode/utf8"
)

func isFormatted(bline []byte) bool {
	return len(bytes.TrimLeft(bline, " \t")) != len(bline)
}

func scanTokens(data []byte, atEOF bool) (advance int, token []byte, err error) {
	// Skip leading spaces.
	start := 0
	for width := 0; start < len(data); start += width {
		var r rune
		r, width = utf8.DecodeRune(data[start:])
		if !unicode.IsSpace(r) {
			break
		}
	}

	i := start
	r, width := utf8.DecodeRune(data[i:])
	i += width

	if unicode.IsPunct(r) {
		// scan punct token
		for ; i < len(data); i += width {
			r, width = utf8.DecodeRune(data[i:])
			if !unicode.IsPunct(r) {
				return i, data[start:i], nil
			}
		}
	} else {
		// scan non-punct token
		for ; i < len(data); i += width {
			r, width = utf8.DecodeRune(data[i:])
			if unicode.IsSpace(r) {
				return i + width, data[start:i], nil
			}
			if unicode.IsPunct(r) {
				switch r {
				case '-', '\'', '"', '`', '‘', '’':
				default:
					return i, data[start:i], nil
				}
			}
		}
	}

	// If we're at EOF, we have a final, non-empty, non-terminated word. Return it.
	if atEOF && len(data) > start {
		return len(data), data[start:], nil
	}

	// Request more data.
	return start, nil, nil
}

type extractResults interface {
	title(string) error
	token([]byte) error
}

type extractorState uint8

const (
	extractorPre extractorState = iota
	extractorBody
	extractorDone
)

type extractResultor interface {
	title(string) error
	info(map[string]string) error
	token([]byte) error
}

func newExtractor(res extractResultor) *extractor {
	return &extractor{
		be: bodyExtractor{
			handler: res.token,
		},
		res: res,
	}
}

var errPrematureClose = errors.New("premature extractor close")

type extractor struct {
	info  map[string]string
	state extractorState
	be    bodyExtractor
	res   extractResultor
}

func (e *extractor) slug(s string) error { return e.meta("SLUG", s) }
func (e *extractor) meta(key, val string) error {
	if e.info == nil {
		e.info = make(map[string]string, 1)
	}
	e.info[key] = val
	return nil
}

func (e *extractor) close() error {
	if e.state < extractorPre {
		return errPrematureClose
	}
	return e.be.close()
}

func (e *extractor) mark(s string) error {
	if e.state == extractorBody {
		e.state = extractorDone
	} // TODO else :shrug:
	return nil
}

func (e *extractor) data(buf []byte) error {
	if e.state != extractorBody {
		return nil // TODO :shrug:
	}
	return e.be.data(buf)
}

func (e *extractor) boundary(end bool, name string) error {
	switch e.state {
	case extractorPre:
		if end {
			e.state = extractorDone
		} else {
			e.state = extractorBody
			if err := e.meta("BOUNDARY_NAME", name); err != nil {
				return err
			}

			if err := e.res.info(e.info); err != nil {
				return err
			}

			title := e.info["Title"]
			if parts := strings.SplitN(title, "\n", 2); len(parts) > 1 {
				title = parts[0]
			}
			if title == "" {
				title = name
			}
			e.be.title = strings.ToLower(title)

			if err := e.res.title(title); err != nil {
				return err
			}
		}
	case extractorBody:
		// TODO: !end :shrug:
		e.state = extractorDone
	case extractorDone:
		// TODO :shrug:
	}
	return nil
}

var _ scanResultor = &extractor{}

var errEmptyBody = errors.New("empty body")

type bodyExtractor struct {
	title   string
	blanks  int
	began   bool
	buf     [][]byte
	procBuf bytes.Buffer
	handler func(token []byte) error
}

func (be *bodyExtractor) close() error {
	if !be.began {
		return errEmptyBody
	}
	return nil
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
	// TODO: expose
	// fmt.Printf("PARA\n")
	return be.proc()
}

func (be *bodyExtractor) mayBegin() error {
	if be.isBegin() {
		be.began = true
		return be.proc()
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

func (be *bodyExtractor) proc() error {
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
	sc.Split(scanTokens)
	for sc.Scan() {
		if err := be.handler(sc.Bytes()); err != nil {
			return err
		}
	}

	return sc.Err()
}
