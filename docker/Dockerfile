# Multi-stage build for AR-13 Go Backend Server
# This Dockerfile builds the server from source and creates a minimal production image

# Build stage
FROM golang:alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /build

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies
# GOTOOLCHAIN=auto will download required Go version if needed
ENV GOTOOLCHAIN=auto
RUN go mod download

# Copy source code
COPY . .

# Build the server binary
# CGO_ENABLED=0 creates a statically linked binary
# -ldflags "-w -s" strips debug info to reduce binary size
ARG TARGETOS=linux
ARG TARGETARCH=amd64
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
    -a -installsuffix cgo \
    -ldflags "-w -s" \
    -o server \
    ./cmd/server/main.go

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS support and curl for health check
RUN apk --no-cache add ca-certificates curl

# Create non-root user for security
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# Create necessary directories
RUN mkdir -p /app/upload && \
    chown -R appuser:appuser /app

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder --chown=appuser:appuser /build/server .

# Make binary executable
RUN chmod +x server

# Ensure all files are owned by appuser
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose port (default 3000, can be overridden via PORT env var)
EXPOSE 3000

# Health check endpoint
HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
  CMD curl -f http://localhost:3000/api/health || exit 1

# Run the server
CMD ["./server"]
