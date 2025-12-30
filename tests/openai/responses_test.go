package openai_test

import (
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
