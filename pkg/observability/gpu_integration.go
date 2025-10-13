package observability

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Finoptimize/agentaflow-sro-community/pkg/gpu"
)

// GPUMetricsIntegration connects GPU metrics collection with observability monitoring
type GPUMetricsIntegration struct {
	monitoringService *MonitoringService
	metricsCollector  *gpu.MetricsCollector
	mu                sync.RWMutex

	// Configuration
	alertThresholds GPUAlertThresholds
	metricsEnabled  bool
	eventsEnabled   bool
	costsEnabled    bool

	// State tracking
	lastKnownState map[string]gpu.GPUMetrics
	alertHistory   map[string][]gpu.GPUAlert
}

// GPUAlertThresholds defines thresholds for GPU monitoring alerts
type GPUAlertThresholds struct {
	HighTemperature     float64 // Temperature in Celsius
	CriticalTemperature float64
	HighMemoryUsage     float64 // Memory usage percentage
	CriticalMemoryUsage float64
	HighPowerUsage      float64 // Power usage percentage of limit
	CriticalPowerUsage  float64
	LowUtilization      float64 // GPU utilization percentage
	HighUtilization     float64
}

// DefaultGPUAlertThresholds returns sensible default alert thresholds
func DefaultGPUAlertThresholds() GPUAlertThresholds {
	return GPUAlertThresholds{
		HighTemperature:     75.0,
		CriticalTemperature: 85.0,
		HighMemoryUsage:     80.0,
		CriticalMemoryUsage: 95.0,
		HighPowerUsage:      80.0,
		CriticalPowerUsage:  95.0,
		LowUtilization:      10.0,
		HighUtilization:     95.0,
	}
}

// NewGPUMetricsIntegration creates a new GPU metrics integration
func NewGPUMetricsIntegration(
	monitoringService *MonitoringService,
	metricsCollector *gpu.MetricsCollector,
) *GPUMetricsIntegration {
	integration := &GPUMetricsIntegration{
		monitoringService: monitoringService,
		metricsCollector:  metricsCollector,
		alertThresholds:   DefaultGPUAlertThresholds(),
		metricsEnabled:    true,
		eventsEnabled:     true,
		costsEnabled:      true,
		lastKnownState:    make(map[string]gpu.GPUMetrics),
		alertHistory:      make(map[string][]gpu.GPUAlert),
	}

	// Register callback with metrics collector
	metricsCollector.RegisterCallback(integration.processGPUMetrics)

	return integration
}

// SetAlertThresholds configures custom alert thresholds
func (gmi *GPUMetricsIntegration) SetAlertThresholds(thresholds GPUAlertThresholds) {
	gmi.mu.Lock()
	defer gmi.mu.Unlock()
	gmi.alertThresholds = thresholds
}

// EnableMetrics enables/disables metrics recording
func (gmi *GPUMetricsIntegration) EnableMetrics(enabled bool) {
	gmi.mu.Lock()
	defer gmi.mu.Unlock()
	gmi.metricsEnabled = enabled
}

// EnableEvents enables/disables event recording
func (gmi *GPUMetricsIntegration) EnableEvents(enabled bool) {
	gmi.mu.Lock()
	defer gmi.mu.Unlock()
	gmi.eventsEnabled = enabled
}

// EnableCostTracking enables/disables cost tracking
func (gmi *GPUMetricsIntegration) EnableCostTracking(enabled bool) {
	gmi.mu.Lock()
	defer gmi.mu.Unlock()
	gmi.costsEnabled = enabled
}

// processGPUMetrics processes incoming GPU metrics and integrates with monitoring
func (gmi *GPUMetricsIntegration) processGPUMetrics(metrics gpu.GPUMetrics) {
	gmi.mu.Lock()
	defer gmi.mu.Unlock()

	gpuID := metrics.GPUID
	lastState, hasLastState := gmi.lastKnownState[gpuID]

	// Record metrics if enabled
	if gmi.metricsEnabled {
		gmi.recordGPUMetrics(metrics)
	}

	// Check for alerts and record events if enabled
	if gmi.eventsEnabled {
		alerts := gmi.checkAlerts(metrics, lastState, hasLastState)
		for _, alert := range alerts {
			gmi.recordAlertEvent(alert, metrics)
		}

		// Store alerts in history
		if _, exists := gmi.alertHistory[gpuID]; !exists {
			gmi.alertHistory[gpuID] = make([]gpu.GPUAlert, 0)
		}
		gmi.alertHistory[gpuID] = append(gmi.alertHistory[gpuID], alerts...)

		// Keep only last 100 alerts per GPU
		if len(gmi.alertHistory[gpuID]) > 100 {
			gmi.alertHistory[gpuID] = gmi.alertHistory[gpuID][len(gmi.alertHistory[gpuID])-100:]
		}
	}

	// Record GPU costs if enabled
	if gmi.costsEnabled {
		gmi.recordGPUCosts(metrics, lastState, hasLastState)
	}

	// Update last known state
	gmi.lastKnownState[gpuID] = metrics
}

