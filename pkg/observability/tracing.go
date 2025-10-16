package observability

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// TracingConfig holds configuration for OpenTelemetry tracing
type TracingConfig struct {
	ServiceName    string            `yaml:"service_name" json:"service_name"`
	ServiceVersion string            `yaml:"service_version" json:"service_version"`
	ExporterType   string            `yaml:"exporter_type" json:"exporter_type"` // "jaeger", "otlp", "stdout", "none"
	JaegerEndpoint string            `yaml:"jaeger_endpoint" json:"jaeger_endpoint"`
	OTLPEndpoint   string            `yaml:"otlp_endpoint" json:"otlp_endpoint"`
	SampleRate     float64           `yaml:"sample_rate" json:"sample_rate"`
	Environment    string            `yaml:"environment" json:"environment"`
	Attributes     map[string]string `yaml:"attributes" json:"attributes"`
	EnabledSpans   map[string]bool   `yaml:"enabled_spans" json:"enabled_spans"`
}

// DefaultTracingConfig returns default tracing configuration
func DefaultTracingConfig() *TracingConfig {
	return &TracingConfig{
		ServiceName:    "agentaflow-sro",
		ServiceVersion: "1.0.0",
		ExporterType:   "stdout",
		JaegerEndpoint: "http://localhost:14268/api/traces",
		OTLPEndpoint:   "http://localhost:4318/v1/traces",
		SampleRate:     1.0, // 100% sampling for development
		Environment:    "development",
		Attributes: map[string]string{
			"deployment.environment": "development",
			"service.namespace":      "agentaflow",
		},
		EnabledSpans: map[string]bool{
			"gpu_scheduling":     true,
			"model_serving":      true,
			"metrics_collection": true,
			"cost_tracking":      true,
			"api_requests":       true,
			"websocket_events":   true,
		},
	}
}

// TracingService manages OpenTelemetry tracing infrastructure
type TracingService struct {
	config   *TracingConfig
	provider *trace.TracerProvider
	tracer   oteltrace.Tracer
	enabled  bool
	logger   *log.Logger
}

// NewTracingService creates a new tracing service with the given configuration
func NewTracingService(config *TracingConfig) (*TracingService, error) {
	if config == nil {
		config = DefaultTracingConfig()
	}

	ts := &TracingService{
		config:  config,
		enabled: config.ExporterType != "none",
		logger:  log.New(log.Writer(), "[TracingService] ", log.LstdFlags|log.Lshortfile),
	}

	if !ts.enabled {
		ts.logger.Printf("Tracing disabled (exporter_type: none)")
		return ts, nil
	}

	if err := ts.initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize tracing: %w", err)
	}

	ts.logger.Printf("Tracing initialized with %s exporter", config.ExporterType)
	return ts, nil
}

// initialize sets up the OpenTelemetry tracing infrastructure
func (ts *TracingService) initialize() error {
	// Create resource with service information
	res, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(ts.config.ServiceName),
			semconv.ServiceVersionKey.String(ts.config.ServiceVersion),
			semconv.DeploymentEnvironmentKey.String(ts.config.Environment),
		),
		resource.WithFromEnv(),
	)
	if err != nil {
		return fmt.Errorf("failed to create resource: %w", err)
	}

	// Add custom attributes
	attributes := make([]attribute.KeyValue, 0, len(ts.config.Attributes))
	for key, value := range ts.config.Attributes {
		attributes = append(attributes, attribute.String(key, value))
	}
	if len(attributes) > 0 {
		res, err = resource.Merge(res, resource.NewWithAttributes(
			semconv.SchemaURL,
			attributes...,
		))
		if err != nil {
			return fmt.Errorf("failed to merge resource attributes: %w", err)
		}
	}

	// Create exporter based on configuration
	var exporter trace.SpanExporter
	switch ts.config.ExporterType {
	case "jaeger":
		exporter, err = ts.createJaegerExporter()
	case "otlp":
		exporter, err = ts.createOTLPExporter()
	case "stdout":
		exporter, err = ts.createStdoutExporter()
	default:
		return fmt.Errorf("unsupported exporter type: %s", ts.config.ExporterType)
	}
	if err != nil {
		return fmt.Errorf("failed to create exporter: %w", err)
	}

	// Create trace provider with sampling
	sampler := trace.TraceIDRatioBased(ts.config.SampleRate)
	ts.provider = trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(res),
		trace.WithSampler(sampler),
	)

	// Set global trace provider and propagator
	otel.SetTracerProvider(ts.provider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// Create tracer for this service
	ts.tracer = otel.Tracer(ts.config.ServiceName)

	return nil
}

// createJaegerExporter creates a Jaeger exporter
func (ts *TracingService) createJaegerExporter() (trace.SpanExporter, error) {
	return jaeger.New(jaeger.WithCollectorEndpoint(
		jaeger.WithEndpoint(ts.config.JaegerEndpoint),
	))
}

