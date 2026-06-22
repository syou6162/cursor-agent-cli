package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/syou6162/cursor-agent-cli/internal/cursor"
)

// streamEvent is the NDJSON output format for each SSE event.
type streamEvent struct {
	Event string          `json:"event"`
	Data  json.RawMessage `json:"data"`
	ID    string          `json:"id,omitempty"`
}

// streamRun connects to the SSE stream and writes NDJSON to w.
// It returns the exit code.
func streamRun(ctx context.Context, client cursor.Client, agentID, runID string, w io.Writer, stderr io.Writer) int {
	stream, err := client.StreamRun(ctx, agentID, runID)
	if err != nil {
		fmt.Fprintf(stderr, "Error: stream connection failed: %v\n", err)
		return ExitAPI
	}
	defer stream.Close()

	enc := json.NewEncoder(w)
	sawTerminal := false
	for {
		evt, err := stream.Next()
		if err != nil {
			if err == io.EOF {
				if sawTerminal {
					return ExitSuccess
				}
				fmt.Fprintln(stderr, "Error: stream ended unexpectedly without a terminal event")
				return ExitAPI
			}
			fmt.Fprintf(stderr, "Error: stream read failed: %v\n", err)
			return ExitAPI
		}

		out := streamEvent{
			Event: evt.Event,
			ID:    evt.ID,
		}

		if json.Valid([]byte(evt.Data)) {
			out.Data = json.RawMessage(evt.Data)
		} else {
			escaped, _ := json.Marshal(evt.Data)
			out.Data = json.RawMessage(escaped)
		}

		if err := enc.Encode(out); err != nil {
			fmt.Fprintf(stderr, "Error: write failed: %v\n", err)
			return ExitError
		}

		switch evt.Event {
		case "done":
			return ExitSuccess
		case "error":
			return ExitAPI
		case "result":
			sawTerminal = true
		}
	}
}
