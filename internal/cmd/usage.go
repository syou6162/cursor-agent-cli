package cmd

import (
	"flag"
	"fmt"
	"io"
)

// Usage callbacks for subcommand help text wired to flag.FlagSet.Usage in root.go.

// modelsUsage returns the Usage callback for the models subcommand.
func modelsUsage(stderr io.Writer) func() {
	return func() {
		fmt.Fprintln(stderr, "Usage: cursor-agent-cli models")
		fmt.Fprintln(stderr)
		fmt.Fprintln(stderr, "List available Cursor Cloud Agent models")
	}
}

// listUsage returns the Usage callback for the list subcommand, including flag defaults.
func listUsage(stderr io.Writer, fs *flag.FlagSet) func() {
	return func() {
		fmt.Fprintln(stderr, "Usage: cursor-agent-cli list [flags]")
		fmt.Fprintln(stderr)
		fmt.Fprintln(stderr, "Flags:")
		fs.PrintDefaults()
	}
}
