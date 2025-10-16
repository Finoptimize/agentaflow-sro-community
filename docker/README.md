# AgentaFlow Docker Documentation

This directory contains Docker configurations for AgentaFlow SRO Community Edition, enabling containerized deployment with minimal setup.

## 📦 Available Images

### 1. Web Dashboard (`Dockerfile.web-dashboard`)
**Purpose**: Real-time GPU monitoring dashboard with WebSocket updates

**Features**:
- Production-ready web interface
- Real-time GPU metrics visualization
- Cost tracking and analytics
- Alert management system

**Ports**:
- `9000`: Web Dashboard UI
- `9001`: Prometheus Metrics

**Usage**:
```bash
# Build locally
docker build -f docker/Dockerfile.web-dashboard -t agentaflow-dashboard .

# Run
docker run -p 9000:9000 -p 9001:9001 agentaflow-dashboard

# Access at http://localhost:9000
```

### 2. Kubernetes GPU Scheduler (`Dockerfile.k8s-scheduler`)
**Purpose**: Kubernetes-native GPU workload scheduler

**Features**:
- Multiple scheduling strategies
- GPU resource optimization
- Custom Resource Definitions (CRDs)
- Real-time workload management

**Ports**:
- `8080`: Metrics and API

**Usage**:
```bash
# Build locally
docker build -f docker/Dockerfile.k8s-scheduler -t k8s-gpu-scheduler .

# Run (requires kubeconfig)
docker run -p 8080:8080 \
  -v ~/.kube:/root/.kube:ro \
  k8s-gpu-scheduler
```

### 3. Prometheus Demo (`Dockerfile.prometheus-demo`)
**Purpose**: Prometheus integration demo with GPU monitoring

**Features**:
- Prometheus metrics export
- GPU simulation for demos
- Cost tracking integration
- Performance analytics

**Ports**:
- `8080`: Prometheus Metrics

**Usage**:
```bash
# Build locally
docker build -f docker/Dockerfile.prometheus-demo -t prometheus-demo .

# Run
docker run -p 8080:8080 prometheus-demo

# Metrics available at http://localhost:8080/metrics
```

## 🚀 Quick Start with Docker Compose

The easiest way to run the complete AgentaFlow stack:

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop all services
docker-compose down

# Stop and remove volumes
docker-compose down -v
```

### Access Services:
- **Web Dashboard**: http://localhost:9000
- **Grafana**: http://localhost:3000 (admin/agentaflow123)
- **Prometheus**: http://localhost:9090

## 🏗️ Build Arguments

### Multi-Architecture Builds

Build for multiple platforms:
```bash
# Build for AMD64 and ARM64
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -f docker/Dockerfile.web-dashboard \
  -t agentaflow-dashboard:latest \
  .
```

### Custom Build Arguments

```bash
# Build with specific Go version
docker build \
  --build-arg GO_VERSION=1.21 \
  -f docker/Dockerfile.web-dashboard \
  -t agentaflow-dashboard:custom \
  .
```

## 🔧 Configuration

### Environment Variables

All containers support these common variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `LOG_LEVEL` | `info` | Logging level (debug, info, warn, error) |
| `GPU_SIMULATION` | `true` | Enable GPU simulation for demo |
| `WEB_PORT` | `9000` | Web dashboard port |
| `METRICS_PORT` | `9001` | Prometheus metrics port |

### Web Dashboard Specific:
```bash
docker run -p 9000:9000 -p 9001:9001 \
  -e LOG_LEVEL=debug \
  -e GPU_SIMULATION=true \
  -e ALERT_THRESHOLDS=temperature:80,utilization:95,memory:90 \
  agentaflow-dashboard
```

### Kubernetes Scheduler Specific:
```bash
docker run -p 8080:8080 \
  -e SCHEDULER_PORT=8080 \
  -e STRATEGY=least-utilized \
  -e LOG_LEVEL=info \
  k8s-gpu-scheduler
```

## 💾 Data Persistence

### Using Volumes

For persistent data storage:

```bash
# Create named volume
docker volume create agentaflow-data

# Run with volume
docker run -p 9000:9000 -p 9001:9001 \
  -v agentaflow-data:/data \
  agentaflow-dashboard
```

### Using Bind Mounts

For development:

```bash
docker run -p 9000:9000 -p 9001:9001 \
  -v $(pwd)/data:/data \
  agentaflow-dashboard
```

## 🔒 Security Best Practices

### Running as Non-Root

All images run as non-root user (UID 65532) by default:

```dockerfile
USER nonroot:nonroot
```

### Read-Only Root Filesystem

For enhanced security:

```bash
docker run -p 9000:9000 \
  --read-only \
  --tmpfs /tmp \
  agentaflow-dashboard
```

### Security Scanning

Scan images for vulnerabilities:

```bash
# Using Trivy
docker run --rm \
  -v /var/run/docker.sock:/var/run/docker.sock \
  aquasec/trivy image agentaflow-dashboard:latest

# Using Docker Scout (Docker Desktop)
docker scout cves agentaflow-dashboard:latest
```

## 🐛 Troubleshooting

### Container Won't Start

Check logs:
```bash
docker logs agentaflow-dashboard
docker-compose logs agentaflow-dashboard
```

### Port Already in Use

Find and stop conflicting process:
```bash
# Linux/Mac
lsof -i :9000
kill -9 <PID>

