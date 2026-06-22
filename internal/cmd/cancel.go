package cmd

import (
	"context"

	"github.com/syou6162/cursor-agent-cli/internal/cursor"
)

// cancelRun cancels an active run on a Cloud Agent.
func cancelRun(ctx context.Context, client cursor.Client, agentID, runID string) (*cursor.CancelRunResponse, error) {
	return client.CancelRun(ctx, agentID, runID)
}
