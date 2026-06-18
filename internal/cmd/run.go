package cmd

import (
	"context"

	"github.com/syou6162/cursor-agent-cli/internal/cursor"
)

// createRun starts a new run on an existing Cloud Agent.
func createRun(ctx context.Context, client cursor.Client, agentID string, req cursor.CreateRunRequest) (*cursor.CreateRunResponse, error) {
	return client.CreateRun(ctx, agentID, req)
}
