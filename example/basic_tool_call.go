// Package main provides an example of using gopher-ai with tool calls.
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/marti-jorda-roca/gopher-ai/gopherai"
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

	myProvider := openai.NewProvider(apiKey).
		SetModel("gpt-4.1")

	myTool := gopherai.Tool{
		Name:        "get_weather",
		Description: "Get the current weather for a location",
		Parameters: map[string]any{
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
		Handler: func(args string) (string, error) {
			var params struct {
				Location string `json:"location"`
				Unit     string `json:"unit"`
			}
			if err := json.Unmarshal([]byte(args), &params); err != nil {
				return "", fmt.Errorf("failed to parse arguments: %w", err)
			}
			return getWeather(params.Location, params.Unit), nil
		},
	}

	myAgent := gopherai.NewAgent(myProvider,
		gopherai.WithInstructions("You are a helpful weather assistant."),
		gopherai.WithTools(myTool),
	)

	response, err := myAgent.Run("What's the weather like in Paris?")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Response: %s\n", response)
}
