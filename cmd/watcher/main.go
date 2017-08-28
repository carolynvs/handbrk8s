package main

import (
	"log"
	"os"
	"os/signal"

	"path/filepath"
	"strings"

	"github.com/carolynvs/handbrk8s/internal/k8s"
	"github.com/carolynvs/handbrk8s/internal/watchers"
	"github.com/pkg/errors"
)

var watchDir = "/watch/raw/"
var claimDir = "/watch/raw/claimed/"
var transcodedDir = "/watch/transcoded/"

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
	// Ignore hidden files
	filename := filepath.Base(path)
	if strings.HasPrefix(".", filename) {
		return
	}

	// Claim the file, prevents attempts to process it a second time
	claimPath := filepath.Join(claimDir, filename)
	log.Println("attempting to claim ", claimPath)
	err := os.Rename(path, claimPath)
	if err != nil {
		log.Println(errors.Wrapf(err, "unable to move %s to %s, skipping for now", path, claimPath))
		return
	}

	transcodedPath := filepath.Join(transcodedDir, filename)
	err = k8s.CreateTranscodeJob(claimPath, transcodedPath)
	if err != nil {
		log.Println(err)
	}

	err = k8s.CreateUploaderJob(transcodedPath)
	if err != nil {
		log.Println(err)
	}
}
