# Multi-stage Dockerfile for SAI CLI

# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X main.version=$(git describe --tags --always --dirty 2>/dev/null || echo 'dev') -X main.buildTime=$(date -u '+%Y-%m-%d_%H:%M:%S') -X main.commit=$(git rev-parse --short HEAD 2>/dev/null || echo 'unknown')" \
    -a -installsuffix cgo \
    -o sai \
    ./cmd/sai

# Final stage
FROM scratch

# Import from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

# Copy the binary
COPY --from=builder /app/sai /usr/local/bin/sai

# Copy configuration files
COPY --from=builder /app/providers /usr/share/sai/providers
COPY --from=builder /app/schemas /usr/share/sai/schemas

# Create non-root user
USER nobody:nobody

# Set entrypoint
ENTRYPOINT ["/usr/local/bin/sai"]

# Default command
CMD ["--help"]

# Labels
LABEL org.opencontainers.image.title="SAI CLI"
LABEL org.opencontainers.image.description="Universal software management CLI tool"
LABEL org.opencontainers.image.url="https://github.com/sai-cli/sai"
LABEL org.opencontainers.image.source="https://github.com/sai-cli/sai"
LABEL org.opencontainers.image.vendor="SAI CLI Team"
LABEL org.opencontainers.image.licenses="MIT"