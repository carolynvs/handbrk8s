package cmd

import (
	"flag"
	"fmt"
	"os"

	"github.com/carolynvs/handbrk8s/internal/watcher"
)

const (
	InvalidArgument int = iota
	Interrupted
	RuntimeError
)

// PlexArgs represents the configuration flags for accessing a Plex library
type PlexArgs struct {
	watcher.LibraryConfig
}

// Populate with Plex command-line arguments.
func (args *PlexArgs) Parse() {
	flag.StringVar(&args.Config.Server, "plex-server", "",
		"Base URL of the Plex server, for example http://192.168.0.105:32400")
	flag.StringVar(&args.Config.Token, "plex-token", os.Getenv("PLEX_TOKEN"), "Plex authentication token [PLEX_TOKEN]")
	flag.StringVar(&args.Name, "plex-library", "", "Name of a Plex library")
	flag.StringVar(&args.Share, "plex-share", "", "Location of to the Plex library share")
	flag.Parse()

	ExitOnMissingFlag(args.Config.Server, "-plex-server")
	ExitOnMissingFlag(args.Config.Token, "-plex-token")
	ExitOnMissingFlag(args.Name, "-plex-library")
	ExitOnMissingFlag(args.Share, "-plex-share")
}

// TODO: Replace with a cmd shell that calls an app which could return an error.
// ExitOnRuntimeError checks for an error, then quits, returning a non-zero exit code.
func ExitOnRuntimeError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(RuntimeError)
	}
}

// ExitOnMissingFlag checks for an empty value, then quits, returning a non-zero exit code.
func ExitOnMissingFlag(value, flag string) {
	if value == "" {
		fmt.Printf("%s is required\n", flag)
		os.Exit(InvalidArgument)
	}
}
