package fs

import (
	"io/ioutil"
	"log"
	"os"
	"time"

	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
)

// StableFile watches for new files, waiting for the file to be completely
// written before signaling an event.
type StableFileWatcher struct {
	watchDir   string
	dirWatcher *fsnotify.Watcher
	done       chan struct{}

	// StableThreshold is the duration that a file must not change
	// before a signaling an event for the file. Defaults to 5 seconds.
	StableThreshold time.Duration

	// Events signal when a file has stabilized.
	Events chan FileEvent
}

// FileEvent signals that a file is in the watch directory is ready to be
// processed.
type FileEvent struct {
	// Path to the file
	Path string
}

// NewStableFileWatcher watcher for a directory.
func NewStableFileWatcher(watchDir string) (*StableFileWatcher, error) {
	w := &StableFileWatcher{
		watchDir:        watchDir,
		done:            make(chan struct{}),
		StableThreshold: 5 * time.Second,
		Events:          make(chan FileEvent),
	}

	dw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to create a file system watcher")
	}

	// Note any preexisting files
	existingFiles, err := ioutil.ReadDir(w.watchDir)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to list %s", w.watchDir)
	}

	// Start watching for new files
	err = dw.Add(w.watchDir)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to start watching %s", watchDir)
	}

	go w.start(existingFiles)

	return w, nil
}

func (w *StableFileWatcher) start(existingFiles []os.FileInfo) {
	for _, file := range existingFiles {
		path := filepath.Join(w.watchDir, file.Name())
		go w.waitUntilFileIsStable(path)
	}

	for {
		select {
		case <-w.done:
			close(w.Events)
			return
		case fileEvent := <-w.dirWatcher.Events:
			if fileEvent.Op&fsnotify.Create == fsnotify.Create {
				go w.waitUntilFileIsStable(fileEvent.Name)
			}
		}
	}
}

// Close all channels.
func (w *StableFileWatcher) Close() {
	close(w.done)
}

// waitUntilFileIsStable waits until the file doesn't change for a set amount of
// time. This prevents acting on a file that is still copying, being written.
func (w *StableFileWatcher) waitUntilFileIsStable(path string) {
	// TODO: reuse the directory watcher and filter
	fw, err := fsnotify.NewWatcher()
	if err != nil {
		log.Println(errors.Wrapf(err, "unable to create watcher, skipping %s", path))
		return
	}
	defer fw.Close()
	err = fw.Add(path)
	if err != nil {
		log.Println(errors.Wrapf(err, "unable to watch %s, skipping", path))
		return
	}

	timer := time.NewTimer(w.StableThreshold)
	defer timer.Stop()

	for {
		select {
		case <-w.done:
			return
		case <-fw.Events:
			// Start the wait over again, the file was changed
			if !timer.Stop() {
				<-timer.C
			}
			timer.Reset(w.StableThreshold)
		case <-timer.C:
			// Make sure the file is still present
			_, err := os.Stat(path)
			if err != nil {
				log.Println(errors.Wrapf(err, "unable to stat %s, skipping", path))
			} else {
				w.Events <- FileEvent{Path: path}
			}
			return
		}
	}
}
