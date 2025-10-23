# Use an official Golang image as the builder
FROM golang:1.20-alpine AS builder
WORKDIR /src

# Install git for `go get` if needed and ca-certificates
RUN apk add --no-cache git ca-certificates && update-ca-certificates || true

# Copy Go modules manifests and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source
COPY . .

# Build a statically linked binary suitable for scratch if desired
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags='-s -w' -o /out/backend ./backend

# Final image: small distroless-like base
FROM gcr.io/distroless/static:nonroot
WORKDIR /app
COPY --from=builder /out/backend /app/backend
COPY --from=builder /src/frontend /app/frontend
COPY --from=builder /src/data /app/data

ENV PORT=80
EXPOSE 80

ENTRYPOINT ["/app/backend"]