sources := $(wildcard *.go)

build: fswatch

fswatch: $(sources)
	go build -o $@ ./...
