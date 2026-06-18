package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/syou6162/cursor-agent-cli/internal/cursor"
)

func TestGetRunStatusSuccess(t *testing.T) {
	t.Parallel()

	agentID := "bc-00000000-0000-0000-0000-000000000001"
	runID := "run-00000000-0000-0000-0000-000000000001"
	want := &cursor.RunStatusResponse{
		ID:      runID,
		AgentID: agentID,
		Status:  "RUNNING",
	}
	reader := &stubRunReader{responses: []*cursor.RunStatusResponse{want}}

	got := getRunStatus(context.Background(), newStubClientWithRunReader(reader), agentID, runID)
	if got.err != nil {
		t.Fatalf("getRunStatus() error = %v", got.err)
	}
	if reader.agentID != agentID {
		t.Fatalf("agentID = %q, want %q", reader.agentID, agentID)
	}
	if reader.runID != runID {
		t.Fatalf("runID = %q, want %q", reader.runID, runID)
	}
	if got.pollingCount != 1 {
		t.Fatalf("pollingCount = %d, want 1", got.pollingCount)
	}
	if got.response.Status != want.Status {
		t.Fatalf("status = %q, want %q", got.response.Status, want.Status)
	}
}

func TestGetRunStatusAPIError(t *testing.T) {
	t.Parallel()

	reader := &stubRunReader{
		err: &cursor.APIError{StatusCode: 404, Body: "not found"},
	}

	got := getRunStatus(context.Background(), newStubClientWithRunReader(reader), "bc-1", "run-1")
	if got.err == nil {
		t.Fatal("getRunStatus() error = nil, want API error")
	}
	if got.pollingCount != 1 {
		t.Fatalf("pollingCount = %d, want 1", got.pollingCount)
	}

	var apiErr *cursor.APIError
	if !errors.As(got.err, &apiErr) {
		t.Fatalf("getRunStatus() error = %T, want *cursor.APIError", got.err)
	}
	if apiErr.StatusCode != 404 {
		t.Fatalf("status = %d, want 404", apiErr.StatusCode)
	}
}

