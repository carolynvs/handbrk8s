package k8s

import (
	"log"
	"path/filepath"
	"regexp"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	apiv1 "k8s.io/client-go/pkg/api/v1"
	batchv1 "k8s.io/client-go/pkg/apis/batch/v1"
	"k8s.io/client-go/rest"
	//"k8s.io/kubernetes/pkg/api/testapi"
)

const namespace = "handbrk8s"

func CreateJob(path string) error {
	log.Println("creating a job for ", path)

	// make a friendly name
	filename := filepath.Base(path)
	jobName := sanitizeJobName(filename)

	config, err := rest.InClusterConfig()
	if err != nil {
		return errors.Wrapf(err, "unable to retrieve the current cluster's configuration")
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return errors.Wrapf(err, "unable to create a kubernetes client")
	}

	jobclient := clientset.BatchV1Client.Jobs(namespace)
	jobs, err := jobclient.List(metav1.ListOptions{})
	if err != nil {
		return errors.Wrapf(err, "unable to list jobs")
	}
	log.Printf("There are %d jobs in the cluster\n", len(jobs.Items))

	// TODO: unmarshal this from template + json/yaml
	j := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      jobName,
			Labels:    map[string]string{"jobgroup": "handbrake"},
		},
		Spec: batchv1.JobSpec{
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: namespace,
					Name:      "handbrake-" + jobName,
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
								"-i", "/watch/twister-sample.mkv",
								"-o", "/watch/twister-sample.done.mkv",
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
					Volumes: []apiv1.Volume{
						{
							Name: "mlp",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: "/mlp/movies/raw",
								},
							},
						},
					},
					RestartPolicy: apiv1.RestartPolicyNever,
				},
			},
		},
	}

	result, err := jobclient.Create(j)
	if err != nil {
		/*var buf bytes.Buffer
		jobDef := yaml.NewDecodingSerializer(testapi.Default.Codec())
		err := jobDef.Encode(j, &buf)*/
		return errors.Wrapf(err, "unable to create job: %s", jobName)
	}

	log.Printf("created job: %s", result.Name)
	return nil
}

func sanitizeJobName(name string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9._-]`)
	return re.ReplaceAllString(name, "-")
}
