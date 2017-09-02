package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/carolynvs/handbrk8s/internal/watcher"
)

var watchVolume = "/watch"
var videoPreset = "tivo"

func main() {
	w := watcher.NewVideoWatcher(watchVolume, videoPreset)
	defer w.Close()

	// Only stop watching when our process is killed
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	for range signals {
		// Do any cleanup before being shut down
		log.Println("done watching for videos!")
		return
	}
}
