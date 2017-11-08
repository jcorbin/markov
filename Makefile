GUTENROOT ?= ~/gutenberg

BINS=bin/guten-mine bin/word-demo

bins: $(BINS)

bin/%: tools/%
	[ -d bin ] || mkdir bin
	go build -o $@ ./$<

PHONY: docs.json
docs.json: guten-mine
	time find $(GUTENROOT) -type f -name '*.txt' | ./guten-mine -stdin >$@ 2>$@.log

clean:
	rm docs.json docs.json.log
	rm -f $(BINS)
