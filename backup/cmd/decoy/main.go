package main

import (
	"os"

	"github.com/aaron70/decoy/cli"
)

func main() {
	err := cli.Execute()
	if err != nil {
		os.Exit(1)
	}
}
