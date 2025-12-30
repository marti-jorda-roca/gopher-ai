package gopherai

import (
	"context"
	"fmt"

	"golang.org/x/sync/errgroup"
)

// Agent orchestrates interactions with an AI provider.
type Agent struct {
	provider     Provider
	tools        []Tool
	toolMap      map[string]Tool
	instructions string
}

// AgentOption configures an Agent.
type AgentOption func(*Agent)

// WithInstructions sets the system instructions for the agent.
func WithInstructions(instructions string) AgentOption {
	return func(a *Agent) {
		a.instructions = instructions
	}
}

// WithTools registers tools that the agent can use.
func WithTools(tools ...Tool) AgentOption {
	return func(a *Agent) {
		a.tools = tools
		for _, tool := range tools {
			a.toolMap[tool.Name] = tool
		}
	}
}

// NewAgent creates a new Agent with the given provider and options.
func NewAgent(provider Provider, opts ...AgentOption) *Agent {
	agent := &Agent{
		provider: provider,
		toolMap:  make(map[string]Tool),
	}

	for _, opt := range opts {
		opt(agent)
	}

	return agent
}

// Run executes the agent with the given prompt and returns the final response.
func (a *Agent) Run(ctx context.Context, prompt string) (string, error) {
	providerTools := make([]any, len(a.tools))
	for i, tool := range a.tools {
		providerTools[i] = a.provider.ConvertTool(tool)
	}

	var conversationHistory []any
	var input any = prompt
	maxIterations := 10

	for i := 0; i < maxIterations; i++ {
		req := a.provider.BuildRequest(input, a.instructions, providerTools)
		resp, err := a.provider.CreateResponse(ctx, req)
		if err != nil {
			return "", fmt.Errorf("failed to create response: %w", err)
		}

		toolCalls, err := a.provider.ExtractToolCalls(resp)
		if err != nil {
			return "", fmt.Errorf("failed to extract tool calls: %w", err)
		}

		if len(toolCalls) == 0 {
			text := a.provider.ExtractText(resp)
			return text, nil
		}

		if len(conversationHistory) == 0 {
			conversationHistory = append(conversationHistory, prompt)
		}

		type toolResult struct {
			call   ToolCall
			output string
		}

		results := make([]toolResult, len(toolCalls))
		g, _ := errgroup.WithContext(ctx)

		for i, call := range toolCalls {
			call := call
			i := i

			tool, ok := a.toolMap[call.Name]
			if !ok {
				return "", fmt.Errorf("unknown tool: %s", call.Name)
			}

			g.Go(func() error {
				result, err := tool.Handler(call.Arguments)
				if err != nil {
					return fmt.Errorf("tool %s failed: %w", call.Name, err)
				}
				results[i] = toolResult{call: call, output: result}
				return nil
			})
		}

		if err := g.Wait(); err != nil {
			return "", err
		}

		var inputItems []any
		for _, result := range results {
			callInputItem := a.provider.CreateFunctionCallInput(result.call)
			inputItems = append(inputItems, callInputItem)

			outputItem := a.provider.CreateFunctionCallOutput(result.call.CallID, result.output)
			inputItems = append(inputItems, outputItem)
		}

		conversationHistory = append(conversationHistory, inputItems...)
		input = conversationHistory
	}

	return "", fmt.Errorf("max iterations reached")
}
