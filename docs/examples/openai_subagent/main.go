// Package main provides an example of using subagents with gopher-ai.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/marti-jorda-roca/gopher-ai/gopherai"
	"github.com/marti-jorda-roca/gopher-ai/gopherai/openai"
)

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("OPENAI_API_KEY environment variable is required")
		os.Exit(1)
	}

	provider := openai.NewProvider(apiKey).
		SetModel("gpt-4.1")

	translatorAgent := gopherai.NewAgent(provider,
		gopherai.WithSystemPrompt("You are a translator. Translate the given text to Spanish. Only respond with the translation, nothing else."),
	)

	summarizerAgent := gopherai.NewAgent(provider,
		gopherai.WithSystemPrompt("You are a summarizer. Summarize the given text in 1-2 sentences. Only respond with the summary, nothing else."),
	)

	orchestratorAgent := gopherai.NewAgent(provider,
		gopherai.WithSystemPrompt("You are an orchestrator that can delegate tasks to specialized agents. Use the available tools to help the user."),
		gopherai.WithTools(
			translatorAgent.AsTool("translate_to_spanish", "Translates text to Spanish"),
			summarizerAgent.AsTool("summarize", "Summarizes text in 1-2 sentences"),
		),
	)

	result, err := orchestratorAgent.Run(context.Background(),
		"Please summarize the following text and then translate the summary to Spanish: "+
			"The Go programming language was created at Google in 2007 by Robert Griesemer, Rob Pike, and Ken Thompson. "+
			"It was designed to improve programming productivity in an era of multicore processors, networked machines, and large codebases. "+
			"Go emphasizes simplicity, efficiency, and readability, with features like garbage collection, structural typing, and CSP-style concurrency.",
	)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Response: %s\n", result.Text)
}
