package gopherai

// Provider defines the interface that all LLM providers must implement.
type Provider interface {
	CreateResponse(req any) (any, error)
}