// createOTLPExporter creates an OTLP HTTP exporter
func (ts *TracingService) createOTLPExporter() (trace.SpanExporter, error) {
	return otlptrace.New(
		context.Background(),
		otlptracehttp.NewClient(
			otlptracehttp.WithEndpoint(ts.config.OTLPEndpoint),
			otlptracehttp.WithInsecure(),
		),
	)
}

// createStdoutExporter creates a stdout exporter for development
func (ts *TracingService) createStdoutExporter() (trace.SpanExporter, error) {
	return stdouttrace.New(
		stdouttrace.WithPrettyPrint(),
	)
}

// StartSpan starts a new trace span with the given name and options
func (ts *TracingService) StartSpan(ctx context.Context, spanName string, opts ...oteltrace.SpanStartOption) (context.Context, oteltrace.Span) {
	if !ts.enabled || !ts.isSpanEnabled(spanName) {
		return ctx, oteltrace.SpanFromContext(ctx)
	}

	return ts.tracer.Start(ctx, spanName, opts...)
}

// isSpanEnabled checks if a span category is enabled in configuration
func (ts *TracingService) isSpanEnabled(spanName string) bool {
	// Extract category from span name (e.g., "gpu_scheduling.schedule_workload" -> "gpu_scheduling")
	for category, enabled := range ts.config.EnabledSpans {
		if len(spanName) >= len(category) && spanName[:len(category)] == category {
			return enabled
		}
	}
	return true // Default to enabled if not specifically configured
}

// AddSpanAttributes adds attributes to the current span
func (ts *TracingService) AddSpanAttributes(span oteltrace.Span, attrs ...attribute.KeyValue) {
	if !ts.enabled || span == nil {
		return
	}
	span.SetAttributes(attrs...)
}

// AddSpanEvent adds an event to the current span
func (ts *TracingService) AddSpanEvent(span oteltrace.Span, name string, attrs ...attribute.KeyValue) {
	if !ts.enabled || span == nil {
		return
	}
	span.AddEvent(name, oteltrace.WithAttributes(attrs...))
}

// RecordError records an error in the current span
func (ts *TracingService) RecordError(span oteltrace.Span, err error) {
	if !ts.enabled || span == nil || err == nil {
		return
	}
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
}

// SetSpanStatus sets the status of the current span
func (ts *TracingService) SetSpanStatus(span oteltrace.Span, code codes.Code, description string) {
	if !ts.enabled || span == nil {
		return
	}
	span.SetStatus(code, description)
}

// TraceGPUScheduling creates a span for GPU scheduling operations
func (ts *TracingService) TraceGPUScheduling(ctx context.Context, operation string, gpuID string) (context.Context, oteltrace.Span) {
	spanName := fmt.Sprintf("gpu_scheduling.%s", operation)
	ctx, span := ts.StartSpan(ctx, spanName,
		oteltrace.WithSpanKind(oteltrace.SpanKindInternal),
		oteltrace.WithAttributes(
			attribute.String("gpu.id", gpuID),
			attribute.String("operation", operation),
		),
	)
	return ctx, span
}

// TraceModelServing creates a span for model serving operations
func (ts *TracingService) TraceModelServing(ctx context.Context, operation string, modelID string) (context.Context, oteltrace.Span) {
	spanName := fmt.Sprintf("model_serving.%s", operation)
	ctx, span := ts.StartSpan(ctx, spanName,
		oteltrace.WithSpanKind(oteltrace.SpanKindServer),
		oteltrace.WithAttributes(
			attribute.String("model.id", modelID),
			attribute.String("operation", operation),
		),
	)
	return ctx, span
}

// TraceMetricsCollection creates a span for metrics collection operations
func (ts *TracingService) TraceMetricsCollection(ctx context.Context, operation string, metricsCount int) (context.Context, oteltrace.Span) {
	spanName := fmt.Sprintf("metrics_collection.%s", operation)
	ctx, span := ts.StartSpan(ctx, spanName,
		oteltrace.WithSpanKind(oteltrace.SpanKindInternal),
		oteltrace.WithAttributes(
			attribute.String("operation", operation),
			attribute.Int("metrics.count", metricsCount),
		),
	)
	return ctx, span
}

// TraceAPIRequest creates a span for HTTP API requests
func (ts *TracingService) TraceAPIRequest(ctx context.Context, method string, path string) (context.Context, oteltrace.Span) {
	spanName := fmt.Sprintf("api_requests.%s %s", method, path)
	ctx, span := ts.StartSpan(ctx, spanName,
		oteltrace.WithSpanKind(oteltrace.SpanKindServer),
		oteltrace.WithAttributes(
			attribute.String("http.method", method),
			attribute.String("http.route", path),
		),
	)
	return ctx, span
}

