package observability

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	"github.com/Finoptimize/agentaflow-sro-community/pkg/gpu"
)

// WebDashboard represents the web-based monitoring dashboard
type WebDashboard struct {
	monitoringService  *MonitoringService
	metricsCollector   gpu.MetricsCollectorInterface
	prometheusExporter *PrometheusExporter
	server             *http.Server
	port               int

	// WebSocket management
	wsConnections  map[*websocket.Conn]bool
	wsWriteMutexes map[*websocket.Conn]*sync.Mutex
	wsUpgrader     websocket.Upgrader
	wsMutex        sync.RWMutex

	// Metrics caching
	lastMetrics  map[string]gpu.GPUMetrics
	metricsCache sync.RWMutex
	mu           sync.RWMutex
	lastCostData CostSummary

	// Configuration
	enableRealTimeUpdates bool
	theme                 string
	systemHealth          SystemHealthStatus

	// Lifecycle management
	ctx    context.Context
	cancel context.CancelFunc
}

// WebDashboardConfig holds configuration for the web dashboard
type WebDashboardConfig struct {
	Port                  int
	EnableRealTimeUpdates bool
	Theme                 string // "light" or "dark"
	Title                 string
	RefreshInterval       int
}

// SystemHealthStatus represents overall system health
type SystemHealthStatus struct {
	Status string  `json:"status"` // "healthy", "warning", "critical"
	Score  float64 `json:"score"`  // 0-100
}

// CostSummary represents cost calculation summary
type CostSummary struct {
	TotalCost    float64 `json:"total_cost"`
	Period       string  `json:"period"`
	Currency     string  `json:"currency"`
	GPUHours     float64 `json:"gpu_hours"`
	AvgCostPerHr float64 `json:"avg_cost_per_hr"`
}

// AlertInfo represents an alert/notification
type AlertInfo struct {
	ID           string    `json:"id"`
	Type         string    `json:"type"` // "info", "warning", "error"
	Title        string    `json:"title"`
	Message      string    `json:"message"`
	Timestamp    time.Time `json:"timestamp"`
	Acknowledged bool      `json:"acknowledged"`
}

// NewWebDashboard creates a new web dashboard instance
func NewWebDashboard(monitoringService *MonitoringService, metricsCollector gpu.MetricsCollectorInterface, prometheusExporter *PrometheusExporter, config WebDashboardConfig) *WebDashboard {
	ctx, cancel := context.WithCancel(context.Background())

	wd := &WebDashboard{
		monitoringService:  monitoringService,
		metricsCollector:   metricsCollector,
		prometheusExporter: prometheusExporter,
		port:               config.Port,
		wsConnections:      make(map[*websocket.Conn]bool),
		wsWriteMutexes:     make(map[*websocket.Conn]*sync.Mutex),
		wsUpgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				origin := r.Header.Get("Origin")

				// Allow empty origin (some clients don't send it)
				if origin == "" {
					return true
				}

				// Extract host from request
				requestHost := r.Host

				// Build allowed origins list
				allowedOrigins := []string{
					// Request host with both HTTP and HTTPS
					fmt.Sprintf("http://%s", requestHost),
					fmt.Sprintf("https://%s", requestHost),
					// Common localhost variations
					fmt.Sprintf("http://localhost:%d", config.Port),
					fmt.Sprintf("https://localhost:%d", config.Port),
					fmt.Sprintf("http://127.0.0.1:%d", config.Port),
					fmt.Sprintf("https://127.0.0.1:%d", config.Port),
				}

				// TODO: Make this configurable via WebDashboardConfig
				// Add any additional allowed origins from config here

				for _, allowed := range allowedOrigins {
					if origin == allowed {
						return true
					}
				}

				// Log rejected origins for security monitoring
				log.Printf("WebSocket connection rejected from origin: %s (request host: %s)", origin, requestHost)
				return false
			},
		},
		lastMetrics:           make(map[string]gpu.GPUMetrics),
		enableRealTimeUpdates: config.EnableRealTimeUpdates,
		theme:                 config.Theme,
		systemHealth:          SystemHealthStatus{Status: "healthy", Score: 100},
		ctx:                   ctx,
		cancel:                cancel,
	}

	// Set up HTTP server
	router := mux.NewRouter()
	wd.setupRoutes(router)

	wd.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", config.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	return wd
}

