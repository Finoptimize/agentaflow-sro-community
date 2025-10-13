# AgentaFlow Prometheus/Grafana Integration Demo

This demo showcases the complete Prometheus and Grafana integration for AgentaFlow SRO Community Edition, providing enterprise-grade monitoring and visualization for GPU infrastructure.

## 🎯 Overview

The demo demonstrates:
- Real-time GPU metrics collection and export to Prometheus
- Comprehensive Grafana dashboards for visualization
- Kubernetes-native monitoring stack deployment
- Cost tracking and efficiency analytics
- Alert management and health monitoring

## 🚀 Quick Start

### 1. Run the Demo Application

```bash
# From the project root
cd examples/demo/prometheus-grafana
go run main.go
```

The demo will start and display:
```
🚀 AgentaFlow Prometheus/Grafana Integration Demo
===============================================
📊 Registering Prometheus metrics...
🔧 Starting services...
🌐 Starting Prometheus metrics server on :8080...
✅ All services started successfully!

🎯 Integration Points:
   • Prometheus metrics: http://localhost:8080/metrics
   • Health endpoint: http://localhost:8080/health

📊 Available Metrics:
   • agentaflow_gpu_utilization_percent
   • agentaflow_gpu_temperature_celsius
   • agentaflow_gpu_memory_used_bytes
   • agentaflow_gpu_health_status
   • agentaflow_workloads_pending
   • agentaflow_cost_total_dollars
   • agentaflow_gpu_efficiency_score
```

### 2. Deploy Monitoring Stack (Kubernetes)

Deploy Prometheus:
```bash
kubectl apply -f ../k8s/monitoring/prometheus.yaml
```

Deploy Grafana:
```bash
kubectl apply -f ../k8s/monitoring/grafana.yaml
```

### 3. Access Grafana Dashboard

Port-forward Grafana service:
```bash
kubectl port-forward svc/grafana-service 3000:3000 -n agentaflow-monitoring
```

Open http://localhost:3000 in your browser:
- **Username**: `admin`
- **Password**: `agentaflow123`

### 4. Import Dashboard

1. Go to **Dashboards** > **Import**
2. Upload `../monitoring/grafana-dashboard.json`
3. Select Prometheus data source
4. Click **Import**

## 📊 Metrics Overview

### GPU Metrics
- **Utilization**: `agentaflow_gpu_utilization_percent`
- **Temperature**: `agentaflow_gpu_temperature_celsius`
- **Memory**: `agentaflow_gpu_memory_used_bytes`, `agentaflow_gpu_memory_total_bytes`
- **Power**: `agentaflow_gpu_power_draw_watts`, `agentaflow_gpu_power_limit_watts`
- **Clock Speeds**: `agentaflow_gpu_clock_graphics_mhz`, `agentaflow_gpu_clock_memory_mhz`
- **Health**: `agentaflow_gpu_health_status`
- **Efficiency**: `agentaflow_gpu_efficiency_score`

### Workload Metrics
- **Pending Jobs**: `agentaflow_workloads_pending`
- **Running Jobs**: `agentaflow_workloads_running`
- **Completed Jobs**: `agentaflow_workloads_completed`
- **Scheduling Duration**: `agentaflow_scheduling_duration_seconds`
- **Allocation Efficiency**: `agentaflow_gpu_allocation_efficiency`

### Cost Metrics
- **Total Cost**: `agentaflow_cost_total_dollars`
- **Hourly Rates**: `agentaflow_cost_per_hour_dollars`
- **GPU Hours**: `agentaflow_gpu_hours_consumed`
- **Monthly Estimates**: `agentaflow_estimated_monthly_cost_dollars`

### System Metrics
- **Cluster Utilization**: `agentaflow_cluster_utilization_percent`
- **GPU Availability**: `agentaflow_gpus_available`, `agentaflow_gpus_total`
- **Component Health**: `agentaflow_component_health_status`
- **Uptime**: `agentaflow_system_uptime_seconds`
- **Active Alerts**: `agentaflow_active_alerts`

## 🔧 Configuration

### Prometheus Configuration
The demo uses these key settings:
```go
prometheusConfig := observability.PrometheusConfig{
    MetricsPrefix: "agentaflow",
    EnabledMetrics: map[string]bool{
        "gpu_metrics":        true,
        "scheduling_metrics": true,
        "serving_metrics":    true,
        "cost_metrics":      true,
        "system_metrics":    true,
    },
    MetricLabels: map[string]string{
        "instance": "demo",
        "version":  "community",
    },
}
```

