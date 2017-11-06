package main

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

type builder struct {
	Title string            `json:"title"`
	Info  map[string]string `json:"info"`
	Dict  *dict             `json:"dictionary"`
	Trans transSymbols      `json:"transitions"`

	chain []symbol
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
	bld.chain = append(bld.chain, bld.Dict.add(stok))
	return nil
}

func (bld *builder) flush() error {
	bld.Trans.addChain(bld.chain)
	bld.chain = bld.chain[:0]
	return nil
}

var _ extractResultor = &builder{}
