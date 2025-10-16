#!/bin/bash
# Bash script to build all AgentaFlow Docker images

set -e

echo "🐳 Building AgentaFlow Docker Images"
echo "====================================="
echo ""

# Navigate to project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
cd "$PROJECT_ROOT"

echo "📍 Project root: $PROJECT_ROOT"
echo ""

# Build web-dashboard
echo "🔨 Building web-dashboard image..."
docker build -f docker/Dockerfile.web-dashboard -t agentaflow-sro:web-dashboard .
echo "✅ web-dashboard build successful!"
echo ""

# Build k8s-scheduler
echo "🔨 Building k8s-scheduler image..."
docker build -f docker/Dockerfile.k8s-scheduler -t agentaflow-sro:k8s-scheduler .
echo "✅ k8s-scheduler build successful!"
echo ""

# Build prometheus-demo
echo "🔨 Building prometheus-demo image..."
docker build -f docker/Dockerfile.prometheus-demo -t agentaflow-sro:prometheus-demo .
echo "✅ prometheus-demo build successful!"
echo ""

# List built images
echo "📦 Built Images:"
docker images | grep agentaflow-sro
echo ""

echo "✨ All images built successfully!"
echo ""
echo "🚀 Quick Start Commands:"
echo "  docker run -p 9000:9000 -p 9001:9001 agentaflow-sro:web-dashboard"
echo "  docker-compose up -d"
echo ""
