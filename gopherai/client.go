// Package gopherai provides a Go client for the OpenAI Responses API.
package gopherai

import (
	"github.com/go-resty/resty/v2"
)

const (
	defaultBaseURL = "https://api.openai.com/v1"
)

// Client is the OpenAI API client.
type Client struct {
	apiKey  string
	baseURL string
	http    *resty.Client
}

// NewClient creates a new OpenAI API client with the given API key.
func NewClient(apiKey string) *Client {
	c := &Client{
		apiKey:  apiKey,
		baseURL: defaultBaseURL,
		http:    resty.New(),
	}

	c.http.SetBaseURL(c.baseURL)
	c.http.SetHeader("Authorization", "Bearer "+apiKey)
	c.http.SetHeader("Content-Type", "application/json")

	return c
}

// SetBaseURL sets a custom base URL for API requests.
func (c *Client) SetBaseURL(url string) *Client {
	c.baseURL = url
	c.http.SetBaseURL(url)
	return c
}
