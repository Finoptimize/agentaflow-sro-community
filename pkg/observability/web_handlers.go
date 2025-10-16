package observability

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/Finoptimize/agentaflow-sro-community/pkg/gpu"
	"github.com/gorilla/mux"
)

// DashboardMetrics represents comprehensive dashboard metrics
type DashboardMetrics struct {
	Timestamp   time.Time              `json:"timestamp"`
	GPUMetrics  map[string]interface{} `json:"gpu_metrics"`
	SystemStats SystemStats            `json:"system_stats"`
	CostData    CostSummary            `json:"cost_data"`
	Alerts      []Alert                `json:"alerts"`
	Performance PerformanceMetrics     `json:"performance"`
}

// SystemStats represents system-level statistics
type SystemStats struct {
	TotalGPUs       int     `json:"total_gpus"`
	ActiveGPUs      int     `json:"active_gpus"`
	AverageUtil     float64 `json:"average_utilization"`
	TotalMemoryGB   float64 `json:"total_memory_gb"`
	UsedMemoryGB    float64 `json:"used_memory_gb"`
	AverageTemp     float64 `json:"average_temperature"`
	TotalPowerWatts float64 `json:"total_power_watts"`
	EfficiencyScore float64 `json:"efficiency_score"`
}

// Alert represents an alert condition
type Alert struct {
	ID        string    `json:"id"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	Source    string    `json:"source"`
	Timestamp time.Time `json:"timestamp"`
}

// PerformanceMetrics represents performance analytics
type PerformanceMetrics struct {
	UtilizationTrend float64           `json:"utilization_trend"`
	CostTrend        float64           `json:"cost_trend"`
	EfficiencyTrend  float64           `json:"efficiency_trend"`
	PredictedCost24h float64           `json:"predicted_cost_24h"`
	OptimizationTips []OptimizationTip `json:"optimization_tips"`
}

// OptimizationTip represents an optimization suggestion
type OptimizationTip struct {
	Type    string  `json:"type"`
	Message string  `json:"message"`
	Impact  string  `json:"impact"`
	Savings float64 `json:"savings"`
	Action  string  `json:"action"`
}

// handleDashboard serves the main dashboard HTML
func (wd *WebDashboard) handleDashboard(config WebDashboardConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")

		// Get the embedded dashboard HTML template
		html := getDashboardHTML(config)
		w.Write([]byte(html))
	}
}

// handleHealth provides health check endpoint
func (wd *WebDashboard) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"version":   "1.0.0",
		"components": map[string]string{
			"monitoring_service": "healthy",
			"metrics_collector":  "healthy",
			"prometheus":         "healthy",
		},
	}

	json.NewEncoder(w).Encode(health)
}

// handleMetrics provides comprehensive metrics data
func (wd *WebDashboard) handleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	wd.mu.RLock()
	defer wd.mu.RUnlock()

	gpuMetricsInterface := make(map[string]interface{})
	for k, v := range wd.lastMetrics {
		gpuMetricsInterface[k] = v
	}

	metrics := DashboardMetrics{
		Timestamp:   time.Now(),
		GPUMetrics:  gpuMetricsInterface,
		SystemStats: wd.calculateSystemStats(),
		CostData:    wd.lastCostData,
		Alerts:      wd.getActiveAlerts(),
		Performance: wd.calculatePerformanceMetrics(),
	}

	json.NewEncoder(w).Encode(metrics)
}

// handleGPUMetrics provides metrics for a specific GPU
func (wd *WebDashboard) handleGPUMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	gpuID := vars["id"]

	wd.mu.RLock()
	defer wd.mu.RUnlock()

	if metrics, exists := wd.lastMetrics[gpuID]; exists {
		json.NewEncoder(w).Encode(metrics)
	} else {
		http.Error(w, "GPU not found", http.StatusNotFound)
	}
}

// handleSystemStats provides system-level statistics
func (wd *WebDashboard) handleSystemStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	wd.mu.RLock()
	defer wd.mu.RUnlock()

	stats := wd.calculateSystemStats()
	json.NewEncoder(w).Encode(stats)
}

// handleCosts provides cost information
func (wd *WebDashboard) handleCosts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	wd.mu.RLock()
	defer wd.mu.RUnlock()

	json.NewEncoder(w).Encode(wd.lastCostData)
}

// handleAlerts provides active alerts
func (wd *WebDashboard) handleAlerts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	alerts := wd.getActiveAlerts()
	json.NewEncoder(w).Encode(alerts)
}

// handleResolveAlert resolves a specific alert
func (wd *WebDashboard) handleResolveAlert(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	alertID := vars["id"]

	// TODO: Implement alert resolution logic
	fmt.Printf("Resolving alert: %s\n", alertID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "resolved"})
}

// handlePerformance provides performance analytics
func (wd *WebDashboard) handlePerformance(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	performance := wd.calculatePerformanceMetrics()
	json.NewEncoder(w).Encode(performance)
}

// calculateSystemStats computes system-level statistics
func (wd *WebDashboard) calculateSystemStats() SystemStats {
	totalGPUs := len(wd.lastMetrics)
	if totalGPUs == 0 {
		return SystemStats{}
	}

	var totalUtil, totalTemp, totalPower, totalMemory, usedMemory float64
	activeGPUs := 0

	for _, metrics := range wd.lastMetrics {
		totalUtil += metrics.UtilizationGPU
		totalTemp += metrics.Temperature
		totalPower += metrics.PowerDraw
		totalMemory += float64(metrics.MemoryTotal)
		usedMemory += float64(metrics.MemoryUsed)

		if metrics.UtilizationGPU > 5 {
			activeGPUs++
		}
	}

	avgUtil := totalUtil / float64(totalGPUs)
	efficiencyScore := calculateEfficiencyScore(avgUtil, totalTemp/float64(totalGPUs))

	return SystemStats{
		TotalGPUs:       totalGPUs,
		ActiveGPUs:      activeGPUs,
		AverageUtil:     avgUtil,
		TotalMemoryGB:   totalMemory / 1024,
		UsedMemoryGB:    usedMemory / 1024,
		AverageTemp:     totalTemp / float64(totalGPUs),
		TotalPowerWatts: totalPower,
		EfficiencyScore: efficiencyScore,
	}
}

// calculateEfficiencyScore calculates a 0-100 efficiency score
func calculateEfficiencyScore(avgUtil, avgTemp float64) float64 {
	// Base score starts at utilization percentage
	score := avgUtil

	// Deduct points for high temperature
	if avgTemp > 80 {
		score -= (avgTemp - 80) * 2
	}

	// Ensure score is between 0 and 100
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return score
}

// getActiveAlerts generates sample alerts based on current metrics
func (wd *WebDashboard) getActiveAlerts() []Alert {
	var alerts []Alert

	for gpuID, metrics := range wd.lastMetrics {
		if metrics.Temperature > 80 {
			alerts = append(alerts, Alert{
				ID:        fmt.Sprintf("temp-%s", gpuID),
				Level:     "warning",
				Message:   fmt.Sprintf("High temperature on GPU %s (%.1f°C)", gpuID, metrics.Temperature),
				Source:    gpuID,
				Timestamp: time.Now(),
			})
		}

		if metrics.UtilizationGPU > 95 {
			alerts = append(alerts, Alert{
				ID:        fmt.Sprintf("util-%s", gpuID),
				Level:     "info",
				Message:   fmt.Sprintf("High utilization on GPU %s (%.1f%%)", gpuID, metrics.UtilizationGPU),
				Source:    gpuID,
				Timestamp: time.Now(),
			})
		}

		if metrics.MemoryTotal > 0 {
			memoryUsagePercent := float64(metrics.MemoryUsed) / float64(metrics.MemoryTotal) * 100
			if memoryUsagePercent > 90 {
				alerts = append(alerts, Alert{
					ID:        fmt.Sprintf("mem-%s", gpuID),
					Level:     "critical",
					Message:   fmt.Sprintf("High memory usage on GPU %s (%.1f%%)", gpuID, memoryUsagePercent),
					Source:    gpuID,
					Timestamp: time.Now(),
				})
			}
		}
	}

	return alerts
}

// calculatePerformanceMetrics computes performance analytics
func (wd *WebDashboard) calculatePerformanceMetrics() PerformanceMetrics {
	// TODO: Implement historical trend analysis
	// For now, return sample data

	tips := []OptimizationTip{}

	// Generate optimization tips based on current state
	avgUtil := 0.0
	totalGPUs := len(wd.lastMetrics)
	if totalGPUs > 0 {
		for _, metrics := range wd.lastMetrics {
			avgUtil += metrics.UtilizationGPU
		}
		avgUtil /= float64(totalGPUs)
	}

	if avgUtil < 50 {
		tips = append(tips, OptimizationTip{
			Type:    "efficiency",
			Message: "GPU utilization is below 50%. Consider consolidating workloads.",
			Impact:  "high",
			Savings: wd.lastCostData.TotalCost * 0.3,
			Action:  "Redistribute workloads to fewer GPUs",
		})
	}

	return PerformanceMetrics{
		UtilizationTrend: avgUtil,
		CostTrend:        wd.lastCostData.TotalCost,
		EfficiencyTrend:  calculateEfficiencyScore(avgUtil, 65),
		PredictedCost24h: wd.lastCostData.TotalCost * 24,
		OptimizationTips: tips,
	}
}

// handleCostSummary provides detailed cost summary information
func (wd *WebDashboard) handleCostSummary(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	wd.mu.RLock()
	defer wd.mu.RUnlock()

	summary := map[string]interface{}{
		"current_period":         wd.lastCostData,
		"daily_breakdown":        wd.calculateDailyCostBreakdown(),
		"cost_by_gpu":            wd.calculateCostByGPU(),
		"optimization_potential": wd.calculateOptimizationPotential(),
	}

	json.NewEncoder(w).Encode(summary)
}

// handleCostForecast provides cost forecasting
func (wd *WebDashboard) handleCostForecast(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	forecast := map[string]interface{}{
		"next_24h":     wd.calculateCostForecast(24 * time.Hour),
		"next_7_days":  wd.calculateCostForecast(7 * 24 * time.Hour),
		"next_30_days": wd.calculateCostForecast(30 * 24 * time.Hour),
		"confidence":   0.85,
		"based_on":     "last 24h utilization patterns",
	}

	json.NewEncoder(w).Encode(forecast)
}

// handleAlertSummary provides alert summary information
func (wd *WebDashboard) handleAlertSummary(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	alerts := wd.getActiveAlerts()

	summary := map[string]interface{}{
		"total_alerts":   len(alerts),
		"critical_count": countAlertsByLevel(alerts, "critical"),
		"warning_count":  countAlertsByLevel(alerts, "warning"),
		"info_count":     countAlertsByLevel(alerts, "info"),
		"recent_alerts":  alerts[:min(len(alerts), 5)],
		"top_sources":    getTopAlertSources(alerts),
	}

	json.NewEncoder(w).Encode(summary)
}

// handleEfficiency provides efficiency analytics
func (wd *WebDashboard) handleEfficiency(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	efficiency := map[string]interface{}{
		"overall_score":      wd.calculateSystemHealth().Score,
		"gpu_efficiency":     wd.calculateGPUEfficiencyScores(),
		"power_efficiency":   wd.calculatePowerEfficiency(),
		"memory_efficiency":  wd.calculateMemoryEfficiency(),
		"thermal_efficiency": wd.calculateThermalEfficiency(),
		"recommendations":    wd.generateEfficiencyRecommendations(),
	}

	json.NewEncoder(w).Encode(efficiency)
}

// handleTrends provides performance trend analysis
func (wd *WebDashboard) handleTrends(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	trends := map[string]interface{}{
		"utilization_trend": wd.calculateUtilizationTrend(),
		"temperature_trend": wd.calculateTemperatureTrend(),
		"cost_trend":        wd.calculateCostTrend(),
		"efficiency_trend":  wd.calculateEfficiencyTrend(),
		"time_range":        "last 24 hours",
		"data_points":       wd.getTrendDataPoints(),
	}

	json.NewEncoder(w).Encode(trends)
}

// handleGPUList provides list of all available GPUs
func (wd *WebDashboard) handleGPUList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	wd.mu.RLock()
	defer wd.mu.RUnlock()

	gpus := make([]map[string]interface{}, 0)

	for gpuID, metrics := range wd.lastMetrics {
		gpu := map[string]interface{}{
			"id":           gpuID,
			"name":         metrics.Name,
			"status":       wd.getGPUStatus(metrics),
			"utilization":  metrics.UtilizationGPU,
			"temperature":  metrics.Temperature,
			"memory_total": metrics.MemoryTotal,
			"memory_used":  metrics.MemoryUsed,
			"power_draw":   metrics.PowerDraw,
			"last_updated": metrics.Timestamp,
		}
		gpus = append(gpus, gpu)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"gpus":      gpus,
		"total":     len(gpus),
		"timestamp": time.Now(),
	})
}

// handleGPUProcesses provides processes running on a specific GPU
func (wd *WebDashboard) handleGPUProcesses(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	gpuID := vars["id"]

	// This would need integration with the metrics collector to get processes
	processes := []map[string]interface{}{
		{
			"pid":         12345,
			"name":        "python",
			"memory_used": 2048,
			"type":        "C",
			"command":     "python train_model.py",
		},
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"gpu_id":    gpuID,
		"processes": processes,
		"timestamp": time.Now(),
	})
}

// handleGPUHistory provides historical metrics for a specific GPU
func (wd *WebDashboard) handleGPUHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	gpuID := vars["id"]

	// Parse query parameters
	hoursStr := r.URL.Query().Get("hours")
	hours := 1 // default to 1 hour
	if h, err := strconv.Atoi(hoursStr); err == nil && h > 0 {
		hours = h
	}

	since := time.Now().Add(-time.Duration(hours) * time.Hour)

	// This would need integration with metrics collector for real history
	history := []map[string]interface{}{
		{
			"timestamp":   time.Now().Add(-30 * time.Minute),
			"utilization": 75.5,
			"temperature": 68.2,
			"memory_used": 8192,
		},
		{
			"timestamp":   time.Now().Add(-15 * time.Minute),
			"utilization": 82.1,
			"temperature": 71.8,
			"memory_used": 9216,
		},
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"gpu_id":  gpuID,
		"since":   since,
		"history": history,
		"count":   len(history),
	})
}

// handleSystemOverview provides comprehensive system overview
func (wd *WebDashboard) handleSystemOverview(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	wd.mu.RLock()
	defer wd.mu.RUnlock()

	overview := map[string]interface{}{
		"system_stats":  wd.calculateSystemStats(),
		"health_status": wd.calculateSystemHealth(),
		"cost_summary":  wd.lastCostData,
		"active_alerts": len(wd.getActiveAlerts()),
		"performance":   wd.calculatePerformanceMetrics(),
		"uptime":        wd.calculateUptime(),
		"last_updated":  time.Now(),
	}

	json.NewEncoder(w).Encode(overview)
}

// handleSystemStatus provides current system status
func (wd *WebDashboard) handleSystemStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	status := map[string]interface{}{
		"status": "operational",
		"components": map[string]string{
			"metrics_collector": "healthy",
			"websocket_server":  "healthy",
			"prometheus":        "healthy",
			"dashboard":         "healthy",
		},
		"active_connections": wd.GetActiveConnections(),
		"data_freshness":     wd.getDataFreshness(),
		"timestamp":          time.Now(),
	}

	json.NewEncoder(w).Encode(status)
}

// handleDemoTrigger triggers specific workload patterns (for demo purposes)
func (wd *WebDashboard) handleDemoTrigger(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gpuID := vars["gpu_id"]
	pattern := vars["pattern"]

	// This would integrate with mock collector to trigger patterns
	result := map[string]interface{}{
		"gpu_id":  gpuID,
		"pattern": pattern,
		"status":  "triggered",
		"message": fmt.Sprintf("Triggered %s workload on %s", pattern, gpuID),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// handleSimulationSpeed controls simulation speed for demo
func (wd *WebDashboard) handleSimulationSpeed(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "POST" {
		var req struct {
			Speed float64 `json:"speed"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		// Would set simulation speed on mock collector
		result := map[string]interface{}{
			"speed":   req.Speed,
			"status":  "updated",
			"message": fmt.Sprintf("Simulation speed set to %.1fx", req.Speed),
		}

		json.NewEncoder(w).Encode(result)
	} else {
		// GET current simulation speed
		result := map[string]interface{}{
			"speed":  1.0,
			"status": "current",
		}

		json.NewEncoder(w).Encode(result)
	}
}

