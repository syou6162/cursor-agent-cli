package cmd

import (
	"context"
	"errors"
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

var errTimeout = errors.New("timeout waiting for run to complete")

func isTerminalRunStatus(status string) bool {
	_, ok := terminalRunStatuses[status]
	return ok
}

type statusOutcome struct {
	response       *cursor.RunStatusResponse
	pollingCount   int
	elapsedSeconds int
	err            error
}

func elapsedSecondsSince(start time.Time) int {
	elapsed := int(time.Since(start).Seconds())
	if elapsed < 0 {
		return 0
	}
	return elapsed
}

// getRunStatus fetches the current status of an agent run.
func getRunStatus(ctx context.Context, client cursor.Client, agentID, runID string) statusOutcome {
	start := time.Now()
	resp, err := client.GetRunStatus(ctx, agentID, runID)
	return statusOutcome{
		response:       resp,
		pollingCount:   1,
		elapsedSeconds: elapsedSecondsSince(start),
		err:            err,
	}
}

// waitForRunStatus polls until the run reaches a terminal status or the timeout elapses.
func waitForRunStatus(ctx context.Context, client cursor.Client, agentID, runID string, interval, timeout time.Duration) statusOutcome {
	start := time.Now()
	pollingCount := 0
	var lastStatus *cursor.RunStatusResponse

	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	for {
		pollingCount++
		status, err := client.GetRunStatus(ctx, agentID, runID)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				return statusOutcome{
					response:       lastStatus,
					pollingCount:   pollingCount,
					elapsedSeconds: elapsedSecondsSince(start),
					err:            errTimeout,
				}
			}
			return statusOutcome{
				response:       lastStatus,
				pollingCount:   pollingCount,
				elapsedSeconds: elapsedSecondsSince(start),
				err:            err,
			}
		}
		lastStatus = status
		if isTerminalRunStatus(status.Status) {
			return statusOutcome{
				response:       status,
				pollingCount:   pollingCount,
				elapsedSeconds: elapsedSecondsSince(start),
			}
		}

		sleep := interval
		if deadline, ok := ctx.Deadline(); ok {
			remaining := time.Until(deadline)
			if remaining <= 0 {
				return statusOutcome{
					response:       lastStatus,
					pollingCount:   pollingCount,
					elapsedSeconds: elapsedSecondsSince(start),
					err:            errTimeout,
				}
			}
			if remaining < sleep {
				sleep = remaining
			}
		}

		select {
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				return statusOutcome{
					response:       lastStatus,
					pollingCount:   pollingCount,
					elapsedSeconds: elapsedSecondsSince(start),
					err:            errTimeout,
				}
			}
			return statusOutcome{
				response:       lastStatus,
				pollingCount:   pollingCount,
				elapsedSeconds: elapsedSecondsSince(start),
				err:            ctx.Err(),
			}
		case <-sleepAfter(sleep):
		}
	}
}
