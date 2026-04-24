# ── Build stage ──────────────────────────────────────────────────────────────
# Must match the `go` directive in go.mod (currently 1.26)
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Download dependencies first (layer-cache friendly)
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Fully static binaries — no libc, no CGO.
# -trimpath removes local file-system paths from the binary (smaller + safer).
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /app/backend ./main.go
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /app/migrate ./cmd/migrate
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /app/healthcheck ./cmd/healthcheck
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /app/resetpw ./cmd/resetpw

# ── Run stage ────────────────────────────────────────────────────────────────
# Use alpine:latest as a base image (often more accessible than gcr.io)
# We manually add a non-root user (uid 65532) to maintain security.
FROM alpine:latest

# Install standard CA certificates and timezone data (equivalent to distroless/static)
RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S -g 65532 nonroot \
    && adduser -S -u 65532 -G nonroot nonroot

WORKDIR /app

# Copy all binaries from the builder stage
COPY --from=builder /app/backend .
COPY --from=builder /app/migrate .
COPY --from=builder /app/healthcheck .
COPY --from=builder /app/resetpw .

# Copy migration SQL files
COPY --from=builder /app/migrations ./migrations

# Switch to the non-root user
USER 65532:65532

EXPOSE 8080

# The healthcheck binary makes a single HTTP GET to /health and exits 0/1.
# This works in distroless because it is a static binary — no shell needed.
HEALTHCHECK --interval=10s --timeout=5s --start-period=15s --retries=3 \
  CMD ["/app/healthcheck"]

ENTRYPOINT ["/app/backend"]
