package llm

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient(t *testing.T) {
	c := New("https://api.openai.com/v1", "sk-test", "gpt-4o-mini")
	if c.BaseURL != "https://api.openai.com/v1" {
		t.Errorf("unexpected BaseURL: %s", c.BaseURL)
	}
	if c.Model != "gpt-4o-mini" {
		t.Errorf("unexpected Model: %s", c.Model)
	}
	if c.APIKey != "sk-test" {
		t.Errorf("unexpected APIKey: %s", c.APIKey)
	}
}

func TestClientChatSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer sk-test" {
			t.Errorf("unexpected auth header")
		}
		resp := Response{
			Choices: []struct {
				Message Message `json:"message"`
			}{
				{Message: Message{Role: "assistant", Content: `{"triples":[]}`}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := New(server.URL, "sk-test", "gpt-4o-mini")
	resp, err := c.Chat([]Message{{Role: "user", Content: "hello"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Choices) == 0 {
		t.Fatal("expected at least one choice")
	}
	if resp.Choices[0].Message.Content != `{"triples":[]}` {
		t.Errorf("unexpected content: %s", resp.Choices[0].Message.Content)
	}
}

func TestClientChatAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]string{"message": "invalid model"},
		})
	}))
	defer server.Close()

	c := New(server.URL, "sk-test", "gpt-4o-mini")
	_, err := c.Chat([]Message{{Role: "user", Content: "hello"}})
	if err == nil {
		t.Fatal("expected error for API error response")
	}
}

func TestClientChatHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	defer server.Close()

	c := New(server.URL, "sk-test", "gpt-4o-mini")
	_, err := c.Chat([]Message{{Role: "user", Content: "hello"}})
	if err == nil {
		t.Fatal("expected error for HTTP 500")
	}
}

func TestClientChatRetry(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		resp := Response{
			Choices: []struct {
				Message Message `json:"message"`
			}{
				{Message: Message{Role: "assistant", Content: "ok"}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := New(server.URL, "sk-test", "gpt-4o-mini")
	c.HTTPClient.Timeout = 0
	resp, err := c.Chat([]Message{{Role: "user", Content: "hello"}})
	if err != nil {
		t.Fatalf("expected retry to succeed: %v", err)
	}
	if resp.Choices[0].Message.Content != "ok" {
		t.Errorf("unexpected content: %s", resp.Choices[0].Message.Content)
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestPromptNotEmpty(t *testing.T) {
	if AssessmentPrompt == "" {
		t.Error("AssessmentPrompt should not be empty")
	}
	if len(AssessmentPrompt) < 100 {
		t.Error("AssessmentPrompt seems too short")
	}
}
