package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/carolynvs/handbrk8s/internal/k8s"
	"github.com/carolynvs/handbrk8s/internal/watchers"
	"github.com/pkg/errors"
)

var watchDir = "/tmp"

func main() {
	done := make(chan struct{})

	// watch directory
	watcher, err := watchers.NewStableFile(watchDir)
	if err != nil {
		log.Fatal(errors.Wrapf(err, "Unable to watch %s", watchDir))
	}
	defer watcher.Close()

	// Wait for the kill
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	go waitForTheKill(signals, done)

	for {
		select {
		case <-done:
			return
		case file := <-watcher.Events:
			go handleFile(file.Path)
		}
	}
}

func waitForTheKill(signals <-chan os.Signal, done chan struct{}) {
	for range signals {
		log.Println("okay, everyone out of the pool!")
		close(done)
	}
}

func handleFile(path string) {
	log.Println("handling ", path)
	err := k8s.CreateJob(path)
	if err != nil {
		log.Println(err)
	}
}
