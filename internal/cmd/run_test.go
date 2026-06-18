package cmd

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/syou6162/cursor-agent-cli/internal/cursor"
)

func TestCreateRunSuccess(t *testing.T) {
	t.Parallel()

	agentID := "bc-00000000-0000-0000-0000-000000000001"
	want := &cursor.CreateRunResponse{
		Run: cursor.Run{
			ID:      "run-00000000-0000-0000-0000-000000000002",
			AgentID: agentID,
			Status:  "CREATING",
		},
	}
	writer := &stubRunWriter{response: want}
	req := cursor.CreateRunRequest{
		Prompt: cursor.AgentPrompt{Text: "Fix the failing test"},
	}

	got, err := createRun(context.Background(), newStubClientWithRunWriter(writer), agentID, req)
	if err != nil {
		t.Fatalf("createRun() error = %v", err)
	}
	if writer.agentID != agentID {
		t.Fatalf("CreateRun agentID = %q, want %q", writer.agentID, agentID)
	}
	if !reflect.DeepEqual(writer.req, req) {
		t.Fatalf("CreateRun request = %+v, want %+v", writer.req, req)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("createRun() = %+v, want %+v", got, want)
	}
}

func TestCreateRunAgentBusy(t *testing.T) {
	t.Parallel()

	writer := &stubRunWriter{err: cursor.ErrAgentBusy}
	req := cursor.CreateRunRequest{
		Prompt: cursor.AgentPrompt{Text: "Fix the failing test"},
	}

	_, err := createRun(context.Background(), newStubClientWithRunWriter(writer), "bc-00000000-0000-0000-0000-000000000001", req)
	if err == nil {
		t.Fatal("createRun() error = nil, want agent busy error")
	}
	if !errors.Is(err, cursor.ErrAgentBusy) {
		t.Fatalf("createRun() error = %v, want ErrAgentBusy", err)
	}
}

func TestCreateRunAPIError(t *testing.T) {
	t.Parallel()

	writer := &stubRunWriter{
		err: &cursor.APIError{StatusCode: 500, Body: "internal error"},
	}
	req := cursor.CreateRunRequest{
		Prompt: cursor.AgentPrompt{Text: "Fix the failing test"},
	}

	_, err := createRun(context.Background(), newStubClientWithRunWriter(writer), "bc-00000000-0000-0000-0000-000000000001", req)
	if err == nil {
		t.Fatal("createRun() error = nil, want API error")
	}

	var apiErr *cursor.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("createRun() error = %T, want *cursor.APIError", err)
	}
	if apiErr.StatusCode != 500 {
		t.Fatalf("status = %d, want 500", apiErr.StatusCode)
	}
}
