package main

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"
)

type scanResultor interface {
	slug(s string) error
	meta(key, val string) error
	mark(s string) error
	boundary(end bool, name string) error
	data(buf []byte) error
}

var (
	markPat     = regexp.MustCompile(`^\*{3}\s+(.+?)\s+\*{3}$`)
	standEndPat = regexp.MustCompile(`(?:START|(END)) OF THIS PROJECT GUTENBERG EBOOK (.+)$`)
)

type gutenScan struct {
	sc  *bufio.Scanner
	res scanResultor
}

func (gs gutenScan) scan() error {
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
	for {
		var key, val string
		var off int

		// scan key
		for gs.sc.Scan() {
			// mark ends meta section
			if mark, err := gs.handleMark(); err != nil {
				return err
			} else if mark {
				return nil
			}

			// skip blank lines
			line := gs.sc.Text()
			if t := strings.TrimSpace(line); len(t) == 0 {
				continue
			}

			// detect key: val
			if off = strings.Index(line, ": "); off > 0 {
				key = line[:off]
				off += 2 // len(": ")
				val = line[off:]
				break
			}

			// emit non key-val
			if err := gs.res.data(gs.sc.Bytes()); err != nil {
				return err
			}
		}

		// continue scanning val
		for gs.sc.Scan() {
			if line := gs.sc.Text(); len(line) >= off && len(strings.TrimSpace(line[:off])) == 0 {
				val += "\n" + line[off:]
				continue
			}
			if err := gs.res.meta(key, val); err != nil {
				return err
			}
			break
		}
	}
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
		line := gs.sc.Text()
		t := strings.TrimSpace(line)
		if len(t) == 0 {
			break
		}
		if len(first) > 0 {
			first += " " + t
		} else {
			first = t
		}
	}
	return gs.res.slug(first)
}

type dumper struct{}

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
