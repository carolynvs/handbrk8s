package main

import (
	"log"

	"github.com/carolynvs/handbrk8s/internal/dashboard"
)

func main() {
	log.Fatal(dashboard.Serve())
}
