package gopherai

import (
	"context"
	"fmt"

	"golang.org/x/sync/errgroup"
)

// Agent orchestrates interactions with an AI provider.
type Agent struct {
	provider            Provider
	tools               []Tool
	toolMap             map[string]Tool
	systemPrompt        string
	conversationHistory []any
}

// AgentOption configures an Agent.
type AgentOption func(*Agent)

// WithSystemPrompt sets the system prompt for the agent.
func WithSystemPrompt(systemPrompt string) AgentOption {
	return func(a *Agent) {
		a.systemPrompt = systemPrompt
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

// WithConversationHistory sets the initial conversation history for the agent.
func WithConversationHistory(history []any) AgentOption {
	return func(a *Agent) {
		a.conversationHistory = history
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

// RunResult holds the result of an agent run, including the response text and conversation history.
type RunResult struct {
	Text    string
	history []any
}

// MessageHistory returns the conversation history from this run.
func (r *RunResult) MessageHistory() []any {
	return r.history
}

// Run executes the agent with the given prompt and optional conversation history, returning the final response and updated history.
func (a *Agent) Run(ctx context.Context, prompt string, history ...[]any) (*RunResult, error) {
	providerTools := make([]any, len(a.tools))
	for i, tool := range a.tools {
		providerTools[i] = a.provider.ConvertTool(tool)
	}

	var conversationHistory []any
	if len(history) > 0 && len(history[0]) > 0 {
		conversationHistory = make([]any, len(history[0]))
		copy(conversationHistory, history[0])
	} else if len(a.conversationHistory) > 0 {
		conversationHistory = make([]any, len(a.conversationHistory))
		copy(conversationHistory, a.conversationHistory)
	}

	conversationHistory = append(conversationHistory, prompt)
	var input any = prompt
	if len(conversationHistory) > 1 {
		input = conversationHistory
	}
	maxIterations := 10

	for i := 0; i < maxIterations; i++ {
		req := a.provider.BuildRequest(input, a.systemPrompt, providerTools)
		resp, err := a.provider.CreateResponse(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("failed to create response: %w", err)
		}

		toolCalls, err := a.provider.ExtractToolCalls(resp)
		if err != nil {
			return nil, fmt.Errorf("failed to extract tool calls: %w", err)
		}

		if len(toolCalls) == 0 {
			text := a.provider.ExtractText(resp)
			assistantMessage := a.provider.CreateAssistantMessage(text)
			conversationHistory = append(conversationHistory, assistantMessage)
			return &RunResult{
				Text:    text,
				history: conversationHistory,
			}, nil
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
				return nil, fmt.Errorf("unknown tool: %s", call.Name)
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
			return nil, err
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

	return nil, fmt.Errorf("max iterations reached")
}

// RunStream executes the agent with streaming output, returning a channel of StreamEvents.
func (a *Agent) RunStream(ctx context.Context, prompt string, history ...[]any) (<-chan StreamEvent, error) {
	streamProvider, ok := a.provider.(StreamProvider)
	if !ok {
		return nil, fmt.Errorf("provider does not support streaming")
	}

	providerTools := make([]any, len(a.tools))
	for i, tool := range a.tools {
		providerTools[i] = a.provider.ConvertTool(tool)
	}

	var conversationHistory []any
	if len(history) > 0 && len(history[0]) > 0 {
		conversationHistory = make([]any, len(history[0]))
		copy(conversationHistory, history[0])
	} else if len(a.conversationHistory) > 0 {
		conversationHistory = make([]any, len(a.conversationHistory))
		copy(conversationHistory, a.conversationHistory)
	}

	conversationHistory = append(conversationHistory, prompt)
	var input any = prompt
	if len(conversationHistory) > 1 {
		input = conversationHistory
	}

	outEvents := make(chan StreamEvent, 100)

	go a.runStreamLoop(ctx, streamProvider, input, conversationHistory, providerTools, outEvents)

	return outEvents, nil
}

func (a *Agent) runStreamLoop(
	ctx context.Context,
	streamProvider StreamProvider,
	input any,
	conversationHistory []any,
	providerTools []any,
	outEvents chan<- StreamEvent,
) {
	defer close(outEvents)

	maxIterations := 10
	for i := 0; i < maxIterations; i++ {
		req := a.provider.BuildRequest(input, a.systemPrompt, providerTools)
		events, err := streamProvider.CreateResponseStream(ctx, req)
		if err != nil {
			outEvents <- StreamEvent{
				Type:  StreamEventTypeError,
				Error: fmt.Errorf("failed to create response stream: %w", err),
			}
			return
		}

		var toolCalls []ToolCall
		var fullText string

		for event := range events {
			switch event.Type {
			case StreamEventTypeTextDelta:
				outEvents <- event

			case StreamEventTypeTextDone:
				fullText = event.Text
				outEvents <- event

			case StreamEventTypeToolCall:
				if event.ToolCall != nil {
					toolCalls = append(toolCalls, *event.ToolCall)
					outEvents <- event
				}

			case StreamEventTypeError:
				outEvents <- event
				return

			case StreamEventTypeDone:
			}
		}

		if len(toolCalls) == 0 {
			outEvents <- StreamEvent{Type: StreamEventTypeDone}
			return
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
				outEvents <- StreamEvent{
					Type:  StreamEventTypeError,
					Error: fmt.Errorf("unknown tool: %s", call.Name),
				}
				return
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
			outEvents <- StreamEvent{
				Type:  StreamEventTypeError,
				Error: err,
			}
			return
		}

		var inputItems []any
		for _, result := range results {
			callInputItem := a.provider.CreateFunctionCallInput(result.call)
			inputItems = append(inputItems, callInputItem)

			outputItem := a.provider.CreateFunctionCallOutput(result.call.CallID, result.output)
			inputItems = append(inputItems, outputItem)
		}

		if fullText != "" {
			assistantMessage := a.provider.CreateAssistantMessage(fullText)
			conversationHistory = append(conversationHistory, assistantMessage)
		}

		conversationHistory = append(conversationHistory, inputItems...)
		input = conversationHistory
	}

	outEvents <- StreamEvent{
		Type:  StreamEventTypeError,
		Error: fmt.Errorf("max iterations reached"),
	}
}
