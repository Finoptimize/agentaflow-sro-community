package observability

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

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
var dashboardTemplate = `<!DOCTYPE html>
<html><head><title>{{.Title}}</title></head>
<body><h1>{{.Title}}</h1><p>Dashboard placeholder</p></body></html>`

func (wd *WebDashboard) handleDashboard(config WebDashboardConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")

		// Replace template variables
		html := strings.ReplaceAll(dashboardTemplate, "{{.Title}}", config.Title)
		html = strings.ReplaceAll(html, "{{.RefreshInterval}}", strconv.Itoa(config.RefreshInterval))

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
