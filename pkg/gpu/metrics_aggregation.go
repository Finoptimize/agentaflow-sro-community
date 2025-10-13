package gpu

import (
	"context"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"
)

// MetricsAggregationService provides advanced GPU metrics aggregation and analytics
type MetricsAggregationService struct {
	metricsCollector *MetricsCollector
	mu               sync.RWMutex

	// Aggregated data
	gpuStats       map[string]*GPUStats
	clusterMetrics *ClusterMetrics

	// Configuration
	aggregationInterval time.Duration
	retentionPeriod     time.Duration

	// State
	ctx             context.Context
	cancel          context.CancelFunc
	running         bool
	lastAggregation time.Time
}

// NewMetricsAggregationService creates a new metrics aggregation service
func NewMetricsAggregationService(
	metricsCollector *MetricsCollector,
	aggregationInterval time.Duration,
	retentionPeriod time.Duration,
) *MetricsAggregationService {
	ctx, cancel := context.WithCancel(context.Background())

	return &MetricsAggregationService{
		metricsCollector:    metricsCollector,
		gpuStats:            make(map[string]*GPUStats),
		aggregationInterval: aggregationInterval,
		retentionPeriod:     retentionPeriod,
		ctx:                 ctx,
		cancel:              cancel,
	}
}

// Start begins the metrics aggregation process
func (mas *MetricsAggregationService) Start() error {
	mas.mu.Lock()
	defer mas.mu.Unlock()

	if mas.running {
		return fmt.Errorf("metrics aggregation service is already running")
	}

	mas.running = true
	go mas.aggregationLoop()

	return nil
}

// Stop stops the metrics aggregation process
func (mas *MetricsAggregationService) Stop() {
	mas.mu.Lock()
	defer mas.mu.Unlock()

	if mas.running {
		mas.cancel()
		mas.running = false
	}
}

// aggregationLoop performs periodic metrics aggregation
func (mas *MetricsAggregationService) aggregationLoop() {
	ticker := time.NewTicker(mas.aggregationInterval)
	defer ticker.Stop()

	for {
		select {
		case <-mas.ctx.Done():
			return
		case <-ticker.C:
			mas.performAggregation()
		}
	}
}

// performAggregation aggregates metrics for all GPUs
func (mas *MetricsAggregationService) performAggregation() {
	now := time.Now()

	// Get latest metrics from collector
	latestMetrics := mas.metricsCollector.GetLatestMetrics()

	mas.mu.Lock()
	defer mas.mu.Unlock()

	// Update GPU stats for each GPU
	for gpuID, metrics := range latestMetrics {
		mas.updateGPUStats(gpuID, metrics, now)
	}

	// Update cluster-wide metrics
	mas.updateClusterMetrics(latestMetrics, now)

	mas.lastAggregation = now
}

// updateGPUStats updates statistics for a single GPU
func (mas *MetricsAggregationService) updateGPUStats(gpuID string, metrics GPUMetrics, now time.Time) {
	stats, exists := mas.gpuStats[gpuID]
	if !exists {
		stats = &GPUStats{
			GPUID:     gpuID,
			StartTime: now,
		}
		mas.gpuStats[gpuID] = stats
	}

	// Get historical metrics for the retention period
	history := mas.metricsCollector.GetMetricsHistory(gpuID, now.Add(-mas.retentionPeriod))
	if len(history) == 0 {
		return
	}

	// Calculate aggregated statistics
	mas.calculateGPUStatistics(stats, history, now)
}

