package main

import (
	"log"

	"github.com/mrcrgl/check-graceful-shutdown/pkg/cli/check-graceful-shutdown/cmd"
)

func main() {
	root := cmd.NewRootCommand()

	if err := root.Execute(); err != nil {
		log.Fatalf("Execution failed: %s", err)
	}
}
