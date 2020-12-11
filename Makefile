.SUFFIXES:

GOPATH ?= $(HOME)/go
GOBIN := $(GOPATH)/bin

build: validate watcher jobchain uploader dashboard test

watcher:
	cd ./cmd/watcher; CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build
	cd ./cmd/watcher; docker build -t carolynvs/handbrk8s-watcher .

dashboard:
	cd ./cmd/dashboard; CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build
	cd ./cmd/dashboard; docker build -t carolynvs/handbrk8s-dashboard .

jobchain:
	cd ./cmd/jobchain; CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build
	cd ./cmd/jobchain; docker build -t carolynvs/jobchain .

uploader:
	cd ./cmd/uploader; CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build
	cd ./cmd/uploader; docker build -t carolynvs/handbrk8s-uploader .

test:
	go test ./...

validate:
	go fmt ./...
	go vet ./...

publish:
	docker push carolynvs/handbrk8s-watcher
	docker push carolynvs/handbrk8s-dashboard
	docker push carolynvs/jobchain
	docker push carolynvs/handbrk8s-uploader

init:
	kubectl apply -f manifests/namespace.yaml
	kubectl apply -f manifests/work.volumes.yaml
	kubectl apply -f manifests/plex.volumes.yaml
	kubectl apply -f manifests/rbac.yaml
	kubectl create configmap handbrakecli -n handbrk8s --from-file=cmd/handbrakecli/presets.json
	kubectl create configmap job-templates -n handbrk8s --from-file=manifests/job-templates/
	kubectl apply -f manifests/watcher-secret.yaml
	kubectl apply -f manifests/watcher.yaml
	kubectl apply -f manifests/dashboard.yaml

config:
	# HACK: https://github.com/kubernetes/kubernetes/issues/30558
	kubectl create configmap handbrakecli -n handbrk8s --dry-run -o yaml --from-file=cmd/handbrakecli/presets.json \
	  | kubectl replace -f -
	kubectl create configmap job-templates -n handbrk8s --dry-run -o yaml --from-file=manifests/job-templates/ \
	  | kubectl replace -f -
	kubectl apply -f manifests/plex.secrets.yaml

deploy: config
	# HACK: force the container to be recreated with the latest image
	-kubectl delete -f manifests/watcher.yaml
	-kubectl delete -f manifests/dashboard.yaml
	kubectl apply -f manifests/watcher.yaml
	kubectl apply -f manifests/dashboard.yaml

local-dashboard:
	open http://localhost
	KUBECONFIG=~/.kube/config go run ./cmd/dashboard/main.go

open-dashboard:
	open http://localhost:8080
	kubectl port-forward svc/dashboard 8080:8080

.PHONY: watcher uploader jobchain dashboard test validate deploy publish open-dashboard local-dashboard
