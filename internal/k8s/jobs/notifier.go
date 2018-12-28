package jobs

import (
	"log"

	"github.com/carolynvs/handbrk8s/internal/k8s/api"
	"github.com/pkg/errors"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	watchapi "k8s.io/apimachinery/pkg/watch"
)

func WaitUntilComplete(done <-chan struct{}, namespace, name string) (<-chan *batchv1.Job, <-chan error) {
	jobChan := make(chan *batchv1.Job)
	errChan := make(chan error)

	go func() {
		defer close(jobChan)
		defer close(errChan)

		clusterClient, err := api.GetCurrentClusterClient()
		if err != nil {
			errChan <- err
			return
		}

		jobclient := clusterClient.BatchV1().Jobs(namespace)

		opts := metav1.ListOptions{
			FieldSelector: fields.OneTermEqualSelector("metadata.name", name).String(),
		}
		watch, err := jobclient.Watch(opts)
		if err != nil {
			errChan <- errors.Wrapf(err, "Unable to watch %v:jobs for %#v", namespace, opts)
			return
		}
		defer watch.Stop()
		events := watch.ResultChan()

		for {
			select {
			case <-done:
				return
			case e := <-events:
				job, ok := e.Object.(*batchv1.Job)
				if !ok {
					errChan <- errors.Errorf("watch returned a non-job:\n%#v", e.Object)
					continue
				}
				if job.Status.Succeeded > 0 {
					jobChan <- job
				} else {
					log.Printf("job hasn't suceeded yet, current status is %#v", job.Status)
				}
			}
		}
	}()

	return jobChan, errChan
}

func WaitUntilDeleted(done <-chan struct{}, namespace, name string) <-chan error {
	errChan := make(chan error)

	go func() {
		defer close(errChan)

		clusterClient, err := api.GetCurrentClusterClient()
		if err != nil {
			errChan <- err
			return
		}

		jobclient := clusterClient.BatchV1().Jobs(namespace)

		opts := metav1.ListOptions{
			FieldSelector: fields.OneTermEqualSelector("metadata.name", name).String(),
		}
		watch, err := jobclient.Watch(opts)
		if err != nil {
			errChan <- errors.Wrapf(err, "Unable to watch %v:jobs for %#v", namespace, opts)
			return
		}
		defer watch.Stop()
		events := watch.ResultChan()

		for {
			select {
			case <-done:
				return
			case e := <-events:
				if e.Type == watchapi.Deleted {
					return
				}
			}
		}
	}()

	return errChan
}
