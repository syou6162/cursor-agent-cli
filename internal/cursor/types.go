package cursor

// ListModelsResponse is the response body for GET /v1/models.
type ListModelsResponse struct {
	Items []Model `json:"items"`
}

// Model describes one available model.
type Model struct {
	ID          string           `json:"id"`
	DisplayName string           `json:"displayName"`
	Description *string          `json:"description,omitempty"`
	Aliases     []string         `json:"aliases,omitempty"`
	Parameters  []ModelParameter `json:"parameters,omitempty"`
	Variants    []ModelVariant   `json:"variants,omitempty"`
}

// ModelParameter is a per-model parameter definition.
type ModelParameter struct {
	ID          string                `json:"id"`
	DisplayName *string               `json:"displayName,omitempty"`
	Values      []ModelParameterValue `json:"values"`
}

// ModelParameterValue is a permitted value for a model parameter.
type ModelParameterValue struct {
	Value       string  `json:"value"`
	DisplayName *string `json:"displayName,omitempty"`
}

// ModelVariant is a concrete id+params combination a model accepts.
type ModelVariant struct {
	Params      []ModelVariantParam `json:"params"`
	DisplayName string              `json:"displayName"`
	Description *string             `json:"description,omitempty"`
	IsDefault   *bool               `json:"isDefault,omitempty"`
}

// ModelVariantParam is a parameter value within a model variant.
type ModelVariantParam struct {
	ID    string `json:"id"`
	Value string `json:"value"`
}

// ListAgentsResponse is the response body for GET /v1/agents.
type ListAgentsResponse struct {
	Items      []Agent `json:"items"`
	NextCursor string  `json:"nextCursor,omitempty"`
}

// Agent describes one Cloud Agent in a list response.
type Agent struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Status      string   `json:"status"`
	Env         AgentEnv `json:"env"`
	URL         string   `json:"url"`
	CreatedAt   string   `json:"createdAt"`
	UpdatedAt   string   `json:"updatedAt"`
	LatestRunID string   `json:"latestRunId"`
}

// AgentEnv describes the runtime environment for an agent.
type AgentEnv struct {
	Type string `json:"type"`
}

// AgentPrompt is the task prompt for creating or continuing an agent run.
type AgentPrompt struct {
	Text string `json:"text"`
}

// AgentRepo describes a repository configuration for an agent.
type AgentRepo struct {
	URL         string  `json:"url"`
	StartingRef *string `json:"startingRef,omitempty"`
	PRURL       *string `json:"prUrl,omitempty"`
}

// CreateAgentRequest is the body for POST /v1/agents.
type CreateAgentRequest struct {
	Prompt       AgentPrompt `json:"prompt"`
	Repos        []AgentRepo `json:"repos,omitempty"`
	AutoCreatePR *bool       `json:"autoCreatePR,omitempty"`
}

// CreateAgentResponse is the response body for POST /v1/agents.
type CreateAgentResponse struct {
	Agent CreatedAgent `json:"agent"`
	Run   Run          `json:"run"`
}

// CreatedAgent is the full agent record returned by POST /v1/agents.
type CreatedAgent struct {
	ID                  string      `json:"id"`
	Name                string      `json:"name"`
	Status              string      `json:"status"`
	Env                 AgentEnv    `json:"env"`
	Repos               []AgentRepo `json:"repos,omitempty"`
	WorkOnCurrentBranch *bool       `json:"workOnCurrentBranch,omitempty"`
	AutoCreatePR        *bool       `json:"autoCreatePR,omitempty"`
	URL                 string      `json:"url"`
	CreatedAt           string      `json:"createdAt"`
	UpdatedAt           string      `json:"updatedAt"`
	LatestRunID         string      `json:"latestRunId"`
}

// Run describes one agent run.
type Run struct {
	ID        string `json:"id"`
	AgentID   string `json:"agentId"`
	Status    string `json:"status"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// CreateRunRequest is the body for POST /v1/agents/{agent_id}/runs.
type CreateRunRequest struct {
	Prompt AgentPrompt `json:"prompt"`
}

// CreateRunResponse is the response body for POST /v1/agents/{agent_id}/runs.
type CreateRunResponse struct {
	Run Run `json:"run"`
}
