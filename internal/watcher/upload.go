package watcher

import (
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/carolynvs/handbrk8s/internal/k8s/jobs"
	"github.com/pkg/errors"
)

type uploadJobValues struct {
	WaitForJob                    string
	Name, TranscodedFile, RawFile string
	DestinationSuffix             string
	PlexServer, PlexToken         string
	PlexLibrary, PlexShare        string
}

// CreateUploadJob creates a job to upload a video to Plex
func (w *VideoWatcher) createUploadJob(waitForJob, transcodedFile, rawFile, pathSuffix, library string) (jobName string, err error) {
	templateFile := filepath.Join(w.TemplatesDir, "upload.yaml")
	template, err := ioutil.ReadFile(templateFile)
	if err != nil {
		return "", errors.Wrapf(err, "could not read %s", templateFile)
	}

	filename := filepath.Base(transcodedFile)

	log.Printf("creating upload job for %s\n", filename)
	values := uploadJobValues{
		Name:              jobs.SanitizeJobName(filename),
		WaitForJob:        waitForJob,
		TranscodedFile:    transcodedFile,
		RawFile:           rawFile,
		DestinationSuffix: pathSuffix,
		PlexServer:        w.PlexCfg.URL,
		PlexToken:         w.PlexCfg.Token,
		PlexLibrary:       library,
		PlexShare:         w.PlexCfg.Share, // Assume that the library name is the share path
	}
	return jobs.CreateFromTemplate(string(template), values)
}
