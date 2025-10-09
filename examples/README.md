# Examples

This directory contains standalone examples demonstrating the core features of AgentaFlow SRO.

## Running Examples

Each example is a standalone program. Run them individually:

### GPU Scheduling Example

Demonstrates GPU orchestration and workload scheduling with different strategies:

```bash
cd examples
go run gpu_scheduling.go
```

This example shows:
- Registering multiple GPUs with different specifications
- Submitting various AI workloads (training, inference, fine-tuning)
- Scheduling workloads using different strategies
- Monitoring GPU utilization and memory usage
- Comparing scheduling algorithm performance

### Model Serving Example

Demonstrates AI model serving optimization with batching, caching, and routing:

```bash
cd examples
go run model_serving.go
```

This example shows:
- Registering multiple AI models
- Setting up model instances with load balancing
- Processing inference requests with automatic caching
- Batch processing for improved throughput
- Comparing different routing strategies
- Monitoring cache hit rates and performance metrics

### Observability Example

Demonstrates comprehensive monitoring, cost tracking, and debugging:

```bash
cd examples
go run observability.go
```

This example shows:
- Recording various performance metrics
- Tracking system events with different severity levels
- Detailed cost tracking for inference and training operations
- Distributed tracing for request analysis
- Performance bottleneck detection
- Multi-level debug logging
- Querying metrics, events, and logs

## What You'll Learn

### GPU Scheduling
- How to optimize GPU utilization across workloads
- Different scheduling strategies and their tradeoffs
- Resource allocation and memory management
- Monitoring cluster health and performance

### Model Serving
- Request batching for improved throughput
- Intelligent caching to reduce latency
- Load balancing across model instances
- Cache management and optimization

### Observability
- Comprehensive metrics collection
- Cost tracking and analysis
- Distributed request tracing
- Debug logging and performance profiling

## Integration

These examples can be used as templates for integrating AgentaFlow into your own AI infrastructure:

1. **GPU Scheduler** - Integrate with your GPU cluster management
2. **Model Serving** - Add to your inference service
3. **Observability** - Incorporate into your monitoring stack

## Next Steps

After running the examples:

1. Review the [main documentation](../DOCUMENTATION.md) for detailed API reference
2. Explore the source code in `pkg/` directories
3. Customize the examples for your use case
4. Check out the main demo: `go run ../cmd/agentaflow/main.go`

## Notes

- Each example is self-contained and can run independently
- Examples use simulated data - integrate with real GPU metrics in production
- Performance numbers are illustrative - actual results depend on your infrastructure
