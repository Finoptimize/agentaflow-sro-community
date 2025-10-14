package observability

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// handleDashboard serves the main dashboard HTML
func (wd *WebDashboard) handleDashboard(config WebDashboardConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Serve the dashboard HTML template
		w.Header().Set("Content-Type", "text/html")

		dashboardHTML := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>` + config.Title + `</title>
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/date-fns@2.29.3/index.min.js"></script>
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.1.3/dist/css/bootstrap.min.css" rel="stylesheet">
    <link href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.0.0/css/all.min.css" rel="stylesheet">
    <style>
        :root {
            --primary-color: #007bff;
            --success-color: #28a745;
            --warning-color: #ffc107;
            --danger-color: #dc3545;
            --dark-bg: #1a1a1a;
            --card-bg: #2d2d2d;
            --text-light: #ffffff;
        }
        
        body {
            background-color: var(--dark-bg);
            color: var(--text-light);
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
        }
        
        .dashboard-card {
            background-color: var(--card-bg);
            border: 1px solid #444;
            border-radius: 8px;
            padding: 20px;
            margin-bottom: 20px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.3);
        }
        
        .metric-card {
            text-align: center;
            padding: 15px;
        }
        
        .metric-value {
            font-size: 2rem;
            font-weight: bold;
            margin-bottom: 5px;
        }
        
        .metric-label {
            color: #888;
            font-size: 0.9rem;
        }
        
        .status-indicator {
            width: 12px;
            height: 12px;
            border-radius: 50%;
            display: inline-block;
            margin-right: 8px;
        }
        
        .status-healthy { background-color: var(--success-color); }
        .status-warning { background-color: var(--warning-color); }
        .status-critical { background-color: var(--danger-color); }
        
        .chart-container {
            position: relative;
            height: 300px;
            margin: 20px 0;
        }
        
        .gpu-card {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            border-radius: 10px;
            padding: 15px;
            margin: 10px 0;
        }
        
        .alert-item {
            border-left: 4px solid;
            padding: 10px 15px;
            margin: 5px 0;
            border-radius: 0 4px 4px 0;
        }
        
        .alert-critical { border-color: var(--danger-color); background-color: rgba(220, 53, 69, 0.1); }
        .alert-warning { border-color: var(--warning-color); background-color: rgba(255, 193, 7, 0.1); }
        .alert-info { border-color: var(--primary-color); background-color: rgba(0, 123, 255, 0.1); }
        
        .navbar-brand {
            font-weight: bold;
            font-size: 1.5rem;
        }
        
        .refresh-indicator {
            animation: spin 2s linear infinite;
        }
        
        @keyframes spin {
            0% { transform: rotate(0deg); }
            100% { transform: rotate(360deg); }
        }
        
        .connection-status {
            position: fixed;
            top: 70px;
            right: 20px;
            padding: 8px 12px;
            border-radius: 4px;
            font-size: 0.8rem;
            z-index: 1000;
        }
        
        .connected { background-color: var(--success-color); }
        .disconnected { background-color: var(--danger-color); }
    </style>
