BIN_DIR := $(GOPATH)/bin
DEP := $(BINDIR)/dep

default: build test

$(DEP):
	go get -u github.com/golang/dep/cmd/dep

build: validate
	go build ./cmd/watcher

test:
	go test ./...

validate: $(DEP)
	go fmt ./...
	go vet ./...
	dep status | grep -v "mismatch"

watch: build
	./watcher

PHONY: build test validate