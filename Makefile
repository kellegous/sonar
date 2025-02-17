SHA := $(shell git rev-parse HEAD)
TAG := $(shell git rev-parse --short HEAD)

ASSETS := \
	internal/ui/assets/index.html

.PHONY: ALL test clean nuke publish push-docker sonar-$(TAG).tar

ALL: bin/sonard

bin/%: cmd/%/main.go $(ASSETS) $(shell find internal -type f)
	go build -o $@ ./cmd/$*

bin/buildname:
	GOBIN="$(CURDIR)/bin" go install github.com/kellegous/buildname/cmd/buildname@latest

bin/buildimg:
	GOBIN="$(CURDIR)/bin" go install github.com/kellegous/buildimg@latest

internal/ui/assets/index.html: node_modules/.build bin/buildname $(shell find ui -type f)
	SHA="$(SHA)" BUILD_NAME="$(shell bin/buildname $(SHA))" npm run build

node_modules/.build:
	npm install
	touch $@

develop: bin/sonard bin/devserver
	bin/devserver

test:
	go test ./internal/...

clean:
	rm -rf bin internal/ui/assets

nuke: clean
	rm -rf node_modules

push-docker: bin/buildimg
	bin/buildimg --tag=$(TAG) --target=linux/amd64 --target=linux/arm64 kellegous/sonar

sonar-$(TAG).tar: bin/buildimg
	bin/buildimg --tag=$(TAG) --target=linux/amd64:$@ kellegous/sonar

publish: sonar-$(TAG).tar
	sup host image load @ $<