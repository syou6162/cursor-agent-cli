package cmd

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/syou6162/cursor-agent-cli/internal/cursor"
)

type spyRunClient struct {
	createRunAgentID string
	createRunReq     cursor.CreateRunRequest
	createRunResp    *cursor.CreateRunResponse
	err              error
}

func (s *spyRunClient) ListModels(context.Context) (*cursor.ListModelsResponse, error) {
	return nil, nil
}

func (s *spyRunClient) ListAgents(context.Context, int) (*cursor.ListAgentsResponse, error) {
	return nil, nil
}

func (s *spyRunClient) CreateAgent(context.Context, cursor.CreateAgentRequest) (*cursor.CreateAgentResponse, error) {
	return nil, nil
}

func (s *spyRunClient) CreateRun(_ context.Context, agentID string, req cursor.CreateRunRequest) (*cursor.CreateRunResponse, error) {
	s.createRunAgentID = agentID
	s.createRunReq = req
	if s.err != nil {
		return nil, s.err
	}
	return s.createRunResp, nil
}

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
	spy := &spyRunClient{createRunResp: want}
	req := cursor.CreateRunRequest{
		Prompt: cursor.AgentPrompt{Text: "Fix the failing test"},
	}

	got, err := createRun(context.Background(), spy, agentID, req)
	if err != nil {
		t.Fatalf("createRun() error = %v", err)
	}
	if spy.createRunAgentID != agentID {
		t.Fatalf("CreateRun agentID = %q, want %q", spy.createRunAgentID, agentID)
	}
	if !reflect.DeepEqual(spy.createRunReq, req) {
		t.Fatalf("CreateRun request = %+v, want %+v", spy.createRunReq, req)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("createRun() = %+v, want %+v", got, want)
	}
}

func TestCreateRunAgentBusy(t *testing.T) {
	t.Parallel()

	spy := &spyRunClient{err: cursor.ErrAgentBusy}
	req := cursor.CreateRunRequest{
		Prompt: cursor.AgentPrompt{Text: "Fix the failing test"},
	}

	_, err := createRun(context.Background(), spy, "bc-00000000-0000-0000-0000-000000000001", req)
	if err == nil {
		t.Fatal("createRun() error = nil, want agent busy error")
	}
	if !errors.Is(err, cursor.ErrAgentBusy) {
		t.Fatalf("createRun() error = %v, want ErrAgentBusy", err)
	}
}

func TestCreateRunAPIError(t *testing.T) {
	t.Parallel()

	spy := &spyRunClient{
		err: &cursor.APIError{StatusCode: 500, Body: "internal error"},
	}
	req := cursor.CreateRunRequest{
		Prompt: cursor.AgentPrompt{Text: "Fix the failing test"},
	}

	_, err := createRun(context.Background(), spy, "bc-00000000-0000-0000-0000-000000000001", req)
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
