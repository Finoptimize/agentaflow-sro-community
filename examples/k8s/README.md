# Kubernetes GPU Scheduling Integration

This directory contains the Kubernetes integration for AgentaFlow GPU scheduling, enabling intelligent GPU resource management across Kubernetes clusters.

## Features

- **Kubernetes-Native GPU Scheduling**: Native integration with Kubernetes for GPU workload scheduling
- **Multiple Scheduling Strategies**: Support for least-utilized, best-fit, priority-based, and round-robin scheduling
- **Real-time GPU Monitoring**: DaemonSet-based GPU monitoring on each node
- **Custom Resource Definitions**: GPUWorkload and GPUNode CRDs for managing GPU resources
- **Health Monitoring**: Comprehensive GPU health checks and alerting
- **CLI Management**: Command-line interface for managing GPU workloads and monitoring

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    Kubernetes Cluster                           │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐ │
│  │   GPU Node 1    │  │   GPU Node 2    │  │   GPU Node N    │ │
│  │ ┌─────────────┐ │  │ ┌─────────────┐ │  │ ┌─────────────┐ │ │
│  │ │GPU Monitor  │ │  │ │GPU Monitor  │ │  │ │GPU Monitor  │ │ │
│  │ │ DaemonSet   │ │  │ │ DaemonSet   │ │  │ │ DaemonSet   │ │ │
│  │ └─────────────┘ │  │ └─────────────┘ │  │ └─────────────┘ │ │
│  │ ┌─────────────┐ │  │ ┌─────────────┐ │  │ ┌─────────────┐ │ │
│  │ │   GPUs      │ │  │ │   GPUs      │ │  │ │   GPUs      │ │ │
│  │ │ A100/V100   │ │  │ │ A100/V100   │ │  │ │ A100/V100   │ │ │
│  │ └─────────────┘ │  │ └─────────────┘ │  │ └─────────────┘ │ │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘ │
├─────────────────────────────────────────────────────────────────┤
│                 ┌─────────────────────────────────┐             │
│                 │     GPU Scheduler Service       │             │
│                 │  ┌─────────────────────────────┐│             │
│                 │  │   Scheduling Engine         ││             │
│                 │  │ - Least Utilized            ││             │
│                 │  │ - Best Fit                  ││             │
│                 │  │ - Priority Based            ││             │
│                 │  │ - Round Robin               ││             │
│                 │  └─────────────────────────────┘│             │
│                 │  ┌─────────────────────────────┐│             │
│                 │  │   Resource Manager          ││             │
│                 │  │ - GPU Discovery             ││             │
│                 │  │ - Workload Queue            ││             │
│                 │  │ - Status Tracking           ││             │
│                 │  └─────────────────────────────┘│             │
│                 └─────────────────────────────────┘             │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐ │
│  │  GPUWorkload    │  │   GPUNode       │  │      CLI        │ │
│  │      CRD        │  │     CRD         │  │   Interface     │ │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

## Installation

### Prerequisites

- Kubernetes cluster v1.20+
- NVIDIA GPU nodes with drivers installed
- nvidia-docker runtime configured
- `kubectl` configured to access your cluster

### Quick Setup

1. **Apply RBAC and Namespace**:
   ```bash
   kubectl apply -f examples/k8s/scheduler-deployment.yaml
   ```

2. **Build and Deploy the Scheduler**:
   ```bash
   # Build Docker image
   docker build -f examples/k8s/Dockerfile -t agentaflow/k8s-gpu-scheduler:latest .
   
   # Push to your registry
   docker push agentaflow/k8s-gpu-scheduler:latest
   
   # Deploy scheduler
   kubectl apply -f examples/k8s/scheduler-deployment.yaml
   ```

3. **Verify Installation**:
   ```bash
   kubectl get pods -n agentaflow
   kubectl logs -f deployment/agentaflow-gpu-scheduler -n agentaflow
   ```

### Manual Installation

1. **Create Namespace**:
   ```bash
   kubectl create namespace agentaflow
   ```

2. **Apply RBAC**:
   ```bash
   kubectl apply -f examples/k8s/rbac.yaml
   ```

3. **Deploy Scheduler**:
   ```bash
   kubectl apply -f examples/k8s/scheduler.yaml
   ```

4. **Deploy GPU Monitor DaemonSet**:
   ```bash
   kubectl apply -f examples/k8s/monitor-daemonset.yaml
   ```

## Usage

### CLI Interface

The scheduler includes a comprehensive CLI for managing GPU workloads:

```bash
# Show overall status
./k8s-gpu-scheduler --mode=cli status

# List GPU nodes
./k8s-gpu-scheduler --mode=cli nodes

# List workloads
./k8s-gpu-scheduler --mode=cli workloads

# Submit a workload
./k8s-gpu-scheduler --mode=cli submit examples/k8s/pytorch-training-workload.yaml

# Change scheduling strategy
./k8s-gpu-scheduler --mode=cli strategy best_fit

# Watch status in real-time
./k8s-gpu-scheduler --mode=cli watch

# Check GPU health
./k8s-gpu-scheduler --mode=cli health

# Generate workload template
./k8s-gpu-scheduler --mode=cli generate-template my-workload.yaml
```

### Submitting GPU Workloads

Create a GPUWorkload YAML file:

```yaml
apiVersion: agentaflow.io/v1
kind: GPUWorkload
metadata:
  name: my-training-job
  namespace: agentaflow
spec:
  priority: 5
  gpuMemoryRequired: 8192  # 8GB
  schedulingStrategy: "least_utilized"
  gpuRequirements:
    minGPUMemory: 8192
    gpuCount: 1
    exclusiveAccess: true
  podTemplate:
    spec:
      containers:
      - name: training
        image: pytorch/pytorch:latest-gpu
        resources:
          limits:
            nvidia.com/gpu: 1
```

