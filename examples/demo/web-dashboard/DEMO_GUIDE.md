# AgentaFlow Web Dashboard Demo - Complete Setup

## 🎯 Overview

This comprehensive demo showcases a **production-ready GPU monitoring dashboard** that can run on **any laptop** without requiring actual NVIDIA GPUs. The demo simulates realistic GPU workloads and provides a complete monitoring experience identical to what you'd see in a production environment.

## ✨ Key Features Demonstrated

### 🖥️ **Modern Web Dashboard**
- **Responsive design** that works on desktop, tablet, and mobile
- **Real-time charts** using Chart.js with WebSocket updates every 2 seconds
- **Interactive GPU cards** showing utilization, temperature, and memory usage
- **System overview metrics** with efficiency scoring
- **Dark theme** optimized for monitoring environments

### 📊 **Realistic GPU Simulation**
- **4 different GPU types**: RTX 4090, RTX 4080, RTX 4070 Ti, Tesla V100
- **Realistic specifications**: 8GB to 32GB memory, 50W to 400W power consumption
- **Dynamic workload patterns**: Idle → Light Inference → Training → Heavy Inference → Batch Processing
- **Temperature modeling**: Thermal throttling and fan speed curves
- **Memory management**: Realistic allocation patterns

### 💰 **Cost Tracking & Analytics**
- **Multi-operation tracking**: Training, inference, model serving, batch processing
- **Real-time cost calculation** with different rates per operation type
- **Cost forecasting** and optimization recommendations
- **GPU hour tracking** with utilization-based pricing

### 🚨 **Alert Management**
- **Real-time alerts** for temperature (>80°C), utilization (>95%), memory (>90%)
- **WebSocket notifications** broadcast to all connected clients
- **Alert history** and management interface
- **Browser notifications** (when permitted)

### 📈 **Performance Analytics**
- **Trend analysis** for utilization, temperature, and costs
- **Efficiency scoring** based on multiple factors
- **System health monitoring** with comprehensive metrics
- **Historical data** visualization

## 🚀 Running the Demo

### 1. Start the Demo
```bash
cd examples/demo/web-dashboard
go run main.go
```

### 2. Access the Dashboard
- **Web Dashboard**: http://localhost:8090
- **Prometheus Metrics**: http://localhost:8080/metrics

### 3. Explore the Features
- Watch real-time GPU metrics update every 2-3 seconds
- Observe automatic workload pattern changes every 45 seconds
- Check for temperature and utilization alerts
- Monitor cost accumulation over time
- Test WebSocket connectivity (connection status in top-right)

## 🎮 Demo Highlights

### **Simulated Hardware**
```
📊 GPU Fleet:
   • gpu-0: NVIDIA GeForce RTX 4090 (24GB VRAM, ~350W)
   • gpu-1: NVIDIA GeForce RTX 4080 (16GB VRAM, ~320W)  
   • gpu-2: NVIDIA GeForce RTX 4070 Ti (12GB VRAM, ~285W)
   • gpu-3: NVIDIA Tesla V100 (32GB VRAM, ~300W)
```

### **Workload Patterns**
- **Idle**: 0-15% utilization, minimal memory usage
- **Light Inference**: 20-45% utilization, 30-55% memory
- **Training**: 70-98% utilization, 75-95% memory, high temperature
- **Heavy Inference**: 45-75% utilization, 50-70% memory
- **Batch Processing**: 85-100% utilization, 80-98% memory

### **Alert Triggers**
```
🔥 Temperature Alerts: > 80°C (Critical)
⚡ High Utilization: > 95% (Warning)  
💾 Memory Usage: > 90% (Warning)
```

### **Cost Structure**
```
💰 Operation Costs (per GPU hour):
   • Training: $2.50/hour
   • Inference: $1.80/hour
   • Model Serving: $2.00/hour
   • Batch Processing: $2.20/hour
```

## 🔧 Technical Implementation

### **Architecture**
```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Web Dashboard │◄───┤  WebSocket       │◄───┤ Mock Collector  │
│   (Port 8090)   │    │  Real-time       │    │ (4 GPUs)        │
└─────────────────┘    │  Updates         │    └─────────────────┘
                       └──────────────────┘              │
┌─────────────────┐    ┌──────────────────┐              │
│   Prometheus    │◄───┤  Monitoring      │◄─────────────┘
│   (Port 8080)   │    │  Service         │
└─────────────────┘    └──────────────────┘
```

### **Core Components**
1. **MockMetricsCollector**: Generates realistic GPU metrics without hardware
2. **WebDashboard**: Modern HTML5 interface with Chart.js and WebSockets
3. **MonitoringService**: Cost tracking and system health monitoring
4. **PrometheusExporter**: Standard metrics export for observability stack

### **API Endpoints**
```
GET  /                           # Main dashboard interface
GET  /ws                         # WebSocket for real-time updates
GET  /health                     # Health check
GET  /api/v1/metrics             # Complete metrics data
GET  /api/v1/system/stats        # System statistics
GET  /api/v1/gpus               # GPU list and status
GET  /api/v1/alerts             # Active alerts
GET  /api/v1/costs              # Cost information
GET  /api/v1/performance        # Performance analytics
```

## 🌟 Production Readiness Features

### **Scalability**
- **Multi-GPU support**: Easily scales to dozens of GPUs
- **WebSocket management**: Handles multiple concurrent dashboard connections
- **Memory efficient**: Circular buffers with configurable history limits
- **Background processing**: Non-blocking metrics collection

### **Reliability**
- **Graceful shutdown**: Proper cleanup of resources
- **Error handling**: Comprehensive error management
- **Connection recovery**: Automatic WebSocket reconnection
- **Health monitoring**: Self-monitoring and status reporting

