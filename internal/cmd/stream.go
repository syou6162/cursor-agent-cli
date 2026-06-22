package cmd

import (
	"context"
	"encoding/json"
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
func streamRun(ctx context.Context, client cursor.Client, agentID, runID string, w io.Writer) int {
	stream, err := client.StreamRun(ctx, agentID, runID)
	if err != nil {
		return ExitAPI
	}
	defer stream.Close()

	enc := json.NewEncoder(w)
	for {
		evt, err := stream.Next()
		if err != nil {
			if err == io.EOF {
				return ExitSuccess
			}
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

		_ = enc.Encode(out)

		if evt.Event == "done" || evt.Event == "error" {
			if evt.Event == "error" {
				return ExitAPI
			}
			return ExitSuccess
		}
	}
}
