## Multi-stage Dockerfile for kanban-backend
# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install build deps
RUN apk add --no-cache git ca-certificates

# Cache modules
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest and build
COPY . .

# Build static binary (no cgo)
ENV CGO_ENABLED=0
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    go build -ldflags "-s -w" -o /kanban-backend ./cmd/api

# Runtime stage
FROM gcr.io/distroless/base-debian12:nonroot

WORKDIR /

COPY --from=builder /kanban-backend /kanban-backend

# Default port used by the app (can be overridden by env HTTP_PORT)
EXPOSE 8083

USER nonroot:nonroot

# Environment variables (override in your deployment)
ENV HTTP_PORT=8083
# DB_DSN, JWT_SECRET, JWT_TTL are expected to be provided at runtime

ENTRYPOINT ["/kanban-backend"]