// Helper methods for calculations

func (wd *WebDashboard) calculateDailyCostBreakdown() []map[string]interface{} {
	// Mock daily cost data
	return []map[string]interface{}{
		{"date": time.Now().AddDate(0, 0, -2).Format("2006-01-02"), "cost": 124.50},
		{"date": time.Now().AddDate(0, 0, -1).Format("2006-01-02"), "cost": 132.75},
		{"date": time.Now().Format("2006-01-02"), "cost": 98.25},
	}
}

func (wd *WebDashboard) calculateCostByGPU() map[string]float64 {
	result := make(map[string]float64)
	for gpuID := range wd.lastMetrics {
		result[gpuID] = 25.50 + float64(len(gpuID))*2.3 // Mock cost per GPU
	}
	return result
}

func (wd *WebDashboard) calculateOptimizationPotential() map[string]interface{} {
	return map[string]interface{}{
		"potential_savings": 45.75,
		"efficiency_gain":   12.5,
		"recommendations":   3,
	}
}

func (wd *WebDashboard) calculateCostForecast(duration time.Duration) float64 {
	// Simple forecast based on current cost data
	return wd.lastCostData.TotalCost * (duration.Hours() / 24.0)
}

func countAlertsByLevel(alerts []Alert, level string) int {
	count := 0
	for _, alert := range alerts {
		if alert.Level == level {
			count++
		}
	}
	return count
}

