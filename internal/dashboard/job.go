package dashboard

import (
	"time"

	"k8s.io/api/batch/v1"
)

type Data struct {
	Jobs []DisplayJob
}
type DisplayJob v1.Job

func (j DisplayJob) Duration() string {
	d := time.Now().Sub(j.Status.StartTime.Time)

	return d.String()
}

func (j DisplayJob) StatusDescription() string {
	if j.Status.Active > 0 {
		return "Running"
	}
	if j.Status.Failed == 0 {
		return "Succeeded"
	}
	return "Failed"
}
