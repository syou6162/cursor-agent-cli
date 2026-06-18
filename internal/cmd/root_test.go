package cmd

import (
	"bytes"
	"context"
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
			return newStubClientWithModel(stubModelReader{
				response: &cursor.ListModelsResponse{
					Items: []cursor.Model{{ID: "composer-2", DisplayName: "Composer 2"}},
				},
			}), nil
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
			return newStubClientWithAgent(&stubAgentReader{
				response: &cursor.ListAgentsResponse{
					Items: []cursor.Agent{
						{ID: "bc-00000000-0000-0000-0000-000000000001", Name: "Test agent"},
					},
				},
			}), nil
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

	reader := &stubAgentReader{
		response: &cursor.ListAgentsResponse{Items: []cursor.Agent{}},
	}
	var stdout bytes.Buffer
	root := &Root{
		stdout: &stdout,
		stderr: &bytes.Buffer{},
		clientFactory: func() (cursor.Client, error) {
			return newStubClientWithAgent(reader), nil
		},
	}

	if got := root.Run([]string{"list", "--limit", "5"}); got != ExitSuccess {
		t.Fatalf("Run(list --limit 5) = %d, want %d", got, ExitSuccess)
	}
	if reader.limit != 5 {
		t.Fatalf("limit = %d, want 5", reader.limit)
	}
}

