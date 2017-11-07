package model

import (
	"encoding/json"
	"math"
	"math/rand"

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
		ws = make(WeightedSymbols, 1)
		ts[a] = ws
	}
	ws[b] += d
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

// GenChain generates a chain through the transition table, using the supplied,
// and calling the given function with each generated symbol. If the function
// returns an error, generation stop annd it is returned.
func (ts Trans) GenChain(rng *rand.Rand, f func(symbol.Symbol) error) error {
	return ts.GenReducedChain(rng, func(sym symbol.Symbol) (symbol.Symbol, error) {
		return sym, f(sym)
	})
}

// GenReducedChain generates a reduced chain through the transition table. The
// only difference from GenChain is that the function may influence the next
// symbol; i.e. by using some sort of reduction logic to combine symbols under
// higher-order language semantics (e.g. n-grams).
func (ts Trans) GenReducedChain(rng *rand.Rand, f func(symbol.Symbol) (symbol.Symbol, error)) error {
	var last symbol.Symbol
	for {
		var next symbol.Symbol
		best := 1.0
		for sym, w := range ts[last] {
			if score := math.Pow(rng.Float64(), 1/float64(w)); score < best {
				best, next = score, sym
			}
		}
		next, err := f(next)
		if err != nil {
			return err
		}
		if next == symbol.Symbol(0) {
			return nil
		}
		last = next
	}
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
