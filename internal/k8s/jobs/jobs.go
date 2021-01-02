package jobs

import (
	"context"
	"log"
	"regexp"
	"strings"

	"github.com/carolynvs/handbrk8s/internal/k8s/api"
	"github.com/pkg/errors"
	batchv1 "k8s.io/api/batch/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SanitizeJobName replaces characters that aren't allowed in a k8s name with dashes.
func SanitizeJobName(name string) string {

	name = strings.ToLower(name)
	re := regexp.MustCompile(`[^a-z0-9-]`)
	return re.ReplaceAllString(name, "-")
}

// Delete a job.
func Delete(name, namespace string) error {
	log.Printf("deleting job: %s/%s", namespace, name)
	clusterClient, err := api.GetCurrentClusterClient()
	if err != nil {
		return err
	}
	jobclient := clusterClient.BatchV1().Jobs(namespace)

	// Wait for the associated pods to delete
	foreground := v1.DeletePropagationForeground
	opts := v1.DeleteOptions{
		PropagationPolicy: &foreground,
	}
	err = jobclient.Delete(context.TODO(), name, opts)
	if !apierrors.IsNotFound(err) {
		return errors.Wrapf(err, "unable to delete %s/%s", namespace, name)
	}

	return nil
}

// CreateFromTemplate creates a job on the current cluster from a template
// and set of replacement values.
func CreateFromTemplate(yamlTemplate string, values interface{}) (jobName string, err error) {
	j, err := BuildFromTemplate(yamlTemplate, values)
	if err != nil {
		return "", err
	}

	return CreateOrReplace(j)
}

func CreateOrReplace(j *batchv1.Job) (jobName string, err error) {
	clusterClient, err := api.GetCurrentClusterClient()
	if err != nil {
		return "", err
	}
	jobclient := clusterClient.BatchV1().Jobs(j.Namespace)

	result, err := jobclient.Create(context.TODO(), j, v1.CreateOptions{})
	if apierrors.IsAlreadyExists(err) {
		delerr := Delete(j.Name, j.Namespace)
		if delerr != nil {
			return "", errors.Wrapf(delerr, "unable to delete existing job %s so that it can be recreated", j.Name)
		}

		errChan := WaitUntilDeleted(nil, j.Namespace, j.Name)
		select {
		case delerr, waiting := <-errChan:
			if waiting && delerr != nil {
				return "", errors.Wrapf(delerr, "unable to wait for the %s job to be deleted", j.Name)
			}
		}

		return CreateOrReplace(j)
	} else if err != nil {
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
