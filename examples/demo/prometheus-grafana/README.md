# Prometheus/Grafana Integration Demo

This demo showcases the complete Prometheus and Grafana integration for AgentaFlow SRO Community Edition.

## Usage

```bash
# Run the demo
go run main.go

# Or build and run
go build -o prometheus-grafana main.go
./prometheus-grafana
```

## What This Demo Shows

- Real-time GPU metrics export to Prometheus format
- HTTP endpoints for Prometheus scraping (`http://localhost:8080/metrics`)
- Realistic metric generation with GPU utilization patterns
- Cost tracking with AWS pricing integration
- Alert scenarios and health monitoring
- Integration with Kubernetes monitoring stack

## Next Steps

1. Deploy Prometheus: `kubectl apply -f ../../k8s/monitoring/prometheus.yaml`
2. Deploy Grafana: `kubectl apply -f ../../k8s/monitoring/grafana.yaml`
3. Import dashboard: `../../monitoring/grafana-dashboard.json`

For complete setup instructions, see `../PROMETHEUS_GRAFANA_DEMO.md`.