// calculateGPUStatistics calculates comprehensive statistics for a GPU
func (mas *MetricsAggregationService) calculateGPUStatistics(stats *GPUStats, history []GPUMetrics, now time.Time) {
	if len(history) == 0 {
		return
	}

	// Initialize tracking variables
	totalUtilization := 0.0
	totalMemoryUsage := 0.0
	totalTemperature := 0.0
	totalPowerDraw := 0.0
	totalEnergyConsumed := 0.0

	maxUtilization := 0.0
	maxMemoryUsage := uint64(0)
	maxTemperature := 0.0
	maxPowerDraw := 0.0

	idleTimeSeconds := 0.0
	processSwitches := 0
	lastProcessCount := -1

	// Calculate time-based metrics
	for i, metric := range history {
		totalUtilization += metric.UtilizationGPU
		totalMemoryUsage += float64(metric.MemoryUsed)
		totalTemperature += metric.Temperature
		totalPowerDraw += metric.PowerDraw

		// Track maximums
		if metric.UtilizationGPU > maxUtilization {
			maxUtilization = metric.UtilizationGPU
		}
		if metric.MemoryUsed > maxMemoryUsage {
			maxMemoryUsage = metric.MemoryUsed
		}
		if metric.Temperature > maxTemperature {
			maxTemperature = metric.Temperature
		}
		if metric.PowerDraw > maxPowerDraw {
			maxPowerDraw = metric.PowerDraw
		}

		// Calculate idle time (utilization < 5%)
		if i > 0 {
			timeDiff := metric.Timestamp.Sub(history[i-1].Timestamp).Seconds()
			const IdleThreshold = 5.0

						if metric.UtilizationGPU < IdleThreshold {
							idleTimeSeconds += timeDiff
			}
		} // Track process switches
		if lastProcessCount != -1 && metric.ProcessCount != lastProcessCount {
			processSwitches++
		}
		lastProcessCount = metric.ProcessCount

		// Calculate energy consumption (power × time)
		if i > 0 {
			timeDiffHours := metric.Timestamp.Sub(history[i-1].Timestamp).Hours()
			avgPower := (metric.PowerDraw + history[i-1].PowerDraw) / 2
			totalEnergyConsumed += avgPower * timeDiffHours / 1000 // Convert to kWh
		}
	}

	// Calculate averages
	count := float64(len(history))
	stats.AverageUtilization = totalUtilization / count
	stats.AverageMemoryUsage = totalMemoryUsage / count
	stats.AverageTemperature = totalTemperature / count
	stats.AveragePowerDraw = totalPowerDraw / count

	// Set maximums
	stats.PeakUtilization = maxUtilization
	stats.PeakMemoryUsage = maxMemoryUsage
	stats.MaxTemperature = maxTemperature
	stats.MaxPowerDraw = maxPowerDraw

	// Calculate derived metrics
	totalTimeSeconds := history[len(history)-1].Timestamp.Sub(history[0].Timestamp).Seconds()
	stats.IdleTimePercentage = (idleTimeSeconds / totalTimeSeconds) * 100

	// Calculate efficiency score (utilization per watt)
	if stats.AveragePowerDraw > 0 {
		stats.EfficiencyScore = stats.AverageUtilization / stats.AveragePowerDraw
	}

	stats.ProcessSwitches = processSwitches
	stats.TotalEnergyConsumed = totalEnergyConsumed
	stats.UptimeHours = totalTimeSeconds / 3600
	stats.Period = mas.retentionPeriod
	stats.EndTime = now
}

// updateClusterMetrics updates cluster-wide metrics
func (mas *MetricsAggregationService) updateClusterMetrics(latestMetrics map[string]GPUMetrics, now time.Time) {
	clusterMetrics := &ClusterMetrics{
		GPUStats:  make(map[string]GPUStats),
		GPUHealth: make(map[string]GPUHealthStatus),
		Timestamp: now,
	}

	totalMemoryMB := uint64(0)
	usedMemoryMB := uint64(0)
	totalUtilization := 0.0
	totalTemperature := 0.0
	totalPowerDraw := 0.0
	totalProcesses := 0
	activeGPUs := 0
	healthyGPUs := 0

	// Aggregate metrics across all GPUs
	for gpuID, metrics := range latestMetrics {
		totalMemoryMB += metrics.MemoryTotal
		usedMemoryMB += metrics.MemoryUsed
		totalUtilization += metrics.UtilizationGPU
		totalTemperature += metrics.Temperature
		totalPowerDraw += metrics.PowerDraw
		totalProcesses += metrics.ProcessCount

		if metrics.UtilizationGPU > 5.0 {
			activeGPUs++
		}

		// Copy GPU stats if available
		if stats, exists := mas.gpuStats[gpuID]; exists {
			clusterMetrics.GPUStats[gpuID] = *stats
		}

		// Generate health status (simplified)
		healthStatus := mas.calculateSimpleHealthStatus(metrics)
		clusterMetrics.GPUHealth[gpuID] = healthStatus

		if healthStatus.Status == "healthy" {
			healthyGPUs++
		}
	}

	// Calculate cluster averages
	gpuCount := len(latestMetrics)
	if gpuCount > 0 {
		clusterMetrics.AverageUtilization = totalUtilization / float64(gpuCount)
		clusterMetrics.AverageTemperature = totalTemperature / float64(gpuCount)
	}

	clusterMetrics.TotalGPUs = gpuCount
	clusterMetrics.AvailableGPUs = gpuCount // Simplified - all discovered GPUs are available
	clusterMetrics.ActiveGPUs = activeGPUs
	clusterMetrics.HealthyGPUs = healthyGPUs
	clusterMetrics.TotalMemoryMB = totalMemoryMB
	clusterMetrics.UsedMemoryMB = usedMemoryMB
	clusterMetrics.TotalPowerDraw = totalPowerDraw
	clusterMetrics.TotalProcesses = totalProcesses

	mas.clusterMetrics = clusterMetrics
}

