package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/syou6162/cursor-agent-cli/internal/cursor"
)

func TestRunDefaultHelloWorld(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	root := &Root{stdout: &stdout, stderr: &bytes.Buffer{}}

	if got := root.Run(nil); got != ExitSuccess {
		t.Fatalf("Run(nil) = %d, want %d", got, ExitSuccess)
	}

	if !strings.Contains(stdout.String(), "hello from cursor-agent-cli") {
		t.Fatalf("stdout = %q, want hello message", stdout.String())
	}
}

func TestRunHelp(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer
	root := &Root{stdout: &bytes.Buffer{}, stderr: &stderr}

	if got := root.Run([]string{"help"}); got != ExitSuccess {
		t.Fatalf("Run(help) = %d, want %d", got, ExitSuccess)
	}

	if !strings.Contains(stderr.String(), "Usage: cursor-agent-cli") {
		t.Fatalf("stderr = %q, want usage text", stderr.String())
	}
}

func TestRunUnknownCommand(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer
	root := &Root{stdout: &bytes.Buffer{}, stderr: &stderr}

	if got := root.Run([]string{"missing"}); got != ExitUsage {
		t.Fatalf("Run(missing) = %d, want %d", got, ExitUsage)
	}

	if !strings.Contains(stderr.String(), "unknown command: missing") {
		t.Fatalf("stderr = %q, want unknown command message", stderr.String())
	}
}

func TestRunModelsMissingAPIKey(t *testing.T) {
	t.Setenv("CURSOR_CLOUD_AGENT_API_KEY", "")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	root := NewRoot()
	root.stdout = &stdout
	root.stderr = &stderr

	if got := root.Run([]string{"models"}); got != ExitConfig {
		t.Fatalf("Run(models) = %d, want %d", got, ExitConfig)
	}
	if !strings.Contains(stderr.String(), "CURSOR_CLOUD_AGENT_API_KEY") {
		t.Fatalf("stderr = %q, want missing API key message", stderr.String())
	}
}

func TestRunModelsSuccess(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	root := &Root{
		stdout: &stdout,
		stderr: &stderr,
		clientFactory: func() (cursor.Client, error) {
			return &spyCursorClient{
				response: &cursor.ListModelsResponse{
					Items: []cursor.Model{{ID: "composer-2", DisplayName: "Composer 2"}},
				},
			}, nil
		},
	}

	if got := root.Run([]string{"models"}); got != ExitSuccess {
		t.Fatalf("Run(models) = %d, want %d", got, ExitSuccess)
	}
	if !strings.Contains(stdout.String(), "composer-2") {
		t.Fatalf("stdout = %q, want composer-2", stdout.String())
	}
}
