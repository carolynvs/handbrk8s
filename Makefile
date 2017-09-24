.SUFFIXES:

GOPATH ?= $(HOME)/go
GOBIN := $(GOPATH)/bin
DEP := $(GOBIN)/dep

default: validate watcher jobchain uploader

$(DEP):
	go get -u github.com/golang/dep/cmd/dep

watcher:
	cd ./cmd/watcher; CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build
	cd ./cmd/watcher; docker build -t carolynvs/handbrk8s-watcher .

jobchain:
	cd ./cmd/jobchain; CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build
	cd ./cmd/jobchain; docker build -t carolynvs/jobchain .

uploader:
	cd ./cmd/uploader; CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build
	cd ./cmd/uploader; docker build -t carolynvs/handbrk8s-uploader .

test:
	go test ./...

validate: $(DEP) test
	go fmt ./...
	go vet ./...
	dep status | grep -v "mismatch"

publish:
	docker push carolynvs/handbrk8s-watcher
	docker push carolynvs/jobchain
	docker push carolynvs/handbrk8s-uploader

deploy:
	kubectl apply -f manifests/namespace.yaml
	kubectl apply -f manifests/work.volumes.yaml
	kubectl apply -f manifests/plex.volumes.yaml
	kubectl apply -f manifests/rbac.yaml
	# HACK: https://github.com/kubernetes/kubernetes/issues/30558
	kubectl create configmap handbrakecli --dry-run -o yaml --from-file=cmd/handbrakecli/presets.json \
	  | kubectl replace -f -
	kubectl create configmap job-templates --dry-run -o yaml --from-file=manifests/job-templates/ \
   	  | kubectl replace -f -
   	# HACK: force the container to be recreated with the latest image
	-kubectl delete -f manifests/watcher.yaml
	kubectl create -f manifests/watcher.yaml

tail:
	kubectl logs -f deploy/watcher

.PHONY: watcher uploader jobchain test validate deploy publish tail
