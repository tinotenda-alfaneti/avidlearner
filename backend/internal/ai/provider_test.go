package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewOpenAIProvider(t *testing.T) {
	t.Run("creates provider with custom values", func(t *testing.T) {
		provider := NewOpenAIProvider("test-key", "gpt-3.5-turbo")

		if provider.apiKey != "test-key" {
			t.Errorf("expected apiKey 'test-key', got '%s'", provider.apiKey)
		}

		if provider.model != "gpt-3.5-turbo" {
			t.Errorf("expected model 'gpt-3.5-turbo', got '%s'", provider.model)
		}

		if provider.httpClient == nil {
			t.Error("expected httpClient to be initialized")
		}

		if provider.httpClient.Timeout != 30*time.Second {
			t.Errorf("expected timeout 30s, got %v", provider.httpClient.Timeout)
		}
	})

	t.Run("uses default model when empty", func(t *testing.T) {
		provider := NewOpenAIProvider("test-key", "")

		if provider.model != "gpt-4" {
			t.Errorf("expected default model 'gpt-4', got '%s'", provider.model)
		}
	})

	t.Run("uses environment variable when apiKey empty", func(t *testing.T) {
		// Note: This test depends on environment, may need to set OPENAI_API_KEY
		provider := NewOpenAIProvider("", "gpt-4")

		// Just verify it doesn't crash
		if provider == nil {
			t.Error("expected provider to be created")
		}
	})
}

func TestNewAnthropicProvider(t *testing.T) {
	t.Run("creates provider with custom values", func(t *testing.T) {
		provider := NewAnthropicProvider("test-key", "claude-3-opus")

		if provider.apiKey != "test-key" {
			t.Errorf("expected apiKey 'test-key', got '%s'", provider.apiKey)
		}

		if provider.model != "claude-3-opus" {
			t.Errorf("expected model 'claude-3-opus', got '%s'", provider.model)
		}

		if provider.httpClient == nil {
			t.Error("expected httpClient to be initialized")
		}
	})

	t.Run("uses default model when empty", func(t *testing.T) {
		provider := NewAnthropicProvider("test-key", "")

		if provider.model != "claude-3-5-sonnet-20241022" {
			t.Errorf("expected default model 'claude-3-5-sonnet-20241022', got '%s'", provider.model)
		}
	})
}

func TestGetProviderName(t *testing.T) {
	t.Run("OpenAI provider returns correct name", func(t *testing.T) {
		provider := NewOpenAIProvider("test-key", "gpt-4")
		if provider.GetProviderName() != "openai" {
			t.Errorf("expected 'openai', got '%s'", provider.GetProviderName())
		}
	})

	t.Run("Anthropic provider returns correct name", func(t *testing.T) {
		provider := NewAnthropicProvider("test-key", "claude-3-5-sonnet-20241022")
		if provider.GetProviderName() != "anthropic" {
			t.Errorf("expected 'anthropic', got '%s'", provider.GetProviderName())
		}
	})
}

func TestGetProvider(t *testing.T) {
	t.Run("returns OpenAI provider", func(t *testing.T) {
		provider, err := GetProvider("openai", "gpt-4")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if provider.GetProviderName() != "openai" {
			t.Errorf("expected openai provider, got %s", provider.GetProviderName())
		}
	})

	t.Run("returns Anthropic provider", func(t *testing.T) {
		provider, err := GetProvider("anthropic", "claude-3-5-sonnet-20241022")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if provider.GetProviderName() != "anthropic" {
			t.Errorf("expected anthropic provider, got %s", provider.GetProviderName())
		}
	})

	t.Run("returns error for unknown provider", func(t *testing.T) {
		_, err := GetProvider("unknown", "model")
		if err == nil {
			t.Error("expected error for unknown provider, got nil")
		}

		expectedMsg := "unknown AI provider: unknown"
		if err.Error() != expectedMsg {
			t.Errorf("expected error message '%s', got '%s'", expectedMsg, err.Error())
		}
	})
}

