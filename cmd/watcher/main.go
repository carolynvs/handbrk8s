package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/carolynvs/handbrk8s/cmd"
	"github.com/carolynvs/handbrk8s/internal/watcher"
)

var watchVolume = "/watch"
var videoPreset = "tivo"

func main() {
	plexCfg := parseArgs()

	w := watcher.NewVideoWatcher(watchVolume, videoPreset, plexCfg)
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

// parseArgs reads and validates flags and environment variables.
func parseArgs() (watcher.LibraryConfig) {
	var args cmd.PlexArgs
	args.Parse()
	return args.LibraryConfig
}