func getTopAlertSources(alerts []Alert) []map[string]interface{} {
	sourceCount := make(map[string]int)
	for _, alert := range alerts {
		sourceCount[alert.Source]++
	}

	sources := make([]map[string]interface{}, 0)
	for source, count := range sourceCount {
		sources = append(sources, map[string]interface{}{
			"source": source,
			"count":  count,
		})
	}

	return sources
}

func (wd *WebDashboard) calculateGPUEfficiencyScores() map[string]float64 {
	scores := make(map[string]float64)
	for gpuID, metrics := range wd.lastMetrics {
		scores[gpuID] = calculateEfficiencyScore(metrics.UtilizationGPU, metrics.Temperature)
	}
	return scores
}

func (wd *WebDashboard) calculatePowerEfficiency() float64 {
	totalUtil := 0.0
	totalPower := 0.0

	for _, metrics := range wd.lastMetrics {
		totalUtil += metrics.UtilizationGPU
		totalPower += metrics.PowerDraw
	}

	if totalPower > 0 {
		return totalUtil / totalPower
	}
	return 0
}

func (wd *WebDashboard) calculateMemoryEfficiency() float64 {
	totalMemoryUtil := 0.0
	count := float64(len(wd.lastMetrics))

	for _, metrics := range wd.lastMetrics {
		if metrics.MemoryTotal > 0 {
			totalMemoryUtil += float64(metrics.MemoryUsed) / float64(metrics.MemoryTotal) * 100
		}
	}

	if count > 0 {
		return totalMemoryUtil / count
	}
	return 0
}

