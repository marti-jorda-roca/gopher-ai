package gopherai

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// NewTool creates a Tool from a Go function by automatically generating
// the JSON schema from the function's parameter struct type.
// The type parameter T must be a struct with json tags on its fields.
// Fields can have optional "description" and "enum" tags for enhanced schema generation.
func NewTool[T any](name, description string, fn func(T) (string, error)) Tool {
	var zero T
	t := reflect.TypeOf(zero)

	if t.Kind() != reflect.Struct {
		panic(fmt.Sprintf("NewTool: type parameter T must be a struct, got %s", t.Kind()))
	}

	schema := generateSchema(t)

	handler := func(args string) (string, error) {
		var params T
		if err := json.Unmarshal([]byte(args), &params); err != nil {
			return "", fmt.Errorf("failed to parse arguments: %w", err)
		}
		return fn(params)
	}

	return Tool{
		Name:        name,
		Description: description,
		Parameters:  schema,
		Handler:     handler,
	}
}

func generateSchema(t reflect.Type) map[string]any {
	schema := map[string]any{
		"type":       "object",
		"properties": make(map[string]any),
	}

	var required []string
	properties := schema["properties"].(map[string]any)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		jsonName := strings.Split(jsonTag, ",")[0]
		if jsonName == "" {
			jsonName = field.Name
		}

		fieldType := field.Type
		isOptional := false
		if fieldType.Kind() == reflect.Ptr {
			isOptional = true
			fieldType = fieldType.Elem()
		}

		prop := map[string]any{
			"type": goTypeToJSONType(fieldType),
		}

		if desc := field.Tag.Get("description"); desc != "" {
			prop["description"] = desc
		}

		if enum := field.Tag.Get("enum"); enum != "" {
			enumValues := strings.Split(enum, ",")
			for i := range enumValues {
				enumValues[i] = strings.TrimSpace(enumValues[i])
			}
			prop["enum"] = enumValues
		}

		if fieldType.Kind() == reflect.Slice {
			elemType := fieldType.Elem()
			prop["items"] = map[string]any{
				"type": goTypeToJSONType(elemType),
			}
		}

		properties[jsonName] = prop

		if !isOptional {
			required = append(required, jsonName)
		}
	}

	if len(required) > 0 {
		schema["required"] = required
	}

	schema["additionalProperties"] = false

	return schema
}

func goTypeToJSONType(t reflect.Type) string {
	switch t.Kind() {
	case reflect.String:
		return "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "integer"
	case reflect.Float32, reflect.Float64:
		return "number"
	case reflect.Bool:
		return "boolean"
	case reflect.Slice, reflect.Array:
		return "array"
	default:
		return "string"
	}
}
