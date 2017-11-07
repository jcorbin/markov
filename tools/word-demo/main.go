package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"

	"github.com/jcorbin/markov/internal/model"
	"github.com/jcorbin/markov/internal/symbol"
)

func main() {
	if err := func(r io.Reader) error {
		dec := json.NewDecoder(r)
		var db model.DocDB
		if err := dec.Decode(&db); err != nil {
			return err
		}

		for _, di := range db.Docs {
			log.Println(di.Title)
			d, err := di.Load()
			if err != nil {
				return err
			}

			// build a 2-gram markov model of the "word language"
			wordLang := model.MakeLang()
			_ = d.Lang.Dict.Each(func(sym symbol.Symbol, str string) error {
				var last symbol.Symbol
				var prior = []rune{0, 0}
				for _, r := range str {
					next := wordLang.Dict.Add(string(r))
					wordLang.Trans.Add(last, next, 1)
					prior[0], prior[1] = prior[1], r
					last = wordLang.Dict.Add(string(prior))
				}
				wordLang.Trans.Add(last, symbol.Symbol(0), 1)
				return nil
			})

			// XXX QED: generate 10 random "words" by chaining 2-grams; use a
			// couple of quality heuristics (at least 5 chars, no non-ascii
			// codepoints...)
			for i := 0; i < 10; i++ {
				word := ""
				var prior = []rune{0, 0}
				_ = wordLang.Trans.GenReducedChain(
					rand.New(rand.NewSource(rand.Int63())),
					func(sym symbol.Symbol) (symbol.Symbol, error) {
						if sym == 0 {
							return sym, nil
						}

						str, def := wordLang.Dict.Get(sym)
						if !def {
							return sym, fmt.Errorf("undefined next symbol %v", sym)
						}
						fmt.Printf("%q => %q\n", prior, str)

						for _, r := range str {
							prior[0], prior[1] = prior[1], r
							sym, def = wordLang.Dict.GetSym(string(prior))
							if !def {
								return sym, fmt.Errorf("undefined next-gen symbol %v", sym)
							}
							word += string(r)
							return sym, nil
						}

						return 0, nil
					},
				)

				if !func() bool {
					for _, r := range word {
						if r > 0x7f {
							return false
						}
					}
					if len(word) < 5 {
						return false
					}
					return true

				}() {
					fmt.Printf("NOPE: %q\n\n", word)
					i--
					continue
				}

				fmt.Printf("WHY NOT: %q\n\n", word)
			}

			break
		}

		return nil
	}(os.Stdin); err != nil {
		log.Fatalln(err)
	}
}
