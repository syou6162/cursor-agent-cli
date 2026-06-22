package cursor

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
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

// AgentReader groups read operations for agents.
type AgentReader interface {
	ListAgents(ctx context.Context, limit int) (*ListAgentsResponse, error)
}

// AgentWriter groups write operations for agents.
type AgentWriter interface {
	CreateAgent(ctx context.Context, req CreateAgentRequest) (*CreateAgentResponse, error)
}

// RunWriter groups write operations for agent runs.
type RunWriter interface {
	CreateRun(ctx context.Context, agentID string, req CreateRunRequest) (*CreateRunResponse, error)
}

// RunReader groups read operations for agent runs.
type RunReader interface {
	GetRunStatus(ctx context.Context, agentID, runID string) (*RunStatusResponse, error)
}

// Client defines the capabilities needed by CLI commands.
type Client interface {
	ModelReader
	AgentReader
	AgentWriter
	RunWriter
	RunReader
}

// Config holds settings for the API client.
type Config struct {
	APIKey     string
	BaseURL    string
	ModelsURL  string
	AgentsURL  string
	HTTPClient *http.Client
}

type apiClient struct {
	apiKey     func() string
	httpClient *http.Client
	modelsURL  string
	agentsURL  string
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
	agentsURL := strings.TrimSpace(cfg.AgentsURL)
	if agentsURL == "" {
		agentsURL = strings.TrimRight(baseURL, "/") + "/agents"
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
		agentsURL:  agentsURL,
	}
}

var _ ModelReader = (*apiClient)(nil)
var _ AgentReader = (*apiClient)(nil)
var _ AgentWriter = (*apiClient)(nil)
var _ RunWriter = (*apiClient)(nil)
var _ RunReader = (*apiClient)(nil)
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
	defer func() { _ = resp.Body.Close() }()

	var data ListModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("cursor API response parse failed: %w", err)
	}
	return &data, nil
}

func (c *apiClient) ListAgents(ctx context.Context, limit int) (*ListAgentsResponse, error) {
	if limit <= 0 {
		return nil, fmt.Errorf("limit must be greater than 0")
	}

	reqURL, err := url.Parse(c.agentsURL)
	if err != nil {
		return nil, fmt.Errorf("invalid agents URL: %w", err)
	}
	q := reqURL.Query()
	q.Set("limit", strconv.Itoa(limit))
	reqURL.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL.String(), nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.apiKey(), "")

	resp, err := c.sendAndParseAPIError(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	var data ListAgentsResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("cursor API response parse failed: %w", err)
	}
	return &data, nil
}

func (c *apiClient) GetRunStatus(ctx context.Context, agentID, runID string) (*RunStatusResponse, error) {
	runURL := strings.TrimRight(c.agentsURL, "/") + "/" + url.PathEscape(agentID) + "/runs/" + url.PathEscape(runID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, runURL, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.apiKey(), "")

	resp, err := c.sendAndParseAPIError(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	var data RunStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("cursor API response parse failed: %w", err)
	}
	return &data, nil
}

func (c *apiClient) CreateRun(ctx context.Context, agentID string, req CreateRunRequest) (*CreateRunResponse, error) {
	runURL := strings.TrimRight(c.agentsURL, "/") + "/" + url.PathEscape(agentID) + "/runs"
	resp, err := c.postJSON(ctx, runURL, req)
	if err != nil {
		var apiErr *APIError
		if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusConflict {
			return nil, ErrAgentBusy
		}
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	var data CreateRunResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("cursor API response parse failed: %w", err)
	}
	return &data, nil
}

func (c *apiClient) CreateAgent(ctx context.Context, req CreateAgentRequest) (*CreateAgentResponse, error) {
	resp, err := c.postJSON(ctx, c.agentsURL, req)
	if err != nil {
		var apiErr *APIError
		if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusConflict {
			return nil, ErrAgentBusy
		}
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	var data CreateAgentResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("cursor API response parse failed: %w", err)
	}
	return &data, nil
}

func (c *apiClient) postJSON(ctx context.Context, url string, reqBody any) (*http.Response, error) {
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("request encode failed: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.apiKey(), "")

	return c.sendAndParseAPIError(req)
}

func (c *apiClient) sendAndParseAPIError(req *http.Request) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cursor API request failed: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer func() { _ = resp.Body.Close() }()
		body, readErr := io.ReadAll(io.LimitReader(resp.Body, maxErrorBodyBytes))
		truncated := truncateBody(string(body), maxErrorBodyDisplay)
		if readErr != nil {
			return nil, fmt.Errorf("cursor API error (status=%d): body read failed: %w", resp.StatusCode, readErr)
		}
		return nil, &APIError{StatusCode: resp.StatusCode, Body: truncated}
	}
	return resp, nil
}

func truncateBody(body string, maxLen int) string {
	if len(body) <= maxLen {
		return body
	}
	return body[:maxLen] + "..."
}
