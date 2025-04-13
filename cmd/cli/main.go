package main

import (
	"ObservabilityServer/internal/cli"
	"os"
)

func main() {
	args := os.Args[1:]

	if ok, cmd := cli.ParseFlags(args); ok {
		cmd.Run()
	}
}