### **Integration Ready**
- **Prometheus compatibility**: Standard metrics export
- **REST API**: Complete programmatic access
- **WebSocket API**: Real-time event streaming
- **CORS support**: Cross-origin resource sharing enabled

### **Security Considerations**
- **Input validation**: All API inputs validated
- **Rate limiting**: WebSocket connection limits
- **Origin checking**: Configurable origin validation
- **Logging**: Comprehensive request logging

## 📱 Dashboard Interface Guide

### **System Overview Cards**
- **Total GPUs**: Count of available GPUs
- **Active GPUs**: Number of GPUs with >5% utilization
- **Average Utilization**: Fleet-wide average utilization
- **Efficiency Score**: 0-100 system efficiency rating
- **Total Power**: Aggregate power consumption
- **Memory Usage**: System-wide memory utilization

### **GPU Status Cards**
Each GPU displays:
- **Name and Model**: GPU identification
- **Status Badge**: idle/active/warning/critical
- **Utilization Bar**: Real-time utilization percentage
- **Memory Bar**: Used/Total memory with percentage
- **Temperature Bar**: Current temperature with color coding

### **Performance Charts**
- **GPU Performance**: Line chart showing utilization and temperature trends
- **Cost Analytics**: Doughnut chart breaking down cost categories
- **Time Range Selector**: 1H/6H/24H data views

### **Alerts Panel**
- **Real-time alerts** with severity levels
- **Alert details** including source and timestamp
- **One-click resolution** for alert management
- **Alert counter** in the header

## 🔬 Demo Scenarios

### **Scenario 1: Normal Operations**
- Monitor steady-state workloads
- Observe utilization patterns
- Track cost accumulation
- System efficiency monitoring

### **Scenario 2: High Load Training**
- Watch training workload trigger
- Observe temperature increases
- Monitor memory allocation
- See power consumption rise

### **Scenario 3: Alert Management**
- Wait for temperature >80°C alert
- Observe real-time dashboard notification
- Check alert in alerts panel
- Monitor system response

### **Scenario 4: API Integration**
```bash
# System status
curl http://localhost:8090/api/v1/system/stats

# GPU details
curl http://localhost:8090/api/v1/gpus

# Current alerts
curl http://localhost:8090/api/v1/alerts

# Cost information
curl http://localhost:8090/api/v1/costs
```

## 🛠️ Customization Options

### **GPU Configuration**
Modify `numGPUs` in `main.go` to simulate different cluster sizes:
```go
numGPUs := 8 // Simulate 8 GPUs instead of 4
```

### **Update Intervals**
Adjust refresh rates in `dashboardConfig`:
```go
RefreshInterval: 1000, // 1 second updates
```

### **Alert Thresholds**
Modify alert triggers in the callback function:
```go
if metrics.Temperature > 75 { // Lower temperature threshold
    // Generate alert
}
```

### **Cost Rates**
Update cost calculations in the cost tracking goroutine:
```go
cost = gpuHours * 3.50 // Higher training rate
```

## 🎯 Production Deployment Considerations

### **Infrastructure Requirements**
- **CPU**: 2+ cores (4+ recommended for high-throughput)
- **Memory**: 4GB+ RAM (8GB+ for large clusters)
- **Network**: Low latency for WebSocket performance
- **Storage**: Minimal (metrics stored in memory)

### **Scaling Guidelines**
- **Up to 50 GPUs**: Single instance handles easily
- **50-200 GPUs**: Consider connection pooling
- **200+ GPUs**: Implement horizontal scaling

### **Production Enhancements**
- **Authentication**: Add user authentication and authorization
- **TLS/SSL**: Enable HTTPS for production security
- **Database**: Persist historical data to database
- **Caching**: Implement Redis for session management
- **Load Balancing**: Use nginx/HAProxy for multiple instances

## 🌟 Value Proposition

### **For Development Teams**
- **No Hardware Dependencies**: Test monitoring without expensive GPUs
- **Realistic Simulation**: Production-like behavior patterns
- **API Testing**: Complete REST and WebSocket APIs
- **Integration Ready**: Prometheus and standard metrics

### **For Demos & Sales**
- **Impressive Visuals**: Modern, professional dashboard
- **Real-time Updates**: Engaging live demonstrations
- **Comprehensive Features**: Full monitoring stack showcase
- **Easy Setup**: Runs anywhere, no prerequisites

### **For Production Planning**
- **Architecture Preview**: Exact production interface
- **Performance Baseline**: Understanding of metrics and costs
- **Alert Testing**: Comprehensive alerting system
- **Capacity Planning**: Resource usage patterns

## 🚀 Next Steps

1. **Explore the Dashboard**: Spend 10-15 minutes with the live interface
2. **Test API Endpoints**: Use curl or Postman to explore the APIs
3. **Monitor Patterns**: Watch workload changes and alert generation
4. **Check Prometheus**: View metrics at http://localhost:8080/metrics
5. **Customize Settings**: Modify GPU count, thresholds, or update rates

## 📞 Support & Documentation

- **GitHub Repository**: https://github.com/Finoptimize/agentaflow-sro-community
- **API Documentation**: Available at `/api/v1/*` endpoints
- **WebSocket Protocol**: Connect to `/ws` for real-time events
- **Prometheus Metrics**: Standard exposition at `/metrics`

---

**🎉 Congratulations!** You now have a comprehensive GPU monitoring dashboard demo that showcases enterprise-grade monitoring capabilities without requiring any specialized hardware. The demo provides a complete preview of what AgentaFlow SRO Community Edition offers for production GPU infrastructure monitoring.