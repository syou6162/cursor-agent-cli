package cmd

import (
	"bytes"
	"strings"
	"testing"
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
