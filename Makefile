GOBIN := $(GOPATH)/bin
DEP := $(GOBIN)/dep

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

pkg:
	cd ./cmd/watcher; CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build
	cd ./cmd/watcher; docker build -t carolynvs/handbrk8s-watcher .

watch: build
	./watcher

PHONY: build test validate