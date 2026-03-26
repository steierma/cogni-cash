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
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /app/cogni-cash-backend ./main.go
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /app/cogni-cash-migrate ./cmd/migrate
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /app/cogni-cash-healthcheck ./cmd/healthcheck
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /app/cogni-cash-resetpw ./cmd/resetpw

# ── Run stage ────────────────────────────────────────────────────────────────
# distroless/static is ~2 MB, has no shell, no package manager, and runs as
# a non-root user (uid 65532) out of the box.
# It already bundles ca-certificates and tzdata so no apk install is needed.
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

# Copy all binaries from the builder stage
COPY --from=builder /app/cogni-cash-backend .
COPY --from=builder /app/cogni-cash-migrate .
COPY --from=builder /app/cogni-cash-healthcheck .
COPY --from=builder /app/cogni-cash-resetpw .

# Copy migration SQL files so the migrate binary can find them at ./migrations
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080

# The healthcheck binary makes a single HTTP GET to /health and exits 0/1.
# This works in distroless because it is a static binary — no shell needed.
HEALTHCHECK --interval=10s --timeout=5s --start-period=15s --retries=3 \
  CMD ["/app/cogni-cash-healthcheck"]

ENTRYPOINT ["/app/cogni-cash-backend"]
