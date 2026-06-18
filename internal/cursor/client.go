package cursor

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	defaultAPIBaseURL   = "https://api.cursor.com/v1"
	defaultTimeoutSecs  = 30
	maxErrorBodyBytes   = 1 << 20
	maxErrorBodyDisplay = 500

	envAPIKey  = "CURSOR_CLOUD_AGENT_API_KEY"
	envBaseURL = "CURSOR_API_BASE_URL"
)

// ModelReader groups read operations for models.
type ModelReader interface {
	ListModels(ctx context.Context) (*ListModelsResponse, error)
}

// Client defines the capabilities needed by CLI commands.
type Client interface {
	ModelReader
}

// Config holds settings for the API client.
type Config struct {
	APIKey     string
	BaseURL    string
	ModelsURL  string
	HTTPClient *http.Client
}

type apiClient struct {
	apiKey     func() string
	httpClient *http.Client
	modelsURL  string
}

// NewClient creates a Client from the given configuration.
func NewClient(cfg Config) Client {
	return newAPIClient(cfg)
}

// ClientFromEnv creates a Client using CURSOR_CLOUD_AGENT_API_KEY and
// CURSOR_API_BASE_URL environment variables.
func ClientFromEnv() (Client, error) {
	apiKey := strings.TrimSpace(os.Getenv(envAPIKey))
	if apiKey == "" {
		return nil, fmt.Errorf("%s is not set", envAPIKey)
	}
	return NewClient(Config{
		APIKey:  apiKey,
		BaseURL: os.Getenv(envBaseURL),
	}), nil
}

func newAPIClient(cfg Config) *apiClient {
	baseURL := strings.TrimSpace(cfg.BaseURL)
	if baseURL == "" {
		baseURL = defaultAPIBaseURL
	}
	modelsURL := strings.TrimSpace(cfg.ModelsURL)
	if modelsURL == "" {
		modelsURL = strings.TrimRight(baseURL, "/") + "/models"
	}
	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: defaultTimeoutSecs * time.Second}
	}

	apiKey := cfg.APIKey
	return &apiClient{
		apiKey:     func() string { return strings.TrimSpace(apiKey) },
		httpClient: httpClient,
		modelsURL:  modelsURL,
	}
}

var _ ModelReader = (*apiClient)(nil)
var _ Client = (*apiClient)(nil)

func (c *apiClient) ListModels(ctx context.Context) (*ListModelsResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.modelsURL, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.apiKey(), "")

	resp, err := c.sendAndParseAPIError(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data ListModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("Cursor API response parse failed: %w", err)
	}
	return &data, nil
}

func (c *apiClient) sendAndParseAPIError(req *http.Request) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Cursor API request failed: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		body, readErr := io.ReadAll(io.LimitReader(resp.Body, maxErrorBodyBytes))
		truncated := truncateBody(string(body), maxErrorBodyDisplay)
		if readErr != nil {
			if truncated == "" {
				return nil, fmt.Errorf("Cursor API error (status=%d): %w", resp.StatusCode, readErr)
			}
			return nil, fmt.Errorf("Cursor API error (status=%d): %s (body read failed: %w)", resp.StatusCode, truncated, readErr)
		}
		return nil, fmt.Errorf("Cursor API error (status=%d): %s", resp.StatusCode, truncated)
	}
	return resp, nil
}

func truncateBody(body string, maxLen int) string {
	if len(body) <= maxLen {
		return body
	}
	return body[:maxLen] + "..."
}
