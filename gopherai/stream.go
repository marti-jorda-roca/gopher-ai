package gopherai

import "context"

// StreamEventType represents the type of streaming event.
type StreamEventType string

// Stream event type constants.
const (
	StreamEventTypeTextDelta StreamEventType = "text_delta"
	StreamEventTypeTextDone  StreamEventType = "text_done"
	StreamEventTypeToolCall  StreamEventType = "tool_call"
	StreamEventTypeError     StreamEventType = "error"
	StreamEventTypeDone      StreamEventType = "done"
)

// StreamEvent represents an event emitted during streaming.
type StreamEvent struct {
	Type     StreamEventType
	Delta    string
	Text     string
	ToolCall *ToolCall
	Error    error
}

// StreamProvider extends Provider with streaming capabilities.
type StreamProvider interface {
	Provider
	CreateResponseStream(ctx context.Context, req any) (<-chan StreamEvent, error)
}