// calculateSimpleHealthStatus creates a basic health status for a GPU
func (mas *MetricsAggregationService) calculateSimpleHealthStatus(metrics GPUMetrics) GPUHealthStatus {
	status := GPUHealthStatus{
		GPUID:           metrics.GPUID,
		Timestamp:       metrics.Timestamp,
		Status:          "healthy",
		Issues:          make([]string, 0),
		Recommendations: make([]string, 0),
		Alerts:          make([]GPUAlert, 0),
	}

	// Simple health checks
	if metrics.Temperature > 85.0 {
		status.Status = "critical"
		status.TemperatureStatus = "critical"
		status.Issues = append(status.Issues, fmt.Sprintf("Temperature too high: %.1f°C", metrics.Temperature))
	} else if metrics.Temperature > 75.0 {
		status.Status = "warning"
		status.TemperatureStatus = "warning"
		status.Issues = append(status.Issues, fmt.Sprintf("Temperature elevated: %.1f°C", metrics.Temperature))
	} else {
		status.TemperatureStatus = "healthy"
	}

	memoryUsagePercent := float64(metrics.MemoryUsed) / float64(metrics.MemoryTotal) * 100
	if memoryUsagePercent > 95.0 {
		status.Status = "critical"
		status.MemoryStatus = "critical"
		status.Issues = append(status.Issues, fmt.Sprintf("Memory usage critically high: %.1f%%", memoryUsagePercent))
	} else if memoryUsagePercent > 80.0 {
		if status.Status == "healthy" {
			status.Status = "warning"
		}
		status.MemoryStatus = "warning"
		status.Issues = append(status.Issues, fmt.Sprintf("Memory usage high: %.1f%%", memoryUsagePercent))
	} else {
		status.MemoryStatus = "healthy"
	}

	return status
}

// GetGPUStats returns aggregated statistics for a specific GPU
func (mas *MetricsAggregationService) GetGPUStats(gpuID string) (*GPUStats, error) {
	mas.mu.RLock()
	defer mas.mu.RUnlock()

	stats, exists := mas.gpuStats[gpuID]
	if !exists {
		return nil, fmt.Errorf("no statistics available for GPU %s", gpuID)
	}

	// Return a copy to avoid race conditions
	statsCopy := *stats
	return &statsCopy, nil
}

// GetAllGPUStats returns aggregated statistics for all GPUs
func (mas *MetricsAggregationService) GetAllGPUStats() map[string]GPUStats {
	mas.mu.RLock()
	defer mas.mu.RUnlock()

	result := make(map[string]GPUStats)
	for gpuID, stats := range mas.gpuStats {
		result[gpuID] = *stats
	}

	return result
}

// GetClusterMetrics returns cluster-wide metrics
func (mas *MetricsAggregationService) GetClusterMetrics() *ClusterMetrics {
	mas.mu.RLock()
	defer mas.mu.RUnlock()

	if mas.clusterMetrics == nil {
		return nil
	}

	// Return a copy to avoid race conditions
	clusterCopy := *mas.clusterMetrics
	return &clusterCopy
}

// GetEfficiencyReport generates a comprehensive efficiency report
func (mas *MetricsAggregationService) GetEfficiencyReport() map[string]interface{} {
	mas.mu.RLock()
	defer mas.mu.RUnlock()

	report := make(map[string]interface{})

	// Cluster-level efficiency metrics
	totalIdleTime := 0.0
	totalEfficiency := 0.0
	gpuCount := 0

	for _, stats := range mas.gpuStats {
		totalIdleTime += stats.IdleTimePercentage
		totalEfficiency += stats.EfficiencyScore
		gpuCount++
	}

	if gpuCount > 0 {
		avgIdleTime := totalIdleTime / float64(gpuCount)
		avgEfficiency := totalEfficiency / float64(gpuCount)

		report["cluster_efficiency"] = map[string]interface{}{
			"average_idle_time_percent": avgIdleTime,
			"average_efficiency_score":  avgEfficiency,
			"total_gpus":                gpuCount,
			"utilization_potential":     100.0 - avgIdleTime,
		}
	}

	// GPU-specific efficiency rankings
	type gpuEfficiency struct {
		GPUID           string
		EfficiencyScore float64
		IdleTime        float64
		Utilization     float64
	}

	efficiencies := make([]gpuEfficiency, 0, len(mas.gpuStats))
	for gpuID, stats := range mas.gpuStats {
		efficiencies = append(efficiencies, gpuEfficiency{
			GPUID:           gpuID,
			EfficiencyScore: stats.EfficiencyScore,
			IdleTime:        stats.IdleTimePercentage,
			Utilization:     stats.AverageUtilization,
		})
	}

	// Sort by efficiency score (descending)
	sort.Slice(efficiencies, func(i, j int) bool {
		return efficiencies[i].EfficiencyScore > efficiencies[j].EfficiencyScore
	})

	report["gpu_rankings"] = efficiencies
	report["generated_at"] = time.Now()
	report["retention_period"] = mas.retentionPeriod.String()

	return report
}

