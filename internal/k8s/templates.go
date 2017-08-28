package k8s

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
        command: ["/usr/bin/HandBrakeCLI"]
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
      # Do not restart containers after they exit
      restartPolicy: Never #OnFailure
      volumes:
      - name: mlp
        hostPath:
          path: /mlp/movies/raw
`

// TranscodeJobValues are the set of values to replace in transcodeJobYaml
type transcodeJobValues struct {
	Name, InputPath, OutputPath, Preset string
}

const uploadJobYaml = `

`

type uploadJobValues struct {
	Name, Path string
}
