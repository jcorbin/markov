package model

import "github.com/jcorbin/markov/internal/symbol"

// Lang represents a language as its dictionary and transition table.
type Lang struct {
	Dict  *symbol.Dict `json:"dictionary"`
	Trans Trans        `json:"transitions"`
}

// MakeLang creates a new lang.
func MakeLang() Lang {
	return Lang{
		Dict:  symbol.NewDict(),
		Trans: make(Trans),
	}
}

// TODO: merging