func TestOpenAIGenerateLesson(t *testing.T) {
	t.Run("returns error when API key not configured", func(t *testing.T) {
		provider := NewOpenAIProvider("", "gpt-4")
		provider.apiKey = "" // Ensure it's empty

		ctx := context.Background()
		_, err := provider.GenerateLesson(ctx, "Go", "concurrency")

		if err == nil {
			t.Error("expected error when API key not configured")
		}

		if err.Error() != "OpenAI API key not configured" {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("handles successful response", func(t *testing.T) {
		// Mock server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Authorization") != "Bearer test-api-key" {
				t.Error("expected Authorization header with Bearer token")
			}

			response := map[string]interface{}{
				"choices": []map[string]interface{}{
					{
						"message": map[string]string{
							"content": `{
								"title": "Go Concurrency",
								"category": "Go",
								"text": "Concurrency in Go",
								"explain": "Go provides goroutines and channels",
								"useCases": ["web servers", "data processing"],
								"tips": ["use channels", "avoid race conditions"]
							}`,
						},
					},
				},
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		provider := NewOpenAIProvider("test-api-key", "gpt-4")
		// Override the httpClient to use our test server
		provider.httpClient = &http.Client{
			Timeout:   5 * time.Second,
			Transport: &testTransport{baseURL: server.URL},
		}

		ctx := context.Background()
		lesson, err := provider.GenerateLesson(ctx, "Go", "concurrency")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if lesson.Title != "Go Concurrency" {
			t.Errorf("expected title 'Go Concurrency', got '%s'", lesson.Title)
		}

		if lesson.Category != "Go" {
			t.Errorf("expected category 'Go', got '%s'", lesson.Category)
		}

		if len(lesson.UseCases) != 2 {
			t.Errorf("expected 2 use cases, got %d", len(lesson.UseCases))
		}

		if len(lesson.Tips) != 2 {
			t.Errorf("expected 2 tips, got %d", len(lesson.Tips))
		}
	})

	t.Run("handles API error response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": {"message": "Invalid API key"}}`))
		}))
		defer server.Close()

		provider := NewOpenAIProvider("invalid-key", "gpt-4")
		provider.httpClient = &http.Client{
			Timeout:   5 * time.Second,
			Transport: &testTransport{baseURL: server.URL},
		}

		ctx := context.Background()
		_, err := provider.GenerateLesson(ctx, "Go", "concurrency")

		if err == nil {
			t.Error("expected error for API error response")
		}
	})
}

func TestAnthropicGenerateLesson(t *testing.T) {
	t.Run("returns error when API key not configured", func(t *testing.T) {
		provider := NewAnthropicProvider("", "claude-3-5-sonnet-20241022")
		provider.apiKey = ""

		ctx := context.Background()
		_, err := provider.GenerateLesson(ctx, "Python", "decorators")

		if err == nil {
			t.Error("expected error when API key not configured")
		}

		if err.Error() != "Anthropic API key not configured" {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("handles successful response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("x-api-key") != "test-api-key" {
				t.Error("expected x-api-key header")
			}

			if r.Header.Get("anthropic-version") != "2023-06-01" {
				t.Error("expected anthropic-version header")
			}

			response := map[string]interface{}{
				"content": []map[string]string{
					{
						"text": `{
							"title": "Python Decorators",
							"category": "Python",
							"text": "Decorators modify functions",
							"explain": "Python decorators are a powerful feature",
							"useCases": ["logging", "authentication"],
							"tips": ["use @decorator syntax", "understand closures"]
						}`,
					},
				},
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		provider := NewAnthropicProvider("test-api-key", "claude-3-5-sonnet-20241022")
		provider.httpClient = &http.Client{
			Timeout:   5 * time.Second,
			Transport: &testTransport{baseURL: server.URL},
		}

		ctx := context.Background()
		lesson, err := provider.GenerateLesson(ctx, "Python", "decorators")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if lesson.Title != "Python Decorators" {
			t.Errorf("expected title 'Python Decorators', got '%s'", lesson.Title)
		}

		if lesson.Category != "Python" {
			t.Errorf("expected category 'Python', got '%s'", lesson.Category)
		}
	})
}

// testTransport is a custom RoundTripper that redirects all requests to a test server
type testTransport struct {
	baseURL string
}

func (t *testTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Redirect all API calls to our test server
	newReq := req.Clone(req.Context())
	newReq.URL.Scheme = "http"
	newReq.URL.Host = req.URL.Host
	if t.baseURL != "" {
		// Parse the test server URL and use its host
		newReq.URL.Scheme = "http"
		// Extract host from baseURL (remove http://)
		host := t.baseURL[7:]
		newReq.URL.Host = host
	}
	return http.DefaultTransport.RoundTrip(newReq)
}
