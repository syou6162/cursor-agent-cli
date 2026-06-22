package cursor

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestListModelsSuccess(t *testing.T) {
	t.Parallel()

	want := ListModelsResponse{
		Items: []Model{
			{ID: "composer-2", DisplayName: "Composer 2"},
		},
	}
	body, err := json.Marshal(want)
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}

	var gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/models" {
			t.Errorf("path = %q, want /models", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("method = %q, want GET", r.Method)
		}
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
	defer server.Close()

	client := NewClient(Config{
		APIKey:    "test-api-key",
		ModelsURL: server.URL + "/models",
	})

	got, err := client.ListModels(context.Background())
	if err != nil {
		t.Fatalf("ListModels() error = %v", err)
	}
	if !reflect.DeepEqual(*got, want) {
		t.Fatalf("ListModels() = %+v, want %+v", got, want)
	}

	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("test-api-key:"))
	if gotAuth != wantAuth {
		t.Fatalf("Authorization = %q, want %q", gotAuth, wantAuth)
	}
}

func TestListModelsTrimsBaseURLWhitespace(t *testing.T) {
	t.Parallel()

	want := ListModelsResponse{Items: []Model{{ID: "m1", DisplayName: "M1"}}}
	body, err := json.Marshal(want)
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
	defer server.Close()

	client := NewClient(Config{
		APIKey:    "test-api-key",
		ModelsURL: " " + server.URL + "/models ",
	})

	got, err := client.ListModels(context.Background())
	if err != nil {
		t.Fatalf("ListModels() error = %v", err)
	}
	if len(got.Items) != 1 {
		t.Fatalf("ListModels() = %+v, want one item", got)
	}
}

func TestListModelsAPIError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	defer server.Close()

	client := NewClient(Config{
		APIKey:    "bad-key",
		ModelsURL: server.URL + "/models",
	})

	_, err := client.ListModels(context.Background())
	if err == nil {
		t.Fatal("ListModels() error = nil, want API error")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error = %T, want *APIError", err)
	}
	if apiErr.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", apiErr.StatusCode)
	}
}

func TestClientFromEnvMissingAPIKey(t *testing.T) {
	t.Setenv(envAPIKey, "")
	t.Setenv(envBaseURL, "")

	_, err := ClientFromEnv()
	if err == nil {
		t.Fatal("ClientFromEnv() error = nil, want missing API key error")
	}
	if !strings.Contains(err.Error(), envAPIKey) {
		t.Fatalf("error = %q, want %q in message", err.Error(), envAPIKey)
	}
}

func TestClientFromEnvSuccess(t *testing.T) {
	t.Setenv(envAPIKey, "  my-key  ")
	t.Setenv(envBaseURL, "")

	client, err := ClientFromEnv()
	if err != nil {
		t.Fatalf("ClientFromEnv() error = %v", err)
	}
	if client == nil {
		t.Fatal("ClientFromEnv() client = nil")
	}
}

type spyClient struct {
	response       *ListModelsResponse
	err            error
	called         bool
	agentsResponse *ListAgentsResponse
	agentsErr      error
	agentsLimit    int
}

func (s *spyClient) ListModels(_ context.Context) (*ListModelsResponse, error) {
	s.called = true
	if s.err != nil {
		return nil, s.err
	}
	return s.response, nil
}

func (s *spyClient) ListAgents(_ context.Context, limit int) (*ListAgentsResponse, error) {
	s.agentsLimit = limit
	if s.agentsErr != nil {
		return nil, s.agentsErr
	}
	return s.agentsResponse, nil
}

func TestSpyClientListModels(t *testing.T) {
	t.Parallel()

	spy := &spyClient{
		response: &ListModelsResponse{
			Items: []Model{{ID: "gpt-5", DisplayName: "GPT-5"}},
		},
	}

	got, err := spy.ListModels(context.Background())
	if err != nil {
		t.Fatalf("ListModels() error = %v", err)
	}
	if !spy.called {
		t.Fatal("spy was not called")
	}
	if len(got.Items) != 1 || got.Items[0].ID != "gpt-5" {
		t.Fatalf("ListModels() = %+v, want gpt-5", got)
	}
}

