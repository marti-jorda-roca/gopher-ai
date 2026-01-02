// Package main provides an example of running multiple gopher-ai agents in parallel.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

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

type AgentResult struct {
	Query    string
	Response string
	Err      error
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

	queries := []string{
		"What's the weather like in Paris?",
		"What's the weather like in Tokyo?",
	}

	results := make(chan AgentResult, len(queries))
	var wg sync.WaitGroup

	for _, query := range queries {
		wg.Add(1)
		go func(q string) {
			defer wg.Done()

			agent := gopherai.NewAgent(myProvider,
				gopherai.WithSystemPrompt("You are a helpful weather assistant."),
				gopherai.WithTools(myTool),
			)

			result, err := agent.Run(context.Background(), q)
			if err != nil {
				results <- AgentResult{Query: q, Err: err}
				return
			}
			results <- AgentResult{Query: q, Response: result.Text}
		}(query)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for res := range results {
		if res.Err != nil {
			fmt.Printf("Query: %s\nError: %v\n\n", res.Query, res.Err)
			continue
		}
		fmt.Printf("Query: %s\nResponse: %s\n\n", res.Query, res.Response)
	}
}

