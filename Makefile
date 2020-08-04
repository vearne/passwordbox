VERSION = v0.0.10

CONTAINER=pwbox
BUILD_TIME = $(shell date +%Y%m%d%H%M%S)
LDFLAGS = -ldflags "-w -s -X main.Version=$(VERSION)-$(BUILD_TIME)"
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

