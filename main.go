package main

import (
	"bufio"
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
)

func closeup(name string, f *os.File, rerr *error) func() {
	return func() {
		if cerr := f.Close(); *rerr == nil {
			*rerr = fmt.Errorf("failed to close %q: %v", name, cerr)
		}
	}
}

func procio(r io.Reader, w io.Writer, info map[string]string) error {
	bld := builder{
		Trans: make(transSymbols),
		Info:  info,
		Dict:  newDict(),
	}
	gs := gutenScan{
		sc: bufio.NewScanner(r),
		// res: dumper{},
		res: newExtractor(&bld),
	}
	if err := gs.scan(); err != nil {
		return err
	}
	enc := json.NewEncoder(w)
	// enc.SetIndent("", "  ")
	return enc.Encode(&bld)
}

func process(nin string) {
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

		if err := procio(fin, fout, map[string]string{
			"sourceFile": nin,
		}); err != nil {
			return err
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
		if err := procio(os.Stdin, os.Stdout, map[string]string{
			"sourceFile": "<stdin>",
		}); err != nil {
			log.Fatalln(err)
		}
	}

	var wg sync.WaitGroup

	N := runtime.GOMAXPROCS(-1)
	toProc := make(chan string, N)

	for i := 0; i < N; i++ {
		go func(toProc <-chan string) {
			for arg := range toProc {
				process(arg)
				wg.Done()
			}
		}(toProc)
	}

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
	return

}
