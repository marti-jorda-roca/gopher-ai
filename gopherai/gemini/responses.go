package gemini

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/marti-jorda-roca/gopher-ai/gopherai"
)

// CreateResponse sends a request to the Gemini API.
func (p *Provider) CreateResponse(ctx context.Context, req any) (any, error) {
	generateReq, ok := req.(*GenerateContentRequest)
	if !ok {
		return nil, fmt.Errorf("invalid request type: expected *GenerateContentRequest")
	}

	var result GenerateContentResponse
	var apiErr APIError

	endpoint := fmt.Sprintf("/models/%s:generateContent", p.model)
	resp, err := p.http.R().
		SetContext(ctx).
		SetBody(generateReq).
		SetResult(&result).
		SetError(&apiErr).
		Post(endpoint)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("API error: %d - %s", apiErr.Error.Code, apiErr.Error.Message)
	}

	return &result, nil
}

// ConvertTool converts a gopherai.Tool to a Gemini FunctionDeclaration.
func (p *Provider) ConvertTool(tool gopherai.Tool) any {
	return FunctionDeclaration{
		Name:        tool.Name,
		Description: tool.Description,
		Parameters:  convertParameters(tool.Parameters),
	}
}

// convertParameters converts a map[string]any parameters to FunctionParameters.
func convertParameters(params map[string]any) *FunctionParameters {
	if params == nil {
		return &FunctionParameters{
			Type:       "object",
			Properties: make(map[string]PropertySchema),
		}
	}

	properties := make(map[string]PropertySchema)
	required := []string{}

	if requiredList, ok := params["required"].([]any); ok {
		for _, req := range requiredList {
			if reqStr, ok := req.(string); ok {
				required = append(required, reqStr)
			}
		}
	}

	if props, ok := params["properties"].(map[string]any); ok {
		for name, prop := range props {
			if propMap, ok := prop.(map[string]any); ok {
				schema := PropertySchema{}
				if typ, ok := propMap["type"].(string); ok {
					schema.Type = typ
				}
				if desc, ok := propMap["description"].(string); ok {
					schema.Description = desc
				}
				if items, ok := propMap["items"].(map[string]any); ok {
					if itemType, ok := items["type"].(string); ok {
						schema.Items = &PropertySchema{Type: itemType}
					}
				}
				if enum, ok := propMap["enum"].([]any); ok {
					enumStrs := make([]string, len(enum))
					for i, e := range enum {
						if eStr, ok := e.(string); ok {
							enumStrs[i] = eStr
						}
					}
					schema.Enum = enumStrs
				}
				properties[name] = schema
			}
		}
	}

	return &FunctionParameters{
		Type:       "object",
		Properties: properties,
		Required:   required,
	}
}

// ExtractToolCalls extracts tool calls from a response.
func (p *Provider) ExtractToolCalls(resp any) ([]gopherai.ToolCall, error) {
	response, ok := resp.(*GenerateContentResponse)
	if !ok {
		return nil, fmt.Errorf("invalid response type: expected *GenerateContentResponse")
	}

	var calls []gopherai.ToolCall
	for _, candidate := range response.Candidates {
		for _, part := range candidate.Content.Parts {
			if part.FunctionCall != nil {
				argsJSON, err := json.Marshal(part.FunctionCall.Args)
				if err != nil {
					return nil, fmt.Errorf("failed to marshal function call args: %w", err)
				}
				calls = append(calls, gopherai.ToolCall{
					Name:      part.FunctionCall.Name,
					Arguments: string(argsJSON),
					CallID:    fmt.Sprintf("%d", len(calls)),
				})
			}
		}
	}
	return calls, nil
}

// ExtractText extracts text content from a response.
func (p *Provider) ExtractText(resp any) string {
	response, ok := resp.(*GenerateContentResponse)
	if !ok {
		return ""
	}

	for _, candidate := range response.Candidates {
		for _, part := range candidate.Content.Parts {
			if part.Text != "" {
				return part.Text
			}
		}
	}
	return ""
}

// BuildRequest builds a GenerateContentRequest from the given parameters.
func (p *Provider) BuildRequest(input any, systemPrompt string, tools []any) any {
	var contents []Content

	if inputStr, ok := input.(string); ok {
		contents = []Content{
			{
				Role: "user",
				Parts: []Part{
					{Text: inputStr},
				},
			},
		}
	} else if inputItems, ok := input.([]any); ok {
		contents = make([]Content, 0, len(inputItems))
		for _, item := range inputItems {
			if content, ok := item.(Content); ok {
				contents = append(contents, content)
			} else if inputStr, ok := item.(string); ok {
				contents = append(contents, Content{
					Role: "user",
					Parts: []Part{
						{Text: inputStr},
					},
				})
			}
		}
		contents = p.fixFunctionResponseNames(contents)
	}

	functionDeclarations := make([]FunctionDeclaration, len(tools))
	for i, tool := range tools {
		functionDeclarations[i] = tool.(FunctionDeclaration)
	}

	var toolsList []Tool
	if len(functionDeclarations) > 0 {
		toolsList = []Tool{
			{FunctionDeclarations: functionDeclarations},
		}
	}

	req := &GenerateContentRequest{
		Contents: contents,
		Tools:    toolsList,
		GenerationConfig: &GenerationConfig{
			Temperature:     p.temperature,
			MaxOutputTokens: p.maxTokens,
		},
	}

	if systemPrompt != "" {
		req.SystemInstruction = &SystemInstruction{
			Parts: []Part{
				{Text: systemPrompt},
			},
		}
	}

	return req
}

