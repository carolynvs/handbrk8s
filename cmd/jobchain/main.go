package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/carolynvs/handbrk8s/cmd"
	"github.com/carolynvs/handbrk8s/internal/k8s/jobs"
)

// jobchain -name JOBNAME [-namespace NAMESPACE]
// Exit with 0 only when the job completes successfully
func main() {
	name, namespace := parseFlags()

	done := make(chan struct{})
	jobChan, errChan := jobs.WaitUntilComplete(done, namespace, name)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	for {
		select {
		case <-signals:
			fmt.Println("Stopping...")
			os.Exit(cmd.Interrupted)
		case job, ok := <-jobChan:
			if !ok {
				fmt.Println("Giving up...")
				os.Exit(cmd.RuntimeError)
			}
			fmt.Printf("Job completed sucessfully at %s\n", job.Status.CompletionTime)
			return
		case err, ok := <-errChan:
			if ok {
				fmt.Printf("%#v", err)
			}
			fmt.Println("Giving up...")
			os.Exit(cmd.RuntimeError)
		}
	}
}

func parseFlags() (name, namespace string) {
	fs := flag.NewFlagSet("jobchain", flag.ExitOnError)
	fs.StringVar(&name, "name", "", "job to wait for")
	fs.StringVar(&namespace, "namespace", "", "namespace of the job")
	fs.Parse(os.Args[1:])

	cmd.ExitOnMissingFlag(name, "-name")

	return name, namespace
}