func TestListAgentsSuccess(t *testing.T) {
	t.Parallel()

	want := ListAgentsResponse{
		Items: []Agent{
			{
				ID:     "bc-00000000-0000-0000-0000-000000000001",
				Name:   "Add README with setup instructions",
				Status: "ACTIVE",
			},
		},
		NextCursor: "bc-00000000-0000-0000-0000-000000000002",
	}
	body, err := json.Marshal(want)
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}

	var gotLimit string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/agents" {
			t.Errorf("path = %q, want /agents", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("method = %q, want GET", r.Method)
		}
		gotLimit = r.URL.Query().Get("limit")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
	defer server.Close()

	client := NewClient(Config{
		APIKey:    "test-api-key",
		AgentsURL: server.URL + "/agents",
	})

	got, err := client.ListAgents(context.Background(), 20)
	if err != nil {
		t.Fatalf("ListAgents() error = %v", err)
	}
	if gotLimit != "20" {
		t.Fatalf("limit query = %q, want 20", gotLimit)
	}
	if !reflect.DeepEqual(*got, want) {
		t.Fatalf("ListAgents() = %+v, want %+v", got, want)
	}
}

func TestListAgentsTrimsAgentsURLWhitespace(t *testing.T) {
	t.Parallel()

	want := ListAgentsResponse{Items: []Agent{{ID: "bc-1", Name: "Agent 1"}}}
	body, err := json.Marshal(want)
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
	defer server.Close()

	client := NewClient(Config{
		APIKey:    "test-api-key",
		AgentsURL: " " + server.URL + "/agents ",
	})

	got, err := client.ListAgents(context.Background(), 10)
	if err != nil {
		t.Fatalf("ListAgents() error = %v", err)
	}
	if len(got.Items) != 1 {
		t.Fatalf("ListAgents() = %+v, want one item", got)
	}
}

func TestListAgentsAPIError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	defer server.Close()

	client := NewClient(Config{
		APIKey:    "bad-key",
		AgentsURL: server.URL + "/agents",
	})

	_, err := client.ListAgents(context.Background(), 20)
	if err == nil {
		t.Fatal("ListAgents() error = nil, want API error")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error = %T, want *APIError", err)
	}
	if apiErr.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", apiErr.StatusCode)
	}
}

func TestListAgentsInvalidLimit(t *testing.T) {
	t.Parallel()

	client := NewClient(Config{APIKey: "test-api-key"})

	_, err := client.ListAgents(context.Background(), 0)
	if err == nil {
		t.Fatal("ListAgents(0) error = nil, want validation error")
	}
	if !strings.Contains(err.Error(), "greater than 0") {
		t.Fatalf("error = %q, want greater than 0", err.Error())
	}
}

func TestListAgentsMergesQueryParams(t *testing.T) {
	t.Parallel()

	want := ListAgentsResponse{Items: []Agent{{ID: "bc-1", Name: "Agent 1"}}}
	body, err := json.Marshal(want)
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}

	var gotLimit string
	var gotIncludeArchived string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotLimit = r.URL.Query().Get("limit")
		gotIncludeArchived = r.URL.Query().Get("includeArchived")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
	defer server.Close()

	client := NewClient(Config{
		APIKey:    "test-api-key",
		AgentsURL: server.URL + "/agents?includeArchived=true",
	})

	got, err := client.ListAgents(context.Background(), 15)
	if err != nil {
		t.Fatalf("ListAgents() error = %v", err)
	}
	if gotLimit != "15" {
		t.Fatalf("limit query = %q, want 15", gotLimit)
	}
	if gotIncludeArchived != "true" {
		t.Fatalf("includeArchived query = %q, want true", gotIncludeArchived)
	}
	if len(got.Items) != 1 {
		t.Fatalf("ListAgents() = %+v, want one item", got)
	}
}

