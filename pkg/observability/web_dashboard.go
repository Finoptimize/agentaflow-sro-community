package observability

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	"github.com/Finoptimize/agentaflow-sro-community/pkg/gpu"
)

// WebDashboard provides a web-based monitoring interface
type WebDashboard struct {
	monitoringService  *MonitoringService
	metricsCollector   *gpu.MetricsCollector
	prometheusExporter *PrometheusExporter

	// Web server configuration
	port   int
	server *http.Server
	router *mux.Router

	// WebSocket connections
	wsConnections map[*websocket.Conn]bool
	wsMutex       sync.RWMutex
	wsUpgrader    websocket.Upgrader

	// Dashboard state
	lastMetrics  map[string]gpu.GPUMetrics
	lastCostData CostSummary
	systemHealth SystemHealthStatus
	mu           sync.RWMutex
}

// WebDashboardConfig configures the web dashboard
type WebDashboardConfig struct {
	Port                  int    `json:"port"`
	Title                 string `json:"title"`
	RefreshInterval       int    `json:"refresh_interval_ms"`
	EnableRealTimeUpdates bool   `json:"enable_realtime_updates"`
	Theme                 string `json:"theme"`
}

// DashboardMetrics represents metrics data for the dashboard
type DashboardMetrics struct {
	Timestamp   time.Time                 `json:"timestamp"`
	GPUMetrics  map[string]gpu.GPUMetrics `json:"gpu_metrics"`
	SystemStats SystemStats               `json:"system_stats"`
	CostData    CostSummary               `json:"cost_data"`
	Alerts      []Alert                   `json:"alerts"`
	Performance PerformanceMetrics        `json:"performance"`
}

// SystemStats provides system-level statistics
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

// Alert represents a system alert
type Alert struct {
	ID         string     `json:"id"`
	Level      string     `json:"level"` // info, warning, error, critical
	Message    string     `json:"message"`
	Source     string     `json:"source"` // gpu_id, system, scheduler
	Timestamp  time.Time  `json:"timestamp"`
	Resolved   bool       `json:"resolved"`
	ResolvedAt *time.Time `json:"resolved_at,omitempty"`
}

// PerformanceMetrics provides performance analytics
type PerformanceMetrics struct {
	UtilizationTrend float64           `json:"utilization_trend"`
	CostTrend        float64           `json:"cost_trend"`
	EfficiencyTrend  float64           `json:"efficiency_trend"`
	PredictedCost24h float64           `json:"predicted_cost_24h"`
	OptimizationTips []OptimizationTip `json:"optimization_tips"`
}

// OptimizationTip suggests improvements
type OptimizationTip struct {
	Type    string  `json:"type"` // cost, performance, efficiency
	Message string  `json:"message"`
	Impact  string  `json:"impact"`  // low, medium, high
	Savings float64 `json:"savings"` // potential cost savings
	Action  string  `json:"action"`  // recommended action
}

// SystemHealthStatus represents overall system health
type SystemHealthStatus struct {
	Status    string    `json:"status"` // healthy, warning, critical
	Score     float64   `json:"score"`  // 0-100
	LastCheck time.Time `json:"last_check"`
	Issues    []string  `json:"issues"`
	Uptime    string    `json:"uptime"`
}

// NewWebDashboard creates a new web dashboard instance
func NewWebDashboard(config WebDashboardConfig, monitoringService *MonitoringService,
	metricsCollector *gpu.MetricsCollector, prometheusExporter *PrometheusExporter) *WebDashboard {

	if config.Port == 0 {
		config.Port = 8090
	}
	if config.RefreshInterval == 0 {
		config.RefreshInterval = 5000 // 5 seconds
	}
	if config.Title == "" {
		config.Title = "AgentaFlow GPU Monitoring Dashboard"
	}
	if config.Theme == "" {
		config.Theme = "dark"
	}

	wd := &WebDashboard{
		monitoringService:  monitoringService,
		metricsCollector:   metricsCollector,
		prometheusExporter: prometheusExporter,
		port:               config.Port,
		wsConnections:      make(map[*websocket.Conn]bool),
		wsUpgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow connections from any origin in development
			},
		},
		lastMetrics:  make(map[string]gpu.GPUMetrics),
		systemHealth: SystemHealthStatus{Status: "healthy", Score: 100},
	}

	wd.setupRouter(config)
	return wd
}

