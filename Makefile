GUTENROOT ?= ~/gutenberg

guten-mine:
	go build ./tools/guten-mine

PHONY: docs.json
docs.json: guten-mine
	time find $(GUTENROOT) -type f -name '*.txt' | ./guten-mine -stdin >$@ 2>$@.log

clean:
	rm guten-mine
	rm docs.json docs.json.log
