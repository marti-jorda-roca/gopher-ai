// Package gemini provides a Google Gemini API provider implementation.
package gemini

import (
	"github.com/go-resty/resty/v2"
)

const (
	defaultBaseURL = "https://generativelanguage.googleapis.com/v1beta"
)

// Provider is the Google Gemini API provider.
type Provider struct {
	apiKey      string
	baseURL     string
	http        *resty.Client
	model       string
	temperature *float64
	maxTokens   *int
}

// NewProvider creates a new Gemini API provider with the given API key.
func NewProvider(apiKey string) *Provider {
	p := &Provider{
		apiKey:  apiKey,
		baseURL: defaultBaseURL,
		http:    resty.New(),
		model:   "gemini-2.5-flash",
	}

	p.http.SetBaseURL(p.baseURL)
	p.http.SetHeader("x-goog-api-key", apiKey)
	p.http.SetHeader("Content-Type", "application/json")

	return p
}

// SetModel sets the model to use for requests.
func (p *Provider) SetModel(model string) *Provider {
	p.model = model
	return p
}

// SetTemperature sets the temperature for requests.
func (p *Provider) SetTemperature(temperature float64) *Provider {
	p.temperature = &temperature
	return p
}

// SetMaxTokens sets the maximum output tokens for requests.
func (p *Provider) SetMaxTokens(maxTokens int) *Provider {
	p.maxTokens = &maxTokens
	return p
}

// SetBaseURL sets a custom base URL for API requests.
func (p *Provider) SetBaseURL(url string) *Provider {
	p.baseURL = url
	p.http.SetBaseURL(url)
	return p
}
