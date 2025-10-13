package observability

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// PrometheusExporter exports metrics in Prometheus format
type PrometheusExporter struct {
	monitoringService *MonitoringService
	mu                sync.RWMutex

	// Metric registries
	gaugeMetrics     map[string]float64
	counterMetrics   map[string]float64
	histogramMetrics map[string][]float64

	// Metric metadata
	metricHelp   map[string]string
	metricTypes  map[string]string
	metricLabels map[string]map[string]string

	// Configuration
	metricsPrefix  string
	enabledMetrics map[string]bool
}

// PrometheusConfig configures the Prometheus exporter
type PrometheusConfig struct {
	MetricsPrefix  string            `json:"metrics_prefix"`
	EnabledMetrics map[string]bool   `json:"enabled_metrics"`
	MetricLabels   map[string]string `json:"metric_labels"`
}

// DefaultPrometheusConfig returns default Prometheus configuration
func DefaultPrometheusConfig() PrometheusConfig {
	return PrometheusConfig{
		MetricsPrefix: "agentaflow",
		EnabledMetrics: map[string]bool{
			"gpu_metrics":        true,
			"scheduling_metrics": true,
			"serving_metrics":    true,
			"cost_metrics":       true,
			"system_metrics":     true,
		},
		MetricLabels: map[string]string{
			"instance": "agentaflow",
			"version":  "community",
		},
	}
}

// NewPrometheusExporter creates a new Prometheus metrics exporter
func NewPrometheusExporter(monitoringService *MonitoringService, config PrometheusConfig) *PrometheusExporter {
	return &PrometheusExporter{
		monitoringService: monitoringService,
		gaugeMetrics:      make(map[string]float64),
		counterMetrics:    make(map[string]float64),
		histogramMetrics:  make(map[string][]float64),
		metricHelp:        make(map[string]string),
		metricTypes:       make(map[string]string),
		metricLabels:      make(map[string]map[string]string),
		metricsPrefix:     config.MetricsPrefix,
		enabledMetrics:    config.EnabledMetrics,
	}
}

// RegisterGPUMetrics registers GPU-related metrics
func (pe *PrometheusExporter) RegisterGPUMetrics() {
	if !pe.enabledMetrics["gpu_metrics"] {
		return
	}

	// GPU utilization metrics
	pe.registerMetric("gpu_utilization_percent", "gauge",
		"GPU utilization percentage", []string{"gpu_id", "gpu_name", "node"})
	pe.registerMetric("gpu_memory_utilization_percent", "gauge",
		"GPU memory utilization percentage", []string{"gpu_id", "gpu_name", "node"})
	pe.registerMetric("gpu_memory_used_bytes", "gauge",
		"GPU memory used in bytes", []string{"gpu_id", "gpu_name", "node"})
	pe.registerMetric("gpu_memory_total_bytes", "gauge",
		"GPU memory total in bytes", []string{"gpu_id", "gpu_name", "node"})

	// GPU temperature and power metrics
	pe.registerMetric("gpu_temperature_celsius", "gauge",
		"GPU temperature in Celsius", []string{"gpu_id", "gpu_name", "node"})
	pe.registerMetric("gpu_power_draw_watts", "gauge",
		"GPU power draw in watts", []string{"gpu_id", "gpu_name", "node"})
	pe.registerMetric("gpu_power_limit_watts", "gauge",
		"GPU power limit in watts", []string{"gpu_id", "gpu_name", "node"})
	pe.registerMetric("gpu_fan_speed_percent", "gauge",
		"GPU fan speed percentage", []string{"gpu_id", "gpu_name", "node"})

	// GPU clock metrics
	pe.registerMetric("gpu_clock_graphics_mhz", "gauge",
		"GPU graphics clock in MHz", []string{"gpu_id", "gpu_name", "node"})
	pe.registerMetric("gpu_clock_memory_mhz", "gauge",
		"GPU memory clock in MHz", []string{"gpu_id", "gpu_name", "node"})

	// GPU efficiency and process metrics
	pe.registerMetric("gpu_process_count", "gauge",
		"Number of processes running on GPU", []string{"gpu_id", "gpu_name", "node"})
	pe.registerMetric("gpu_efficiency_score", "gauge",
		"GPU efficiency score (0-1)", []string{"gpu_id", "gpu_name", "node"})
	pe.registerMetric("gpu_idle_time_percent", "gauge",
		"GPU idle time percentage", []string{"gpu_id", "gpu_name", "node"})

	// GPU health status
	pe.registerMetric("gpu_health_status", "gauge",
		"GPU health status (0=unhealthy, 1=warning, 2=healthy)", []string{"gpu_id", "gpu_name", "node", "status"})
}

