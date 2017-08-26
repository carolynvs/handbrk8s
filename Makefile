default: build test

build: validate
	go build ./cmd/watcher

test:
	go test ./...

validate:
	go fmt ./...
	go vet ./...

watch: build
	./watcher

PHONY: build test validate