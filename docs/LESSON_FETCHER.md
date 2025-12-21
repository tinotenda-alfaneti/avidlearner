# Hybrid Lesson Fetcher

## Overview

The application now uses a **hybrid approach** for lesson content, combining:
- **Local lessons** from `data/lessons.json` (always available, instant access)
- **External lessons** from free APIs (auto-refreshed, no cost)

## Features

### Free External Sources

1. **GitHub API** - System Design Primer
   - Source: `donnemartin/system-design-primer`
   - Content: System design patterns, architecture concepts
   - No API key required

2. **Dev.to API** - Engineering Articles
   - Tags: `architecture`, `systemdesign`, `designpatterns`
   - Content: Real-world tutorials and best practices
   - No API key required

### Smart Caching

- **Cache TTL**: 6 hours (configurable)
- **Background Refresh**: Automatic every 6 hours
- **Immediate Availability**: Returns local lessons instantly, fetches external in background
- **Resilient**: Falls back to local lessons if external APIs fail

### How It Works

```
┌─────────────────────────────────────────────┐
│         Application Startup                 │
├─────────────────────────────────────────────┤
│ 1. Load local lessons from JSON             │
│ 2. Initialize lesson fetcher                │
│ 3. Start background refresh goroutine       │
│ 4. Fetch external lessons (non-blocking)    │
└─────────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────┐
│         Runtime Behavior                    │
├─────────────────────────────────────────────┤
│ • API requests get: local + cached external │
│ • Every 6h: refresh external sources        │
│ • On error: continue with cached/local      │
└─────────────────────────────────────────────┘
```

## Configuration

### Cache Duration

Edit `backend/main.go` to adjust cache TTL:

```go
// Initialize with 6-hour cache
lessonFetcher = lessons.NewFetcher(localLessons, 6*time.Hour)

// Start refresh every 6 hours
lessonFetcher.StartBackgroundRefresh(context.Background(), 6*time.Hour)
```

### Add More Sources

Edit `backend/lessons/fetcher.go` in the `refreshCache()` method:

```go
// Add new source
wg.Add(1)
go func() {
    defer wg.Done()
    lessons, err := f.fetchFromYourSource(ctx)
    if err != nil {
        log.Printf("Error fetching from YourSource: %v", err)
        return
    }
    lessonsChan <- lessons
}()
```

## API Response

Lessons now include a `source` field:

```json
{
  "title": "Circuit Breaker",
  "category": "system-design",
  "text": "Circuit breaker stops cascading failures...",
  "explain": "A circuit breaker wraps a remote call...",
  "useCases": ["Calling flaky services", "..."],
  "tips": ["Track half-open state", "..."],
  "source": "local"  // "local", "github", or "devto"
}
```

## Testing

Run the lesson fetcher tests:

```bash
cd backend
go test ./lessons -v
```

Run all tests:

```bash
go test ./... -v
```
