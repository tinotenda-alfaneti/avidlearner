
# ---------- Frontend build ----------
FROM node:18-alpine AS frontend
WORKDIR /app
COPY frontend/package.json frontend/vite.config.js ./
RUN npm ci || npm i
COPY frontend ./
RUN npm run build

# ---------- Backend build ----------
FROM golang:1.24-alpine AS backend
WORKDIR /src
COPY backend/go.mod ./
RUN go mod download
COPY backend ./backend
WORKDIR /src/backend
RUN go build -o /out/server

# ---------- Final minimal image ----------
FROM alpine:3.20
WORKDIR /app
COPY --from=backend /out/server /app/server
COPY --from=frontend /app/dist /app/frontend/dist
COPY data /app/data
COPY backend/protests /app/backend/protests
ENV PORT=8081 LESSONS_FILE=/app/data/lessons.json
EXPOSE 8081
CMD ["/app/server"]
