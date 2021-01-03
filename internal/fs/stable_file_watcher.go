package fs

import (
	"log"
	"os"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/radovskyb/watcher"
)

// StableFile watches for new files, waiting for the file to be completely
// written before signaling an event.
type StableFileWatcher struct {
	watchDir      string
	dirWatcher    *watcher.Watcher
	pollingPeriod time.Duration
	done          chan struct{}
	unstableFiles sync.Map

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
		dirWatcher:      watcher.New(),
		pollingPeriod:   100 * time.Millisecond,
		done:            make(chan struct{}),
		StableThreshold: stableThreshold,
		Events:          make(chan FileEvent),
	}

	// Start watching for new files
	err := w.dirWatcher.AddRecursive(w.watchDir)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to watch directory %s", watchDir)
	}

	go w.start()

	return w, nil
}

func (w *StableFileWatcher) start() {
	for path, fi := range w.dirWatcher.WatchedFiles() {
		if !fi.IsDir() {
			go w.waitUntilFileIsStable(path)
		}
	}

	go func() {
		for {
			select {
			case <-w.done:
				close(w.Events)
				return
			case e := <-w.dirWatcher.Event:
				//log.Println(e.Op, e.Name())
				if e.IsDir() {
					continue
				}

				if e.Op == watcher.Remove {
					// Attempt to stop watching a deleted directory or file
					w.unstableFiles.Delete(e.Path)
					continue
				}

				go w.waitUntilFileIsStable(e.Path)
			}
		}
	}()

	go func() {
		err := w.dirWatcher.Start(w.pollingPeriod)
		if err != nil {
			panic(errors.Wrap(err, "unable to start watcher"))
		}
	}()

	w.dirWatcher.Wait()
}

// Close all channels.
func (w *StableFileWatcher) Close() {
	w.dirWatcher.Close()
	close(w.done)
}

// waitUntilFileIsStable waits until the file doesn't change for a set amount of
// time. This prevents acting on a file that is still copying, being written.
func (w *StableFileWatcher) waitUntilFileIsStable(path string) {
	if _, ok := w.unstableFiles.LoadOrStore(path, struct{}{}); ok {
		return
	}

	log.Printf("waiting for %s to stabilize\n", path)
	fw := watcher.New()
	defer fw.Close()
	err := fw.Add(path)
	if err != nil {
		log.Println(errors.Wrapf(err, "unable to watch %s, skipping", path))
		w.unstableFiles.Delete(path)
		return
	}
	go fw.Start(w.pollingPeriod)

	timer := time.NewTimer(w.StableThreshold)
	defer timer.Stop()

	for {
		select {
		case <-w.done:
			w.unstableFiles.Delete(path)
			return
		case e := <-fw.Event:
			if e.Op == watcher.Remove {
				w.unstableFiles.Delete(path)
				return
			}

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
		}
	}
}