// Start starts the web dashboard server
func (wd *WebDashboard) Start() error {
	log.Printf("Starting web dashboard on port %d", wd.port)
	log.Printf("Dashboard will be accessible at: http://localhost:%d", wd.port)

	// Start background metrics collection
	go wd.startMetricsCollection()

	// Start WebSocket broadcast routine
	go wd.startWebSocketBroadcast()

	log.Printf("HTTP server starting on :%d...", wd.port)
	err := wd.server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Printf("Error starting web dashboard server: %v", err)
		return err
	}
	return err
}

// Stop stops the web dashboard server
func (wd *WebDashboard) Stop() error {
	// Cancel the context to stop background routines
	wd.cancel()

	// Create a timeout context for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown the HTTP server gracefully
	return wd.server.Shutdown(ctx)
}

// startMetricsCollection runs background metrics collection
func (wd *WebDashboard) startMetricsCollection() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			wd.updateMetrics()
		case <-wd.ctx.Done():
			return
		}
	}
}

// updateMetrics fetches and caches the latest metrics
func (wd *WebDashboard) updateMetrics() {
	if wd.metricsCollector == nil {
		return
	}

	wd.metricsCache.Lock()
	defer wd.metricsCache.Unlock()

	// Collect metrics for all GPUs
	// This is a simplified version - you'd need to implement actual GPU discovery
	gpuIDs := []string{"gpu-0", "gpu-1"} // Example GPU IDs

	for _, gpuID := range gpuIDs {
		latestMetrics := wd.metricsCollector.GetLatestMetrics()
		metrics, exists := latestMetrics[gpuID]
		if !exists {
			log.Printf("No metrics available for GPU %s", gpuID)
			continue
		}

		wd.lastMetrics[gpuID] = metrics
	}
}

// getLatestMetrics returns cached metrics data
func (wd *WebDashboard) getLatestMetrics() map[string]gpu.GPUMetrics {
	wd.metricsCache.RLock()
	defer wd.metricsCache.RUnlock()

	result := make(map[string]gpu.GPUMetrics)
	for k, v := range wd.lastMetrics {
		result[k] = v
	}

	return result
}

// startWebSocketBroadcast starts broadcasting updates to connected WebSocket clients
func (wd *WebDashboard) startWebSocketBroadcast() {
	if !wd.enableRealTimeUpdates {
		return
	}

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Only broadcast if there are connections to avoid race conditions
			if wd.GetActiveConnections() > 0 {
				wd.broadcastMetrics()
			}
		case <-wd.ctx.Done():
			return
		}
	}
}

// broadcastMetricsUpdate sends current metrics to all connected WebSocket clients
func (wd *WebDashboard) broadcastMetricsUpdate() {
	metrics := wd.getLatestMetrics()

	message := map[string]interface{}{
		"type":      "metrics_update",
		"timestamp": time.Now().Unix(),
		"data":      metrics,
	}

	// Note: broadcastToWebSockets method will be implemented in web_websocket.go
	// For now, we'll use a placeholder
	_ = message
}

// Helper functions for cost calculation and health monitoring
func (wd *WebDashboard) calculateCostSummary() CostSummary {
	// Simplified cost calculation
	// In a real implementation, this would integrate with cloud provider APIs
	metrics := wd.getLatestMetrics()

	totalGPUs := float64(len(metrics))
	hoursInDay := 24.0
	costPerGPUHour := 2.50 // Example cost

	totalCost := totalGPUs * hoursInDay * costPerGPUHour

	return CostSummary{
		TotalCost:    totalCost,
		Period:       "24h",
		Currency:     "USD",
		GPUHours:     totalGPUs * hoursInDay,
		AvgCostPerHr: costPerGPUHour,
	}
}

func (wd *WebDashboard) calculateSystemHealth() SystemHealthStatus {
	metrics := wd.getLatestMetrics()

	if len(metrics) == 0 {
		return SystemHealthStatus{Status: "warning", Score: 50}
	}

	totalUtilization := 0.0
	highTempCount := 0

	for _, metric := range metrics {
		totalUtilization += metric.UtilizationGPU
		if metric.Temperature > 80.0 { // Threshold for high temperature
			highTempCount++
		}
	}

	// Calculate average utilization (not currently used in health calculation)
	_ = totalUtilization / float64(len(metrics))
	healthScore := 100.0

	// Reduce score for high temperatures
	if highTempCount > 0 {
		healthScore -= float64(highTempCount) * 20.0
	}

	// Determine status based on score
	status := "healthy"
	if healthScore < 80 {
		status = "warning"
	}
	if healthScore < 50 {
		status = "critical"
	}

	return SystemHealthStatus{
		Status: status,
		Score:  healthScore,
	}
}

