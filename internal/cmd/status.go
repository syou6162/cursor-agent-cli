package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/syou6162/cursor-agent-cli/internal/cursor"
)

var terminalRunStatuses = map[string]struct{}{
	"FINISHED":  {},
	"ERROR":     {},
	"CANCELLED": {},
	"EXPIRED":   {},
}

var sleepAfter = time.After

func isTerminalRunStatus(status string) bool {
	_, ok := terminalRunStatuses[status]
	return ok
}

// getRunStatus fetches the current status of an agent run.
func getRunStatus(ctx context.Context, client cursor.Client, agentID, runID string) (*cursor.RunStatusResponse, error) {
	return client.GetRunStatus(ctx, agentID, runID)
}

// waitForRunStatus polls until the run reaches a terminal status or the timeout elapses.
func waitForRunStatus(ctx context.Context, client cursor.Client, agentID, runID string, interval, timeout time.Duration) (*cursor.RunStatusResponse, error) {
	deadline := time.Time{}
	if timeout > 0 {
		deadline = time.Now().Add(timeout)
	}

	for {
		status, err := client.GetRunStatus(ctx, agentID, runID)
		if err != nil {
			return nil, err
		}
		if isTerminalRunStatus(status.Status) {
			return status, nil
		}
		if !deadline.IsZero() && time.Now().After(deadline) {
			return nil, fmt.Errorf("timeout waiting for run to complete")
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-sleepAfter(interval):
		}
	}
}
