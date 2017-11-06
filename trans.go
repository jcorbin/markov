package main

import (
	"encoding/json"

	"github.com/jcorbin/markov/internal/symbol"
)

type weightedSymbols map[symbol.Symbol]uint
type transSymbols map[symbol.Symbol]weightedSymbols

func (ts transSymbols) addChain(chain []symbol.Symbol) {
	var last symbol.Symbol
	for _, sym := range chain {
		ts.add(last, sym)
		last = sym
	}
	ts.add(last, symbol.Symbol(0))
}

func (ts transSymbols) add(a, b symbol.Symbol) {
	ws := ts[a]
	if ws == nil {
		ws = make(weightedSymbols, 1)
		ts[a] = ws
	}
	ws[b]++
}

func (ts transSymbols) MarshalJSON() ([]byte, error) {
	type jsonWS struct {
		Weight uint          `json:"weight"`
		Symbol symbol.Symbol `json:"symbol"`
	}
	type jsonTS struct {
		FromSym symbol.Symbol `json:"fromSym"`
		ToSym   []jsonWS      `json:"toSym"`
	}

	d := make([]jsonTS, 0, len(ts))
	for fromSym, ws := range ts {
		jws := make([]jsonWS, 0, len(ws))
		for toSym, weight := range ws {
			jws = append(jws, jsonWS{weight, toSym})
		}
		d = append(d, jsonTS{fromSym, jws})
	}
	return json.Marshal(d)
}
