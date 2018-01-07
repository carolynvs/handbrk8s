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
	testcases := []struct {
		Name        string
		LibraryName string
		WantType    MediaType
	}{
		{Name: "Movies", LibraryName: "Movies", WantType: Movie},
		{Name: "TV Shows", LibraryName: "TV", WantType: Show},
	}

	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			c := buildClient(t)
			lib, err := c.FindLibrary(tc.LibraryName)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			if lib.Type != tc.WantType {
				t.Fatalf("Library type was not deserialized, expected: %v, got: %v", tc.WantType, lib.Type)
			}
		})
	}
}

func TestLibrary_List(t *testing.T) {
	testcases := []struct {
		Name        string
		LibraryName string
		WantType    MediaType
	}{
		{Name: "Movies", LibraryName: "Movies", WantType: Movie},
		{Name: "TV Shows", LibraryName: "TV", WantType: Episode},
	}

	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			c := buildClient(t)
			lib, err := c.FindLibrary(tc.LibraryName)
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

			if video.Type != tc.WantType {
				t.Fatalf("Video type was not deserialized, expected: %v, got: %v", tc.WantType, video.Type)
			}
		})
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
