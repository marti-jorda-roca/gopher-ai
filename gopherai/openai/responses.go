package openai

import (
	"fmt"
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