</head>
<body>
    <nav class="navbar navbar-dark bg-dark">
        <div class="container-fluid">
            <span class="navbar-brand">
                <i class="fas fa-tachometer-alt me-2"></i>
                ` + config.Title + `
            </span>
            <div class="d-flex align-items-center">
                <span class="me-3">
                    <i id="refresh-icon" class="fas fa-sync-alt"></i>
                    <small class="text-muted ms-2" id="last-update">Never</small>
                </span>
                <div id="connection-status" class="connection-status disconnected">
                    <i class="fas fa-circle"></i> Connecting...
                </div>
            </div>
        </div>
    </nav>

    <div class="container-fluid mt-4">
        <!-- System Overview -->
        <div class="row mb-4">
            <div class="col-12">
                <div class="dashboard-card">
                    <div class="d-flex justify-content-between align-items-center mb-3">
                        <h4><i class="fas fa-server me-2"></i>System Overview</h4>
                        <div id="system-health" class="d-flex align-items-center">
                            <span class="status-indicator status-healthy"></span>
                            <span>System Healthy</span>
                        </div>
                    </div>
                    
                    <div class="row">
                        <div class="col-md-2">
                            <div class="metric-card">
                                <div class="metric-value text-primary" id="total-gpus">0</div>
                                <div class="metric-label">Total GPUs</div>
                            </div>
                        </div>
                        <div class="col-md-2">
                            <div class="metric-card">
                                <div class="metric-value text-success" id="active-gpus">0</div>
                                <div class="metric-label">Active GPUs</div>
                            </div>
                        </div>
                        <div class="col-md-2">
                            <div class="metric-card">
                                <div class="metric-value text-info" id="avg-utilization">0%</div>
                                <div class="metric-label">Avg Utilization</div>
                            </div>
                        </div>
                        <div class="col-md-2">
                            <div class="metric-card">
                                <div class="metric-value text-warning" id="total-power">0W</div>
                                <div class="metric-label">Total Power</div>
                            </div>
                        </div>
                        <div class="col-md-2">
                            <div class="metric-card">
                                <div class="metric-value text-success" id="efficiency-score">100%</div>
                                <div class="metric-label">Efficiency</div>
                            </div>
                        </div>
                        <div class="col-md-2">
                            <div class="metric-card">
                                <div class="metric-value text-danger" id="total-cost">$0.00</div>
                                <div class="metric-label">Daily Cost</div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <div class="row">
            <!-- GPU Metrics Charts -->
            <div class="col-lg-8">
                <div class="dashboard-card">
                    <h5><i class="fas fa-chart-line me-2"></i>GPU Performance</h5>
                    <div class="chart-container">
                        <canvas id="gpu-chart"></canvas>
                    </div>
                </div>
                
                <div class="dashboard-card">
                    <h5><i class="fas fa-dollar-sign me-2"></i>Cost Analytics</h5>
                    <div class="chart-container">
                        <canvas id="cost-chart"></canvas>
                    </div>
                </div>
            </div>

            <!-- GPU Status & Alerts -->
            <div class="col-lg-4">
                <div class="dashboard-card">
                    <h5><i class="fas fa-microchip me-2"></i>GPU Status</h5>
                    <div id="gpu-list">
                        <!-- GPU cards will be populated here -->
                    </div>
                </div>

                <div class="dashboard-card">
                    <h5><i class="fas fa-exclamation-triangle me-2"></i>Active Alerts</h5>
                    <div id="alerts-list">
                        <div class="text-muted text-center py-3">No active alerts</div>
                    </div>
                </div>
            </div>
        </div>
    </div>

    <script>
        // Dashboard JavaScript
        class AgentaFlowDashboard {
            constructor() {
                this.ws = null;
                this.charts = {};
                this.lastUpdate = null;
                this.reconnectAttempts = 0;
                this.maxReconnectAttempts = 5;
                
                this.init();
            }

            init() {
                this.setupWebSocket();
                this.initCharts();
                this.startPeriodicUpdate();
            }

            setupWebSocket() {
                const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
                const wsUrl = wsProtocol + '//' + window.location.host + '/ws';
                
                this.ws = new WebSocket(wsUrl);
                
                this.ws.onopen = () => {
                    console.log('WebSocket connected');
                    this.updateConnectionStatus(true);
                    this.reconnectAttempts = 0;
                };
                
                this.ws.onmessage = (event) => {
                    try {
                        const data = JSON.parse(event.data);
                        this.handleMetricsUpdate(data);
                    } catch (e) {
                        console.error('Error parsing WebSocket message:', e);
                    }
                };
                
                this.ws.onclose = () => {
                    console.log('WebSocket disconnected');
                    this.updateConnectionStatus(false);
                    this.scheduleReconnect();
                };
                
                this.ws.onerror = (error) => {
                    console.error('WebSocket error:', error);
                };
            }

            scheduleReconnect() {
                if (this.reconnectAttempts < this.maxReconnectAttempts) {
                    this.reconnectAttempts++;
                    const delay = Math.min(1000 * Math.pow(2, this.reconnectAttempts), 30000);
                    setTimeout(() => this.setupWebSocket(), delay);
                }
            }

            updateConnectionStatus(connected) {
                const statusEl = document.getElementById('connection-status');
                if (connected) {
                    statusEl.className = 'connection-status connected';
                    statusEl.innerHTML = '<i class="fas fa-circle"></i> Connected';
                } else {
                    statusEl.className = 'connection-status disconnected';
                    statusEl.innerHTML = '<i class="fas fa-circle"></i> Disconnected';
                }
            }

            initCharts() {
                // GPU Performance Chart
                const gpuCtx = document.getElementById('gpu-chart').getContext('2d');
                this.charts.gpu = new Chart(gpuCtx, {
                    type: 'line',
                    data: {
                        labels: [],
                        datasets: [{
                            label: 'GPU Utilization (%)',
                            data: [],
                            borderColor: '#007bff',
                            backgroundColor: 'rgba(0, 123, 255, 0.1)',
                            tension: 0.4
                        }, {
                            label: 'Temperature (°C)',
                            data: [],
                            borderColor: '#dc3545',
                            backgroundColor: 'rgba(220, 53, 69, 0.1)',
                            yAxisID: 'y1',
                            tension: 0.4
                        }]
                    },
                    options: {
                        responsive: true,
                        maintainAspectRatio: false,
                        plugins: {
                            legend: {
                                labels: { color: '#ffffff' }
                            }
                        },
                        scales: {
                            x: {
                                ticks: { color: '#888' },
                                grid: { color: '#444' }
                            },
                            y: {
                                type: 'linear',
                                display: true,
                                position: 'left',
                                ticks: { color: '#888' },
                                grid: { color: '#444' }
                            },
                            y1: {
                                type: 'linear',
                                display: true,
                                position: 'right',
                                ticks: { color: '#888' },
                                grid: { drawOnChartArea: false }
                            }
                        }
                    }
                });

                // Cost Chart
                const costCtx = document.getElementById('cost-chart').getContext('2d');
                this.charts.cost = new Chart(costCtx, {
                    type: 'bar',
                    data: {
                        labels: [],
                        datasets: [{
                            label: 'Hourly Cost ($)',
                            data: [],
                            backgroundColor: 'rgba(255, 193, 7, 0.8)',
                            borderColor: '#ffc107',
                            borderWidth: 1
                        }]
                    },
                    options: {
                        responsive: true,
                        maintainAspectRatio: false,
                        plugins: {
                            legend: {
                                labels: { color: '#ffffff' }
                            }
                        },
                        scales: {
                            x: {
                                ticks: { color: '#888' },
                                grid: { color: '#444' }
                            },
                            y: {
                                ticks: { color: '#888' },
                                grid: { color: '#444' }
                            }
                        }
                    }
                });
            }

            async fetchMetrics() {
                try {
                    const response = await fetch('/api/v1/metrics');
                    const data = await response.json();
                    this.handleMetricsUpdate(data);
                } catch (error) {
                    console.error('Error fetching metrics:', error);
                }
            }

            handleMetricsUpdate(data) {
                this.lastUpdate = new Date();
                document.getElementById('last-update').textContent = 
                    this.lastUpdate.toLocaleTimeString();

                // Update system overview
                if (data.system_stats) {
                    this.updateSystemStats(data.system_stats);
                }

                // Update GPU list
                if (data.gpu_metrics) {
                    this.updateGPUList(data.gpu_metrics);
                    this.updateCharts(data.gpu_metrics);
                }

                // Update alerts
                if (data.alerts) {
                    this.updateAlerts(data.alerts);
                }

                // Update cost data
                if (data.cost_data) {
                    this.updateCostData(data.cost_data);
                }

                // Animate refresh icon
                const refreshIcon = document.getElementById('refresh-icon');
                refreshIcon.classList.add('refresh-indicator');
                setTimeout(() => refreshIcon.classList.remove('refresh-indicator'), 1000);
            }

            updateSystemStats(stats) {
                document.getElementById('total-gpus').textContent = stats.total_gpus || 0;
                document.getElementById('active-gpus').textContent = stats.active_gpus || 0;
                document.getElementById('avg-utilization').textContent = 
                    (stats.average_utilization || 0).toFixed(1) + '%';
                document.getElementById('total-power').textContent = 
                    (stats.total_power_watts || 0).toFixed(0) + 'W';
                document.getElementById('efficiency-score').textContent = 
                    (stats.efficiency_score || 0).toFixed(0) + '%';
                
                // Update system health indicator
                const healthEl = document.getElementById('system-health');
                const indicator = healthEl.querySelector('.status-indicator');
                const text = healthEl.querySelector('span:last-child');
                
                if (stats.efficiency_score > 80) {
                    indicator.className = 'status-indicator status-healthy';
                    text.textContent = 'System Healthy';
                } else if (stats.efficiency_score > 60) {
                    indicator.className = 'status-indicator status-warning';
                    text.textContent = 'System Warning';
                } else {
                    indicator.className = 'status-indicator status-critical';
                    text.textContent = 'System Critical';
                }
            }

            updateGPUList(gpuMetrics) {
                const container = document.getElementById('gpu-list');
                container.innerHTML = '';

                Object.entries(gpuMetrics).forEach(([gpuId, metrics]) => {
                    const gpuCard = document.createElement('div');
                    gpuCard.className = 'gpu-card mb-2';
                    
                    const tempColor = metrics.temperature > 80 ? 'text-danger' : 
                                     metrics.temperature > 70 ? 'text-warning' : 'text-success';
                    
                    gpuCard.innerHTML = ` + "`" + `
                        <div class="d-flex justify-content-between align-items-center">
                            <div>
                                <strong>${metrics.name || gpuId}</strong>
                                <br>
                                <small>ID: ${gpuId}</small>
                            </div>
                            <div class="text-end">
                                <div>${(metrics.utilization_gpu || 0).toFixed(1)}%</div>
                                <small class="${tempColor}">${(metrics.temperature || 0).toFixed(1)}°C</small>
                            </div>
                        </div>
                        <div class="mt-2">
                            <div class="progress mb-1" style="height: 6px;">
                                <div class="progress-bar" style="width: ${metrics.utilization_gpu || 0}%"></div>
                            </div>
                            <small class="text-light">
                                ${(metrics.memory_used || 0)}MB / ${(metrics.memory_total || 0)}MB
                            </small>
                        </div>
                    ` + "`" + `;
                    
                    container.appendChild(gpuCard);
                });
            }

            updateCharts(gpuMetrics) {
                // Update GPU performance chart with latest data
                const now = new Date().toLocaleTimeString();
                
                // Calculate average metrics across all GPUs
                const gpuArray = Object.values(gpuMetrics);
                const avgUtil = gpuArray.reduce((sum, gpu) => sum + (gpu.utilization_gpu || 0), 0) / gpuArray.length;
                const avgTemp = gpuArray.reduce((sum, gpu) => sum + (gpu.temperature || 0), 0) / gpuArray.length;

                // Update charts
                this.addDataPoint(this.charts.gpu, now, [avgUtil, avgTemp]);
            }

            updateAlerts(alerts) {
                const container = document.getElementById('alerts-list');
                
                if (!alerts || alerts.length === 0) {
                    container.innerHTML = '<div class="text-muted text-center py-3">No active alerts</div>';
                    return;
                }

                container.innerHTML = '';
                alerts.forEach(alert => {
                    const alertEl = document.createElement('div');
                    alertEl.className = ` + "`alert-item alert-${alert.level}`" + `;
                    alertEl.innerHTML = ` + "`" + `
                        <div class="d-flex justify-content-between">
                            <div>
                                <strong>${alert.message}</strong>
                                <br>
                                <small class="text-muted">${alert.source} • ${new Date(alert.timestamp).toLocaleTimeString()}</small>
                            </div>
                            <button class="btn btn-sm btn-outline-secondary" onclick="dashboard.resolveAlert('${alert.id}')">
                                <i class="fas fa-check"></i>
                            </button>
                        </div>
                    ` + "`" + `;
                    container.appendChild(alertEl);
                });
            }

            updateCostData(costData) {
                if (costData && costData.total_cost !== undefined) {
                    document.getElementById('total-cost').textContent = 
                        '$' + (costData.total_cost || 0).toFixed(2);
                }
            }

            addDataPoint(chart, label, data) {
                chart.data.labels.push(label);
                data.forEach((value, index) => {
                    chart.data.datasets[index].data.push(value);
                });

                // Keep only last 20 data points
                if (chart.data.labels.length > 20) {
                    chart.data.labels.shift();
                    chart.data.datasets.forEach(dataset => dataset.data.shift());
                }

                chart.update('none');
            }

            async resolveAlert(alertId) {
                try {
                    await fetch(` + "`/api/v1/alerts/${alertId}/resolve`" + `, { method: 'POST' });
                    this.fetchMetrics(); // Refresh to get updated alerts
                } catch (error) {
                    console.error('Error resolving alert:', error);
                }
            }

            startPeriodicUpdate() {
                // Fallback periodic update if WebSocket fails
                setInterval(() => {
                    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
                        this.fetchMetrics();
                    }
                }, ` + strconv.Itoa(config.RefreshInterval) + `);
            }
        }

        // Initialize dashboard when page loads
        let dashboard;
        document.addEventListener('DOMContentLoaded', () => {
            dashboard = new AgentaFlowDashboard();
        });
    </script>
</body>
</html>`

		w.Write([]byte(dashboardHTML))
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

	metrics := DashboardMetrics{
		Timestamp:   time.Now(),
		GPUMetrics:  wd.lastMetrics,
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
