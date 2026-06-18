package cmd

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/syou6162/cursor-agent-cli/internal/cursor"
)

func TestCreateAgentSuccess(t *testing.T) {
	t.Parallel()

	autoCreatePR := true
	branch := "main"
	want := &cursor.CreateAgentResponse{
		Agent: cursor.CreatedAgent{
			ID:     "bc-00000000-0000-0000-0000-000000000001",
			Name:   "Add README",
			Status: "ACTIVE",
			URL:    "https://cursor.com/agents/bc-00000000-0000-0000-0000-000000000001",
		},
		Run: cursor.Run{
			ID:      "run-00000000-0000-0000-0000-000000000001",
			AgentID: "bc-00000000-0000-0000-0000-000000000001",
			Status:  "CREATING",
		},
	}
	writer := &stubAgentWriter{response: want}
	req := cursor.CreateAgentRequest{
		Prompt: cursor.AgentPrompt{Text: "Add README"},
		Repos: []cursor.AgentRepo{
			{URL: "https://github.com/org/repo", StartingRef: &branch},
		},
		AutoCreatePR: &autoCreatePR,
	}

	got, err := createAgent(context.Background(), newStubClientWithAgentWriter(writer), req)
	if err != nil {
		t.Fatalf("createAgent() error = %v", err)
	}
	if !reflect.DeepEqual(writer.req, req) {
		t.Fatalf("CreateAgent request = %+v, want %+v", writer.req, req)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("createAgent() = %+v, want %+v", got, want)
	}
}

func TestCreateAgentAPIError(t *testing.T) {
	t.Parallel()

	writer := &stubAgentWriter{
		err: &cursor.APIError{StatusCode: 500, Body: "internal error"},
	}
	req := cursor.CreateAgentRequest{
		Prompt: cursor.AgentPrompt{Text: "Add README"},
	}

	_, err := createAgent(context.Background(), newStubClientWithAgentWriter(writer), req)
	if err == nil {
		t.Fatal("createAgent() error = nil, want API error")
	}

	var apiErr *cursor.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("createAgent() error = %T, want *cursor.APIError", err)
	}
	if apiErr.StatusCode != 500 {
		t.Fatalf("status = %d, want 500", apiErr.StatusCode)
	}
}
