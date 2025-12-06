# Multi-stage build for minimal image size

# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum* ./

# Download dependencies (GOTOOLCHAIN=auto will download required Go version if needed)
ENV GOTOOLCHAIN=auto
RUN go mod download

# Copy source code
COPY . .

# Build the binary (automatically detects target platform)
# TARGETOS and TARGETARCH are set by Docker BuildKit for multi-platform builds
ARG TARGETOS=linux
ARG TARGETARCH=amd64
ARG TARGETVARIANT=""
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -a -installsuffix cgo -o filebrowser main.go

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS support (if needed in future) and wget for health check
RUN apk --no-cache add ca-certificates wget

# Create non-root user
RUN addgroup -g 1000 filebrowser && \
    adduser -D -u 1000 -G filebrowser filebrowser

# Set working directory
WORKDIR /app

# Copy binary from builder (use --platform to ensure correct architecture)
ARG TARGETOS=linux
ARG TARGETARCH=amd64
COPY --from=builder --chown=filebrowser:filebrowser /build/filebrowser .

# Change ownership to non-root user
RUN chown -R filebrowser:filebrowser /app

# Switch to non-root user
USER filebrowser

# Expose port
EXPOSE 8082

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD wget --quiet --tries=1 --spider http://localhost:8082/health || exit 1

# Run the binary
CMD ["./filebrowser"]

