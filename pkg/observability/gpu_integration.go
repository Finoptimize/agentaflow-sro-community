package observability

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Finoptimize/agentaflow-sro-community/pkg/gpu"
)

// GPU cost configuration constants
const (
	// Default GPU hourly costs (USD) - based on common cloud provider pricing
	DefaultCostA100     = 3.06   // AWS p4d.xlarge approximate cost
	DefaultCostV100     = 3.06   // AWS p3.2xlarge approximate cost  
	DefaultCostT4       = 0.526  // AWS g4dn.xlarge approximate cost
	DefaultCostRTX      = 1.00   // Estimate for RTX series
	DefaultCostGeneric  = 1.50   // Default fallback cost
	
	// Utilization cost factors
	MinUtilizationFactor = 0.1   // Minimum cost factor for idle GPUs
	MaxUtilizationFactor = 1.0   // Maximum cost factor for full utilization
)

// GPUCostConfiguration defines cost settings for GPU types and pricing models
type GPUCostConfiguration struct {
	// Cost per hour by GPU type (USD)
	CostPerHour map[string]float64
	
	// Pricing model settings
	UseUtilizationFactor bool    // Whether to adjust cost based on utilization
	MinUtilizationFactor float64 // Minimum cost factor for idle GPUs
	IdleCostReduction    float64 // Cost reduction factor for idle time (0.0-1.0)
	
	// Cloud provider settings
	CloudProvider        string             // AWS, GCP, Azure, etc.
	Region              string             // Cloud region for regional pricing
	CustomPricing       map[string]float64 // Custom pricing overrides
	
	// Currency and billing
	Currency            string  // Cost currency (USD, EUR, etc.)
	TaxRate             float64 // Tax rate to apply (0.0-1.0)
	
	// Advanced pricing features
	SpotInstanceDiscount float64            // Discount for spot instances (0.0-1.0)
	ReservedInstanceCost map[string]float64 // Reserved instance pricing
	VolumeDiscounts     []VolumeDiscount   // Volume-based discounts
}

// VolumeDiscount defines volume-based pricing discounts
type VolumeDiscount struct {
	MinHours     float64 // Minimum hours for discount to apply
	DiscountRate float64 // Discount rate (0.0-1.0)
}

// DefaultGPUCostConfiguration returns a default cost configuration
func DefaultGPUCostConfiguration() GPUCostConfiguration {
	return GPUCostConfiguration{
		CostPerHour: map[string]float64{
			"a100":    DefaultCostA100,
			"v100":    DefaultCostV100, 
			"t4":      DefaultCostT4,
			"rtx":     DefaultCostRTX,
			"generic": DefaultCostGeneric,
		},
		UseUtilizationFactor: true,
		MinUtilizationFactor: MinUtilizationFactor,
		IdleCostReduction:    0.1, // 10% cost for idle time
		CloudProvider:        "aws",
		Region:              "us-west-2",
		Currency:            "USD",
		TaxRate:             0.0,
		SpotInstanceDiscount: 0.0,
		CustomPricing:       make(map[string]float64),
		ReservedInstanceCost: make(map[string]float64),
		VolumeDiscounts:     []VolumeDiscount{},
	}
}

// GPUMetricsIntegration connects GPU metrics collection with observability monitoring
type GPUMetricsIntegration struct {
	monitoringService *MonitoringService
	metricsCollector  *gpu.MetricsCollector
	mu                sync.RWMutex

	// Configuration
	alertThresholds GPUAlertThresholds
	costConfig      GPUCostConfiguration // Add cost configuration
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
		costConfig:        DefaultGPUCostConfiguration(), // Initialize with defaults
		metricsEnabled:    true,
		eventsEnabled:     true,
		costsEnabled:      true,
		lastKnownState:    make(map[string]gpu.GPUMetrics),
		alertHistory:      make(map[string][]gpu.GPUAlert),
	}

	// Register callback with metrics collector
	if metricsCollector != nil {
		metricsCollector.RegisterCallback(integration.processGPUMetrics)
	}

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

// SetCostConfiguration configures custom cost settings
func (gmi *GPUMetricsIntegration) SetCostConfiguration(config GPUCostConfiguration) {
	gmi.mu.Lock()
	defer gmi.mu.Unlock()
	gmi.costConfig = config
}

// GetCostConfiguration returns the current cost configuration
func (gmi *GPUMetricsIntegration) GetCostConfiguration() GPUCostConfiguration {
	gmi.mu.RLock()
	defer gmi.mu.RUnlock()
	return gmi.costConfig
}

// UpdateGPUCost updates the cost for a specific GPU type
func (gmi *GPUMetricsIntegration) UpdateGPUCost(gpuType string, costPerHour float64) {
	gmi.mu.Lock()
	defer gmi.mu.Unlock()
	if gmi.costConfig.CostPerHour == nil {
		gmi.costConfig.CostPerHour = make(map[string]float64)
	}
	gmi.costConfig.CostPerHour[gpuType] = costPerHour
}

