package symbol

import (
	"encoding/json"
	"fmt"
)

// Symbol is a symbolicated string in some dictionary.
type Symbol uint

// Dict maps between Symbols and strings, allowing for lower memory usage. It
// is not safe to use from multiple goroutines.
type Dict struct {
	str2sym map[string]Symbol
	sym2str []string
}

// GS is the group separator symbol, used to signal paragraphs.
const GS = Symbol(1)
const gs = "\x1d"

// EOF is the end-of-file symbol.
const EOF = Symbol(2)
const eof = "\x1a"

// NewDict creates a new dict with the 0 Symbol mapped to "".
func NewDict() *Dict {
	return &Dict{
		str2sym: map[string]Symbol{
			"":  0,
			gs:  GS,
			eof: EOF,
			".": 3,
			"!": 4,
			"?": 5,
		},
		sym2str: []string{
			"", gs, eof,
			".", "!", "?",
		},
	}
}

// Len returns the number of defined symbols in the dictionary.
func (d *Dict) Len() int {
	return len(d.sym2str)
}

// Merge merges another dictionary into a copy of this dictionary, returning
// the new merged copy.
func (d *Dict) Merge(other *Dict) (map[Symbol]Symbol, *Dict) {
	rewrite := make(map[Symbol]Symbol, len(other.sym2str))
	for isym, otherStr := range other.sym2str {
		sym := Symbol(isym)
		myStr, def := d.Get(sym)
		if !def || myStr != otherStr {
			rewrite[sym] = sym
		}
	}

	n := len(d.sym2str) + len(rewrite)
	out := Dict{
		str2sym: make(map[string]Symbol, n),
		sym2str: make([]string, 0, n),
	}
	out.sym2str = append(out.sym2str, d.sym2str...)
	for isym, str := range d.sym2str {
		out.str2sym[str] = Symbol(isym)
	}

	for isym, str := range other.sym2str {
		sym := Symbol(isym)
		if _, def := rewrite[sym]; def {
			newSym := Symbol(len(out.sym2str))
			rewrite[sym] = newSym
			out.sym2str = append(out.sym2str, str)
			out.str2sym[str] = newSym
		}
	}

	return rewrite, &out
}

// Add adds a string, returning its Symbol.
func (d *Dict) Add(str string) Symbol {
	sym, def := d.str2sym[str]
	if !def {
		sym = Symbol(len(d.sym2str))
		d.sym2str = append(d.sym2str, str)
		d.str2sym[str] = sym
	}
	return sym
}

// GetSym gets any defined symbol for the given string, returning false and
// empty string if none.
func (d *Dict) GetSym(str string) (Symbol, bool) {
	sym, def := d.str2sym[str]
	return sym, def
}

// Get looks up a Symbol, returing its string and a bool defined flag; if not
// defined, the empty sting is returned.
func (d *Dict) Get(sym Symbol) (string, bool) {
	if int(sym) < len(d.sym2str) {
		return d.sym2str[sym], true
	}
	return "", false
}

// Each calls the given function on each symbol in the dictionary,
// in a random order, stopping on and returning any error.
func (d *Dict) Each(f func(sym Symbol, str string) error) error {
	for str, sym := range d.str2sym {
		if err := f(sym, str); err != nil {
			return err
		}
	}
	return nil
}

// ToString turns a Symbol into a string; if the Symbol isn't defined, a
// "?+HEX" string is returned.
func (d *Dict) ToString(sym Symbol) string {
	if str, def := d.Get(sym); def {
		return str
	}
	return fmt.Sprintf("?+%X", sym)
}

// MarshalJSON marshals the dictionary as JSON.
func (d *Dict) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.str2sym)
}

// UnmarshalJSON marshals the dictionary as JSON.
func (d *Dict) UnmarshalJSON(data []byte) error {
	err := json.Unmarshal(data, &d.str2sym)
	if err == nil {
		d.sym2str = make([]string, len(d.str2sym)+1)
		for str, sym := range d.str2sym {
			if int(sym) >= len(d.sym2str) {
				continue
			}
			d.sym2str[sym] = str
		}
	}
	return err
}
