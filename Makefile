SHA := $(shell git rev-parse HEAD)
TAG := $(shell git rev-parse --short HEAD)

ASSETS := \
	pkg/web/ui/index.html

ALL: bin/sonard

bin/%: cmd/%/main.go $(ASSETS) $(shell find pkg -type f)
	go build -o $@ ./cmd/$*

bin/buildname:
	GOBIN="$(CURDIR)/bin" go install github.com/kellegous/buildname/cmd/buildname@latest

bin/buildimg:
	GOBIN="$(CURDIR)/bin" go install github.com/kellegous/buildimg@latest

pkg/web/ui/index.html: node_modules/.build bin/buildname $(shell find ui -type f)
	SHA="$(SHA)" BUILD_NAME="$(shell bin/buildname $(SHA))" npm run build

node_modules/.build:
	npm install
	touch $@

test:
	go test ./pkg/...

clean:
	rm -rf bin pkg/web/ui/

nuke: clean
	rm -rf node_modules

push-docker: bin/buildimg Dockerfile $(shell find cmd pkg ui -type f)
	bin/buildimg --tag=$(TAG) --target=linux/amd64 --target=linux/arm64 kellegous/sonar

sonar-$(TAG).tar: bin/buildimg Dockerfile $(shell find cmd pkg ui -type f)
	bin/buildimg --tag=$(TAG) --target=linux/amd64:$@ kellegous/sonar

publish: sonar-$(TAG).tar
	sup host image load @ $<