package gopherai

// ToolCall represents a function call request from the AI.
type ToolCall struct {
	Name      string
	Arguments string
	CallID    string
}

// FunctionCallOutput represents the result of a function call.
type FunctionCallOutput struct {
	CallID string
	Output string
}

// Provider defines the interface for AI providers.
type Provider interface {
	CreateResponse(req any) (any, error)
	BuildRequest(input any, instructions string, tools []any) any
	ConvertTool(tool Tool) any
	ExtractToolCalls(resp any) ([]ToolCall, error)
	ExtractText(resp any) string
	CreateFunctionCallInput(call ToolCall) any
	CreateFunctionCallOutput(callID, output string) any
}
