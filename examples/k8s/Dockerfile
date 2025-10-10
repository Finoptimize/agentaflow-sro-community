# Dockerfile for AgentaFlow Kubernetes GPU Scheduler
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o k8s-gpu-scheduler ./cmd/k8s-gpu-scheduler

FROM nvidia/cuda:11.8-base-ubuntu20.04

# Install nvidia-smi and other required tools
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    nvidia-utils-470 \
    curl \
    ca-certificates && \
    rm -rf /var/lib/apt/lists/*

# Create non-root user
RUN groupadd -r agentaflow && useradd -r -g agentaflow agentaflow

# Copy binary from builder
COPY --from=builder /app/k8s-gpu-scheduler /usr/local/bin/k8s-gpu-scheduler

# Set permissions
RUN chmod +x /usr/local/bin/k8s-gpu-scheduler

# Switch to non-root user for scheduler mode
USER agentaflow

EXPOSE 8080 9090

ENTRYPOINT ["/usr/local/bin/k8s-gpu-scheduler"]