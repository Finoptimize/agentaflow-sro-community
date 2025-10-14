package observability

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/Finoptimize/agentaflow-sro-community/pkg/gpu"
)

// TracedGPUScheduler wraps a GPU scheduler with OpenTelemetry tracing
type TracedGPUScheduler struct {
	scheduler *gpu.Scheduler
	tracer    *TracingService
}

// NewTracedGPUScheduler creates a new traced GPU scheduler
func NewTracedGPUScheduler(scheduler *gpu.Scheduler, tracer *TracingService) *TracedGPUScheduler {
	return &TracedGPUScheduler{
		scheduler: scheduler,
		tracer:    tracer,
	}
}

// RegisterGPU registers a GPU with tracing
func (tgs *TracedGPUScheduler) RegisterGPU(ctx context.Context, gpu *gpu.GPU) error {
	ctx, span := tgs.tracer.TraceGPUScheduling(ctx, "register_gpu", gpu.ID)
	defer span.End()

	tgs.tracer.AddSpanAttributes(span,
		attribute.String("gpu.name", gpu.Name),
		attribute.Int64("gpu.memory_total", int64(gpu.MemoryTotal)),
		attribute.Bool("gpu.available", gpu.Available),
	)

	start := time.Now()
	err := tgs.scheduler.RegisterGPU(gpu)
	duration := time.Since(start)

	tgs.tracer.AddSpanAttributes(span,
		attribute.Int64("operation.duration_ms", duration.Milliseconds()),
	)

	if err != nil {
		tgs.tracer.RecordError(span, err)
	} else {
		tgs.tracer.SetSpanStatus(span, codes.Ok, "GPU registered successfully")
		tgs.tracer.AddSpanEvent(span, "gpu.registered",
			attribute.String("gpu.id", gpu.ID),
			attribute.String("gpu.name", gpu.Name),
		)
	}

	return err
}

// SubmitWorkload submits a workload with tracing
func (tgs *TracedGPUScheduler) SubmitWorkload(ctx context.Context, workload *gpu.Workload) error {
	ctx, span := tgs.tracer.TraceGPUScheduling(ctx, "submit_workload", "")
	defer span.End()

	tgs.tracer.AddSpanAttributes(span,
		attribute.String("workload.id", workload.ID),
		attribute.String("workload.name", workload.Name),
		attribute.Int("workload.priority", workload.Priority),
		attribute.Int64("workload.memory_required", int64(workload.MemoryRequired)),
	)

	start := time.Now()
	err := tgs.scheduler.SubmitWorkload(workload)
	duration := time.Since(start)

	tgs.tracer.AddSpanAttributes(span,
		attribute.Int64("operation.duration_ms", duration.Milliseconds()),
	)

	if err != nil {
		tgs.tracer.RecordError(span, err)
	} else {
		tgs.tracer.SetSpanStatus(span, codes.Ok, "Workload submitted successfully")
		tgs.tracer.AddSpanEvent(span, "workload.submitted",
			attribute.String("workload.id", workload.ID),
		)
	}

	return err
}

// Schedule performs scheduling with tracing
func (tgs *TracedGPUScheduler) Schedule(ctx context.Context) error {
	ctx, span := tgs.tracer.TraceGPUScheduling(ctx, "schedule", "")
	defer span.End()

	// Get queue information before scheduling
	metrics := tgs.scheduler.GetUtilizationMetrics()
	queueSize := metrics["pending_workloads"].(int)
	totalGPUs := metrics["total_gpus"].(int)

	tgs.tracer.AddSpanAttributes(span,
		attribute.Int("scheduler.queue_size", queueSize),
		attribute.Int("scheduler.total_gpus", totalGPUs),
	)

	start := time.Now()
	err := tgs.scheduler.Schedule()
	duration := time.Since(start)

	// Get scheduling results
	metricsAfter := tgs.scheduler.GetUtilizationMetrics()
	queueSizeAfter := metricsAfter["pending_workloads"].(int)
	scheduledCount := queueSize - queueSizeAfter

	tgs.tracer.AddSpanAttributes(span,
		attribute.Int64("operation.duration_ms", duration.Milliseconds()),
		attribute.Int("scheduler.scheduled_count", scheduledCount),
	)

	if err != nil {
		tgs.tracer.RecordError(span, err)
	} else {
		tgs.tracer.SetSpanStatus(span, codes.Ok, fmt.Sprintf("Scheduled %d workloads", scheduledCount))
		tgs.tracer.AddSpanEvent(span, "scheduling.completed",
			attribute.Int("workloads.scheduled", scheduledCount),
		)
	}

	return err
}

