package openai

import (
	"fmt"

	"github.com/marti-jorda-roca/gopher-ai/gopherai"
)

// CreateResponse sends a request to the OpenAI Responses API.
func (p *Provider) CreateResponse(req any) (any, error) {
	createReq, ok := req.(*CreateResponseRequest)
	if !ok {
		return nil, fmt.Errorf("invalid request type: expected *CreateResponseRequest")
	}

	var result Response
	var apiErr APIError

	resp, err := p.http.R().
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
func (p *Provider) CreateResponseTyped(req *CreateResponseRequest) (*Response, error) {
	resp, err := p.CreateResponse(req)
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
func (p *Provider) BuildRequest(input any, instructions string, tools []any) any {
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
		Instructions:    instructions,
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
