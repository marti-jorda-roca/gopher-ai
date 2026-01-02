// Package main provides an example of using gopher-ai with Gemini streaming output.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/marti-jorda-roca/gopher-ai/gopherai"
	"github.com/marti-jorda-roca/gopher-ai/gopherai/gemini"
)

type GetWeatherParams struct {
	Location string `json:"location" description:"The city and state, e.g. San Francisco, CA"`
	Unit     string `json:"unit" description:"Temperature unit" enum:"celsius,fahrenheit"`
}

func GetWeather(params GetWeatherParams) (string, error) {
	weatherData := map[string]any{
		"location":    params.Location,
		"temperature": 22,
		"unit":        params.Unit,
		"conditions":  "partly cloudy",
		"humidity":    65,
	}
	data, _ := json.Marshal(weatherData)
	return string(data), nil
}

func main() {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		fmt.Println("GEMINI_API_KEY environment variable is required")
		os.Exit(1)
	}

	myProvider := gemini.NewProvider(apiKey).
		SetModel("gemini-2.5-flash")

	myTool := gopherai.NewTool("get_weather", "Get the current weather for a location", GetWeather)

	myAgent := gopherai.NewAgent(myProvider,
		gopherai.WithSystemPrompt("You are a helpful weather assistant."),
		gopherai.WithTools(myTool),
	)

	events, err := myAgent.RunStream(context.Background(), "What's the weather like in Paris?")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Print("Response: ")
	for event := range events {
		switch event.Type {
		case gopherai.StreamEventTypeTextDelta:
			fmt.Printf("\nText Delta: %+v\n", event.Delta)
		case gopherai.StreamEventTypeToolCall:
			fmt.Printf("\n[Calling tool: %s]\n", event.ToolCall.Name)
		case gopherai.StreamEventTypeError:
			fmt.Printf("\nError: %v\n", event.Error)
			os.Exit(1)
		case gopherai.StreamEventTypeDone:
			fmt.Println()
		}
	}
}