// SetCloudProviderPricing configures pricing for a specific cloud provider
func (gmi *GPUMetricsIntegration) SetCloudProviderPricing(provider, region string, pricing map[string]float64) {
	gmi.mu.Lock()
	defer gmi.mu.Unlock()
	gmi.costConfig.CloudProvider = provider
	gmi.costConfig.Region = region
	for gpuType, cost := range pricing {
		if gmi.costConfig.CostPerHour == nil {
			gmi.costConfig.CostPerHour = make(map[string]float64)
		}
		gmi.costConfig.CostPerHour[gpuType] = cost
	}
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
		// Limit alert history size
		const maxAlertsPerGPU = 100
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
	if hours <= 0 {
		return // Invalid duration
	}

	// Get cost per hour for this GPU type using the new configuration system
	costPerHour := gmi.getGPUCostPerHour(metrics.Name)
	
	// Calculate utilization factor if enabled
	utilizationFactor := 1.0
	if gmi.costConfig.UseUtilizationFactor {
		utilizationFactor = gmi.calculateUtilizationFactor(metrics, lastState)
	}
	
	// Apply spot instance discount if configured
	spotDiscount := 1.0 - gmi.costConfig.SpotInstanceDiscount
	
	// Calculate base cost
	baseCost := costPerHour * hours * utilizationFactor * spotDiscount
	
	// Apply volume discounts if any
	finalCost := gmi.applyVolumeDiscounts(baseCost, hours)
	
	// Apply tax if configured
	if gmi.costConfig.TaxRate > 0 {
		finalCost *= (1.0 + gmi.costConfig.TaxRate)
	}

	// Record cost entry
	costEntry := CostEntry{
		ID:        fmt.Sprintf("gpu-%s-%d", metrics.GPUID, time.Now().Unix()),
		Operation: "gpu_compute",
		ModelID:   fmt.Sprintf("gpu_%s", gmi.normalizeGPUType(metrics.Name)),
		Duration:  duration,
		GPUHours:  hours,
		Cost:      finalCost,
		Currency:  gmi.costConfig.Currency,
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

// getGPUCostPerHour returns the cost per hour for a given GPU type
func (gmi *GPUMetricsIntegration) getGPUCostPerHour(gpuName string) float64 {
	gpuType := gmi.normalizeGPUType(gpuName)
	
	// Check custom pricing first
	if cost, exists := gmi.costConfig.CustomPricing[gpuType]; exists {
		return cost
	}
	
	// Check reserved instance pricing
	if cost, exists := gmi.costConfig.ReservedInstanceCost[gpuType]; exists {
		return cost
	}
	
	// Check standard pricing
	if cost, exists := gmi.costConfig.CostPerHour[gpuType]; exists {
		return cost
	}
	
	// Return generic cost as fallback
	if cost, exists := gmi.costConfig.CostPerHour["generic"]; exists {
		return cost
	}
	
	return DefaultCostGeneric
}

// normalizeGPUType extracts and normalizes GPU type from GPU name
func (gmi *GPUMetricsIntegration) normalizeGPUType(gpuName string) string {
	lowerName := strings.ToLower(gpuName)
	
	switch {
	case strings.Contains(lowerName, "a100"):
		return "a100"
	case strings.Contains(lowerName, "v100"):
		return "v100"
	case strings.Contains(lowerName, "t4"):
		return "t4"
	case strings.Contains(lowerName, "rtx"):
		return "rtx"
	case strings.Contains(lowerName, "h100"):
		return "h100"
	case strings.Contains(lowerName, "a10"):
		return "a10"
	case strings.Contains(lowerName, "k80"):
		return "k80"
	default:
		return "generic"
	}
}

// calculateUtilizationFactor calculates cost adjustment based on GPU utilization
func (gmi *GPUMetricsIntegration) calculateUtilizationFactor(current, previous gpu.GPUMetrics) float64 {
	// Average utilization over the measurement period
	avgUtilization := (current.UtilizationGPU + previous.UtilizationGPU) / 2.0 / 100.0
	
	// Apply idle cost reduction for underutilized GPUs
	if avgUtilization < 0.1 { // Less than 10% utilization
		return gmi.costConfig.MinUtilizationFactor + 
			(gmi.costConfig.IdleCostReduction * avgUtilization)
	}
	
	// Linear factor based on utilization
	factor := gmi.costConfig.MinUtilizationFactor + 
		(avgUtilization * (MaxUtilizationFactor - gmi.costConfig.MinUtilizationFactor))
	
	// Ensure factor is within bounds
	if factor < gmi.costConfig.MinUtilizationFactor {
		return gmi.costConfig.MinUtilizationFactor
	}
	if factor > MaxUtilizationFactor {
		return MaxUtilizationFactor
	}
	
	return factor
}

// applyVolumeDiscounts applies volume-based discounts to the cost
func (gmi *GPUMetricsIntegration) applyVolumeDiscounts(cost, hours float64) float64 {
	if len(gmi.costConfig.VolumeDiscounts) == 0 {
		return cost
	}
	
	// Find the highest applicable discount
	var bestDiscount float64
	for _, discount := range gmi.costConfig.VolumeDiscounts {
		if hours >= discount.MinHours && discount.DiscountRate > bestDiscount {
			bestDiscount = discount.DiscountRate
		}
	}
	
	return cost * (1.0 - bestDiscount)
}
