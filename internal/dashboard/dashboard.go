package dashboard

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"os"

	"github.com/carolynvs/handbrk8s/internal/watcher"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func Serve() error {
	var config *rest.Config
	var err error

	if kubeconfig, ok := os.LookupEnv("KUBECONFIG"); ok {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return err
		}
	} else {
		config, err = rest.InClusterConfig()
		if err != nil {
			return err
		}
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	jobs, err := client.BatchV1().Jobs(watcher.Namespace).List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	t, err := template.New("dashboard").Parse(dashboardTemplate)
	if err != nil {
		return err
	}

	data := Data{
		Jobs: make([]DisplayJob, len(jobs.Items)),
	}
	for i, j := range jobs.Items {
		data.Jobs[i] = DisplayJob(j)
	}

	helloHandler := func(w http.ResponseWriter, req *http.Request) {
		b := &bytes.Buffer{}
		err := t.Execute(b, data)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(500)
		} else {
			w.Write(b.Bytes())
		}

	}

	http.HandleFunc("/", helloHandler)
	return http.ListenAndServe(":80", nil)
}
