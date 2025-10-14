# OpenTelemetry Distributed Tracing - AgentaFlow SRO Community Edition

## Overview

AgentaFlow SRO Community Edition includes comprehensive OpenTelemetry distributed tracing support, providing visibility into GPU scheduling operations, model serving requests, metrics collection, and cost calculations across the entire platform.

## Features

- **🔍 Distributed Tracing**: Full request tracing across all AgentaFlow components
- **🏷️  Rich Span Attributes**: GPU IDs, workload details, performance metrics, costs
- **📊 Multiple Exporters**: Support for Jaeger, OTLP, and stdout exporters
- **🔗 Trace Correlation**: Automatic correlation of related operations
- **⚡ Performance Monitoring**: Operation durations, queue sizes, utilization metrics
- **🎯 Configurable Sampling**: Control trace volume with sampling ratios
- **🏥 Health Checks**: Built-in tracing system health monitoring

## Quick Start

### 1. Setup Jaeger (Local Development)

```bash
# Run Jaeger all-in-one container
docker run -d \
  --name jaeger \
  -p 16686:16686 \
  -p 14268:14268 \
  jaegertracing/all-in-one:latest

# Access Jaeger UI at http://localhost:16686
```

### 2. Configure Tracing

Create a tracing configuration file (`tracing-config.yaml`):

```yaml
service_name: "agentaflow-sro"
service_version: "0.1.0"
environment: "development"

tracing:
  enabled: true
  sampling_ratio: 1.0  # Trace 100% of requests
  
  exporters:
    jaeger:
      enabled: true
      endpoint: "http://localhost:14268/api/traces"
    
    stdout:
      enabled: false
      pretty_print: true
```

### 3. Initialize Tracing in Your Application

```go
package main

import (
    "context"
    "log"
    
    "github.com/Finoptimize/agentaflow-sro-community/pkg/gpu"
    "github.com/Finoptimize/agentaflow-sro-community/pkg/observability"
)

func main() {
    // Initialize tracing
    config := &observability.TracingConfig{
        ServiceName:    "agentaflow-sro",
        ServiceVersion: "0.1.0",
        Environment:    "production",
        JaegerEndpoint: "http://localhost:14268/api/traces",
        SamplingRatio:  1.0,
        EnableJaeger:   true,
    }
    
    tracingIntegration, err := observability.NewTracingIntegration(config)
    if err != nil {
        log.Fatalf("Failed to initialize tracing: %v", err)
    }
    defer tracingIntegration.Shutdown(context.Background())
    
    // Wrap your components with tracing
    scheduler := &gpu.Scheduler{}
    tracedScheduler := tracingIntegration.WrapGPUScheduler(scheduler)
    
    // Your application logic with automatic tracing
    ctx := context.Background()
    gpu := &gpu.GPU{ID: "gpu-001", Name: "NVIDIA A100"}
    tracedScheduler.RegisterGPU(ctx, gpu)
}
```

## Component Tracing

### GPU Scheduling Operations

The tracing integration automatically captures:

- **GPU Registration**: GPU details, memory capacity, availability
- **Workload Submission**: Workload IDs, memory requirements, priorities
- **Scheduling Decisions**: Queue sizes, scheduling duration, success/failure
- **Utilization Queries**: Current utilization percentages

```go
// All operations are automatically traced when using TracedGPUScheduler
tracedScheduler.RegisterGPU(ctx, gpu)
tracedScheduler.SubmitWorkload(ctx, workload)
tracedScheduler.Schedule(ctx)
utilization, _ := tracedScheduler.GetGPUUtilization(ctx)
```

### Metrics Collection

Traces capture metrics collection operations:

- **Collection Cycles**: Start/stop of collection processes
- **Individual GPU Metrics**: Per-GPU utilization, temperature, memory usage
- **Collection Performance**: Duration and success rates

```go
tracedCollector.Start(ctx)
metrics, _ := tracedCollector.CollectMetrics(ctx)
latest := tracedCollector.GetLatestMetrics(ctx)
```

### Cost Calculation

Financial operations are fully traced:

