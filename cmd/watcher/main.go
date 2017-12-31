package main

import (
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/carolynvs/handbrk8s/cmd"
	"github.com/carolynvs/handbrk8s/internal/watcher"
)

var configVolume = "/config"
var watchVolume = "/watch"
var workVolume = "/work"
var videoPreset = "tivo"

func main() {
	plexCfg := parseArgs()

	w, err := watcher.NewVideoWatcher(configVolume, watchVolume, workVolume, videoPreset, plexCfg)
	if err != nil {
		cmd.ExitOnRuntimeError(err)
	}
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
func parseArgs() watcher.LibraryConfig {
	fs := flag.NewFlagSet("watcher", flag.ExitOnError)
	var args cmd.PlexArgs
	args.Parse(fs)
	return args.LibraryConfig
}
