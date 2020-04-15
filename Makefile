build:
	go build -ldflags '-w -s' -o pwbox

install: build
	cp -f pwbox /usr/local/bin/


