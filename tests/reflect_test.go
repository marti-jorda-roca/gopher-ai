package gopherai_test

import (
	"testing"

	"github.com/marti-jorda-roca/gopher-ai/gopherai"
)

func TestNewTool_PanicsForNonStructType(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for non-struct type")
		}
	}()

	gopherai.NewTool("test", "description", func(s string) (string, error) {
		return s, nil
	})
}

func TestNewTool_CreatesToolWithSchemaFromStruct(t *testing.T) {
	type testParams struct {
		Name string `json:"name"`
	}

	tool := gopherai.NewTool("test_tool", "A test tool", func(p testParams) (string, error) {
		return p.Name, nil
	})

	if tool.Name != "test_tool" {
		t.Errorf("expected name 'test_tool', got '%s'", tool.Name)
	}

	if tool.Description != "A test tool" {
		t.Errorf("expected description 'A test tool', got '%s'", tool.Description)
	}

	if tool.Parameters == nil {
		t.Fatal("expected parameters to be set")
	}

	if tool.Parameters["type"] != "object" {
		t.Errorf("expected parameters type 'object', got '%v'", tool.Parameters["type"])
	}
}

func TestNewTool_HandlerUnmarshalsAndCallsFunction(t *testing.T) {
	type testParams struct {
		Value string `json:"value"`
	}

	called := false
	tool := gopherai.NewTool("test", "description", func(p testParams) (string, error) {
		called = true
		return "result: " + p.Value, nil
	})

	result, err := tool.Handler(`{"value": "test"}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !called {
		t.Error("expected function to be called")
	}

	if result != "result: test" {
		t.Errorf("expected 'result: test', got '%s'", result)
	}
}

func TestNewTool_CreatesSchemaWithRequiredFields(t *testing.T) {
	type testParams struct {
		Required string `json:"required"`
	}

	tool := gopherai.NewTool("test", "description", func(_ testParams) (string, error) {
		return "", nil
	})

	required, ok := tool.Parameters["required"].([]string)
	if !ok {
		t.Fatal("required should be []string")
	}

	if len(required) != 1 || required[0] != "required" {
		t.Errorf("expected required to contain 'required', got %v", required)
	}
}

func TestNewTool_OmitsPointerFieldsFromRequired(t *testing.T) {
	type testParams struct {
		Optional *string `json:"optional"`
	}

	tool := gopherai.NewTool("test", "description", func(_ testParams) (string, error) {
		return "", nil
	})

	_, hasRequired := tool.Parameters["required"]
	if hasRequired {
		t.Error("expected no required fields for pointer fields")
	}
}

func TestNewTool_IncludesDescriptionFromTag(t *testing.T) {
	type testParams struct {
		Field string `json:"field" description:"A test field"`
	}

	tool := gopherai.NewTool("test", "description", func(_ testParams) (string, error) {
		return "", nil
	})

	props := tool.Parameters["properties"].(map[string]any)
	fieldProp := props["field"].(map[string]any)

	if fieldProp["description"] != "A test field" {
		t.Errorf("expected description 'A test field', got '%v'", fieldProp["description"])
	}
}

func TestNewTool_IncludesEnumValuesFromTag(t *testing.T) {
	type testParams struct {
		Field string `json:"field" enum:"value1,value2,value3"`
	}

	tool := gopherai.NewTool("test", "description", func(_ testParams) (string, error) {
		return "", nil
	})

	props := tool.Parameters["properties"].(map[string]any)
	fieldProp := props["field"].(map[string]any)

	enum, ok := fieldProp["enum"].([]string)
	if !ok {
		t.Fatal("enum should be []string")
	}

	expected := []string{"value1", "value2", "value3"}
	if len(enum) != len(expected) {
		t.Errorf("expected %d enum values, got %d", len(expected), len(enum))
	}

	for i, val := range expected {
		if enum[i] != val {
			t.Errorf("expected enum[%d] to be '%s', got '%s'", i, val, enum[i])
		}
	}
}

func TestNewTool_IncludesItemsForSliceFields(t *testing.T) {
	type testParams struct {
		List []string `json:"list"`
	}

	tool := gopherai.NewTool("test", "description", func(_ testParams) (string, error) {
		return "", nil
	})

	props := tool.Parameters["properties"].(map[string]any)
	listProp := props["list"].(map[string]any)

	items, ok := listProp["items"].(map[string]any)
	if !ok {
		t.Fatal("items should be map[string]any")
	}

	if items["type"] != "string" {
		t.Errorf("expected items type 'string', got '%v'", items["type"])
	}
}

func TestNewTool_SkipsFieldsWithoutJSONTag(t *testing.T) {
	type testParams struct {
		WithTag    string `json:"with_tag"`
		WithoutTag string
	}

	tool := gopherai.NewTool("test", "description", func(_ testParams) (string, error) {
		return "", nil
	})

	props := tool.Parameters["properties"].(map[string]any)

	if _, exists := props["with_tag"]; !exists {
		t.Error("expected 'with_tag' to exist")
	}

	if _, exists := props["WithoutTag"]; exists {
		t.Error("expected field without json tag to be skipped")
	}
}
