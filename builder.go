package main

import (
	"encoding/json"
	"strings"
	"unicode"
	"unicode/utf8"
)

type builder struct {
	title string
	dict  *dict
	chain []symbol
	ts    transSymbols
}

func (bld *builder) setTitle(title string) error {
	bld.title = title
	return nil
}

func (bld *builder) handle(tok []byte) error {
	// TODO: handle numeric tokens specially

	if r, width := utf8.DecodeRune(tok); width == len(tok) {
		switch r {

		case '.', '!', '?':
			return bld.flush()

		case ':': // TODO: could be a register/mode switch
			return nil

		case ',': // TODO also ingest without comma
			return nil

		case ';': // TODO also ingest as an end/start-of-chain
			return nil

		default:
			if unicode.IsPunct(r) {
				// fmt.Printf(" SKIP<%s>", tok)
				return nil
			}

		}
	}

	stok := string(tok)
	stok = strings.ToLower(stok)
	bld.chain = append(bld.chain, bld.dict.add(stok))
	return nil
}

func (bld *builder) flush() error {
	bld.ts.addChain(bld.chain)
	bld.chain = bld.chain[:0]
	return nil
}

func (bld *builder) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Title string       `json:"title"`
		Dict  *dict        `json:"dictionary"`
		Trans transSymbols `json:"transitions"`
	}{bld.title, bld.dict, bld.ts})
}
