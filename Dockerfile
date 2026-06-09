# Aetra Blockchain Docker Image
# Multi-stage build for aetrad validator node

# ===========================================
# Stage 1: Builder
# ===========================================
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache \
    git \
    make \
    ca-certificates \
    tzdata

WORKDIR /build

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build arguments
ARG VERSION=dev
ARG COMMIT=unknown
ARG DATE=unknown
ARG APP_NAME=aetrad

# Build the binary with version ldflags
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
    -ldflags="-X github.com/sovereign-l1/l1/app.Version=${VERSION} \
              -X github.com/sovereign-l1/l1/app.Commit=${COMMIT} \
              -X github.com/sovereign-l1/l1/app.BuildDate=${DATE}" \
    -o /aetrad \
    ./cmd/aetrad

# ===========================================
# Stage 2: Runner
# ===========================================
FROM alpine:3.19 AS runner

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    curl \
    bash

# Create non-root user for security
RUN addgroup -g 1000 aetra && \
    adduser -u 1000 -G aetra -s /bin/bash -D aetra

WORKDIR /home/aetra

# Copy binary from builder
COPY --from=builder /aetrad /usr/local/bin/aetrad

# Create data directory
RUN mkdir -p /home/aetra/.aetrad && \
    chown -R aetra:aetra /home/aetra

# Set non-root user
USER aetra

# Expose ports
# CometBFT P2P
EXPOSE 26656
# CometBFT RPC
EXPOSE 26657
# REST API
EXPOSE 1317
# gRPC
EXPOSE 9090
# Prometheus metrics
EXPOSE 6060

# Healthcheck
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:26657/health || exit 1

# Default command: show version and help
ENTRYPOINT ["/usr/local/bin/aetrad"]
CMD ["--help"]

# Labels
LABEL org.opencontainers.image.title="Aetra Blockchain"
LABEL org.opencontainers.image.description="Aetra testnet validator node"
LABEL org.opencontainers.image.source="https://github.com/sovereign-l1/l1"
LABEL org.opencontainers.image.version="${VERSION}"
LABEL org.opencontainers.image.created="${DATE}"
LABEL maintainer="Aetra Team"