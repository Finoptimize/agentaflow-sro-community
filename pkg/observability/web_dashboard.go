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

// CostSummary represents cost tracking data
type CostSummary struct {
	TotalCost     float64            `json:"total_cost"`
	GPUHours      float64            `json:"gpu_hours"`
	TokensUsed    int64              `json:"tokens_used"`
	Period        string             `json:"period"`
	StartTime     time.Time          `json:"start_time"`
	EndTime       time.Time          `json:"end_time"`
	CostBreakdown map[string]float64 `json:"cost_breakdown"`
}

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
	lastMetrics           map[string]gpu.GPUMetrics
	lastCostData          CostSummary
	systemHealth          SystemHealthStatus
	mu                    sync.RWMutex
	enableRealTimeUpdates bool
	theme                 string

	// Context for graceful shutdown
	ctx    context.Context
	cancel context.CancelFunc
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
	ctx, cancel := context.WithCancel(context.Background())

	wd := &WebDashboard{
		monitoringService:  monitoringService,
		metricsCollector:   metricsCollector,
		prometheusExporter: prometheusExporter,
		port:               config.Port,
		wsConnections:      make(map[*websocket.Conn]bool),
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

	wd.setupRouter(config)
	return wd
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
// Stop stops the web dashboard server
func (wd *WebDashboard) Stop() error {
	// Cancel background goroutines first
	wd.cancel()
	
	// Then shutdown the HTTP server
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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
	return wd.server.Shutdown(ctx)
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
		costData := wd.monitoringService.GetCostSummary(startTime, endTime)

		// Convert map to CostSummary struct
		wd.lastCostData = CostSummary{
			TotalCost:  getFloat64FromMap(costData, "total_cost"),
			GPUHours:   getFloat64FromMap(costData, "gpu_hours"),
			TokensUsed: getInt64FromMap(costData, "tokens_used"),
			Period:     "24h",
			StartTime:  startTime,
			EndTime:    endTime,
		}
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

	// Guard against division by zero when no GPUs are available
	if totalGPUs == 0 {
		wd.systemHealth = SystemHealthStatus{
			Status:    "warning",
			Score:     50,
			LastCheck: time.Now(),
			Issues:    []string{"No GPU metrics available"},
			Uptime:    "24h 15m", // TODO: Calculate actual uptime
		}
		return
	}

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

// Helper functions for type conversion
func getFloat64FromMap(m map[string]interface{}, key string) float64 {
	if val, ok := m[key]; ok {
		if f, ok := val.(float64); ok {
			return f
		}
		if i, ok := val.(int); ok {
			return float64(i)
		}
	}
	return 0
}

func getInt64FromMap(m map[string]interface{}, key string) int64 {
	if val, ok := m[key]; ok {
		if i, ok := val.(int64); ok {
			return i
		}
		if i, ok := val.(int); ok {
			return int64(i)
		}
		if f, ok := val.(float64); ok {
			return int64(f)
		}
	}
	return 0
}
