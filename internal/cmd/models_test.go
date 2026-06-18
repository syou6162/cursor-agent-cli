package cmd

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/syou6162/cursor-agent-cli/internal/cursor"
)

type stubModelReader struct {
	response *cursor.ListModelsResponse
	err      error
}

func (s stubModelReader) ListModels(_ context.Context) (*cursor.ListModelsResponse, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.response, nil
}

func TestListModelsSuccess(t *testing.T) {
	t.Parallel()

	want := &cursor.ListModelsResponse{
		Items: []cursor.Model{
			{ID: "composer-2", DisplayName: "Composer 2"},
		},
	}
	client := newStubClient(stubModelReader{response: want}, nil)

	got, err := listModels(context.Background(), client)
	if err != nil {
		t.Fatalf("listModels() error = %v", err)
	}
	if got.Items[0].ID != want.Items[0].ID {
		t.Fatalf("listModels() = %+v, want %+v", got, want)
	}
}

func TestListModelsAPIError(t *testing.T) {
	t.Parallel()

	client := newStubClient(stubModelReader{
		err: errors.New("Cursor API error (status=401): unauthorized"),
	}, nil)

	_, err := listModels(context.Background(), client)
	if err == nil {
		t.Fatal("listModels() error = nil, want API error")
	}
	if !strings.Contains(err.Error(), "status=401") {
		t.Fatalf("error = %q, want status=401", err.Error())
	}
}