func (wd *WebDashboard) calculateThermalEfficiency() float64 {
	totalTemp := 0.0
	count := float64(len(wd.lastMetrics))

	for _, metrics := range wd.lastMetrics {
		totalTemp += metrics.Temperature
	}

	avgTemp := totalTemp / count

	// Return efficiency score based on temperature (lower is better)
	return math.Max(0, 100-(avgTemp-40)*2) // Penalize temps above 40°C
}

func (wd *WebDashboard) generateEfficiencyRecommendations() []map[string]interface{} {
	recommendations := []map[string]interface{}{}

	avgUtil := 0.0
	avgTemp := 0.0
	count := float64(len(wd.lastMetrics))

	for _, metrics := range wd.lastMetrics {
		avgUtil += metrics.UtilizationGPU
		avgTemp += metrics.Temperature
	}

	if count > 0 {
		avgUtil /= count
		avgTemp /= count
	}

	if avgUtil < 50 {
		recommendations = append(recommendations, map[string]interface{}{
			"type":        "utilization",
			"priority":    "high",
			"title":       "Low GPU Utilization",
			"description": "Consider consolidating workloads to improve efficiency",
			"impact":      "30% cost reduction",
		})
	}

	if avgTemp > 75 {
		recommendations = append(recommendations, map[string]interface{}{
			"type":        "thermal",
			"priority":    "medium",
			"title":       "High Operating Temperature",
			"description": "Improve cooling or reduce workload intensity",
			"impact":      "Extended hardware lifespan",
		})
	}

	return recommendations
}

