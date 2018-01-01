package plex

import (
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path/filepath"

	"github.com/pkg/errors"
)

// ServerConfig is the set of information necessary to connect to a Plex server.
type ServerConfig struct {
	URL   string
	Token string
}

// LibraryConfig is the set of information necessary to upload videos to a Plex library.
type LibraryConfig struct {
	ServerConfig
	Name  string
	Share string
}

type Client struct {
	ServerConfig
}

func NewClient(cfg ServerConfig) Client {
	return Client{ServerConfig: cfg}
}

type Library struct {
	c    Client
	Id   string    `xml:"key,attr"`
	Name string    `xml:"title,attr"`
	Type MediaType `xml:"type,attr"`
}

type MediaType string

func (t MediaType) ToFilter() string {
	switch t {
	case Movie:
		return "1"
	case Show:
		return "4"
	case Season:
		return "4"
	case Episode:
		return "4"
	case Artist:
		return "10"
	case Album:
		return "10"
	case Track:
		return "10"
	case Photo:
		return "14"
	default:
		return ""
	}
}

const (
	Movie   MediaType = "movie"
	Show    MediaType = "show"
	Season  MediaType = "season"
	Episode MediaType = "episode"
	Artist  MediaType = "artist"
	Album   MediaType = "album"
	Track   MediaType = "track"
	Photo   MediaType = "photo"
)

type Video struct {
	Name  string      `xml:"title,attr"`
	Type  MediaType   `xml:"type,attr"`
	Files []VideoFile `xml:"Media>Part"`
}

type VideoFile struct {
	Path string `xml:"file,attr"`
}

func (vf VideoFile) FileName() string {
	return filepath.Base(vf.Path)
}

func (c Client) Get(format string, query map[string]string, result interface{}, a ...interface{}) error {
	baseUrl := fmt.Sprintf(c.URL+"/"+format, a...)
	u, err := url.Parse(baseUrl)
	if err != nil {
		return errors.Wrapf(err, "invalid url %s", baseUrl)
	}

	qs := u.Query()
	for key, val := range query {
		qs.Add(key, val)
	}
	qs.Add("X-Plex-Token", c.Token)
	u.RawQuery = qs.Encode()

	resp, err := http.Get(u.String())
	if err != nil {
		return errors.Wrapf(err, "unable to get %s", u)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return errors.Errorf("%d(%s) %s", resp.StatusCode, resp.Status, u)
	} else {
		log.Printf("%d(%s) %s", resp.StatusCode, resp.Status, u)
	}
	defer resp.Body.Close()

	if result != nil {
		err = xml.NewDecoder(resp.Body).Decode(result)
		if err != nil {
			return errors.Wrapf(err, "Cannot decode result from %s into %T", u, result)
		}
	}

	return nil
}

// lookup library id from name
func (c Client) FindLibrary(name string) (library Library, err error) {
	var result struct {
		Libraries []Library `xml:"Directory"`
	}
	err = c.Get("library/sections", nil, &result)
	if err != nil {
		return library, errors.Wrap(err, "unable to list Plex libraries")
	}

	for _, l := range result.Libraries {
		if l.Name == name {
			l.c = c
			return l, nil
		}
	}

	return library, errors.Errorf("library not found: %s", name)
}

// list library contents
func (l Library) List() ([]Video, error) {
	var result struct {
		Videos []Video `xml:"Video"`
	}

	query := map[string]string{"type": l.Type.ToFilter()}
	err := l.c.Get("library/sections/%s/all", query, &result, l.Id)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list videos in the %s library")
	}

	return result.Videos, nil
}

func (l Library) HasVideo(filename string) (bool, error) {
	videos, err := l.List()
	if err != nil {
		return false, err
	}

	for _, video := range videos {
		for _, file := range video.Files {
			if file.FileName() == filename {
				return true, nil
			}
		}

	}
	return false, nil
}

// refresh library
func (l *Library) Update() error {
	err := l.c.Get("library/sections/%s/refresh", nil, nil, l.Id)
	return errors.Wrapf(err, "unable to update the %s library", l.Name)
}
