# AgentaFlow Web Dashboard Demo

This demo showcases the comprehensive web-based monitoring dashboard for AgentaFlow SRO Community Edition, providing real-time GPU monitoring, cost tracking, and system analytics through a modern web interface.

## 🎯 Overview

The web dashboard provides:
- **Real-time GPU monitoring** with live charts and metrics
- **Interactive visualizations** using Chart.js
- **WebSocket-based updates** for instant data refresh
- **Cost analytics** with live tracking and forecasting
- **Alert management** with real-time notifications
- **System health monitoring** with efficiency scoring
- **Responsive design** optimized for desktop and mobile

## 🚀 Quick Start

### Run the Demo

```bash
# From the project root
cd examples/demo/web-dashboard
go run main.go
```

### Access the Dashboard

Once started, the demo provides:

- **🌐 Web Dashboard**: http://localhost:8090
- **📈 Prometheus Metrics**: http://localhost:8080/metrics

## 📊 Dashboard Features

### Real-time Monitoring
- Live GPU utilization, temperature, and power consumption
- Memory usage tracking across all GPUs
- System-wide performance metrics
- Automatic refresh every 3 seconds via WebSocket

### Interactive Charts
- **GPU Performance Chart**: Real-time utilization and temperature trends
- **Cost Analytics Chart**: Hourly cost tracking and forecasting
- **Historical Data**: Last 20 data points with smooth animations

### System Overview Cards
- Total and active GPU count
- Average utilization across cluster
- Total power consumption
- Efficiency scoring (0-100%)
- Daily cost tracking

### GPU Status Cards
- Individual GPU monitoring
- Real-time utilization bars
- Temperature color coding
- Memory usage indicators

### Alert Management
- Real-time alert notifications
- Temperature warnings (>75°C)
- High utilization alerts (>95%)
- Memory usage warnings (>90%)
- One-click alert resolution

## 🎮 Demo Data

The demo generates realistic simulated data:

### GPU Metrics (every 5 seconds)
- Simulated GPU utilization (0-100%)
- Temperature readings (40-85°C)
- Memory usage patterns
- Power consumption data

### Cost Tracking (every 10 seconds)
- Inference operation costs
- GPU hour tracking
- Token usage counting
- Real-time cost accumulation

### Alerts
- Automatic temperature alerts when >75°C
- High utilization notifications
- Memory usage warnings
- WebSocket broadcast to all clients

## 🌐 API Endpoints

The dashboard exposes REST API endpoints:

### System Information
- `GET /` - Main dashboard interface
- `GET /health` - System health check
- `GET /api/v1/metrics` - Complete metrics data
- `GET /api/v1/system/stats` - System statistics

### GPU Specific
- `GET /api/v1/gpu/{id}/metrics` - Individual GPU metrics
- `GET /api/v1/costs` - Cost information
- `GET /api/v1/performance` - Performance analytics

### Alert Management
- `GET /api/v1/alerts` - Active alerts
- `POST /api/v1/alerts/{id}/resolve` - Resolve alert

### Real-time Updates
- `GET /ws` - WebSocket endpoint for live updates

## 🔧 Configuration

The dashboard can be configured via `WebDashboardConfig`:

```go
dashboardConfig := observability.WebDashboardConfig{
    Port:                  8090,           // Web server port
    Title:                "Custom Title",   // Dashboard title
    RefreshInterval:       3000,           // Update interval (ms)
    EnableRealTimeUpdates: true,           // WebSocket updates
    Theme:                "dark",          // UI theme
}
```

## 📱 Responsive Design

The dashboard is fully responsive and works on:
- **Desktop computers** (1920x1080 and higher)
- **Tablets** (768px and up)
- **Mobile devices** (576px and up)
- **Large displays** (4K and ultra-wide monitors)

## 🎨 Themes

Currently supports:
- **Dark Theme** (default) - Optimized for monitoring environments
- Light theme support planned for future releases

## 🔌 Integration

### With Prometheus
The dashboard integrates seamlessly with Prometheus:
- Metrics automatically exported to `/metrics` endpoint
- Compatible with existing Prometheus configurations
- Works alongside Grafana for advanced analytics

### With GPU Metrics Collector
- Real-time data from `gpu.MetricsCollector`
- Automatic GPU discovery and monitoring
- Health status tracking and alerting

### With Cost Tracking
- Integration with `MonitoringService` cost tracking
- Real-time cost calculations
- Historical cost analysis and forecasting

## 🛠️ Development

### Adding Custom Metrics
```go
// Add custom metric to dashboard
dashboard.BroadcastSystemUpdate(map[string]interface{}{
    "custom_metric": "value",
    "timestamp": time.Now(),
})
```

### Custom Alerts
```go
// Send custom alert
alert := observability.Alert{
    ID:        "custom-alert-123",
    Level:     "warning",
    Message:   "Custom alert message",
    Source:    "system",
    Timestamp: time.Now(),
}
dashboard.BroadcastAlert(alert)
```

### Notifications
```go
// Send notification to all connected clients
dashboard.SendNotification("Title", "Message", "success")
```

## 🚀 Production Deployment

For production use:

1. **Reverse Proxy**: Use nginx or similar for SSL termination
2. **Authentication**: Add authentication middleware
3. **Monitoring**: Monitor WebSocket connections and memory usage
4. **Scaling**: Consider horizontal scaling for multiple dashboard instances

### Docker Deployment
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o dashboard examples/demo/web-dashboard/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/dashboard .
EXPOSE 8090
CMD ["./dashboard"]
```

## 🔍 Troubleshooting

### WebSocket Connection Issues
- Check firewall settings for port 8090
- Ensure WebSocket support in proxy configuration
- Verify browser WebSocket support

### Performance Issues
- Monitor memory usage with multiple connections
- Adjust refresh intervals for lower-powered systems
- Consider connection limits for high-traffic scenarios

### Data Issues
- Verify GPU metrics collector is running
- Check Prometheus exporter configuration
- Ensure monitoring service is recording data

## 📈 Metrics Available

### GPU Metrics
- `utilization_gpu` - GPU utilization percentage
- `utilization_memory` - Memory utilization percentage
- `temperature` - GPU temperature in Celsius
- `power_draw` - Current power consumption
- `memory_used` - Used memory in MB
- `memory_total` - Total memory in MB
- `fan_speed` - Fan speed percentage

### System Metrics
- `total_gpus` - Total number of GPUs
- `active_gpus` - Number of active GPUs
- `average_utilization` - Average GPU utilization
- `efficiency_score` - System efficiency score (0-100)
- `total_power_watts` - Total power consumption

### Cost Metrics
- `total_cost` - Total accumulated cost
- `hourly_cost` - Current hourly cost rate
- `gpu_hours` - GPU hours consumed
- `tokens_used` - Total tokens processed

## 🎯 Use Cases

1. **Data Center Monitoring** - Real-time GPU cluster oversight
2. **Cost Optimization** - Track and optimize GPU spending
3. **Performance Debugging** - Identify bottlenecks and inefficiencies
4. **Capacity Planning** - Monitor utilization for scaling decisions
5. **Alert Management** - Proactive issue detection and resolution
6. **Resource Allocation** - Optimize workload distribution

## 🔮 Future Enhancements

Planned features for upcoming releases:
- **Multi-cluster support** - Monitor multiple GPU clusters
- **Historical analytics** - Long-term trend analysis
- **Custom dashboards** - User-configurable layouts
- **Advanced alerting** - Configurable alert rules
- **Export capabilities** - PDF and CSV export options
- **Mobile app** - Native mobile application