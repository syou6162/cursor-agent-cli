package cursor

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
	if len(got.Items) != 1 || got.Items[0].ID != "composer-2" {
		t.Fatalf("ListModels() = %+v, want composer-2", got)
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
	if !strings.Contains(err.Error(), "status=401") {
		t.Fatalf("error = %q, want status=401", err.Error())
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
	if len(got.Items) != 1 || got.Items[0].ID != "bc-00000000-0000-0000-0000-000000000001" {
		t.Fatalf("ListAgents() = %+v, want agent id", got)
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
	if !strings.Contains(err.Error(), "status=401") {
		t.Fatalf("error = %q, want status=401", err.Error())
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
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotLimit = r.URL.Query().Get("limit")
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
	if len(got.Items) != 1 {
		t.Fatalf("ListAgents() = %+v, want one item", got)
	}
}
