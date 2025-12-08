# syntax=docker/dockerfile:1.7
## Multi-stage Dockerfile for kanban-backend
# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install build deps
RUN apk add --no-cache git ca-certificates

# Avoid auto-downloading toolchains and allow tidy in container
ENV GOTOOLCHAIN=local
ENV GOFLAGS -mod=mod

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
# WARNING: This is a development-only secret. ALWAYS override in production!
ENV JWT_SECRET=dev-secret-please-change-in-production
# DB_DSN and JWT_TTL can be provided at runtime (JWT_TTL defaults to 24h in code)

ENTRYPOINT ["/kanban-backend"]
