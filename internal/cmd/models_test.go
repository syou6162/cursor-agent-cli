package cmd

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/syou6162/cursor-agent-cli/internal/cursor"
)

type spyCursorClient struct {
	response      *cursor.ListModelsResponse
	err           error
	agentsResponse *cursor.ListAgentsResponse
	agentsErr      error
	agentsLimit    int
}

func (s *spyCursorClient) ListModels(_ context.Context) (*cursor.ListModelsResponse, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.response, nil
}

func (s *spyCursorClient) ListAgents(_ context.Context, limit int) (*cursor.ListAgentsResponse, error) {
	s.agentsLimit = limit
	if s.agentsErr != nil {
		return nil, s.agentsErr
	}
	return s.agentsResponse, nil
}

func TestListModelsSuccess(t *testing.T) {
	t.Parallel()

	want := &cursor.ListModelsResponse{
		Items: []cursor.Model{
			{ID: "composer-2", DisplayName: "Composer 2"},
		},
	}
	spy := &spyCursorClient{response: want}

	got, err := listModels(context.Background(), spy)
	if err != nil {
		t.Fatalf("listModels() error = %v", err)
	}
	if got.Items[0].ID != want.Items[0].ID {
		t.Fatalf("listModels() = %+v, want %+v", got, want)
	}
}

func TestListModelsAPIError(t *testing.T) {
	t.Parallel()

	spy := &spyCursorClient{
		err: errors.New("Cursor API error (status=401): unauthorized"),
	}

	_, err := listModels(context.Background(), spy)
	if err == nil {
		t.Fatal("listModels() error = nil, want API error")
	}
	if !strings.Contains(err.Error(), "status=401") {
		t.Fatalf("error = %q, want status=401", err.Error())
	}
}
