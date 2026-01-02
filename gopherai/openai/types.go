package openai

// FunctionTool represents a function tool that can be called by the model.
type FunctionTool struct {
	Type        string         `json:"type"`
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Parameters  map[string]any `json:"parameters,omitempty"`
	Strict      *bool          `json:"strict,omitempty"`
}

// NewFunctionTool creates a new function tool with the given name, description, and parameters.
func NewFunctionTool(name, description string, parameters map[string]any) FunctionTool {
	strict := true
	return FunctionTool{
		Type:        "function",
		Name:        name,
		Description: description,
		Parameters:  parameters,
		Strict:      &strict,
	}
}

// InputMessage represents an input message in a conversation.
type InputMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// InputItem represents a generic input item for the conversation.
type InputItem struct {
	Type      string `json:"type"`
	Role      string `json:"role,omitempty"`
	Content   string `json:"content,omitempty"`
	CallID    string `json:"call_id,omitempty"`
	Name      string `json:"name,omitempty"`
	Arguments string `json:"arguments,omitempty"`
	Output    string `json:"output,omitempty"`
}

// NewFunctionCallOutput creates a new function call output item.
func NewFunctionCallOutput(callID, output string) InputItem {
	return InputItem{
		Type:   "function_call_output",
		CallID: callID,
		Output: output,
	}
}

// ToInputItem converts an OutputItem (function_call) to an InputItem for continuing the conversation.
func (o *OutputItem) ToInputItem() InputItem {
	return InputItem{
		Type:      o.Type,
		CallID:    o.CallID,
		Name:      o.Name,
		Arguments: o.Arguments,
	}
}

// CreateResponseRequest represents a request to create a response.
type CreateResponseRequest struct {
	Model             string         `json:"model"`
	Input             any            `json:"input"`
	Instructions      string         `json:"instructions,omitempty"`
	Tools             []FunctionTool `json:"tools,omitempty"`
	ToolChoice        any            `json:"tool_choice,omitempty"`
	ParallelToolCalls *bool          `json:"parallel_tool_calls,omitempty"`
	Temperature       *float64       `json:"temperature,omitempty"`
	MaxOutputTokens   *int           `json:"max_output_tokens,omitempty"`
	Store             *bool          `json:"store,omitempty"`
	Stream            *bool          `json:"stream,omitempty"`
}

// Response represents the API response from a create response request.
type Response struct {
	ID                 string             `json:"id"`
	Object             string             `json:"object"`
	CreatedAt          int64              `json:"created_at"`
	Status             string             `json:"status"`
	Error              *ResponseError     `json:"error"`
	IncompleteDetails  *IncompleteDetails `json:"incomplete_details"`
	Instructions       *string            `json:"instructions"`
	MaxOutputTokens    *int               `json:"max_output_tokens"`
	Model              string             `json:"model"`
	Output             []OutputItem       `json:"output"`
	ParallelToolCalls  bool               `json:"parallel_tool_calls"`
	PreviousResponseID *string            `json:"previous_response_id"`
	Reasoning          *Reasoning         `json:"reasoning"`
	Store              bool               `json:"store"`
	Temperature        float64            `json:"temperature"`
	ToolChoice         any                `json:"tool_choice"`
	Tools              []FunctionTool     `json:"tools"`
	TopP               float64            `json:"top_p"`
	Truncation         string             `json:"truncation"`
	Usage              *Usage             `json:"usage"`
	User               *string            `json:"user"`
	Metadata           map[string]any     `json:"metadata"`
}

// ResponseError represents an error in the response.
type ResponseError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// IncompleteDetails provides details when a response is incomplete.
type IncompleteDetails struct {
	Reason string `json:"reason"`
}

// OutputItem represents an item in the response output.
type OutputItem struct {
	Type      string        `json:"type"`
	ID        string        `json:"id"`
	Status    string        `json:"status"`
	Role      string        `json:"role,omitempty"`
	Content   []ContentItem `json:"content,omitempty"`
	Name      string        `json:"name,omitempty"`
	Arguments string        `json:"arguments,omitempty"`
	CallID    string        `json:"call_id,omitempty"`
}

// ContentItem represents a content item within an output item.
type ContentItem struct {
	Type        string       `json:"type"`
	Text        string       `json:"text,omitempty"`
	Annotations []Annotation `json:"annotations,omitempty"`
}

// Annotation represents an annotation on content.
type Annotation struct {
	Type  string `json:"type"`
	Text  string `json:"text,omitempty"`
	URL   string `json:"url,omitempty"`
	Title string `json:"title,omitempty"`
}

// Reasoning represents reasoning information in the response.
type Reasoning struct {
	Effort  *string `json:"effort"`
	Summary *string `json:"summary"`
}

// Usage represents token usage information.
type Usage struct {
	InputTokens         int                 `json:"input_tokens"`
	InputTokensDetails  InputTokensDetails  `json:"input_tokens_details"`
	OutputTokens        int                 `json:"output_tokens"`
	OutputTokensDetails OutputTokensDetails `json:"output_tokens_details"`
	TotalTokens         int                 `json:"total_tokens"`
}

// InputTokensDetails provides details about input token usage.
type InputTokensDetails struct {
	CachedTokens int `json:"cached_tokens"`
}

// OutputTokensDetails provides details about output token usage.
type OutputTokensDetails struct {
	ReasoningTokens int `json:"reasoning_tokens"`
}

// APIError represents an error response from the API.
type APIError struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Param   string `json:"param"`
		Code    string `json:"code"`
	} `json:"error"`
}

// StreamEventData represents the data payload of a streaming event.
type StreamEventData struct {
	Type           string             `json:"type"`
	SequenceNumber int                `json:"sequence_number"`
	Response       *Response          `json:"response,omitempty"`
	OutputIndex    int                `json:"output_index,omitempty"`
	ItemID         string             `json:"item_id,omitempty"`
	ContentIndex   int                `json:"content_index,omitempty"`
	Delta          string             `json:"delta,omitempty"`
	Text           string             `json:"text,omitempty"`
	Item           *StreamOutputItem  `json:"item,omitempty"`
	Part           *StreamContentPart `json:"part,omitempty"`
	Name           string             `json:"name,omitempty"`
	Arguments      string             `json:"arguments,omitempty"`
	Code           string             `json:"code,omitempty"`
	Message        string             `json:"message,omitempty"`
}

// StreamOutputItem represents an item in the streaming response output.
type StreamOutputItem struct {
	Type      string        `json:"type"`
	ID        string        `json:"id"`
	Status    string        `json:"status"`
	Role      string        `json:"role,omitempty"`
	Content   []ContentItem `json:"content,omitempty"`
	Name      string        `json:"name,omitempty"`
	Arguments string        `json:"arguments,omitempty"`
	CallID    string        `json:"call_id,omitempty"`
}

// StreamContentPart represents a content part in the streaming response.
type StreamContentPart struct {
	Type        string       `json:"type"`
	Text        string       `json:"text,omitempty"`
	Annotations []Annotation `json:"annotations,omitempty"`
}
