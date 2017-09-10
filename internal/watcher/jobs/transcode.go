package jobs

import (
	"log"
	"path/filepath"

	"github.com/carolynvs/handbrk8s/internal/k8s/jobs"
)

const transcodeJobYaml = `
apiVersion: batch/v1
kind: Job
metadata:
  name: transcode-{{.Name}}
  namespace: handbrk8s
spec:
  template:
    metadata:
      name: handbrake-{{.Name}}
    spec:
      containers:
      - name: handbrake
        image: carolynvs/handbrakecli:latest
        imagePullPolicy: Always
        args:
        - "--preset-import-file"
        - "/config/ghb/presets.json"
        - "-i"
        - "{{.InputPath}}"
        - "-o"
        - "{{.OutputPath}}"
        - "--preset"
        - "{{.Preset}}"
        volumeMounts:
        - mountPath: /watch
          name: mlp
      nodeSelector:
        samba: "yes"
      restartPolicy: OnFailure
      volumes:
      - name: mlp
        hostPath:
          path: /mlp/movies/raw
`

// TranscodeJobValues are the set of values to replace in transcodeJobYaml
type transcodeJobValues struct {
	Name, InputPath, OutputPath, Preset string
}

// CreateTranscodeJob creates a job to transcode a video
func CreateTranscodeJob(inputPath string, outputPath string, preset string) (jobName string, err error) {
	filename := filepath.Base(inputPath)

	log.Println("creating transcode job for ", filename)
	values := transcodeJobValues{
		Name:       sanitizeJobName(filename),
		InputPath:  inputPath,
		OutputPath: outputPath,
		Preset:     preset,
	}
	return jobs.CreateFromTemplate(transcodeJobYaml, values)
}
