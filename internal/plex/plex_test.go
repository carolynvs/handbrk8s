package plex

import (
	"os"
	"testing"
)

func buildClient(t *testing.T) (c Client) {
	c.Token = os.Getenv("PLEX_TOKEN")
	c.URL = os.Getenv("PLEX_SERVER")

	if c.Token == "" {
		t.Skip("skipping: PLEX_TOKEN is not set")
	}

	if c.URL == "" {
		t.Skip("skipping: PLEX_SERVER is not set")
	}

	return c
}

func TestClient_FindLibrary(t *testing.T) {
	c := buildClient(t)
	_, err := c.FindLibrary("Movies")
	if err != nil {
		t.Fatalf("%#v", err)
	}
}

func TestLibrary_List(t *testing.T) {
	c := buildClient(t)
	lib, err := c.FindLibrary("Movies")
	if err != nil {
		t.Fatalf("%#v", err)
	}

	videos, err := lib.List()
	if err != nil {
		t.Fatalf("%#v", err)
	}

	if len(videos) < 1 {
		t.Fatal("expected a video in the library")
	}

	video := videos[0]
	if video.Files[0].FileName() == "" {
		t.Fatal("expected a video filename")
	}
}

func TestLibrary_HasVideo(t *testing.T) {
	c := buildClient(t)
	lib, err := c.FindLibrary("Movies")
	if err != nil {
		t.Fatalf("%#v", err)
	}

	ok, err := lib.HasVideo("THE_ANIMATRIX.mkv")
	if err != nil {
		t.Fatalf("%#v", err)
	}

	if !ok {
		t.Fatal("Expected the Movies library to contain the animatrix")
	}
}

func TestLibrary_Update(t *testing.T) {
	c := buildClient(t)
	lib, err := c.FindLibrary("Movies")
	if err != nil {
		t.Fatalf("%#v", err)
	}

	err = lib.Update()
	if err != nil {
		t.Fatalf("%#v", err)
	}
}
