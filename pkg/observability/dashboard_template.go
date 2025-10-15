package observability

import (
	"fmt"
	"strings"
)

// getDashboardHTML returns the complete HTML for the web dashboard
func getDashboardHTML(config WebDashboardConfig) string {
	tmpl := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
    
    <!-- CSS Frameworks and Icons -->
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
    <link href="https://cdn.jsdelivr.net/npm/bootstrap-icons@1.11.0/font/bootstrap-icons.css" rel="stylesheet">
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700&display=swap" rel="stylesheet">
    
    <!-- Chart.js -->
    <script src="https://cdn.jsdelivr.net/npm/chart.js@3.9.1/dist/chart.min.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/date-fns@2.29.3/index.min.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/chartjs-adapter-date-fns@2.0.1/dist/chartjs-adapter-date-fns.bundle.min.js"></script>
    
    <style>
        :root {
            --bg-primary: #0f1419;
            --bg-secondary: #1e2329;
            --bg-tertiary: #2b313b;
            --text-primary: #ffffff;
            --text-secondary: #c5ccd6;
            --text-muted: #8a9199;
            --accent-blue: #1890ff;
            --accent-green: #52c41a;
            --accent-yellow: #faad14;
            --accent-red: #ff4d4f;
            --accent-purple: #722ed1;
            --border-color: #434a56;
            --shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.2);
            --shadow-lg: 0 10px 15px -3px rgba(0, 0, 0, 0.3);
        }

        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
            transition: all 0.3s ease;
        }

        body {
            font-family: 'Inter', -apple-system, BlinkMacSystemFont, sans-serif;
            background: var(--bg-primary);
            color: var(--text-primary);
            line-height: 1.6;
            overflow-x: hidden;
        }

        .navbar {
            background: var(--bg-secondary) !important;
            border-bottom: 1px solid var(--border-color);
            box-shadow: var(--shadow);
        }

        .navbar-brand {
            color: var(--text-primary) !important;
            font-weight: 600;
            font-size: 1.25rem;
        }

        .navbar-text {
            color: var(--text-secondary) !important;
        }

        .status-indicator {
            width: 8px;
            height: 8px;
            border-radius: 50%;
            display: inline-block;
            margin-right: 0.5rem;
        }

        .status-healthy { background-color: var(--accent-green); }
        .status-warning { background-color: var(--accent-yellow); }
        .status-critical { background-color: var(--accent-red); }

        .main-container {
            padding: 2rem 1rem;
            max-width: 1400px;
            margin: 0 auto;
        }

        .metrics-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
            gap: 1.5rem;
            margin-bottom: 2rem;
        }

        .metric-card {
            background: linear-gradient(135deg, var(--bg-secondary) 0%, var(--bg-tertiary) 100%);
            border: 1px solid var(--border-color);
            border-radius: 16px;
            padding: 1.75rem;
            transition: all 0.3s ease;
            box-shadow: var(--shadow);
            position: relative;
            overflow: hidden;
        }

        .metric-card::before {
            content: '';
            position: absolute;
            top: 0;
            left: 0;
            right: 0;
            height: 3px;
            background: linear-gradient(90deg, var(--accent-blue), var(--accent-purple), var(--accent-green));
            opacity: 0.8;
        }

        .metric-card:hover {
            transform: translateY(-2px);
            box-shadow: var(--shadow-lg);
        }

        .metric-header {
            display: flex;
            align-items: center;
            justify-content: space-between;
            margin-bottom: 1rem;
        }

        .metric-title {
            font-size: 0.875rem;
            color: var(--text-primary);
            text-transform: uppercase;
            letter-spacing: 0.05em;
            font-weight: 600;
            opacity: 0.9;
        }

        .metric-value {
            font-size: 2.25rem;
            font-weight: 700;
            color: var(--text-primary);
            line-height: 1;
            text-shadow: 0 2px 4px rgba(0,0,0,0.3);
        }

        .metric-change {
            font-size: 0.875rem;
            font-weight: 500;
            margin-top: 0.25rem;
        }

        .metric-change.positive { color: var(--accent-green); }
        .metric-change.negative { color: var(--accent-red); }
        .metric-change.neutral { color: var(--text-secondary); }

        .gpu-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
            gap: 1.5rem;
            margin-bottom: 2rem;
        }

        .gpu-card {
            background: linear-gradient(135deg, var(--bg-secondary) 0%, var(--bg-tertiary) 100%);
            border: 1px solid var(--border-color);
            border-radius: 16px;
            padding: 1.75rem;
            box-shadow: var(--shadow);
            transition: all 0.3s ease;
            position: relative;
            overflow: hidden;
        }

        .gpu-card::before {
            content: '';
            position: absolute;
            top: 0;
            left: 0;
            right: 0;
            height: 3px;
            background: linear-gradient(90deg, var(--accent-green), var(--accent-blue), var(--accent-purple));
            opacity: 0.8;
        }

        .gpu-card:hover {
            transform: translateY(-2px);
            box-shadow: var(--shadow-lg);
        }

        .gpu-header {
            display: flex;
            align-items: center;
            justify-content: between;
            margin-bottom: 1rem;
        }

        .gpu-name {
            font-weight: 600;
            color: var(--text-primary);
            font-size: 1rem;
        }

        .gpu-status {
            padding: 0.25rem 0.75rem;
            border-radius: 1rem;
            font-size: 0.75rem;
            font-weight: 500;
            text-transform: uppercase;
            letter-spacing: 0.05em;
        }

        .gpu-status.idle { background-color: rgba(108, 117, 125, 0.2); color: var(--text-muted); }
        .gpu-status.active { background-color: rgba(0, 123, 255, 0.2); color: var(--accent-blue); }
        .gpu-status.warning { background-color: rgba(255, 193, 7, 0.2); color: var(--accent-yellow); }
        .gpu-status.critical { background-color: rgba(220, 53, 69, 0.2); color: var(--accent-red); }

        .progress-group {
            margin-bottom: 1rem;
        }

        .progress-label {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 0.5rem;
            font-size: 0.875rem;
            color: var(--text-secondary);
        }

        .progress {
            height: 8px;
            background-color: var(--bg-tertiary);
            border-radius: 4px;
            overflow: hidden;
        }

        .progress-bar {
            transition: width 0.6s ease, background-color 0.3s ease;
        }

        .gpu-temp, .gpu-util, .gpu-memory {
            transition: color 0.2s ease;
            font-weight: 600;
        }

        .status-badge {
            padding: 0.25rem 0.6rem;
            border-radius: 6px;
            font-size: 0.75rem;
            font-weight: 600;
            text-transform: uppercase;
            letter-spacing: 0.5px;
            transition: all 0.2s ease;
        }

        .status-healthy { background: var(--accent-green); color: white; }
        .status-warning { background: var(--accent-yellow); color: #333; }
        .status-critical { background: var(--accent-red); color: white; }

        .chart-container {
            background: linear-gradient(135deg, #2c3e50 0%, #34495e 50%, #2c3e50 100%);
            border: 1px solid rgba(106, 135, 219, 0.3);
            border-radius: 16px;
            padding: 2rem;
            margin-bottom: 2rem;
            box-shadow: 0 8px 25px rgba(0, 0, 0, 0.2), 0 2px 8px rgba(106, 135, 219, 0.1);
            position: relative;
            overflow: hidden;
        }

        .chart-container::before {
            content: '';
            position: absolute;
            top: 0;
            left: 0;
            right: 0;
            height: 4px;
            background: linear-gradient(90deg, #e74c3c, #f39c12, #2ecc71, #3498db, #9b59b6);
            opacity: 0.9;
            border-radius: 16px 16px 0 0;
        }

        .chart-header {
            display: flex;
            align-items: center;
            justify-content: space-between;
            margin-bottom: 1.5rem;
        }

        .chart-title {
            font-size: 1.125rem;
            font-weight: 600;
            color: #ecf0f1;
            text-shadow: 0 1px 3px rgba(0, 0, 0, 0.3);
        }

        .chart-canvas {
            position: relative;
            height: 300px;
            background: transparent;
        }

        .chart-canvas canvas {
            background: rgba(255, 255, 255, 0.02) !important;
            border-radius: 8px;
        }

        .alerts-container {
            background: var(--bg-secondary);
            border: 1px solid var(--border-color);
            border-radius: 12px;
            padding: 1.5rem;
            margin-bottom: 2rem;
            box-shadow: var(--shadow);
        }

        .alert-item {
            display: flex;
            align-items: center;
            padding: 0.75rem;
            margin-bottom: 0.5rem;
            background: var(--bg-tertiary);
            border-radius: 8px;
            transition: all 0.3s ease;
        }

        .alert-item:hover {
            background: rgba(255, 255, 255, 0.05);
        }

        .alert-icon {
            width: 24px;
            height: 24px;
            border-radius: 50%;
            display: flex;
            align-items: center;
            justify-content: center;
            margin-right: 0.75rem;
            font-size: 0.75rem;
        }

        .alert-info .alert-icon { background-color: var(--accent-blue); }
        .alert-warning .alert-icon { background-color: var(--accent-yellow); }
        .alert-critical .alert-icon { background-color: var(--accent-red); }

        .alert-content {
            flex: 1;
        }

        .alert-title {
            font-weight: 500;
            color: var(--text-primary);
            font-size: 0.875rem;
        }

        .alert-message {
            color: var(--text-secondary);
            font-size: 0.75rem;
            margin-top: 0.25rem;
        }

        .alert-time {
            color: var(--text-muted);
            font-size: 0.75rem;
        }

        .connection-status {
            position: fixed;
            top: 1rem;
            right: 1rem;
            padding: 0.5rem 1rem;
            background: var(--bg-secondary);
            border: 1px solid var(--border-color);
            border-radius: 8px;
            font-size: 0.75rem;
            color: var(--text-secondary);
            z-index: 1000;
            display: flex;
            align-items: center;
            gap: 0.5rem;
            box-shadow: var(--shadow);
        }

        .connection-status.connected { border-color: var(--accent-green); }
        .connection-status.disconnected { border-color: var(--accent-red); }

        .loading {
            display: inline-block;
            width: 16px;
            height: 16px;
            border: 2px solid var(--bg-tertiary);
            border-radius: 50%;
            border-top-color: var(--accent-blue);
            animation: spin 1s ease-in-out infinite;
        }

        @keyframes spin {
            to { transform: rotate(360deg); }
        }

        .pulse {
            animation: pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
        }

        @keyframes pulse {
            0%, 100% { opacity: 1; }
            50% { opacity: 0.5; }
        }

        .fade-in {
            animation: fadeIn 0.5s ease-in-out;
        }

        @keyframes fadeIn {
            from { opacity: 0; transform: translateY(10px); }
            to { opacity: 1; transform: translateY(0); }
        }

        /* Responsive Design */
        @media (max-width: 768px) {
            .main-container {
                padding: 1rem 0.5rem;
            }
            
            .metrics-grid {
                grid-template-columns: 1fr;
                gap: 1rem;
            }
            
            .gpu-grid {
                grid-template-columns: 1fr;
            }
            
            .chart-canvas {
                height: 250px;
            }
        }
    </style>
</head>
<body>
    <!-- Navigation -->
    <nav class="navbar navbar-expand-lg">
        <div class="container-fluid">
            <a class="navbar-brand" href="#">
                <i class="bi bi-gpu-card me-2"></i>
                {{.Title}}
            </a>
            <div class="navbar-text">
                <span class="status-indicator status-healthy"></span>
                <span id="current-time"></span>
            </div>
        </div>
    </nav>

    <!-- Connection Status -->
    <div id="connection-status" class="connection-status disconnected">
        <div class="loading"></div>
        <span>Connecting...</span>
    </div>

    <!-- Main Container -->
    <div class="main-container">
        <!-- System Overview Metrics -->
        <div class="metrics-grid" id="system-metrics">
            <!-- Metrics cards will be populated by JavaScript -->
        </div>

        <!-- GPU Status Grid -->
        <div class="gpu-grid" id="gpu-grid">
            <!-- GPU cards will be populated by JavaScript -->
        </div>

        <!-- Performance Charts -->
        <div class="row">
            <div class="col-lg-8">
                <div class="chart-container">
                    <div class="chart-header">
                        <h3 class="chart-title">
                            <i class="bi bi-graph-up me-2"></i>
                            GPU Performance
                        </h3>
                        <div class="btn-group btn-group-sm" role="group">
                            <button type="button" class="btn btn-outline-secondary active" data-timerange="1h">1H</button>
                            <button type="button" class="btn btn-outline-secondary" data-timerange="6h">6H</button>
                            <button type="button" class="btn btn-outline-secondary" data-timerange="24h">24H</button>
                        </div>
                    </div>
                    <div class="chart-canvas">
                        <canvas id="performance-chart"></canvas>
                    </div>
                </div>
            </div>
            <div class="col-lg-4">
                <div class="chart-container">
                    <div class="chart-header">
                        <h3 class="chart-title">
                            <i class="bi bi-currency-dollar me-2"></i>
                            Cost Analytics
                        </h3>
                    </div>
                    <div class="chart-canvas">
                        <canvas id="cost-chart"></canvas>
                    </div>
                </div>
            </div>
        </div>

        <!-- Alerts -->
        <div class="alerts-container">
            <div class="chart-header">
                <h3 class="chart-title">
                    <i class="bi bi-exclamation-triangle me-2"></i>
                    Active Alerts
                    <span id="alert-count" class="badge bg-warning ms-2">0</span>
                </h3>
                <button class="btn btn-sm btn-outline-secondary" onclick="clearAllAlerts()">
                    Clear All
                </button>
            </div>
            <div id="alerts-list">
                <div class="text-muted text-center py-3">
                    <i class="bi bi-check-circle-fill me-2"></i>
                    No active alerts
                </div>
            </div>
        </div>
    </div>

    <!-- Bootstrap JS -->
    <script src="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/js/bootstrap.bundle.min.js"></script>

    <script>
        // Dashboard state
        let wsConnection = null;
        let performanceChart = null;
        let costChart = null;
        let isConnected = false;
        let metricsData = {};
        let reconnectAttempts = 0;
        const maxReconnectAttempts = 5;
        
        // Initialize dashboard
        document.addEventListener('DOMContentLoaded', function() {
            // Wait for Chart.js to load
            if (typeof Chart !== 'undefined') {
                initializeCharts();
            } else {
                console.log('Waiting for Chart.js to load...');
                setTimeout(function() {
                    if (typeof Chart !== 'undefined') {
                        initializeCharts();
                    } else {
                        console.error('Chart.js failed to load');
                    }
                }, 1000);
            }
            connectWebSocket();
            startDataRefresh();
            updateTime();
            setInterval(updateTime, 1000);
        });

        // WebSocket connection
        function connectWebSocket() {
            const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
            const wsUrl = protocol + '//' + window.location.host + '/ws';
            
            wsConnection = new WebSocket(wsUrl);
            
            wsConnection.onopen = function(event) {
                console.log('WebSocket connected');
                isConnected = true;
                reconnectAttempts = 0;
                updateConnectionStatus(true);
                
                // Request initial data
                wsConnection.send(JSON.stringify({
                    type: 'get_metrics',
                    timestamp: new Date().toISOString()
                }));
            };
            
            wsConnection.onmessage = function(event) {
                try {
                    const message = JSON.parse(event.data);
                    handleWebSocketMessage(message);
                } catch (error) {
                    console.error('Error parsing WebSocket message:', error);
                }
            };
            
            wsConnection.onclose = function(event) {
                console.log('WebSocket disconnected');
                isConnected = false;
                updateConnectionStatus(false);
                
                // Attempt to reconnect
                if (reconnectAttempts < maxReconnectAttempts) {
                    reconnectAttempts++;
                    setTimeout(connectWebSocket, Math.pow(2, reconnectAttempts) * 1000);
                }
            };
            
            wsConnection.onerror = function(error) {
                console.error('WebSocket error:', error);
            };
        }

        // Handle WebSocket messages
        function handleWebSocketMessage(message) {
            switch (message.type) {
                case 'metrics_update':
                    updateDashboard(message.data);
                    break;
                case 'alert':
                    addAlert(message.data);
                    break;
                case 'notification':
                    showNotification(message.data);
                    break;
                case 'pong':
                    // Handle pong response
                    break;
                default:
                    console.log('Unknown message type:', message.type);
            }
        }

        // Update connection status indicator
        function updateConnectionStatus(connected) {
            const statusEl = document.getElementById('connection-status');
            if (connected) {
                statusEl.className = 'connection-status connected';
                statusEl.innerHTML = '<i class="bi bi-wifi text-success"></i><span>Connected</span>';
            } else {
                statusEl.className = 'connection-status disconnected';
                statusEl.innerHTML = '<div class="loading"></div><span>Reconnecting...</span>';
            }
        }

        // Update current time display
        function updateTime() {
            const now = new Date();
            document.getElementById('current-time').textContent = now.toLocaleTimeString();
        }

        // Fetch data via REST API (fallback)
        async function fetchMetrics() {
            try {
                const response = await fetch('/api/v1/metrics');
                const data = await response.json();
                updateDashboard(data);
            } catch (error) {
                console.error('Error fetching metrics:', error);
            }
        }

        // Start periodic data refresh
        function startDataRefresh() {
            // Fetch data immediately
            fetchMetrics();
            
            // Refresh every 3 seconds regardless of WebSocket status for demo purposes
            setInterval(() => {
                fetchMetrics();
            }, 3000);
            
            // Also try WebSocket connection every 5 seconds if not connected
            setInterval(() => {
                if (!isConnected) {
                    connectWebSocket();
                }
            }, 5000);
        }

        // Update dashboard with new data
        function updateDashboard(data) {
            metricsData = data;
            updateSystemMetrics(data.system_stats);
            updateGPUCards(data.gpu_metrics);
            updateCharts(data);
            updateAlerts(data.alerts);
        }

        // Update system metrics cards
        function updateSystemMetrics(systemStats) {
            const metricsContainer = document.getElementById('system-metrics');
            
            const metrics = [
                {
                    title: 'Total GPUs',
                    value: systemStats.total_gpus || 0,
                    change: null,
                    icon: 'bi-gpu-card',
                    color: 'var(--accent-blue)'
                },
                {
                    title: 'Active GPUs',
                    value: systemStats.active_gpus || 0,
                    change: null,
                    icon: 'bi-power',
                    color: 'var(--accent-green)'
                },
                {
                    title: 'Avg Utilization',
                    value: (systemStats.average_utilization || 0).toFixed(1) + '%',
                    change: '+2.3%',
                    icon: 'bi-speedometer2',
                    color: 'var(--accent-purple)'
                },
                {
                    title: 'Efficiency Score',
                    value: (systemStats.efficiency_score || 0).toFixed(0) + '/100',
                    change: '+5.7%',
                    icon: 'bi-star-fill',
                    color: 'var(--accent-yellow)'
                },
                {
                    title: 'Total Power',
                    value: (systemStats.total_power_watts || 0).toFixed(0) + 'W',
                    change: '-1.2%',
                    icon: 'bi-lightning-fill',
                    color: 'var(--accent-red)'
                },
                {
                    title: 'Memory Usage',
                    value: ((systemStats.used_memory_gb || 0) / (systemStats.total_memory_gb || 1) * 100).toFixed(1) + '%',
                    change: '+0.8%',
                    icon: 'bi-memory',
                    color: 'var(--accent-blue)'
                }
            ];

            // Update existing metric cards or create new ones
            metrics.forEach((metric, index) => {
                const existingCard = metricsContainer.children[index];
                if (existingCard) {
                    updateExistingMetricCard(existingCard, metric);
                } else {
                    metricsContainer.insertAdjacentHTML('beforeend', createMetricCard(metric));
                }
            });
        }

        // Create metric card HTML
        function createMetricCard(metric) {
            const changeClass = metric.change ? 
                (metric.change.startsWith('+') ? 'positive' : 'negative') : 'neutral';
            
            return '<div class="metric-card fade-in">' +
                    '<div class="metric-header">' +
                        '<span class="metric-title">' + metric.title + '</span>' +
                        '<i class="' + metric.icon + '" style="color: ' + metric.color + '; font-size: 1.25rem;"></i>' +
                    '</div>' +
                    '<div class="metric-value">' + metric.value + '</div>' +
                    (metric.change ? '<div class="metric-change ' + changeClass + '">' + metric.change + '</div>' : '') +
                '</div>';
        }

        // Update GPU cards
        function updateGPUCards(gpuMetrics) {
            const gpuContainer = document.getElementById('gpu-grid');
            
            if (!gpuMetrics || Object.keys(gpuMetrics).length === 0) {
                gpuContainer.innerHTML = '<div class="col-12 text-center text-muted">No GPU data available</div>';
                return;
            }

            Object.entries(gpuMetrics).forEach(([gpuId, metrics]) => {
                var gpuCard = document.querySelector('[data-gpu-id="' + gpuId + '"]');
                if (!gpuCard) {
                    // Create new card if it doesn't exist
                    const cardHTML = createGPUCard(gpuId, metrics);
                    gpuContainer.insertAdjacentHTML('beforeend', cardHTML);
                    gpuCard = document.querySelector('[data-gpu-id="' + gpuId + '"]');
                } else {
                    // Update existing card smoothly
                    updateExistingGPUCard(gpuCard, metrics);
                }
            });
        }

        // Update existing metric card without replacing DOM
        function updateExistingMetricCard(cardElement, metric) {
            const valueElement = cardElement.querySelector('.metric-value');
            const changeElement = cardElement.querySelector('.metric-change');
            
            if (valueElement) valueElement.textContent = metric.value;
            if (changeElement && metric.change) changeElement.textContent = metric.change;
        }

        // Update existing GPU card without replacing DOM
        function updateExistingGPUCard(cardElement, metrics) {
            const utilization = metrics.utilization_gpu || 0;
            const temperature = metrics.temperature || 0;
            const memoryUsed = metrics.memory_used || 0;
            const memoryTotal = metrics.memory_total || 1;
            const memoryPercent = (memoryUsed / memoryTotal) * 100;
            
            // Update values smoothly
            const tempElement = cardElement.querySelector('.gpu-temp');
            const utilElement = cardElement.querySelector('.gpu-util');
            const memoryElement = cardElement.querySelector('.gpu-memory');
            
            if (tempElement) tempElement.textContent = Math.round(temperature) + '°C';
            if (utilElement) utilElement.textContent = Math.round(utilization) + '%';
            if (memoryElement) memoryElement.textContent = 
                (memoryUsed / 1024).toFixed(1) + '/' + (memoryTotal / 1024).toFixed(1) + ' GB';
            
            // Update progress bars smoothly
            const tempBar = cardElement.querySelector('.progress-bar.temp-bar');
            const utilBar = cardElement.querySelector('.progress-bar.util-bar');
            const memBar = cardElement.querySelector('.progress-bar.memory-bar');
            
            if (tempBar) tempBar.style.width = Math.min(temperature / 100 * 100, 100) + '%';
            if (utilBar) utilBar.style.width = utilization + '%';
            if (memBar) memBar.style.width = memoryPercent + '%';
            
            // Update status badge
            const status = getGPUStatus(temperature, utilization);
            const statusBadge = cardElement.querySelector('.status-badge');
            if (statusBadge) {
                statusBadge.className = 'status-badge status-' + status.class;
                statusBadge.textContent = status.text;
            }
        }

        // Create GPU card HTML
        function createGPUCard(gpuId, metrics) {
            const utilization = metrics.utilization_gpu || 0;
            const temperature = metrics.temperature || 0;
            const memoryUsed = metrics.memory_used || 0;
            const memoryTotal = metrics.memory_total || 1;
            const memoryPercent = (memoryUsed / memoryTotal) * 100;
            
            const status = getGPUStatus(temperature, utilization);
            
            return '<div class="gpu-card fade-in" data-gpu-id="' + gpuId + '">' +
                    '<div class="gpu-header">' +
                        '<div>' +
                            '<div class="gpu-name">' + (metrics.name || gpuId) + '</div>' +
                            '<small class="text-muted">' + gpuId + '</small>' +
                        '</div>' +
                        '<span class="status-badge status-' + status.class + '">' + status.text + '</span>' +
                    '</div>' +
                    
                    '<div class="progress-group">' +
                        '<div class="progress-label">' +
                            '<span>Utilization</span>' +
                            '<span class="gpu-util">' + utilization.toFixed(1) + '%</span>' +
                        '</div>' +
                        '<div class="progress">' +
                            '<div class="progress-bar util-bar" style="width: ' + utilization + '%; background-color: ' + getUtilizationColor(utilization) + ';"></div>' +
                        '</div>' +
                    '</div>' +
                    
                    '<div class="progress-group">' +
                        '<div class="progress-label">' +
                            '<span>Memory</span>' +
                            '<span class="gpu-memory">' + (memoryUsed / 1024).toFixed(1) + 'GB / ' + (memoryTotal / 1024).toFixed(1) + 'GB</span>' +
                        '</div>' +
                        '<div class="progress">' +
                            '<div class="progress-bar memory-bar" style="width: ' + memoryPercent + '%; background-color: ' + getMemoryColor(memoryPercent) + ';"></div>' +
                        '</div>' +
                    '</div>' +
                    
                    '<div class="progress-group">' +
                        '<div class="progress-label">' +
                            '<span>Temperature</span>' +
                            '<span class="gpu-temp">' + temperature.toFixed(1) + '°C</span>' +
                        '</div>' +
                        '<div class="progress">' +
                            '<div class="progress-bar temp-bar" style="width: ' + Math.min(temperature / 100 * 100, 100) + '%; background-color: ' + getTemperatureColor(temperature) + ';"></div>' +
                        '</div>' +
                    '</div>' +
                '</div>';
        }

        // Get GPU status based on metrics
        function getGPUStatus(temperature, utilization) {
            if (temperature > 85 || utilization > 95) return 'critical';
            if (temperature > 75 || utilization > 80) return 'warning';
            if (utilization > 5) return 'active';
            return 'idle';
        }

        // Color functions for progress bars
        function getUtilizationColor(util) {
            if (util > 90) return 'var(--accent-red)';
            if (util > 70) return 'var(--accent-yellow)';
            if (util > 30) return 'var(--accent-blue)';
            return 'var(--accent-green)';
        }

        function getMemoryColor(mem) {
            if (mem > 90) return 'var(--accent-red)';
            if (mem > 75) return 'var(--accent-yellow)';
            return 'var(--accent-blue)';
        }

        function getTemperatureColor(temp) {
            if (temp > 80) return 'var(--accent-red)';
            if (temp > 70) return 'var(--accent-yellow)';
            return 'var(--accent-green)';
        }

        // Initialize charts
        function initializeCharts() {
            if (typeof Chart === 'undefined') {
                console.warn('Chart.js not available - charts disabled');
                return;
            }
            
            // Performance chart
            const performanceCtx = document.getElementById('performance-chart').getContext('2d');
            // Add some initial sample data to make chart visible
            const now = new Date();
            const initialLabels = [];
            const initialUtilData = [];
            const initialTempData = [];
            
            for (let i = 9; i >= 0; i--) {
                const time = new Date(now.getTime() - i * 30000); // 30 second intervals
                initialLabels.push(time);
                initialUtilData.push(Math.random() * 60 + 20); // 20-80%
                initialTempData.push(Math.random() * 20 + 50); // 50-70°C
            }
            
            performanceChart = new Chart(performanceCtx, {
                type: 'line',
                data: {
                    labels: initialLabels,
                    datasets: [{
                        label: 'GPU Utilization %',
                        data: initialUtilData,
                        borderColor: '#1890ff',
                        backgroundColor: 'rgba(24, 144, 255, 0.1)',
                        tension: 0.4,
                        fill: true
                    }, {
                        label: 'Temperature °C',
                        data: initialTempData,
                        borderColor: '#ff4d4f',
                        backgroundColor: 'rgba(255, 77, 79, 0.1)',
                        tension: 0.4,
                        yAxisID: 'y1',
                        fill: true
                    }]
                },
                options: {
                    responsive: true,
                    maintainAspectRatio: false,
                    interaction: {
                        mode: 'index',
                        intersect: false
                    },
                    plugins: {
                        legend: {
                            labels: { 
                                color: '#ffffff',
                                font: { size: 12, weight: '500' }
                            }
                        }
                    },
                    scales: {
                        x: {
                            type: 'time',
                            time: { unit: 'minute' },
                            grid: { 
                                color: 'rgba(255, 255, 255, 0.1)',
                                drawBorder: false
                            },
                            ticks: { 
                                color: '#c5ccd6',
                                font: { size: 11 }
                            }
                        },
                        y: {
                            type: 'linear',
                            display: true,
                            position: 'left',
                            grid: { 
                                color: 'rgba(255, 255, 255, 0.1)',
                                drawBorder: false
                            },
                            ticks: { 
                                color: '#c5ccd6',
                                font: { size: 11 }
                            },
                            title: { 
                                display: true, 
                                text: 'Utilization %', 
                                color: '#c5ccd6',
                                font: { size: 12, weight: '500' }
                            }
                        },
                        y1: {
                            type: 'linear',
                            display: true,
                            position: 'right',
                            grid: { drawOnChartArea: false },
                            ticks: { 
                                color: '#c5ccd6',
                                font: { size: 11 }
                            },
                            title: { 
                                display: true, 
                                text: 'Temperature °C', 
                                color: '#c5ccd6',
                                font: { size: 12, weight: '500' }
                            }
                        }
                    }
                }
            });

            // Cost chart
            const costCtx = document.getElementById('cost-chart').getContext('2d');
            costChart = new Chart(costCtx, {
                type: 'doughnut',
                data: {
                    labels: ['GPU Hours', 'Storage', 'Network', 'Other'],
                    datasets: [{
                        data: [65, 20, 10, 5],
                        backgroundColor: [
                            '#1890ff',  // Blue
                            '#52c41a',  // Green  
                            '#faad14',  // Yellow
                            '#722ed1'   // Purple
                        ],
                        borderColor: '#2c3e50',
                        borderWidth: 2,
                        hoverOffset: 8
                    }]
                },
                options: {
                    responsive: true,
                    maintainAspectRatio: false,
                    cutout: '60%',
                    plugins: {
                        legend: {
                            position: 'bottom',
                            labels: { 
                                color: '#ffffff',
                                usePointStyle: true,
                                padding: 15,
                                font: { size: 12, weight: '500' }
                            }
                        },
                        tooltip: {
                            backgroundColor: 'rgba(44, 62, 80, 0.9)',
                            titleColor: '#ffffff',
                            bodyColor: '#c5ccd6',
                            borderColor: '#1890ff',
                            borderWidth: 1
                        }
                    }
                }
            });
        }

        // Update charts with new data
        function updateCharts(data) {
            if (data.gpu_metrics && performanceChart) {
                const now = new Date();
                const gpuMetrics = Object.values(data.gpu_metrics);
                
                if (gpuMetrics.length > 0) {
                    const avgUtilization = gpuMetrics.reduce((sum, gpu) => sum + (gpu.utilization_gpu || 0), 0) / gpuMetrics.length;
                    const avgTemperature = gpuMetrics.reduce((sum, gpu) => sum + (gpu.temperature || 0), 0) / gpuMetrics.length;
                    
                    // Add new data point
                    performanceChart.data.labels.push(now);
                    performanceChart.data.datasets[0].data.push(avgUtilization);
                    performanceChart.data.datasets[1].data.push(avgTemperature);
                    
                    // Keep only last 20 data points
                    if (performanceChart.data.labels.length > 20) {
                        performanceChart.data.labels.shift();
                        performanceChart.data.datasets[0].data.shift();
                        performanceChart.data.datasets[1].data.shift();
                    }
                    
                    performanceChart.update('quiet');
                }
            }
            
            // Update cost chart if cost data is available
            if (data.cost_data && costChart) {
                // Update cost breakdown (this would use real data in production)
                costChart.update('quiet');
            }
        }

        // Update alerts
        function updateAlerts(alerts) {
            const alertsList = document.getElementById('alerts-list');
            const alertCount = document.getElementById('alert-count');
            
            if (!alerts || alerts.length === 0) {
                alertsList.innerHTML = ` + "`" + `
                    <div class="text-muted text-center py-3">
                        <i class="bi bi-check-circle-fill me-2"></i>
                        No active alerts
                    </div>
                ` + "`" + `;
                alertCount.textContent = '0';
                alertCount.className = 'badge bg-success ms-2';
            } else {
                alertsList.innerHTML = alerts.map(alert => createAlertItem(alert)).join('');
                alertCount.textContent = alerts.length;
                alertCount.className = 'badge bg-warning ms-2';
            }
        }

        // Create alert item HTML
        function createAlertItem(alert) {
            const timeAgo = getTimeAgo(new Date(alert.timestamp));
            
            return ` + "`" + `
                <div class="alert-item alert-${alert.level} fade-in">
                    <div class="alert-icon">
                        <i class="bi ${getAlertIcon(alert.level)}"></i>
                    </div>
                    <div class="alert-content">
                        <div class="alert-title">${alert.message}</div>
                        <div class="alert-message">Source: ${alert.source}</div>
                    </div>
                    <div class="alert-time">${timeAgo}</div>
                </div>
            ` + "`" + `;
        }

        // Get alert icon based on level
        function getAlertIcon(level) {
            switch (level) {
                case 'critical': return 'bi-exclamation-triangle-fill';
                case 'warning': return 'bi-exclamation-circle-fill';
                case 'info': return 'bi-info-circle-fill';
                default: return 'bi-info-circle';
            }
        }

        // Get time ago string
        function getTimeAgo(date) {
            const seconds = Math.floor((new Date() - date) / 1000);
            
            if (seconds < 60) return seconds + 's ago';
            
            const minutes = Math.floor(seconds / 60);
            if (minutes < 60) return minutes + 'm ago';
            
            const hours = Math.floor(minutes / 60);
            if (hours < 24) return hours + 'h ago';
            
            const days = Math.floor(hours / 24);
            return days + 'd ago';
        }

        // Add new alert
        function addAlert(alert) {
            // Add to current alerts and refresh display
            if (!metricsData.alerts) metricsData.alerts = [];
            metricsData.alerts.unshift(alert);
            updateAlerts(metricsData.alerts);
            
            // Show browser notification if supported
            if (Notification.permission === 'granted') {
                new Notification('AgentaFlow Alert', {
                    body: alert.message,
                    icon: '/favicon.ico'
                });
            }
        }

        // Show notification
        function showNotification(notification) {
            // Create toast notification
            const toast = document.createElement('div');
            toast.className = 'toast show position-fixed top-0 end-0 m-3';
            toast.style.zIndex = '9999';
            toast.innerHTML = ` + "`" + `
                <div class="toast-header">
                    <strong class="me-auto">${notification.title}</strong>
                    <small class="text-muted">now</small>
                    <button type="button" class="btn-close" data-bs-dismiss="toast"></button>
                </div>
                <div class="toast-body">${notification.message}</div>
            ` + "`" + `;
            
            document.body.appendChild(toast);
            
            // Auto-remove after 5 seconds
            setTimeout(() => {
                toast.remove();
            }, 5000);
        }

        // Clear all alerts
        function clearAllAlerts() {
            if (metricsData.alerts) {
                metricsData.alerts = [];
                updateAlerts([]);
            }
        }

        // Request browser notification permission
        if (Notification.permission === 'default') {
            Notification.requestPermission();
        }

        // Handle page visibility changes
        document.addEventListener('visibilitychange', function() {
            if (!document.hidden && isConnected) {
                // Refresh data when page becomes visible
                wsConnection.send(JSON.stringify({
                    type: 'get_metrics',
                    timestamp: new Date().toISOString()
                }));
            }
        });

        // Handle window beforeunload
        window.addEventListener('beforeunload', function() {
            if (wsConnection) {
                wsConnection.close();
            }
        });

        // Time range selector for charts
        document.querySelectorAll('[data-timerange]').forEach(button => {
            button.addEventListener('click', function() {
                document.querySelectorAll('[data-timerange]').forEach(b => b.classList.remove('active'));
                this.classList.add('active');
                
                // In a real implementation, this would fetch historical data for the selected time range
                console.log('Selected time range:', this.dataset.timerange);
            });
        });
    </script>
</body>
</html>`

	// Replace template variables
	html := strings.ReplaceAll(tmpl, "{{.Title}}", config.Title)
	html = strings.ReplaceAll(html, "{{.RefreshInterval}}", fmt.Sprintf("%d", config.RefreshInterval))
	html = strings.ReplaceAll(html, "{{.Theme}}", config.Theme)

	return html
}
