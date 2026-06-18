package cmd

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/syou6162/cursor-agent-cli/internal/cursor"
)

const (
	cliStateSuccess     = "success"
	cliStateTimeout     = "timeout"
	cliStateAPIError    = "api_error"
	cliStateUsageError  = "usage_error"
	cliStateConfigError = "config_error"
)

// CLIStatusInfo holds CLI-specific metadata for the status subcommand.
type CLIStatusInfo struct {
	State          string  `json:"state"`
	ExitCode       int     `json:"exitCode"`
	PollingCount   int     `json:"pollingCount"`
	ElapsedSeconds int     `json:"elapsedSeconds"`
	Error          *string `json:"error,omitempty"`
}

// StatusResponse is the agent-friendly JSON output for the status subcommand.
type StatusResponse struct {
	Run     *cursor.RunStatusResponse
	AgentID string
	RunID   string
	CLI     CLIStatusInfo
	mode    statusResponseMode
}

type statusResponseMode int

const (
	statusModeCLIOnly statusResponseMode = iota
	statusModeAPIError
	statusModeFull
)

func cliErrorMessage(err error) *string {
	if err == nil {
		return nil
	}
	msg := err.Error()
	return &msg
}

func newCLIOnlyStatus(state string, exitCode int, err error) StatusResponse {
	return StatusResponse{
		mode: statusModeCLIOnly,
		CLI: CLIStatusInfo{
			State:          state,
			ExitCode:       exitCode,
			PollingCount:   0,
			ElapsedSeconds: 0,
			Error:          cliErrorMessage(err),
		},
	}
}

func newAPIErrorStatus(agentID, runID string, pollingCount, elapsedSeconds int, err error) StatusResponse {
	return StatusResponse{
		mode:    statusModeAPIError,
		AgentID: agentID,
		RunID:   runID,
		CLI: CLIStatusInfo{
			State:          cliStateAPIError,
			ExitCode:       ExitAPI,
			PollingCount:   pollingCount,
			ElapsedSeconds: elapsedSeconds,
			Error:          cliErrorMessage(err),
		},
	}
}

func newSuccessStatus(run *cursor.RunStatusResponse, pollingCount, elapsedSeconds int) StatusResponse {
	return StatusResponse{
		mode: statusModeFull,
		Run:  run,
		CLI: CLIStatusInfo{
			State:          cliStateSuccess,
			ExitCode:       ExitSuccess,
			PollingCount:   pollingCount,
			ElapsedSeconds: elapsedSeconds,
			Error:          nil,
		},
	}
}

func newTimeoutStatus(run *cursor.RunStatusResponse, pollingCount, elapsedSeconds int, err error) StatusResponse {
	return StatusResponse{
		mode: statusModeFull,
		Run:  run,
		CLI: CLIStatusInfo{
			State:          cliStateTimeout,
			ExitCode:       ExitError,
			PollingCount:   pollingCount,
			ElapsedSeconds: elapsedSeconds,
			Error:          cliErrorMessage(err),
		},
	}
}

func isTimeoutError(err error) bool {
	return errors.Is(err, errTimeout)
}

func statusAPIError(err error) error {
	var apiErr *cursor.APIError
	if errors.As(err, &apiErr) && apiErr.IsRateLimit() {
		return fmt.Errorf("rate limit exceeded: please wait before retrying")
	}
	return err
}

func statusResponseFromOutcome(agentID, runID string, outcome statusOutcome) StatusResponse {
	if outcome.err != nil {
		var apiErr *cursor.APIError
		if errors.As(outcome.err, &apiErr) {
			return newAPIErrorStatus(agentID, runID, outcome.pollingCount, outcome.elapsedSeconds, statusAPIError(outcome.err))
		}
		if isTimeoutError(outcome.err) {
			return newTimeoutStatus(outcome.response, outcome.pollingCount, outcome.elapsedSeconds, outcome.err)
		}
		return newAPIErrorStatus(agentID, runID, outcome.pollingCount, outcome.elapsedSeconds, outcome.err)
	}
	return newSuccessStatus(outcome.response, outcome.pollingCount, outcome.elapsedSeconds)
}

// MarshalJSON implements custom JSON output for agent-friendly status responses.
func (s StatusResponse) MarshalJSON() ([]byte, error) {
	switch s.mode {
	case statusModeCLIOnly:
		return json.Marshal(struct {
			CLI CLIStatusInfo `json:"_cli"`
		}{CLI: s.CLI})
	case statusModeAPIError:
		return json.Marshal(struct {
			ID      string        `json:"id"`
			AgentID string        `json:"agentId"`
			CLI     CLIStatusInfo `json:"_cli"`
		}{
			ID:      s.RunID,
			AgentID: s.AgentID,
			CLI:     s.CLI,
		})
	default:
		if s.Run == nil {
			return json.Marshal(struct {
				CLI CLIStatusInfo `json:"_cli"`
			}{CLI: s.CLI})
		}
		return json.Marshal(struct {
			cursor.RunStatusResponse
			CLI CLIStatusInfo `json:"_cli"`
		}{
			RunStatusResponse: *s.Run,
			CLI:               s.CLI,
		})
	}
}
