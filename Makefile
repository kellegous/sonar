PROTOC_GEN_GO_VERSION := v1.36.5
PROTOC_GEN_CONNECT_GO_VERSION := v1.19.1
PROTOC_VERSION := 33.0

SHA = $(shell go run github.com/kellegous/glue/build/info --format="{{.SHA}}")
BUILD_NAME = $(shell go run github.com/kellegous/glue/build/info --format="{{.Name}}")

GO_MOD := $(shell go list -m)

ASSETS := \
	internal/ui/assets/index.html

BE_PROTOS := \
	sonar.pb.go \
	sonar_connect/sonar.connect.go

FE_PROTOS := \
	ui/src/gen/sonar_pb.ts

.PHONY: ALL test clean nuke

ALL: bin/sonard

bin/sonard: cmd/sonard/main.go $(BE_PROTOS) $(ASSETS) $(shell find internal -type f -name '*.go')
	go build -o $@ ./cmd/sonard

bin/protoc-gen-go:
	GOBIN="$(CURDIR)/bin" go install google.golang.org/protobuf/cmd/protoc-gen-go@$(PROTOC_GEN_GO_VERSION)

bin/protoc-gen-connect-go:
	GOBIN="$(CURDIR)/bin" go install connectrpc.com/connect/cmd/protoc-gen-connect-go@$(PROTOC_GEN_CONNECT_GO_VERSION)

bin/protoc:
	etc/download-protoc $(PROTOC_VERSION)

%.pb.go: %.proto bin/protoc-gen-go bin/protoc
	bin/protoc --proto_path=. \
		--plugin=protoc-gen-go=bin/protoc-gen-go \
		--go_out=. \
		--go_opt=module=$(GO_MOD) \
		$<

sonar_connect/sonar.connect.go: sonar.proto bin/protoc-gen-connect-go bin/protoc
	bin/protoc --proto_path=. \
		--plugin=protoc-gen-connect-go=bin/protoc-gen-connect-go \
		--connect-go_out=. \
		--connect-go_opt=module=$(GO_MOD) \
		--connect-go_opt=package_suffix=_connect \
		$<

ui/src/gen/%_pb.ts: %.proto node_modules/.build bin/protoc
	mkdir -p $(dir $@)
	bin/protoc --proto_path=. \
		--plugin=protoc-gen-es=node_modules/.bin/protoc-gen-es \
		--es_out=ui/src/gen \
		--es_opt=target=ts \
		$<

internal/ui/assets/index.html: node_modules/.build $(FE_PROTOS) $(shell find ui -type f)
	SHA="$(SHA)" BUILD_NAME="$(BUILD_NAME)" npm run build

node_modules/.build:
	npm install
	touch $@

develop: bin/sonard
	sudo bin/sonard --dev-mode=.:4066

test:
	go test ./internal/...

clean:
	rm -rf bin internal/ui/assets $(BE_PROTOS)

nuke: clean
	rm -rf node_modules