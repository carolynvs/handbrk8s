package plex

import (
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	"github.com/pkg/errors"
)

type Config struct {
	Server string
	Token  string
}

type Client struct {
	Config
}

func NewClient(cfg Config) Client {
	return Client{Config: cfg}
}

type Library struct {
	c    Client
	Id   string `xml:"key,attr"`
	Name string `xml:"title,attr"`
}

type Video struct {
	Name  string      `xml:"title,attr"`
	Files []VideoFile `xml:"Media>Part"`
}

type VideoFile struct {
	Path string `xml:"file,attr"`
}

func (vf VideoFile) FileName() string {
	return filepath.Base(vf.Path)
}

func (c Client) Get(path string, result interface{}) error {
	url := fmt.Sprintf("%s/%s?X-Plex-Token=%s", c.Server, path, c.Token)

	resp, err := http.Get(url)
	if err != nil {
		return errors.Wrapf(err, "unable to get %s", url)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return errors.Errorf("%d(%s) %s", resp.StatusCode, resp.Status, url)
	} else {
		log.Printf("%d(%s) %s", resp.StatusCode, resp.Status, url)
	}
	defer resp.Body.Close()

	if result != nil {
		err = xml.NewDecoder(resp.Body).Decode(result)
		if err != nil {
			return errors.Wrapf(err, "Cannot decode result from %s into %T", url, result)
		}
	}

	return nil
}

// lookup library id from name
func (c Client) FindLibrary(name string) (library Library, err error) {
	var result struct {
		Libraries []Library `xml:"Directory"`
	}
	err = c.Get("library/sections", &result)
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
	err := l.c.Get(fmt.Sprintf("library/sections/%s/all", l.Id), &result)
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
	err := l.c.Get(fmt.Sprintf("library/sections/%s/refresh", l.Id), nil)
	return errors.Wrapf(err, "unable to update the %s library", l.Name)
}
