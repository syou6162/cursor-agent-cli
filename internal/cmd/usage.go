package cmd

import (
	"flag"
	"fmt"
	"io"
)

func modelsUsage(stderr io.Writer) func() {
	return func() {
		fmt.Fprintln(stderr, "Usage: cursor-agent-cli models")
		fmt.Fprintln(stderr)
		fmt.Fprintln(stderr, "List available Cursor Cloud Agent models")
	}
}

func listUsage(stderr io.Writer, fs *flag.FlagSet) func() {
	return func() {
		fmt.Fprintln(stderr, "Usage: cursor-agent-cli list [flags]")
		fmt.Fprintln(stderr)
		fmt.Fprintln(stderr, "Flags:")
		fs.PrintDefaults()
	}
}
