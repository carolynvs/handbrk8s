package main

import (
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/carolynvs/handbrk8s/cmd"
	"github.com/carolynvs/handbrk8s/internal/plex"
	"github.com/carolynvs/handbrk8s/internal/watcher"
)

var configVolume = "/config"
var watchVolume = "/watch"
var workVolume = "/work"
var plexVolume = "/plex"
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
func parseArgs() (plexCfg plex.LibraryConfig) {
	fs := flag.NewFlagSet("watcher", flag.ExitOnError)

	fs.StringVar(&plexCfg.URL, "plex-server", "",
		"Base URL of the Plex server, for example http://192.168.0.105:32400")
	fs.StringVar(&plexCfg.Token, "plex-token", os.Getenv("PLEX_TOKEN"), "Plex authentication token [PLEX_TOKEN]")
	fs.StringVar(&plexCfg.Share, "plex-share", "", "Location of the Plex share")
	fs.Parse(os.Args[1:])

	cmd.ExitOnMissingFlag(plexCfg.URL, "-plex-server")
	cmd.ExitOnMissingFlag(plexCfg.Token, "-plex-token")

	return plexCfg
}
