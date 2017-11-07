package main

import (
	"bytes"
	"encoding/json"
	"errors"
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

func genBook(title string, lng model.Lang) error {
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
				fmt.Printf("%s\n", buf.Bytes())
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
		fmt.Printf("%s\n", buf.Bytes())
		buf.Reset()
	}
	return nil
}

func main() {
	if err := func(r io.Reader) error {
		dec := json.NewDecoder(r)
		if err := dec.Decode(&db); err != nil {
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
		return genBook(title, lng)
	}(os.Stdin); err != nil {
		log.Fatalln(err)
	}
}
