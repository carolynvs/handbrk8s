package jobs

import (
	"log"
	"path/filepath"

	"github.com/carolynvs/handbrk8s/internal/k8s"
)

const uploadJobYaml = `

`

type uploadJobValues struct {
	Name, InitJob, Path string
}

// CreateUploadJob creates a job to upload a video to Plex
func CreateUploadJob(initJob string, path string) (jobName string, err error) {
	filename := filepath.Base(path)

	log.Println("creating upload job for ", filename)
	values := uploadJobValues{
		Name:    sanitizeJobName(filename),
		InitJob: initJob,
		Path:    path,
	}
	return k8s.CreateJobFromTemplate(uploadJobYaml, values)
}