// GetGPUUtilization gets GPU utilization with tracing
func (tgs *TracedGPUScheduler) GetGPUUtilization(ctx context.Context) (float64, error) {
	ctx, span := tgs.tracer.TraceGPUScheduling(ctx, "get_utilization", "")
	defer span.End()

	start := time.Now()
	metrics := tgs.scheduler.GetUtilizationMetrics()
	val, ok := metrics["average_utilization"]
	if !ok {
		return 0, fmt.Errorf("average_utilization key not found in metrics")
	}
	utilization, ok := val.(float64)
	if !ok {
		return 0, fmt.Errorf("average_utilization value is not a float64")
	}
	duration := time.Since(start)

	tgs.tracer.AddSpanAttributes(span,
		attribute.Int64("operation.duration_ms", duration.Milliseconds()),
		attribute.Float64("gpu.utilization", utilization),
	)

	tgs.tracer.SetSpanStatus(span, codes.Ok, "Utilization retrieved")

	return utilization, nil
}

// TracedMetricsCollector wraps a metrics collector with OpenTelemetry tracing
type TracedMetricsCollector struct {
	collector *gpu.MetricsCollector
	tracer    *TracingService
}

// NewTracedMetricsCollector creates a new traced metrics collector
func NewTracedMetricsCollector(collector *gpu.MetricsCollector, tracer *TracingService) *TracedMetricsCollector {
	return &TracedMetricsCollector{
		collector: collector,
		tracer:    tracer,
	}
}

// Start starts the metrics collector with tracing
func (tmc *TracedMetricsCollector) Start(ctx context.Context) error {
	ctx, span := tmc.tracer.TraceMetricsCollection(ctx, "start_collector", 0)
	defer span.End()

	err := tmc.collector.Start()

	if err != nil {
		tmc.tracer.RecordError(span, err)
	} else {
		tmc.tracer.SetSpanStatus(span, codes.Ok, "Metrics collector started")
		tmc.tracer.AddSpanEvent(span, "collector.started")
	}

	return err
}

// CollectMetrics collects metrics with tracing
func (tmc *TracedMetricsCollector) CollectMetrics(ctx context.Context) (*gpu.GPUMetrics, error) {
	ctx, span := tmc.tracer.TraceMetricsCollection(ctx, "collect_metrics", 1)
	defer span.End()

	start := time.Now()
	metrics, err := tmc.collector.CollectMetrics()
	duration := time.Since(start)

	tmc.tracer.AddSpanAttributes(span,
		attribute.Int64("operation.duration_ms", duration.Milliseconds()),
	)

	if err != nil {
		tmc.tracer.RecordError(span, err)
	} else {
		tmc.tracer.SetSpanStatus(span, codes.Ok, "Metrics collected")
		tmc.tracer.AddSpanAttributes(span,
			attribute.String("gpu.id", metrics.GPUID),
			attribute.Float64("gpu.utilization", metrics.UtilizationGPU),
			attribute.Float64("gpu.temperature", metrics.Temperature),
			attribute.Int64("gpu.memory_used", int64(metrics.MemoryUsed)),
		)
		tmc.tracer.AddSpanEvent(span, "metrics.collected",
			attribute.String("gpu.id", metrics.GPUID),
		)
	}

	return metrics, err
}

// GetLatestMetrics gets latest metrics with tracing
func (tmc *TracedMetricsCollector) GetLatestMetrics(ctx context.Context) map[string]gpu.GPUMetrics {
	ctx, span := tmc.tracer.TraceMetricsCollection(ctx, "get_latest_metrics", 0)
	defer span.End()

	start := time.Now()
	metrics := tmc.collector.GetLatestMetrics()
	duration := time.Since(start)

	tmc.tracer.AddSpanAttributes(span,
		attribute.Int64("operation.duration_ms", duration.Milliseconds()),
		attribute.Int("metrics.gpu_count", len(metrics)),
	)

	tmc.tracer.SetSpanStatus(span, codes.Ok, fmt.Sprintf("Retrieved metrics for %d GPUs", len(metrics)))

	return metrics
}

// TracedMonitoringService wraps a monitoring service with OpenTelemetry tracing
type TracedMonitoringService struct {
	monitoring *MonitoringService
	tracer     *TracingService
}

