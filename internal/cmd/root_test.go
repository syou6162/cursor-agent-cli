package cmd

import (
	"bytes"
	"errors"
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

func TestRunListMissingAPIKey(t *testing.T) {
	t.Setenv("CURSOR_CLOUD_AGENT_API_KEY", "")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	root := NewRoot()
	root.stdout = &stdout
	root.stderr = &stderr

	if got := root.Run([]string{"list"}); got != ExitConfig {
		t.Fatalf("Run(list) = %d, want %d", got, ExitConfig)
	}
	if !strings.Contains(stderr.String(), "CURSOR_CLOUD_AGENT_API_KEY") {
		t.Fatalf("stderr = %q, want missing API key message", stderr.String())
	}
}

func TestRunListSuccess(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	root := &Root{
		stdout: &stdout,
		stderr: &stderr,
		clientFactory: func() (cursor.Client, error) {
			return &spyCursorClient{
				agentsResponse: &cursor.ListAgentsResponse{
					Items: []cursor.Agent{
						{ID: "bc-00000000-0000-0000-0000-000000000001", Name: "Test agent"},
					},
				},
			}, nil
		},
	}

	if got := root.Run([]string{"list"}); got != ExitSuccess {
		t.Fatalf("Run(list) = %d, want %d", got, ExitSuccess)
	}
	if !strings.Contains(stdout.String(), "bc-00000000-0000-0000-0000-000000000001") {
		t.Fatalf("stdout = %q, want agent id", stdout.String())
	}
}

func TestRunListWithLimit(t *testing.T) {
	t.Parallel()

	spy := &spyCursorClient{
		agentsResponse: &cursor.ListAgentsResponse{Items: []cursor.Agent{}},
	}
	var stdout bytes.Buffer
	root := &Root{
		stdout: &stdout,
		stderr: &bytes.Buffer{},
		clientFactory: func() (cursor.Client, error) {
			return spy, nil
		},
	}

	if got := root.Run([]string{"list", "--limit", "5"}); got != ExitSuccess {
		t.Fatalf("Run(list --limit 5) = %d, want %d", got, ExitSuccess)
	}
	if spy.agentsLimit != 5 {
		t.Fatalf("agentsLimit = %d, want 5", spy.agentsLimit)
	}
}

func TestRunListAPIError(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer
	root := &Root{
		stdout: &bytes.Buffer{},
		stderr: &stderr,
		clientFactory: func() (cursor.Client, error) {
			return &spyCursorClient{
				agentsErr: errors.New("Cursor API error (status=500): internal error"),
			}, nil
		},
	}

	if got := root.Run([]string{"list"}); got != ExitAPI {
		t.Fatalf("Run(list) = %d, want %d", got, ExitAPI)
	}
	if !strings.Contains(stderr.String(), "status=500") {
		t.Fatalf("stderr = %q, want API error message", stderr.String())
	}
}

func TestRunListHelp(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer
	root := &Root{stdout: &bytes.Buffer{}, stderr: &stderr}

	if got := root.Run([]string{"list", "--help"}); got != ExitSuccess {
		t.Fatalf("Run(list --help) = %d, want %d", got, ExitSuccess)
	}
	if !strings.Contains(stderr.String(), "Usage: cursor-agent-cli list") {
		t.Fatalf("stderr = %q, want list usage text", stderr.String())
	}
}

func TestRunListInvalidLimit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		args  []string
		want  string
	}{
		{name: "zero", args: []string{"list", "--limit", "0"}, want: "greater than 0"},
		{name: "negative", args: []string{"list", "--limit", "-1"}, want: "greater than 0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var stderr bytes.Buffer
			root := &Root{
				stdout: &bytes.Buffer{},
				stderr: &stderr,
				clientFactory: func() (cursor.Client, error) {
					t.Fatal("client should not be called for invalid limit")
					return nil, nil
				},
			}

			if got := root.Run(tt.args); got != ExitUsage {
				t.Fatalf("Run(%v) = %d, want %d", tt.args, got, ExitUsage)
			}
			if !strings.Contains(stderr.String(), tt.want) {
				t.Fatalf("stderr = %q, want %q", stderr.String(), tt.want)
			}
		})
	}
}
