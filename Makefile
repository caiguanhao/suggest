GOFILES = $(shell find . -name '*.go' -not -path './vendor/*')
HTMLFILES = $(shell find web -name '*.html')
HTMLGOFILES = $(patsubst %.html,%.html.go,$(HTMLFILES))
WEBGO = ./web/web.go

suggest: ./suggest/suggest

html: $(WEBGO)

serve: suggest
	./suggest/suggest serve --local

./suggest/suggest: $(GOFILES) $(WEBGO)
	go build -race -o $@ ./suggest

$(WEBGO): $(HTMLGOFILES)
	(echo "package web" && \
		echo "var Files = map[string][]byte{" && \
			for i in $^; do i=$${i%%.go} && \
				echo "\"$$i\": []byte($${i//[^A-Za-z0-9]/_}),"; \
					done && \
						echo "}") | gofmt > $@

$(HTMLGOFILES): $(HTMLFILES)
	for i in $^; do \
		(echo "package web" && \
			echo "const $${i//[^A-Za-z0-9]/_} = \`$$(cat $$i)\n\`") | gofmt > $$i.go; \
				done
