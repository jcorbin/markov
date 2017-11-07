package model

import (
	"encoding/json"
	"math/rand"
	"os"

	"github.com/jcorbin/markov/internal/symbol"
)

// DocDB represents a database of extracted documents. It contains a markov
// language for generated a plausible document title, and an inverted index to
// map back to supporting document info from such a title.
type DocDB struct {
	Docs      map[string]DocInfo  `json:"docs"`
	TitleLang Lang                `json:"titleLang"`
	InvTW     map[string][]string `json:"invertedTitleWords"`
}

// DocInfo contains meta data for a documnet in a DocDB.
type DocInfo struct {
	SourceFile string            `json:"sourceFile"`
	TransFile  string            `json:"transFile"`
	Title      string            `json:"title"`
	Info       map[string]string `json:"info"`
}

// Doc represents an extracted document loaded from a DocInfo in a DocDB.
type Doc struct {
	Title string            `json:"title"`
	Info  map[string]string `json:"info"`
	Lang  Lang              `json:"language"`
}

// SupportDocIDs is a set of supporting document ids mapped to their longest
// supporting word.
type SupportDocIDs map[string]string

// GenTitle generates a random document title, and returns a set of supporting
// document ids.
func (db DocDB) GenTitle(rng *rand.Rand) (string, SupportDocIDs) {
	title, docs := "", make(map[string]string)
	_ = db.TitleLang.Trans.GenChain(rng, func(sym symbol.Symbol) error {
		if sym == 0 {
			return nil
		}
		word := db.TitleLang.Dict.ToString(sym)
		for _, id := range db.InvTW[word] {
			if len(word) > len(docs[id]) {
				docs[id] = word
			}
		}
		if title == "" {
			title = word
		} else {
			title += " " + word
		}
		return nil
	})
	return title, docs
}

// Load loads the extracted document from TransFile; subsequent calls to Load
// may return the same *Doc pointer.
func (di DocInfo) Load() (rd *Doc, rerr error) {
	// TODO: caching
	f, err := os.Open(di.TransFile)
	if err != nil {
		return nil, err
	}
	defer func() {
		if cerr := f.Close(); rerr == nil {
			rerr = cerr
		}
	}()
	var d Doc
	dec := json.NewDecoder(f)
	if err := dec.Decode(&d); err != nil {
		return nil, err
	}
	return &d, nil
}
