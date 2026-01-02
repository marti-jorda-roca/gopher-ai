package gopherai_test

import (
	"context"
	"errors"
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

type mockStreamProvider struct {
	mockProvider
	events      []gopherai.StreamEvent
	createError error
}

func (m *mockStreamProvider) CreateResponseStream(_ context.Context, _ any) (<-chan gopherai.StreamEvent, error) {
	if m.createError != nil {
		return nil, m.createError
	}

	ch := make(chan gopherai.StreamEvent, len(m.events))
	for _, event := range m.events {
		ch <- event
	}
	close(ch)
	return ch, nil
}

func TestRunStream_ProviderDoesNotSupportStreaming(t *testing.T) {
	provider := &mockProvider{}
	agent := gopherai.NewAgent(provider)

	_, err := agent.RunStream(context.Background(), "test prompt")
	if err == nil {
		t.Fatal("expected error when provider does not support streaming")
	}

	if err.Error() != "provider does not support streaming" {
		t.Errorf("expected 'provider does not support streaming' error, got: %v", err)
	}
}

func TestRunStream_ReturnsTextEvents(t *testing.T) {
	provider := &mockStreamProvider{
		events: []gopherai.StreamEvent{
			{Type: gopherai.StreamEventTypeTextDelta, Delta: "Hello"},
			{Type: gopherai.StreamEventTypeTextDelta, Delta: " World"},
			{Type: gopherai.StreamEventTypeTextDone, Text: "Hello World"},
			{Type: gopherai.StreamEventTypeDone},
		},
	}
	agent := gopherai.NewAgent(provider)

	events, err := agent.RunStream(context.Background(), "test prompt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var receivedEvents []gopherai.StreamEvent
	for event := range events {
		receivedEvents = append(receivedEvents, event)
	}

	if len(receivedEvents) != 4 {
		t.Errorf("expected 4 events, got %d", len(receivedEvents))
	}

	deltaCount := 0
	var doneEvent bool
	for _, event := range receivedEvents {
		if event.Type == gopherai.StreamEventTypeTextDelta {
			deltaCount++
		}
		if event.Type == gopherai.StreamEventTypeDone {
			doneEvent = true
		}
	}

	if deltaCount != 2 {
		t.Errorf("expected 2 delta events, got %d", deltaCount)
	}
	if !doneEvent {
		t.Error("expected done event")
	}
}

func TestRunStream_WithToolCalls(t *testing.T) {
	type testParams struct {
		Name string `json:"name"`
	}
	tool := gopherai.NewTool("greet", "greets a person", func(p testParams) (string, error) {
		return "Hello, " + p.Name, nil
	})

	callCount := 0
	provider := &mockStreamProviderWithToolCalls{
		mockStreamProvider: mockStreamProvider{},
		tool:               tool,
		callCount:          &callCount,
	}
	agent := gopherai.NewAgent(provider, gopherai.WithTools(tool))

	events, err := agent.RunStream(context.Background(), "greet John")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var receivedEvents []gopherai.StreamEvent
	for event := range events {
		receivedEvents = append(receivedEvents, event)
	}

	hasToolCall := false
	hasDone := false
	for _, event := range receivedEvents {
		if event.Type == gopherai.StreamEventTypeToolCall {
			hasToolCall = true
		}
		if event.Type == gopherai.StreamEventTypeDone {
			hasDone = true
		}
	}

	if !hasToolCall {
		t.Error("expected tool call event")
	}
	if !hasDone {
		t.Error("expected done event")
	}
}

func TestRunStream_ErrorDuringStream(t *testing.T) {
	provider := &mockStreamProvider{
		events: []gopherai.StreamEvent{
			{Type: gopherai.StreamEventTypeTextDelta, Delta: "Hello"},
			{Type: gopherai.StreamEventTypeError, Error: errors.New("stream error")},
		},
	}
	agent := gopherai.NewAgent(provider)

	events, err := agent.RunStream(context.Background(), "test prompt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var errorEvent *gopherai.StreamEvent
	for event := range events {
		if event.Type == gopherai.StreamEventTypeError {
			errorEvent = &event
		}
	}

	if errorEvent == nil {
		t.Fatal("expected error event")
	}

	if errorEvent.Error == nil || errorEvent.Error.Error() != "stream error" {
		t.Errorf("expected 'stream error', got: %v", errorEvent.Error)
	}
}

func TestRunStream_CreateResponseStreamError(t *testing.T) {
	provider := &mockStreamProvider{
		createError: errors.New("connection failed"),
	}
	agent := gopherai.NewAgent(provider)

	events, err := agent.RunStream(context.Background(), "test prompt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var errorEvent *gopherai.StreamEvent
	for event := range events {
		if event.Type == gopherai.StreamEventTypeError {
			errorEvent = &event
		}
	}

	if errorEvent == nil {
		t.Fatal("expected error event")
	}

	if errorEvent.Error == nil {
		t.Error("expected error in event")
	}
}

func TestRunStream_UnknownTool(t *testing.T) {
	provider := &mockStreamProvider{
		events: []gopherai.StreamEvent{
			{Type: gopherai.StreamEventTypeToolCall, ToolCall: &gopherai.ToolCall{
				Name:      "unknown_tool",
				Arguments: "{}",
				CallID:    "call_123",
			}},
			{Type: gopherai.StreamEventTypeDone},
		},
	}
	agent := gopherai.NewAgent(provider)

	events, err := agent.RunStream(context.Background(), "test prompt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var errorEvent *gopherai.StreamEvent
	for event := range events {
		if event.Type == gopherai.StreamEventTypeError {
			errorEvent = &event
		}
	}

	if errorEvent == nil {
		t.Fatal("expected error event for unknown tool")
	}
}

func TestRunStream_ToolExecutionError(t *testing.T) {
	type testParams struct {
		Name string `json:"name"`
	}
	tool := gopherai.NewTool("failing_tool", "always fails", func(_ testParams) (string, error) {
		return "", errors.New("tool execution failed")
	})

	provider := &mockStreamProvider{
		events: []gopherai.StreamEvent{
			{Type: gopherai.StreamEventTypeToolCall, ToolCall: &gopherai.ToolCall{
				Name:      "failing_tool",
				Arguments: `{"name":"test"}`,
				CallID:    "call_123",
			}},
			{Type: gopherai.StreamEventTypeDone},
		},
	}
	agent := gopherai.NewAgent(provider, gopherai.WithTools(tool))

	events, err := agent.RunStream(context.Background(), "test prompt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var errorEvent *gopherai.StreamEvent
	for event := range events {
		if event.Type == gopherai.StreamEventTypeError {
			errorEvent = &event
		}
	}

	if errorEvent == nil {
		t.Fatal("expected error event for tool execution failure")
	}
}

func TestRunStream_WithHistory(t *testing.T) {
	provider := &mockStreamProvider{
		events: []gopherai.StreamEvent{
			{Type: gopherai.StreamEventTypeTextDelta, Delta: "Response"},
			{Type: gopherai.StreamEventTypeTextDone, Text: "Response"},
			{Type: gopherai.StreamEventTypeDone},
		},
	}
	agent := gopherai.NewAgent(provider)
	history := []any{"previous message", "previous response"}

	events, err := agent.RunStream(context.Background(), "new prompt", history)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var receivedEvents []gopherai.StreamEvent
	for event := range events {
		receivedEvents = append(receivedEvents, event)
	}

	if len(receivedEvents) == 0 {
		t.Error("expected to receive events")
	}
}

func TestRunStream_WithConversationHistoryOption(t *testing.T) {
	provider := &mockStreamProvider{
		events: []gopherai.StreamEvent{
			{Type: gopherai.StreamEventTypeTextDelta, Delta: "Response"},
			{Type: gopherai.StreamEventTypeTextDone, Text: "Response"},
			{Type: gopherai.StreamEventTypeDone},
		},
	}
	history := []any{"preset message"}
	agent := gopherai.NewAgent(provider, gopherai.WithConversationHistory(history))

	events, err := agent.RunStream(context.Background(), "new prompt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var receivedEvents []gopherai.StreamEvent
	for event := range events {
		receivedEvents = append(receivedEvents, event)
	}

	if len(receivedEvents) == 0 {
		t.Error("expected to receive events")
	}
}

type mockStreamProviderWithToolCalls struct {
	mockStreamProvider
	tool      gopherai.Tool
	callCount *int
}

func (m *mockStreamProviderWithToolCalls) CreateResponseStream(_ context.Context, _ any) (<-chan gopherai.StreamEvent, error) {
	ch := make(chan gopherai.StreamEvent, 10)

	go func() {
		defer close(ch)

		if *m.callCount == 0 {
			*m.callCount++
			ch <- gopherai.StreamEvent{
				Type: gopherai.StreamEventTypeToolCall,
				ToolCall: &gopherai.ToolCall{
					Name:      m.tool.Name,
					Arguments: `{"name":"John"}`,
					CallID:    "call_123",
				},
			}
			ch <- gopherai.StreamEvent{Type: gopherai.StreamEventTypeDone}
		} else {
			ch <- gopherai.StreamEvent{Type: gopherai.StreamEventTypeTextDelta, Delta: "Done"}
			ch <- gopherai.StreamEvent{Type: gopherai.StreamEventTypeTextDone, Text: "Done"}
			ch <- gopherai.StreamEvent{Type: gopherai.StreamEventTypeDone}
		}
	}()

	return ch, nil
}
