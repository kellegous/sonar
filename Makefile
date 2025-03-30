SHA := $(shell git rev-parse HEAD)
BUILD_TIME := $(shell git show -s --format=%ct $(SHA))

GOMOD := $(shell go list -m)
GOBUILD_FLAGS := -ldflags "-X $(GOMOD)/internal/build.vcsInfo=$(SHA),$(BUILD_TIME)"

ASSETS := \
	internal/ui/assets/index.html

.PHONY: ALL test clean nuke

ALL: bin/sonard

bin/sonard: cmd/sonard/main.go $(ASSETS) $(shell find internal -type f)
	go build -o $@ $(GOBUILD_FLAGS) ./cmd/sonard

bin/devserver: cmd/devserver/main.go $(shell find internal -type f)
	go build -o $@ $(GOBUILD_FLAGS) ./cmd/devserver

bin/buildname:
	GOBIN="$(CURDIR)/bin" go install github.com/kellegous/buildname/cmd/buildname@latest

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