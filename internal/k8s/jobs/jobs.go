package jobs

import (
	"log"

	"github.com/carolynvs/handbrk8s/internal/k8s/api"
	"github.com/pkg/errors"
	_ "k8s.io/client-go/pkg/apis/batch/install"
	batchv1 "k8s.io/client-go/pkg/apis/batch/v1"
)

// CreateFromTemplate creates a job on the current cluster from a template
// and set of replacement values.
func CreateFromTemplate(yamlTemplate string, values interface{}) (jobName string, err error) {
	j, err := BuildFromTemplate(yamlTemplate, values)
	if err != nil {
		return "", err
	}

	clusterClient, err := api.GetCurrentClusterClient()
	if err != nil {
		return "", err
	}
	jobclient := clusterClient.BatchV1Client.Jobs(j.Namespace)

	result, err := jobclient.Create(j)
	if err != nil {
		yaml, _ := api.SerializeObject(j)
		return "", errors.Wrapf(err, "unable to create job from:\n%s", yaml)
	}

	log.Printf("created job: %s", result.Name)
	return result.Name, nil
}

// BuildFromTemplate builds a job definition from a template
// and set of replacement values.
func BuildFromTemplate(yamlTemplate string, values interface{}) (*batchv1.Job, error) {
	yaml, err := api.ProcessTemplate(yamlTemplate, values)
	if err != nil {
		return nil, err
	}
	return Deserialize(yaml)
}

// Deserialize reads a job definition from yaml.
func Deserialize(yaml []byte) (*batchv1.Job, error) {
	obj, err := api.DeserializeObject(yaml)
	if err != nil {
		return nil, err
	}

	j, ok := obj.(*batchv1.Job)
	if !ok {
		return nil, errors.Errorf("yaml does not deserialize into a batch/v1 job\n%s", string(yaml))
	}

	return j, err
}