// TraceWebSocketEvent creates a span for WebSocket events
func (ts *TracingService) TraceWebSocketEvent(ctx context.Context, event string, connectionID string) (context.Context, oteltrace.Span) {
	spanName := fmt.Sprintf("websocket_events.%s", event)
	ctx, span := ts.StartSpan(ctx, spanName,
		oteltrace.WithSpanKind(oteltrace.SpanKindServer),
		oteltrace.WithAttributes(
			attribute.String("websocket.event", event),
			attribute.String("connection.id", connectionID),
		),
	)
	return ctx, span
}

// TraceCostCalculation creates a span for cost calculation operations
func (ts *TracingService) TraceCostCalculation(ctx context.Context, operation string, gpuHours float64) (context.Context, oteltrace.Span) {
	spanName := fmt.Sprintf("cost_tracking.%s", operation)
	ctx, span := ts.StartSpan(ctx, spanName,
		oteltrace.WithSpanKind(oteltrace.SpanKindInternal),
		oteltrace.WithAttributes(
			attribute.String("operation", operation),
			attribute.Float64("cost.gpu_hours", gpuHours),
		),
	)
	return ctx, span
}

// GetTracer returns the underlying OpenTelemetry tracer
func (ts *TracingService) GetTracer() oteltrace.Tracer {
	return ts.tracer
}

// IsEnabled returns whether tracing is enabled
func (ts *TracingService) IsEnabled() bool {
	return ts.enabled
}

// Shutdown gracefully shuts down the tracing service
func (ts *TracingService) Shutdown(ctx context.Context) error {
	if !ts.enabled || ts.provider == nil {
		return nil
	}

	ts.logger.Printf("Shutting down tracing service")
	return ts.provider.Shutdown(ctx)
}

// TraceMiddleware returns an HTTP middleware that adds tracing to requests
func (ts *TracingService) TraceMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !ts.enabled {
				next.ServeHTTP(w, r)
				return
			}

			// Extract trace context from headers
			ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))

			// Start span for this request
			ctx, span := ts.TraceAPIRequest(ctx, r.Method, r.URL.Path)
			defer span.End()

			// Add request attributes
			ts.AddSpanAttributes(span,
				attribute.String("http.url", r.URL.String()),
				attribute.String("http.user_agent", r.UserAgent()),
				attribute.String("http.remote_addr", r.RemoteAddr),
			)

			// Create response writer wrapper to capture status
			rw := &responseWriter{ResponseWriter: w, statusCode: 200}

			// Continue with request processing
			next.ServeHTTP(rw, r.WithContext(ctx))

			// Add response attributes
			ts.AddSpanAttributes(span,
				attribute.Int("http.status_code", rw.statusCode),
			)

			// Set span status based on HTTP status code
			if rw.statusCode >= 400 {
				ts.SetSpanStatus(span, codes.Error, fmt.Sprintf("HTTP %d", rw.statusCode))
			} else {
				ts.SetSpanStatus(span, codes.Ok, "")
			}
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// TracingHealthCheck returns tracing service health information
func (ts *TracingService) TracingHealthCheck() map[string]interface{} {
	health := map[string]interface{}{
		"enabled":       ts.enabled,
		"service_name":  ts.config.ServiceName,
		"exporter_type": ts.config.ExporterType,
		"sample_rate":   ts.config.SampleRate,
		"environment":   ts.config.Environment,
	}

	if ts.enabled {
		health["status"] = "active"
		health["enabled_spans"] = ts.config.EnabledSpans
	} else {
		health["status"] = "disabled"
	}

	return health
}

// Example helper functions for common tracing patterns

// TraceFunction is a helper to trace entire function calls
func (ts *TracingService) TraceFunction(ctx context.Context, functionName string, fn func(ctx context.Context) error) error {
	ctx, span := ts.StartSpan(ctx, functionName)
	defer span.End()

	start := time.Now()
	err := fn(ctx)
	duration := time.Since(start)

	ts.AddSpanAttributes(span,
		attribute.Int64("function.duration_ms", duration.Milliseconds()),
	)

	if err != nil {
		ts.RecordError(span, err)
	} else {
		ts.SetSpanStatus(span, codes.Ok, "")
	}

	return err
}

// TraceAsync is a helper to trace asynchronous operations
func (ts *TracingService) TraceAsync(ctx context.Context, spanName string, fn func(ctx context.Context)) {
	if !ts.enabled {
		go fn(ctx)
		return
	}

	// Create a new context for the goroutine to avoid context cancellation issues
	spanCtx := oteltrace.SpanContextFromContext(ctx)
	newCtx := oteltrace.ContextWithSpanContext(context.Background(), spanCtx)

	go func() {
		ctx, span := ts.StartSpan(newCtx, spanName)
		defer span.End()

		start := time.Now()
		fn(ctx)
		duration := time.Since(start)

		ts.AddSpanAttributes(span,
			attribute.Int64("async.duration_ms", duration.Milliseconds()),
		)
		ts.SetSpanStatus(span, codes.Ok, "")
	}()
}
