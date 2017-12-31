package cmd

import (
	"fmt"
	"os"
)

const (
	InvalidArgument int = iota
	Interrupted
	RuntimeError
)

// TODO: Replace with a cmd shell that calls an app which could return an error.
// ExitOnRuntimeError checks for an error, then quits, returning a non-zero exit code.
func ExitOnRuntimeError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(RuntimeError)
	}
}

// ExitOnMissingFlag checks for an empty value, then quits, returning a non-zero exit code.
func ExitOnMissingFlag(value, flag string) {
	if value == "" {
		fmt.Printf("%s is required\n", flag)
		os.Exit(InvalidArgument)
	}
}
