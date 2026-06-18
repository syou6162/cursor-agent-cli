package cmd

import (
	"context"

	"github.com/syou6162/cursor-agent-cli/internal/cursor"
)

type stubClient struct {
	cursor.ModelReader
	cursor.AgentReader
}

func newStubClient(model cursor.ModelReader, agent cursor.AgentReader) cursor.Client {
	if model == nil {
		model = noopModelReader{}
	}
	if agent == nil {
		agent = noopAgentReader{}
	}
	return stubClient{
		ModelReader: model,
		AgentReader: agent,
	}
}

func newStubClientWithModel(model cursor.ModelReader) cursor.Client {
	return newStubClient(model, noopAgentReader{})
}

func newStubClientWithAgent(agent cursor.AgentReader) cursor.Client {
	return newStubClient(noopModelReader{}, agent)
}

type noopModelReader struct{}

func (noopModelReader) ListModels(context.Context) (*cursor.ListModelsResponse, error) {
	panic("unexpected ListModels call")
}

type noopAgentReader struct{}

func (noopAgentReader) ListAgents(context.Context, int) (*cursor.ListAgentsResponse, error) {
	panic("unexpected ListAgents call")
}
