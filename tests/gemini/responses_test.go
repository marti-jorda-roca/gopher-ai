package gemini_test

import (
	"context"
	"strings"
	"testing"

	"github.com/marti-jorda-roca/gopher-ai/gopherai"
	"github.com/marti-jorda-roca/gopher-ai/gopherai/gemini"
)

func TestConvertTool_CreatesFunctionDeclaration(t *testing.T) {
	provider := gemini.NewProvider("test-key")
	tool := gopherai.Tool{
		Name:        "test_tool",
		Description: "A test tool",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{
					"type":        "string",
					"description": "Name parameter",
				},
			},
			"required": []any{"name"},
		},
	}

	result := provider.ConvertTool(tool)
	funcDecl, ok := result.(gemini.FunctionDeclaration)
	if !ok {
		t.Fatal("expected result to be FunctionDeclaration")
	}

	if funcDecl.Name != "test_tool" {
		t.Errorf("expected name 'test_tool', got '%s'", funcDecl.Name)
	}

	if funcDecl.Description != "A test tool" {
		t.Errorf("expected description 'A test tool', got '%s'", funcDecl.Description)
	}

	if funcDecl.Parameters == nil {
		t.Fatal("expected parameters to be set")
	}
}

func TestExtractToolCalls_ReturnsErrorForInvalidType(t *testing.T) {
	provider := gemini.NewProvider("test-key")
	_, err := provider.ExtractToolCalls("invalid")

	if err == nil {
		t.Error("expected error for invalid response type")
	}
}

func TestExtractToolCalls_ExtractsFunctionCallsFromResponse(t *testing.T) {
	provider := gemini.NewProvider("test-key")
	response := &gemini.GenerateContentResponse{
		Candidates: []gemini.Candidate{
			{
				Content: gemini.Content{
					Parts: []gemini.Part{
						{
							FunctionCall: &gemini.FunctionCall{
								Name: "test_func",
								Args: map[string]any{
									"param": "value",
								},
							},
						},
					},
				},
			},
		},
	}

	calls, err := provider.ExtractToolCalls(response)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(calls) != 1 {
		t.Errorf("expected 1 tool call, got %d", len(calls))
	}

	if calls[0].Name != "test_func" {
		t.Errorf("expected call name 'test_func', got '%s'", calls[0].Name)
	}
}

func TestExtractText_ReturnsTextFromCandidate(t *testing.T) {
	provider := gemini.NewProvider("test-key")
	response := &gemini.GenerateContentResponse{
		Candidates: []gemini.Candidate{
			{
				Content: gemini.Content{
					Parts: []gemini.Part{
						{
							Text: "Hello from Gemini!",
						},
					},
				},
			},
		},
	}

	text := provider.ExtractText(response)
	if text != "Hello from Gemini!" {
		t.Errorf("expected 'Hello from Gemini!', got '%s'", text)
	}
}

func TestExtractText_ReturnsEmptyForInvalidType(t *testing.T) {
	provider := gemini.NewProvider("test-key")
	text := provider.ExtractText("invalid")

	if text != "" {
		t.Errorf("expected empty string, got '%s'", text)
	}
}

func TestBuildRequest_CreatesRequestWithStringInput(t *testing.T) {
	provider := gemini.NewProvider("test-key")

	req := provider.BuildRequest("test message", "", []any{})
	genReq, ok := req.(*gemini.GenerateContentRequest)
	if !ok {
		t.Fatal("expected GenerateContentRequest")
	}

	if len(genReq.Contents) != 1 {
		t.Errorf("expected 1 content, got %d", len(genReq.Contents))
	}

	if genReq.Contents[0].Role != "user" {
		t.Errorf("expected role 'user', got '%s'", genReq.Contents[0].Role)
	}

	if len(genReq.Contents[0].Parts) != 1 {
		t.Errorf("expected 1 part, got %d", len(genReq.Contents[0].Parts))
	}

	if genReq.Contents[0].Parts[0].Text != "test message" {
		t.Errorf("expected text 'test message', got '%s'", genReq.Contents[0].Parts[0].Text)
	}
}

