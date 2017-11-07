package extractor

import (
	"bytes"
	"errors"
	"strings"

	"github.com/jcorbin/markov/internal/guten/scanner"
)

func isFormatted(bline []byte) bool {
	return len(bytes.TrimLeft(bline, " \t")) != len(bline)
}

type state uint8

const (
	statePre state = iota
	stateBody
	stateDone
)

// Resultor is the interface implemented to receive results from Extractor.
type Resultor interface {
	SetTitle(string) error
	SetInfo(map[string]string) error
	BodyResultor
}

// New creates a new extractor that will call the given Resultor; it implements
// scanner.Resultor, and expects to be called by one run of a scanner.Scanner.
func New(res Resultor) *Extractor {
	return &Extractor{
		be:  bodyExtractor{res: res},
		res: res,
	}
}

var errPrematureClose = errors.New("premature Extractor close")

// Extractor is a structural extractor for a Project Gutenberg e-book.
type Extractor struct {
	info  map[string]string
	state state
	be    bodyExtractor
	res   Resultor
}

// Slug stores the slug alongside other meta info.
func (e *Extractor) Slug(s string) error { return e.Meta("SLUG", s) }

// Meta stores key/val meta info for later use.
func (e *Extractor) Meta(key, val string) error {
	if e.info == nil {
		e.info = make(map[string]string, 1)
	}
	e.info[key] = val
	return nil
}

// Close halts extraction, returning an error if we haven't even gotten started
// (still scanning meta data).
func (e *Extractor) Close() error {
	if e.state < statePre {
		return errPrematureClose
	}
	return e.be.close()
}

// Mark sets state to terminal, closing the body area.
func (e *Extractor) Mark(s string) error {
	if e.state == stateBody {
		e.state = stateDone
	} // TODO else :shrug:
	return nil
}

// Boundary handles opening/closing the body area; it flushes meta info to the
// Resultor upon opening the body.
func (e *Extractor) Boundary(end bool, name string) error {
	switch e.state {
	case statePre:
		if end {
			e.state = stateDone
		} else {
			e.state = stateBody
			if err := e.Meta("BOUNDARY_NAME", name); err != nil {
				return err
			}

			if err := e.res.SetInfo(e.info); err != nil {
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

			if err := e.res.SetTitle(title); err != nil {
				return err
			}
		}
	case stateBody:
		// TODO: !end :shrug:
		e.state = stateDone
	case stateDone:
		// TODO :shrug:
	}
	return nil
}

// Data processes a buffer of line data: passing it on to the body extractor if
// in body state.
func (e *Extractor) Data(buf []byte) error {
	if e.state != stateBody {
		return nil // TODO :shrug:
	}
	return e.be.data(buf)
}

var _ scanner.Resultor = &Extractor{}
