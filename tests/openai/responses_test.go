package openai_test

import (
	"context"
	"strings"
	"testing"

	"github.com/marti-jorda-roca/gopher-ai/gopherai"
	"github.com/marti-jorda-roca/gopher-ai/gopherai/openai"
)

func TestGetOutputText_ReturnsTextFromMessageOutput(t *testing.T) {
	response := &openai.Response{
		Output: []openai.OutputItem{
			{
				Type: "message",
				Content: []openai.ContentItem{
					{
						Type: "output_text",
						Text: "Hello, world!",
					},
				},
			},
		},
	}

	text := response.GetOutputText()
	if text != "Hello, world!" {
		t.Errorf("expected 'Hello, world!', got '%s'", text)
	}
}

func TestGetOutputText_ReturnsEmptyWhenNoTextOutput(t *testing.T) {
	response := &openai.Response{
		Output: []openai.OutputItem{
			{
				Type: "function_call",
			},
		},
	}

	text := response.GetOutputText()
	if text != "" {
		t.Errorf("expected empty string, got '%s'", text)
	}
}

func TestGetToolCalls_ReturnsFunctionCallOutputItems(t *testing.T) {
	response := &openai.Response{
		Output: []openai.OutputItem{
			{
				Type:      "function_call",
				Name:      "test_func",
				Arguments: `{"arg":"value"}`,
			},
			{
				Type: "message",
			},
		},
	}

	calls := response.GetToolCalls()
	if len(calls) != 1 {
		t.Errorf("expected 1 tool call, got %d", len(calls))
	}

	if calls[0].Name != "test_func" {
		t.Errorf("expected name 'test_func', got '%s'", calls[0].Name)
	}
}

func TestGetToolCalls_ReturnsEmptySliceWhenNoFunctionCalls(t *testing.T) {
	response := &openai.Response{
		Output: []openai.OutputItem{
			{
				Type: "message",
			},
		},
	}

	calls := response.GetToolCalls()
	if len(calls) != 0 {
		t.Errorf("expected 0 tool calls, got %d", len(calls))
	}
}

func TestConvertTool_CreatesStrictFunctionTool(t *testing.T) {
	provider := openai.NewProvider("test-key")
	tool := gopherai.Tool{
		Name:        "test_tool",
		Description: "A test tool",
		Parameters: map[string]any{
			"type": "object",
		},
	}

	result := provider.ConvertTool(tool)
	funcTool, ok := result.(openai.FunctionTool)
	if !ok {
		t.Fatal("expected result to be FunctionTool")
	}

	if funcTool.Type != "function" {
		t.Errorf("expected type 'function', got '%s'", funcTool.Type)
	}

	if funcTool.Name != "test_tool" {
		t.Errorf("expected name 'test_tool', got '%s'", funcTool.Name)
	}

	if funcTool.Strict == nil || !*funcTool.Strict {
		t.Error("expected strict to be true")
	}
}

func TestExtractToolCalls_ReturnsErrorForInvalidType(t *testing.T) {
	provider := openai.NewProvider("test-key")
	_, err := provider.ExtractToolCalls("invalid")

	if err == nil {
		t.Error("expected error for invalid response type")
	}
}

func TestExtractToolCalls_ExtractsFunctionCallsFromResponse(t *testing.T) {
	provider := openai.NewProvider("test-key")
	response := &openai.Response{
		Output: []openai.OutputItem{
			{
				Type:      "function_call",
				Name:      "func1",
				Arguments: `{"key":"value"}`,
				CallID:    "call_123",
			},
			{
				Type:      "function_call",
				Name:      "func2",
				Arguments: `{}`,
				CallID:    "call_456",
			},
		},
	}

	calls, err := provider.ExtractToolCalls(response)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(calls) != 2 {
		t.Errorf("expected 2 tool calls, got %d", len(calls))
	}

	if calls[0].Name != "func1" {
		t.Errorf("expected first call name 'func1', got '%s'", calls[0].Name)
	}

	if calls[0].CallID != "call_123" {
		t.Errorf("expected first call ID 'call_123', got '%s'", calls[0].CallID)
	}
}

func TestExtractText_ReturnsEmptyForInvalidType(t *testing.T) {
	provider := openai.NewProvider("test-key")
	text := provider.ExtractText("invalid")

	if text != "" {
		t.Errorf("expected empty string, got '%s'", text)
	}
}

