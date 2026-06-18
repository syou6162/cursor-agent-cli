package cmd

import (
	"context"

	"github.com/syou6162/cursor-agent-cli/internal/cursor"
)

// listModels fetches available models from the Cursor Cloud Agent API.
func listModels(ctx context.Context, client cursor.Client) (*cursor.ListModelsResponse, error) {
	return client.ListModels(ctx)
}
