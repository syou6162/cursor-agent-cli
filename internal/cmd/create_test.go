package cmd

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/syou6162/cursor-agent-cli/internal/cursor"
)

type spyCreateClient struct {
	createAgentReq  cursor.CreateAgentRequest
	createAgentResp *cursor.CreateAgentResponse
	err             error
}

type stubAgentWriter struct {
	response *cursor.CreateAgentResponse
	err      error
	req      cursor.CreateAgentRequest
}

func (s *stubAgentWriter) CreateAgent(_ context.Context, req cursor.CreateAgentRequest) (*cursor.CreateAgentResponse, error) {
	s.req = req
	if s.err != nil {
		return nil, s.err
	}
	return s.response, nil
}

func (s *spyCreateClient) ListModels(context.Context) (*cursor.ListModelsResponse, error) {
	return nil, nil
}

func (s *spyCreateClient) ListAgents(context.Context, int) (*cursor.ListAgentsResponse, error) {
	return nil, nil
}

func (s *spyCreateClient) CreateAgent(_ context.Context, req cursor.CreateAgentRequest) (*cursor.CreateAgentResponse, error) {
	s.createAgentReq = req
	if s.err != nil {
		return nil, s.err
	}
	return s.createAgentResp, nil
}

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
	spy := &spyCreateClient{createAgentResp: want}
	req := cursor.CreateAgentRequest{
		Prompt: cursor.AgentPrompt{Text: "Add README"},
		Repos: []cursor.AgentRepo{
			{URL: "https://github.com/org/repo", StartingRef: &branch},
		},
		AutoCreatePR: &autoCreatePR,
	}

	got, err := createAgent(context.Background(), spy, req)
	if err != nil {
		t.Fatalf("createAgent() error = %v", err)
	}
	if !reflect.DeepEqual(spy.createAgentReq, req) {
		t.Fatalf("CreateAgent request = %+v, want %+v", spy.createAgentReq, req)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("createAgent() = %+v, want %+v", got, want)
	}
}

func TestCreateAgentAPIError(t *testing.T) {
	t.Parallel()

	spy := &spyCreateClient{
		err: &cursor.APIError{StatusCode: 500, Body: "internal error"},
	}
	req := cursor.CreateAgentRequest{
		Prompt: cursor.AgentPrompt{Text: "Add README"},
	}

	_, err := createAgent(context.Background(), spy, req)
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
