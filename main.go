package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"unicode/utf8"
)

type chainBuilder struct {
	chain []symbol
}

func isTerminal(r rune) bool {
	switch r {
	case '.', '!', '?':
		return true
	}
	return false
}

func main() {
	if err := func(r io.Reader) error {
		return gutenScan{
			sc: bufio.NewScanner(r),
			// res: dumper{},
			res: newExtractor(
				func(title string) error {
					fmt.Printf("EXTRACTING %q\n", title)
					return nil
				},
				func(tok []byte) error {
					if r, width := utf8.DecodeRune(tok); width == len(tok) {
						if isTerminal(r) {
							fmt.Printf("EoC: %q\n\n", tok)
						} else {
							fmt.Printf("SKIP: %q\n", tok)
						}
					} else {
						fmt.Printf("WORD: %q\n", tok)
					}
					return nil
				},
			),
		}.scan()
	}(os.Stdin); err != nil {
		log.Fatalln(err)
	}
}
