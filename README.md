
# AvidLearner — Software Engineering Coach (Go + React)

Single container app: Go backend + React frontend (Vite) with optional PWA install.

## Project Layout
```
.
├── Dockerfile           # root-level, builds frontend and backend
├── CHANGELOG.md
├── .env.example         # Environment configuration template
├── docs/
│   └── AI_FEATURE.md    # AI feature documentation
├── charts/              # Helm chart + k8s manifests
├── data/
│   ├── lessons.json     # 70+ lessons (expanded dataset)
│   └── pro_challenges.json
├── backend/
│   ├── main.go          # API endpoints
│   ├── go.mod
│   ├── config/
│   │   └── features.go  # Feature flag system
│   └── ai/
│       └── provider.go  # AI provider interface & implementations
├── frontend/            # Vite + React app w/ PWA manifest + SW
│   ├── package.json
│   ├── public/icon.svg
│   ├── src/
│   │   ├── components/
│   │   │   ├── AILessonGenerator.jsx
│   │   │   ├── Dashboard.jsx
│   │   │   ├── LessonView.jsx
│   │   │   └── ...
│   │   ├── api.js       # API client (with AI endpoints)
│   │   └── App.jsx
│   └── vite.config.js   # proxies /api to :8081 in dev
├── scripts/
│   └── run.ps1          # convenience launcher
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

A production build exposes the service worker and manifest, so the site behaves as a PWA (installable / offline support).

## API
- `GET /api/lessons` → `{ categories, lessons }`
- `GET /api/random?category=any|<name>` → one lesson
- `GET /api/session?stage=lesson` → returns a lesson and primes a quiz
- `GET /api/session?stage=quiz` → returns question + options
- `GET /api/session?stage=result&answer=A|B|C|D` → evaluates, updates coins/streak

State is kept per-browser via a cookie (`sid`) and in-memory on the server runtime.
You can replace `data/lessons.json` with a model-generated dataset using the same schema without breaking the UI.
```
[
  { "title": "...", "category": "...", "text": "...", "explain": "...", "useCases": [], "tips": [] }
]
```
