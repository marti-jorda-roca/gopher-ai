package openai

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/marti-jorda-roca/gopher-ai/gopherai"
)

// CreateResponse sends a request to the OpenAI Responses API.
func (p *Provider) CreateResponse(ctx context.Context, req any) (any, error) {
	createReq, ok := req.(*CreateResponseRequest)
	if !ok {
		return nil, fmt.Errorf("invalid request type: expected *CreateResponseRequest")
	}

	var result Response
	var apiErr APIError

	resp, err := p.http.R().
		SetContext(ctx).
		SetBody(createReq).
		SetResult(&result).
		SetError(&apiErr).
		Post("/responses")
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("API error: %s - %s", apiErr.Error.Type, apiErr.Error.Message)
	}

	return &result, nil
}

// CreateResponseTyped sends a request to the OpenAI Responses API with typed request and response.
func (p *Provider) CreateResponseTyped(ctx context.Context, req *CreateResponseRequest) (*Response, error) {
	resp, err := p.CreateResponse(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*Response), nil
}

// GetOutputText returns the text content from the response output.
func (r *Response) GetOutputText() string {
	for _, item := range r.Output {
		if item.Type == "message" {
			for _, content := range item.Content {
				if content.Type == "output_text" {
					return content.Text
				}
			}
		}
	}
	return ""
}

// GetToolCalls returns all function call outputs from the response.
func (r *Response) GetToolCalls() []OutputItem {
	var calls []OutputItem
	for _, item := range r.Output {
		if item.Type == "function_call" {
			calls = append(calls, item)
		}
	}
	return calls
}

// ConvertTool converts a gopherai.Tool to an OpenAI FunctionTool.
func (p *Provider) ConvertTool(tool gopherai.Tool) any {
	strict := true
	return FunctionTool{
		Type:        "function",
		Name:        tool.Name,
		Description: tool.Description,
		Parameters:  tool.Parameters,
		Strict:      &strict,
	}
}

// ExtractToolCalls extracts tool calls from a response.
func (p *Provider) ExtractToolCalls(resp any) ([]gopherai.ToolCall, error) {
	response, ok := resp.(*Response)
	if !ok {
		return nil, fmt.Errorf("invalid response type: expected *Response")
	}

	var calls []gopherai.ToolCall
	for _, item := range response.Output {
		if item.Type == "function_call" {
			calls = append(calls, gopherai.ToolCall{
				Name:      item.Name,
				Arguments: item.Arguments,
				CallID:    item.CallID,
			})
		}
	}
	return calls, nil
}

// ExtractText extracts text content from a response.
func (p *Provider) ExtractText(resp any) string {
	response, ok := resp.(*Response)
	if !ok {
		return ""
	}
	return response.GetOutputText()
}

// BuildRequest builds a CreateResponseRequest from the given parameters.
func (p *Provider) BuildRequest(input any, systemPrompt string, tools []any) any {
	functionTools := make([]FunctionTool, len(tools))
	for i, tool := range tools {
		functionTools[i] = tool.(FunctionTool)
	}

	convertedInput := input
	if inputItems, ok := input.([]any); ok {
		convertedItems := make([]InputItem, 0, len(inputItems))
		for _, item := range inputItems {
			if inputItem, ok := item.(InputItem); ok {
				convertedItems = append(convertedItems, inputItem)
			} else if inputStr, ok := item.(string); ok {
				convertedItems = append(convertedItems, InputItem{
					Type:    "message",
					Role:    "user",
					Content: inputStr,
				})
			}
		}
		convertedInput = convertedItems
	}

	req := &CreateResponseRequest{
		Model:           p.model,
		Input:           convertedInput,
		Instructions:    systemPrompt,
		Tools:           functionTools,
		Temperature:     p.temperature,
		MaxOutputTokens: p.maxTokens,
	}

	return req
}

// CreateFunctionCallInput creates a function call input item from a ToolCall.
func (p *Provider) CreateFunctionCallInput(call gopherai.ToolCall) any {
	return InputItem{
		Type:      "function_call",
		CallID:    call.CallID,
		Name:      call.Name,
		Arguments: call.Arguments,
	}
}

// CreateFunctionCallOutput creates a function call output item.
func (p *Provider) CreateFunctionCallOutput(callID, output string) any {
	return NewFunctionCallOutput(callID, output)
}

// CreateAssistantMessage creates an assistant message input item.
func (p *Provider) CreateAssistantMessage(text string) any {
	return InputItem{
		Type:    "message",
		Role:    "assistant",
		Content: text,
	}
}

// CreateResponseStream sends a streaming request to the OpenAI Responses API.
func (p *Provider) CreateResponseStream(ctx context.Context, req any) (<-chan gopherai.StreamEvent, error) {
	createReq, ok := req.(*CreateResponseRequest)
	if !ok {
		return nil, fmt.Errorf("invalid request type: expected *CreateResponseRequest")
	}

	stream := true
	createReq.Stream = &stream

	events := make(chan gopherai.StreamEvent, 100)

	resp, err := p.http.R().
		SetContext(ctx).
		SetBody(createReq).
		SetDoNotParseResponse(true).
		Post("/responses")
	if err != nil {
		close(events)
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.IsError() {
		body := resp.RawBody()
		defer func() { _ = body.Close() }()
		bodyBytes, _ := io.ReadAll(body)
		close(events)
		return nil, fmt.Errorf("API error: %s", string(bodyBytes))
	}

	go p.parseSSEStream(resp.RawBody(), events)

	return events, nil
}

func (p *Provider) parseSSEStream(body io.ReadCloser, events chan<- gopherai.StreamEvent) {
	defer close(events)
	defer func() { _ = body.Close() }()

	reader := bufio.NewReader(body)
	var fullText strings.Builder
	pendingToolCalls := make(map[string]*gopherai.ToolCall)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				events <- gopherai.StreamEvent{
					Type:  gopherai.StreamEventTypeError,
					Error: err,
				}
			}
			break
		}

		line = strings.TrimSpace(line)
		if line == "" || !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var eventData StreamEventData
		if err := json.Unmarshal([]byte(data), &eventData); err != nil {
			continue
		}

		switch eventData.Type {
		case "response.output_text.delta":
			fullText.WriteString(eventData.Delta)
			events <- gopherai.StreamEvent{
				Type:  gopherai.StreamEventTypeTextDelta,
				Delta: eventData.Delta,
			}

		case "response.output_text.done":
			events <- gopherai.StreamEvent{
				Type: gopherai.StreamEventTypeTextDone,
				Text: eventData.Text,
			}

		case "response.output_item.added":
			if eventData.Item != nil && eventData.Item.Type == "function_call" {
				pendingToolCalls[eventData.Item.ID] = &gopherai.ToolCall{
					CallID: eventData.Item.CallID,
					Name:   eventData.Item.Name,
				}
			}

		case "response.function_call_arguments.done":
			if tc, ok := pendingToolCalls[eventData.ItemID]; ok {
				tc.Arguments = eventData.Arguments
				if eventData.Name != "" {
					tc.Name = eventData.Name
				}
				events <- gopherai.StreamEvent{
					Type:     gopherai.StreamEventTypeToolCall,
					ToolCall: tc,
				}
				delete(pendingToolCalls, eventData.ItemID)
			}

		case "response.completed":
			events <- gopherai.StreamEvent{
				Type: gopherai.StreamEventTypeDone,
			}

		case "error":
			events <- gopherai.StreamEvent{
				Type:  gopherai.StreamEventTypeError,
				Error: fmt.Errorf("%s: %s", eventData.Code, eventData.Message),
			}
		}
	}
}
