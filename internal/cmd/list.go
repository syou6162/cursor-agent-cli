package cmd

import (
	"context"

	"github.com/syou6162/cursor-agent-cli/internal/cursor"
)

// listAgents fetches agents from the Cursor Cloud Agent API.
func listAgents(ctx context.Context, client cursor.Client, limit int) (*cursor.ListAgentsResponse, error) {
	return client.ListAgents(ctx, limit)
}
