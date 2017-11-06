package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"runtime"
	"strings"
	"sync"

	"github.com/jcorbin/markov/internal/guten/extractor"
	"github.com/jcorbin/markov/internal/guten/scanner"
	"github.com/jcorbin/markov/internal/symbol"
)

func closeup(name string, f *os.File, rerr *error) func() {
	return func() {
		if cerr := f.Close(); *rerr == nil {
			*rerr = fmt.Errorf("failed to close %q: %v", name, cerr)
		}
	}
}

func prefer(a, b docInfo) bool {
	aenc := a.Info["Character set encoding"]
	benc := b.Info["Character set encoding"]

	// TODO: update once we correct parse other encodings
	for _, enc := range []string{"UTF-8", "Unicode", "ASCII", "US-ASCII"} {
		if aenc == enc {
			return false
		}
		if benc == enc {
			return true
		}
	}

	// log.Printf("??? PREFER A: %+v", di.Info)
	// log.Printf("??? PREFER B: %+v", prior.Info)

	return false
}

func procio(r io.Reader, w io.Writer, info map[string]string) (builder, error) {
	bld := builder{
		Info: info,
		Lang: makeMarkovLang(),
	}
	gs := scanner.NewScanner(r, extractor.New(&bld)) // scanner.Dumper{}
	err := gs.Scan()
	if err == nil {
		enc := json.NewEncoder(w)
		// enc.SetIndent("", "  ")
		err = enc.Encode(&bld)
	}
	return bld, err
}

type docInfo struct {
	SourceFile string            `json:"sourceFile"`
	TransFile  string            `json:"transFile"`
	Title      string            `json:"title"`
	Info       map[string]string `json:"info"`
}

type docDB struct {
	Docs      map[string]docInfo  `json:"docs"`
	TitleLang markovLang          `json:"titleLang"`
	InvTW     map[string][]string `json:"invertedTitleWords"`
}

func process(nin string, doneDocs chan<- docInfo) {
	if err := func() (rerr error) {
		nout := strings.TrimSuffix(nin, path.Ext(nin)) + ".markov.json"

		fin, err := os.Open(nin)
		if err != nil {
			return fmt.Errorf("failed to open %q: %v", nin, err)
		}
		defer closeup(nin, fin, &rerr)

		fout, err := os.Create(nout)
		if err != nil {
			return fmt.Errorf("failed to create %q: %v", nout, err)
		}
		defer closeup(nout, fout, &rerr)
		defer func() {
			if rerr != nil {
				_ = os.Remove(nout)
			}
		}()

		log.Printf("processing %q", nin)

		bld, err := procio(fin, fout, map[string]string{
			"sourceFile": nin,
		})
		if err != nil {
			return err
		}

		doneDocs <- docInfo{
			SourceFile: nin,
			TransFile:  nout,
			Title:      bld.Title,
			Info:       bld.Info,
		}

		log.Printf("processed %q", nin)
		return nil
	}(); err != nil {
		log.Printf("failed to process %q: %v", nin, err)
	}
}

func main() {
	argsFromStdin := false
	flag.BoolVar(&argsFromStdin, "stdin", false, "read path args from stdin")
	flag.Parse()

	if !argsFromStdin && len(flag.Args()) == 0 {
		if _, err := procio(os.Stdin, os.Stdout, map[string]string{
			"sourceFile": "<stdin>",
		}); err != nil {
			log.Fatalln(err)
		}
	}

	var wg sync.WaitGroup

	N := runtime.GOMAXPROCS(-1)
	toProc := make(chan string, N)

	doneDocs := make(chan docInfo, 10*N)

	for i := 0; i < N; i++ {
		go func(toProc <-chan string) {
			for arg := range toProc {
				process(arg, doneDocs)
				wg.Done()
			}
		}(toProc)
	}

	db := docDB{
		Docs:      make(map[string]docInfo),
		TitleLang: makeMarkovLang(),
		InvTW:     make(map[string][]string),
	}

	docDBDone := make(chan struct{})
	go func() {
		var buf bytes.Buffer
		for di := range doneDocs {
			id := di.Title
			prior, def := db.Docs[id]

			if !def {
				// ingest the title for markov generation and inverted lookup
				buf.Reset()
				buf.WriteString(strings.ToLower(di.Title))
				sc := bufio.NewScanner(&buf)
				sc.Split(extractor.ScanTokens)
				var last symbol.Symbol
				for sc.Scan() {
					word := sc.Text()

					db.InvTW[word] = append(db.InvTW[word], id)

					sym := db.TitleLang.Dict.Add(word)
					db.TitleLang.Trans.add(last, sym)
					last = sym
				}
				db.TitleLang.Trans.add(last, symbol.Symbol(0))
			}

			if !def {
				db.Docs[id] = di
				log.Printf("Indexed %v => %+v", id, di.Info)
			} else if prefer(di, prior) {
				db.Docs[id] = di
				log.Printf("Re-Indexed %v => %+v", id, di.Info)
			}
		}
		docDBDone <- struct{}{}
	}()

	if argsFromStdin {
		sc := bufio.NewScanner(os.Stdin)
		for sc.Scan() {
			wg.Add(1)
			toProc <- sc.Text()
		}
		if err := sc.Err(); err != nil {
			log.Printf("error reading path args: %v", err)
		}
	} else {
		for _, arg := range flag.Args() {
			wg.Add(1)
			toProc <- arg
		}
	}

	close(toProc)
	wg.Wait()
	close(doneDocs)

	<-docDBDone
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(db); err != nil {
		log.Fatalln("Failed to write doc db index:", err)
	}

	return

}
