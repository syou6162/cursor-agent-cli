package cmd

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/syou6162/cursor-agent-cli/internal/cursor"
)

func parseStatusOutput(t *testing.T, stdout string) map[string]any {
	t.Helper()

	var parsed map[string]any
	if err := json.Unmarshal([]byte(stdout), &parsed); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, stdout = %q", err, stdout)
	}
	return parsed
}

func parseStatusCLI(t *testing.T, stdout string) map[string]any {
	t.Helper()

	parsed := parseStatusOutput(t, stdout)
	cli, ok := parsed["_cli"].(map[string]any)
	if !ok {
		t.Fatalf("_cli missing in %v", parsed)
	}
	return cli
}

func TestRunDefaultShowsUsage(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer
	root := &Root{stdout: &bytes.Buffer{}, stderr: &stderr}

	if got := root.Run(nil); got != ExitSuccess {
		t.Fatalf("Run(nil) = %d, want %d", got, ExitSuccess)
	}

	if !strings.Contains(stderr.String(), "Usage: cursor-agent-cli") {
		t.Fatalf("stderr = %q, want usage text", stderr.String())
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

func TestRunModelsHelp(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer
	root := &Root{stdout: &bytes.Buffer{}, stderr: &stderr}

	if got := root.Run([]string{"models", "--help"}); got != ExitSuccess {
		t.Fatalf("Run(models --help) = %d, want %d", got, ExitSuccess)
	}
	if !strings.Contains(stderr.String(), "Usage: cursor-agent-cli models") {
		t.Fatalf("stderr = %q, want models usage text", stderr.String())
	}
	if !strings.Contains(stderr.String(), "List available Cursor Cloud Agent models") {
		t.Fatalf("stderr = %q, want models description", stderr.String())
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
		tt := tt
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

func TestRunStatusMissingAPIKey(t *testing.T) {
	t.Setenv("CURSOR_CLOUD_AGENT_API_KEY", "")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	root := NewRoot()
	root.stdout = &stdout
	root.stderr = &stderr

	if got := root.Run([]string{"status", "bc-1", "run-1"}); got != ExitConfig {
		t.Fatalf("Run(status) = %d, want %d", got, ExitConfig)
	}
	cli := parseStatusCLI(t, stdout.String())
	if cli["state"] != cliStateConfigError {
		t.Fatalf("_cli.state = %v, want %s", cli["state"], cliStateConfigError)
	}
	if !strings.Contains(stderr.String(), "CURSOR_CLOUD_AGENT_API_KEY") {
		t.Fatalf("stderr = %q, want missing API key message", stderr.String())
	}
}

func TestRunStatusSuccess(t *testing.T) {
	t.Parallel()

	agentID := "bc-00000000-0000-0000-0000-000000000001"
	runID := "run-00000000-0000-0000-0000-000000000001"
	result := "Added README.md"
	prURL := "https://github.com/org/repo/pull/123"
	reader := &stubRunReader{
		responses: []*cursor.RunStatusResponse{
			{
				ID:      runID,
				AgentID: agentID,
				Status:  "FINISHED",
				Result:  &result,
				Git: &cursor.RunGitInfo{
					Branches: []cursor.RunGitBranch{
						{PRURL: &prURL},
					},
				},
			},
		},
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	root := &Root{
		stdout: &stdout,
		stderr: &stderr,
		clientFactory: func() (cursor.Client, error) {
			return newStubClientWithRunReader(reader), nil
		},
	}

	if got := root.Run([]string{"status", agentID, runID}); got != ExitSuccess {
		t.Fatalf("Run(status) = %d, want %d", got, ExitSuccess)
	}
	parsed := parseStatusOutput(t, stdout.String())
	if parsed["status"] != "FINISHED" {
		t.Fatalf("status = %v, want FINISHED", parsed["status"])
	}
	cli := parsed["_cli"].(map[string]any)
	if cli["state"] != cliStateSuccess {
		t.Fatalf("_cli.state = %v, want %s", cli["state"], cliStateSuccess)
	}
	if cli["pollingCount"] != float64(1) {
		t.Fatalf("_cli.pollingCount = %v, want 1", cli["pollingCount"])
	}
	if !strings.Contains(stdout.String(), result) {
		t.Fatalf("stdout = %q, want result text", stdout.String())
	}
	if !strings.Contains(stdout.String(), prURL) {
		t.Fatalf("stdout = %q, want prUrl", stdout.String())
	}
	if reader.agentID != agentID {
		t.Fatalf("agentID = %q, want %q", reader.agentID, agentID)
	}
	if reader.runID != runID {
		t.Fatalf("runID = %q, want %q", reader.runID, runID)
	}
}

func TestRunStatusWatchPollsUntilTerminal(t *testing.T) {
	agentID := "bc-00000000-0000-0000-0000-000000000001"
	runID := "run-00000000-0000-0000-0000-000000000001"
	reader := &stubRunReader{
		responses: []*cursor.RunStatusResponse{
			{ID: runID, AgentID: agentID, Status: "RUNNING"},
			{ID: runID, AgentID: agentID, Status: "FINISHED"},
		},
	}
	var stdout bytes.Buffer
	root := &Root{
		stdout: &stdout,
		stderr: &bytes.Buffer{},
		clientFactory: func() (cursor.Client, error) {
			return newStubClientWithRunReader(reader), nil
		},
	}

	origSleepAfter := sleepAfter
	sleepAfter = func(time.Duration) <-chan time.Time {
		ch := make(chan time.Time, 1)
		ch <- time.Now()
		return ch
	}
	t.Cleanup(func() { sleepAfter = origSleepAfter })

	if got := root.Run([]string{"status", agentID, runID, "--watch", "--interval", "5"}); got != ExitSuccess {
		t.Fatalf("Run(status --watch) = %d, want %d", got, ExitSuccess)
	}
	if reader.calls != 2 {
		t.Fatalf("GetRunStatus calls = %d, want 2", reader.calls)
	}
	parsed := parseStatusOutput(t, stdout.String())
	if parsed["status"] != "FINISHED" {
		t.Fatalf("status = %v, want FINISHED", parsed["status"])
	}
	cli := parsed["_cli"].(map[string]any)
	if cli["state"] != cliStateSuccess {
		t.Fatalf("_cli.state = %v, want %s", cli["state"], cliStateSuccess)
	}
	if cli["pollingCount"] != float64(2) {
		t.Fatalf("_cli.pollingCount = %v, want 2", cli["pollingCount"])
	}
	if strings.Count(stdout.String(), "\"status\"") != 1 {
		t.Fatalf("stdout should contain exactly one JSON object, got %q", stdout.String())
	}
}

func TestRunStatusWatchTimeout(t *testing.T) {
	agentID := "bc-1"
	runID := "run-1"
	reader := &stubRunReader{
		responses: []*cursor.RunStatusResponse{
			{ID: runID, AgentID: agentID, Status: "RUNNING"},
		},
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	root := &Root{
		stdout: &stdout,
		stderr: &stderr,
		clientFactory: func() (cursor.Client, error) {
			return newStubClientWithRunReader(reader), nil
		},
	}

	origSleepAfter := sleepAfter
	sleepAfter = func(time.Duration) <-chan time.Time {
		ch := make(chan time.Time, 1)
		ch <- time.Now()
		return ch
	}
	t.Cleanup(func() { sleepAfter = origSleepAfter })

	if got := root.Run([]string{"status", agentID, runID, "--watch", "--interval", "5", "--timeout", "1"}); got != ExitError {
		t.Fatalf("Run(status --watch --timeout) = %d, want %d", got, ExitError)
	}
	parsed := parseStatusOutput(t, stdout.String())
	if parsed["status"] != "RUNNING" {
		t.Fatalf("status = %v, want RUNNING", parsed["status"])
	}
	cli := parsed["_cli"].(map[string]any)
	if cli["state"] != cliStateTimeout {
		t.Fatalf("_cli.state = %v, want %s", cli["state"], cliStateTimeout)
	}
	if cli["exitCode"] != float64(ExitError) {
		t.Fatalf("_cli.exitCode = %v, want %d", cli["exitCode"], ExitError)
	}
	if !strings.Contains(stderr.String(), "timeout waiting for run to complete") {
		t.Fatalf("stderr = %q, want timeout message", stderr.String())
	}
}

func TestRunStatusInvalidInterval(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
	}{
		{name: "zero", args: []string{"status", "bc-1", "run-1", "--interval", "0"}},
		{name: "one", args: []string{"status", "bc-1", "run-1", "--interval", "1"}},
		{name: "four", args: []string{"status", "bc-1", "run-1", "--interval", "4"}},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var stdout bytes.Buffer
			var stderr bytes.Buffer
			root := &Root{
				stdout: &stdout,
				stderr: &stderr,
				clientFactory: func() (cursor.Client, error) {
					t.Fatal("client should not be called for invalid interval")
					return nil, nil
				},
			}

			if got := root.Run(tt.args); got != ExitUsage {
				t.Fatalf("Run(%v) = %d, want %d", tt.args, got, ExitUsage)
			}
			cli := parseStatusCLI(t, stdout.String())
			if cli["state"] != cliStateUsageError {
				t.Fatalf("_cli.state = %v, want %s", cli["state"], cliStateUsageError)
			}
			if !strings.Contains(stderr.String(), "at least 5 seconds") {
				t.Fatalf("stderr = %q, want interval minimum message", stderr.String())
			}
		})
	}
}

func TestRunStatusRateLimit(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	root := &Root{
		stdout: &stdout,
		stderr: &stderr,
		clientFactory: func() (cursor.Client, error) {
			return newStubClientWithRunReader(&stubRunReader{
				err: &cursor.APIError{StatusCode: 429, Body: "too many requests"},
			}), nil
		},
	}

	if got := root.Run([]string{"status", "bc-1", "run-1"}); got != ExitAPI {
		t.Fatalf("Run(status) = %d, want %d", got, ExitAPI)
	}
	cli := parseStatusCLI(t, stdout.String())
	if cli["state"] != cliStateAPIError {
		t.Fatalf("_cli.state = %v, want %s", cli["state"], cliStateAPIError)
	}
	if !strings.Contains(stderr.String(), "rate limit exceeded: please wait before retrying") {
		t.Fatalf("stderr = %q, want rate limit message", stderr.String())
	}
}

func TestRunStatusAPIError(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	root := &Root{
		stdout: &stdout,
		stderr: &stderr,
		clientFactory: func() (cursor.Client, error) {
			return newStubClientWithRunReader(&stubRunReader{
				err: &cursor.APIError{StatusCode: 500, Body: "internal error"},
			}), nil
		},
	}

	if got := root.Run([]string{"status", "bc-1", "run-1"}); got != ExitAPI {
		t.Fatalf("Run(status) = %d, want %d", got, ExitAPI)
	}
	parsed := parseStatusOutput(t, stdout.String())
	if parsed["id"] != "run-1" {
		t.Fatalf("id = %v, want run-1", parsed["id"])
	}
	if parsed["agentId"] != "bc-1" {
		t.Fatalf("agentId = %v, want bc-1", parsed["agentId"])
	}
	cli := parsed["_cli"].(map[string]any)
	if cli["state"] != cliStateAPIError {
		t.Fatalf("_cli.state = %v, want %s", cli["state"], cliStateAPIError)
	}
	if cli["pollingCount"] != float64(1) {
		t.Fatalf("_cli.pollingCount = %v, want 1", cli["pollingCount"])
	}
	if !strings.Contains(stderr.String(), "status=500") {
		t.Fatalf("stderr = %q, want API error message", stderr.String())
	}
}

func TestRunStatusHelp(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer
	root := &Root{stdout: &bytes.Buffer{}, stderr: &stderr}

	if got := root.Run([]string{"status", "--help"}); got != ExitSuccess {
		t.Fatalf("Run(status --help) = %d, want %d", got, ExitSuccess)
	}
	if !strings.Contains(stderr.String(), "Usage: cursor-agent-cli status") {
		t.Fatalf("stderr = %q, want status usage text", stderr.String())
	}
}

func TestRunStatusMissingAgentID(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	root := &Root{
		stdout: &stdout,
		stderr: &stderr,
		clientFactory: func() (cursor.Client, error) {
			t.Fatal("client should not be called for missing agent_id")
			return nil, nil
		},
	}

	if got := root.Run([]string{"status"}); got != ExitUsage {
		t.Fatalf("Run(status) = %d, want %d", got, ExitUsage)
	}
	cli := parseStatusCLI(t, stdout.String())
	if cli["state"] != cliStateUsageError {
		t.Fatalf("_cli.state = %v, want %s", cli["state"], cliStateUsageError)
	}
	if !strings.Contains(stderr.String(), "agent_id is required") {
		t.Fatalf("stderr = %q, want missing agent_id message", stderr.String())
	}
}

func TestRunStatusMissingRunID(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	root := &Root{
		stdout: &stdout,
		stderr: &stderr,
		clientFactory: func() (cursor.Client, error) {
			t.Fatal("client should not be called for missing run_id")
			return nil, nil
		},
	}

	if got := root.Run([]string{"status", "bc-1"}); got != ExitUsage {
		t.Fatalf("Run(status) = %d, want %d", got, ExitUsage)
	}
	cli := parseStatusCLI(t, stdout.String())
	if cli["state"] != cliStateUsageError {
		t.Fatalf("_cli.state = %v, want %s", cli["state"], cliStateUsageError)
	}
	if !strings.Contains(stderr.String(), "run_id is required") {
		t.Fatalf("stderr = %q, want missing run_id message", stderr.String())
	}
}

func TestRunStatusWatchFlagsBeforeArgsRequiresRunID(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	root := &Root{
		stdout: &stdout,
		stderr: &stderr,
		clientFactory: func() (cursor.Client, error) {
			t.Fatal("client should not be called when run_id is missing")
			return nil, nil
		},
	}

	if got := root.Run([]string{"status", "--watch", "bc-1"}); got != ExitUsage {
		t.Fatalf("Run(status --watch bc-1) = %d, want %d", got, ExitUsage)
	}
	cli := parseStatusCLI(t, stdout.String())
	if cli["state"] != cliStateUsageError {
		t.Fatalf("_cli.state = %v, want %s", cli["state"], cliStateUsageError)
	}
	if !strings.Contains(stderr.String(), "run_id is required") {
		t.Fatalf("stderr = %q, want missing run_id message", stderr.String())
	}
}

func TestRunStatusWatchFlagsBeforeArgs(t *testing.T) {
	agentID := "bc-00000000-0000-0000-0000-000000000001"
	runID := "run-00000000-0000-0000-0000-000000000001"
	reader := &stubRunReader{
		responses: []*cursor.RunStatusResponse{
			{ID: runID, AgentID: agentID, Status: "FINISHED"},
		},
	}
	var stdout bytes.Buffer
	root := &Root{
		stdout: &stdout,
		stderr: &bytes.Buffer{},
		clientFactory: func() (cursor.Client, error) {
			return newStubClientWithRunReader(reader), nil
		},
	}

	if got := root.Run([]string{"status", "--watch", "--interval", "5", agentID, runID}); got != ExitSuccess {
		t.Fatalf("Run(status --watch agent run) = %d, want %d", got, ExitSuccess)
	}
	if reader.agentID != agentID {
		t.Fatalf("agentID = %q, want %q", reader.agentID, agentID)
	}
	if reader.runID != runID {
		t.Fatalf("runID = %q, want %q", reader.runID, runID)
	}
}

func TestRunCancelSuccess(t *testing.T) {
	t.Parallel()

	agentID := "bc-00000000-0000-0000-0000-000000000001"
	runID := "run-00000000-0000-0000-0000-000000000001"
	writer := &stubCancelWriter{
		response: &cursor.CancelRunResponse{ID: runID},
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	root := &Root{
		stdout: &stdout,
		stderr: &stderr,
		clientFactory: func() (cursor.Client, error) {
			return newStubClientWithCancelWriter(writer), nil
		},
	}

	if got := root.Run([]string{"cancel", agentID, runID}); got != ExitSuccess {
		t.Fatalf("Run(cancel) = %d, want %d", got, ExitSuccess)
	}
	if !strings.Contains(stdout.String(), runID) {
		t.Fatalf("stdout = %q, want run id", stdout.String())
	}
	if writer.agentID != agentID {
		t.Fatalf("agentID = %q, want %q", writer.agentID, agentID)
	}
	if writer.runID != runID {
		t.Fatalf("runID = %q, want %q", writer.runID, runID)
	}
}

func TestRunCancelAPIError(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer
	root := &Root{
		stdout: &bytes.Buffer{},
		stderr: &stderr,
		clientFactory: func() (cursor.Client, error) {
			return newStubClientWithCancelWriter(&stubCancelWriter{
				err: &cursor.APIError{StatusCode: 409, Body: "run_not_cancellable"},
			}), nil
		},
	}

	if got := root.Run([]string{"cancel", "bc-1", "run-1"}); got != ExitAPI {
		t.Fatalf("Run(cancel) = %d, want %d", got, ExitAPI)
	}
	if !strings.Contains(stderr.String(), "status=409") {
		t.Fatalf("stderr = %q, want API error message", stderr.String())
	}
}

func TestRunCancelHelp(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer
	root := &Root{stdout: &bytes.Buffer{}, stderr: &stderr}

	if got := root.Run([]string{"cancel", "--help"}); got != ExitSuccess {
		t.Fatalf("Run(cancel --help) = %d, want %d", got, ExitSuccess)
	}
	if !strings.Contains(stderr.String(), "Usage: cursor-agent-cli cancel") {
		t.Fatalf("stderr = %q, want cancel usage text", stderr.String())
	}
}

func TestRunCancelMissingAgentID(t *testing.T) {
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

	if got := root.Run([]string{"cancel"}); got != ExitUsage {
		t.Fatalf("Run(cancel) = %d, want %d", got, ExitUsage)
	}
	if !strings.Contains(stderr.String(), "agent_id is required") {
		t.Fatalf("stderr = %q, want missing agent_id message", stderr.String())
	}
}

func TestRunCancelMissingRunID(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer
	root := &Root{
		stdout: &bytes.Buffer{},
		stderr: &stderr,
		clientFactory: func() (cursor.Client, error) {
			t.Fatal("client should not be called for missing run_id")
			return nil, nil
		},
	}

	if got := root.Run([]string{"cancel", "bc-1"}); got != ExitUsage {
		t.Fatalf("Run(cancel) = %d, want %d", got, ExitUsage)
	}
	if !strings.Contains(stderr.String(), "run_id is required") {
		t.Fatalf("stderr = %q, want missing run_id message", stderr.String())
	}
}

func TestRunCancelMissingAPIKey(t *testing.T) {
	t.Setenv("CURSOR_CLOUD_AGENT_API_KEY", "")

	var stderr bytes.Buffer
	root := NewRoot()
	root.stdout = &bytes.Buffer{}
	root.stderr = &stderr

	if got := root.Run([]string{"cancel", "bc-1", "run-1"}); got != ExitConfig {
		t.Fatalf("Run(cancel) = %d, want %d", got, ExitConfig)
	}
	if !strings.Contains(stderr.String(), "CURSOR_CLOUD_AGENT_API_KEY") {
		t.Fatalf("stderr = %q, want missing API key message", stderr.String())
	}
}

func TestRunStreamSuccess(t *testing.T) {
	t.Parallel()

	agentID := "bc-00000000-0000-0000-0000-000000000001"
	runID := "run-00000000-0000-0000-0000-000000000001"
	reader := &stubStreamReader{
		stream: &stubSSEStream{
			events: []cursor.SSEEvent{
				{Event: "status", Data: `{"runId":"run-1","status":"RUNNING"}`},
				{Event: "assistant", Data: `{"text":"hello"}`, ID: "1713033006000-0"},
				{Event: "done", Data: `{}`},
			},
		},
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	root := &Root{
		stdout: &stdout,
		stderr: &stderr,
		clientFactory: func() (cursor.Client, error) {
			return newStubClientWithStreamReader(reader), nil
		},
	}

	if got := root.Run([]string{"stream", agentID, runID}); got != ExitSuccess {
		t.Fatalf("Run(stream) = %d, want %d", got, ExitSuccess)
	}
	if reader.agentID != agentID {
		t.Fatalf("agentID = %q, want %q", reader.agentID, agentID)
	}
	if reader.runID != runID {
		t.Fatalf("runID = %q, want %q", reader.runID, runID)
	}

	lines := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	if len(lines) != 3 {
		t.Fatalf("got %d lines, want 3: %q", len(lines), stdout.String())
	}

	var first map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &first); err != nil {
		t.Fatalf("json.Unmarshal line 0: %v", err)
	}
	if first["event"] != "status" {
		t.Fatalf("line 0 event = %v, want status", first["event"])
	}
}

func TestRunStreamErrorEvent(t *testing.T) {
	t.Parallel()

	reader := &stubStreamReader{
		stream: &stubSSEStream{
			events: []cursor.SSEEvent{
				{Event: "error", Data: `{"code":"run_not_found","message":"not found"}`},
			},
		},
	}
	var stdout bytes.Buffer
	root := &Root{
		stdout: &stdout,
		stderr: &bytes.Buffer{},
		clientFactory: func() (cursor.Client, error) {
			return newStubClientWithStreamReader(reader), nil
		},
	}

	if got := root.Run([]string{"stream", "bc-1", "run-1"}); got != ExitAPI {
		t.Fatalf("Run(stream error) = %d, want %d", got, ExitAPI)
	}
}

func TestRunStreamUnexpectedEOF(t *testing.T) {
	t.Parallel()

	reader := &stubStreamReader{
		stream: &stubSSEStream{
			events: []cursor.SSEEvent{
				{Event: "status", Data: `{"runId":"run-1","status":"RUNNING"}`},
				{Event: "assistant", Data: `{"text":"hello"}`},
			},
		},
	}
	var stderr bytes.Buffer
	root := &Root{
		stdout: &bytes.Buffer{},
		stderr: &stderr,
		clientFactory: func() (cursor.Client, error) {
			return newStubClientWithStreamReader(reader), nil
		},
	}

	if got := root.Run([]string{"stream", "bc-1", "run-1"}); got != ExitAPI {
		t.Fatalf("Run(stream unexpected EOF) = %d, want %d", got, ExitAPI)
	}
	if !strings.Contains(stderr.String(), "stream ended unexpectedly") {
		t.Fatalf("stderr = %q, want unexpected EOF message", stderr.String())
	}
}

func TestRunStreamConnectionError(t *testing.T) {
	t.Parallel()

	reader := &stubStreamReader{
		err: &cursor.APIError{StatusCode: 404, Body: "not found"},
	}
	var stderr bytes.Buffer
	root := &Root{
		stdout: &bytes.Buffer{},
		stderr: &stderr,
		clientFactory: func() (cursor.Client, error) {
			return newStubClientWithStreamReader(reader), nil
		},
	}

	if got := root.Run([]string{"stream", "bc-1", "run-1"}); got != ExitAPI {
		t.Fatalf("Run(stream conn error) = %d, want %d", got, ExitAPI)
	}
	if !strings.Contains(stderr.String(), "stream connection failed") {
		t.Fatalf("stderr = %q, want connection error message", stderr.String())
	}
}

func TestRunStreamHelp(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer
	root := &Root{stdout: &bytes.Buffer{}, stderr: &stderr}

	if got := root.Run([]string{"stream", "--help"}); got != ExitSuccess {
		t.Fatalf("Run(stream --help) = %d, want %d", got, ExitSuccess)
	}
	if !strings.Contains(stderr.String(), "Usage: cursor-agent-cli stream") {
		t.Fatalf("stderr = %q, want stream usage text", stderr.String())
	}
}

func TestRunStreamMissingAgentID(t *testing.T) {
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

	if got := root.Run([]string{"stream"}); got != ExitUsage {
		t.Fatalf("Run(stream) = %d, want %d", got, ExitUsage)
	}
	if !strings.Contains(stderr.String(), "agent_id is required") {
		t.Fatalf("stderr = %q, want missing agent_id message", stderr.String())
	}
}

func TestRunStreamMissingRunID(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer
	root := &Root{
		stdout: &bytes.Buffer{},
		stderr: &stderr,
		clientFactory: func() (cursor.Client, error) {
			t.Fatal("client should not be called for missing run_id")
			return nil, nil
		},
	}

	if got := root.Run([]string{"stream", "bc-1"}); got != ExitUsage {
		t.Fatalf("Run(stream) = %d, want %d", got, ExitUsage)
	}
	if !strings.Contains(stderr.String(), "run_id is required") {
		t.Fatalf("stderr = %q, want missing run_id message", stderr.String())
	}
}

func TestRunStreamMissingAPIKey(t *testing.T) {
	t.Setenv("CURSOR_CLOUD_AGENT_API_KEY", "")

	var stderr bytes.Buffer
	root := NewRoot()
	root.stdout = &bytes.Buffer{}
	root.stderr = &stderr

	if got := root.Run([]string{"stream", "bc-1", "run-1"}); got != ExitConfig {
		t.Fatalf("Run(stream) = %d, want %d", got, ExitConfig)
	}
	if !strings.Contains(stderr.String(), "CURSOR_CLOUD_AGENT_API_KEY") {
		t.Fatalf("stderr = %q, want missing API key message", stderr.String())
	}
}

func TestRunHelpIncludesNewCommands(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer
	root := &Root{stdout: &bytes.Buffer{}, stderr: &stderr}

	root.Run([]string{"help"})
	if !strings.Contains(stderr.String(), "stream") {
		t.Fatalf("stderr = %q, want 'stream' in help output", stderr.String())
	}
	if !strings.Contains(stderr.String(), "cancel") {
		t.Fatalf("stderr = %q, want 'cancel' in help output", stderr.String())
	}
}
