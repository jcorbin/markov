package main

import (
	"encoding/json"
)

type weightedSymbols map[symbol]uint
type transSymbols map[symbol]weightedSymbols

func (ts transSymbols) addChain(chain []symbol) {
	var last symbol
	for _, sym := range chain {
		ts.add(last, sym)
		last = sym
	}
	ts.add(last, symbol(0))
}

func (ts transSymbols) add(a, b symbol) {
	ws := ts[a]
	if ws == nil {
		ws = make(weightedSymbols, 1)
		ts[a] = ws
	}
	ws[b]++
}

func (ts transSymbols) MarshalJSON() ([]byte, error) {
	type jsonWS struct {
		Weight uint   `json:"weight"`
		Symbol symbol `json:"symbol"`
	}
	type jsonTS struct {
		FromSym symbol   `json:"fromSym"`
		ToSym   []jsonWS `json:"toSym"`
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