func TestBuildRequest_CreatesRequestWithSliceInput(t *testing.T) {
	provider := gemini.NewProvider("test-key")

	input := []any{
		"user message",
		gemini.Content{
			Role:  "model",
			Parts: []gemini.Part{{Text: "model response"}},
		},
	}

	req := provider.BuildRequest(input, "", []any{})
	genReq := req.(*gemini.GenerateContentRequest)

	if len(genReq.Contents) != 2 {
		t.Errorf("expected 2 contents, got %d", len(genReq.Contents))
	}
}

func TestBuildRequest_AddsSystemInstructionWhenProvided(t *testing.T) {
	provider := gemini.NewProvider("test-key")

	req := provider.BuildRequest("test", "You are a helpful assistant", []any{})
	genReq := req.(*gemini.GenerateContentRequest)

	if genReq.SystemInstruction == nil {
		t.Fatal("expected system instruction to be set")
	}

	if len(genReq.SystemInstruction.Parts) != 1 {
		t.Errorf("expected 1 system instruction part, got %d", len(genReq.SystemInstruction.Parts))
	}

	if genReq.SystemInstruction.Parts[0].Text != "You are a helpful assistant" {
		t.Errorf("expected system instruction 'You are a helpful assistant', got '%s'", genReq.SystemInstruction.Parts[0].Text)
	}
}

func TestCreateFunctionCallInput_CreatesModelContent(t *testing.T) {
	provider := gemini.NewProvider("test-key")
	call := gopherai.ToolCall{
		Name:      "test_func",
		Arguments: `{"key":"value"}`,
		CallID:    "call_123",
	}

	result := provider.CreateFunctionCallInput(call)
	content, ok := result.(gemini.Content)
	if !ok {
		t.Fatal("expected result to be Content")
	}

	if content.Role != "model" {
		t.Errorf("expected role 'model', got '%s'", content.Role)
	}

	if len(content.Parts) != 1 {
		t.Errorf("expected 1 part, got %d", len(content.Parts))
	}

	if content.Parts[0].FunctionCall == nil {
		t.Fatal("expected FunctionCall to be set")
	}

	if content.Parts[0].FunctionCall.Name != "test_func" {
		t.Errorf("expected function name 'test_func', got '%s'", content.Parts[0].FunctionCall.Name)
	}
}

func TestCreateFunctionCallOutput_CreatesUserContent(t *testing.T) {
	provider := gemini.NewProvider("test-key")
	result := provider.CreateFunctionCallOutput("call_123", `{"result":"success"}`)

	content, ok := result.(gemini.Content)
	if !ok {
		t.Fatal("expected result to be Content")
	}

	if content.Role != "user" {
		t.Errorf("expected role 'user', got '%s'", content.Role)
	}

	if len(content.Parts) != 1 {
		t.Errorf("expected 1 part, got %d", len(content.Parts))
	}

	if content.Parts[0].FunctionResponse == nil {
		t.Fatal("expected FunctionResponse to be set")
	}

	if content.Parts[0].FunctionResponse.Name != "call_123" {
		t.Errorf("expected function name 'call_123', got '%s'", content.Parts[0].FunctionResponse.Name)
	}
}

func TestCreateAssistantMessage_CreatesModelContent(t *testing.T) {
	provider := gemini.NewProvider("test-key")
	result := provider.CreateAssistantMessage("Response text")

	content, ok := result.(gemini.Content)
	if !ok {
		t.Fatal("expected result to be Content")
	}

	if content.Role != "model" {
		t.Errorf("expected role 'model', got '%s'", content.Role)
	}

	if len(content.Parts) != 1 {
		t.Errorf("expected 1 part, got %d", len(content.Parts))
	}

	if content.Parts[0].Text != "Response text" {
		t.Errorf("expected text 'Response text', got '%s'", content.Parts[0].Text)
	}
}

