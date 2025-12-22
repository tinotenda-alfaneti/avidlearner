# Lesson Loading Sources

AvidLearner supports **four ways** to load lessons, giving you diverse, always-current content from multiple sources.

## The Four Lesson Sources

### 1. 📚 **Local Lessons** (Immediate)
**File**: `data/lessons.json`

70+ curated lessons on core software engineering topics:
- System Design patterns (Circuit Breaker, Idempotency, etc.)
- Database concepts (Sharding, Replication, CAP theorem)
- API design (REST, GraphQL, Rate Limiting)
- Cloud architecture (Load balancing, Caching, CDN)
- Security best practices

**Loads**: Instantly on startup (baked into Docker image)

---

### 2. 🌐 **GitHub Sources** (Background)

#### System Design Primer
**Source**: [donnemartin/system-design-primer](https://github.com/donnemartin/system-design-primer)

Automatically fetches architecture and scalability lessons from this popular repository.

#### Book of Secret Knowledge
**Source**: [trimstray/the-book-of-secret-knowledge](https://github.com/trimstray/the-book-of-secret-knowledge)

Curated collection of DevOps tools, security frameworks, and best practices:
- **Static**: 20 hand-picked lessons in `data/secret_knowledge_lessons.json`
- **Dynamic**: 114+ lessons fetched live (htop, Burp Suite, Docker, OWASP guides, etc.)

**Loads**: 15 seconds after startup, refreshes every 6 hours

---

### 3. 📰 **Dev.to API** (Background)
**Tags**: `architecture`, `systemdesign`, `designpatterns`

Pulls top articles from Dev.to's community for fresh perspectives on:
- Modern architecture patterns
- Real-world system design case studies
- Design pattern implementations

**Loads**: 15 seconds after startup, refreshes every 6 hours

---

### 4. 🤖 **AI Generation** (On-Demand)
**Providers**: OpenAI GPT-4 or Anthropic Claude

Generate custom lessons on any software engineering topic:
1. User enters topic (e.g., "GraphQL subscriptions")
2. Selects category
3. AI creates lesson with explanations, use cases, and tips
4. Lesson added to current session

**Requires**: `ENABLE_AI_LESSONS=true` and API key configured  
See [AI_FEATURE.md](AI_FEATURE.md) for setup details

---

## Total Lessons Available

When all sources are active:
- **70** Local lessons
- **114** Book of Secret Knowledge
- **25** System Design Primer  
- **30** Dev.to articles
- **Unlimited** AI-generated (on demand)

**≈240 lessons** ready to learn!

---

## Configuration

### Backend (main.go)
```go
// External sources refresh every 6 hours
lessonFetcher = lessons.NewFetcher(localLessons, 6*time.Hour)
lessonFetcher.StartBackgroundRefresh(context.Background(), 6*time.Hour)
```

### Enable AI
In `.env`:
```bash
ENABLE_AI_LESSONS=true
AI_PROVIDER=openai  # or anthropic
OPENAI_API_KEY=sk-your-key
```

---

## Monitoring Logs

On startup:
```
Loaded 20 lessons from Book of Secret Knowledge
Loaded 223 lessons total (223 local + external)
Cache refreshed: 139 lessons from external sources
```
