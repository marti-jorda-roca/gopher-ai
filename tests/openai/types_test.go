package openai_test

import (
	"testing"

	"github.com/marti-jorda-roca/gopher-ai/gopherai/openai"
)

func TestNewFunctionTool_CreatesToolWithStrictTrue(t *testing.T) {
	params := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{
				"type": "string",
			},
		},
	}

	tool := openai.NewFunctionTool("test_tool", "A test tool", params)

	if tool.Type != "function" {
		t.Errorf("expected type 'function', got '%s'", tool.Type)
	}

	if tool.Name != "test_tool" {
		t.Errorf("expected name 'test_tool', got '%s'", tool.Name)
	}

	if tool.Description != "A test tool" {
		t.Errorf("expected description 'A test tool', got '%s'", tool.Description)
	}

	if tool.Strict == nil {
		t.Fatal("expected strict to be set")
	}

	if !*tool.Strict {
		t.Error("expected strict to be true")
	}
}

func TestNewFunctionCallOutput_CreatesFunctionCallOutputItem(t *testing.T) {
	output := openai.NewFunctionCallOutput("call_123", "result data")

	if output.Type != "function_call_output" {
		t.Errorf("expected type 'function_call_output', got '%s'", output.Type)
	}

	if output.CallID != "call_123" {
		t.Errorf("expected CallID 'call_123', got '%s'", output.CallID)
	}

	if output.Output != "result data" {
		t.Errorf("expected output 'result data', got '%s'", output.Output)
	}
}

func TestOutputItem_ToInputItem_CopiesFields(t *testing.T) {
	outputItem := &openai.OutputItem{
		Type:      "function_call",
		Name:      "test_func",
		Arguments: `{"arg":"value"}`,
		CallID:    "call_456",
	}

	inputItem := outputItem.ToInputItem()

	if inputItem.Type != "function_call" {
		t.Errorf("expected type 'function_call', got '%s'", inputItem.Type)
	}

	if inputItem.Name != "test_func" {
		t.Errorf("expected name 'test_func', got '%s'", inputItem.Name)
	}

	if inputItem.Arguments != `{"arg":"value"}` {
		t.Errorf("expected arguments '{\"arg\":\"value\"}', got '%s'", inputItem.Arguments)
	}

	if inputItem.CallID != "call_456" {
		t.Errorf("expected CallID 'call_456', got '%s'", inputItem.CallID)
	}
}