func TestBuildRequest_CreatesRequestWithStringInput(t *testing.T) {
	provider := openai.NewProvider("test-key")
	provider.SetModel("gpt-4o")

	req := provider.BuildRequest("test input", "system prompt", []any{})
	createReq, ok := req.(*openai.CreateResponseRequest)
	if !ok {
		t.Fatal("expected CreateResponseRequest")
	}

	if createReq.Model != "gpt-4o" {
		t.Errorf("expected model 'gpt-4o', got '%s'", createReq.Model)
	}

	if createReq.Instructions != "system prompt" {
		t.Errorf("expected instructions 'system prompt', got '%s'", createReq.Instructions)
	}
}

func TestBuildRequest_ConvertsSliceInputToInputItems(t *testing.T) {
	provider := openai.NewProvider("test-key")

	input := []any{
		"user message",
		openai.InputItem{Type: "message", Role: "assistant", Content: "assistant message"},
	}

	req := provider.BuildRequest(input, "", []any{})
	createReq := req.(*openai.CreateResponseRequest)

	items, ok := createReq.Input.([]openai.InputItem)
	if !ok {
		t.Fatal("expected input to be []InputItem")
	}

	if len(items) != 2 {
		t.Errorf("expected 2 input items, got %d", len(items))
	}

	if items[0].Content != "user message" {
		t.Errorf("expected first item content 'user message', got '%s'", items[0].Content)
	}
}

func TestCreateFunctionCallInput_CreatesInputItem(t *testing.T) {
	provider := openai.NewProvider("test-key")
	call := gopherai.ToolCall{
		Name:      "test_func",
		Arguments: `{"arg":"val"}`,
		CallID:    "call_123",
	}

	result := provider.CreateFunctionCallInput(call)
	inputItem, ok := result.(openai.InputItem)
	if !ok {
		t.Fatal("expected result to be InputItem")
	}

	if inputItem.Type != "function_call" {
		t.Errorf("expected type 'function_call', got '%s'", inputItem.Type)
	}

	if inputItem.Name != "test_func" {
		t.Errorf("expected name 'test_func', got '%s'", inputItem.Name)
	}

	if inputItem.CallID != "call_123" {
		t.Errorf("expected CallID 'call_123', got '%s'", inputItem.CallID)
	}
}

func TestCreateFunctionCallOutput_CreatesOutputItem(t *testing.T) {
	provider := openai.NewProvider("test-key")
	result := provider.CreateFunctionCallOutput("call_123", "output value")

	inputItem, ok := result.(openai.InputItem)
	if !ok {
		t.Fatal("expected result to be InputItem")
	}

	if inputItem.Type != "function_call_output" {
		t.Errorf("expected type 'function_call_output', got '%s'", inputItem.Type)
	}

	if inputItem.CallID != "call_123" {
		t.Errorf("expected CallID 'call_123', got '%s'", inputItem.CallID)
	}

	if inputItem.Output != "output value" {
		t.Errorf("expected output 'output value', got '%s'", inputItem.Output)
	}
}

func TestCreateAssistantMessage_CreatesMessageItem(t *testing.T) {
	provider := openai.NewProvider("test-key")
	result := provider.CreateAssistantMessage("Hello!")

	inputItem, ok := result.(openai.InputItem)
	if !ok {
		t.Fatal("expected result to be InputItem")
	}

	if inputItem.Type != "message" {
		t.Errorf("expected type 'message', got '%s'", inputItem.Type)
	}

	if inputItem.Role != "assistant" {
		t.Errorf("expected role 'assistant', got '%s'", inputItem.Role)
	}

	if inputItem.Content != "Hello!" {
		t.Errorf("expected content 'Hello!', got '%s'", inputItem.Content)
	}
}

