.SUFFIXES:

GOPATH ?= $(HOME)/go
GOBIN := $(GOPATH)/bin
DEP := $(GOBIN)/dep

default: validate watcher jobchain handbrakecli uploader

$(DEP):
	go get -u github.com/golang/dep/cmd/dep

handbrakecli:
	go build ./cmd/handbrakecli
	cd ./cmd/handbrakecli; CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build
	cd ./cmd/handbrakecli; docker build -t carolynvs/handbrakecli .
	docker push carolynvs/handbrakecli

watcher:
	go build ./cmd/watcher
	cd ./cmd/watcher; CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build
	cd ./cmd/watcher; docker build -t carolynvs/handbrk8s-watcher .
	docker push carolynvs/handbrk8s-watcher

jobchain:
	go build ./cmd/jobchain
	cd ./cmd/jobchain; CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build
	cd ./cmd/jobchain; docker build -t carolynvs/jobchain .
	docker push carolynvs/jobchain

uploader:
	go build ./cmd/uploader
	cd ./cmd/uploader; CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build
	cd ./cmd/uploader; docker build -t carolynvs/handbrk8s-uploader .
	docker push carolynvs/handbrk8s-uploader

test:
	go test ./...

validate: $(DEP) test
	go fmt ./...
	go vet ./...
	dep status | grep -v "mismatch"

deploy:
	kubectl apply -f manifests/namespace.yaml
	kubectl apply -f manifests/work.volumes.yaml
	kubectl apply -f manifests/plex.volumes.yaml
	kubectl apply -f manifests/rbac.yaml
	# HACK: create/delete to force a new container
	-kubectl delete -f manifests/watcher.yaml
	kubectl create -f manifests/watcher.yaml

tail:
	kubectl logs -f deploy/watcher

.PHONY: handbrakecli watcher uploader jobchain test validate deploy tail
