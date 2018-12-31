package watcher

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/carolynvs/handbrk8s/internal/fs"
	"github.com/carolynvs/handbrk8s/internal/k8s/jobs"
	"github.com/carolynvs/handbrk8s/internal/plex"
	"github.com/pkg/errors"
)

const Namespace = "handbrk8s"

type VideoWatcher struct {
	done chan struct{}

	// WatchDir contains raw (untranscoded) video files.
	WatchDir string

	// ClaimDir temporarily holds raw video files while they are being transcoded.
	ClaimDir string

	// TranscodedDir contains completed (transcoded) video files.
	TranscodedDir string

	// TemplatesDir contains templates for jobs that are created by the watcher.
	TemplatesDir string

	FailedDir string

	// VideoPreset is the name of a HandBrake preset.
	VideoPreset string

	// PlexCfg contains connection information upload a file to a Plex server.
	PlexCfg plex.LibraryConfig
}

// NewVideoWatcher begins watching for new videos to transcode.
func NewVideoWatcher(configVolume, watchVolume, workVolume string, videoPreset string, plexCfg plex.LibraryConfig) (*VideoWatcher, error) {
	if _, err := os.Stat(configVolume); os.IsNotExist(err) {
		return nil, errors.Errorf("config volume, %s, is not mounted", configVolume)
	}

	if _, err := os.Stat(watchVolume); os.IsNotExist(err) {
		return nil, errors.Errorf("watch volume, %s, is not mounted", watchVolume)
	}

	if _, err := os.Stat(workVolume); os.IsNotExist(err) {
		return nil, errors.Errorf("work volume, %s, is not mounted", workVolume)
	}

	w := &VideoWatcher{
		done:          make(chan struct{}),
		WatchDir:      filepath.Join(watchVolume, "watch"),
		FailedDir:     filepath.Join(watchVolume, "fail"),
		ClaimDir:      filepath.Join(workVolume, "claim"),
		TranscodedDir: filepath.Join(workVolume, "work"),
		TemplatesDir:  filepath.Join(configVolume, "templates"),
		VideoPreset:   videoPreset,
		PlexCfg:       plexCfg,
	}

	err := os.MkdirAll(w.WatchDir, 0755)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to create watch directory %s", w.WatchDir)
	}

	err = os.MkdirAll(w.FailedDir, 0755)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to create failed directory %s", w.FailedDir)
	}

	err = os.MkdirAll(w.ClaimDir, 0755)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to create claim directory %s", w.ClaimDir)
	}

	err = os.MkdirAll(w.TranscodedDir, 0755)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to create transcoded directory %s", w.TranscodedDir)
	}

	log.Printf("watching %s for new videos\n", w.WatchDir)
	go w.start()
	return w, nil
}

func (w *VideoWatcher) start() {
	dirWatcher, err := fs.NewStableFileWatcher(w.WatchDir, 5*time.Second)
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
	if strings.HasPrefix(".", filepath.Base(path)) {
		return
	}

	// Preserve the directory nesting of the video relative to the watch directory
	// Example: /watch/Movies/Foo/bar.mkv -> Movies/Foo/bar.mkv
	pathSuffix, err := filepath.Rel(w.WatchDir, path)
	if err != nil {
		log.Println(errors.Wrapf(err, "unable to determine path suffix of %s, skipping for now",
			path))
		return
	}

	// Claim the file by moving it out of the watch directory,
	// prevents attempts to process it a second time
	claimPath := filepath.Join(w.ClaimDir, pathSuffix)
	log.Printf("attempting to claim %s\n", path)
	err = fs.MoveFile(path, claimPath)
	if err != nil {
		log.Println(errors.Wrapf(err, "unable to move %s to %s, skipping for now",
			path, claimPath))
		return
	}

	transcodedPath := filepath.Join(w.TranscodedDir, pathSuffix)
	transcodeJobName, err := w.createTranscodeJob(claimPath, transcodedPath)
	if err != nil {
		log.Println(err)
		w.cleanupFailedClaim(claimPath)
		return
	}

	// Assume that the library is the first segment of the path, e.g. /watch/LIBRARY/../video.mkv
	library := strings.Split(pathSuffix, string(os.PathSeparator))[0]

	_, err = w.createUploadJob(transcodeJobName, transcodedPath, claimPath, pathSuffix, library)
	if err != nil {
		log.Println(err)
		err = jobs.Delete(transcodeJobName, Namespace)
		if err != nil {
			log.Println(err)
		}
		w.cleanupFailedClaim(claimPath)
		return
	}
}

func (w *VideoWatcher) cleanupFailedClaim(claimPath string) {
	pathSuffix := strings.Replace(claimPath, w.ClaimDir, "", 1)

	log.Printf("cleaning up failed claim: %s\n", claimPath)
	failedPath := filepath.Join(w.FailedDir, pathSuffix)
	err := fs.MoveFile(claimPath, failedPath)
	if err != nil {
		log.Println(errors.Wrap(err, "unable to cleanup failed claim"))
	}
}
