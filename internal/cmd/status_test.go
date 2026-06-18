package cmd

import (
	"context"
	"errors"
	"reflect"
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

	got, err := getRunStatus(context.Background(), newStubClientWithRunReader(reader), agentID, runID)
	if err != nil {
		t.Fatalf("getRunStatus() error = %v", err)
	}
	if reader.agentID != agentID {
		t.Fatalf("agentID = %q, want %q", reader.agentID, agentID)
	}
	if reader.runID != runID {
		t.Fatalf("runID = %q, want %q", reader.runID, runID)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("getRunStatus() = %+v, want %+v", got, want)
	}
}

func TestGetRunStatusAPIError(t *testing.T) {
	t.Parallel()

	reader := &stubRunReader{
		err: &cursor.APIError{StatusCode: 404, Body: "not found"},
	}

	_, err := getRunStatus(context.Background(), newStubClientWithRunReader(reader), "bc-1", "run-1")
	if err == nil {
		t.Fatal("getRunStatus() error = nil, want API error")
	}

	var apiErr *cursor.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("getRunStatus() error = %T, want *cursor.APIError", err)
	}
	if apiErr.StatusCode != 404 {
		t.Fatalf("status = %d, want 404", apiErr.StatusCode)
	}
}

func TestWaitForRunStatusPollsUntilTerminal(t *testing.T) {
	t.Parallel()

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

	got, err := waitForRunStatus(context.Background(), newStubClientWithRunReader(reader), agentID, runID, time.Millisecond, 0)
	if err != nil {
		t.Fatalf("waitForRunStatus() error = %v", err)
	}
	if reader.calls != 3 {
		t.Fatalf("GetRunStatus calls = %d, want 3", reader.calls)
	}
	if got.Status != "FINISHED" {
		t.Fatalf("status = %q, want FINISHED", got.Status)
	}
	if got.Result == nil || *got.Result != result {
		t.Fatalf("result = %v, want %q", got.Result, result)
	}
}

func TestWaitForRunStatusTimeout(t *testing.T) {
	t.Parallel()

	reader := &stubRunReader{
		responses: []*cursor.RunStatusResponse{
			{ID: "run-1", Status: "RUNNING"},
		},
	}

	origSleepAfter := sleepAfter
	sleepAfter = func(time.Duration) <-chan time.Time {
		ch := make(chan time.Time, 1)
		ch <- time.Now()
		return ch
	}
	t.Cleanup(func() { sleepAfter = origSleepAfter })

	_, err := waitForRunStatus(context.Background(), newStubClientWithRunReader(reader), "bc-1", "run-1", time.Millisecond, time.Millisecond)
	if err == nil {
		t.Fatal("waitForRunStatus() error = nil, want timeout error")
	}
	if !errors.Is(err, context.DeadlineExceeded) && err.Error() != "timeout waiting for run to complete" {
		t.Fatalf("waitForRunStatus() error = %v, want timeout", err)
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
