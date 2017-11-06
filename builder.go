package main

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/jcorbin/markov/internal/symbol"
)

type markovLang struct {
	Dict  *symbol.Dict `json:"dictionary"`
	Trans transSymbols `json:"transitions"`
}

func makeMarkovLang() markovLang {
	return markovLang{
		Dict:  symbol.NewDict(),
		Trans: make(transSymbols),
	}
}

type builder struct {
	Title string            `json:"title"`
	Info  map[string]string `json:"info"`
	Lang  markovLang        `json:"language"`

	chain []symbol.Symbol
}

func (bld *builder) title(title string) error {
	bld.Title = title
	return nil
}

func (bld *builder) info(info map[string]string) error {
	if bld.Info == nil {
		bld.Info = info
		return nil
	}
	for k, v := range info {
		bld.Info[k] = v
	}
	return nil
}

func (bld *builder) token(tok []byte) error {
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
	bld.chain = append(bld.chain, bld.Lang.Dict.Add(stok))
	return nil
}

func (bld *builder) flush() error {
	bld.Lang.Trans.addChain(bld.chain)
	bld.chain = bld.chain[:0]
	return nil
}

var _ extractResultor = &builder{}
