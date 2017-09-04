package jobs

import (
	"testing"
)

func TestDeserializeJob(t *testing.T) {
	yaml := `
apiVersion: batch/v1
kind: Job
metadata:
  name: jobname
spec:
  template:
    metadata:
      name: podname
    spec:
      containers:
      - name: containername
        image: containerimg
`
	j, err := Deserialize([]byte(yaml))
	if err != nil {
		t.Fatalf("%#v", err)
	}

	if j == nil {
		t.Fatal("didn't deserialize into a job instance")
	}
}
