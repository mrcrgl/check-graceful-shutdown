package main

import (
	"log"

	"os"

	"github.com/fid-dev/check-graceful-shutdown/pkg/cli/check-graceful-shutdown/cmd"
)

func main() {
	root := cmd.NewRootCommand()

	root.ParseFlags(os.Args)

	if err := root.Execute(); err != nil {
		log.Fatalf("Execution failed: %s", err)
	}
}
