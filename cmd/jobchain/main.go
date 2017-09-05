package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/carolynvs/handbrk8s/internal/k8s/jobs"
)

const (
	retCodeInvalidArguments int = iota
	retCodeInterrupted
	retCodeRuntimeError
)

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
			os.Exit(retCodeInterrupted)
		case job, ok := <-jobChan:
			if ok {
				fmt.Printf("Job completed sucessfully at %s\n", job.Status.CompletionTime)
				return
			} else {
				fmt.Println("Giving up...")
				os.Exit(retCodeRuntimeError)
			}
		case err, ok := <-errChan:
			if ok {
				fmt.Println(err)
			}
		}
	}
}

func parseFlags() (name, namespace string) {
	flag.StringVar(&name, "name", "", "job to wait for")
	flag.StringVar(&namespace, "namespace", "", "namespace of the job")
	flag.Parse()

	if name == "" {
		fmt.Println("Waits for a job to complete successfully")
		fmt.Println("jobchain [-namespace] -name")
		os.Exit(retCodeInvalidArguments)
	}

	return name, namespace
}
