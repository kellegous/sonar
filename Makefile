SHA := $(shell git rev-parse HEAD)

ifndef SHA
	SHA := $(shell git rev-parse HEAD)
endif

ifndef BUILD_TIME
	BUILD_TIME := $(shell git show -s --format=%ct $(SHA))
endif

GOMOD := $(shell go list -m)
GOBUILD_FLAGS := -ldflags "-X $(GOMOD)/internal/build.vcsInfo=$(SHA),$(BUILD_TIME)"

ASSETS := \
	internal/ui/assets/index.html

.PHONY: ALL test clean nuke publish push-docker sonar.tar

ALL: bin/sonard

bin/sonard: cmd/sonard/main.go $(ASSETS) $(shell find internal -type f)
	go build -o $@ $(GOBUILD_FLAGS) ./cmd/sonard

bin/devserver: cmd/devserver/main.go $(shell find internal -type f)
	go build -o $@ $(GOBUILD_FLAGS) ./cmd/devserver

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

sonar.tar: bin/buildimg
	bin/buildimg --tag=$(TAG) --target=linux/amd64:$@  --build-arg=SHA=${SHA} --build-arg=BUILD_TIME=${BUILD_TIME} kellegous/sonar

publish: sonar.tar
	sup host image load @ $<