// setupRouter configures HTTP routes for the dashboard
func (wd *WebDashboard) setupRouter(config WebDashboardConfig) {
	wd.router = mux.NewRouter()

	// Static file serving
	wd.router.PathPrefix("/static/").Handler(http.StripPrefix("/static/",
		http.FileServer(http.Dir("./pkg/observability/web/static/"))))

	// Dashboard routes
	wd.router.HandleFunc("/", wd.handleDashboard(config)).Methods("GET")
	wd.router.HandleFunc("/health", wd.handleHealth).Methods("GET")

	// API routes
	api := wd.router.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/metrics", wd.handleMetrics).Methods("GET")
	api.HandleFunc("/gpu/{id}/metrics", wd.handleGPUMetrics).Methods("GET")
	api.HandleFunc("/system/stats", wd.handleSystemStats).Methods("GET")
	api.HandleFunc("/costs", wd.handleCosts).Methods("GET")
	api.HandleFunc("/alerts", wd.handleAlerts).Methods("GET")
	api.HandleFunc("/alerts/{id}/resolve", wd.handleResolveAlert).Methods("POST")
	api.HandleFunc("/performance", wd.handlePerformance).Methods("GET")

	// WebSocket endpoint
	wd.router.HandleFunc("/ws", wd.handleWebSocket).Methods("GET")

	wd.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", wd.port),
		Handler: wd.router,
	}
}

// Start starts the web dashboard server
func (wd *WebDashboard) Start() error {
	log.Printf("🌐 Starting AgentaFlow Web Dashboard on port %d", wd.port)
	log.Printf("📊 Dashboard available at: http://localhost:%d", wd.port)

	// Start background metrics collection
	go wd.startMetricsCollection()

	// Start WebSocket broadcast routine
	go wd.startWebSocketBroadcast()

	return wd.server.ListenAndServe()
}

// Stop stops the web dashboard server
func (wd *WebDashboard) Stop() error {
	return wd.server.Shutdown(nil)
}

// startMetricsCollection runs background metrics collection
func (wd *WebDashboard) startMetricsCollection() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		wd.updateMetrics()
	}
}

// updateMetrics updates internal metrics cache
func (wd *WebDashboard) updateMetrics() {
	wd.mu.Lock()
	defer wd.mu.Unlock()

	// Update GPU metrics
	if wd.metricsCollector != nil {
		overview := wd.metricsCollector.GetSystemOverview()
		for gpuID, metrics := range overview {
			if gpuMetrics, ok := metrics.(gpu.GPUMetrics); ok {
				wd.lastMetrics[gpuID] = gpuMetrics
			}
		}
	}

	// Update cost data
	if wd.monitoringService != nil {
		startTime := time.Now().Add(-24 * time.Hour)
		endTime := time.Now()
		wd.lastCostData = wd.monitoringService.GetCostSummary(startTime, endTime)
	}

	// Update system health
	wd.updateSystemHealth()
}

// updateSystemHealth calculates system health status
func (wd *WebDashboard) updateSystemHealth() {
	var totalScore float64 = 100
	var issues []string

	// Check GPU health
	highTempCount := 0
	highUtilCount := 0
	totalGPUs := len(wd.lastMetrics)

	for _, metrics := range wd.lastMetrics {
		if metrics.Temperature > 80 {
			highTempCount++
			issues = append(issues, fmt.Sprintf("GPU %s running hot (%.1f°C)", metrics.GPUID, metrics.Temperature))
		}
		if metrics.UtilizationGPU > 95 {
			highUtilCount++
		}
	}

	// Adjust score based on issues
	if highTempCount > 0 {
		totalScore -= float64(highTempCount) / float64(totalGPUs) * 20
	}

	// Determine status
	status := "healthy"
	if totalScore < 80 {
		status = "warning"
	}
	if totalScore < 60 {
		status = "critical"
	}

	wd.systemHealth = SystemHealthStatus{
		Status:    status,
		Score:     totalScore,
		LastCheck: time.Now(),
		Issues:    issues,
		Uptime:    "24h 15m", // TODO: Calculate actual uptime
	}
}
