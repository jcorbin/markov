package main

import (
	"encoding/json"
	"flag"
	"io"
	"log"
	"os"

	"github.com/jcorbin/markov/internal/gen"
	"github.com/jcorbin/markov/internal/model"
)

func main() {
	// useful to prepaare a small set of documents to iterate on extraction
	var atLeast int
	flag.IntVar(&atLeast, "atLeast", 10, "list at least this many documents")

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
		var db model.DocDB
		dec := json.NewDecoder(r)
		if err := dec.Decode(&db); err != nil {
			return err
		}
		g := gen.New(db)

		suchDocs := make(model.SupportDocIDs, 2*atLeast)

		i := 0
		for ; i < 4*atLeast && len(suchDocs) < atLeast; i++ {
			title, docs, err := g.GenTitle()
			if err != nil {
				return err
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

		var err error
		for enc, i, ids := json.NewEncoder(w), 0, suchDocs.SortedIDs(); err == nil && i < len(ids); {
			id := ids[i]
			di := db.Docs[id]
			err = enc.Encode(di)
			i++
		}
		return err
	}(dbFile, os.Stdout); err != nil {
		log.Fatalln(err)
	}
}
