package model

import (
	"encoding/json"

	"github.com/jcorbin/markov/internal/symbol"
)

// Trans is a symbol transition table.
type Trans map[symbol.Symbol]WeightedSymbols

// WeightedSymbols is a weighted set of symbols.
type WeightedSymbols map[symbol.Symbol]uint

// Add adds a transition to the table, incrementing the weight for a -> b by
// the given delta.
func (ts Trans) Add(a, b symbol.Symbol, d uint) {
	ws := ts[a]
	if ws == nil {
		ws = make(WeightedSymbols, d)
		ts[a] = ws
	}
	ws[b]++
}

// AddChain adds a chain of symbols to the table.
func (ts Trans) AddChain(chain []symbol.Symbol) {
	var last symbol.Symbol
	for _, sym := range chain {
		ts.Add(last, sym, 1)
		last = sym
	}
	ts.Add(last, symbol.Symbol(0), 1)
}

type jsonWS struct {
	Weight uint          `json:"weight"`
	Symbol symbol.Symbol `json:"symbol"`
}
type jsonTS struct {
	FromSym symbol.Symbol `json:"fromSym"`
	ToSym   []jsonWS      `json:"toSym"`
}

// MarshalJSON marshals the table to JSON
func (ts Trans) MarshalJSON() ([]byte, error) {
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

// UnmarshalJSON marshals the table to JSON
func (ts *Trans) UnmarshalJSON(data []byte) error {
	var d []jsonTS
	if err := json.Unmarshal(data, &d); err != nil {
		return err
	}
	if len(d) > 0 {
		*ts = make(Trans, len(d))
	} else {
		*ts = nil
	}
	for _, jts := range d {
		if len(jts.ToSym) > 0 {
			ws := make(WeightedSymbols, len(jts.ToSym))
			for _, jws := range jts.ToSym {
				ws[jws.Symbol] = jws.Weight
			}
			(*ts)[jts.FromSym] = ws
		}
	}
	return nil
}

// TODO: generating from a table