// RegisterSchedulingMetrics registers GPU scheduling metrics
func (pe *PrometheusExporter) RegisterSchedulingMetrics() {
	if !pe.enabledMetrics["scheduling_metrics"] {
		return
	}

	// Workload metrics
	pe.registerMetric("workloads_total", "counter",
		"Total number of workloads submitted", []string{"status", "priority"})
	pe.registerMetric("workloads_pending", "gauge",
		"Number of pending workloads", []string{"priority"})
	pe.registerMetric("workloads_running", "gauge",
		"Number of running workloads", []string{"gpu_id", "priority"})
	pe.registerMetric("workloads_completed", "counter",
		"Number of completed workloads", []string{"status", "priority"})

	// Scheduling performance metrics
	pe.registerMetric("scheduling_duration_seconds", "histogram",
		"Time taken for scheduling decisions", []string{"strategy"})
	pe.registerMetric("scheduling_decisions_total", "counter",
		"Total scheduling decisions made", []string{"strategy", "outcome"})
	pe.registerMetric("gpu_allocation_efficiency", "gauge",
		"GPU allocation efficiency percentage", []string{"strategy"})

	// Queue metrics
	pe.registerMetric("workload_queue_time_seconds", "histogram",
		"Time workloads spend in queue", []string{"priority"})
	pe.registerMetric("workload_execution_time_seconds", "histogram",
		"Workload execution time", []string{"workload_type", "gpu_type"})
}

// RegisterServingMetrics registers model serving metrics
func (pe *PrometheusExporter) RegisterServingMetrics() {
	if !pe.enabledMetrics["serving_metrics"] {
		return
	}

	// Request metrics
	pe.registerMetric("inference_requests_total", "counter",
		"Total inference requests", []string{"model_id", "status"})
	pe.registerMetric("inference_latency_seconds", "histogram",
		"Inference request latency", []string{"model_id"})
	pe.registerMetric("inference_batch_size", "histogram",
		"Inference batch sizes", []string{"model_id"})

	// Cache metrics
	pe.registerMetric("cache_hits_total", "counter",
		"Total cache hits", []string{"model_id"})
	pe.registerMetric("cache_misses_total", "counter",
		"Total cache misses", []string{"model_id"})
	pe.registerMetric("cache_hit_rate", "gauge",
		"Cache hit rate percentage", []string{"model_id"})
	pe.registerMetric("cache_size_bytes", "gauge",
		"Cache size in bytes", []string{"model_id"})

	// Throughput metrics
	pe.registerMetric("inference_throughput_requests_per_second", "gauge",
		"Inference throughput (requests/second)", []string{"model_id"})
	pe.registerMetric("tokens_processed_per_second", "gauge",
		"Tokens processed per second", []string{"model_id"})
}

// RegisterCostMetrics registers cost tracking metrics
func (pe *PrometheusExporter) RegisterCostMetrics() {
	if !pe.enabledMetrics["cost_metrics"] {
		return
	}

	// Cost metrics
	pe.registerMetric("cost_total_dollars", "counter",
		"Total costs in dollars", []string{"operation", "model_id", "currency"})
	pe.registerMetric("cost_per_hour_dollars", "gauge",
		"Current cost per hour in dollars", []string{"resource_type", "currency"})
	pe.registerMetric("gpu_hours_consumed", "counter",
		"Total GPU hours consumed", []string{"gpu_type", "workload_type"})
	pe.registerMetric("tokens_processed_total", "counter",
		"Total tokens processed", []string{"model_id", "operation"})

	// Cost efficiency metrics
	pe.registerMetric("cost_per_token_dollars", "gauge",
		"Cost per token in dollars", []string{"model_id", "operation"})
	pe.registerMetric("cost_efficiency_score", "gauge",
		"Cost efficiency score (output/cost)", []string{"model_id", "operation"})
	pe.registerMetric("estimated_monthly_cost_dollars", "gauge",
		"Estimated monthly cost in dollars", []string{"resource_type"})
}

