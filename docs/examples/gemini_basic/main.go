// Package main provides an example of using gopher-ai with Gemini and tool calls.
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

	result, err := myAgent.Run(context.Background(), "What's the weather like in Paris?")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Response: %s\n", result.Text)

	followUp, err := myAgent.Run(context.Background(), "And in London?", result.MessageHistory())
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Follow-up Response: %s\n", followUp.Text)
}
