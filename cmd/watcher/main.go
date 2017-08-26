package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/carolynvs/handbrk8s/internal/watcher"
	"github.com/pkg/errors"
)

var watchDir = "/mlp/movies/raw"

func main() {
	done := make(chan bool)

	// watch directory
	watcher, err := watcher.NewCopyFileWatcher(watchDir)
	if err != nil {
		log.Fatal(errors.Wrapf(err, "Unable to watch %s", watchDir))
	}
	defer watcher.Close()

	// Wait for the kill
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	go waitForTheKill(signals, done)
	<-done
}

func waitForTheKill(signals <-chan os.Signal, done chan<- bool) {
	for range signals {
		log.Println("okay, everyone out of the pool!")
		done <- true
	}
}