// RegisterSystemMetrics registers system-level metrics
func (pe *PrometheusExporter) RegisterSystemMetrics() {
	if !pe.enabledMetrics["system_metrics"] {
		return
	}

	// System overview metrics
	pe.registerMetric("gpus_total", "gauge",
		"Total number of GPUs in cluster", []string{"node", "gpu_type"})
	pe.registerMetric("gpus_available", "gauge",
		"Number of available GPUs", []string{"node", "gpu_type"})
	pe.registerMetric("cluster_utilization_percent", "gauge",
		"Overall cluster utilization percentage", []string{})
	pe.registerMetric("cluster_efficiency_score", "gauge",
		"Overall cluster efficiency score", []string{})

	// Alert metrics
	pe.registerMetric("alerts_total", "counter",
		"Total number of alerts generated", []string{"severity", "type", "source"})
	pe.registerMetric("active_alerts", "gauge",
		"Number of currently active alerts", []string{"severity", "type"})

	// System health metrics
	pe.registerMetric("system_uptime_seconds", "gauge",
		"System uptime in seconds", []string{"component"})
	pe.registerMetric("component_health_status", "gauge",
		"Component health status (0=down, 1=degraded, 2=healthy)", []string{"component"})
}

// registerMetric registers a metric with metadata
func (pe *PrometheusExporter) registerMetric(name, metricType, help string, labels []string) {
	fullName := fmt.Sprintf("%s_%s", pe.metricsPrefix, name)
	pe.metricTypes[fullName] = metricType
	pe.metricHelp[fullName] = help
	pe.metricLabels[fullName] = make(map[string]string)

	// Initialize metric based on type
	switch metricType {
	case "gauge":
		pe.gaugeMetrics[fullName] = 0.0
	case "counter":
		pe.counterMetrics[fullName] = 0.0
	case "histogram":
		pe.histogramMetrics[fullName] = make([]float64, 0)
	}
}

// UpdateMetric updates a specific metric value
func (pe *PrometheusExporter) UpdateMetric(name string, value float64, labels map[string]string) {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	fullName := fmt.Sprintf("%s_%s", pe.metricsPrefix, name)
	metricKey := pe.buildMetricKey(fullName, labels)

	metricType := pe.metricTypes[fullName]
	switch metricType {
	case "gauge":
		pe.gaugeMetrics[metricKey] = value
	case "counter":
		pe.counterMetrics[metricKey] += value
	case "histogram":
		if pe.histogramMetrics[metricKey] == nil {
			pe.histogramMetrics[metricKey] = make([]float64, 0)
		}
		pe.histogramMetrics[metricKey] = append(pe.histogramMetrics[metricKey], value)
	}
}

// buildMetricKey creates a unique key for metric with labels
func (pe *PrometheusExporter) buildMetricKey(name string, labels map[string]string) string {
	if len(labels) == 0 {
		return name
	}

	var labelPairs []string
	for key, value := range labels {
		labelPairs = append(labelPairs, fmt.Sprintf("%s=%s", key, value))
	}

	return fmt.Sprintf("%s{%s}", name, strings.Join(labelPairs, ","))
}

