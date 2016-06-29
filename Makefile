GOFILES := $(shell find . -name '*.go' -not -path './vendor/*')
HTMLFILES := $(shell find . -name '*.html' -not -path './vendor/*')

suggest: ./suggest/suggest

html: ./web.go

serve: suggest
	./suggest/suggest serve

./suggest/suggest: $(GOFILES)
	go build -o $@ ./suggest

./web.go: $(HTMLFILES)
	(echo "package suggest\nvar web = map[string][]byte{" && \
		for i in $^; do \
			echo "\"$$i\": []byte{\n$$(cat $$i | gzip -n | xxd -i),\n},"; done && echo "\n}") | \
				gofmt > $@