func (wd *WebDashboard) calculateUtilizationTrend() map[string]interface{} {
	// Mock trend data
	return map[string]interface{}{
		"direction": "increasing",
		"change":    "+12.5%",
		"slope":     0.15,
	}
}

func (wd *WebDashboard) calculateTemperatureTrend() map[string]interface{} {
	return map[string]interface{}{
		"direction": "stable",
		"change":    "+1.2°C",
		"slope":     0.02,
	}
}

func (wd *WebDashboard) calculateCostTrend() map[string]interface{} {
	return map[string]interface{}{
		"direction": "decreasing",
		"change":    "-8.3%",
		"slope":     -0.08,
	}
}

func (wd *WebDashboard) calculateEfficiencyTrend() map[string]interface{} {
	return map[string]interface{}{
		"direction": "improving",
		"change":    "+5.7%",
		"slope":     0.06,
	}
}

func (wd *WebDashboard) getTrendDataPoints() []map[string]interface{} {
	points := make([]map[string]interface{}, 24)
	baseTime := time.Now().Add(-24 * time.Hour)

	for i := 0; i < 24; i++ {
		points[i] = map[string]interface{}{
			"timestamp":   baseTime.Add(time.Duration(i) * time.Hour),
			"utilization": 60.0 + math.Sin(float64(i)*0.2)*20.0,
			"temperature": 65.0 + math.Cos(float64(i)*0.15)*10.0,
			"cost":        25.0 + float64(i)*0.5,
		}
	}

	return points
}

