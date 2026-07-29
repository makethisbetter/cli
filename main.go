package main

import (
	"os"
	_ "time/tzdata"

	"github.com/makethisbetter/cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
