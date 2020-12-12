package fs

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
)

// StableFile watches for new files, waiting for the file to be completely
// written before signaling an event.
type StableFileWatcher struct {
	watchDir     string
	done         chan struct{}
	unwatchFiles chan string
	stableFiles  chan string

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
		unwatchFiles:    make(chan string),
		stableFiles:     make(chan string),
		StableThreshold: stableThreshold,
		Events:          make(chan FileEvent),
	}

	go w.start()

	return w, nil
}

func (w *StableFileWatcher) readFiles() ([]string, error) {
	var files []string

	filepath.Walk(w.watchDir, func(path string, item os.FileInfo, err error) error {
		if !item.IsDir() {
			log.Printf("found video: %s\n", path)
			files = append(files, path)
		}
		return nil
	})

	return files, nil
}

func (w *StableFileWatcher) start() {
	watchedFiles := map[string]struct{}{}

	for {
		select {
		case <-w.done:
			close(w.unwatchFiles)
			close(w.stableFiles)
			close(w.Events)
			return
		case path := <-w.unwatchFiles:
			delete(watchedFiles, path)
		case path := <-w.stableFiles:
			delete(watchedFiles, path)
			w.Events <- FileEvent{path}
		default:
			files, err := w.readFiles()
			if err != nil {
				log.Println(err)
				return
			}

			for _, f := range files {
				if _, ok := watchedFiles[f]; ok {
					continue
				}

				info, err := os.Stat(f)
				if err != nil {
					log.Println(errors.Wrapf(err, "could not stat %q, skipping", f))
					continue
				}

				if !info.IsDir() {
					watchedFiles[f] = struct{}{}
					go w.waitUntilFileIsStable(f)
				}
			}
		}

		time.Sleep(time.Second)
	}
}

// Close all channels.
func (w *StableFileWatcher) Close() {
	close(w.done)
}

// waitUntilFileIsStable waits until the file doesn't change for a set amount of
// time. This prevents acting on a file that is still copying, being written.
func (w *StableFileWatcher) waitUntilFileIsStable(path string) {
	timer := time.NewTimer(w.StableThreshold)
	defer timer.Stop()

	var lastSize int64
	for {
		select {
		case <-w.done:
			return
		case <-timer.C:
			// Make sure the file is still present
			_, err := os.Stat(path)
			if err != nil {
				w.unwatchFiles <- path
				log.Println(errors.Wrapf(err, "unable to stat %s, skipping", path))
			} else {
				w.stableFiles <- path
			}
			return
		default:
			info, err := os.Stat(path)
			if err != nil {
				w.unwatchFiles <- path
				log.Println(errors.Wrapf(err, "unable to stat %s, skipping", path))
			}

			if lastSize != info.Size() {
				lastSize = info.Size()
				// Start the wait over again, the file was changed
				if !timer.Stop() {
					<-timer.C
				}
				timer.Reset(w.StableThreshold)
			}
		}

		time.Sleep(time.Second)
	}
}