func TestCreateResponseStream_ReturnsErrorForInvalidRequestType(t *testing.T) {
	provider := openai.NewProvider("test-key")

	_, err := provider.CreateResponseStream(context.Background(), "invalid")
	if err == nil {
		t.Error("expected error for invalid request type")
	}

	expectedMsg := "invalid request type: expected *CreateResponseRequest"
	if err.Error() != expectedMsg {
		t.Errorf("expected error '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestParseSSEStream_ParsesTextDeltaEvents(t *testing.T) {
	sseData := `data: {"type":"response.output_text.delta","delta":"Hello"}

data: {"type":"response.output_text.delta","delta":" World"}

data: {"type":"response.output_text.done","text":"Hello World"}

data: {"type":"response.completed"}

data: [DONE]
`
	events := openai.ParseSSEStreamForTest(strings.NewReader(sseData))

	var receivedEvents []gopherai.StreamEvent
	for event := range events {
		receivedEvents = append(receivedEvents, event)
	}

	if len(receivedEvents) != 4 {
		t.Errorf("expected 4 events, got %d", len(receivedEvents))
	}

	if receivedEvents[0].Type != gopherai.StreamEventTypeTextDelta {
		t.Errorf("expected first event type TextDelta, got %s", receivedEvents[0].Type)
	}
	if receivedEvents[0].Delta != "Hello" {
		t.Errorf("expected delta 'Hello', got '%s'", receivedEvents[0].Delta)
	}

	if receivedEvents[1].Type != gopherai.StreamEventTypeTextDelta {
		t.Errorf("expected second event type TextDelta, got %s", receivedEvents[1].Type)
	}
	if receivedEvents[1].Delta != " World" {
		t.Errorf("expected delta ' World', got '%s'", receivedEvents[1].Delta)
	}

	if receivedEvents[2].Type != gopherai.StreamEventTypeTextDone {
		t.Errorf("expected third event type TextDone, got %s", receivedEvents[2].Type)
	}
	if receivedEvents[2].Text != "Hello World" {
		t.Errorf("expected text 'Hello World', got '%s'", receivedEvents[2].Text)
	}

	if receivedEvents[3].Type != gopherai.StreamEventTypeDone {
		t.Errorf("expected fourth event type Done, got %s", receivedEvents[3].Type)
	}
}

func TestParseSSEStream_ParsesToolCallEvents(t *testing.T) {
	sseData := `data: {"type":"response.output_item.added","item":{"id":"item_123","type":"function_call","call_id":"call_456","name":"get_weather"}}

data: {"type":"response.function_call_arguments.done","item_id":"item_123","arguments":"{\"location\":\"NYC\"}"}

data: {"type":"response.completed"}

data: [DONE]
`
	events := openai.ParseSSEStreamForTest(strings.NewReader(sseData))

	var toolCallEvent *gopherai.StreamEvent
	for event := range events {
		if event.Type == gopherai.StreamEventTypeToolCall {
			toolCallEvent = &event
		}
	}

	if toolCallEvent == nil {
		t.Fatal("expected tool call event")
	}

	if toolCallEvent.ToolCall == nil {
		t.Fatal("expected ToolCall to be set")
	}

	if toolCallEvent.ToolCall.Name != "get_weather" {
		t.Errorf("expected name 'get_weather', got '%s'", toolCallEvent.ToolCall.Name)
	}

	if toolCallEvent.ToolCall.CallID != "call_456" {
		t.Errorf("expected CallID 'call_456', got '%s'", toolCallEvent.ToolCall.CallID)
	}

	if toolCallEvent.ToolCall.Arguments != `{"location":"NYC"}` {
		t.Errorf("expected arguments '{\"location\":\"NYC\"}', got '%s'", toolCallEvent.ToolCall.Arguments)
	}
}

func TestParseSSEStream_ParsesErrorEvents(t *testing.T) {
	sseData := `data: {"type":"error","code":"rate_limit","message":"Rate limit exceeded"}

`
	events := openai.ParseSSEStreamForTest(strings.NewReader(sseData))

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
		t.Fatal("expected Error to be set")
	}

	expectedErr := "rate_limit: Rate limit exceeded"
	if errorEvent.Error.Error() != expectedErr {
		t.Errorf("expected error '%s', got '%s'", expectedErr, errorEvent.Error.Error())
	}
}

func TestParseSSEStream_SkipsEmptyAndNonDataLines(t *testing.T) {
	sseData := `
event: ping

data: {"type":"response.output_text.delta","delta":"test"}

: this is a comment
data: {"type":"response.completed"}

data: [DONE]
`
	events := openai.ParseSSEStreamForTest(strings.NewReader(sseData))

	var count int
	for range events {
		count++
	}

	if count != 2 {
		t.Errorf("expected 2 events, got %d", count)
	}
}
