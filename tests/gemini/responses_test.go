package gemini_test

import (
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
