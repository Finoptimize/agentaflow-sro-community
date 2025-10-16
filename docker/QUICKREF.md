# 🐳 AgentaFlow Docker Quick Reference

## One-Line Demos

```bash
# Web Dashboard (most popular)
docker run -p 9000:9000 -p 9001:9001 agentaflow-sro:web-dashboard

# Complete Stack
docker-compose up -d

# Build All Images
docker-compose build
```

## Common Commands

### Build
```bash
# Single image
docker build -f docker/Dockerfile.web-dashboard -t agentaflow-sro:web-dashboard .

# All images
./docker/build.ps1  # Windows
./docker/build.sh   # Linux/Mac
```

### Run
```bash
# Foreground (see logs)
docker run -p 9000:9000 -p 9001:9001 agentaflow-sro:web-dashboard

# Background (detached)
docker run -d -p 9000:9000 -p 9001:9001 --name agentaflow agentaflow-sro:web-dashboard

# With volume
docker run -d -p 9000:9000 -p 9001:9001 -v agentaflow-data:/data agentaflow-sro:web-dashboard
```

### Manage
```bash
# View logs
docker logs -f agentaflow

# Stop container
docker stop agentaflow

# Remove container
docker rm agentaflow

# Shell access (debug images only)
docker exec -it agentaflow sh
```

### Docker Compose
```bash
# Start
docker-compose up -d

# Stop
docker-compose down

# Restart
docker-compose restart

# View logs
docker-compose logs -f

# Rebuild and start
docker-compose up -d --build
```

## Ports

| Service | Port | Description |
|---------|------|-------------|
| Web Dashboard | 9000 | Main UI |
| Metrics | 9001 | Prometheus metrics |
| Grafana | 3000 | Visualization |
| Prometheus | 9090 | Metrics database |
| K8s Scheduler | 8080 | Kubernetes API |

## Environment Variables

```bash
# Debug mode
docker run -p 9000:9000 -e LOG_LEVEL=debug agentaflow-sro:web-dashboard

# Custom thresholds
docker run -p 9000:9000 -e ALERT_THRESHOLDS=temperature:85,utilization:90 agentaflow-sro:web-dashboard

# Disable GPU simulation
docker run -p 9000:9000 -e GPU_SIMULATION=false agentaflow-sro:web-dashboard
```

## Troubleshooting

```bash
# Check if running
docker ps | grep agentaflow

# View all containers (including stopped)
docker ps -a

# Check logs
docker logs agentaflow

# Check health
docker inspect --format='{{.State.Health.Status}}' agentaflow

# Resource usage
docker stats agentaflow

# Remove all stopped containers
docker container prune

# Remove unused images
docker image prune -a

# Remove all (nuclear option)
docker system prune -a --volumes
```

## Access URLs

- **Dashboard**: http://localhost:9000
- **Metrics**: http://localhost:9001/metrics
- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000 (admin/agentaflow123)

## Quick Tests

```bash
# Test dashboard is running
curl http://localhost:9000/health

# Test metrics endpoint
curl http://localhost:9001/metrics

# Test WebSocket
wscat -c ws://localhost:9000/ws
```
