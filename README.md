
# AvidLearner — Software Engineering Coach (Go + React)

Single container app: Go backend + React frontend (Vite) with optional PWA install.

## Project Layout
```
.
├── Dockerfile            # root-level, builds frontend and backend
├── CHANGELOG.md
├── Jenkinsfile
├── Makefile
├── .env.example          # Environment configuration template
├── docs/
│   ├── AI_FEATURE.md     # AI feature documentation
│   ├── LESSON_FETCHER.md # Lesson fetcher behavior
│   ├── LESSON_SOURCES.md # Lesson sources + filtering
│   └── TECH_NEWS_FEATURE.md
├── charts/               # Helm chart + k8s manifests
│   └── app/
│       └── templates/
├── ci/
│   └── kubernetes/
├── data/
│   ├── lessons.json              # 70+ lessons (expanded dataset)
│   ├── challenges.json           # sample coding challenges for autograder
│   ├── pro_challenges.json
│   ├── secret_knowledge_lessons.json # Curated content from Book of Secret Knowledge
│   └── leaderboard.json          # Persistent leaderboard storage
├── backend/
│   ├── main.go                   # API entrypoint
│   ├── go.mod
│   ├── ai/
│   │   └── provider.go           # AI provider interface & implementations
│   ├── config/
│   │   └── features.go           # Feature flag system
│   ├── internal/
│   │   ├── models/
│   │   │   └── models.go
│   │   └── routes/
│   │       ├── routes.go
│   │       └── state.go
│   ├── lessons/
│   │   └── fetcher.go             # External lesson fetcher
│   └── protests/                  # Go practice exercises (see subfolders)
├── frontend/             # Vite + React app w/ PWA manifest + SW
│   ├── package.json
│   ├── public/icon.svg
│   ├── src/
│   │   ├── components/
│   │   │   ├── AILessonGenerator.jsx
│   │   │   ├── Dashboard.jsx
│   │   │   ├── Leaderboard.jsx        # Global leaderboard view
│   │   │   ├── LessonView.jsx
│   │   │   ├── NameInputModal.jsx     # Score submission modal
│   │   │   ├── TechNews.jsx
│   │   │   └── ...
│   │   ├── api.js        # API client (with AI & leaderboard endpoints)
│   │   ├── App.jsx
│   │   └── main.jsx
│   └── vite.config.js    # proxies /api to :8081 in dev
├── scripts/
│   ├── run.ps1           # convenience launcher
│   ├── setup.ps1         # local setup helper
│   └── test-all.ps1      # test runner
├── tools/                # small developer CLIs and utilities
│   ├── tiny-tutor/       # terminal-first lesson tutor (tools/tiny-tutor)
│   └── autograder/       # simple Go autograder for coding challenges
└── README.md
```

## Quick Start

### Development

Single command (Windows PowerShell or pwsh on macOS/Linux):
```bash
pwsh scripts/run.ps1
```
This builds the Go backend, starts it on port 8081, and runs `npm run dev` (Vite) on port 5173.

Override ports if needed:
```bash
pwsh scripts/run.ps1 -BackendPort 9090 -FrontendPort 3000
```

Manual start (if you prefer):
```bash
cd backend && go run main.go
# in another shell
cd frontend && npm install && npm run dev
```

### Build & Run (Docker)
```bash
docker build -t avidlearner .
docker run -p 8081:8081 avidlearner
```
Open http://localhost:8081

## Testing

Comprehensive unit tests are available for both backend and frontend.

```bash
# Run all tests
pwsh scripts/test-all.ps1
```

## AI Feature Configuration

The AI lesson generation feature is **disabled by default** and controlled via environment variables.

### Setup

1. Copy the example environment file:
   ```bash
   cp .env.example .env
   ```

2. Configure the feature flag and API keys:
   ```bash
   # Enable AI lesson generation
   ENABLE_AI_LESSONS=true
   
   # Choose provider: openai or anthropic
   AI_PROVIDER=openai
   
   # Set your API key (only for the provider you choose)
   OPENAI_API_KEY=sk-your-key-here
   # OR
   ANTHROPIC_API_KEY=sk-ant-your-key-here
   ```

3. Restart the backend

