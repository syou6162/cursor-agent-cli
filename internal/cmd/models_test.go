package cmd

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/syou6162/cursor-agent-cli/internal/cursor"
)

func TestListModelsSuccess(t *testing.T) {
	t.Parallel()

	want := &cursor.ListModelsResponse{
		Items: []cursor.Model{
			{ID: "composer-2", DisplayName: "Composer 2"},
		},
	}
	client := newStubClientWithModel(stubModelReader{response: want})

	got, err := listModels(context.Background(), client)
	if err != nil {
		t.Fatalf("listModels() error = %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("listModels() = %+v, want %+v", got, want)
	}
}

func TestListModelsAPIError(t *testing.T) {
	t.Parallel()

	client := newStubClientWithModel(stubModelReader{
		err: &cursor.APIError{StatusCode: 401, Body: "unauthorized"},
	})

	_, err := listModels(context.Background(), client)
	if err == nil {
		t.Fatal("listModels() error = nil, want API error")
	}

	var apiErr *cursor.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("listModels() error = %T, want *cursor.APIError", err)
	}
	if apiErr.StatusCode != 401 {
		t.Fatalf("status = %d, want 401", apiErr.StatusCode)
	}
}
