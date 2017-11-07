package main

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/jcorbin/markov/internal/guten/extractor"
	"github.com/jcorbin/markov/internal/model"
	"github.com/jcorbin/markov/internal/symbol"
)

type builder struct {
	model.Doc

	last symbol.Symbol
}

func (bld *builder) SetTitle(title string) error {
	title = strings.Trim(title, `"'?!.`)
	bld.Title = title
	return nil
}

func (bld *builder) SetInfo(info map[string]string) error {
	if bld.Info == nil {
		bld.Info = info
		return nil
	}
	for k, v := range info {
		bld.Info[k] = v
	}
	return nil
}

func (bld *builder) OnToken(tok []byte) error {
	// TODO: handle numeric tokens specially

	if r, width := utf8.DecodeRune(tok); width == len(tok) {
		switch r {

		case '.', '!', '?':
			return bld.advance(bld.Lang.Dict.Add(string(r)))

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
	return bld.advance(bld.Lang.Dict.Add(stok))
}

func (bld *builder) advance(sym symbol.Symbol) error {
	bld.Lang.Trans.Add(bld.last, sym, 1)
	bld.last = sym
	return nil
}

var _ extractor.Resultor = &builder{}