func (wd *WebDashboard) getRecentAlerts() []AlertInfo {
	// In a real implementation, this would fetch from a persistent store
	// For now, return some example alerts
	return []AlertInfo{
		{
			ID:        "alert-001",
			Type:      "warning",
			Title:     "High GPU Temperature",
			Message:   "GPU-0 temperature is above 85°C",
			Timestamp: time.Now().Add(-10 * time.Minute),
		},
		{
			ID:        "alert-002",
			Type:      "info",
			Title:     "Workload Completed",
			Message:   "Training job job-xyz completed successfully",
			Timestamp: time.Now().Add(-1 * time.Hour),
		},
	}
}

// setupRoutes configures the HTTP routes for the web dashboard
func (wd *WebDashboard) setupRoutes(router *mux.Router) {
	// Create a config to pass to handlers
	config := WebDashboardConfig{
		Port:                  wd.port,
		EnableRealTimeUpdates: wd.enableRealTimeUpdates,
		Theme:                 wd.theme,
		Title:                 "AgentaFlow GPU Monitoring Dashboard",
		RefreshInterval:       3000,
	}

	// Main dashboard route
	router.HandleFunc("/", wd.handleDashboard(config)).Methods("GET")

	// Health check endpoint
	router.HandleFunc("/health", wd.handleHealth).Methods("GET")

	// WebSocket endpoint for real-time updates
	router.HandleFunc("/ws", wd.handleWebSocket).Methods("GET")

	// API v1 routes
	api := router.PathPrefix("/api/v1").Subrouter()

	// Metrics endpoints
	api.HandleFunc("/metrics", wd.handleMetrics).Methods("GET")
	api.HandleFunc("/gpu/{id}/metrics", wd.handleGPUMetrics).Methods("GET")
	api.HandleFunc("/system/stats", wd.handleSystemStats).Methods("GET")

	// Cost endpoints
	api.HandleFunc("/costs", wd.handleCosts).Methods("GET")
	api.HandleFunc("/costs/summary", wd.handleCostSummary).Methods("GET")
	api.HandleFunc("/costs/forecast", wd.handleCostForecast).Methods("GET")

	// Alert endpoints
	api.HandleFunc("/alerts", wd.handleAlerts).Methods("GET")
	api.HandleFunc("/alerts/{id}/resolve", wd.handleResolveAlert).Methods("POST")
	api.HandleFunc("/alerts/summary", wd.handleAlertSummary).Methods("GET")

	// Performance endpoints
	api.HandleFunc("/performance", wd.handlePerformance).Methods("GET")
	api.HandleFunc("/performance/efficiency", wd.handleEfficiency).Methods("GET")
	api.HandleFunc("/performance/trends", wd.handleTrends).Methods("GET")

	// GPU management endpoints
	api.HandleFunc("/gpus", wd.handleGPUList).Methods("GET")
	api.HandleFunc("/gpu/{id}/processes", wd.handleGPUProcesses).Methods("GET")
	api.HandleFunc("/gpu/{id}/history", wd.handleGPUHistory).Methods("GET")

	// System endpoints
	api.HandleFunc("/system/overview", wd.handleSystemOverview).Methods("GET")
	api.HandleFunc("/system/status", wd.handleSystemStatus).Methods("GET")

	// Demo endpoints (for testing/simulation)
	api.HandleFunc("/demo/trigger/{gpu_id}/{pattern}", wd.handleDemoTrigger).Methods("POST")
	api.HandleFunc("/demo/simulation/speed", wd.handleSimulationSpeed).Methods("POST", "GET")

	// Static file serving for dashboard assets
	staticDir := "/static/"
	router.PathPrefix(staticDir).Handler(http.StripPrefix(staticDir, http.FileServer(http.Dir("./static/"))))

	// CORS middleware for development
	router.Use(wd.corsMiddleware)

	// Logging middleware
	router.Use(wd.loggingMiddleware)
}
