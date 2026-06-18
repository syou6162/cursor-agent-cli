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
