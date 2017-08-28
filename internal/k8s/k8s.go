package k8s

import (
	"log"
	"path/filepath"
	"regexp"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api"
	_ "k8s.io/client-go/pkg/api/install"
	apiv1 "k8s.io/client-go/pkg/api/v1"
	_ "k8s.io/client-go/pkg/apis/batch/install"
	batchv1 "k8s.io/client-go/pkg/apis/batch/v1"
	"k8s.io/client-go/rest"
)

const namespace = "handbrk8s"

const jobYaml = `
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

// GetCurrentClusterClient gets a client for the current cluster upon which
// we are currently executing upon. Only works when running in a cluster.
func GetCurrentClusterClient() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, errors.Wrapf(err, "unable to retrieve the current cluster's configuration")
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to create a kubernetes client")
	}

	return clientset, nil
}

// CreateTranscodeJob creates a job to transcode the specified video
func CreateTranscodeJob(inputPath string, outputPath string) (jobName string, err error) {
	// TODO: make a job from a yaml template and a map of replacement values
	log.Println("creating a job for ", inputPath)

	// make a friendly name
	filename := filepath.Base(inputPath)
	jobName = sanitizeJobName(filename)

	clusterClient, err := GetCurrentClusterClient()
	if err != nil {
		return jobName, err
	}

	jobclient := clusterClient.BatchV1Client.Jobs(namespace)

	j := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      jobName,
		},
		Spec: batchv1.JobSpec{
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: namespace,
					Name:      "transcode-" + jobName,
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:            "handbrake",
							Image:           "carolynvs/handbrakecli:latest",
							ImagePullPolicy: apiv1.PullAlways,
							Command:         []string{"/usr/bin/HandBrakeCLI"},
							Args: []string{
								"--preset-import-file", "/config/ghb/presets.json",
								"-i", inputPath,
								"-o", outputPath,
								"--preset", "tivo",
							},
							VolumeMounts: []apiv1.VolumeMount{
								{
									Name:      "mlp",
									MountPath: "/watch",
								},
							},
						},
					},
					NodeSelector: map[string]string{
						"samba": "yes",
					},
					Volumes: []apiv1.Volume{ // TODO: The volume source needs to be configurable
						{
							Name: "mlp",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: "/mlp/movies/raw",
								},
							},
						},
					},
					RestartPolicy: apiv1.RestartPolicyOnFailure,
				},
			},
		},
	}

	result, err := jobclient.Create(j)
	if err != nil {
		return jobName, errors.Wrapf(err, "unable to create job: %s", jobName)
	}

	log.Printf("created job: %s", result.Name)
	return jobName, nil
}

func CreateUploaderJob(path string) error {
	log.Println("creating stub uploader job")

	return nil
}

func DeserializeJob(yaml []byte) (*batchv1.Job, error) {
	decode := api.Codecs.UniversalDeserializer().Decode

	obj, _, err := decode(yaml, nil, nil)
	if err != nil {
		return nil, err
	}

	j, ok := obj.(*batchv1.Job)
	if !ok {
		return nil, errors.New("yaml didn't not deserialize into a batch/v1 job")
	}

	return j, err
}

func sanitizeJobName(name string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9._-]`)
	return re.ReplaceAllString(name, "-")
}
