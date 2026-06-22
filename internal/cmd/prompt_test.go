package cmd

import (
	"bytes"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/syou6162/cursor-agent-cli/internal/cursor"
)

func TestResolvePromptFromFlag(t *testing.T) {
	t.Parallel()

	got, err := resolvePrompt("flag prompt", strings.NewReader("stdin prompt"), func(io.Reader) bool { return false })
	if err != nil {
		t.Fatalf("resolvePrompt() error = %v", err)
	}
	if got != "flag prompt" {
		t.Fatalf("resolvePrompt() = %q, want %q", got, "flag prompt")
	}
}

func TestResolvePromptFromStdin(t *testing.T) {
	t.Parallel()

	got, err := resolvePrompt("", strings.NewReader("prompt from stdin"), func(io.Reader) bool { return false })
	if err != nil {
		t.Fatalf("resolvePrompt() error = %v", err)
	}
	if got != "prompt from stdin" {
		t.Fatalf("resolvePrompt() = %q, want %q", got, "prompt from stdin")
	}
}

func TestResolvePromptFromStdinTrimsWhitespace(t *testing.T) {
	t.Parallel()

	got, err := resolvePrompt("", strings.NewReader("  prompt from stdin\n"), func(io.Reader) bool { return false })
	if err != nil {
		t.Fatalf("resolvePrompt() error = %v", err)
	}
	if got != "prompt from stdin" {
		t.Fatalf("resolvePrompt() = %q, want %q", got, "prompt from stdin")
	}
}

func TestResolvePromptFlagTakesPriorityOverStdin(t *testing.T) {
	t.Parallel()

	got, err := resolvePrompt("flag wins", strings.NewReader("stdin prompt"), func(io.Reader) bool { return false })
	if err != nil {
		t.Fatalf("resolvePrompt() error = %v", err)
	}
	if got != "flag wins" {
		t.Fatalf("resolvePrompt() = %q, want %q", got, "flag wins")
	}
}

