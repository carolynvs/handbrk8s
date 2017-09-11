package jobs

import (
	"log"
	"path/filepath"

	"github.com/carolynvs/handbrk8s/internal/k8s/jobs"
	"github.com/carolynvs/handbrk8s/internal/plex"
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
        - mountPath: /movies
          name: mlp
        - mountPath: /plex
          name: deathstar
      nodeSelector:
        samba: "yes"
      # Do not restart containers after they exit
      restartPolicy: Never #OnFailure
      volumes:
      - name: mlp
        hostPath:
          path: /mlp/movies
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
func CreateUploadJob(waitForJob, transcodedFile, rawFile string, plexCfg plex.Config, destLib, destShare string) (jobName string, err error) {
	filename := filepath.Base(transcodedFile)

	log.Println("creating upload job for ", filename)
	values := uploadJobValues{
		Name:           sanitizeJobName(filename),
		WaitForJob:     waitForJob,
		TranscodedFile: transcodedFile,
		RawFile:        rawFile,
		PlexServer:     plexCfg.Server,
		PlexToken:      plexCfg.Token,
		PlexLibrary:    destLib,
		PlexShare:      destShare,
	}
	return jobs.CreateFromTemplate(uploadJobYaml, values)
}
