package main

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
	"os"
)

func main() {
	if err := func(r io.Reader, w io.Writer) error {
		bld := builder{
			ts:   make(transSymbols),
			dict: newDict(),
		}
		gs := gutenScan{
			sc: bufio.NewScanner(r),
			// res: dumper{},
			res: newExtractor(bld.setTitle, bld.handle),
		}
		if err := gs.scan(); err != nil {
			return err
		}
		enc := json.NewEncoder(w)
		// enc.SetIndent("", "  ")
		return enc.Encode(&bld)
	}(os.Stdin, os.Stdout); err != nil {
		log.Fatalln(err)
	}
}
