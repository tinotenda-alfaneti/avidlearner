# Use an official Golang image as the builder
FROM golang:1.20-alpine AS builder
WORKDIR /src

# Install git and certificates
RUN apk add --no-cache git ca-certificates && update-ca-certificates

# Copy go module files
COPY go.mod ./
RUN if [ -f go.sum ]; then go mod download; fi

# Copy the source code
COPY . .

RUN echo "🏗️ Building ARM64 binary..." && \
    CGO_ENABLED=0 GOOS=linux GOARCH=arm64 \
    go build -ldflags='-s -w' -o /out/backend ./backend

# Final lightweight image
FROM gcr.io/distroless/static:nonroot
WORKDIR /app

COPY --from=builder /out/backend /app/backend
COPY --from=builder /src/frontend /app/frontend
COPY --from=builder /src/data /app/data

ENV PORT=80
EXPOSE 80

ENTRYPOINT ["/app/backend"]
