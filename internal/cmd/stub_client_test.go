package cmd

import (
	"context"

	"github.com/syou6162/cursor-agent-cli/internal/cursor"
)

type stubClient struct {
	cursor.ModelReader
	cursor.AgentReader
	cursor.AgentWriter
	cursor.RunWriter
}

func newStubClient(model cursor.ModelReader, agent cursor.AgentReader, writer cursor.AgentWriter, runWriter cursor.RunWriter) cursor.Client {
	if model == nil {
		model = noopModelReader{}
	}
	if agent == nil {
		agent = noopAgentReader{}
	}
	if writer == nil {
		writer = noopAgentWriter{}
	}
	if runWriter == nil {
		runWriter = noopRunWriter{}
	}
	return stubClient{
		ModelReader: model,
		AgentReader: agent,
		AgentWriter: writer,
		RunWriter:   runWriter,
	}
}

func newStubClientWithModel(model cursor.ModelReader) cursor.Client {
	return newStubClient(model, noopAgentReader{}, noopAgentWriter{}, noopRunWriter{})
}

func newStubClientWithAgent(agent cursor.AgentReader) cursor.Client {
	return newStubClient(noopModelReader{}, agent, noopAgentWriter{}, noopRunWriter{})
}

func newStubClientWithAgentWriter(writer cursor.AgentWriter) cursor.Client {
	return newStubClient(noopModelReader{}, noopAgentReader{}, writer, noopRunWriter{})
}

func newStubClientWithRunWriter(writer cursor.RunWriter) cursor.Client {
	return newStubClient(noopModelReader{}, noopAgentReader{}, noopAgentWriter{}, writer)
}

type noopModelReader struct{}

func (noopModelReader) ListModels(context.Context) (*cursor.ListModelsResponse, error) {
	panic("unexpected ListModels call")
}

type noopAgentReader struct{}

func (noopAgentReader) ListAgents(context.Context, int) (*cursor.ListAgentsResponse, error) {
	panic("unexpected ListAgents call")
}

type noopAgentWriter struct{}

func (noopAgentWriter) CreateAgent(context.Context, cursor.CreateAgentRequest) (*cursor.CreateAgentResponse, error) {
	panic("unexpected CreateAgent call")
}

type noopRunWriter struct{}

func (noopRunWriter) CreateRun(context.Context, string, cursor.CreateRunRequest) (*cursor.CreateRunResponse, error) {
	panic("unexpected CreateRun call")
}
