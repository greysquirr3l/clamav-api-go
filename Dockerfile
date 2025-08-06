# Multi-stage build for ClamAV API Go
# Stage 1: Build the Go application
FROM golang:1.24-alpine AS builder

# Install git for version info and ca-certificates for SSL
RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build arguments for version information
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_TIME=unknown

# Build the application with version info and optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.buildTime=${BUILD_TIME}" \
    -a -installsuffix cgo \
    -o clamav-api-go \
    .

# Stage 2: Create minimal runtime image
FROM alpine:3.20

# Install ca-certificates for SSL connections and timezone data
RUN apk --no-cache add ca-certificates tzdata && \
    addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/clamav-api-go /app/clamav-api-go

# Copy ClamAV configuration files (if they exist)
COPY --from=builder /app/config/clamd.conf /app/config/clamd.conf
COPY --from=builder /app/config/freshclam.conf /app/config/freshclam.conf

# Change ownership to non-root user
RUN chown -R appuser:appgroup /app

# Switch to non-root user for security
USER appuser

# Expose the API port (updated from 8080 to 8888)
EXPOSE 8888

# Add labels for better container management
LABEL maintainer="ClamAV API Go" \
      description="Production-ready ClamAV REST API with authentication" \
      version="${VERSION}" \
      commit="${COMMIT}" \
      build-time="${BUILD_TIME}"

# Health check for container orchestration
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8888/rest/v1/ping || exit 1

# Set default environment variables
ENV SERVER_ADDR=0.0.0.0:8888 \
    LOGGER_LOG_LEVEL=info \
    LOGGER_FORMAT=json \
    CLAMAV_ADDR=clamav:3310 \
    CLAMAV_TIMEOUT=30s

# Run the application
CMD ["/app/clamav-api-go"]