// NewTracedMonitoringService creates a new traced monitoring service
func NewTracedMonitoringService(monitoring *MonitoringService, tracer *TracingService) *TracedMonitoringService {
	return &TracedMonitoringService{
		monitoring: monitoring,
		tracer:     tracer,
	}
}

// RecordCost records a cost entry with tracing
func (tms *TracedMonitoringService) RecordCost(ctx context.Context, entry CostEntry) {
	ctx, span := tms.tracer.TraceCostCalculation(ctx, "record_cost", entry.GPUHours)
	defer span.End()

	tms.tracer.AddSpanAttributes(span,
		attribute.String("cost.operation", entry.Operation),
		attribute.Float64("cost.gpu_hours", entry.GPUHours),
		attribute.Int64("cost.tokens_used", entry.TokensUsed),
		attribute.Float64("cost.amount", entry.Cost),
	)

	start := time.Now()
	tms.monitoring.RecordCost(entry)
	duration := time.Since(start)

	tms.tracer.AddSpanAttributes(span,
		attribute.Int64("operation.duration_ms", duration.Milliseconds()),
	)

	tms.tracer.SetSpanStatus(span, codes.Ok, "Cost recorded")
	tms.tracer.AddSpanEvent(span, "cost.recorded",
		attribute.String("operation", entry.Operation),
		attribute.Float64("cost", entry.Cost),
	)
}

// GetCostSummary gets cost summary with tracing
func (tms *TracedMonitoringService) GetCostSummary(ctx context.Context, startTime, endTime time.Time) map[string]interface{} {
	ctx, span := tms.tracer.TraceCostCalculation(ctx, "get_cost_summary", 0)
	defer span.End()

	tms.tracer.AddSpanAttributes(span,
		attribute.String("cost.start_time", startTime.Format(time.RFC3339)),
		attribute.String("cost.end_time", endTime.Format(time.RFC3339)),
	)

	start := time.Now()
	summary := tms.monitoring.GetCostSummary(startTime, endTime)
	duration := time.Since(start)

	totalCost, _ := summary["total_cost"].(float64)
	gpuHours, _ := summary["gpu_hours"].(float64)

	tms.tracer.AddSpanAttributes(span,
		attribute.Int64("operation.duration_ms", duration.Milliseconds()),
		attribute.Float64("cost.total", totalCost),
		attribute.Float64("cost.gpu_hours", gpuHours),
	)

	tms.tracer.SetSpanStatus(span, codes.Ok, "Cost summary retrieved")

	return summary
}

// TracingIntegration manages tracing integration across all AgentaFlow components
type TracingIntegration struct {
	tracingService *TracingService
	scheduler      *TracedGPUScheduler
	collector      *TracedMetricsCollector
	monitoring     *TracedMonitoringService
}

// NewTracingIntegration creates a new tracing integration
func NewTracingIntegration(config *TracingConfig) (*TracingIntegration, error) {
	tracingService, err := NewTracingService(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create tracing service: %w", err)
	}

	return &TracingIntegration{
		tracingService: tracingService,
	}, nil
}

// WrapGPUScheduler wraps a GPU scheduler with tracing
func (ti *TracingIntegration) WrapGPUScheduler(scheduler *gpu.Scheduler) *TracedGPUScheduler {
	ti.scheduler = NewTracedGPUScheduler(scheduler, ti.tracingService)
	return ti.scheduler
}

// WrapMetricsCollector wraps a metrics collector with tracing
func (ti *TracingIntegration) WrapMetricsCollector(collector *gpu.MetricsCollector) *TracedMetricsCollector {
	ti.collector = NewTracedMetricsCollector(collector, ti.tracingService)
	return ti.collector
}

// WrapMonitoringService wraps a monitoring service with tracing
func (ti *TracingIntegration) WrapMonitoringService(monitoring *MonitoringService) *TracedMonitoringService {
	ti.monitoring = NewTracedMonitoringService(monitoring, ti.tracingService)
	return ti.monitoring
}

// GetTracingService returns the underlying tracing service
func (ti *TracingIntegration) GetTracingService() *TracingService {
	return ti.tracingService
}

// Shutdown gracefully shuts down the tracing integration
func (ti *TracingIntegration) Shutdown(ctx context.Context) error {
	return ti.tracingService.Shutdown(ctx)
}

// HealthCheck returns health information for the tracing integration
func (ti *TracingIntegration) HealthCheck() map[string]interface{} {
	return ti.tracingService.TracingHealthCheck()
}
