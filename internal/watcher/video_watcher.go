package watcher

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/carolynvs/handbrk8s/internal/fs"
	"github.com/carolynvs/handbrk8s/internal/watcher/jobs"
	"github.com/pkg/errors"
)

type VideoWatcher struct {
	done chan struct{}

	// WatchDir contains raw (untranscoded) video files.
	WatchDir string

	// ClaimDir temporarily holds raw video files while they are being transcoded.
	ClaimDir string

	// TranscodedDir contains completed (transcoded) video files.
	TranscodedDir string

	// VideoPreset is the name of a HandBrake preset.
	VideoPreset string
}

// NewVideoWatcher begins watching for new videos to transcode.
func NewVideoWatcher(rootDir string, videoPreset string) *VideoWatcher {
	w := &VideoWatcher{
		done:          make(chan struct{}),
		WatchDir:      filepath.Join(rootDir, "raw"),
		ClaimDir:      filepath.Join(rootDir, "raw", "claimed"),
		TranscodedDir: filepath.Join(rootDir, "transcoded"),
		VideoPreset:   videoPreset,
	}
	go w.start()
	return w
}

func (w *VideoWatcher) start() {
	dirWatcher, err := fs.NewStableFileWatcher(w.WatchDir)
	if err != nil {
		log.Fatal(errors.Wrapf(err, "Unable to watch %s", w.WatchDir))
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
	log.Println("attempting to claim ", claimPath)
	err := os.Rename(path, claimPath)
	if err != nil {
		log.Println(errors.Wrapf(err, "unable to move %s to %s, skipping for now", path, claimPath))
		return
	}

	transcodedPath := filepath.Join(w.TranscodedDir, filename)
	tj, err := jobs.CreateTranscodeJob(claimPath, transcodedPath, w.VideoPreset)
	if err != nil {
		log.Println(err)
	}

	_, err = jobs.CreateUploadJob(tj, transcodedPath)
	if err != nil {
		log.Println(err)
	}
}