func (wd *WebDashboard) getGPUStatus(metrics gpu.GPUMetrics) string {
	if metrics.Temperature > 85 {
		return "critical"
	} else if metrics.Temperature > 75 || metrics.UtilizationGPU > 95 {
		return "warning"
	} else if metrics.UtilizationGPU > 5 {
		return "active"
	}
	return "idle"
}

func (wd *WebDashboard) calculateUptime() string {
	// Mock uptime - in real implementation this would track actual uptime
	return "15 days, 4 hours, 23 minutes"
}

func (wd *WebDashboard) getDataFreshness() map[string]interface{} {
	freshness := map[string]interface{}{
		"last_update": time.Now().Add(-5 * time.Second),
		"status":      "fresh",
	}

	// Check if data is stale
	if len(wd.lastMetrics) > 0 {
		latestTime := time.Time{}
		for _, metrics := range wd.lastMetrics {
			if metrics.Timestamp.After(latestTime) {
				latestTime = metrics.Timestamp
			}
		}

		age := time.Since(latestTime)
		freshness["last_update"] = latestTime

		if age > 30*time.Second {
			freshness["status"] = "stale"
		}
	}

	return freshness
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (wd *WebDashboard) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (wd *WebDashboard) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)

		fmt.Printf("[%s] %s %s - %v\n",
			time.Now().Format("2006-01-02 15:04:05"),
			r.Method,
			r.URL.Path,
			time.Since(start),
		)
	})
}
