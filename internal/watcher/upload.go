package watcher

import (
	"log"
	"path/filepath"

	"github.com/carolynvs/handbrk8s/internal/k8s/jobs"
)

const uploadJobYaml = `
apiVersion: batch/v1
kind: Job
metadata:
  name: upload-{{.Name}}
  namespace: handbrk8s
spec:
  template:
    metadata:
      name: handbrake-{{.Name}}
    spec:
      initContainers:
      - name: jobchain
        image: carolynvs/jobchain:latest
        imagePullPolicy: Always
        args:
        - "--namespace"
        - "handbrk8s"
        - "--name"
        - "{{.WaitForJob}}"
      containers:
      - name: uploader
        image: carolynvs/handbrk8s-uploader:latest
        imagePullPolicy: Always
        args:
        - "-f"
        - "{{.TranscodedFile}}"
        - "--plex-server"
        - "{{.PlexServer}}"
        - "--plex-library"
        - "{{.PlexLibrary}}"
        - "--plex-share"
        - "{{.PlexShare}}"
        - "--raw"
        - "{{.RawFile}}"
        env:
        - name: PLEX_TOKEN
          value: {{.PlexToken}}
        volumeMounts:
        - mountPath: /work
          name: ponyshare
        - mountPath: /plex
          name: deathstar
      # Do not restart containers after they exit
      restartPolicy: Never #OnFailure
      volumes:
      - name: ponyshare
        hostPath:
          path: /mlp
      - name: deathstar
        hostPath:
          path: /deathstar/Multimedia
`

type uploadJobValues struct {
	WaitForJob                    string
	Name, TranscodedFile, RawFile string
	PlexServer, PlexToken         string
	PlexLibrary, PlexShare        string
}

// CreateUploadJob creates a job to upload a video to Plex
func (w *VideoWatcher) createUploadJob(waitForJob, transcodedFile, rawFile string) (jobName string, err error) {
	filename := filepath.Base(transcodedFile)

	log.Printf("creating upload job for %s\n", filename)
	values := uploadJobValues{
		Name:           jobs.SanitizeJobName(filename),
		WaitForJob:     waitForJob,
		TranscodedFile: transcodedFile,
		RawFile:        rawFile,
		PlexServer:     w.DestLib.Config.Server,
		PlexToken:      w.DestLib.Config.Token,
		PlexLibrary:    w.DestLib.Name,
		PlexShare:      w.DestLib.Share,
	}
	return jobs.CreateFromTemplate(uploadJobYaml, values)
}
