package watcher

import (
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/carolynvs/handbrk8s/internal/k8s/jobs"
	"github.com/pkg/errors"
)

// TranscodeJobValues are the set of values to replace in transcodeJobYaml
type transcodeJobValues struct {
	Name, InputPath, OutputDir, OutputPath, Preset string
}

// CreateTranscodeJob creates a job to transcode a video
func (w *VideoWatcher) createTranscodeJob(inputPath string, outputPath string) (jobName string, err error) {
	templateFile := filepath.Join(w.TemplatesDir, "transcode.yaml")
	template, err := ioutil.ReadFile(templateFile)
	if err != nil {
		return "", errors.Wrapf(err, "could not read %s", templateFile)
	}

	filename := filepath.Base(inputPath)

	log.Printf("creating transcode job for %s\n", filename)
	values := transcodeJobValues{
		Name:       jobs.SanitizeJobName(filename),
		InputPath:  inputPath,
		OutputDir:  filepath.Dir(outputPath),
		OutputPath: outputPath,
		Preset:     w.VideoPreset,
	}
	return jobs.CreateFromTemplate(string(template), values)
}
