# gopher-ai

![Gopher Army](docs/imgs/gopher-army.svg)

A simple Go Agentic framework for building AI agents.

```go
package main

import (
	"github.com/marti-jorda-roca/gopher-ai/gopherai"
	"github.com/marti-jorda-roca/gopher-ai/gopherai/openai"
)

type GetWeatherParams struct {
	Location string `json:"location" description:"The city and state, e.g. San Francisco, CA"`
	Unit     string `json:"unit" description:"Temperature unit" enum:"celsius,fahrenheit"`
}

func GetWeather(params GetWeatherParams) (string, error) {
	return `{"temperature": 22, "conditions": "sunny"}`, nil
}

func main() {
	provider := openai.NewProvider("your-api-key").SetModel("gpt-4.1")
	tool := gopherai.NewTool("get_weather", "Get the current weather", GetWeather)
	agent := gopherai.NewAgent(provider,
		gopherai.WithSystemPrompt("You are a helpful weather assistant."),
		gopherai.WithTools(tool),
	)
	result, _ := agent.Run("What's the weather like in Paris?")
	println(result.Text)
}
```

## Installation

```bash
go get github.com/marti-jorda-roca/gopher-ai/gopherai
```

## Providers

| Provider | Import |
|----------|--------|
| OpenAI | `gopherai/openai` |
| Google Gemini | `gopherai/gemini` |

## Run Examples

### OpenAI

```bash
export OPENAI_API_KEY="your-api-key"
go run ./docs/examples/openai_basic
```

### Google Gemini

```bash
export GEMINI_API_KEY="your-api-key"
go run ./docs/examples/gemini_basic
```

## Build

```bash
go build ./gopherai/...
```

## Lint & Format

```bash
golangci-lint run --fix
```
