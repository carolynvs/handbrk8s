package k8s

import (
	"log"
	"text/template"

	"bytes"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api"
	_ "k8s.io/client-go/pkg/api/install"
	_ "k8s.io/client-go/pkg/apis/batch/install"
	batchv1 "k8s.io/client-go/pkg/apis/batch/v1"
	"k8s.io/client-go/rest"
)

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

// CreateJobFromTemplate creates a job on the current cluster from a template
// and set of replacement values.
func CreateJobFromTemplate(yamlTemplate string, values interface{}) (jobName string, err error) {
	j, err := BuildJobFromTemplate(yamlTemplate, values)
	if err != nil {
		return "", err
	}

	clusterClient, err := GetCurrentClusterClient()
	if err != nil {
		return "", err
	}
	jobclient := clusterClient.BatchV1Client.Jobs(j.Namespace)

	result, err := jobclient.Create(j)
	if err != nil {
		yaml, _ := SerializeObject(j)
		return "", errors.Wrapf(err, "unable to create job from:\n%s", yaml)
	}

	log.Printf("created job: %s", result.Name)
	return result.Name, nil
}

// BuildJobFromTemplate builds a job definition from a template
// and set of replacement values.
func BuildJobFromTemplate(yamlTemplate string, values interface{}) (*batchv1.Job, error) {
	yaml, err := ProcessTemplate(yamlTemplate, values)
	if err != nil {
		return nil, err
	}
	return DeserializeJob(yaml)
}

// ProcessTemplate substitutes the supplied values into a template.
func ProcessTemplate(yamlTemplate string, values interface{}) (yaml []byte, err error) {
	t, err := template.New("").Parse(yamlTemplate)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to parse the yaml template\n%s", yamlTemplate)
	}

	var buf bytes.Buffer
	err = t.Execute(&buf, values)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to execute the yaml template with the values:\n%#v", values)
	}

	return buf.Bytes(), nil
}

// DeserializeJob reads a job definition from yaml.
func DeserializeJob(yaml []byte) (*batchv1.Job, error) {
	obj, err := DeserializeObject(yaml)
	if err != nil {
		return nil, err
	}

	j, ok := obj.(*batchv1.Job)
	if !ok {
		return nil, errors.Errorf("yaml does not deserialize into a batch/v1 job\n%s", string(yaml))
	}

	return j, err
}

// DeserializeObject reads a k8s object from yaml.
func DeserializeObject(yaml []byte) (runtime.Object, error) {
	serializer := api.Codecs.UniversalDeserializer()

	obj, _, err := serializer.Decode(yaml, nil, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot decode the yaml into a k8s runtime object\n%s", string(yaml))
	}
	return obj, err
}

// SerializeObject returns the yaml representation of a k8s object.
func SerializeObject(obj runtime.Object) (string, error) {
	serializer := json.NewYAMLSerializer(json.DefaultMetaFactory, nil, nil)

	var buf bytes.Buffer
	err := serializer.Encode(obj, &buf)
	if err != nil {
		return "", errors.Wrapf(err, "unable to serialize object:\n%#v", obj)
	}
	return buf.String(), nil
}