- **Cost Recording**: Operations, GPU hours, token usage, amounts
- **Cost Summaries**: Time ranges, total costs, GPU hour calculations

```go
entry := observability.CostEntry{
    Operation:  "LLM Training",
    GPUHours:   2.5,
    TokensUsed: 1500000,
    Cost:       12.50,
}
tracedMonitoring.RecordCost(ctx, entry)
```

### API Requests

HTTP requests and WebSocket events can be traced:

```go
// Trace API requests
ctx, span := tracingService.TraceAPIRequest(ctx, "gpu_status", "/api/gpu/status")
defer span.End()

// Trace WebSocket events  
ctx, span = tracingService.TraceWebSocketEvent(ctx, "metrics_update")
defer span.End()
```

## Span Attributes

AgentaFlow automatically adds rich metadata to traces:

### GPU Operations

- `gpu.id`: GPU identifier
- `gpu.name`: GPU model name
- `gpu.memory_total`: Total GPU memory in MB
- `gpu.memory_used`: Current memory usage in MB
- `gpu.utilization`: GPU utilization percentage
- `gpu.temperature`: GPU temperature in Celsius

### Workload Operations

- `workload.id`: Unique workload identifier
- `workload.name`: Human-readable workload name
- `workload.priority`: Scheduling priority
- `workload.memory_required`: Required GPU memory in MB

### Scheduler Operations

- `scheduler.queue_size`: Number of pending workloads
- `scheduler.available_gpus`: Number of available GPUs
- `scheduler.scheduled_count`: Number of workloads scheduled

### Cost Operations

- `cost.operation`: Type of operation (training, inference, etc.)
- `cost.gpu_hours`: GPU hours consumed
- `cost.tokens_used`: Number of tokens processed
- `cost.amount`: Cost in dollars

### Performance Metrics

- `operation.duration_ms`: Operation duration in milliseconds
- `operation.start_time`: Operation start timestamp
- `operation.status`: Success or error status

## Exporters

### Jaeger (Recommended for Development)

Perfect for local development and debugging:

```yaml
exporters:
  jaeger:
    enabled: true
    endpoint: "http://localhost:14268/api/traces"
```

Access the Jaeger UI at `http://localhost:16686` to view traces.

### OTLP (Recommended for Production)

For production environments with observability platforms:

```yaml
exporters:
  otlp_http:
    enabled: true
    endpoint: "https://api.honeycomb.io/v1/traces"
    headers:
      "x-honeycomb-team": "your-api-key"
```

Supports platforms like:

- Honeycomb
- New Relic
- Datadog
- AWS X-Ray
- Google Cloud Trace

### Stdout (Debugging)

For debugging and development:

```yaml
exporters:
  stdout:
    enabled: true
    pretty_print: true
```

## Configuration Options

### Sampling

Control trace volume with sampling ratios:

```yaml
tracing:
  sampling_ratio: 0.1  # Trace 10% of requests
```

### Component-Specific Settings

Fine-tune tracing per component:

```yaml
components:
  gpu_scheduler:
    trace_workload_submission: true
    trace_scheduling_decisions: true
    
  metrics_collector:
    trace_collection_cycles: true
    trace_individual_gpus: false  # Reduce noise
    
  web_dashboard:
    trace_api_requests: true
    trace_websocket_events: false  # Can be very noisy
```

### Performance Tuning

Optimize tracing performance:

```yaml
performance:
  max_export_batch_size: 512
  max_export_timeout: "30s"
  export_timeout: "30s"
  max_queue_size: 2048
```

## Running the Demo

A comprehensive tracing demo is available:

```bash
# Ensure Jaeger is running
docker run -d --name jaeger -p 16686:16686 -p 14268:14268 jaegertracing/all-in-one:latest

# Run the tracing demo
cd examples
go run tracing_demo.go
```

The demo demonstrates:

- GPU registration and scheduling with tracing
- Metrics collection tracing
- Cost calculation tracing  
- API request tracing
- Nested span correlation
- Health check tracing

## Best Practices

### 1. Use Meaningful Operation Names

