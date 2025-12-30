package gopherai

import (
	"fmt"
)

// CreateResponse sends a request to the OpenAI Responses API.
func (c *Client) CreateResponse(req *CreateResponseRequest) (*Response, error) {
	var result Response
	var apiErr APIError

	resp, err := c.http.R().
		SetBody(req).
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