// CreateFunctionCallInput creates a function call input content from a ToolCall.
func (p *Provider) CreateFunctionCallInput(call gopherai.ToolCall) any {
	var args map[string]any
	if err := json.Unmarshal([]byte(call.Arguments), &args); err != nil {
		args = make(map[string]any)
	}

	return Content{
		Role: "model",
		Parts: []Part{
			{
				FunctionCall: &FunctionCall{
					Name: call.Name,
					Args: args,
				},
			},
		},
	}
}

// CreateFunctionCallOutput creates a function call output content.
func (p *Provider) CreateFunctionCallOutput(callID, output string) any {
	var response map[string]any
	if err := json.Unmarshal([]byte(output), &response); err != nil {
		response = map[string]any{"result": output}
	}

	return Content{
		Role: "user",
		Parts: []Part{
			{
				FunctionResponse: &FunctionResponse{
					Name:     callID,
					Response: response,
				},
			},
		},
	}
}

// CreateAssistantMessage creates an assistant message content.
func (p *Provider) CreateAssistantMessage(text string) any {
	return Content{
		Role: "model",
		Parts: []Part{
			{Text: text},
		},
	}
}

// fixFunctionResponseNames fixes function response names by matching them
// with the preceding function calls in the conversation.
func (p *Provider) fixFunctionResponseNames(contents []Content) []Content {
	result := make([]Content, len(contents))
	copy(result, contents)

	for i := 0; i < len(result); i++ {
		if result[i].Role == "function" {
			for j := i - 1; j >= 0; j-- {
				if result[j].Role == "model" {
					for _, part := range result[j].Parts {
						if part.FunctionCall != nil {
							for k := range result[i].Parts {
								if result[i].Parts[k].FunctionResponse != nil {
									result[i].Parts[k].FunctionResponse.Name = part.FunctionCall.Name
								}
							}
							break
						}
					}
					break
				}
			}
		}
	}

	return result
}

// CreateResponseStream sends a streaming request to the Gemini API.
func (p *Provider) CreateResponseStream(ctx context.Context, req any) (<-chan gopherai.StreamEvent, error) {
	generateReq, ok := req.(*GenerateContentRequest)
	if !ok {
		return nil, fmt.Errorf("invalid request type: expected *GenerateContentRequest")
	}

	events := make(chan gopherai.StreamEvent, 100)

	endpoint := fmt.Sprintf("/models/%s:streamGenerateContent?alt=sse", p.model)
	resp, err := p.http.R().
		SetContext(ctx).
		SetBody(generateReq).
		SetDoNotParseResponse(true).
		Post(endpoint)
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

	go p.parseGeminiStream(resp.RawBody(), events)

	return events, nil
}

func (p *Provider) parseGeminiStream(body io.ReadCloser, events chan<- gopherai.StreamEvent) {
	defer close(events)
	defer func() { _ = body.Close() }()

	reader := bufio.NewReader(body)
	var fullText strings.Builder

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

		var response GenerateContentResponse
		if err := json.Unmarshal([]byte(data), &response); err != nil {
			continue
		}

		for _, candidate := range response.Candidates {
			for _, part := range candidate.Content.Parts {
				if part.Text != "" {
					fullText.WriteString(part.Text)
					events <- gopherai.StreamEvent{
						Type:  gopherai.StreamEventTypeTextDelta,
						Delta: part.Text,
					}
				}

				if part.FunctionCall != nil {
					argsJSON, err := json.Marshal(part.FunctionCall.Args)
					if err != nil {
						continue
					}
					events <- gopherai.StreamEvent{
						Type: gopherai.StreamEventTypeToolCall,
						ToolCall: &gopherai.ToolCall{
							Name:      part.FunctionCall.Name,
							Arguments: string(argsJSON),
							CallID:    part.FunctionCall.Name,
						},
					}
				}
			}

			if candidate.FinishReason == "STOP" {
				if fullText.Len() > 0 {
					events <- gopherai.StreamEvent{
						Type: gopherai.StreamEventTypeTextDone,
						Text: fullText.String(),
					}
				}
				events <- gopherai.StreamEvent{
					Type: gopherai.StreamEventTypeDone,
				}
				return
			}
		}
	}
}
