default: build test

build: validate
	go build ./cmd/watcher

test:
	go test ./...

validate:
	go fmt ./...
	go vet ./...

PHONY: build test validate