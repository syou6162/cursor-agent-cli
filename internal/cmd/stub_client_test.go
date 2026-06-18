package cmd

import (
	"context"
	"errors"

	"github.com/syou6162/cursor-agent-cli/internal/cursor"
)

var errUnexpectedAPICall = errors.New("unexpected API call")

type stubClient struct {
	listModels  func(context.Context) (*cursor.ListModelsResponse, error)
	listAgents  func(context.Context, int) (*cursor.ListAgentsResponse, error)
	createAgent func(context.Context, cursor.CreateAgentRequest) (*cursor.CreateAgentResponse, error)
	createRun   func(context.Context, string, cursor.CreateRunRequest) (*cursor.CreateRunResponse, error)
}

func (s stubClient) ListModels(ctx context.Context) (*cursor.ListModelsResponse, error) {
	if s.listModels == nil {
		return nil, errUnexpectedAPICall
	}
	return s.listModels(ctx)
}

func (s stubClient) ListAgents(ctx context.Context, limit int) (*cursor.ListAgentsResponse, error) {
	if s.listAgents == nil {
		return nil, errUnexpectedAPICall
	}
	return s.listAgents(ctx, limit)
}

func (s stubClient) CreateAgent(ctx context.Context, req cursor.CreateAgentRequest) (*cursor.CreateAgentResponse, error) {
	if s.createAgent == nil {
		return nil, errUnexpectedAPICall
	}
	return s.createAgent(ctx, req)
}

func (s stubClient) CreateRun(ctx context.Context, agentID string, req cursor.CreateRunRequest) (*cursor.CreateRunResponse, error) {
	if s.createRun == nil {
		return nil, errUnexpectedAPICall
	}
	return s.createRun(ctx, agentID, req)
}

type stubModelReader struct {
	response *cursor.ListModelsResponse
	err      error
}

func (s stubModelReader) bind() func(context.Context) (*cursor.ListModelsResponse, error) {
	return func(context.Context) (*cursor.ListModelsResponse, error) {
		if s.err != nil {
			return nil, s.err
		}
		return s.response, nil
	}
}

type stubAgentReader struct {
	response *cursor.ListAgentsResponse
	err      error
	limit    int
}

func (s *stubAgentReader) bind() func(context.Context, int) (*cursor.ListAgentsResponse, error) {
	return func(_ context.Context, limit int) (*cursor.ListAgentsResponse, error) {
		s.limit = limit
		if s.err != nil {
			return nil, s.err
		}
		return s.response, nil
	}
}

type stubAgentWriter struct {
	response *cursor.CreateAgentResponse
	err      error
	req      cursor.CreateAgentRequest
}

func (s *stubAgentWriter) bind() func(context.Context, cursor.CreateAgentRequest) (*cursor.CreateAgentResponse, error) {
	return func(_ context.Context, req cursor.CreateAgentRequest) (*cursor.CreateAgentResponse, error) {
		s.req = req
		if s.err != nil {
			return nil, s.err
		}
		return s.response, nil
	}
}

type stubRunWriter struct {
	agentID  string
	req      cursor.CreateRunRequest
	response *cursor.CreateRunResponse
	err      error
}

func (s *stubRunWriter) bind() func(context.Context, string, cursor.CreateRunRequest) (*cursor.CreateRunResponse, error) {
	return func(_ context.Context, agentID string, req cursor.CreateRunRequest) (*cursor.CreateRunResponse, error) {
		s.agentID = agentID
		s.req = req
		if s.err != nil {
			return nil, s.err
		}
		return s.response, nil
	}
}

func newStubClientWithModel(reader stubModelReader) cursor.Client {
	return stubClient{listModels: reader.bind()}
}

func newStubClientWithAgent(reader *stubAgentReader) cursor.Client {
	return stubClient{listAgents: reader.bind()}
}

func newStubClientWithAgentWriter(writer *stubAgentWriter) cursor.Client {
	return stubClient{createAgent: writer.bind()}
}

func newStubClientWithRunWriter(writer *stubRunWriter) cursor.Client {
	return stubClient{createRun: writer.bind()}
}
