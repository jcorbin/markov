GUTENROOT ?= ~/gutenberg

.PHONY: guten-mine
guten-mine: tools/guten-mine
	go build ./$<

.PHONY: word-demo
word-demo: tools/word-demo
	go build ./$<

PHONY: docs.json
docs.json: guten-mine
	time find $(GUTENROOT) -type f -name '*.txt' | ./guten-mine -stdin >$@ 2>$@.log

clean:
	rm guten-mine word-demo
	rm docs.json docs.json.log
