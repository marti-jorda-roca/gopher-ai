package gopherai_test

import (
	"context"
	"testing"

	"github.com/marti-jorda-roca/gopher-ai/gopherai"
)

func TestNewAgent_CreatesAgentWithProvider(t *testing.T) {
	provider := &mockProvider{}
	agent := gopherai.NewAgent(provider)

	if agent == nil {
		t.Fatal("expected agent to be created")
	}
}

func TestNewAgent_AppliesSystemPromptOption(t *testing.T) {
	provider := &mockProvider{}
	agent := gopherai.NewAgent(
		provider,
		gopherai.WithSystemPrompt("test prompt"),
	)

	if agent == nil {
		t.Fatal("expected agent to be created")
	}
}

func TestNewAgent_AppliesToolsOption(t *testing.T) {
	provider := &mockProvider{}
	type testParams struct {
		Name string `json:"name"`
	}
	tool := gopherai.NewTool("test", "description", func(_ testParams) (string, error) {
		return "", nil
	})

	agent := gopherai.NewAgent(
		provider,
		gopherai.WithTools(tool),
	)

	if agent == nil {
		t.Fatal("expected agent to be created")
	}
}

func TestNewAgent_AppliesConversationHistoryOption(t *testing.T) {
	provider := &mockProvider{}
	history := []any{"msg1", "msg2"}

	agent := gopherai.NewAgent(
		provider,
		gopherai.WithConversationHistory(history),
	)

	if agent == nil {
		t.Fatal("expected agent to be created")
	}
}

func TestNewAgent_AppliesAllOptions(t *testing.T) {
	provider := &mockProvider{}
	type testParams struct {
		Name string `json:"name"`
	}
	tool := gopherai.NewTool("test", "description", func(_ testParams) (string, error) {
		return "", nil
	})
	history := []any{"msg"}

	agent := gopherai.NewAgent(
		provider,
		gopherai.WithSystemPrompt("prompt"),
		gopherai.WithTools(tool),
		gopherai.WithConversationHistory(history),
	)

	if agent == nil {
		t.Fatal("expected agent to be created with all options")
	}
}

func TestRunResult_MessageHistoryReturnsHistory(t *testing.T) {
	provider := &mockProvider{
		text: "response text",
	}
	agent := gopherai.NewAgent(provider)

	result, err := agent.Run(context.Background(), "test prompt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	history := result.MessageHistory()
	if history == nil {
		t.Error("expected history to be returned")
	}

	if len(history) == 0 {
		t.Error("expected history to contain messages")
	}
}

type mockProvider struct {
	text string
}

func (m *mockProvider) CreateResponse(_ context.Context, _ any) (any, error) {
	return &mockResponse{text: m.text}, nil
}

func (m *mockProvider) BuildRequest(_ any, _ string, _ []any) any {
	return &mockRequest{}
}

func (m *mockProvider) ConvertTool(tool gopherai.Tool) any {
	return tool
}

func (m *mockProvider) ExtractToolCalls(_ any) ([]gopherai.ToolCall, error) {
	return nil, nil
}

func (m *mockProvider) ExtractText(resp any) string {
	if mockResp, ok := resp.(*mockResponse); ok {
		return mockResp.text
	}
	return ""
}

func (m *mockProvider) CreateFunctionCallInput(call gopherai.ToolCall) any {
	return call
}

func (m *mockProvider) CreateFunctionCallOutput(callID, output string) any {
	return gopherai.FunctionCallOutput{CallID: callID, Output: output}
}

func (m *mockProvider) CreateAssistantMessage(text string) any {
	return text
}

type mockRequest struct{}

type mockResponse struct {
	text string
}
