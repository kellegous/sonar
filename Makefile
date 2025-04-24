PROTOC_GEN_GO_VERSION := v1.36.5
PROTOC_GEN_TWIRP_VERSION := v8.1.3
PROTOC_VERSION := 29.1

SHA := $(shell go run github.com/kellegous/glue/build/info --format="{{.SHA}}")
BUILD_TIME := $(shell go run github.com/kellegous/glue/build/info --format="{{.CommitTime|timestamp}}")
BUILD_NAME := $(shell go run github.com/kellegous/glue/build/info --format="{{.Name}}")

GO_MOD := $(shell go list -m)

ASSETS := \
	internal/ui/assets/index.html

BE_PROTOS := \
	sonar.pb.go \
	sonar.twirp.go

FE_PROTOS := \
	ui/src/gen/sonar.ts \
	ui/src/gen/sonar.twirp.ts

.PHONY: ALL test clean nuke

ALL: bin/sonard

bin/sonard: cmd/sonard/main.go $(BE_PROTOS) $(ASSETS) $(shell find internal -type f)
	go build -o $@ ./cmd/sonard

bin/devserver: cmd/devserver/main.go $(shell find internal -type f)
	go build -o $@ ./cmd/devserver

bin/protoc-gen-go:
	GOBIN="$(CURDIR)/bin" go install google.golang.org/protobuf/cmd/protoc-gen-go@$(PROTOC_GEN_GO_VERSION)

bin/protoc-gen-twirp:
	GOBIN="$(CURDIR)/bin" go install github.com/twitchtv/twirp/protoc-gen-twirp@$(PROTOC_GEN_TWIRP_VERSION)

bin/protoc:
	etc/download-protoc $(PROTOC_VERSION)

%.pb.go: %.proto bin/protoc-gen-go bin/protoc
	bin/protoc --proto_path=. \
		--plugin=protoc-gen-go=bin/protoc-gen-go \
		--go_out=. \
		--go_opt=module=$(GO_MOD) \
		$<

%.twirp.go: %.proto bin/protoc-gen-twirp bin/protoc
	bin/protoc --proto_path=. \
		--plugin=protoc-gen-twirp=bin/protoc-gen-twirp \
		--twirp_out=. \
		--twirp_opt=module=$(GO_MOD) \
		$<

ui/src/gen/%.ts: %.proto node_modules/.build
	mkdir -p $(dir $@)
	protoc --proto_path=. \
		--plugin=protoc-gen-ts=node_modules/.bin/protoc-gen-ts \
		--ts_out=ui/src/gen \
		--ts_opt=ts_nocheck,force_server_none \
		$<

ui/src/gen/%.twirp.ts: %.proto node_modules/.build
	mkdir -p $(dir $@)
	protoc --proto_path=. \
		--plugin=protoc-gen-ts=node_modules/.bin/protoc-gen-ts \
		--plugin=protoc-gen-twirp_ts=node_modules/.bin/protoc-gen-twirp_ts \
		--twirp_ts_out=ui/src/gen \
		--ts_out=ui/src/gen \
		--ts_opt=ts_nocheck,force_server_none \
		$<
	
internal/ui/assets/index.html: node_modules/.build $(FE_PROTOS) $(shell find ui -type f)
	SHA="$(SHA)" BUILD_NAME="$(BUILD_NAME)" npm run build

node_modules/.build:
	npm install
	touch $@

develop: bin/sonard bin/devserver
	bin/devserver

test:
	go test ./internal/...

clean:
	rm -rf bin internal/ui/assets $(BE_PROTOS)

nuke: clean
	rm -rf node_modules