package cmd

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/syou6162/cursor-agent-cli/internal/cursor"
)

type spyCursorClient struct {
	response *cursor.ListModelsResponse
	err      error
}

func (s *spyCursorClient) ListModels(_ context.Context) (*cursor.ListModelsResponse, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.response, nil
}

func TestModelsRunSuccess(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	spy := &spyCursorClient{
		response: &cursor.ListModelsResponse{
			Items: []cursor.Model{
				{ID: "composer-2", DisplayName: "Composer 2"},
			},
		},
	}

	cmd := NewModelsWithClient(&stdout, &stderr, spy)
	if got := cmd.Run(nil); got != ExitSuccess {
		t.Fatalf("Run() = %d, want %d", got, ExitSuccess)
	}
	if !strings.Contains(stdout.String(), "composer-2") {
		t.Fatalf("stdout = %q, want composer-2", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestModelsRunAPIError(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	spy := &spyCursorClient{
		err: errors.New("Cursor API error (status=401): unauthorized"),
	}

	cmd := NewModelsWithClient(&stdout, &stderr, spy)
	if got := cmd.Run(nil); got != ExitAPI {
		t.Fatalf("Run() = %d, want %d", got, ExitAPI)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if !strings.Contains(stderr.String(), "status=401") {
		t.Fatalf("stderr = %q, want API error message", stderr.String())
	}
}

func TestModelsRunMissingAPIKey(t *testing.T) {
	t.Setenv("CURSOR_CLOUD_AGENT_API_KEY", "")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := NewModels(&stdout, &stderr)

	if got := cmd.Run(nil); got != ExitConfig {
		t.Fatalf("Run() = %d, want %d", got, ExitConfig)
	}
	if !strings.Contains(stderr.String(), "CURSOR_CLOUD_AGENT_API_KEY") {
		t.Fatalf("stderr = %q, want missing API key message", stderr.String())
	}
}
