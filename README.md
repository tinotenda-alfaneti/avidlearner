
# AvidLearner — Software Engineering Coach (Go + React)

Single container app: Go backend + React frontend.

## Project Layout
```
.
├── Dockerfile           # root-level, builds frontend and backend
├── data/
│   └── lessons.json     # rich dataset (same schema, model-friendly)
├── backend/
│   ├── main.go          # /api/lessons, /api/random, /api/session (stateful)
│   └── go.mod
└── frontend/            # Vite + React app
    ├── package.json
    ├── vite.config.js   # proxies /api to :8081 in dev
    └── src/...
```

## Dev
Backend:
```
cd backend
go run main.go
```
Frontend (Vite dev server with API proxy):
```
cd frontend
npm i
npm run dev
```

## Build & Run (Docker)
```
docker build -t avidlearner .
docker run -p 8081:8081 avidlearner
```
Open http://localhost:8081

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
