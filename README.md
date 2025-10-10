# AgentaFlow SRO Community Edition

**AI Infrastructure Tooling & Optimization Platform**

Deploy and manage AI infrastructure more efficiently with tools for GPU orchestration, model serving optimization, and comprehensive observability.

[![License: Apache 2.0](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org/dl/)

## 🚀 Features

### GPU Orchestration & Scheduling
Tools that optimize GPU utilization across workloads, reducing waste:
- **Smart Scheduling**: Multiple strategies (least-utilized, best-fit, priority, round-robin)
- **Resource Optimization**: Reduce GPU idle time by up to 40%
- **Workload Management**: Efficient queuing and distribution across GPU clusters
- **Real-time Monitoring**: Track utilization, memory, temperature, and power

### AI Model Serving Optimization
Software that reduces inference costs through better batching, caching, and routing:
- **Request Batching**: Improve throughput by 3-5x with intelligent batching
- **Smart Caching**: Reduce latency by up to 50% with TTL-based caching
- **Load Balancing**: Multiple routing strategies for optimal distribution
- **Cost Reduction**: Minimize inference costs through efficient resource use

### Observability Tools for AI Systems
Monitoring, debugging, and cost tracking for LLM applications and training runs:
- **Comprehensive Metrics**: Counters, gauges, and histograms for all operations
- **Cost Tracking**: Detailed tracking of GPU hours, tokens, and operational costs
- **Distributed Tracing**: Full request tracing across distributed systems
- **Debug Utilities**: Multi-level logging with performance analysis

## 📦 Installation

```bash
go get github.com/Finoptimize/agentaflow-sro-community
```

## 🎯 Quick Start

Run the comprehensive demo:

```bash
cd cmd/agentaflow
go run main.go
```

This demonstrates all three core components working together.

## 💡 Usage Examples

### GPU Scheduling

```go
import "github.com/Finoptimize/agentaflow-sro-community/pkg/gpu"

scheduler := gpu.NewScheduler(gpu.StrategyLeastUtilized)

// Register GPU
gpu1 := &gpu.GPU{
    ID:          "gpu-0",
    Name:        "NVIDIA A100",
    MemoryTotal: 40960,
    Available:   true,
}
scheduler.RegisterGPU(gpu1)

// Submit and schedule workload
workload := &gpu.Workload{
    ID:             "training-job-1",
    MemoryRequired: 32768,
    Priority:       1,
}
scheduler.SubmitWorkload(workload)
scheduler.Schedule()
```

### Model Serving

```go
import "github.com/Finoptimize/agentaflow-sro-community/pkg/serving"

servingMgr := serving.NewServingManager(&serving.BatchConfig{
    MaxBatchSize: 32,
    MaxWaitTime:  100 * time.Millisecond,
}, 5*time.Minute)

// Process inference with automatic caching
response, _ := servingMgr.SubmitInferenceRequest(&serving.InferenceRequest{
    ModelID: "gpt-model",
    Input:   []byte("Your prompt"),
})
```

### Observability

```go
import "github.com/Finoptimize/agentaflow-sro-community/pkg/observability"

monitor := observability.NewMonitoringService(10000)

// Track costs
monitor.RecordCost(observability.CostEntry{
    Operation:  "inference",
    GPUHours:   0.5,
    TokensUsed: 1000,
    Cost:       2.50,
})

// Get cost summary
summary := monitor.GetCostSummary(startTime, endTime)
```

## 📊 Key Benefits

| Component | Benefit | Impact |
|-----------|---------|--------|
| GPU Scheduling | Optimized utilization | Up to 40% reduction in GPU idle time |
| Request Batching | Improved throughput | 3-5x increase in requests/second |
| Response Caching | Reduced latency | Up to 50% faster responses |
| Cost Tracking | Better budgeting | Full visibility into AI infrastructure costs |

## 🏗️ Architecture

```
agentaflow-sro-community/
├── pkg/
│   ├── gpu/           # GPU orchestration and scheduling
│   ├── serving/       # Model serving optimization
│   └── observability/ # Monitoring and debugging
├── cmd/agentaflow/    # Main CLI application
└── examples/          # Usage examples
```

## 📖 Documentation

For detailed documentation, see [DOCUMENTATION.md](DOCUMENTATION.md)

Topics covered:
- Detailed API reference
- Scheduling strategies
- Performance optimization
- Configuration options
- Use cases and examples

## 🎓 Use Cases

1. **ML Training Clusters** - Optimize GPU allocation across multiple training jobs
2. **LLM Inference Services** - Reduce costs with intelligent batching and caching
3. **Multi-Model Deployments** - Load balance requests across model instances
4. **Cost Optimization** - Track and minimize AI infrastructure spending
5. **Performance Debugging** - Identify and resolve bottlenecks

## 🛠️ Requirements

- Go 1.21 or higher
- No external dependencies for core functionality

## 📝 License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## 🤝 Contributing

Contributions are welcome! This is a community edition focused on providing accessible AI infrastructure optimization tools.

## 🗺️ Roadmap

- Kubernetes integration for GPU scheduling
- Real-time GPU metrics collection
- Prometheus/Grafana integration
- Web dashboard for monitoring
- OpenTelemetry support for tracing

## 🚀 Enterprise Edition (Coming Soon)

Looking for advanced features for production environments? Our **Enterprise Edition** will include:

- **Multi-cluster Orchestration**: Manage GPU resources across multiple Kubernetes clusters
- **Multi-cloud GPU resource support**: Support for running in Azure, Google Cloud, Vercel, or DigitalOcean 
- **Advanced Scheduling Algorithms**: Cost optimization algorithms and priority queues for enterprise workloads  
- **RBAC and Audit Logs**: Role-based access control and comprehensive audit logging
- **Enterprise Integrations**: Slack alerts, DataDog monitoring, and other enterprise tools
- **SLA Support**: Guaranteed service levels with dedicated support
- **Usage-based Billing Features**: Advanced cost tracking and billing automation

*Contact us for early access and enterprise pricing.*

## 📞 Support

For questions, issues, or contributions, please open an issue on GitHub.

---

**Built with ❤️ by FinOptimize for AgentaFlow**