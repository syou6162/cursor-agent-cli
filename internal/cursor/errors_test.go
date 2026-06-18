package cursor

import (
	"net/http"
	"testing"
)

func TestAPIErrorIsRateLimit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		statusCode int
		want       bool
	}{
		{name: "429", statusCode: http.StatusTooManyRequests, want: true},
		{name: "500", statusCode: http.StatusInternalServerError, want: false},
		{name: "404", statusCode: http.StatusNotFound, want: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := &APIError{StatusCode: tt.statusCode, Body: "error"}
			if got := err.IsRateLimit(); got != tt.want {
				t.Fatalf("IsRateLimit() = %v, want %v", got, tt.want)
			}
		})
	}
}
