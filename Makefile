GOBIN := $(GOPATH)/bin
DEP := $(GOBIN)/dep

default: validate watcher jobchain

$(DEP):
	go get -u github.com/golang/dep/cmd/dep

watcher:
	go build ./cmd/watcher
	go test ./cmd/watcher
	cd ./cmd/watcher; CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build
	cd ./cmd/watcher; docker build -t carolynvs/handbrk8s-watcher .
	docker push carolynvs/handbrk8s-watcher

jobchain:
	go build ./cmd/jobchain
	go test ./cmd/jobchain
	cd ./cmd/jobchain; CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build
	cd ./cmd/jobchain; docker build -t carolynvs/jobchain .
	docker push carolynvs/jobchain

test:
	go test ./...

validate: $(DEP) test
	go fmt ./...
	go vet ./...
	dep status | grep -v "mismatch"

deploy:
	kubectl apply -f manifests/handbrk8s.namespace.yaml
	kubectl apply -f manifests/handbrk8s.rbac.yaml
	# HACK: create/delete to force a new container
	-kubectl delete -f manifests/handbrk8s.deploy.yaml
	kubectl create -f manifests/handbrk8s.deploy.yaml

tail:
	kubectl logs -f deploy/watcher

.PHONY: watcher jobchain test validate deploy tail