func TestCreateResponseStream_ReturnsErrorForInvalidRequestType(t *testing.T) {
	provider := gemini.NewProvider("test-key")

	_, err := provider.CreateResponseStream(context.Background(), "invalid")
	if err == nil {
		t.Error("expected error for invalid request type")
	}

	expectedMsg := "invalid request type: expected *GenerateContentRequest"
	if err.Error() != expectedMsg {
		t.Errorf("expected error '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestParseGeminiStream_ParsesTextDeltaEvents(t *testing.T) {
	sseData := `data: {"candidates":[{"content":{"parts":[{"text":"Hello"}],"role":"model"}}]}

data: {"candidates":[{"content":{"parts":[{"text":" World"}],"role":"model"}}]}

data: {"candidates":[{"content":{"parts":[{"text":"!"}],"role":"model"},"finishReason":"STOP"}]}

`
	events := gemini.ParseGeminiStreamForTest(strings.NewReader(sseData))

	var receivedEvents []gopherai.StreamEvent
	for event := range events {
		receivedEvents = append(receivedEvents, event)
	}

	if len(receivedEvents) != 5 {
		t.Errorf("expected 5 events, got %d", len(receivedEvents))
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

	if receivedEvents[2].Type != gopherai.StreamEventTypeTextDelta {
		t.Errorf("expected third event type TextDelta, got %s", receivedEvents[2].Type)
	}
	if receivedEvents[2].Delta != "!" {
		t.Errorf("expected delta '!', got '%s'", receivedEvents[2].Delta)
	}

	if receivedEvents[3].Type != gopherai.StreamEventTypeTextDone {
		t.Errorf("expected fourth event type TextDone, got %s", receivedEvents[3].Type)
	}
	if receivedEvents[3].Text != "Hello World!" {
		t.Errorf("expected text 'Hello World!', got '%s'", receivedEvents[3].Text)
	}

	if receivedEvents[4].Type != gopherai.StreamEventTypeDone {
		t.Errorf("expected fifth event type Done, got %s", receivedEvents[4].Type)
	}
}

func TestParseGeminiStream_ParsesToolCallEvents(t *testing.T) {
	sseData := `data: {"candidates":[{"content":{"parts":[{"functionCall":{"name":"get_weather","args":{"location":"NYC"}}}],"role":"model"},"finishReason":"STOP"}]}

`
	events := gemini.ParseGeminiStreamForTest(strings.NewReader(sseData))

	var toolCallEvent *gopherai.StreamEvent
	var doneEvent *gopherai.StreamEvent
	for event := range events {
		if event.Type == gopherai.StreamEventTypeToolCall {
			toolCallEvent = &event
		}
		if event.Type == gopherai.StreamEventTypeDone {
			doneEvent = &event
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

	if toolCallEvent.ToolCall.CallID != "get_weather" {
		t.Errorf("expected CallID 'get_weather', got '%s'", toolCallEvent.ToolCall.CallID)
	}

	if toolCallEvent.ToolCall.Arguments != `{"location":"NYC"}` {
		t.Errorf("expected arguments '{\"location\":\"NYC\"}', got '%s'", toolCallEvent.ToolCall.Arguments)
	}

	if doneEvent == nil {
		t.Fatal("expected done event")
	}
}

func TestParseGeminiStream_SkipsEmptyAndNonDataLines(t *testing.T) {
	sseData := `
event: ping

data: {"candidates":[{"content":{"parts":[{"text":"test"}],"role":"model"}}]}

: this is a comment
data: {"candidates":[{"content":{"parts":[{"text":"!"}],"role":"model"},"finishReason":"STOP"}]}

`
	events := gemini.ParseGeminiStreamForTest(strings.NewReader(sseData))

	var count int
	for range events {
		count++
	}

	if count != 4 {
		t.Errorf("expected 4 events (2 deltas + textDone + done), got %d", count)
	}
}

func TestParseGeminiStream_NoTextDoneIfNoText(t *testing.T) {
	sseData := `data: {"candidates":[{"content":{"parts":[{"functionCall":{"name":"test","args":{}}}],"role":"model"},"finishReason":"STOP"}]}

`
	events := gemini.ParseGeminiStreamForTest(strings.NewReader(sseData))

	var hasTextDone bool
	for event := range events {
		if event.Type == gopherai.StreamEventTypeTextDone {
			hasTextDone = true
		}
	}

	if hasTextDone {
		t.Error("expected no TextDone event when there's no text")
	}
}
