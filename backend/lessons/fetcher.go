package lessons

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
)

// Lesson represents a single lesson
type Lesson struct {
	Title    string   `json:"title"`
	Category string   `json:"category"`
	Text     string   `json:"text"`
	Explain  string   `json:"explain"`
	UseCases []string `json:"useCases"`
	Tips     []string `json:"tips"`
	Source   string   `json:"source,omitempty"` // "local", "github", "devto"
}

// Fetcher manages lesson sources with caching
type Fetcher struct {
	mu            sync.RWMutex
	localLessons  []Lesson
	cachedLessons []Lesson
	lastFetch     time.Time
	cacheTTL      time.Duration
	httpClient    *http.Client
}

// NewFetcher creates a new lesson fetcher
func NewFetcher(localLessons []Lesson, cacheTTL time.Duration) *Fetcher {
	return &Fetcher{
		localLessons: localLessons,
		cacheTTL:     cacheTTL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetLessons returns all lessons (local + cached external)
func (f *Fetcher) GetLessons(ctx context.Context) []Lesson {
	f.mu.RLock()
	needsRefresh := time.Since(f.lastFetch) > f.cacheTTL
	cached := append([]Lesson{}, f.cachedLessons...)
	f.mu.RUnlock()

	// Return current cache + local immediately
	combined := append([]Lesson{}, f.localLessons...)
	combined = append(combined, cached...)

	// Refresh in background if needed
	if needsRefresh {
		go f.refreshCache(context.Background())
	}

	return combined
}

// refreshCache fetches lessons from external sources
func (f *Fetcher) refreshCache(ctx context.Context) {
	log.Println("Refreshing lesson cache from external sources...")

	var wg sync.WaitGroup
	lessonsChan := make(chan []Lesson, 2)

	// Fetch from GitHub
	wg.Add(1)
	go func() {
		defer wg.Done()
		lessons, err := f.fetchFromGitHub(ctx)
		if err != nil {
			log.Printf("Error fetching from GitHub: %v", err)
			return
		}
		lessonsChan <- lessons
	}()

	// Fetch from Dev.to
	wg.Add(1)
	go func() {
		defer wg.Done()
		lessons, err := f.fetchFromDevTo(ctx)
		if err != nil {
			log.Printf("Error fetching from Dev.to: %v", err)
			return
		}
		lessonsChan <- lessons
	}()

	// Close channel when all fetches complete
	go func() {
		wg.Wait()
		close(lessonsChan)
	}()

	// Collect all lessons
	var allExternal []Lesson
	for lessons := range lessonsChan {
		allExternal = append(allExternal, lessons...)
	}

	// Update cache
	f.mu.Lock()
	f.cachedLessons = allExternal
	f.lastFetch = time.Now()
	f.mu.Unlock()

	log.Printf("Cache refreshed: %d lessons from external sources", len(allExternal))
}

// fetchFromGitHub pulls lessons from system-design-primer
func (f *Fetcher) fetchFromGitHub(ctx context.Context) ([]Lesson, error) {
	// Fetch README from system-design-primer
	url := "https://raw.githubusercontent.com/donnemartin/system-design-primer/master/README.md"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return parseGitHubMarkdown(string(body)), nil
}

// parseGitHubMarkdown extracts lessons from markdown content
func parseGitHubMarkdown(markdown string) []Lesson {
	var lessons []Lesson

	// Find sections with ## headers
	sectionRegex := regexp.MustCompile(`(?m)^##\s+(.+)$`)
	matches := sectionRegex.FindAllStringSubmatchIndex(markdown, -1)

	for i, match := range matches {
		titleStart, titleEnd := match[2], match[3]
		title := strings.TrimSpace(markdown[titleStart:titleEnd])

		// Skip navigation sections
		lowerTitle := strings.ToLower(title)
		if strings.Contains(lowerTitle, "index") ||
			strings.Contains(lowerTitle, "contribut") ||
			strings.Contains(lowerTitle, "credit") ||
			strings.Contains(lowerTitle, "license") {
			continue
		}

		// Get content until next section
		contentStart := match[1]
		contentEnd := len(markdown)
		if i+1 < len(matches) {
			contentEnd = matches[i+1][0]
		}

		content := strings.TrimSpace(markdown[contentStart:contentEnd])

		// Extract first paragraph as summary
		paragraphs := strings.Split(content, "\n\n")
		summary := ""
		explain := ""

		for _, p := range paragraphs {
			cleaned := strings.TrimSpace(p)
			if cleaned != "" && !strings.HasPrefix(cleaned, "#") && !strings.HasPrefix(cleaned, "<") {
				if summary == "" {
					summary = cleaned
					if len(summary) > 200 {
						summary = summary[:197] + "..."
					}
				} else if explain == "" {
					explain = cleaned
					if len(explain) > 300 {
						explain = explain[:297] + "..."
					}
					break
				}
			}
		}

		if summary == "" {
			continue
		}

		lessons = append(lessons, Lesson{
			Title:    title,
			Category: "system-design",
			Text:     summary,
			Explain:  explain,
			UseCases: []string{"Distributed systems", "Scalable architectures"},
			Tips:     []string{"Review trade-offs", "Consider CAP theorem"},
			Source:   "github",
		})
	}

	return lessons
}

// fetchFromDevTo pulls articles from Dev.to
func (f *Fetcher) fetchFromDevTo(ctx context.Context) ([]Lesson, error) {
	tags := []string{"architecture", "systemdesign", "designpatterns"}
	var allLessons []Lesson

	for _, tag := range tags {
		url := fmt.Sprintf("https://dev.to/api/articles?tag=%s&per_page=10&top=7", tag)

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			continue
		}

		resp, err := f.httpClient.Do(req)
		if err != nil {
			log.Printf("Error fetching Dev.to tag %s: %v", tag, err)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			continue
		}

		var articles []DevToArticle
		if err := json.NewDecoder(resp.Body).Decode(&articles); err != nil {
			resp.Body.Close()
			continue
		}
		resp.Body.Close()

		for _, article := range articles {
			lesson := Lesson{
				Title:    article.Title,
				Category: categorizeDevToArticle(article.Tags),
				Text:     truncate(article.Description, 200),
				Explain:  fmt.Sprintf("Read more at: %s", article.URL),
				UseCases: extractUseCases(article.Description),
				Tips:     []string{"Check the full article for details", "Consider practical applications"},
				Source:   "devto",
			}
			allLessons = append(allLessons, lesson)
		}

		// Rate limiting - be nice to Dev.to
		time.Sleep(500 * time.Millisecond)
	}

	return allLessons, nil
}

// DevToArticle represents a Dev.to article response
type DevToArticle struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	URL         string   `json:"url"`
	Tags        []string `json:"tags"`
}

