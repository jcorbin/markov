GUTENROOT ?= ~/gutenberg

%: tools/%
	go build ./$<

PHONY: docs.json
docs.json: guten-mine
	time find $(GUTENROOT) -type f -name '*.txt' | ./guten-mine -stdin >$@ 2>$@.log

clean:
	rm guten-mine word-demo
	rm docs.json docs.json.log
