package api

import (
	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
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
