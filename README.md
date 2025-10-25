
# AvidLearner — Software Engineering Coach (Go + React)

Single container app: Go backend + React frontend (Vite) with optional PWA install.

## Project Layout
```
.
├── Dockerfile           # root-level, builds frontend and backend
├── CHANGELOG.md
├── charts/              # Helm chart + k8s manifests
├── data/
│   └── lessons.json     # rich dataset (same schema, model-friendly)
├── backend/
│   ├── main.go          # /api/lessons, /api/random, /api/session (stateful)
│   └── go.mod
├── frontend/            # Vite + React app w/ PWA manifest + SW
│   ├── package.json
│   ├── public/icon.svg
│   ├── src/...
│   └── vite.config.js   # proxies /api to :8081 in dev
├── scripts/
│   └── run.ps1          # convenience launcher (backend + frontend)
└── README.md
```

## Dev
Single command (Windows PowerShell or pwsh on macOS/Linux):
```
pwsh scripts/run.ps1
```
This builds the Go backend, starts it on port 8081, and runs `npm run dev` (Vite) on port 5173. Override ports if needed:
```
pwsh scripts/run.ps1 -BackendPort 9090 -FrontendPort 3000
```

Manual start (if you prefer):
```
cd backend && go run main.go
# in another shell
cd frontend && npm install && npm run dev
```

## Build & Run (Docker)
```
docker build -t avidlearner .
docker run -p 8081:8081 avidlearner
```
Open http://localhost:8081

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
