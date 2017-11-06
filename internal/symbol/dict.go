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

// NewDict creates a new dict with the 0 Symbol mapped to "".
func NewDict() *Dict {
	return &Dict{
		str2sym: map[string]Symbol{"": 0},
		sym2str: []string{""},
	}
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

// Get looks up a Symbol, returing its string and a bool defined flag; if not
// defined, the empty sting is returned.
func (d *Dict) Get(sym Symbol) (string, bool) {
	if int(sym) < len(d.sym2str) {
		return d.sym2str[sym], true
	}
	return "", false
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
