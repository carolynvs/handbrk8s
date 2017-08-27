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
	docker push carolynvs/handbrk8s-watcher

deploy:
	kubectl apply -f manifests/handbrk8s.namespace.yaml
	kubectl apply -f manifests/handbrk8s.rbac.yaml
	# HACK: create/delete to force a new container
	-kubectl delete -f manifests/handbrk8s.deploy.yaml
	kubectl create -f manifests/handbrk8s.deploy.yaml

tail:
	kubectl logs -f deploy/watcher

watch: pkg
	docker run --rm -it -v `pwd`/tmp:/tmp carolynvs/handbrk8s-watcher

PHONY: build test validate