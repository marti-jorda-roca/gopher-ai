package gopherai

// Tool represents a function that can be called by the AI.
type Tool struct {
	Name        string
	Description string
	Parameters  map[string]any
	Handler     func(args string) (string, error)
}
