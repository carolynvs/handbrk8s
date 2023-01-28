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


# Troubleshooting

## Transcode fails writing file to disk

When the transcode job fails because either the init container (prep) can't make the work directory, or the transcode job fails with the following error because it can't write to the work directory (below), the problem is that the directories on /ponyshare should be owned by carolynvs.ponies. Check that it's not owned by root which can happen when making directories pre-emptively and you aren't logged in as the right user.

```
# prep container
mkdir: can't create directory '/ponyshare/handbrk8s/': Permission denied
```

```
# transcode container
[15:06:21] sync: got 0 frames, 74637 expected
```