```go
// Good
ctx, span := tracingService.TraceGPUScheduling(ctx, "schedule_llm_training", gpu.ID)

// Avoid
ctx, span := tracingService.TraceGPUScheduling(ctx, "operation", "")
```

### 2. Add Context-Specific Attributes

```go
tracingService.AddSpanAttributes(span,
    attribute.String("model.name", "gpt-3.5-turbo"),
    attribute.Int("batch.size", 32),
    attribute.String("user.id", userID),
)
```

### 3. Record Important Events

```go
tracingService.AddSpanEvent(span, "model.loaded",
    attribute.String("model.path", modelPath),
    attribute.Int64("model.size_mb", modelSizeMB),
)
```

### 4. Handle Errors Properly

```go
if err != nil {
    tracingService.RecordError(span, err)
    tracingService.SetSpanStatus(span, codes.Error, "Failed to load model")
}
```

### 5. Use Appropriate Sampling

- Development: `sampling_ratio: 1.0` (trace everything)
- Production: `sampling_ratio: 0.1` (trace 10%)
- High-traffic: `sampling_ratio: 0.01` (trace 1%)

## Integration with Monitoring

Tracing works seamlessly with existing monitoring:

- **Prometheus Metrics**: Correlate metrics with trace spans
- **Grafana Dashboards**: Link to traces from dashboard panels
- **Alerting**: Include trace IDs in alert notifications

## Troubleshooting

### Traces Not Appearing

1. **Check Exporter Configuration**:
   ```bash
   # Verify Jaeger is accessible
   curl http://localhost:14268/api/traces
   ```

2. **Verify Sampling**:
   ```yaml
   tracing:
     sampling_ratio: 1.0  # Ensure you're sampling traces
   ```

3. **Check Logs**:
   ```go
   config.EnableStdout = true  // Enable stdout exporter to see traces in logs
   ```

### High Trace Volume

1. **Reduce Sampling**:
   ```yaml
   tracing:
     sampling_ratio: 0.1  # Sample 10% instead of 100%
   ```

2. **Disable Noisy Components**:
   ```yaml
   components:
     web_dashboard:
       trace_websocket_events: false
   ```

### Performance Impact

1. **Async Export**: Traces are exported asynchronously with minimal performance impact
2. **Batch Export**: Spans are batched for efficient export
3. **Sampling**: Use appropriate sampling ratios for your environment

## API Reference

### TracingService

Core tracing service with span creation methods:

- `TraceGPUScheduling(ctx, operation, gpuID)`: GPU operations
- `TraceModelServing(ctx, operation, modelName)`: Model serving
- `TraceMetricsCollection(ctx, operation, gpuCount)`: Metrics collection
- `TraceAPIRequest(ctx, operation, endpoint)`: HTTP requests
- `TraceCostCalculation(ctx, operation, gpuHours)`: Cost operations

### TracedGPUScheduler

GPU scheduler with automatic tracing:

- `RegisterGPU(ctx, gpu)`: Register GPU with tracing
- `SubmitWorkload(ctx, workload)`: Submit workload with tracing
- `Schedule(ctx)`: Execute scheduling with tracing
- `GetGPUUtilization(ctx)`: Get utilization with tracing

### TracingIntegration

Main integration manager:

- `WrapGPUScheduler(scheduler)`: Add tracing to scheduler
- `WrapMetricsCollector(collector)`: Add tracing to collector
- `WrapMonitoringService(monitoring)`: Add tracing to monitoring
- `HealthCheck()`: Get tracing system health

## Conclusion

OpenTelemetry tracing in AgentaFlow SRO provides comprehensive observability into your AI infrastructure operations. From GPU scheduling decisions to cost calculations, every operation can be traced, correlated, and analyzed to optimize performance and troubleshoot issues.

The integration is designed to be:

- **Zero-config for common cases**: Works out of the box with Jaeger
- **Highly configurable**: Extensive configuration options for production
- **Performance-conscious**: Minimal overhead with async export
- **Standards-compliant**: Full OpenTelemetry compatibility

Start with the demo, configure for your environment, and gain deep insights into your AI infrastructure operations!
 
 