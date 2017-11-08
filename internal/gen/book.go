package gen

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/jcorbin/markov/internal/model"
	"github.com/jcorbin/markov/internal/symbol"
)

var errStop = errors.New("done")

func (g gen) GenBook(title string, docs model.SupportDocIDs, w io.Writer) error {
	title = strings.Title(title)
	if _, err := fmt.Fprintf(w, "Title: %q\n", title); err != nil {
		return err
	}
	if err := g.writeDocIDs(docs, w); err != nil {
		return err
	}
	lng, err := g.db.MergedDocLang(docs)
	if err != nil {
		return err
	}
	return g.write(title, lng, w)
}

func (g gen) writeDocIDs(docs model.SupportDocIDs, w io.Writer) error {
	if _, err := fmt.Fprintf(w, "\nSupporting Docs:\n"); err != nil {
		return err
	}
	for id := range docs {
		if _, err := fmt.Fprintf(w, "- %q\n", id); err != nil {
			return err
		}
	}
	_, err := w.Write([]byte("\n"))
	return err
}

func (g gen) write(title string, lng model.Lang, w io.Writer) error {
	const (
		limit    = 10000
		lineWrap = 80 - 1
	)

	head := strings.ToUpper(title)
	if n := (lineWrap - len(head)) / 2; n > 0 {
		head = strings.Repeat(" ", n) + head
	}
	if _, err := fmt.Fprintf(w, "%s\n\n", head); err != nil {
		return err
	}

	// TODO: handle punctuation better
	// TODO: maybe section headers

	// TODO: reify state, break apart the chain callback
	lg := lineGen{w: w}
	lg.buf.Grow(lineWrap + 2)
	chainLength := 0
	first := true

	err := lng.Trans.GenChain(g.rng, func(sym symbol.Symbol) error {
		switch sym {
		case 0, symbol.EOF:
			return errStop
		case symbol.GS:
			first = true
			lg.buf.WriteRune('\n')
			return lg.flush()
		}

		word := lng.Dict.ToString(sym)
		if err := lg.flushIfExceeds(len(word), lineWrap); err != nil {
			return err
		}

		if lg.buf.Len() > 0 {
			_, _ = lg.buf.WriteRune(' ')
		}
		if first {
			_, _ = lg.buf.WriteString(strings.Title(word))
			first = false
		} else {
			_, _ = lg.buf.WriteString(word)
		}

		chainLength++
		switch word {
		case ".", "!", "?":
			first = true
			if chainLength >= limit {
				// TODO: approach / generate / work-in EOF more naturally
				lg.buf.WriteRune('\n')
				if err := lg.flush(); err != nil {
					return err
				}
				fmt.Fprintf(&lg.buf, "-- Cut off by editorial oversight: exceeded %v words", limit)
				if err := lg.flush(); err != nil {
					return err
				}
				return errStop
			}
		}

		return nil
	})

	if err == nil || err == errStop {
		err = lg.flush()
	}

	return err
}
