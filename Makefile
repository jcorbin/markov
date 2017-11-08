GUTENROOT ?= ~/gutenberg
DOC_LISTS = $(shell ls *.list)
DOC_DBS = $(DOC_LISTS:.list=.db)

BINS=bin/guten-mine bin/word-demo bin/gen-book bin/gen-doc-list

bins: $(BINS)

bin/%: tools/%
	[ -d bin ] || mkdir bin
	go build -o $@ ./$<

.PHONY: tiny.list
tiny.list: bin/gen-doc-list
	./bin/gen-doc-list -atLeast 400 all.db/index.json \
		| jq -r .sourceFile >$@

all.list:
	find $(GUTENROOT) -type f -name '*.txt' >$@

%.db: %.list bin/guten-mine
	! [ -e $@ ] || rm $@ -rf
	mkdir $@
	time ./bin/guten-mine -dbDir $@ -stdin <$< >$@/index.json 2>$@/extract.log

clean:
	rm -rf $(DOC_DBS)
	rm all.list tiny.list
	rm -f $(BINS)
