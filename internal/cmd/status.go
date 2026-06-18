package cmd

import (
	"context"
	"errors"
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
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	for {
		status, err := client.GetRunStatus(ctx, agentID, runID)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				return nil, fmt.Errorf("timeout waiting for run to complete")
			}
			return nil, err
		}
		if isTerminalRunStatus(status.Status) {
			return status, nil
		}

		sleep := interval
		if deadline, ok := ctx.Deadline(); ok {
			remaining := time.Until(deadline)
			if remaining <= 0 {
				return nil, fmt.Errorf("timeout waiting for run to complete")
			}
			if remaining < sleep {
				sleep = remaining
			}
		}

		select {
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				return nil, fmt.Errorf("timeout waiting for run to complete")
			}
			return nil, ctx.Err()
		case <-sleepAfter(sleep):
		}
	}
}
