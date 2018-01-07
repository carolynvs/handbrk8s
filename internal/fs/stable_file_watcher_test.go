package fs

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

// An atomic counter
type counter struct {
	val int32
}

func (c *counter) increment() {
	atomic.AddInt32(&c.val, 1)
}

func (c *counter) value() int32 {
	return atomic.LoadInt32(&c.val)
}

var testStableThreshold = 1 * time.Second

func TestCopyFileWatcher_NewFile(t *testing.T) {
	t.Parallel()

	tmpDir, err := ioutil.TempDir("", t.Name())
	if err != nil {
		t.Fatalf("%#v", err)
	}
	defer os.RemoveAll(tmpDir)
	t.Log("watching ", tmpDir)

	w, err := NewStableFileWatcher(tmpDir, testStableThreshold)

	// Track how many times an event is raised
	var gotEvents counter
	done := make(chan bool)
	go func() {
		for e := range w.Events {
			t.Log(e)
			gotEvents.increment()
		}

		// Stop the goroutine once the events has been closed
		done <- true
	}()

	// Create a file in the watched directory
	tmpfile := filepath.Join(tmpDir, "foo.txt")
	f, err := os.Create(tmpfile)
	if err != nil {
		t.Fatalf("%#v", err)
	}
	err = f.Close()
	if err != nil {
		t.Fatalf("%#v", err)
	}
	time.Sleep(50 * time.Millisecond)

	// Write to it a few times
	for i := 0; i < 10; i++ {
		if gotEvents.value() > 0 {
			t.Fatalf("expected no events to be raised until the StableThreshold has been reached but got %d events", gotEvents)
		}

		f, err = os.OpenFile(tmpfile, os.O_WRONLY, 0666)
		if err != nil {
			t.Fatalf("%#v", err)
		}
		_, err = f.WriteString(fmt.Sprintf("%d", i))
		if err != nil {
			t.Fatalf("%#v", err)
		}

		err = f.Sync()
		if err != nil {
			t.Fatalf("%#v", err)
		}

		err = f.Close()
		if err != nil {
			t.Fatalf("%#v", err)
		}

		time.Sleep(50 * time.Millisecond)
	}

	// Give the file time to be considered stable
	time.Sleep(w.StableThreshold)

	// Stop listening for events
	w.Close()

	// Wait for all the events to be processed
	t.Log("wait for all events to be processed")
	<-done

	var wantEvents int32 = 1
	if gotEvents.value() != wantEvents {
		t.Fatalf("expected %d events, got %d", wantEvents, gotEvents)
	}
}

func TestCopyFileWatcher_ExistingFile(t *testing.T) {
	t.Parallel()

	tmpDir, err := ioutil.TempDir("", t.Name())
	if err != nil {
		t.Fatalf("%#v", err)
	}
	defer os.RemoveAll(tmpDir)
	t.Log("watching ", tmpDir)

	// Create a file in the watched directory
	tmpfile := filepath.Join(tmpDir, "foo.txt")
	f, err := os.Create(tmpfile)
	if err != nil {
		t.Fatalf("%#v", err)
	}
	err = f.Close()
	if err != nil {
		t.Fatalf("%#v", err)
	}

	w, err := NewStableFileWatcher(tmpDir, testStableThreshold)

	// Track how many times an event is raised
	var gotEvents counter
	done := make(chan bool)
	go func() {
		for e := range w.Events {
			t.Log(e)
			gotEvents.increment()
		}

		// Stop the goroutine once the events has been closed
		done <- true
	}()

	// Give the file time to be considered stable
	time.Sleep(w.StableThreshold * 2)

	// Stop listening for events
	w.Close()

	// Wait for all the events to be processed
	t.Log("wait for all events to be processed")
	<-done

	var wantEvents int32 = 1
	if gotEvents.value() != wantEvents {
		t.Fatalf("expected %d events, got %d", wantEvents, gotEvents)
	}
}

func TestCopyFileWatcher_DeletedFile(t *testing.T) {
	t.Parallel()

	tmpDir, err := ioutil.TempDir("", t.Name())
	if err != nil {
		t.Fatalf("%#v", err)
	}
	defer os.RemoveAll(tmpDir)
	t.Log("watching ", tmpDir)

	w, err := NewStableFileWatcher(tmpDir, testStableThreshold)

	// Track how many times an event is raised
	var gotEvents counter
	done := make(chan bool)
	go func() {
		for e := range w.Events {
			t.Log(e)
			gotEvents.increment()
		}

		// Stop the goroutine once the events has been closed
		done <- true
	}()

	// Create a file in the watched directory
	tmpfile := filepath.Join(tmpDir, "foo.txt")
	f, err := os.Create(tmpfile)
	if err != nil {
		t.Fatalf("%#v", err)
	}
	err = f.Close()
	if err != nil {
		t.Fatalf("%#v", err)
	}
	time.Sleep(50 * time.Millisecond)

	err = os.Remove(tmpfile)
	if err != nil {
		t.Fatalf("%#v", err)
	}

	// Give the file time to be considered stable
	time.Sleep(w.StableThreshold)

	// Stop listening for events
	w.Close()

	// Wait for all the events to be processed
	t.Log("wait for all events to be processed")
	<-done

	var wantEvents int32 = 0
	if gotEvents.value() != wantEvents {
		t.Fatalf("expected %d events, got %d", wantEvents, gotEvents)
	}
}

func TestCopyFileWatcher_NestedFile(t *testing.T) {
	t.Parallel()

	tmpDir, err := ioutil.TempDir("", t.Name())
	if err != nil {
		t.Fatalf("%#v", err)
	}
	defer os.RemoveAll(tmpDir)
	t.Log("watching", tmpDir)

	w, err := NewStableFileWatcher(tmpDir, testStableThreshold)

	// Track how many times an event is raised
	var gotEvents counter
	done := make(chan bool)
	go func() {
		for e := range w.Events {
			t.Log(e)
			gotEvents.increment()
		}

		// Stop the goroutine once the events has been closed
		done <- true
	}()

	// Create a file in the watched directory
	err = os.MkdirAll(filepath.Join(tmpDir, "movies"), 0755)
	if err != nil {
		t.Fatalf("%#v", err)
	}

	tmpfile := filepath.Join(tmpDir, "/movies/foo.txt")
	f, err := os.Create(tmpfile)
	if err != nil {
		t.Fatalf("%#v", err)
	}
	err = f.Close()
	if err != nil {
		t.Fatalf("%#v", err)
	}
	time.Sleep(50 * time.Millisecond)

	// Write to it a few times
	for i := 0; i < 10; i++ {
		if gotEvents.value() > 0 {
			t.Fatalf("expected no events to be raised until the StableThreshold has been reached but got %d events", gotEvents)
		}

		f, err = os.OpenFile(tmpfile, os.O_WRONLY, 0666)
		if err != nil {
			t.Fatalf("%#v", err)
		}
		_, err = f.WriteString(fmt.Sprintf("%d", i))
		if err != nil {
			t.Fatalf("%#v", err)
		}

		err = f.Sync()
		if err != nil {
			t.Fatalf("%#v", err)
		}

		err = f.Close()
		if err != nil {
			t.Fatalf("%#v", err)
		}

		time.Sleep(50 * time.Millisecond)
	}

	// Give the file time to be considered stable
	time.Sleep(w.StableThreshold)

	// Stop listening for events
	w.Close()

	// Wait for all the events to be processed
	t.Log("wait for all events to be processed")
	<-done

	var wantEvents int32 = 1
	if gotEvents.value() != wantEvents {
		t.Fatalf("expected %d events, got %d", wantEvents, gotEvents)
	}
}
