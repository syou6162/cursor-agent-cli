package cursor

import "fmt"

// APIError represents a non-2xx response from the Cursor Cloud Agent API.
type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("Cursor API error (status=%d): %s", e.StatusCode, e.Body)
}