func TestCreateAgentSuccess(t *testing.T) {
	t.Parallel()

	want := CreateAgentResponse{
		Agent: CreatedAgent{
			ID:     "bc-00000000-0000-0000-0000-000000000001",
			Name:   "Add README with setup instructions",
			Status: "ACTIVE",
			URL:    "https://cursor.com/agents/bc-00000000-0000-0000-0000-000000000001",
		},
		Run: Run{
			ID:      "run-00000000-0000-0000-0000-000000000001",
			AgentID: "bc-00000000-0000-0000-0000-000000000001",
			Status:  "CREATING",
		},
	}
	respBody, err := json.Marshal(want)
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}

	autoCreatePR := true
	branch := "main"
	wantReq := CreateAgentRequest{
		Prompt: AgentPrompt{Text: "Add README with setup instructions"},
		Repos: []AgentRepo{
			{URL: "https://github.com/org/repo", StartingRef: &branch},
		},
		AutoCreatePR: &autoCreatePR,
	}
	wantReqBody, err := json.Marshal(wantReq)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	var gotAuth string
	var gotBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/agents" {
			t.Errorf("path = %q, want /agents", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", r.Header.Get("Content-Type"))
		}
		gotAuth = r.Header.Get("Authorization")
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(respBody)
	}))
	defer server.Close()

	client := NewClient(Config{
		APIKey:    "test-api-key",
		AgentsURL: server.URL + "/agents",
	})

	got, err := client.CreateAgent(context.Background(), wantReq)
	if err != nil {
		t.Fatalf("CreateAgent() error = %v", err)
	}
	if !reflect.DeepEqual(*got, want) {
		t.Fatalf("CreateAgent() = %+v, want %+v", got, want)
	}

	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("test-api-key:"))
	if gotAuth != wantAuth {
		t.Fatalf("Authorization = %q, want %q", gotAuth, wantAuth)
	}
	if !reflect.DeepEqual(gotBody, wantReqBody) {
		t.Fatalf("request body = %s, want %s", gotBody, wantReqBody)
	}
}

func TestCreateAgentAPIError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "bad request", http.StatusBadRequest)
	}))
	defer server.Close()

	client := NewClient(Config{
		APIKey:    "bad-key",
		AgentsURL: server.URL + "/agents",
	})

	_, err := client.CreateAgent(context.Background(), CreateAgentRequest{
		Prompt: AgentPrompt{Text: "test"},
	})
	if err == nil {
		t.Fatal("CreateAgent() error = nil, want API error")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error = %T, want *APIError", err)
	}
	if apiErr.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", apiErr.StatusCode)
	}
}

func TestCreateAgentAgentBusy(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, `{"error":"agent_busy"}`, http.StatusConflict)
	}))
	defer server.Close()

	client := NewClient(Config{
		APIKey:    "test-api-key",
		AgentsURL: server.URL + "/agents",
	})

	_, err := client.CreateAgent(context.Background(), CreateAgentRequest{
		Prompt: AgentPrompt{Text: "Add README"},
	})
	if err == nil {
		t.Fatal("CreateAgent() error = nil, want agent busy error")
	}
	if !errors.Is(err, ErrAgentBusy) {
		t.Fatalf("error = %v, want ErrAgentBusy", err)
	}
}

func TestCreateRunSuccess(t *testing.T) {
	t.Parallel()

	agentID := "bc-00000000-0000-0000-0000-000000000001"
	want := CreateRunResponse{
		Run: Run{
			ID:      "run-00000000-0000-0000-0000-000000000002",
			AgentID: agentID,
			Status:  "CREATING",
		},
	}
	respBody, err := json.Marshal(want)
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}

	wantReq := CreateRunRequest{
		Prompt: AgentPrompt{Text: "Fix the failing test"},
	}
	wantReqBody, err := json.Marshal(wantReq)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	var gotAuth string
	var gotBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/agents/"+agentID+"/runs" {
			t.Errorf("path = %q, want /agents/%s/runs", r.URL.Path, agentID)
		}
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", r.Header.Get("Content-Type"))
		}
		gotAuth = r.Header.Get("Authorization")
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(respBody)
	}))
	defer server.Close()

	client := NewClient(Config{
		APIKey:    "test-api-key",
		AgentsURL: server.URL + "/agents",
	})

	got, err := client.CreateRun(context.Background(), agentID, wantReq)
	if err != nil {
		t.Fatalf("CreateRun() error = %v", err)
	}
	if !reflect.DeepEqual(*got, want) {
		t.Fatalf("CreateRun() = %+v, want %+v", got, want)
	}

	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("test-api-key:"))
	if gotAuth != wantAuth {
		t.Fatalf("Authorization = %q, want %q", gotAuth, wantAuth)
	}
	if !reflect.DeepEqual(gotBody, wantReqBody) {
		t.Fatalf("request body = %s, want %s", gotBody, wantReqBody)
	}
}

