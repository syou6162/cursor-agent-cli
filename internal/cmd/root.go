package cmd

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
)

const (
	ExitSuccess = 0
	ExitError   = 1
	ExitAPI     = 2
	ExitUsage   = 2
	ExitConfig  = 3
)

// Root is the top-level command dispatcher for cursor-agent-cli.
type Root struct {
	stdout io.Writer
	stderr io.Writer
}

// NewRoot creates a Root command with stdout and stderr wired to os.Stdout and os.Stderr.
func NewRoot() *Root {
	return &Root{
		stdout: os.Stdout,
		stderr: os.Stderr,
	}
}

// Run dispatches to a subcommand or prints the default hello-world response.
func (r *Root) Run(args []string) int {
	if len(args) == 0 {
		return r.runHello()
	}

	switch args[0] {
	case "help", "-h", "--help":
		return r.runHelp(args[1:])
	case "models":
		return NewModels(r.stdout, r.stderr).Run(args[1:])
	default:
		return r.runUnknown(args[0])
	}
}

func (r *Root) runHello() int {
	return r.writeJSON(map[string]string{
		"message": "hello from cursor-agent-cli",
	})
}

func (r *Root) runHelp(_ []string) int {
	fs := flag.NewFlagSet("cursor-agent-cli", flag.ContinueOnError)
	fs.SetOutput(r.stderr)
	fs.Usage = func() {
		fmt.Fprintln(r.stderr, "Usage: cursor-agent-cli [command]")
		fmt.Fprintln(r.stderr)
		fmt.Fprintln(r.stderr, "Commands:")
		fmt.Fprintln(r.stderr, "  help     Show usage information")
		fmt.Fprintln(r.stderr, "  models   List available models")
	}
	fs.Usage()
	return ExitSuccess
}

func (r *Root) runUnknown(name string) int {
	fmt.Fprintf(r.stderr, "unknown command: %s\n", name)
	return ExitUsage
}

func (r *Root) writeJSON(v any) int {
	enc := json.NewEncoder(r.stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		fmt.Fprintf(r.stderr, "error: %v\n", err)
		return ExitError
	}
	return ExitSuccess
}
