package fs

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
)

// StableFile watches for new files, waiting for the file to be completely
// written before signaling an event.
type StableFileWatcher struct {
	watchDir      string
	dirWatcher    *fsnotify.Watcher
	done          chan struct{}
	unstableFiles map[string]struct{}

	// StableThreshold is the duration that a file must not change
	// before a signaling an event for the file.
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
func NewStableFileWatcher(watchDir string, stableThreshold time.Duration) (*StableFileWatcher, error) {
	w := &StableFileWatcher{
		watchDir:        watchDir,
		done:            make(chan struct{}),
		unstableFiles:   make(map[string]struct{}),
		StableThreshold: stableThreshold,
		Events:          make(chan FileEvent),
	}

	dw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, errors.Wrapf(err, "unable to create a file system watcher")
	}
	w.dirWatcher = dw

	// Note any preexisting files
	existingFiles, err := w.readFiles()
	if err != nil {
		return nil, err
	}

	// Start watching for new files
	err = w.dirWatcher.Add(w.watchDir)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to start watching %s", watchDir)
	}

	go w.start(existingFiles)

	return w, nil
}

func (w *StableFileWatcher) readFiles() ([]string, error) {
	var files []string

	filepath.Walk(w.watchDir, func(path string, item os.FileInfo, err error) error {
		if item.IsDir() {
			w.dirWatcher.Add(path)
		} else {
			log.Printf("found existing video: %s\n", path)
			files = append(files, path)
		}
		return nil
	})

	return files, nil
}

func (w *StableFileWatcher) start(existingFiles []string) {
	for _, file := range existingFiles {
		go w.waitUntilFileIsStable(file)
	}

	for {
		select {
		case <-w.done:
			close(w.Events)
			return
		case e := <-w.dirWatcher.Events:
			info, err := os.Stat(e.Name)
			if err != nil {
				// Attempt to stop watching a deleted directory or file
				w.dirWatcher.Remove(e.Name)
				delete(w.unstableFiles, e.Name)
				continue
			}

			if info.IsDir() {
				w.dirWatcher.Add(e.Name)
			} else {
				go w.waitUntilFileIsStable(e.Name)
			}
		}
	}
}

// Close all channels.
func (w *StableFileWatcher) Close() {
	w.dirWatcher.Close()
	close(w.done)
}

// waitUntilFileIsStable waits until the file doesn't change for a set amount of
// time. This prevents acting on a file that is still copying, being written.
func (w *StableFileWatcher) waitUntilFileIsStable(path string) {
	if _, ok := w.unstableFiles[path]; ok {
		return
	}

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
	w.unstableFiles[path] = struct{}{}

	timer := time.NewTimer(w.StableThreshold)
	defer timer.Stop()

	for {
		select {
		case <-w.done:
			delete(w.unstableFiles, path)
			return
		case <-fw.Events:
			// Start the wait over again, the file was changed
			if !timer.Stop() {
				<-timer.C
			}
			timer.Reset(w.StableThreshold)
		case <-timer.C:
			delete(w.unstableFiles, path)
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
