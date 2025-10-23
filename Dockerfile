# --- Stage 1: Builder ---
FROM --platform=$BUILDPLATFORM golang:1.20-alpine AS builder
ARG TARGETOS
ARG TARGETARCH

WORKDIR /src

# Install git & certificates
RUN apk add --no-cache git ca-certificates && update-ca-certificates

# Copy go.mod (and go.sum if it exists)
COPY go.mod ./
RUN if [ -f go.sum ]; then go mod download; fi

# Copy source
COPY . .

# Build architecture-specific binary
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -ldflags='-s -w' -o /out/backend ./backend

# --- Stage 2: Final image ---
FROM gcr.io/distroless/static:nonroot
WORKDIR /app

COPY --from=builder /out/backend /app/backend
COPY --from=builder /src/frontend /app/frontend
COPY --from=builder /src/data /app/data

ENV PORT=80
EXPOSE 80

ENTRYPOINT ["/app/backend"]
