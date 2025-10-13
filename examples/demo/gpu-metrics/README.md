# GPU Metrics Collection Demo

This demo showcases AgentaFlow's real-time GPU metrics collection capabilities.

## Usage

```bash
# Run the demo
go run main.go

# Or build and run
go build -o gpu-metrics main.go
./gpu-metrics
```

## Features Demonstrated

- Real-time GPU metrics collection using nvidia-smi
- Cost tracking with configurable hourly rates
- System overview and efficiency metrics
- Alert thresholds and health monitoring
- Integration with AgentaFlow's monitoring service

For more comprehensive monitoring with Prometheus/Grafana, see the `../prometheus-grafana/` demo.