// GetPerformanceTrends analyzes performance trends over time
func (mas *MetricsAggregationService) GetPerformanceTrends(gpuID string, period time.Duration) map[string]interface{} {
	history := mas.metricsCollector.GetMetricsHistory(gpuID, time.Now().Add(-period))

	if len(history) < 2 {
		return map[string]interface{}{
			"error": "insufficient data for trend analysis",
		}
	}

	// Calculate trends
	utilizationTrend := mas.calculateTrend(history, func(m GPUMetrics) float64 { return m.UtilizationGPU })
	temperatureTrend := mas.calculateTrend(history, func(m GPUMetrics) float64 { return m.Temperature })
	memoryTrend := mas.calculateTrend(history, func(m GPUMetrics) float64 { return float64(m.MemoryUsed) })
	powerTrend := mas.calculateTrend(history, func(m GPUMetrics) float64 { return m.PowerDraw })

	return map[string]interface{}{
		"gpu_id":            gpuID,
		"period_hours":      period.Hours(),
		"sample_count":      len(history),
		"utilization_trend": utilizationTrend,
		"temperature_trend": temperatureTrend,
		"memory_trend":      memoryTrend,
		"power_trend":       powerTrend,
		"analysis_time":     time.Now(),
	}
}

// calculateTrend calculates the trend (slope) of a metric over time using linear regression
func (mas *MetricsAggregationService) calculateTrend(history []GPUMetrics, valueFunc func(GPUMetrics) float64) map[string]float64 {
	if len(history) < 2 {
		return map[string]float64{"slope": 0, "r_squared": 0}
	}

	n := float64(len(history))

	// Convert timestamps to hours since start
	startTime := history[0].Timestamp

	var sumX, sumY, sumXY, sumX2, sumY2 float64

	for _, metric := range history {
		x := metric.Timestamp.Sub(startTime).Hours()
		y := valueFunc(metric)

		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
		sumY2 += y * y
	}

	// Calculate linear regression
	slope := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)

	// Calculate R-squared (coefficient of determination)
	meanY := sumY / n
	var ssRes, ssTot float64

	for _, metric := range history {
		x := metric.Timestamp.Sub(startTime).Hours()
		y := valueFunc(metric)

		predicted := slope*x + (sumY-slope*sumX)/n
		ssRes += (y - predicted) * (y - predicted)
		ssTot += (y - meanY) * (y - meanY)
	}

	rSquared := 1.0 - (ssRes / ssTot)
	if math.IsNaN(rSquared) || math.IsInf(rSquared, 0) {
		rSquared = 0.0
	}

	return map[string]float64{
		"slope":     slope,
		"r_squared": rSquared,
	}
}

// GetCostAnalysis provides cost analysis based on GPU metrics
func (mas *MetricsAggregationService) GetCostAnalysis() map[string]interface{} {
	mas.mu.RLock()
	defer mas.mu.RUnlock()

	analysis := make(map[string]interface{})

	totalCostEstimate := 0.0
	totalPotentialSavings := 0.0

	gpuCosts := make(map[string]interface{})

	for gpuID, stats := range mas.gpuStats {
		// Estimate cost based on uptime and utilization
		// This is a simplified model - real implementations would use actual cloud pricing

		var costPerHour float64
		switch {
		case stats.AverageUtilization > 80:
			costPerHour = 3.06 // High-performance instance
		case stats.AverageUtilization > 40:
			costPerHour = 1.53 // Medium instance
		default:
			costPerHour = 0.76 // Basic instance
		}

		actualCost := costPerHour * stats.UptimeHours

		// Calculate potential savings from better utilization
		utilizationFactor := stats.AverageUtilization / 100.0
		if utilizationFactor < 0.1 {
			utilizationFactor = 0.1
		}

		optimizedCost := actualCost * utilizationFactor
		potentialSavings := actualCost - optimizedCost

		gpuCosts[gpuID] = map[string]interface{}{
			"actual_cost":       actualCost,
			"optimized_cost":    optimizedCost,
			"potential_savings": potentialSavings,
			"cost_per_hour":     costPerHour,
			"uptime_hours":      stats.UptimeHours,
			"avg_utilization":   stats.AverageUtilization,
			"efficiency_score":  stats.EfficiencyScore,
		}

		totalCostEstimate += actualCost
		totalPotentialSavings += potentialSavings
	}

	analysis["total_estimated_cost"] = totalCostEstimate
	analysis["total_potential_savings"] = totalPotentialSavings
	analysis["savings_percentage"] = (totalPotentialSavings / totalCostEstimate) * 100
	analysis["gpu_costs"] = gpuCosts
	analysis["analysis_time"] = time.Now()

	return analysis
}