func TestCreateRunAgentBusy(t *testing.T) {
	t.Parallel()

	agentID := "bc-00000000-0000-0000-0000-000000000001"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, `{"error":"agent_busy"}`, http.StatusConflict)
	}))
	defer server.Close()

	client := NewClient(Config{
		APIKey:    "test-api-key",
		AgentsURL: server.URL + "/agents",
	})

	_, err := client.CreateRun(context.Background(), agentID, CreateRunRequest{
		Prompt: AgentPrompt{Text: "Fix the failing test"},
	})
	if err == nil {
		t.Fatal("CreateRun() error = nil, want agent busy error")
	}
	if !errors.Is(err, ErrAgentBusy) {
		t.Fatalf("error = %v, want ErrAgentBusy", err)
	}
}

func TestCreateRunAPIError(t *testing.T) {
	t.Parallel()

	agentID := "bc-00000000-0000-0000-0000-000000000001"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "bad request", http.StatusBadRequest)
	}))
	defer server.Close()

	client := NewClient(Config{
		APIKey:    "bad-key",
		AgentsURL: server.URL + "/agents",
	})

	_, err := client.CreateRun(context.Background(), agentID, CreateRunRequest{
		Prompt: AgentPrompt{Text: "test"},
	})
	if err == nil {
		t.Fatal("CreateRun() error = nil, want API error")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error = %T, want *APIError", err)
	}
	if apiErr.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", apiErr.StatusCode)
	}
}

func TestGetRunStatusSuccess(t *testing.T) {
	t.Parallel()

	agentID := "bc-00000000-0000-0000-0000-000000000001"
	runID := "run-00000000-0000-0000-0000-000000000001"
	result := "Added README.md"
	prURL := "https://github.com/org/repo/pull/123"
	branch := "cursor/add-readme-a1b2"
	want := RunStatusResponse{
		ID:      runID,
		AgentID: agentID,
		Status:  "FINISHED",
		Result:  &result,
		Git: &RunGitInfo{
			Branches: []RunGitBranch{
				{
					RepoURL: "github.com/org/repo",
					Branch:  &branch,
					PRURL:   &prURL,
				},
			},
		},
	}
	body, err := json.Marshal(want)
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}

	var gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/agents/"+agentID+"/runs/"+runID {
			t.Errorf("path = %q, want /agents/%s/runs/%s", r.URL.Path, agentID, runID)
		}
		if r.Method != http.MethodGet {
			t.Errorf("method = %q, want GET", r.Method)
		}
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
	defer server.Close()

	client := NewClient(Config{
		APIKey:    "test-api-key",
		AgentsURL: server.URL + "/agents",
	})

	got, err := client.GetRunStatus(context.Background(), agentID, runID)
	if err != nil {
		t.Fatalf("GetRunStatus() error = %v", err)
	}
	if !reflect.DeepEqual(*got, want) {
		t.Fatalf("GetRunStatus() = %+v, want %+v", got, want)
	}

	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("test-api-key:"))
	if gotAuth != wantAuth {
		t.Fatalf("Authorization = %q, want %q", gotAuth, wantAuth)
	}
}

func TestGetRunStatusAPIError(t *testing.T) {
	t.Parallel()

	agentID := "bc-00000000-0000-0000-0000-000000000001"
	runID := "run-00000000-0000-0000-0000-000000000001"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClient(Config{
		APIKey:    "bad-key",
		AgentsURL: server.URL + "/agents",
	})

	_, err := client.GetRunStatus(context.Background(), agentID, runID)
	if err == nil {
		t.Fatal("GetRunStatus() error = nil, want API error")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error = %T, want *APIError", err)
	}
	if apiErr.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", apiErr.StatusCode)
	}
}

