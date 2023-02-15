package main

import (
	"os"

	"github.com/katiem0/gh-export-secrets/cmd"
)

func main() {
	// Instantiate and execute root command
	cmd := cmd.NewCmd()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
