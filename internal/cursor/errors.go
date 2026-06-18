package cursor

import (
	"errors"
	"fmt"
	"net/http"
)

// ErrAgentBusy is returned when a new run is requested while the agent is still running.
var ErrAgentBusy = errors.New("agent_busy")

// APIError represents a non-2xx response from the Cursor Cloud Agent API.
type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("Cursor API error (status=%d): %s", e.StatusCode, e.Body)
}

// IsRateLimit reports whether the API returned HTTP 429 Too Many Requests.
func (e *APIError) IsRateLimit() bool {
	return e.StatusCode == http.StatusTooManyRequests
}
