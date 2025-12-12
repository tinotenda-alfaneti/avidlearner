package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// Lesson represents a generated lesson
type Lesson struct {
	Title    string   `json:"title"`
	Category string   `json:"category"`
	Text     string   `json:"text"`
	Explain  string   `json:"explain"`
	UseCases []string `json:"useCases"`
	Tips     []string `json:"tips"`
}

// Provider defines the interface for AI lesson generation
type Provider interface {
	GenerateLesson(ctx context.Context, category, topic string) (*Lesson, error)
	GetProviderName() string
}

// OpenAIProvider implements lesson generation using OpenAI API
type OpenAIProvider struct {
	apiKey     string
	model      string
	httpClient *http.Client
}

// AnthropicProvider implements lesson generation using Anthropic API
type AnthropicProvider struct {
	apiKey     string
	model      string
	httpClient *http.Client
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(apiKey, model string) *OpenAIProvider {
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}
	if model == "" {
		model = "gpt-4"
	}
	return &OpenAIProvider{
		apiKey: apiKey,
		model:  model,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// NewAnthropicProvider creates a new Anthropic provider
func NewAnthropicProvider(apiKey, model string) *AnthropicProvider {
	if apiKey == "" {
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
	}
	if model == "" {
		model = "claude-3-5-sonnet-20241022"
	}
	return &AnthropicProvider{
		apiKey: apiKey,
		model:  model,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (p *OpenAIProvider) GetProviderName() string {
	return "openai"
}

func (p *AnthropicProvider) GetProviderName() string {
	return "anthropic"
}

// GenerateLesson generates a lesson using OpenAI
func (p *OpenAIProvider) GenerateLesson(ctx context.Context, category, topic string) (*Lesson, error) {
	if p.apiKey == "" {
		return nil, errors.New("OpenAI API key not configured")
	}

	prompt := fmt.Sprintf(`Generate a software engineering lesson in JSON format with the following structure:
{
  "title": "concise title",
  "category": "%s",
  "text": "1-2 sentence overview",
  "explain": "detailed explanation (2-3 sentences)",
  "useCases": ["use case 1", "use case 2", "use case 3"],
  "tips": ["tip 1", "tip 2", "tip 3"]
}

Topic: %s
Category: %s

Focus on practical, actionable content suitable for intermediate to advanced engineers. Keep it concise but informative.`, category, topic, category)

	reqBody := map[string]interface{}{
		"model": p.model,
		"messages": []map[string]string{
			{"role": "system", "content": "You are an expert software engineering instructor. Generate educational content in valid JSON format only."},
			{"role": "user", "content": prompt},
		},
		"temperature":     0.7,
		"response_format": map[string]string{"type": "json_object"},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("OpenAI API error %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if len(result.Choices) == 0 {
		return nil, errors.New("no choices in response")
	}

	var lesson Lesson
	if err := json.Unmarshal([]byte(result.Choices[0].Message.Content), &lesson); err != nil {
		return nil, fmt.Errorf("parse lesson JSON: %w", err)
	}

	return &lesson, nil
}

// GenerateLesson generates a lesson using Anthropic
func (p *AnthropicProvider) GenerateLesson(ctx context.Context, category, topic string) (*Lesson, error) {
	if p.apiKey == "" {
		return nil, errors.New("Anthropic API key not configured")
	}

	prompt := fmt.Sprintf(`Generate a software engineering lesson about "%s" in the "%s" category. 

Return ONLY valid JSON in this exact structure:
{
  "title": "concise title",
  "category": "%s",
  "text": "1-2 sentence overview",
  "explain": "detailed explanation (2-3 sentences)",
  "useCases": ["use case 1", "use case 2", "use case 3"],
  "tips": ["tip 1", "tip 2", "tip 3"]
}

Focus on practical, actionable content for intermediate to advanced engineers.`, topic, category, category)

	reqBody := map[string]interface{}{
		"model":      p.model,
		"max_tokens": 1024,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Anthropic API error %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if len(result.Content) == 0 {
		return nil, errors.New("no content in response")
	}

	var lesson Lesson
	if err := json.Unmarshal([]byte(result.Content[0].Text), &lesson); err != nil {
		return nil, fmt.Errorf("parse lesson JSON: %w", err)
	}

	return &lesson, nil
}

// GetProvider returns the appropriate AI provider based on configuration
func GetProvider(providerName, model string) (Provider, error) {
	switch providerName {
	case "openai":
		return NewOpenAIProvider("", model), nil
	case "anthropic":
		return NewAnthropicProvider("", model), nil
	default:
		return nil, fmt.Errorf("unknown AI provider: %s", providerName)
	}
}