// ExportMetrics exports metrics in Prometheus format
func (pe *PrometheusExporter) ExportMetrics() string {
	pe.mu.RLock()
	defer pe.mu.RUnlock()

	var output strings.Builder

	// Export gauge metrics
	for metricKey, value := range pe.gaugeMetrics {
		name, labels := pe.parseMetricKey(metricKey)
		if help, exists := pe.metricHelp[name]; exists {
			output.WriteString(fmt.Sprintf("# HELP %s %s\n", name, help))
			output.WriteString(fmt.Sprintf("# TYPE %s gauge\n", name))
		}

		if labels != "" {
			output.WriteString(fmt.Sprintf("%s{%s} %.2f\n", name, labels, value))
		} else {
			output.WriteString(fmt.Sprintf("%s %.2f\n", name, value))
		}
	}

	// Export counter metrics
	for metricKey, value := range pe.counterMetrics {
		name, labels := pe.parseMetricKey(metricKey)
		if help, exists := pe.metricHelp[name]; exists {
			output.WriteString(fmt.Sprintf("# HELP %s %s\n", name, help))
			output.WriteString(fmt.Sprintf("# TYPE %s counter\n", name))
		}

		if labels != "" {
			output.WriteString(fmt.Sprintf("%s{%s} %.2f\n", name, labels, value))
		} else {
			output.WriteString(fmt.Sprintf("%s %.2f\n", name, value))
		}
	}

	// Export histogram metrics (simplified - just count and sum)
	for metricKey, values := range pe.histogramMetrics {
		if len(values) == 0 {
			continue
		}

		name, labels := pe.parseMetricKey(metricKey)
		if help, exists := pe.metricHelp[name]; exists {
			output.WriteString(fmt.Sprintf("# HELP %s %s\n", name, help))
			output.WriteString(fmt.Sprintf("# TYPE %s histogram\n", name))
		}

		count := float64(len(values))
		sum := 0.0
		for _, v := range values {
			sum += v
		}

		labelStr := labels
		if labelStr != "" {
			output.WriteString(fmt.Sprintf("%s_count{%s} %.0f\n", name, labelStr, count))
			output.WriteString(fmt.Sprintf("%s_sum{%s} %.2f\n", name, labelStr, sum))
		} else {
			output.WriteString(fmt.Sprintf("%s_count %.0f\n", name, count))
			output.WriteString(fmt.Sprintf("%s_sum %.2f\n", name, sum))
		}
	}

	return output.String()
}

// parseMetricKey extracts metric name and labels from metric key
func (pe *PrometheusExporter) parseMetricKey(metricKey string) (string, string) {
	if !strings.Contains(metricKey, "{") {
		return metricKey, ""
	}

	parts := strings.SplitN(metricKey, "{", 2)
	name := parts[0]
	labels := strings.TrimSuffix(parts[1], "}")

	return name, labels
}

// ServeHTTP serves metrics via HTTP for Prometheus scraping
func (pe *PrometheusExporter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/metrics" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	// Add timestamp
	w.Write([]byte(fmt.Sprintf("# Generated at %s\n", time.Now().UTC().Format(time.RFC3339))))

	// Export metrics
	metrics := pe.ExportMetrics()
	w.Write([]byte(metrics))
}

// StartMetricsServer starts an HTTP server for Prometheus metrics
func (pe *PrometheusExporter) StartMetricsServer(port int) error {
	addr := fmt.Sprintf(":%d", port)

	http.Handle("/metrics", pe)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	return http.ListenAndServe(addr, nil)
}

// SyncFromMonitoringService syncs metrics from the monitoring service
func (pe *PrometheusExporter) SyncFromMonitoringService() {
	if pe.monitoringService == nil {
		return
	}

	// Get current metrics from monitoring service
	now := time.Now()
	oneHourAgo := now.Add(-1 * time.Hour)

	// Get recent metrics
	metrics := pe.monitoringService.GetMetrics(oneHourAgo, now, "")
	for _, metric := range metrics {
		pe.UpdateMetric(metric.Name, metric.Value, metric.Labels)
	}

	// Get cost summary
	costSummary := pe.monitoringService.GetCostSummary(oneHourAgo, now)
	if totalCost, ok := costSummary["total_cost"].(float64); ok {
		pe.UpdateMetric("cost_total_dollars", totalCost, map[string]string{
			"operation": "all",
			"currency":  "USD",
		})
	}

	if gpuHours, ok := costSummary["total_gpu_hours"].(float64); ok {
		pe.UpdateMetric("gpu_hours_consumed", gpuHours, map[string]string{
			"gpu_type":      "all",
			"workload_type": "all",
		})
	}
}
