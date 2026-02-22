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

	"avidlearner/internal/httpx"
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
		httpClient:   httpx.NewClient(10 * time.Second),
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

// fetchFromGitHub pulls lessons from system-design-primer and book of secret knowledge
func (f *Fetcher) fetchFromGitHub(ctx context.Context) ([]Lesson, error) {
	var allLessons []Lesson

	// Fetch from system-design-primer
	systemDesignLessons, err := f.fetchSystemDesignPrimer(ctx)
	if err != nil {
		log.Printf("Error fetching system-design-primer: %v", err)
	} else {
		allLessons = append(allLessons, systemDesignLessons...)
	}

	// Fetch from book-of-secret-knowledge
	secretKnowledgeLessons, err := f.fetchSecretKnowledge(ctx)
	if err != nil {
		log.Printf("Error fetching book-of-secret-knowledge: %v", err)
	} else {
		allLessons = append(allLessons, secretKnowledgeLessons...)
	}

	return allLessons, nil
}

// fetchSystemDesignPrimer pulls lessons from system-design-primer
func (f *Fetcher) fetchSystemDesignPrimer(ctx context.Context) ([]Lesson, error) {
	url := "https://raw.githubusercontent.com/donnemartin/system-design-primer/master/README.md"

	resp, err := httpx.DoWithRetry(ctx, f.httpClient, func() (*http.Request, error) {
		return http.NewRequestWithContext(ctx, "GET", url, nil)
	}, func(status int, _ []byte) error {
		return fmt.Errorf("github returned status %d", status)
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return parseGitHubMarkdown(string(body)), nil
}

// fetchSecretKnowledge pulls curated lessons from the-book-of-secret-knowledge
func (f *Fetcher) fetchSecretKnowledge(ctx context.Context) ([]Lesson, error) {
	url := "https://raw.githubusercontent.com/trimstray/the-book-of-secret-knowledge/master/README.md"

	resp, err := httpx.DoWithRetry(ctx, f.httpClient, func() (*http.Request, error) {
		return http.NewRequestWithContext(ctx, "GET", url, nil)
	}, func(status int, _ []byte) error {
		return fmt.Errorf("github returned status %d", status)
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return parseSecretKnowledgeMarkdown(string(body)), nil
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

// parseSecretKnowledgeMarkdown extracts curated lessons from the-book-of-secret-knowledge
func parseSecretKnowledgeMarkdown(markdown string) []Lesson {
	var lessons []Lesson

	// Categories we want to extract from the book
	categoryMap := map[string]struct {
		keywords []string
		category string
		tips     []string
	}{
		"CLI Tools": {
			keywords: []string{"command", "terminal", "shell", "cli"},
			category: "devops",
			tips:     []string{"Practice in a safe environment", "Read man pages", "Use --help flag"},
		},
		"Web Tools": {
			keywords: []string{"browser", "security", "ssl", "http"},
			category: "security",
			tips:     []string{"Bookmark useful tools", "Understand HTTPS/TLS", "Check multiple sources"},
		},
		"Security": {
			keywords: []string{"penetration", "vulnerability", "encryption"},
			category: "security",
			tips:     []string{"Stay ethical", "Get permission before testing", "Keep tools updated"},
		},
		"System Diagnostics": {
			keywords: []string{"debug", "monitor", "performance", "troubleshoot"},
			category: "devops",
			tips:     []string{"Monitor proactively", "Establish baselines", "Use multiple metrics"},
		},
		"Network": {
			keywords: []string{"network", "dns", "http", "tcp"},
			category: "networking",
			tips:     []string{"Understand OSI model", "Use tcpdump/wireshark", "Check DNS first"},
		},
		"Databases": {
			keywords: []string{"database", "sql", "nosql", "query"},
			category: "databases",
			tips:     []string{"Index wisely", "EXPLAIN queries", "Monitor slow queries"},
		},
	}

	// Find main sections (####)
	sectionRegex := regexp.MustCompile(`(?m)^####\s+(.+?)(?:\s+&nbsp;)?\s*\[`)
	matches := sectionRegex.FindAllStringSubmatchIndex(markdown, -1)

	for i, match := range matches {
		titleStart, titleEnd := match[2], match[3]
		title := strings.TrimSpace(markdown[titleStart:titleEnd])

		// Get content until next section
		contentStart := match[1]
		contentEnd := len(markdown)
		if i+1 < len(matches) {
			contentEnd = matches[i+1][0]
		}

		content := markdown[contentStart:contentEnd]

		// Extract tools/resources from the section
		toolRegex := regexp.MustCompile(`<a href="([^"]+)"><b>([^<]+)</b></a>\s*-\s*([^<\n]+)`)
		toolMatches := toolRegex.FindAllStringSubmatch(content, 15) // Limit to 15 per section

		// Determine category for this section
		categoryInfo := struct {
			keywords []string
			category string
			tips     []string
		}{
			keywords: []string{},
			category: "general",
			tips:     []string{"Research before using", "Check documentation", "Start with basics"},
		}

		for key, info := range categoryMap {
			if strings.Contains(title, key) {
				categoryInfo = info
				break
			}
		}

		// Create lessons from tools
		for idx, toolMatch := range toolMatches {
			if idx >= 10 { // Max 10 lessons per section to avoid overwhelming
				break
			}

			url := toolMatch[1]
			name := toolMatch[2]
			description := strings.TrimSpace(toolMatch[3])

			// Clean description
			if len(description) > 150 {
				description = description[:147] + "..."
			}

			// Create use cases based on description
			useCases := extractUseCasesFromDescription(description)
			if len(useCases) == 0 {
				useCases = []string{
					fmt.Sprintf("Learn %s", categoryInfo.category),
					"Improve technical skills",
				}
			}

			lesson := Lesson{
				Title:    name,
				Category: categoryInfo.category,
				Text:     description,
				Explain:  fmt.Sprintf("From The Book of Secret Knowledge: %s. Learn more at %s", title, url),
				UseCases: useCases,
				Tips:     categoryInfo.tips,
				Source:   "secret-knowledge",
			}

			lessons = append(lessons, lesson)
		}
	}

	log.Printf("Parsed %d lessons from Book of Secret Knowledge", len(lessons))
	return lessons
}

// extractUseCasesFromDescription attempts to extract use cases from description text
func extractUseCasesFromDescription(description string) []string {
	var useCases []string
	lowerDesc := strings.ToLower(description)

	// Common patterns
	patterns := map[string][]string{
		"security":    {"Security testing", "Vulnerability assessment", "Penetration testing"},
		"monitor":     {"System monitoring", "Performance tracking", "Resource management"},
		"debug":       {"Debugging", "Troubleshooting", "Error analysis"},
		"test":        {"Testing", "Quality assurance", "Validation"},
		"network":     {"Network analysis", "Traffic monitoring", "Connectivity troubleshooting"},
		"database":    {"Database management", "Query optimization", "Data analysis"},
		"deployment":  {"CI/CD", "Deployment automation", "Release management"},
		"container":   {"Container orchestration", "Microservices", "Cloud native apps"},
		"performance": {"Performance optimization", "Benchmarking", "Load testing"},
		"encrypt":     {"Data encryption", "Secure communication", "Privacy protection"},
	}

	for keyword, cases := range patterns {
		if strings.Contains(lowerDesc, keyword) {
			useCases = append(useCases, cases...)
			break
		}
	}

	// Limit to 3 use cases
	if len(useCases) > 3 {
		useCases = useCases[:3]
	}

	return useCases
}

// fetchFromDevTo pulls articles from Dev.to
func (f *Fetcher) fetchFromDevTo(ctx context.Context) ([]Lesson, error) {
	tags := []string{"architecture", "systemdesign", "designpatterns"}
	var allLessons []Lesson

	for _, tag := range tags {
		url := fmt.Sprintf("https://dev.to/api/articles?tag=%s&per_page=10&top=7", tag)

		resp, err := httpx.DoWithRetry(ctx, f.httpClient, func() (*http.Request, error) {
			return http.NewRequestWithContext(ctx, "GET", url, nil)
		}, func(status int, _ []byte) error {
			return fmt.Errorf("dev.to returned status %d", status)
		})
		if err != nil {
			log.Printf("Error fetching Dev.to tag %s: %v", tag, err)
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
