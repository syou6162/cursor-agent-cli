package cmd

import (
	"flag"
	"fmt"
	"io"
)

func modelsUsage(stderr io.Writer) func() {
	return func() {
		_, _ = fmt.Fprintln(stderr, "Usage: cursor-agent-cli models")
		_, _ = fmt.Fprintln(stderr)
		_, _ = fmt.Fprintln(stderr, "List available Cursor Cloud Agent models")
	}
}

func listUsage(stderr io.Writer, fs *flag.FlagSet) func() {
	return func() {
		_, _ = fmt.Fprintln(stderr, "Usage: cursor-agent-cli list [flags]")
		_, _ = fmt.Fprintln(stderr)
		_, _ = fmt.Fprintln(stderr, "Flags:")
		fs.PrintDefaults()
	}
}
