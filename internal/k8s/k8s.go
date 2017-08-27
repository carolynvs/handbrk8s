package k8s

import (
	"log"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func CreateJob(path string) error {
	log.Println("listing jobs for ", path)
	config, err := rest.InClusterConfig()
	if err != nil {
		return errors.Wrapf(err, "unable to retrieve the current cluster's configuration")
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return errors.Wrapf(err, "unable to create a kubernetes client")
	}

	jobs, err := clientset.BatchV1Client.Jobs("handbrk8s").List(metav1.ListOptions{})
	if err != nil {
		return errors.Wrapf(err, "unable to list jobs")
	}
	log.Printf("There are %d jobs in the cluster\n", len(jobs.Items))
	return nil
}