Submit the workload:

```bash
./k8s-gpu-scheduler --mode=cli submit my-workload.yaml
```

### Monitoring GPU Resources

View GPU node status:

```bash
kubectl get nodes -l agentaflow.gpu/enabled=true
kubectl describe node <gpu-node-name>
```

Check GPU utilization:

```bash
./k8s-gpu-scheduler --mode=cli nodes
./k8s-gpu-scheduler --mode=cli health
```

### Scheduling Strategies

The scheduler supports multiple strategies:

- **`least_utilized`**: Schedule on GPU with lowest utilization (default)
- **`best_fit`**: Schedule on GPU with just enough free memory
- **`priority`**: Schedule high-priority workloads first
- **`round_robin`**: Distribute workloads evenly across GPUs

Change strategy at runtime:

```bash
./k8s-gpu-scheduler --mode=cli strategy priority
```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `KUBERNETES_NAMESPACE` | Namespace to operate in | `agentaflow` |
| `SCHEDULING_STRATEGY` | Default scheduling strategy | `least_utilized` |
| `GPU_UTILIZATION_GOAL` | Target GPU utilization percentage | `80` |
| `MONITOR_INTERVAL` | GPU monitoring interval | `15s` |
| `SCHEDULING_INTERVAL` | Scheduling cycle interval | `5s` |

### Node Labels and Annotations

The system uses these labels and annotations:

**Labels:**
- `agentaflow.gpu/enabled=true`: Marks GPU-enabled nodes
- `agentaflow.gpu/schedulable=true`: Indicates schedulable GPUs available
- `agentaflow.gpu/count=N`: Number of GPUs on node
- `agentaflow.gpu/utilization-tier=low|medium|high`: GPU utilization level

**Annotations:**
- `agentaflow.gpu/devices`: JSON array of GPU device information
- `agentaflow.gpu/status`: Current GPU status and metrics
- `agentaflow.gpu/last-update`: Timestamp of last status update

## Troubleshooting

### Common Issues

1. **Scheduler not finding GPU nodes**:
   ```bash
   # Check if nodes have GPU labels
   kubectl get nodes -l agentaflow.gpu/enabled=true
   
   # Check monitor DaemonSet logs
   kubectl logs -l app=agentaflow-gpu-monitor -n agentaflow
   ```

2. **Workloads stuck in Pending state**:
   ```bash
   # Check GPU availability
   ./k8s-gpu-scheduler --mode=cli nodes
   
   # Check workload requirements vs available resources
   ./k8s-gpu-scheduler --mode=cli workloads
   ```

3. **GPU Monitor not updating status**:
   ```bash
   # Check nvidia-smi is available
   kubectl exec -it <monitor-pod> -- nvidia-smi
   
   # Check monitor permissions
   kubectl describe daemonset agentaflow-gpu-monitor -n agentaflow
   ```

### Debug Mode

Enable debug logging:

```bash
./k8s-gpu-scheduler --mode=scheduler --log-level=debug
```

### Health Checks

The scheduler exposes health endpoints:

- `GET /healthz`: Liveness probe
- `GET /readyz`: Readiness probe
- `GET /metrics`: Prometheus metrics

## Metrics and Monitoring

### Prometheus Metrics

The scheduler exposes Prometheus metrics on port 9090:

- `agentaflow_gpu_total`: Total number of GPUs
- `agentaflow_gpu_available`: Available GPUs
- `agentaflow_gpu_utilization`: Average GPU utilization
- `agentaflow_workloads_pending`: Pending workloads
- `agentaflow_workloads_running`: Running workloads
- `agentaflow_scheduling_duration`: Scheduling cycle duration

### Grafana Dashboard

Import the provided Grafana dashboard for visualization:

```bash
kubectl apply -f examples/k8s/grafana-dashboard.json
```

## Security Considerations

1. **RBAC**: The scheduler requires cluster-wide permissions to manage nodes and pods
2. **Privileged Access**: GPU monitor runs with privileged access to query GPU status
3. **Resource Isolation**: Workloads are isolated using Kubernetes namespaces and resource quotas
4. **Network Policies**: Apply network policies to restrict scheduler communication

## Performance Tuning

### Scheduler Performance

- Adjust `SCHEDULING_INTERVAL` based on cluster size
- Use node selectors to limit scheduling scope
- Configure resource requests/limits for scheduler pods

### GPU Utilization

- Set appropriate `GPU_UTILIZATION_GOAL` for your workloads
- Use priority-based scheduling for critical workloads
- Monitor temperature and power usage thresholds

## Integration with Other Systems

### Kubeflow

The scheduler integrates with Kubeflow for ML pipeline scheduling:

```yaml
apiVersion: kubeflow.org/v1
kind: PyTorchJob
metadata:
  annotations:
    agentaflow.gpu/scheduling-strategy: "priority"
    agentaflow.gpu/priority: "10"
```

### Prometheus/Grafana

Configure monitoring stack:

```bash
# Add Prometheus ServiceMonitor
kubectl apply -f examples/k8s/prometheus-servicemonitor.yaml

# Import Grafana dashboard
kubectl apply -f examples/k8s/grafana-dashboard.yaml
```

### Slack/Teams Integration

Configure alerts for GPU issues:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: alert-webhook
data:
  webhook-url: <base64-encoded-webhook-url>
```

## Contributing

1. Fork the repository
2. Create feature branch
3. Add tests for new functionality
4. Run integration tests with kind/minikube
5. Submit pull request

## License

Apache License 2.0 - see [LICENSE](../../LICENSE) for details.