# Handbrake on Kubernetes
Run handbrake in parallel on a Kubernetes cluster.

[![Build Status](https://travis-ci.org/carolynvs/handbrk8s.svg?branch=master)](https://travis-ci.org/carolynvs/handbrk8s)

# Local Development
Works best with go version 1.9.

1. `go get -u github.com/carolynvs/handbrk8s`
1. `make`
1. `make deploy`
1. `make tail`

# Fun Commands

* `kubectl get pods -o wide` will show you where your pods are running.
  Take a moment and admire having a bunch of computers doing your bidding.