// recordGPUMetrics records GPU metrics with the monitoring service
func (gmi *GPUMetricsIntegration) recordGPUMetrics(metrics gpu.GPUMetrics) {
	labels := map[string]string{
		"gpu_id":   metrics.GPUID,
		"gpu_name": metrics.Name,
	}

	// GPU utilization metrics
	gmi.monitoringService.RecordMetric(Metric{
		Name:   "gpu_utilization_percent",
		Type:   MetricGauge,
		Value:  metrics.UtilizationGPU,
		Labels: labels,
	})

	// Memory metrics
	gmi.monitoringService.RecordMetric(Metric{
		Name:   "gpu_memory_utilization_percent",
		Type:   MetricGauge,
		Value:  metrics.UtilizationMemory,
		Labels: labels,
	})

	gmi.monitoringService.RecordMetric(Metric{
		Name:   "gpu_memory_used_mb",
		Type:   MetricGauge,
		Value:  float64(metrics.MemoryUsed),
		Labels: labels,
	})

	gmi.monitoringService.RecordMetric(Metric{
		Name:   "gpu_memory_total_mb",
		Type:   MetricGauge,
		Value:  float64(metrics.MemoryTotal),
		Labels: labels,
	})

	// Temperature metrics
	gmi.monitoringService.RecordMetric(Metric{
		Name:   "gpu_temperature_celsius",
		Type:   MetricGauge,
		Value:  metrics.Temperature,
		Labels: labels,
	})

	// Power metrics
	gmi.monitoringService.RecordMetric(Metric{
		Name:   "gpu_power_draw_watts",
		Type:   MetricGauge,
		Value:  metrics.PowerDraw,
		Labels: labels,
	})

	gmi.monitoringService.RecordMetric(Metric{
		Name:   "gpu_power_limit_watts",
		Type:   MetricGauge,
		Value:  metrics.PowerLimit,
		Labels: labels,
	})

	// Clock metrics
	gmi.monitoringService.RecordMetric(Metric{
		Name:   "gpu_clock_graphics_mhz",
		Type:   MetricGauge,
		Value:  float64(metrics.ClockGraphics),
		Labels: labels,
	})

	gmi.monitoringService.RecordMetric(Metric{
		Name:   "gpu_clock_memory_mhz",
		Type:   MetricGauge,
		Value:  float64(metrics.ClockMemory),
		Labels: labels,
	})

	// Process metrics
	gmi.monitoringService.RecordMetric(Metric{
		Name:   "gpu_process_count",
		Type:   MetricGauge,
		Value:  float64(metrics.ProcessCount),
		Labels: labels,
	})

	// Efficiency metrics
	powerEfficiency := 0.0
	if metrics.PowerDraw > 0 {
		powerEfficiency = metrics.UtilizationGPU / metrics.PowerDraw
	}

	gmi.monitoringService.RecordMetric(Metric{
		Name:   "gpu_power_efficiency",
		Type:   MetricGauge,
		Value:  powerEfficiency,
		Labels: labels,
	})
}