### Using AI Generation

When enabled, a "Generate AI Lesson" button appears in Learn Mode:
1. Click the button
2. Enter any software engineering topic
3. Select a category
4. Wait 10-20 seconds for generation
5. Study the custom-generated lesson

### Disabling the Feature

Set `ENABLE_AI_LESSONS=false` in your `.env` file and restart. The AI option will disappear from the UI.

For detailed configuration, costs, and troubleshooting, see [docs/AI_FEATURE.md](docs/AI_FEATURE.md).

## Content Sources

AvidLearner curates content from multiple high-quality sources:

### Built-in Lessons
- 70+ core software engineering lessons
- Topics: System Design, Databases, APIs, Cloud, Security, DevOps

### Book of Secret Knowledge Integration
Automatically loads curated content from [The Book of Secret Knowledge](https://github.com/trimstray/the-book-of-secret-knowledge):
- **20 hand-picked lessons** load immediately from static file
- **Dynamic fetcher** pulls latest tools and resources every 6 hours
- **Categories**: DevOps tools, Security frameworks, Network utilities, Best practices
- **Examples**: htop, Burp Suite, Wireshark, Docker, OWASP guides

### External Sources (Dynamic)
- System Design Primer (GitHub)
- Dev.to articles (system design, architecture tags)
- Refreshes automatically in the background every 6 hours

### AI-Generated Lessons (Optional)
- Generate custom lessons on any topic when enabled
- Requires API key configuration

**Source Filtering**: In Learn Mode, use the source selector to focus on specific content:
- All Sources - See everything (~240 lessons)
- Local - 70+ curated core lessons
- GitHub - System Design Primer content
- Secret Knowledge - DevOps/Security tools (134 lessons)
- Dev.to - Community articles
- AI Generated - Custom lessons (shows "Coming Soon" when disabled)

See [docs/LESSON_SOURCES.md](docs/LESSON_SOURCES.md) for details.

## Leaderboard System

AvidLearner includes a **secure, global leaderboard** for all game modes with server-side validation to prevent cheating.


### How to Use

1. **Play Any Game**: Complete a quiz, typing test, or coding challenge
2. **Submit Your Score**: Click the "Submit to Leaderboard" button that appears
3. **Enter Your Name**: Type your name (max 30 characters)
4. **View Rankings**: Click "View Leaderboard" on the dashboard to see top scores

### Score Types

- **Quiz Mode**: Number of correct answers in your quiz session
- **Typing Mode**: Words per minute (WPM) from your typing test
- **Coding Mode**: Total XP earned from completed challenges

### Security

All scores are **validated server-side** to ensure legitimacy:
- The backend tracks your actual gameplay in real-time
- When you submit, your claimed score is validated against the server's record
- Fake scores are rejected with a 403 Forbidden error
- This makes it impossible to cheat using browser dev tools

For detailed information, see:
- [Leaderboard Security Documentation](docs/LEADERBOARD_SECURITY.md)
- [Leaderboard UI Guide](docs/LEADERBOARD_UI_GUIDE.md)
- [Complete Implementation Details](docs/LEADERBOARD_IMPLEMENTATION.md)

A production build exposes the service worker and manifest, so the site behaves as a PWA (installable / offline support).

## API
- `GET /api/lessons` → `{ categories, lessons }`
- `GET /api/random?category=any|<name>` → one lesson
- `GET /api/session?stage=lesson` → returns a lesson and primes a quiz
- `GET /api/session?stage=quiz` → returns question + options
- `GET /api/session?stage=result&answer=A|B|C|D` → evaluates, updates coins/streak
- `GET /api/leaderboard?mode=quiz|typing|coding` → returns top 100 scores
- `POST /api/leaderboard/submit` → submit score (validated server-side)
- `POST /api/typing/score` → update typing score for session

State is kept per-browser via a cookie (`sid`) and in-memory on the server runtime.
Leaderboard data is persisted to `data/leaderboard.json` and survives restarts.

You can replace `data/lessons.json` with a model-generated dataset using the same schema without breaking the UI.
```
[
  { "title": "...", "category": "...", "text": "...", "explain": "...", "useCases": [], "tips": [] }
]
```
