package watcher

import (
	"log"
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
)

// NewFileWatcher watches for new files.
// It waits for a file to stop changing size before signaling the event.
type CopyFileWatcher struct {
	watchDir  string
	fsWatcher *fsnotify.Watcher
	Events    chan FileEvent
}

// FileEvent signals that a file is in the watch directory is ready to be
// processed.
type FileEvent struct {
	// Path to the file
	Path string
}

// NewCopyFileWatcher prepares a new watcher for a directory.
func NewCopyFileWatcher(watchDir string) (*CopyFileWatcher, error) {
	fw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to create a file system watcher")
	}

	// Start watching
	err = fw.Add(watchDir)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to start watching %s", watchDir)
	}

	w := &CopyFileWatcher{
		watchDir:  watchDir,
		fsWatcher: fw,
		Events:    make(chan FileEvent),
	}

	go w.start()

	return w, nil
}

// Close all channels.
func (w *CopyFileWatcher) Close() error {
	err := w.fsWatcher.Close()
	return errors.WithStack(err)
}

// Start sending FileEvents.
func (w *CopyFileWatcher) start() {
	for event := range w.fsWatcher.Events {
		log.Println("event:", event)

		if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
			fi, err := os.Stat(event.Name)
			if err != nil {
				log.Println("unable to stat file:", event.Name)
			} else {
				log.Printf("modified file (%s): %s\n", fi.Size(), event.Name)
			}
			w.Events <- FileEvent{Path: event.Name}
		}
	}
	close(w.Events)
}
