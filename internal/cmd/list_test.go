package cmd

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/syou6162/cursor-agent-cli/internal/cursor"
)

type stubAgentReader struct {
	response *cursor.ListAgentsResponse
	err      error
	limit    int
}

func (s *stubAgentReader) ListAgents(_ context.Context, limit int) (*cursor.ListAgentsResponse, error) {
	s.limit = limit
	if s.err != nil {
		return nil, s.err
	}
	return s.response, nil
}

func TestListAgentsSuccess(t *testing.T) {
	t.Parallel()

	want := &cursor.ListAgentsResponse{
		Items: []cursor.Agent{
			{
				ID:     "bc-00000000-0000-0000-0000-000000000001",
				Name:   "Add README with setup instructions",
				Status: "ACTIVE",
			},
		},
	}
	reader := &stubAgentReader{response: want}
	client := newStubClientWithAgent(reader)

	got, err := listAgents(context.Background(), client, 20)
	if err != nil {
		t.Fatalf("listAgents() error = %v", err)
	}
	if reader.limit != 20 {
		t.Fatalf("limit = %d, want 20", reader.limit)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("listAgents() = %+v, want %+v", got, want)
	}
}

func TestListAgentsAPIError(t *testing.T) {
	t.Parallel()

	client := newStubClientWithAgent(&stubAgentReader{
		err: &cursor.APIError{StatusCode: 401, Body: "unauthorized"},
	})

	_, err := listAgents(context.Background(), client, 20)
	if err == nil {
		t.Fatal("listAgents() error = nil, want API error")
	}

	var apiErr *cursor.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("listAgents() error = %T, want *cursor.APIError", err)
	}
	if apiErr.StatusCode != 401 {
		t.Fatalf("status = %d, want 401", apiErr.StatusCode)
	}
}
