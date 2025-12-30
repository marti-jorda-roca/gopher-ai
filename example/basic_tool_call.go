// Package main provides an example of using the gopherai library.
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/marti-jorda-roca/gopher-ai/gopherai/openai"
)

func getWeather(location, unit string) string {
	weatherData := map[string]any{
		"location":    location,
		"temperature": 22,
		"unit":        unit,
		"conditions":  "partly cloudy",
		"humidity":    65,
	}
	data, _ := json.Marshal(weatherData)
	return string(data)
}

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("OPENAI_API_KEY environment variable is required")
		os.Exit(1)
	}

	client := openai.NewProvider(apiKey)

	weatherTool := openai.NewFunctionTool(
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

	resp, err := client.CreateResponseTyped(&openai.CreateResponseRequest{
		Model: "gpt-4.1",
		Input: "What's the weather like in Paris?",
		Tools: []openai.FunctionTool{weatherTool},
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Response ID: %s\n", resp.ID)
	fmt.Printf("Status: %s\n", resp.Status)

	toolCalls := resp.GetToolCalls()
	if len(toolCalls) == 0 {
		if text := resp.GetOutputText(); text != "" {
			fmt.Printf("\nOutput Text: %s\n", text)
		}
		return
	}

	fmt.Println("\nTool Calls:")
	var input []openai.InputItem
	for _, call := range toolCalls {
		fmt.Printf("  - Function: %s\n", call.Name)
		fmt.Printf("    Arguments: %s\n", call.Arguments)

		input = append(input, call.ToInputItem())

		if call.Name == "get_weather" {
			var args struct {
				Location string `json:"location"`
				Unit     string `json:"unit"`
			}
			if err := json.Unmarshal([]byte(call.Arguments), &args); err != nil {
				fmt.Printf("Error parsing arguments: %v\n", err)
				os.Exit(1)
			}

			result := getWeather(args.Location, args.Unit)
			fmt.Printf("    Result: %s\n", result)
			input = append(input, openai.NewFunctionCallOutput(call.CallID, result))
		}
	}

	finalResp, err := client.CreateResponseTyped(&openai.CreateResponseRequest{
		Model: "gpt-4.1",
		Input: input,
		Tools: []openai.FunctionTool{weatherTool},
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nFinal Response ID: %s\n", finalResp.ID)
	if text := finalResp.GetOutputText(); text != "" {
		fmt.Printf("Output Text: %s\n", text)
	}

	pretty, _ := json.MarshalIndent(finalResp, "", "  ")
	fmt.Printf("\nFull Response:\n%s\n", string(pretty))
}