### Alert Thresholds
```go
customThresholds := observability.GPUAlertThresholds{
    HighTemperature:     70.0,
    CriticalTemperature: 85.0,
    HighMemoryUsage:     80.0,
    CriticalMemoryUsage: 95.0,
    HighPowerUsage:      85.0,
    CriticalPowerUsage:  95.0,
    LowUtilization:      15.0,
    HighUtilization:     90.0,
}
```

### Cost Configuration
```go
awsCostConfig := observability.GPUCostConfiguration{
    CostPerHour: map[string]float64{
        "a100":    3.06,  // AWS p4d.xlarge
        "v100":    3.06,  // AWS p3.2xlarge
        "t4":      0.526, // AWS g4dn.xlarge
        "rtx":     1.20,  // Custom RTX pricing
        "generic": 1.50,  // Default
    },
    UseUtilizationFactor: true,
    MinUtilizationFactor: 0.15,
    IdleCostReduction:    0.20,
    CloudProvider:        "aws",
    Region:              "us-west-2",
    SpotInstanceDiscount: 0.60,
}
```

## 📈 Dashboard Panels

The Grafana dashboard includes 8 comprehensive panels:

1. **GPU Utilization** - Real-time utilization across all GPUs
2. **Temperature Monitoring** - Temperature trends with thresholds
3. **Memory Usage** - Memory utilization and availability
4. **Power Consumption** - Power draw vs limits
5. **Workload Distribution** - Job scheduling and distribution
6. **Cost Tracking** - Real-time cost analysis
7. **System Efficiency** - Performance and efficiency metrics
8. **Health Status** - Overall system health indicators

## 🎮 Interactive Demo Features

### Real-time Metrics Generation
The demo generates realistic metrics that simulate:
- **GPU utilization patterns**: Waves simulating training and inference workloads
- **Temperature correlation**: Temperature increases with utilization
- **Memory patterns**: Dynamic memory allocation and release
- **Cost calculations**: Real-time cost tracking with utilization factors
- **Alert scenarios**: Periodic alerts to demonstrate monitoring

### Endpoints Available
- **Metrics**: http://localhost:8080/metrics - Prometheus format metrics
- **Health**: http://localhost:8080/health - Service health check
- **Debug**: Live debugging output in terminal

### Metric Patterns
- **Utilization**: Sine wave pattern (20-80%) simulating workload cycles
- **Temperature**: Correlated with utilization (35-80°C)
- **Memory**: Independent allocation patterns per GPU
- **Workloads**: Periodic job completion and queuing
- **Costs**: Dynamic cost calculation with AWS pricing

## 🔍 Troubleshooting

### Common Issues

**Metrics not appearing in Prometheus**
```bash
# Check if demo is running
curl http://localhost:8080/metrics

# Verify Prometheus config
kubectl logs -n agentaflow-monitoring prometheus-deployment-xxx
```

**Grafana dashboard not loading**
```bash
# Check Grafana logs
kubectl logs -n agentaflow-monitoring grafana-deployment-xxx

# Verify data source connection
# Go to Configuration > Data Sources in Grafana UI
```

**Kubernetes deployment issues**
```bash
# Check namespace
kubectl get namespace agentaflow-monitoring

# Check all resources
kubectl get all -n agentaflow-monitoring

# Check ConfigMaps
kubectl get configmaps -n agentaflow-monitoring
```

### Debug Mode
Enable verbose logging:
```go
// Add to main function
log.SetLevel(log.DebugLevel)
```

## 🏗️ Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   AgentaFlow    │───▶│   Prometheus     │───▶│    Grafana      │
│  GPU Metrics    │    │   Exporter       │    │   Dashboard     │
│   Collector     │    │  (:8080/metrics) │    │ (localhost:3000)│
└─────────────────┘    └──────────────────┘    └─────────────────┘
         │                        │                       │
         ▼                        ▼                       ▼
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│ GPU Integration │    │ Metrics Storage  │    │  Visualization  │
│    Service      │    │   & Alerting     │    │   & Analysis    │
└─────────────────┘    └──────────────────┘    └─────────────────┘
```

## 🧪 Testing

Run integration tests:
```bash
# Test Prometheus metrics endpoint
curl -s http://localhost:8080/metrics | grep agentaflow

# Test health endpoint
curl -s http://localhost:8080/health

# Test Grafana API
curl -s -u admin:agentaflow123 http://localhost:3000/api/health
```

## 📚 Next Steps

1. **Production Setup**: Adapt configurations for production environments
2. **Custom Dashboards**: Create additional dashboards for specific use cases
3. **Alert Rules**: Implement custom Prometheus alerting rules
4. **Scaling**: Configure for multi-cluster monitoring
5. **Integration**: Connect with existing monitoring infrastructure

For production deployment, see the main [README.md](../../README.md) for complete setup instructions.