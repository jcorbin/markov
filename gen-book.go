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

func genBook(title string, lng model.Lang, w io.Writer) error {
	// TODO: proper content generation: needs paragraph, eof markers, maybe
	// even section headers.
	var buf bytes.Buffer
	for n := 0; n < 100; {
		first := true
		_ = lng.Trans.GenChain(rng, func(sym symbol.Symbol) error {
			if sym == 0 {
				_, _ = buf.WriteRune('.')
				return nil
			}
			word := lng.Dict.ToString(sym)
			if n := buf.Len(); n+len(word) > 79 {
				buf.WriteRune('\n')
				_, err := w.Write(buf.Bytes())
				if err != nil {
					return err
				}
				buf.Reset()
			} else {
				if n > 0 {
					_, _ = buf.WriteRune(' ')
				}
				if first {
					_, _ = buf.WriteString(strings.Title(word))
					first = false
				} else {
					_, _ = buf.WriteString(word)
				}
			}
			n++
			return nil
		})
	}
	if buf.Len() > 0 {
		buf.WriteRune('\n')
		_, err := w.Write(buf.Bytes())
		if err != nil {
			return err
		}
		buf.Reset()
	}
	return nil
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

	if err := func(r io.Reader) error {
		dec := json.NewDecoder(r)
		if err := dec.Decode(&db); err != nil {
			return err
		}

		if numDocsToCollect != 0 {
			suchDocs, err := collectDocs(numDocsToCollect)
			if err == nil {
				for _, id := range suchDocs.SortedIDs() {
					di := db.Docs[id]
					fmt.Printf("%s\n", di.SourceFile)
				}
			}
			return err
		}

		title, docs, err := genTitle()
		if err != nil {
			return err
		}

		title = strings.Title(title)
		fmt.Printf("Title: %q\n", title)
		for id := range docs {
			fmt.Printf("- %q\n", id)
		}

		lng, err := db.MergedDocLang(docs)
		if err != nil {
			return err
		}
		return genBook(title, lng, os.Stdout)
	}(os.Stdin); err != nil {
		log.Fatalln(err)
	}
}
