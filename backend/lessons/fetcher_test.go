package lessons

import (
	"context"
	"testing"
	"time"
)

func TestNewFetcher(t *testing.T) {
	localLessons := []Lesson{
		{
			Title:    "Test Lesson",
			Category: "testing",
			Text:     "This is a test",
			Explain:  "Testing the fetcher",
			UseCases: []string{"Unit tests"},
			Tips:     []string{"Write good tests"},
			Source:   "local",
		},
	}

	fetcher := NewFetcher(localLessons, 1*time.Hour)
	if fetcher == nil {
		t.Fatal("Expected fetcher to be initialized")
	}

	if len(fetcher.localLessons) != 1 {
		t.Errorf("Expected 1 local lesson, got %d", len(fetcher.localLessons))
	}

	if fetcher.cacheTTL != 1*time.Hour {
		t.Errorf("Expected cache TTL of 1 hour, got %v", fetcher.cacheTTL)
	}
}

func TestGetLessons(t *testing.T) {
	localLessons := []Lesson{
		{
			Title:    "Local Lesson 1",
			Category: "local",
			Text:     "Local content",
			Source:   "local",
		},
		{
			Title:    "Local Lesson 2",
			Category: "local",
			Text:     "More local content",
			Source:   "local",
		},
	}

	fetcher := NewFetcher(localLessons, 1*time.Hour)

	ctx := context.Background()
	lessons := fetcher.GetLessons(ctx)

	if len(lessons) < 2 {
		t.Errorf("Expected at least 2 lessons, got %d", len(lessons))
	}

	// Check that local lessons are included
	foundLocal := 0
	for _, l := range lessons {
		if l.Source == "local" {
			foundLocal++
		}
	}

	if foundLocal != 2 {
		t.Errorf("Expected 2 local lessons, found %d", foundLocal)
	}
}

func TestCategorizeDevToArticle(t *testing.T) {
	tests := []struct {
		tags     []string
		expected string
	}{
		{[]string{"architecture", "backend"}, "system-design"},
		{[]string{"microservices"}, "system-design"},
		{[]string{"database", "sql"}, "databases"},
		{[]string{"nosql", "mongodb"}, "databases"},
		{[]string{"api", "rest"}, "apis"},
		{[]string{"graphql"}, "apis"},
		{[]string{"cloud", "aws"}, "cloud"},
		{[]string{"kubernetes"}, "cloud"},
		{[]string{"security"}, "security"},
		{[]string{"random", "other"}, "general"},
	}

	for _, tt := range tests {
		result := categorizeDevToArticle(tt.tags)
		if result != tt.expected {
			t.Errorf("categorizeDevToArticle(%v) = %s; want %s", tt.tags, result, tt.expected)
		}
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"short", 10, "short"},
		{"exactly ten", 11, "exactly ten"},
		{"this is a very long string that needs truncation", 20, "this is a very lo..."},
		{"", 10, ""},
	}

	for _, tt := range tests {
		result := truncate(tt.input, tt.maxLen)
		if result != tt.expected {
			t.Errorf("truncate(%q, %d) = %q; want %q", tt.input, tt.maxLen, result, tt.expected)
		}
	}
}

func TestExtractUseCases(t *testing.T) {
	tests := []struct {
		description string
		expected    int // expected number of use cases
	}{
		{"- Use case 1\n- Use case 2\n- Use case 3", 3},
		{"* First point\n* Second point", 2},
		{"• Bullet one\n• Bullet two", 2},
		{"No bullet points here", 2}, // Should return default
		{"", 2},                      // Should return default
	}

	for _, tt := range tests {
		result := extractUseCases(tt.description)
		if len(result) != tt.expected {
			t.Errorf("extractUseCases(%q) returned %d use cases; want %d", tt.description, len(result), tt.expected)
		}
	}
}

func TestParseGitHubMarkdown(t *testing.T) {
	markdown := `
# Main Title

## Database Sharding

Database sharding splits a large database into smaller, faster, more easily managed parts called shards.

Each shard is a separate database, and together they make up a single logical database.

## Load Balancer

Load balancers distribute incoming client requests to computing resources such as application servers and databases.

In each case, the load balancer returns the response from the computing resource to the appropriate client.

## Index of Topics

This is a navigation section that should be skipped.

## Contributing

How to contribute to this project.
`

	lessons := parseGitHubMarkdown(markdown)

	// Should extract Database Sharding and Load Balancer, skip Index and Contributing
	if len(lessons) < 2 {
		t.Errorf("Expected at least 2 lessons, got %d", len(lessons))
	}

	// Check that lessons have content
	for _, lesson := range lessons {
		if lesson.Title == "" {
			t.Error("Lesson title should not be empty")
		}
		if lesson.Text == "" {
			t.Error("Lesson text should not be empty")
		}
		if lesson.Category != "system-design" {
			t.Errorf("Expected category 'system-design', got %q", lesson.Category)
		}
		if lesson.Source != "github" {
			t.Errorf("Expected source 'github', got %q", lesson.Source)
		}
	}

	// Verify navigation sections were skipped
	for _, lesson := range lessons {
		title := lesson.Title
		if title == "Index of Topics" || title == "Contributing" {
			t.Errorf("Should have skipped navigation section: %s", title)
		}
	}
}

func TestCacheRefresh(t *testing.T) {
	localLessons := []Lesson{
		{
			Title:    "Local",
			Category: "test",
			Text:     "Local lesson",
			Source:   "local",
		},
	}

	// Use very short TTL for testing
	fetcher := NewFetcher(localLessons, 1*time.Millisecond)

	ctx := context.Background()

	// First call - cache is empty
	lessons1 := fetcher.GetLessons(ctx)
	initialCount := len(lessons1)

	// Wait for cache to expire
	time.Sleep(2 * time.Millisecond)

	// Second call - should trigger background refresh
	lessons2 := fetcher.GetLessons(ctx)

	// Give background refresh a moment to start
	time.Sleep(100 * time.Millisecond)

	// Verify we still get lessons (at least local ones)
	if len(lessons2) < initialCount {
		t.Errorf("Expected at least %d lessons after refresh, got %d", initialCount, len(lessons2))
	}
}
