package cmd

import (
	"context"

	"github.com/syou6162/cursor-agent-cli/internal/cursor"
)

// createAgent creates a Cloud Agent via the Cursor Cloud Agent API.
func createAgent(ctx context.Context, client cursor.Client, req cursor.CreateAgentRequest) (*cursor.CreateAgentResponse, error) {
	return client.CreateAgent(ctx, req)
}