// categorizeDevToArticle determines category from tags
func categorizeDevToArticle(tags []string) string {
	for _, tag := range tags {
		switch strings.ToLower(tag) {
		case "architecture", "systemdesign", "microservices":
			return "system-design"
		case "database", "sql", "nosql":
			return "databases"
		case "api", "rest", "graphql":
			return "apis"
		case "cloud", "aws", "azure", "kubernetes":
			return "cloud"
		case "security":
			return "security"
		}
	}
	return "general"
}

// extractUseCases tries to find use cases in description
func extractUseCases(description string) []string {
	// Simple extraction - look for bullet points or numbered lists
	useCases := []string{}
	lines := strings.Split(description, "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "-") || strings.HasPrefix(trimmed, "*") || strings.HasPrefix(trimmed, "•") {
			useCase := strings.TrimSpace(strings.TrimLeft(trimmed, "-*•"))
			if len(useCase) > 0 && len(useCase) < 100 {
				useCases = append(useCases, useCase)
			}
		}
	}

	if len(useCases) == 0 {
		useCases = []string{"General software engineering", "System architecture"}
	}

	return useCases
}

// truncate cuts string to maxLen with ellipsis
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// StartBackgroundRefresh starts a goroutine that refreshes cache periodically
func (f *Fetcher) StartBackgroundRefresh(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		// Initial fetch
		f.refreshCache(ctx)

		for {
			select {
			case <-ticker.C:
				f.refreshCache(ctx)
			case <-ctx.Done():
				log.Println("Stopping background lesson refresh")
				return
			}
		}
	}()
}
