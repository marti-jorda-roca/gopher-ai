// Package main provides an example of using the gopherai library.
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/yourusername/gopher-ai/gopherai"
)

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("OPENAI_API_KEY environment variable is required")
		os.Exit(1)
	}

	client := gopherai.NewClient(apiKey)

	weatherTool := gopherai.NewFunctionTool(
		"get_weather",
		"Get the current weather for a location",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"location": map[string]any{
					"type":        "string",
					"description": "The city and state, e.g. San Francisco, CA",
				},
				"unit": map[string]any{
					"type": "string",
					"enum": []string{"celsius", "fahrenheit"},
				},
			},
			"required":             []string{"location", "unit"},
			"additionalProperties": false,
		},
	)

	resp, err := client.CreateResponse(&gopherai.CreateResponseRequest{
		Model: "gpt-4.1",
		Input: "What's the weather like in Paris?",
		Tools: []gopherai.FunctionTool{weatherTool},
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Response ID: %s\n", resp.ID)
	fmt.Printf("Status: %s\n", resp.Status)

	toolCalls := resp.GetToolCalls()
	if len(toolCalls) > 0 {
		fmt.Println("\nTool Calls:")
		for _, call := range toolCalls {
			fmt.Printf("  - Function: %s\n", call.Name)
			fmt.Printf("    Arguments: %s\n", call.Arguments)
		}
	}

	if text := resp.GetOutputText(); text != "" {
		fmt.Printf("\nOutput Text: %s\n", text)
	}

	pretty, _ := json.MarshalIndent(resp, "", "  ")
	fmt.Printf("\nFull Response:\n%s\n", string(pretty))
}
