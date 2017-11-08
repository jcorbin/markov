# Markov Chaining Around Project Gutenberg

The goal: generate random books

The setup:
- download all of project gutenberg and extract from it
- a markov transition table from each book's text
- a markov transition table from all books' titles

Then generate a random title by chaining through the title transition table.

Collect a list of every document that could have contributed each non-trivial
word to that random title. Now merge all of their markov transition tables into
a new table.

Then generate some body text from the title-derived combined transition table.

## How Do

Download all of Project Gutenberg (BE KIND use a mirror! TODO link
suggestion); unpack it to `$HOME/gutenberg` (or if you choose
another place, `export GUTENROOT=.../that/path`).

Build All The Things ™ with `make bins`.

Mine markov goodness out with a `make all.db`; currently takes ~7
minutes on my machine.

Genearte a book by running `./bin/gen-book all.db/index.json` .

## Rambling on Possibilities

So far title generation has worked better than expected; however it might be
useful to explore adding artificial "mash-up" rules. E.g. currently the book
"Don Quixote" will never generate an interesting title; what if we had "Don
Quixote and The ...".

Improved structural generation, including section headers, better punctuation
etc. Maybe try to extract entities, store them in a list, and then fill
entities in to generated structure. Maybe try a hierarchical model, e.g.
different transition table within quotations...

Explore generating random words from N-grams, or phoneme chaining.
