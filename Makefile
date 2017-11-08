GUTENROOT ?= ~/gutenberg

BINS=bin/guten-mine bin/word-demo

bins: $(BINS)

bin/%: tools/%
	[ -d bin ] || mkdir bin
	go build -o $@ ./$<

all.list:
	find $(GUTENROOT) -type f -name '*.txt' >$@

PHONY: docs.json
docs.json: all.list guten-mine
	time ./guten-mine -stdin <$< >$@ 2>$@.log

clean:
	rm docs.json docs.json.log
	rm all.list
	rm -f $(BINS)