// checkAlerts checks for alert conditions in GPU metrics
func (gmi *GPUMetricsIntegration) checkAlerts(metrics gpu.GPUMetrics, lastState gpu.GPUMetrics, hasLastState bool) []gpu.GPUAlert {
	var alerts []gpu.GPUAlert

	// Temperature alerts
	if metrics.Temperature >= gmi.alertThresholds.CriticalTemperature {
		alerts = append(alerts, gpu.GPUAlert{
			Type:      "temperature",
			Severity:  "critical",
			Message:   fmt.Sprintf("GPU %s temperature critically high", metrics.GPUID),
			Value:     metrics.Temperature,
			Threshold: gmi.alertThresholds.CriticalTemperature,
			Timestamp: metrics.Timestamp,
		})
	} else if metrics.Temperature >= gmi.alertThresholds.HighTemperature {
		alerts = append(alerts, gpu.GPUAlert{
			Type:      "temperature",
			Severity:  "warning",
			Message:   fmt.Sprintf("GPU %s temperature high", metrics.GPUID),
			Value:     metrics.Temperature,
			Threshold: gmi.alertThresholds.HighTemperature,
			Timestamp: metrics.Timestamp,
		})
	}

	// Memory alerts
	memoryUsagePercent := float64(metrics.MemoryUsed) / float64(metrics.MemoryTotal) * 100
	if memoryUsagePercent >= gmi.alertThresholds.CriticalMemoryUsage {
		alerts = append(alerts, gpu.GPUAlert{
			Type:      "memory",
			Severity:  "critical",
			Message:   fmt.Sprintf("GPU %s memory usage critically high", metrics.GPUID),
			Value:     memoryUsagePercent,
			Threshold: gmi.alertThresholds.CriticalMemoryUsage,
			Timestamp: metrics.Timestamp,
		})
	} else if memoryUsagePercent >= gmi.alertThresholds.HighMemoryUsage {
		alerts = append(alerts, gpu.GPUAlert{
			Type:      "memory",
			Severity:  "warning",
			Message:   fmt.Sprintf("GPU %s memory usage high", metrics.GPUID),
			Value:     memoryUsagePercent,
			Threshold: gmi.alertThresholds.HighMemoryUsage,
			Timestamp: metrics.Timestamp,
		})
	}

	// Power alerts
	powerUsagePercent := 0.0
	if metrics.PowerLimit > 0 {
		powerUsagePercent = metrics.PowerDraw / metrics.PowerLimit * 100
	}

	if powerUsagePercent >= gmi.alertThresholds.CriticalPowerUsage {
		alerts = append(alerts, gpu.GPUAlert{
			Type:      "power",
			Severity:  "critical",
			Message:   fmt.Sprintf("GPU %s power usage critically high", metrics.GPUID),
			Value:     powerUsagePercent,
			Threshold: gmi.alertThresholds.CriticalPowerUsage,
			Timestamp: metrics.Timestamp,
		})
	} else if powerUsagePercent >= gmi.alertThresholds.HighPowerUsage {
		alerts = append(alerts, gpu.GPUAlert{
			Type:      "power",
			Severity:  "warning",
			Message:   fmt.Sprintf("GPU %s power usage high", metrics.GPUID),
			Value:     powerUsagePercent,
			Threshold: gmi.alertThresholds.HighPowerUsage,
			Timestamp: metrics.Timestamp,
		})
	}

	// Utilization alerts
	if metrics.UtilizationGPU >= gmi.alertThresholds.HighUtilization {
		alerts = append(alerts, gpu.GPUAlert{
			Type:      "utilization",
			Severity:  "info",
			Message:   fmt.Sprintf("GPU %s utilization very high", metrics.GPUID),
			Value:     metrics.UtilizationGPU,
			Threshold: gmi.alertThresholds.HighUtilization,
			Timestamp: metrics.Timestamp,
		})
	} else if metrics.UtilizationGPU <= gmi.alertThresholds.LowUtilization {
		alerts = append(alerts, gpu.GPUAlert{
			Type:      "utilization",
			Severity:  "info",
			Message:   fmt.Sprintf("GPU %s utilization low", metrics.GPUID),
			Value:     metrics.UtilizationGPU,
			Threshold: gmi.alertThresholds.LowUtilization,
			Timestamp: metrics.Timestamp,
		})
	}

	// State change alerts (if we have previous state)
	if hasLastState {
		// Significant utilization change
		utilizationDiff := metrics.UtilizationGPU - lastState.UtilizationGPU
		if utilizationDiff > 50.0 {
			alerts = append(alerts, gpu.GPUAlert{
				Type:      "utilization",
				Severity:  "info",
				Message:   fmt.Sprintf("GPU %s utilization increased significantly", metrics.GPUID),
				Value:     utilizationDiff,
				Threshold: 50.0,
				Timestamp: metrics.Timestamp,
			})
		}

		// Process count change
		if metrics.ProcessCount > lastState.ProcessCount {
			alerts = append(alerts, gpu.GPUAlert{
				Type:      "process",
				Severity:  "info",
				Message:   fmt.Sprintf("New process started on GPU %s", metrics.GPUID),
				Value:     float64(metrics.ProcessCount),
				Threshold: float64(lastState.ProcessCount),
				Timestamp: metrics.Timestamp,
			})
		}
	}

	return alerts
}

