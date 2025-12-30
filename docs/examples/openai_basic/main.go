// Package main provides an example of using gopher-ai with tool calls.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/marti-jorda-roca/gopher-ai/gopherai"
	"github.com/marti-jorda-roca/gopher-ai/gopherai/openai"
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
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("OPENAI_API_KEY environment variable is required")
		os.Exit(1)
	}

	myProvider := openai.NewProvider(apiKey).
		SetModel("gpt-4.1")

	myTool := gopherai.NewTool("get_weather", "Get the current weather for a location", GetWeather)

	myAgent := gopherai.NewAgent(myProvider,
		gopherai.WithInstructions("You are a helpful weather assistant."),
		gopherai.WithTools(myTool),
	)

	response, err := myAgent.Run(context.Background(), "What's the weather like in Paris?")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Response: %s\n", response)
}
