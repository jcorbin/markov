package main

import (
	"encoding/json"
	"fmt"
)

type symbol uint

type dict struct {
	str2sym map[string]symbol
	sym2str []string
}

func newDict() *dict {
	return &dict{
		str2sym: map[string]symbol{"": 0},
		sym2str: []string{""},
	}
}

func (d *dict) add(str string) symbol {
	sym, def := d.str2sym[str]
	if !def {
		sym = symbol(len(d.sym2str))
		d.sym2str = append(d.sym2str, str)
		d.str2sym[str] = sym
	}
	return sym
}

func (d *dict) get(sym symbol) (string, bool) {
	if int(sym) < len(d.sym2str) {
		return d.sym2str[sym], true
	}
	return "", false
}

func (d *dict) toString(sym symbol) string {
	if str, def := d.get(sym); def {
		return str
	}
	return fmt.Sprintf("?+%X", sym)
}

func (d *dict) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.str2sym)
}