# Windows PowerShell
netstat -ano | findstr :9000
taskkill /F /PID <PID>
```

### Health Check Failing

Verify health endpoint:
```bash
docker exec agentaflow-dashboard /app/agentaflow-dashboard --health-check
```

### Permission Issues

Ensure proper permissions:
```bash
# Fix volume permissions
docker run --rm \
  -v agentaflow-data:/data \
  alpine chown -R 65532:65532 /data
```

## 📊 Monitoring Containers

### View Resource Usage

```bash
# All containers
docker stats

# Specific container
docker stats agentaflow-dashboard
```

### Container Health

```bash
# Check health status
docker inspect --format='{{.State.Health.Status}}' agentaflow-dashboard

# View health logs
docker inspect --format='{{range .State.Health.Log}}{{.Output}}{{end}}' agentaflow-dashboard
```

## 🤖 CI/CD Pipeline & GitHub Packages

### Automated Builds

AgentaFlow uses GitHub Actions for automated container builds:

**Triggers**:
- Push to `main` or `develop` branches
- Version tags (e.g., `v1.0.0`)
- Pull requests to `main`
- Manual workflow dispatch

**Pipeline Stages**:
1. **Security Scan**: Trivy vulnerability scanning on source code
2. **Build & Test**: Go tests with coverage reporting
3. **Container Build**: Multi-arch builds (AMD64 + ARM64) with caching
4. **Container Scan**: Trivy + Grype security scanning
5. **Integration Test**: Health checks and API validation
6. **Publish**: Push to GitHub Container Registry

**Workflow Files**:
- `.github/workflows/container.yml` - Main build pipeline
- `.github/workflows/security-scan.yml` - Scheduled security scans

### GitHub Container Registry

All images are published to GitHub Container Registry (ghcr.io):

```bash
# Pull latest images
docker pull ghcr.io/finoptimize/agentaflow-sro-community:web-dashboard-latest
docker pull ghcr.io/finoptimize/agentaflow-sro-community:k8s-scheduler-latest
docker pull ghcr.io/finoptimize/agentaflow-sro-community:prometheus-demo-latest

# Pull specific version
docker pull ghcr.io/finoptimize/agentaflow-sro-community:web-dashboard-v1.0.0
```

### Image Tags

**Tag Strategy**:
- `latest` - Latest build from main branch
- `<component>-latest` - Latest component-specific build
- `<component>-v1.0.0` - Specific version release
- `<component>-main-<sha>` - Branch + commit SHA
- `<component>-1.0` - Major.minor version

**Examples**:
```bash
# Production: Use version tags
docker pull ghcr.io/finoptimize/agentaflow-sro-community:web-dashboard-v1.0.0

# Development: Use latest
docker pull ghcr.io/finoptimize/agentaflow-sro-community:web-dashboard-latest

# Specific commit: Use SHA tags
docker pull ghcr.io/finoptimize/agentaflow-sro-community:web-dashboard-main-abc1234
```

### Security & Provenance

**Automated Security**:
- Daily vulnerability scans
- CodeQL analysis
- Secret detection with Gitleaks
- SBOM (Software Bill of Materials) generation
- SLSA provenance for release builds

**View Security Reports**:
- [Security Advisories](https://github.com/Finoptimize/agentaflow-sro-community/security/advisories)
- [Dependabot Alerts](https://github.com/Finoptimize/agentaflow-sro-community/security/dependabot)
- [Code Scanning](https://github.com/Finoptimize/agentaflow-sro-community/security/code-scanning)

### Automated Dependency Updates

Dependabot monitors and updates:
- Go module dependencies (weekly)
- Docker base images (weekly)
- GitHub Actions versions (weekly)

**Configuration**: `.github/dependabot.yml`

## 🔄 Updates and Maintenance

### Pulling Latest Images

```bash
# From GitHub Container Registry
docker pull ghcr.io/finoptimize/agentaflow-sro-community:web-dashboard-latest

# Update compose stack to use registry images
docker-compose pull
docker-compose up -d
```

### Pruning Old Images

```bash
# Remove unused images
docker image prune -a

# Remove unused volumes
docker volume prune
```

## 🎯 Production Deployment

### Recommended Docker Compose for Production

```yaml
version: '3.8'
services:
  agentaflow-dashboard:
    image: ghcr.io/finoptimize/agentaflow-sro-community:web-dashboard
    restart: always
    deploy:
      resources:
        limits:
          cpus: '0.5'
          memory: 256M
        reservations:
          cpus: '0.25'
          memory: 128M
    healthcheck:
      test: ["CMD", "/app/agentaflow-dashboard", "--health-check"]
      interval: 30s
      timeout: 10s
      retries: 3
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
```

### Resource Limits

Set appropriate limits:

```bash
docker run -p 9000:9000 -p 9001:9001 \
  --memory="256m" \
  --memory-reservation="128m" \
  --cpus="0.5" \
  agentaflow-dashboard
```

## 📚 Additional Resources

- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [Dockerfile Best Practices](https://docs.docker.com/develop/develop-images/dockerfile_best-practices/)
- [Container Security](https://docs.docker.com/engine/security/)
- [GitHub Packages](https://docs.github.com/en/packages)

## 🆘 Support

For issues or questions:
- GitHub Issues: https://github.com/Finoptimize/agentaflow-sro-community/issues
- Documentation: See main [README.md](../README.md)
- Container Strategy: See [CONTAINER.md](../CONTAINER.md)