func TestWaitForRunStatusPollsUntilTerminal(t *testing.T) {
	agentID := "bc-00000000-0000-0000-0000-000000000001"
	runID := "run-00000000-0000-0000-0000-000000000001"
	result := "Added README.md"
	prURL := "https://github.com/org/repo/pull/123"
	reader := &stubRunReader{
		responses: []*cursor.RunStatusResponse{
			{ID: runID, AgentID: agentID, Status: "RUNNING"},
			{ID: runID, AgentID: agentID, Status: "RUNNING"},
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

	origSleepAfter := sleepAfter
	sleepAfter = func(time.Duration) <-chan time.Time {
		ch := make(chan time.Time, 1)
		ch <- time.Now()
		return ch
	}
	t.Cleanup(func() { sleepAfter = origSleepAfter })

	got := waitForRunStatus(context.Background(), newStubClientWithRunReader(reader), agentID, runID, time.Millisecond, 0)
	if got.err != nil {
		t.Fatalf("waitForRunStatus() error = %v", got.err)
	}
	if reader.calls != 3 {
		t.Fatalf("GetRunStatus calls = %d, want 3", reader.calls)
	}
	if got.pollingCount != 3 {
		t.Fatalf("pollingCount = %d, want 3", got.pollingCount)
	}
	if got.response.Status != "FINISHED" {
		t.Fatalf("status = %q, want FINISHED", got.response.Status)
	}
	if got.response.Result == nil || *got.response.Result != result {
		t.Fatalf("result = %v, want %q", got.response.Result, result)
	}
}

func TestWaitForRunStatusTimeout(t *testing.T) {
	agentID := "bc-1"
	runID := "run-1"
	reader := &stubRunReader{
		responses: []*cursor.RunStatusResponse{
			{ID: runID, AgentID: agentID, Status: "RUNNING"},
		},
	}

	origSleepAfter := sleepAfter
	sleepAfter = func(time.Duration) <-chan time.Time {
		ch := make(chan time.Time, 1)
		ch <- time.Now()
		return ch
	}
	t.Cleanup(func() { sleepAfter = origSleepAfter })

	got := waitForRunStatus(context.Background(), newStubClientWithRunReader(reader), agentID, runID, time.Millisecond, time.Millisecond)
	if got.err == nil {
		t.Fatal("waitForRunStatus() error = nil, want timeout error")
	}
	if !isTimeoutError(got.err) {
		t.Fatalf("waitForRunStatus() error = %v, want timeout", got.err)
	}
	if got.response == nil || got.response.Status != "RUNNING" {
		t.Fatalf("last status = %+v, want RUNNING", got.response)
	}
	if got.pollingCount < 1 {
		t.Fatalf("pollingCount = %d, want >= 1", got.pollingCount)
	}
}

func TestIsTerminalRunStatus(t *testing.T) {
	t.Parallel()

	terminal := []string{"FINISHED", "ERROR", "CANCELLED", "EXPIRED"}
	for _, status := range terminal {
		if !isTerminalRunStatus(status) {
			t.Fatalf("isTerminalRunStatus(%q) = false, want true", status)
		}
	}
	if isTerminalRunStatus("RUNNING") {
		t.Fatal("isTerminalRunStatus(RUNNING) = true, want false")
	}
}

func TestStatusResponseMarshalSuccess(t *testing.T) {
	t.Parallel()

	result := "Added README.md"
	resp := newSuccessStatus(&cursor.RunStatusResponse{
		ID:      "run-1",
		AgentID: "bc-1",
		Status:  "FINISHED",
		Result:  &result,
	}, 5, 45)

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if parsed["status"] != "FINISHED" {
		t.Fatalf("status = %v, want FINISHED", parsed["status"])
	}
	cli, ok := parsed["_cli"].(map[string]any)
	if !ok {
		t.Fatal("_cli field missing")
	}
	if cli["state"] != cliStateSuccess {
		t.Fatalf("_cli.state = %v, want %s", cli["state"], cliStateSuccess)
	}
	if cli["exitCode"] != float64(ExitSuccess) {
		t.Fatalf("_cli.exitCode = %v, want %d", cli["exitCode"], ExitSuccess)
	}
	if cli["pollingCount"] != float64(5) {
		t.Fatalf("_cli.pollingCount = %v, want 5", cli["pollingCount"])
	}
}

func TestStatusResponseMarshalTimeout(t *testing.T) {
	t.Parallel()

	timeoutErr := errors.New("timeout waiting for run to complete")
	resp := newTimeoutStatus(&cursor.RunStatusResponse{
		ID:      "run-1",
		AgentID: "bc-1",
		Status:  "RUNNING",
	}, 12, 180, timeoutErr)

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
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
}

func TestStatusResponseMarshalAPIError(t *testing.T) {
	t.Parallel()

	apiErr := &cursor.APIError{StatusCode: 500, Body: "internal error"}
	resp := newAPIErrorStatus("bc-1", "run-1", 0, 0, apiErr)

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if parsed["id"] != "run-1" {
		t.Fatalf("id = %v, want run-1", parsed["id"])
	}
	if parsed["agentId"] != "bc-1" {
		t.Fatalf("agentId = %v, want bc-1", parsed["agentId"])
	}
	if parsed["status"] != nil {
		t.Fatalf("status = %v, want null", parsed["status"])
	}
	cli := parsed["_cli"].(map[string]any)
	if cli["state"] != cliStateAPIError {
		t.Fatalf("_cli.state = %v, want %s", cli["state"], cliStateAPIError)
	}
}

func TestStatusResponseMarshalUsageError(t *testing.T) {
	t.Parallel()

	resp := newCLIOnlyStatus(cliStateUsageError, ExitUsage, errors.New("agent_id is required"))

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if len(parsed) != 1 {
		t.Fatalf("parsed = %+v, want only _cli field", parsed)
	}
	cli := parsed["_cli"].(map[string]any)
	if cli["state"] != cliStateUsageError {
		t.Fatalf("_cli.state = %v, want %s", cli["state"], cliStateUsageError)
	}
}

func TestStatusResponseMarshalConfigError(t *testing.T) {
	t.Parallel()

	resp := newCLIOnlyStatus(cliStateConfigError, ExitConfig, errors.New("CURSOR_CLOUD_AGENT_API_KEY is not set"))

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	cli := parsed["_cli"].(map[string]any)
	if cli["state"] != cliStateConfigError {
		t.Fatalf("_cli.state = %v, want %s", cli["state"], cliStateConfigError)
	}
	if cli["exitCode"] != float64(ExitConfig) {
		t.Fatalf("_cli.exitCode = %v, want %d", cli["exitCode"], ExitConfig)
	}
}