func TestCancelRunSuccess(t *testing.T) {
	t.Parallel()

	agentID := "bc-00000000-0000-0000-0000-000000000001"
	runID := "run-00000000-0000-0000-0000-000000000001"
	want := CancelRunResponse{ID: runID}
	body, err := json.Marshal(want)
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}

	var gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/agents/"+agentID+"/runs/"+runID+"/cancel" {
			t.Errorf("path = %q, want /agents/%s/runs/%s/cancel", r.URL.Path, agentID, runID)
		}
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
	defer server.Close()

	client := NewClient(Config{
		APIKey:    "test-api-key",
		AgentsURL: server.URL + "/agents",
	})

	got, err := client.CancelRun(context.Background(), agentID, runID)
	if err != nil {
		t.Fatalf("CancelRun() error = %v", err)
	}
	if got.ID != want.ID {
		t.Fatalf("CancelRun() = %+v, want %+v", got, want)
	}

	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("test-api-key:"))
	if gotAuth != wantAuth {
		t.Fatalf("Authorization = %q, want %q", gotAuth, wantAuth)
	}
}

func TestCancelRunAPIError(t *testing.T) {
	t.Parallel()

	agentID := "bc-00000000-0000-0000-0000-000000000001"
	runID := "run-00000000-0000-0000-0000-000000000001"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, `{"error":"run_not_cancellable"}`, http.StatusConflict)
	}))
	defer server.Close()

	client := NewClient(Config{
		APIKey:    "test-api-key",
		AgentsURL: server.URL + "/agents",
	})

	_, err := client.CancelRun(context.Background(), agentID, runID)
	if err == nil {
		t.Fatal("CancelRun() error = nil, want API error")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error = %T, want *APIError", err)
	}
	if apiErr.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want 409", apiErr.StatusCode)
	}
}

func TestStreamRunSuccess(t *testing.T) {
	t.Parallel()

	agentID := "bc-00000000-0000-0000-0000-000000000001"
	runID := "run-00000000-0000-0000-0000-000000000001"

	ssePayload := "event: status\ndata: {\"runId\":\"run-1\",\"status\":\"RUNNING\"}\n\nevent: done\ndata: {}\n\n"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/agents/"+agentID+"/runs/"+runID+"/stream" {
			t.Errorf("path = %q, want /agents/%s/runs/%s/stream", r.URL.Path, agentID, runID)
		}
		if r.Method != http.MethodGet {
			t.Errorf("method = %q, want GET", r.Method)
		}
		if r.Header.Get("Accept") != "text/event-stream" {
			t.Errorf("Accept = %q, want text/event-stream", r.Header.Get("Accept"))
		}
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte(ssePayload))
	}))
	defer server.Close()

	client := NewClient(Config{
		APIKey:    "test-api-key",
		AgentsURL: server.URL + "/agents",
	})

	stream, err := client.StreamRun(context.Background(), agentID, runID)
	if err != nil {
		t.Fatalf("StreamRun() error = %v", err)
	}
	defer func() { _ = stream.Close() }()

	evt1, err := stream.Next()
	if err != nil {
		t.Fatalf("Next() error = %v", err)
	}
	if evt1.Event != "status" {
		t.Fatalf("event = %q, want status", evt1.Event)
	}

	evt2, err := stream.Next()
	if err != nil {
		t.Fatalf("Next() error = %v", err)
	}
	if evt2.Event != "done" {
		t.Fatalf("event = %q, want done", evt2.Event)
	}

	_, err = stream.Next()
	if err != io.EOF {
		t.Fatalf("Next() error = %v, want io.EOF", err)
	}
}

func TestStreamRunAPIError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClient(Config{
		APIKey:    "test-api-key",
		AgentsURL: server.URL + "/agents",
	})

	_, err := client.StreamRun(context.Background(), "bc-1", "run-1")
	if err == nil {
		t.Fatal("StreamRun() error = nil, want API error")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error = %T, want *APIError", err)
	}
	if apiErr.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", apiErr.StatusCode)
	}
}
