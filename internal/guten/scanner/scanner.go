package scanner

import (
	"bufio"
	"bytes"
	"io"
	"regexp"
)

// TODO: support meta "Character set encoding"
// 6923 'ASCII'
// 5742 'ISO-8859-1'
//  709 'UTF-8'
//   83 'US-ASCII'
//   21 'ISO
//    5 'Unicode
//    5 'ISO-8859-2'
//    2 'ISO-8859-3'
//    2 'ISO-8859-15'
//    2 'CP-1252'

var (
	markPat     = regexp.MustCompile(`^\*{3}\s+(.+?)\s+\*{3}$`)
	standEndPat = regexp.MustCompile(`(?:START|(END)) OF THIS PROJECT GUTENBERG EBOOK (.+)$`)
)

// Resultor is the interface implemnetd to receive results from Scanner.
type Resultor interface {
	Slug(s string) error
	Meta(key, val string) error
	Mark(s string) error
	Boundary(end bool, name string) error
	Data(buf []byte) error
	Close() error
}

// Scanner scans a Project Gutenberg e-book.
type Scanner struct {
	sc  *bufio.Scanner
	res Resultor
}

// New creates a new Scanner that will scan from the given io.Reader and call
// methods on the given resultor.
func New(r io.Reader, res Resultor) *Scanner {
	return &Scanner{
		sc:  bufio.NewScanner(r),
		res: res,
	}
}

// Scan performs the scan.
func (sc Scanner) Scan() (err error) {
	defer func() {
		if cerr := sc.res.Close(); err == nil {
			err = cerr
		}
	}()
	for _, f := range []func() error{
		sc.scanFirst,
		sc.scanMeta,
		sc.scanBody,
	} {
		if err := f(); err != nil {
			return err
		}
	}
	return sc.sc.Err()
}

func (sc Scanner) handleMark() (bool, error) {
	m := markPat.FindSubmatch(sc.sc.Bytes())
	if len(m) == 0 {
		return false, nil
	}

	if sem := standEndPat.FindSubmatch(m[1]); len(sem) > 0 {
		end := len(sem[1]) > 0
		return true, sc.res.Boundary(end, string(sem[2]))
	}

	// XXX e.g. "*** START: FULL LICENSE ***"
	return true, sc.res.Mark(string(m[1]))
}

func (sc Scanner) scanMeta() error {

	// scan key
	for sc.sc.Scan() {
		// mark ends meta section
		if mark, err := sc.handleMark(); err != nil {
			return err
		} else if mark {
			return nil
		}

		// skip blank lines
		bline := sc.sc.Bytes()
		if t := bytes.TrimSpace(bline); len(t) == 0 {
			continue
		}

		// detect key: val
		off := bytes.Index(bline, []byte(": "))

		if off <= 0 {
			// emit non key-val
			if err := sc.res.Data(sc.sc.Bytes()); err != nil {
				return err
			}
			continue
		}

		key := string(bline[:off])

		// continue scanning val
		off += 2 // len(": ")
		val := string(bline[off:])
		for sc.sc.Scan() {
			if bline := sc.sc.Bytes(); len(bline) >= off && len(bytes.TrimSpace(bline[:off])) == 0 {
				val += "\n" + string(bline[off:])
				continue
			}
			if err := sc.res.Meta(key, val); err != nil {
				return err
			}
			break
		}
	}

	return sc.sc.Err()
}

func (sc Scanner) scanBody() error {
	for sc.sc.Scan() {
		// handle marks
		if _, err := sc.handleMark(); err != nil {
			return err
		}
		if err := sc.res.Data(sc.sc.Bytes()); err != nil {
			return err
		}
	}
	return sc.sc.Err()
}

func (sc Scanner) scanFirst() error {
	var first string
	for sc.sc.Scan() {
		t := bytes.TrimSpace(sc.sc.Bytes())
		if len(t) == 0 {
			break
		}
		if len(first) > 0 {
			first += " " + string(t)
		} else {
			first = string(t)
		}
	}
	return sc.res.Slug(first)
}