func TestResolvePromptMissingWhenTTY(t *testing.T) {
	t.Parallel()

	_, err := resolvePrompt("", strings.NewReader("ignored"), func(io.Reader) bool { return true })
	if err == nil {
		t.Fatal("resolvePrompt() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "--prompt is required") {
		t.Fatalf("resolvePrompt() error = %v, want --prompt is required", err)
	}
}

func TestResolvePromptMissingWhenNilStdin(t *testing.T) {
	t.Parallel()

	_, err := resolvePrompt("", nil, defaultIsTerminal)
	if err == nil {
		t.Fatal("resolvePrompt() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "--prompt is required") {
		t.Fatalf("resolvePrompt() error = %v, want --prompt is required", err)
	}
}

func TestResolvePromptNilIsTerminalUsesDefault(t *testing.T) {
	t.Parallel()

	got, err := resolvePrompt("", strings.NewReader("prompt from stdin"), nil)
	if err != nil {
		t.Fatalf("resolvePrompt() error = %v", err)
	}
	if got != "prompt from stdin" {
		t.Fatalf("resolvePrompt() = %q, want %q", got, "prompt from stdin")
	}
}

func TestDefaultIsTerminalNonFileReaderIsPipe(t *testing.T) {
	t.Parallel()

	if defaultIsTerminal(strings.NewReader("data")) {
		t.Fatal("defaultIsTerminal(strings.NewReader) = true, want false")
	}
}

func TestDefaultIsTerminalNilReaderIsTTY(t *testing.T) {
	t.Parallel()

	if !defaultIsTerminal(nil) {
		t.Fatal("defaultIsTerminal(nil) = false, want true")
	}
}

func TestRunCreatePromptFromStdin(t *testing.T) {
	t.Parallel()

	writer := &stubAgentWriter{
		response: &cursor.CreateAgentResponse{
			Agent: cursor.CreatedAgent{ID: "bc-1", URL: "https://cursor.com/agents/bc-1"},
			Run:   cursor.Run{ID: "run-1", AgentID: "bc-1"},
		},
	}
	root := &Root{
		stdout: &bytes.Buffer{},
		stderr: &bytes.Buffer{},
		stdin:  strings.NewReader("prompt from stdin"),
		clientFactory: func() (cursor.Client, error) {
			return newStubClientWithAgentWriter(writer), nil
		},
	}

	if got := root.Run([]string{
		"create",
		"--repo", "https://github.com/org/repo",
	}); got != ExitSuccess {
		t.Fatalf("Run(create) = %d, want %d", got, ExitSuccess)
	}
	if writer.req.Prompt.Text != "prompt from stdin" {
		t.Fatalf("prompt = %q, want %q", writer.req.Prompt.Text, "prompt from stdin")
	}
}

func TestRunCreatePromptFlagTakesPriorityOverStdin(t *testing.T) {
	t.Parallel()

	writer := &stubAgentWriter{
		response: &cursor.CreateAgentResponse{
			Agent: cursor.CreatedAgent{ID: "bc-1", URL: "https://cursor.com/agents/bc-1"},
			Run:   cursor.Run{ID: "run-1", AgentID: "bc-1"},
		},
	}
	root := &Root{
		stdout: &bytes.Buffer{},
		stderr: &bytes.Buffer{},
		stdin:  strings.NewReader("prompt from stdin"),
		clientFactory: func() (cursor.Client, error) {
			return newStubClientWithAgentWriter(writer), nil
		},
	}

	if got := root.Run([]string{
		"create",
		"--repo", "https://github.com/org/repo",
		"--prompt", "flag prompt",
	}); got != ExitSuccess {
		t.Fatalf("Run(create) = %d, want %d", got, ExitSuccess)
	}
	if writer.req.Prompt.Text != "flag prompt" {
		t.Fatalf("prompt = %q, want %q", writer.req.Prompt.Text, "flag prompt")
	}
}

func TestRunCreateMissingPromptWithTTYStdin(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer
	root := &Root{
		stdout: &bytes.Buffer{},
		stderr: &stderr,
		stdin:  strings.NewReader("prompt from stdin"),
		isTerminal: func(io.Reader) bool {
			return true
		},
		clientFactory: func() (cursor.Client, error) {
			t.Fatal("client should not be called for missing prompt")
			return nil, nil
		},
	}

	if got := root.Run([]string{
		"create",
		"--repo", "https://github.com/org/repo",
	}); got != ExitUsage {
		t.Fatalf("Run(create) = %d, want %d", got, ExitUsage)
	}
	if !strings.Contains(stderr.String(), "--prompt is required") {
		t.Fatalf("stderr = %q, want missing prompt message", stderr.String())
	}
}

func TestRunRunPromptFromStdin(t *testing.T) {
	t.Parallel()

	agentID := "bc-00000000-0000-0000-0000-000000000001"
	writer := &stubRunWriter{
		response: &cursor.CreateRunResponse{
			Run: cursor.Run{
				ID:      "run-00000000-0000-0000-0000-000000000002",
				AgentID: agentID,
				Status:  "CREATING",
			},
		},
	}
	root := &Root{
		stdout: &bytes.Buffer{},
		stderr: &bytes.Buffer{},
		stdin:  strings.NewReader("Fix the tests"),
		clientFactory: func() (cursor.Client, error) {
			return newStubClientWithRunWriter(writer), nil
		},
	}

	if got := root.Run([]string{"run", agentID}); got != ExitSuccess {
		t.Fatalf("Run(run) = %d, want %d", got, ExitSuccess)
	}
	if writer.req.Prompt.Text != "Fix the tests" {
		t.Fatalf("prompt = %q, want %q", writer.req.Prompt.Text, "Fix the tests")
	}
}

func TestRunRunPromptFlagTakesPriorityOverStdin(t *testing.T) {
	t.Parallel()

	writer := &stubRunWriter{
		response: &cursor.CreateRunResponse{
			Run: cursor.Run{ID: "run-1", AgentID: "bc-1", Status: "CREATING"},
		},
	}
	root := &Root{
		stdout: &bytes.Buffer{},
		stderr: &bytes.Buffer{},
		stdin:  strings.NewReader("stdin prompt"),
		clientFactory: func() (cursor.Client, error) {
			return newStubClientWithRunWriter(writer), nil
		},
	}

	if got := root.Run([]string{"run", "bc-1", "--prompt", "flag prompt"}); got != ExitSuccess {
		t.Fatalf("Run(run) = %d, want %d", got, ExitSuccess)
	}
	if writer.req.Prompt.Text != "flag prompt" {
		t.Fatalf("prompt = %q, want %q", writer.req.Prompt.Text, "flag prompt")
	}
}

func TestRunRunMissingPromptWithTTYStdin(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer
	root := &Root{
		stdout: &bytes.Buffer{},
		stderr: &stderr,
		stdin:  strings.NewReader("stdin prompt"),
		isTerminal: func(io.Reader) bool {
			return true
		},
		clientFactory: func() (cursor.Client, error) {
			t.Fatal("client should not be called for missing prompt")
			return nil, nil
		},
	}

	if got := root.Run([]string{"run", "bc-1"}); got != ExitUsage {
		t.Fatalf("Run(run) = %d, want %d", got, ExitUsage)
	}
	if !strings.Contains(stderr.String(), "--prompt is required") {
		t.Fatalf("stderr = %q, want missing prompt message", stderr.String())
	}
}

func TestRunCreateHelpMentionsStdin(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer
	root := &Root{stdout: &bytes.Buffer{}, stderr: &stderr}

	if got := root.Run([]string{"create", "--help"}); got != ExitSuccess {
		t.Fatalf("Run(create --help) = %d, want %d", got, ExitSuccess)
	}
	if !strings.Contains(stderr.String(), "stdin") {
		t.Fatalf("stderr = %q, want stdin mention", stderr.String())
	}
}

func TestRunRunHelpMentionsStdin(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer
	root := &Root{stdout: &bytes.Buffer{}, stderr: &stderr}

	if got := root.Run([]string{"run", "--help"}); got != ExitSuccess {
		t.Fatalf("Run(run --help) = %d, want %d", got, ExitSuccess)
	}
	if !strings.Contains(stderr.String(), "stdin") {
		t.Fatalf("stderr = %q, want stdin mention", stderr.String())
	}
}

func TestRunCreatePromptFromStdinMatchesFlagRequest(t *testing.T) {
	t.Parallel()

	writer := &stubAgentWriter{
		response: &cursor.CreateAgentResponse{
			Agent: cursor.CreatedAgent{ID: "bc-1", URL: "https://cursor.com/agents/bc-1"},
			Run:   cursor.Run{ID: "run-1", AgentID: "bc-1"},
		},
	}
	root := &Root{
		stdout: &bytes.Buffer{},
		stderr: &bytes.Buffer{},
		stdin:  strings.NewReader("Add README"),
		clientFactory: func() (cursor.Client, error) {
			return newStubClientWithAgentWriter(writer), nil
		},
	}

	if got := root.Run([]string{
		"create",
		"--repo", "https://github.com/org/repo",
	}); got != ExitSuccess {
		t.Fatalf("Run(create) = %d, want %d", got, ExitSuccess)
	}

	autoCreatePR := true
	branch := "main"
	wantReq := cursor.CreateAgentRequest{
		Prompt: cursor.AgentPrompt{Text: "Add README"},
		Repos: []cursor.AgentRepo{
			{URL: "https://github.com/org/repo", StartingRef: &branch},
		},
		AutoCreatePR: &autoCreatePR,
	}
	if !reflect.DeepEqual(writer.req, wantReq) {
		t.Fatalf("CreateAgent request = %+v, want %+v", writer.req, wantReq)
	}
}
