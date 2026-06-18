package cmd

import (
	"context"
	"errors"
	"strings"
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
	client := newStubClient(nil, reader)

	got, err := listAgents(context.Background(), client, 20)
	if err != nil {
		t.Fatalf("listAgents() error = %v", err)
	}
	if reader.limit != 20 {
		t.Fatalf("limit = %d, want 20", reader.limit)
	}
	if got.Items[0].ID != want.Items[0].ID {
		t.Fatalf("listAgents() = %+v, want %+v", got, want)
	}
}

func TestListAgentsAPIError(t *testing.T) {
	t.Parallel()

	client := newStubClient(nil, &stubAgentReader{
		err: errors.New("Cursor API error (status=401): unauthorized"),
	})

	_, err := listAgents(context.Background(), client, 20)
	if err == nil {
		t.Fatal("listAgents() error = nil, want API error")
	}
	if !strings.Contains(err.Error(), "status=401") {
		t.Fatalf("error = %q, want status=401", err.Error())
	}
}