func TestRunListAPIError(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer
	root := &Root{
		stdout: &bytes.Buffer{},
		stderr: &stderr,
		clientFactory: func() (cursor.Client, error) {
			return newStubClientWithAgent(&stubAgentReader{
				err: &cursor.APIError{StatusCode: 500, Body: "internal error"},
			}), nil
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

func TestRunCreateMissingAPIKey(t *testing.T) {
	t.Setenv("CURSOR_CLOUD_AGENT_API_KEY", "")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	root := NewRoot()
	root.stdout = &stdout
	root.stderr = &stderr

	if got := root.Run([]string{"create", "--repo", "https://github.com/org/repo", "--prompt", "test"}); got != ExitConfig {
		t.Fatalf("Run(create) = %d, want %d", got, ExitConfig)
	}
	if !strings.Contains(stderr.String(), "CURSOR_CLOUD_AGENT_API_KEY") {
		t.Fatalf("stderr = %q, want missing API key message", stderr.String())
	}
}

func TestRunCreateSuccess(t *testing.T) {
	t.Parallel()

	writer := &stubAgentWriter{
		response: &cursor.CreateAgentResponse{
			Agent: cursor.CreatedAgent{
				ID:   "bc-00000000-0000-0000-0000-000000000001",
				Name: "Test agent",
				URL:  "https://cursor.com/agents/bc-00000000-0000-0000-0000-000000000001",
			},
			Run: cursor.Run{
				ID:      "run-00000000-0000-0000-0000-000000000001",
				AgentID: "bc-00000000-0000-0000-0000-000000000001",
				Status:  "CREATING",
			},
		},
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	root := &Root{
		stdout: &stdout,
		stderr: &stderr,
		clientFactory: func() (cursor.Client, error) {
			return newStubClientWithAgentWriter(writer), nil
		},
	}

	if got := root.Run([]string{
		"create",
		"--repo", "https://github.com/org/repo",
		"--prompt", "Add README",
	}); got != ExitSuccess {
		t.Fatalf("Run(create) = %d, want %d", got, ExitSuccess)
	}
	if !strings.Contains(stdout.String(), "bc-00000000-0000-0000-0000-000000000001") {
		t.Fatalf("stdout = %q, want agent id", stdout.String())
	}
	if !strings.Contains(stdout.String(), "https://cursor.com/agents/bc-00000000-0000-0000-0000-000000000001") {
		t.Fatalf("stdout = %q, want agent url", stdout.String())
	}
	if writer.req.Prompt.Text != "Add README" {
		t.Fatalf("prompt = %q, want Add README", writer.req.Prompt.Text)
	}
	if len(writer.req.Repos) != 1 || writer.req.Repos[0].URL != "https://github.com/org/repo" {
		t.Fatalf("repos = %+v, want one repo", writer.req.Repos)
	}
	if writer.req.Repos[0].StartingRef == nil || *writer.req.Repos[0].StartingRef != "main" {
		t.Fatalf("startingRef = %+v, want main", writer.req.Repos[0].StartingRef)
	}
	if writer.req.AutoCreatePR == nil || !*writer.req.AutoCreatePR {
		t.Fatalf("autoCreatePR = %+v, want true", writer.req.AutoCreatePR)
	}
}

func TestRunCreateWithBranch(t *testing.T) {
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
		clientFactory: func() (cursor.Client, error) {
			return newStubClientWithAgentWriter(writer), nil
		},
	}

	if got := root.Run([]string{
		"create",
		"--repo", "https://github.com/org/repo",
		"--prompt", "Add README",
		"--branch", "develop",
	}); got != ExitSuccess {
		t.Fatalf("Run(create --branch develop) = %d, want %d", got, ExitSuccess)
	}
	if writer.req.Repos[0].StartingRef == nil || *writer.req.Repos[0].StartingRef != "develop" {
		t.Fatalf("startingRef = %+v, want develop", writer.req.Repos[0].StartingRef)
	}
}

func TestRunCreateAPIError(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer
	root := &Root{
		stdout: &bytes.Buffer{},
		stderr: &stderr,
		clientFactory: func() (cursor.Client, error) {
			return newStubClientWithAgentWriter(&stubAgentWriter{
				err: &cursor.APIError{StatusCode: 500, Body: "internal error"},
			}), nil
		},
	}

	if got := root.Run([]string{
		"create",
		"--repo", "https://github.com/org/repo",
		"--prompt", "Add README",
	}); got != ExitAPI {
		t.Fatalf("Run(create) = %d, want %d", got, ExitAPI)
	}
	if !strings.Contains(stderr.String(), "status=500") {
		t.Fatalf("stderr = %q, want API error message", stderr.String())
	}
}

func TestRunCreateHelp(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer
	root := &Root{stdout: &bytes.Buffer{}, stderr: &stderr}

	if got := root.Run([]string{"create", "--help"}); got != ExitSuccess {
		t.Fatalf("Run(create --help) = %d, want %d", got, ExitSuccess)
	}
	if !strings.Contains(stderr.String(), "Usage: cursor-agent-cli create") {
		t.Fatalf("stderr = %q, want create usage text", stderr.String())
	}
}

func TestRunCreateMissingRepo(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer
	root := &Root{
		stdout: &bytes.Buffer{},
		stderr: &stderr,
		clientFactory: func() (cursor.Client, error) {
			t.Fatal("client should not be called for missing repo")
			return nil, nil
		},
	}

	if got := root.Run([]string{"create", "--prompt", "Add README"}); got != ExitUsage {
		t.Fatalf("Run(create) = %d, want %d", got, ExitUsage)
	}
	if !strings.Contains(stderr.String(), "--repo is required") {
		t.Fatalf("stderr = %q, want missing repo message", stderr.String())
	}
}

func TestRunCreateMissingPrompt(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer
	root := &Root{
		stdout: &bytes.Buffer{},
		stderr: &stderr,
		clientFactory: func() (cursor.Client, error) {
			t.Fatal("client should not be called for missing prompt")
			return nil, nil
		},
	}

	if got := root.Run([]string{"create", "--repo", "https://github.com/org/repo"}); got != ExitUsage {
		t.Fatalf("Run(create) = %d, want %d", got, ExitUsage)
	}
	if !strings.Contains(stderr.String(), "--prompt is required") {
		t.Fatalf("stderr = %q, want missing prompt message", stderr.String())
	}
}

type stubRunWriter struct {
	agentID  string
	req      cursor.CreateRunRequest
	response *cursor.CreateRunResponse
	err      error
}

func (s *stubRunWriter) CreateRun(_ context.Context, agentID string, req cursor.CreateRunRequest) (*cursor.CreateRunResponse, error) {
	s.agentID = agentID
	s.req = req
	if s.err != nil {
		return nil, s.err
	}
	return s.response, nil
}

func TestRunRunMissingAPIKey(t *testing.T) {
	t.Setenv("CURSOR_CLOUD_AGENT_API_KEY", "")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	root := NewRoot()
	root.stdout = &stdout
	root.stderr = &stderr

	if got := root.Run([]string{"run", "bc-1", "--prompt", "test"}); got != ExitConfig {
		t.Fatalf("Run(run) = %d, want %d", got, ExitConfig)
	}
	if !strings.Contains(stderr.String(), "CURSOR_CLOUD_AGENT_API_KEY") {
		t.Fatalf("stderr = %q, want missing API key message", stderr.String())
	}
}

func TestRunRunSuccess(t *testing.T) {
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
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	root := &Root{
		stdout: &stdout,
		stderr: &stderr,
		clientFactory: func() (cursor.Client, error) {
			return newStubClientWithRunWriter(writer), nil
		},
	}

	if got := root.Run([]string{"run", agentID, "--prompt", "Fix the failing test"}); got != ExitSuccess {
		t.Fatalf("Run(run) = %d, want %d", got, ExitSuccess)
	}
	if !strings.Contains(stdout.String(), "run-00000000-0000-0000-0000-000000000002") {
		t.Fatalf("stdout = %q, want run id", stdout.String())
	}
	if writer.agentID != agentID {
		t.Fatalf("agentID = %q, want %q", writer.agentID, agentID)
	}
	if writer.req.Prompt.Text != "Fix the failing test" {
		t.Fatalf("prompt = %q, want Fix the failing test", writer.req.Prompt.Text)
	}
}

func TestRunRunAgentBusy(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer
	root := &Root{
		stdout: &bytes.Buffer{},
		stderr: &stderr,
		clientFactory: func() (cursor.Client, error) {
			return newStubClientWithRunWriter(&stubRunWriter{
				err: cursor.ErrAgentBusy,
			}), nil
		},
	}

	if got := root.Run([]string{"run", "bc-1", "--prompt", "Fix the failing test"}); got != ExitAPI {
		t.Fatalf("Run(run) = %d, want %d", got, ExitAPI)
	}
	if !strings.Contains(stderr.String(), "agent_busy") {
		t.Fatalf("stderr = %q, want agent_busy message", stderr.String())
	}
}

func TestRunRunAPIError(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer
	root := &Root{
		stdout: &bytes.Buffer{},
		stderr: &stderr,
		clientFactory: func() (cursor.Client, error) {
			return newStubClientWithRunWriter(&stubRunWriter{
				err: &cursor.APIError{StatusCode: 500, Body: "internal error"},
			}), nil
		},
	}

	if got := root.Run([]string{"run", "bc-1", "--prompt", "Fix the failing test"}); got != ExitAPI {
		t.Fatalf("Run(run) = %d, want %d", got, ExitAPI)
	}
	if !strings.Contains(stderr.String(), "status=500") {
		t.Fatalf("stderr = %q, want API error message", stderr.String())
	}
}

func TestRunRunHelp(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer
	root := &Root{stdout: &bytes.Buffer{}, stderr: &stderr}

	if got := root.Run([]string{"run", "--help"}); got != ExitSuccess {
		t.Fatalf("Run(run --help) = %d, want %d", got, ExitSuccess)
	}
	if !strings.Contains(stderr.String(), "Usage: cursor-agent-cli run") {
		t.Fatalf("stderr = %q, want run usage text", stderr.String())
	}
}

func TestRunRunMissingAgentID(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer
	root := &Root{
		stdout: &bytes.Buffer{},
		stderr: &stderr,
		clientFactory: func() (cursor.Client, error) {
			t.Fatal("client should not be called for missing agent_id")
			return nil, nil
		},
	}

	if got := root.Run([]string{"run", "--prompt", "Fix the failing test"}); got != ExitUsage {
		t.Fatalf("Run(run) = %d, want %d", got, ExitUsage)
	}
	if !strings.Contains(stderr.String(), "agent_id is required") {
		t.Fatalf("stderr = %q, want missing agent_id message", stderr.String())
	}
}

func TestRunRunMissingPrompt(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer
	root := &Root{
		stdout: &bytes.Buffer{},
		stderr: &stderr,
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
