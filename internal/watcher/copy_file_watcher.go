package watcher

import (
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
)

// NewFileWatcher watches for new files, waiting for the file to stop changing
// for a period of time before signaling an event.
type CopyFileWatcher struct {
	watchDir   string
	dirWatcher *fsnotify.Watcher

	// StableThreshold is the duration that a file must not change
	// before a signaling an event for the file. Defaults to 5seconds.
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

// NewCopyFileWatcher prepares a new watcher for a directory.
func NewCopyFileWatcher(watchDir string) (*CopyFileWatcher, error) {
	dw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to create a file system watcher")
	}

	// Note any preexisting files
	existingFiles, err := ioutil.ReadDir(watchDir)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to list %s", watchDir)
	}

	// Start watching for new files
	err = dw.Add(watchDir)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to start watching %s", watchDir)
	}

	w := &CopyFileWatcher{
		watchDir:        watchDir,
		dirWatcher:      dw,
		StableThreshold: 5 * time.Second,
		Events:          make(chan FileEvent),
	}

	w.watchExistingFiles(existingFiles)
	go w.watchForNewFiles()

	return w, nil
}

// Close all channels.
func (w *CopyFileWatcher) Close() error {
	err := w.dirWatcher.Close()
	return errors.WithStack(err)
}

func (w *CopyFileWatcher) watchExistingFiles(files []os.FileInfo) {
	for _, file := range files {
		go w.waitUntilFileIsStable(file.Name())
	}
}

func (w *CopyFileWatcher) watchForNewFiles() {
	for fileEvent := range w.dirWatcher.Events {
		if fileEvent.Op&fsnotify.Create == fsnotify.Create {
			go w.waitUntilFileIsStable(fileEvent.Name)
		}
	}

	// Tie closing our events to the underlying watcher
	close(w.Events)
}

// waitUntilFileIsStable waits until the file doesn't change for a set amount of
// time. This prevents acting on a file that is still copying, being written.
func (w *CopyFileWatcher) waitUntilFileIsStable(path string) {
	// todo: this entire thing is leaky, need a done channel or context

	fw, err := fsnotify.NewWatcher()
	if err != nil {
		log.Println(errors.Wrapf(err, "unable to create watcher, skipping %s", path))
	}
	defer fw.Close()
	err = fw.Add(path)
	if err != nil {
		log.Println(errors.Wrapf(err, "unable to watch %s, skipping", path))
	}

	timer := time.NewTimer(w.StableThreshold)
	defer timer.Stop()

	for {
		select {
		case <-fw.Events:
			// Start the wait over again, the file was changed
			if !timer.Stop() {
				<-timer.C
			}
			timer.Reset(w.StableThreshold)
		case <-timer.C:
			w.Events <- FileEvent{Path: path}
			return
		}
	}
}