// recordAlertEvent records GPU alerts as events in the monitoring service
func (gmi *GPUMetricsIntegration) recordAlertEvent(alert gpu.GPUAlert, metrics gpu.GPUMetrics) {
	event := Event{
		Type:     "gpu_alert",
		Severity: alert.Severity,
		Message:  alert.Message,
		Source:   "gpu_metrics_integration",
		Metadata: map[string]interface{}{
			"gpu_id":     metrics.GPUID,
			"gpu_name":   metrics.Name,
			"alert_type": alert.Type,
			"value":      alert.Value,
			"threshold":  alert.Threshold,
		},
	}

	gmi.monitoringService.RecordEvent(event)
}

// recordGPUCosts estimates and records GPU operational costs
func (gmi *GPUMetricsIntegration) recordGPUCosts(metrics gpu.GPUMetrics, lastState gpu.GPUMetrics, hasLastState bool) {
	if !hasLastState {
		return // Need previous state to calculate time-based costs
	}

	// Calculate time since last measurement
	duration := metrics.Timestamp.Sub(lastState.Timestamp)
	hours := duration.Hours()

	// Estimate cost based on GPU type and power consumption
	// These are rough estimates - real costs would come from cloud provider APIs
	var costPerHour float64
	switch {
	case strings.Contains(strings.ToLower(metrics.Name), "a100"):
		costPerHour = 3.06 // Approximate AWS p4d.xlarge cost
	case strings.Contains(strings.ToLower(metrics.Name), "v100"):
		costPerHour = 3.06 // Approximate AWS p3.2xlarge cost
	case strings.Contains(strings.ToLower(metrics.Name), "t4"):
		costPerHour = 0.526 // Approximate AWS g4dn.xlarge cost
	case strings.Contains(strings.ToLower(metrics.Name), "rtx"):
		costPerHour = 1.00 // Estimate for RTX series
	default:
		costPerHour = 1.50 // Default estimate
	}

	// Adjust cost based on actual utilization (idle time costs less)
	utilizationFactor := (metrics.UtilizationGPU + lastState.UtilizationGPU) / 200.0 // Average utilization
	if utilizationFactor < 0.1 {
		utilizationFactor = 0.1 // Minimum cost factor for idle GPU
	}

	actualCost := costPerHour * hours * utilizationFactor

	// Record cost entry
	costEntry := CostEntry{
		ID:        fmt.Sprintf("gpu-%s-%d", metrics.GPUID, time.Now().Unix()),
		Operation: "gpu_compute",
		ModelID:   "gpu_workload", // Generic model ID for GPU compute
		Duration:  duration,
		GPUHours:  hours,
		Cost:      actualCost,
		Currency:  "USD",
		Timestamp: time.Now(),
	}

	gmi.monitoringService.RecordCost(costEntry)
}

// GetGPUHealth returns health status for all monitored GPUs
func (gmi *GPUMetricsIntegration) GetGPUHealth() map[string]gpu.GPUHealthStatus {
	gmi.mu.RLock()
	defer gmi.mu.RUnlock()

	health := make(map[string]gpu.GPUHealthStatus)

	for gpuID, metrics := range gmi.lastKnownState {
		status := gmi.calculateHealthStatus(gpuID, metrics)
		health[gpuID] = status
	}

	return health
}

