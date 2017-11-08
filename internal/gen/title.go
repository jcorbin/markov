package gen

import (
	"errors"

	"github.com/jcorbin/markov/internal/model"
)

var errCantEvenTitle = errors.New("unable to produce acceptable title")

func (g gen) GenTitle() (string, model.SupportDocIDs, error) {
	const (
		numAttempts          = 100
		minSupportWordLength = 3
		minSupportDocs       = 9
	)

	for i := 0; i < numAttempts; i++ {
		title, docs := g.db.GenTitle(g.rng)
		for id, word := range docs {
			if len(word) <= minSupportWordLength {
				delete(docs, id)
			}
		}
		if len(docs) > minSupportDocs {
			return title, docs, nil
		}
	}
	return "", nil, errCantEvenTitle
}
