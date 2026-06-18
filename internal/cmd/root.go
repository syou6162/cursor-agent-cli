package cmd

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/syou6162/cursor-agent-cli/internal/cursor"
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
	stdout        io.Writer
	stderr        io.Writer
	clientFactory func() (cursor.Client, error)
}

// NewRoot creates a Root command with stdout and stderr wired to os.Stdout and os.Stderr.
func NewRoot() *Root {
	return &Root{
		stdout:        os.Stdout,
		stderr:        os.Stderr,
		clientFactory: cursor.ClientFromEnv,
	}
}

func (r *Root) apiClient() (cursor.Client, error) {
	if r.clientFactory != nil {
		return r.clientFactory()
	}
	return cursor.ClientFromEnv()
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
		return r.runModels(args[1:])
	case "list":
		return r.runList(args[1:])
	default:
		return r.runUnknown(args[0])
	}
}

func (r *Root) runModels(_ []string) int {
	client, err := r.apiClient()
	if err != nil {
		return r.fail(ExitConfig, err)
	}

	resp, err := listModels(context.Background(), client)
	if err != nil {
		return r.fail(ExitAPI, err)
	}
	return r.writeJSON(resp)
}

func (r *Root) runList(args []string) int {
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	fs.SetOutput(r.stderr)
	limit := fs.Int("limit", 20, "maximum number of agents to return")
	if err := fs.Parse(args); err != nil {
		return r.fail(ExitUsage, err)
	}

	client, err := r.apiClient()
	if err != nil {
		return r.fail(ExitConfig, err)
	}

	resp, err := listAgents(context.Background(), client, *limit)
	if err != nil {
		return r.fail(ExitAPI, err)
	}
	return r.writeJSON(resp)
}

func (r *Root) fail(code int, err error) int {
	fmt.Fprintf(r.stderr, "error: %v\n", err)
	return code
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
		fmt.Fprintln(r.stderr, "  list     List agents")
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