// calculateHealthStatus calculates health status for a GPU
func (gmi *GPUMetricsIntegration) calculateHealthStatus(gpuID string, metrics gpu.GPUMetrics) gpu.GPUHealthStatus {
	status := gpu.GPUHealthStatus{
		GPUID:           gpuID,
		Status:          "healthy",
		Timestamp:       time.Now(),
		Issues:          make([]string, 0),
		Recommendations: make([]string, 0),
	}

	// Check temperature status
	if metrics.Temperature >= gmi.alertThresholds.CriticalTemperature {
		status.TemperatureStatus = "critical"
		status.Status = "critical"
		status.Issues = append(status.Issues, fmt.Sprintf("Temperature critically high: %.1f°C", metrics.Temperature))
		status.Recommendations = append(status.Recommendations, "Check cooling system and reduce workload")
	} else if metrics.Temperature >= gmi.alertThresholds.HighTemperature {
		status.TemperatureStatus = "warning"
		if status.Status == "healthy" {
			status.Status = "warning"
		}
		status.Issues = append(status.Issues, fmt.Sprintf("Temperature elevated: %.1f°C", metrics.Temperature))
		status.Recommendations = append(status.Recommendations, "Monitor temperature trends")
	} else {
		status.TemperatureStatus = "healthy"
	}

	// Check memory status
	memoryUsagePercent := float64(metrics.MemoryUsed) / float64(metrics.MemoryTotal) * 100
	if memoryUsagePercent >= gmi.alertThresholds.CriticalMemoryUsage {
		status.MemoryStatus = "critical"
		status.Status = "critical"
		status.Issues = append(status.Issues, fmt.Sprintf("Memory usage critically high: %.1f%%", memoryUsagePercent))
		status.Recommendations = append(status.Recommendations, "Reduce memory usage or scale workloads")
	} else if memoryUsagePercent >= gmi.alertThresholds.HighMemoryUsage {
		status.MemoryStatus = "warning"
		if status.Status == "healthy" {
			status.Status = "warning"
		}
		status.Issues = append(status.Issues, fmt.Sprintf("Memory usage high: %.1f%%", memoryUsagePercent))
		status.Recommendations = append(status.Recommendations, "Consider memory optimization")
	} else {
		status.MemoryStatus = "healthy"
	}

	// Check power status
	powerUsagePercent := 0.0
	if metrics.PowerLimit > 0 {
		powerUsagePercent = metrics.PowerDraw / metrics.PowerLimit * 100
	}

	if powerUsagePercent >= gmi.alertThresholds.CriticalPowerUsage {
		status.PowerStatus = "critical"
		status.Status = "critical"
		status.Issues = append(status.Issues, fmt.Sprintf("Power usage critically high: %.1f%%", powerUsagePercent))
	} else if powerUsagePercent >= gmi.alertThresholds.HighPowerUsage {
		status.PowerStatus = "warning"
		if status.Status == "healthy" {
			status.Status = "warning"
		}
		status.Issues = append(status.Issues, fmt.Sprintf("Power usage high: %.1f%%", powerUsagePercent))
	} else {
		status.PowerStatus = "healthy"
	}

	// Check utilization status
	if metrics.UtilizationGPU <= gmi.alertThresholds.LowUtilization {
		status.UtilizationStatus = "underutilized"
		status.Issues = append(status.Issues, fmt.Sprintf("Low utilization: %.1f%%", metrics.UtilizationGPU))
		status.Recommendations = append(status.Recommendations, "Consider consolidating workloads or scaling down")
	} else if metrics.UtilizationGPU >= gmi.alertThresholds.HighUtilization {
		status.UtilizationStatus = "high"
		status.Issues = append(status.Issues, fmt.Sprintf("High utilization: %.1f%%", metrics.UtilizationGPU))
		status.Recommendations = append(status.Recommendations, "Monitor for performance bottlenecks")
	} else {
		status.UtilizationStatus = "optimal"
	}

	// Add recent alerts
	if alerts, exists := gmi.alertHistory[gpuID]; exists {
		recentAlerts := make([]gpu.GPUAlert, 0)
		cutoff := time.Now().Add(-5 * time.Minute) // Last 5 minutes

		for _, alert := range alerts {
			if alert.Timestamp.After(cutoff) {
				recentAlerts = append(recentAlerts, alert)
			}
		}

		status.Alerts = recentAlerts
	}

	return status
}

// GetAlertHistory returns alert history for a specific GPU
func (gmi *GPUMetricsIntegration) GetAlertHistory(gpuID string, since time.Time) []gpu.GPUAlert {
	gmi.mu.RLock()
	defer gmi.mu.RUnlock()

	alerts, exists := gmi.alertHistory[gpuID]
	if !exists {
		return []gpu.GPUAlert{}
	}

	result := make([]gpu.GPUAlert, 0)
	for _, alert := range alerts {
		if alert.Timestamp.After(since) {
			result = append(result, alert)
		}
	}

	return result
}
