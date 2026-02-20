# AI Lesson Generation Feature

## Overview

AvidLearner now supports AI-powered lesson generation using OpenAI or Anthropic APIs. This feature is controlled by a feature flag and can be toggled on/off without code changes.

## Architecture

### Components

1. **Feature Flag System** (`backend/internal/featureflag/features.go`)
   - Thread-safe singleton pattern
   - Environment-based configuration
   - Runtime toggle support for testing

2. **AI Provider Interface** (`backend/internal/ai/provider.go`)
   - Abstract provider interface
   - OpenAI implementation
   - Anthropic implementation
   - Easy to extend for other providers

3. **Backend API Endpoints**
   - `GET /api/ai/config` - Returns AI feature status
   - `POST /api/ai/generate` - Generates a lesson with AI

4. **Frontend Components**
   - `AILessonGenerator.jsx` - UI for generating lessons
   - Integration with existing lesson flow
   - Conditional rendering based on feature flag

## Configuration

### Environment Variables

Copy `.env.example` to `.env` and configure:

```bash
# Enable AI lesson generation
ENABLE_AI_LESSONS=true

# Choose provider: openai or anthropic
AI_PROVIDER=openai

# Model selection
AI_MODEL=gpt-4

# Rate limiting
MAX_AI_LESSONS_PER_DAY=10

# API Keys (only needed for the provider you choose)
OPENAI_API_KEY=sk-your-key-here
# OR
ANTHROPIC_API_KEY=sk-ant-your-key-here
```

### Provider Options

#### OpenAI
- **Models**: `gpt-4`, `gpt-3.5-turbo`, `gpt-4-turbo`
- **API Key**: Set `OPENAI_API_KEY`
- **Cost**: ~$0.03 per lesson (gpt-4)

#### Anthropic
- **Models**: `claude-3-5-sonnet-20241022`, `claude-3-opus-20240229`
- **API Key**: Set `ANTHROPIC_API_KEY`
- **Cost**: ~$0.015 per lesson (claude-3.5-sonnet)

## Usage

### Enabling the Feature

1. Set environment variable:
   ```bash
   ENABLE_AI_LESSONS=true
   ```

2. Configure your chosen provider:
   ```bash
   AI_PROVIDER=openai
   OPENAI_API_KEY=sk-your-key-here
   ```

3. Restart the backend

4. The "Generate AI Lesson" button will appear in the dashboard

### Using AI Generation

1. Click "Generate AI Lesson" from the Learn Mode card
2. Enter a topic (e.g., "Database Connection Pooling")
3. Select a category
4. Click "Generate Lesson"
5. Wait 10-20 seconds for generation
6. Review and study the generated lesson

### Disabling the Feature

1. Set environment variable:
   ```bash
   ENABLE_AI_LESSONS=false
   ```

2. Restart the backend

3. The AI option will disappear from the UI

## API Reference

### GET /api/ai/config

Returns the current AI configuration status.

**Response:**
```json
{
  "aiEnabled": true,
  "provider": "openai",
  "maxPerDay": 10
}
```

### POST /api/ai/generate

Generates a custom lesson using AI.

**Request:**
```json
{
  "topic": "Circuit Breaker Pattern",
  "category": "system-design"
}
```

**Response:**
```json
{
  "stage": "lesson",
  "lesson": {
    "title": "Circuit Breaker Pattern",
    "category": "system-design",
    "text": "Brief overview...",
    "explain": "Detailed explanation...",
    "useCases": ["Use case 1", "Use case 2", "Use case 3"],
    "tips": ["Tip 1", "Tip 2", "Tip 3"]
  },
  "message": "AI-generated lesson"
}
```

## Implementation Details

### Thread Safety

The feature flag system uses `sync.RWMutex` for thread-safe reads and writes:

```go
func (f *FeatureFlags) IsAILessonsEnabled() bool {
    f.mu.RLock()
    defer f.mu.RUnlock()
    return f.aiLessonsEnabled
}
```

### Provider Pattern

The provider interface allows easy extension:

```go
type Provider interface {
    GenerateLesson(ctx context.Context, category, topic string) (*Lesson, error)
    GetProviderName() string
}
```

### Manual Testing

1. **Test with feature disabled:**
   ```bash
   ENABLE_AI_LESSONS=false go run main.go
   ```
   Verify AI button doesn't appear

2. **Test with OpenAI:**
   ```bash
   ENABLE_AI_LESSONS=true AI_PROVIDER=openai OPENAI_API_KEY=your-key go run main.go
   ```
   Generate a lesson and verify output

3. **Test with Anthropic:**
   ```bash
   ENABLE_AI_LESSONS=true AI_PROVIDER=anthropic ANTHROPIC_API_KEY=your-key go run main.go
   ```
   Generate a lesson and verify output

4. **Test without API key:**
   ```bash
   ENABLE_AI_LESSONS=true AI_PROVIDER=openai go run main.go
   ```
   Verify graceful error handling

```
