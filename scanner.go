package main

import (
	"bufio"
	"bytes"
	"fmt"
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

type scanResultor interface {
	slug(s string) error
	meta(key, val string) error
	mark(s string) error
	boundary(end bool, name string) error
	data(buf []byte) error
	close() error
}

var (
	markPat     = regexp.MustCompile(`^\*{3}\s+(.+?)\s+\*{3}$`)
	standEndPat = regexp.MustCompile(`(?:START|(END)) OF THIS PROJECT GUTENBERG EBOOK (.+)$`)
)

type gutenScan struct {
	sc  *bufio.Scanner
	res scanResultor
}

func (gs gutenScan) scan() (err error) {
	defer func() {
		if cerr := gs.res.close(); err == nil {
			err = cerr
		}
	}()
	for _, f := range []func() error{
		gs.scanFirst,
		gs.scanMeta,
		gs.scanBody,
	} {
		if err := f(); err != nil {
			return err
		}
	}
	return gs.sc.Err()
}

func (gs gutenScan) handleMark() (bool, error) {
	m := markPat.FindSubmatch(gs.sc.Bytes())
	if len(m) == 0 {
		return false, nil
	}

	if sem := standEndPat.FindSubmatch(m[1]); len(sem) > 0 {
		end := len(sem[1]) > 0
		return true, gs.res.boundary(end, string(sem[2]))
	}

	// XXX e.g. "*** START: FULL LICENSE ***"
	return true, gs.res.mark(string(m[1]))
}

func (gs gutenScan) scanMeta() error {

	// scan key
	for gs.sc.Scan() {
		// mark ends meta section
		if mark, err := gs.handleMark(); err != nil {
			return err
		} else if mark {
			return nil
		}

		// skip blank lines
		bline := gs.sc.Bytes()
		if t := bytes.TrimSpace(bline); len(t) == 0 {
			continue
		}

		// detect key: val
		off := bytes.Index(bline, []byte(": "))

		if off <= 0 {
			// emit non key-val
			if err := gs.res.data(gs.sc.Bytes()); err != nil {
				return err
			}
			continue
		}

		key := string(bline[:off])

		// continue scanning val
		off += 2 // len(": ")
		val := string(bline[off:])
		for gs.sc.Scan() {
			if bline := gs.sc.Bytes(); len(bline) >= off && len(bytes.TrimSpace(bline[:off])) == 0 {
				val += "\n" + string(bline[off:])
				continue
			}
			if err := gs.res.meta(key, val); err != nil {
				return err
			}
			break
		}
	}

	return gs.sc.Err()
}

func (gs gutenScan) scanBody() error {
	for gs.sc.Scan() {
		// handle marks
		if _, err := gs.handleMark(); err != nil {
			return err
		}
		if err := gs.res.data(gs.sc.Bytes()); err != nil {
			return err
		}
	}
	return gs.sc.Err()
}

func (gs gutenScan) scanFirst() error {
	var first string
	for gs.sc.Scan() {
		t := bytes.TrimSpace(gs.sc.Bytes())
		if len(t) == 0 {
			break
		}
		if len(first) > 0 {
			first += " " + string(t)
		} else {
			first = string(t)
		}
	}
	return gs.res.slug(first)
}

type dumper struct{}

func (d dumper) close() error {
	_, err := fmt.Printf("DONE\n")
	return err
}

func (d dumper) slug(s string) error {
	_, err := fmt.Printf("slug: %q\n", s)
	return err
}

func (d dumper) meta(key, val string) error {
	_, err := fmt.Printf("%q => %q\n", key, val)
	return err
}

func (d dumper) mark(s string) error {
	_, err := fmt.Printf("MARK: %q\n", s)
	return err
}

func (d dumper) data(buf []byte) error {
	_, err := fmt.Printf("NOPE: %q\n", buf)
	return err
}

func (d dumper) boundary(end bool, name string) error {
	kind := "Start"
	if end {
		kind = "End"
	}
	_, err := fmt.Printf("# %s: %q\n", kind, name)
	return err
}

var _ scanResultor = dumper{}
