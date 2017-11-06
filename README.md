# Markov Chaining Around Project Gutenberg

The goal: generate a random book

How I plan to get there:
- generate a title by markov chaining through possible titles
- generate its content by markov chaining through a transition table created by
  merging the transition table of each book that could have contributed a word
  to the title

Since generating titles that way probably won't suffice:
- I plan to augment the extracted title transition table with "mash-up" rules,
  e.g. so that singleton titles like "Don Quixote" can chain to "Don Quixote
  and The ..."
- I plan to experiment with word coining, by extracting a another form of state
  (TBD 2-grams, 3-grams, phonemes, etc); this could then provide another form
  of explosion, as tables could cross-pollinate based on contributing corpii to
  a novel title word

Such word coining is probably even more interesting within a fused table, to
coin combined words...

## Status: Under Development ™

- there isn't yet any generation...
- ...the extraction mechanism is barely working:
- on my machine, I can mine all of the ~74K files that I've downloaded from
  gutenberg in ~7 minutes; of those, only ~13K successfully extract
