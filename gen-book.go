package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"strings"

	"github.com/jcorbin/markov/internal/model"
	"github.com/jcorbin/markov/internal/symbol"
)

var (
	db  model.DocDB
	rng *rand.Rand
)

func init() {
	rng = rand.New(rand.NewSource(rand.Int63()))
}

func genTitle() (string, model.SupportDocIDs, error) {
	for i := 0; i < 100; i++ {
		title, docs := db.GenTitle(rng)
		for id, word := range docs {
			if len(word) <= 3 {
				delete(docs, id)
			}
		}
		if len(docs) > 9 {
			return title, docs, nil
		}
	}
	return "", nil, errors.New("unable to produce acceptable title")
}

type bookGen struct {
	buf bytes.Buffer
	w   io.Writer
}

func genBook(title string, lng model.Lang, w io.Writer) error {
	const limit = 10000 // TODO: parameter

	var stop = errors.New("done")
	var bg bookGen
	bg.w = w

	// TODO: handle punctuation better
	// TODO: maybe section headers

	chainLength := 0

	first := true
	err := lng.Trans.GenChain(rng, func(sym symbol.Symbol) error {
		switch sym {
		case 0, symbol.EOF:
			return stop
		case symbol.GS:
			first = true
			bg.buf.WriteRune('\n')
			return bg.flush()
		}

		word := lng.Dict.ToString(sym)
		if n := bg.buf.Len(); n+len(word) > 79 {
			if err := bg.flush(); err != nil {
				return err
			}
		} else {
			if n > 0 {
				_, _ = bg.buf.WriteRune(' ')
			}
			if first {
				_, _ = bg.buf.WriteString(strings.Title(word))
				first = false
			} else {
				_, _ = bg.buf.WriteString(word)
			}
		}

		chainLength++
		switch word {
		case ".", "!", "?":
			first = true
			if chainLength >= limit {
				// TODO: approach / generate / work-in EOF more naturally
				bg.buf.WriteRune('\n')
				if err := bg.flush(); err != nil {
					return err
				}
				fmt.Fprintf(&bg.buf, "-- Cut off by editorial oversight: exceeded %v words", limit)
				if err := bg.flush(); err != nil {
					return err
				}
				return stop
			}
		}

		return nil
	})

	if err == nil || err == stop {
		err = bg.flush()
	}

	return err
}

func (bg *bookGen) flush() error {
	if bg.buf.Len() == 0 {
		return nil
	}
	bg.buf.WriteRune('\n')
	_, err := bg.w.Write(bg.buf.Bytes())
	bg.buf.Reset()
	return err
}

func collectDocs(atLeast int) (model.SupportDocIDs, error) {
	suchDocs := make(model.SupportDocIDs, 2*atLeast)
	i := 0
	for ; i < 4*atLeast && len(suchDocs) < atLeast; i++ {
		title, docs, err := genTitle()
		if err != nil {
			return nil, err
		}
		n := 0
		for id, word := range docs {
			if m := len(suchDocs[id]); m < len(word) {
				suchDocs[id] = word
				if m == 0 {
					n++
				}
			}
		}
		log.Printf("added %v docs from %q", n, title)
	}
	log.Printf("collected %v docs from %v titles", len(suchDocs), i)
	return suchDocs, nil
}

func main() {
	// useful to prepaare a small set of documents to iterate on extraction
	var numDocsToCollect int
	flag.IntVar(&numDocsToCollect, "collectDocs", 0, "collect and print a list of useful document source files")

	flag.Parse()

	dbFile := os.Stdin
	if args := flag.Args(); len(args) > 0 {
		f, err := os.Open(args[0])
		if err != nil {
			log.Fatalf("Failed to read %q: %v", args[0], err)
		}
		dbFile = f
	}

	if err := func(r io.Reader, w io.Writer) error {
		dec := json.NewDecoder(r)
		if err := dec.Decode(&db); err != nil {
			return err
		}

		if numDocsToCollect != 0 {
			suchDocs, err := collectDocs(numDocsToCollect)
			if err == nil {
				for _, id := range suchDocs.SortedIDs() {
					di := db.Docs[id]
					fmt.Fprintf(w, "%s\n", di.SourceFile)
				}
			}
			return err
		}

		title, docs, err := genTitle()
		if err != nil {
			return err
		}

		title = strings.Title(title)
		fmt.Fprintf(w, "Title: %q\n", title)
		for id := range docs {
			fmt.Fprintf(w, "- %q\n", id)
		}

		lng, err := db.MergedDocLang(docs)
		if err != nil {
			return err
		}
		return genBook(title, lng, w)
	}(dbFile, os.Stdout); err != nil {
		log.Fatalln(err)
	}
}
