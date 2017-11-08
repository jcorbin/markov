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

		title, docs, err := g.GenTitle()
		if err != nil {
			return err
		}

		return g.GenBook(title, docs, w)
	}(dbFile, os.Stdout); err != nil {
		log.Fatalln(err)
	}
}
