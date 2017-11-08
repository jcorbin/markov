package gen

import (
	"io"
	"math/rand"

	"github.com/jcorbin/markov/internal/model"
)

// Gen is an interface for generating book
// titles and corresponding content.
type Gen interface {
	// GenTitle generates and returns a book title with supporting
	// document IDs, or fails to do so and return an error.
	GenTitle() (title string, docs model.SupportDocIDs, err error)

	// GenBook generates and writes book content to an io.Writer based
	// on a given title and supporting document ID set. Any io error
	// encountered while writing halts the process, and is returned.
	GenBook(title string, docs model.SupportDocIDs, w io.Writer) error
}

// New constructs a Gen that will generate from a database of extracted
// documents.
func New(db model.DocDB) Gen {
	return gen{
		db:  db,
		rng: rand.New(rand.NewSource(rand.Int63())),
	}
}

type gen struct {
	db  model.DocDB
	rng *rand.Rand
}
