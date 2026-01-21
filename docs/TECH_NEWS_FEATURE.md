# Tech News Feature

## Overview
The Tech News section displays the latest articles from popular developer communities directly on your AvidLearner dashboard.

## News Sources

### HackerNews
- **API**: Official HackerNews Firebase API
- **Content**: Top 10 stories from tech community
- **Cost**: Free, no authentication required
- **Rate Limits**: Generous, suitable for frequent polling

### Dev.to
- **API**: Official Dev.to API
- **Content**: Top articles from the past week
- **Cost**: Free, no authentication required
- **Tags**: Shows article tags for quick categorization

### Features
- Loading states with spinner
- Error handling with retry button
- Time formatting (e.g., "2h ago", "5d ago")
- Points/upvotes display
- Comment counts
- Author information

## Customization

### Add More Sources
Add new sources by:
1. Adding to `NEWS_SOURCES` object
2. Creating a new `fetch{SourceName}` function
3. Adding a case to the switch statement in `fetchNews`