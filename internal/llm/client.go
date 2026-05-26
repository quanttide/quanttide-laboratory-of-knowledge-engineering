package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	DefaultTimeout  = 30 * time.Second
	MaxRetries      = 3
	InitialBackoff  = 100 * time.Millisecond
	MaxBackoff      = 2 * time.Second
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ResponseFormat struct {
	Type string `json:"type"`
}

type Request struct {
	Model       string          `json:"model"`
	Messages    []Message       `json:"messages"`
	Temperature float64         `json:"temperature"`
	MaxTokens   int             `json:"max_tokens"`
	Format      *ResponseFormat `json:"response_format,omitempty"`
}

type Response struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type Client struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
	Model      string
}

func New(baseURL, apiKey, model string) *Client {
	return &Client{
		BaseURL: baseURL,
		APIKey:  apiKey,
		Model:   model,
		HTTPClient: &http.Client{
			Timeout: DefaultTimeout,
		},
	}
}

func (c *Client) Chat(messages []Message) (*Response, error) {
	req := Request{
		Model:       c.Model,
		Messages:    messages,
		Temperature: 0.1,
		MaxTokens:   2048,
		Format:      &ResponseFormat{Type: "json_object"},
	}

	var lastErr error
	backoff := InitialBackoff

	for attempt := 0; attempt <= MaxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(backoff)
			backoff = time.Duration(float64(backoff) * 2.5)
			if backoff > MaxBackoff {
				backoff = MaxBackoff
			}
		}

		resp, err := c.doRequest(req)
		if err != nil {
			lastErr = err
			continue
		}
		if resp.Error != nil {
			lastErr = fmt.Errorf("api error: %s", resp.Error.Message)
			continue
		}
		if len(resp.Choices) == 0 {
			lastErr = fmt.Errorf("empty response: no choices")
			continue
		}
		return resp, nil
	}

	return nil, fmt.Errorf("all %d retries failed: %w", MaxRetries, lastErr)
}

func (c *Client) doRequest(req Request) (*Response, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)

	httpResp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if httpResp.StatusCode != 200 {
		return nil, fmt.Errorf("http %d: %s", httpResp.StatusCode, string(respBody))
	}

	var resp Response
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w (body: %s)", err, string(respBody))
	}

	return &resp, nil
}
