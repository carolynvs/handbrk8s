package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/carolynvs/handbrk8s/cmd"
	"github.com/carolynvs/handbrk8s/internal/fs"
	"github.com/carolynvs/handbrk8s/internal/plex"
	"github.com/dustin/go-humanize"
	"github.com/pkg/errors"
)

// Gracefully handle restarts between upload steps, continuing to the next step
// when the previous is already complete:
// 1. Upload the transcoded video file to the Plex library share
// 2. Refresh the Plex library to include the new video.
// 3. Remove the transcoded video file.
// 4. Remove the original raw video file.
func main() {
	libCfg, transcodedPath, rawPath := parseArgs()

	filename := filepath.Base(transcodedPath)
	uploadPath := filepath.Join(libCfg.Share, filename)

	// Determine if the file should be uploaded
	shouldUpload := false
	destStat, destErr := os.Stat(uploadPath)
	if destErr != nil {
		if os.IsNotExist(destErr) {
			fmt.Println("the video is not in on the Plex share and must be uploaded.")
			shouldUpload = true
		} else {
			err := errors.Wrapf(destErr, "cannot stat %s", uploadPath)
			cmd.ExitOnRuntimeError(err)
		}
	}

	srcStat, srcErr := os.Stat(transcodedPath)
	if srcErr != nil {
		if os.IsNotExist(srcErr) {
			if shouldUpload {
				fmt.Println(errors.Wrapf(srcErr, "cannot stat the transcoded video file '%s'", transcodedPath))
				os.Exit(cmd.RuntimeError)
			}
			fmt.Println("the transcoded video file is gone and was found on the Plex share. Skipping upload.")
		} else {
			err := errors.Wrapf(destErr, "cannot stat %s", uploadPath)
			cmd.ExitOnRuntimeError(err)
		}
	} else if !shouldUpload {
		destSize := uint64(destStat.Size())
		srcSize := uint64(srcStat.Size())
		if destSize != srcSize {
			shouldUpload = true
			fmt.Printf("an existing video file was found on the Plex share, and is a different size than the source video file (%s != %s) and must be re-uploaded.",
				humanize.Bytes(destSize), humanize.Bytes(srcSize))
		}
	}

	shouldRefresh := true
	if shouldUpload {
		shouldRefresh = true
		fmt.Println("uploading the video to Plex...")
		err := fs.CopyFile(transcodedPath, uploadPath)
		cmd.ExitOnRuntimeError(err)
	}

	plexC := plex.NewClient(libCfg.Config)
	lib, err := plexC.FindLibrary(libCfg.Name)
	cmd.ExitOnRuntimeError(err)

	// Determine if the Plex library should be refreshed
	if !shouldRefresh {
		fmt.Println("checking for the video in the Plex library...")
		exists, err := lib.HasVideo(filename)
		cmd.ExitOnRuntimeError(err)
		shouldRefresh = !exists
	}

	if shouldRefresh {
		fmt.Println("updating the Plex library index...")
		err := lib.Update()
		cmd.ExitOnRuntimeError(err)

		fmt.Println("checking that the video in now in the Plex library...")
		exists := false
		for i := 0; i < 3; i++ {
			time.Sleep(1 * time.Second)
			exists, err = lib.HasVideo(filename)
			if err != nil {
				continue
			}
			if exists {
				break
			}
		}
		if !exists {
			err = errors.New("plex was updated but the video is still not in the library")
			cmd.ExitOnRuntimeError(err)
		}
	} else {
		fmt.Println("the video is already in the Plex library. Skipping update.")
	}

	// Determine if the transcoded file should be removed
	_, err = os.Stat(transcodedPath)
	if err != nil {
		if !os.IsNotExist(err) {
			err = errors.Wrapf(err, "cannot stat %s", transcodedPath)
			cmd.ExitOnRuntimeError(err)
		}
	} else {
		fmt.Printf("removing %s\n", transcodedPath)
		err = os.Remove(transcodedPath)
		cmd.ExitOnRuntimeError(err)
	}

	// Determine if the original raw file should be removed
	_, err = os.Stat(rawPath)
	if err != nil {
		if !os.IsNotExist(err) {
			err = errors.Wrapf(err, "cannot stat %s", transcodedPath)
			cmd.ExitOnRuntimeError(err)
		}
	} else {
		fmt.Printf("removing %s\n", rawPath)
		err = os.Remove(rawPath)
		cmd.ExitOnRuntimeError(err)
	}
}

// parseArgs reads and validates flags and environment variables.
func parseArgs() (plexCfg cmd.PlexArgs, transcodedPath, rawPath string) {
	fs := flag.NewFlagSet("uploader", flag.ExitOnError)

	fs.StringVar(&transcodedPath, "f", "", "transcoded video file to upload to Plex")
	fs.StringVar(&rawPath, "raw", "", "original raw video file to cleanup")
	plexCfg.Parse(fs)

	cmd.ExitOnMissingFlag(transcodedPath, "-f")
	cmd.ExitOnMissingFlag(rawPath, "-raw")

	return plexCfg, transcodedPath, rawPath
}
