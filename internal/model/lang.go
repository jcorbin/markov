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

// Merge merges another language into a copy of this one, returning the new
// copy.
func (lng Lang) Merge(other Lang) Lang {
	rewrite, dict := lng.Dict.Merge(other.Dict)
	return Lang{
		Dict:  dict,
		Trans: lng.Trans.Merge(other.Trans, rewrite),
	}
}
