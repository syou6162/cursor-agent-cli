package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/syou6162/cursor-agent-cli/internal/cursor"
)

// Models lists available models from the Cursor Cloud Agent API.
type Models struct {
	stdout io.Writer
	stderr io.Writer
	client cursor.Client
}

// NewModels creates a Models command that reads the API client from the environment.
func NewModels(stdout, stderr io.Writer) *Models {
	return &Models{
		stdout: stdout,
		stderr: stderr,
	}
}

// NewModelsWithClient creates a Models command with an injected API client.
func NewModelsWithClient(stdout, stderr io.Writer, client cursor.Client) *Models {
	return &Models{
		stdout: stdout,
		stderr: stderr,
		client: client,
	}
}

// Run executes the models subcommand.
func (m *Models) Run(_ []string) int {
	client := m.client
	if client == nil {
		var err error
		client, err = cursor.ClientFromEnv()
		if err != nil {
			fmt.Fprintf(m.stderr, "error: %v\n", err)
			return ExitConfig
		}
	}

	resp, err := client.ListModels(context.Background())
	if err != nil {
		fmt.Fprintf(m.stderr, "error: %v\n", err)
		return ExitAPI
	}

	enc := json.NewEncoder(m.stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(resp); err != nil {
		fmt.Fprintf(m.stderr, "error: %v\n", err)
		return ExitError
	}
	return ExitSuccess
}
