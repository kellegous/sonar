SHA := $(shell git rev-parse HEAD)
TAG := $(shell git rev-parse --short HEAD)

ASSETS := pkg/web/ui/index.html \
	pkg/web/ui/index.js

ALL: bin/sonard

bin/%: cmd/%/main.go $(ASSETS) $(shell find pkg ui -type f)
	go build -o $@ ./cmd/$*

bin/render_html:
	GOBIN="$(CURDIR)/bin" go install github.com/kellegous/render_html@latest

bin/buildname:
	GOBIN="$(CURDIR)/bin" go install github.com/kellegous/buildname/cmd/buildname@latest

bin/buildimg:
	GOBIN="$(CURDIR)/bin" go install github.com/kellegous/buildimg@latest

pkg/web/ui/index.html: ui/index.html bin/render_html bin/buildname
	bin/render_html -v build.sha="$(SHA)" -v build.name="$(shell bin/buildname $(SHA))" $< $@

node_modules/build:
	npm install --verbose
	date > $@

pkg/web/ui/index.js: $(shell find ui -type f) node_modules/build
	npx webpack --mode=production

test:
	go test ./pkg/...

clean:
	rm -rf bin $(ASSETS)

nuke: clean
	rm -rf node_modules

push-docker: bin/buildimg Dockerfile $(shell find cmd pkg ui -type f)
	bin/buildimg --tag=$(TAG) --target=linux/amd64 --target=linux/arm64 kellegous/sonar

sonar-$(TAG).tar: bin/buildimg Dockerfile $(shell find cmd pkg ui -type f)
	bin/buildimg --tag=$(TAG) --target=linux/amd64:$@ kellegous/sonar

publish: sonar-$(TAG).tar
	sup host image load @ $<