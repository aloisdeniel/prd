package main

import (
	"os"

	"github.com/aloisdeniel/prd/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
