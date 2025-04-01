SHA := $(shell go run github.com/kellegous/glue/build/info --format="{{.SHA}}")
BUILD_TIME := $(shell go run github.com/kellegous/glue/build/info --format="{{.CommitTime|timestamp}}")
BUILD_NAME := $(shell go run github.com/kellegous/glue/build/info --format="{{.Name}}")

ASSETS := \
	internal/ui/assets/index.html

.PHONY: ALL test clean nuke

ALL: bin/sonard

bin/sonard: cmd/sonard/main.go $(ASSETS) $(shell find internal -type f)
	go build -o $@ ./cmd/sonard

bin/devserver: cmd/devserver/main.go $(shell find internal -type f)
	go build -o $@ ./cmd/devserver

internal/ui/assets/index.html: node_modules/.build $(shell find ui -type f)
	SHA="$(SHA)" BUILD_NAME="$(BUILD_NAME)" npm run build

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