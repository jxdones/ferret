package main

import (
	"os"

	"github.com/jxdones/ferret/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
