package watcher

import (
	"log"
	"path/filepath"
	"strings"

	"github.com/carolynvs/handbrk8s/internal/fs"
	"github.com/carolynvs/handbrk8s/internal/k8s/jobs"
	"github.com/carolynvs/handbrk8s/internal/plex"
	"github.com/pkg/errors"
)

const namespace = "handbrk8s"

type VideoWatcher struct {
	done chan struct{}

	// WatchDir contains raw (untranscoded) video files.
	WatchDir string

	// ClaimDir temporarily holds raw video files while they are being transcoded.
	ClaimDir string

	// TranscodedDir contains completed (transcoded) video files.
	TranscodedDir string

	FailedDir string

	// VideoPreset is the name of a HandBrake preset.
	VideoPreset string

	// DestLib contains connection information to the destination Plex library.
	DestLib LibraryConfig
}

// LibraryConfig is the set of information necessary to upload videos to a Plex library.
type LibraryConfig struct {
	Config plex.Config
	Name   string
	Share  string
}

// NewVideoWatcher begins watching for new videos to transcode.
func NewVideoWatcher(watchDir, failedDir, workVolume string, videoPreset string, destLib LibraryConfig) *VideoWatcher {
	w := &VideoWatcher{
		done:          make(chan struct{}),
		WatchDir:      watchDir,
		FailedDir:     failedDir,
		ClaimDir:      filepath.Join(workVolume, "claimed"),
		TranscodedDir: filepath.Join(workVolume, "transcoded"),
		VideoPreset:   videoPreset,
		DestLib:       destLib,
	}

	log.Printf("watching %s for new videos\n", w.WatchDir)
	go w.start()
	return w
}

func (w *VideoWatcher) start() {
	dirWatcher, err := fs.NewStableFileWatcher(w.WatchDir)
	if err != nil {
		log.Fatal(errors.Wrapf(err, "unable to watch %s", w.WatchDir))
	}
	defer dirWatcher.Close()

	for {
		select {
		case <-w.done:
			return
		case file := <-dirWatcher.Events:
			go w.handleVideo(file.Path)
		}
	}
}

func (w *VideoWatcher) Close() {
	close(w.done)
}

func (w *VideoWatcher) handleVideo(path string) {
	// Ignore hidden files
	filename := filepath.Base(path)
	if strings.HasPrefix(".", filename) {
		return
	}

	// Claim the file, prevents attempts to process it a second time
	claimPath := filepath.Join(w.ClaimDir, filename)
	log.Printf("attempting to claim %s\n", claimPath)
	err := fs.MoveFile(path, claimPath)
	if err != nil {
		log.Println(errors.Wrapf(err, "unable to move %s to %s, skipping for now",
			path, claimPath))
		return
	}

	transcodedPath := filepath.Join(w.TranscodedDir, filename)
	tj, err := w.createTranscodeJob(claimPath, transcodedPath)
	if err != nil {
		log.Println(err)
		w.cleanupFailedClaim(claimPath)
		return
	}

	_, err = w.createUploadJob(tj, transcodedPath, claimPath)
	if err != nil {
		log.Println(err)
		err = jobs.Delete(tj, namespace)
		if err != nil {
			log.Println(err)
		}
		w.cleanupFailedClaim(claimPath)
		return
	}
}

func (w *VideoWatcher) cleanupFailedClaim(claimPath string) {
	log.Printf("cleaning up failed claim: %s\n", claimPath)
	failedPath := filepath.Join(w.FailedDir, filepath.Base(claimPath))
	err := fs.MoveFile(claimPath, failedPath)
	if err != nil {
		log.Println(errors.Wrap(err, "unable to cleanup failed claim"))
	}
}
