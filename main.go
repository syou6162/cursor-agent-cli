package main

import (
	"os"

	"github.com/syou6162/cursor-agent-cli/internal/cmd"
)

func main() {
	os.Exit(cmd.NewRoot().Run(os.Args[1:]))
}
