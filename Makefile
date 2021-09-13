VERSION = v0.0.12

CONTAINER=pwbox
IMPORT_PATH = github.com/vearne/passwordbox

BUILD_TIME = $(shell date +%Y%m%d%H%M%S)
GITTAG = `git log -1 --pretty=format:"%H"`
LDFLAGS = -ldflags "-s -w -X $(IMPORT_PATH)/consts.GitTag=${GITTAG} -X $(IMPORT_PATH)/consts.BuildTime=${BUILD_TIME} -X $(IMPORT_PATH)/consts.Version=${VERSION}"
SOURCE_PATH = /go/src/github.com/vearne/passwordbox/

.PHONY: build install release release-linux release-mac docker-img


build:
	go build $(LDFLAGS) -o pwbox

install: build
	cp -f pwbox /usr/local/bin/


release: release-linux release-mac

release-linux:
	docker run -v `pwd`:$(SOURCE_PATH) -t -e GOOS=linux -e GOARCH=amd64 -i $(CONTAINER) go build $(LDFLAGS) -o pwbox
	tar -zcvf pwbox-$(VERSION)-linux-amd64.tar.gz ./pwbox
	rm pwbox

release-mac:
	env GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o pwbox
	tar -zcvf pwbox-$(VERSION)-darwin-amd64.tar.gz ./pwbox
	rm pwbox

docker-img:
	docker build --rm -t $(CONTAINER) -f Dockerfile.dev .

