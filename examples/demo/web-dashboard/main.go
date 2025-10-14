package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Finoptimize/agentaflow-sro-community/pkg/gpu"
	"github.com/Finoptimize/agentaflow-sro-community/pkg/observability"
)

func main() {
	fmt.Println("🚀 AgentaFlow Web Dashboard Demo")
	fmt.Println("===============================")

	// Create monitoring service
	fmt.Println("📊 Setting up monitoring service...")
	monitoringService := observability.NewMonitoringService(10000)

	// Create GPU metrics collector
	fmt.Println("🔧 Initializing GPU metrics collector...")
	metricsCollector := gpu.NewMetricsCollector(5 * time.Second)

	// Create Prometheus exporter
	fmt.Println("📈 Setting up Prometheus exporter...")
	prometheusConfig := observability.DefaultPrometheusConfig()
	prometheusExporter := observability.NewPrometheusExporter(monitoringService, prometheusConfig)

	// Register metrics
	prometheusExporter.RegisterGPUMetrics()
	prometheusExporter.RegisterCostMetrics()
	prometheusExporter.RegisterSchedulingMetrics()

	// Start Prometheus metrics server
	go func() {
		fmt.Println("🌐 Starting Prometheus metrics server on :8080...")
		if err := prometheusExporter.StartMetricsServer(8080); err != nil {
			log.Printf("Error starting Prometheus server: %v", err)
		}
	}()

	// Create GPU integration
	fmt.Println("🔌 Setting up GPU metrics integration...")
	integration := observability.NewGPUMetricsIntegration(monitoringService, metricsCollector)
	integration.SetPrometheusExporter(prometheusExporter)
	integration.EnablePrometheusExport(true)

	// Configure web dashboard
	dashboardConfig := observability.WebDashboardConfig{
		Port:                  8090,
		Title:                 "AgentaFlow GPU Monitoring Dashboard",
		RefreshInterval:       3000, // 3 seconds
		EnableRealTimeUpdates: true,
		Theme:                 "dark",
	}

	// Create web dashboard
	fmt.Println("🌐 Setting up web dashboard...")
	dashboard := observability.NewWebDashboard(dashboardConfig, monitoringService,
		metricsCollector, prometheusExporter)

	// Start metrics collection
	fmt.Println("📡 Starting GPU metrics collection...")
	metricsCollector.Start()

	// Register callback for real-time monitoring
	metricsCollector.RegisterCallback(func(metrics gpu.GPUMetrics) {
		fmt.Printf("📊 GPU %s: %.1f%% util, %.1f°C, %dMB used\n",
			metrics.GPUID, metrics.UtilizationGPU, metrics.Temperature, metrics.MemoryUsed)

		// Generate alerts for demonstration
		if metrics.Temperature > 75 {
			alert := observability.Alert{
				ID:        fmt.Sprintf("temp-%s-%d", metrics.GPUID, time.Now().Unix()),
				Level:     "warning",
				Message:   fmt.Sprintf("High temperature detected on GPU %s: %.1f°C", metrics.GPUID, metrics.Temperature),
				Source:    metrics.GPUID,
				Timestamp: time.Now(),
			}
			dashboard.BroadcastAlert(alert)
		}
	})

	// Start the web dashboard server
	go func() {
		fmt.Println("🌐 Starting web dashboard server...")
		if err := dashboard.Start(); err != nil {
			log.Printf("Error starting dashboard server: %v", err)
		}
	}()

	// Add some demo cost tracking
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			monitoringService.RecordCost(observability.CostEntry{
				Operation:  "gpu_inference",
				GPUHours:   0.0028, // ~10 seconds
				TokensUsed: 150,
				Cost:       0.012,
				Timestamp:  time.Now(),
			})
		}
	}()

	// Print access information
	time.Sleep(2 * time.Second)
	fmt.Println()
	fmt.Println("✅ All services started successfully!")
	fmt.Println()
	fmt.Println("📊 Access Points:")
	fmt.Println("   🌐 Web Dashboard:     http://localhost:8090")
	fmt.Println("   📈 Prometheus Metrics: http://localhost:8080/metrics")
	fmt.Println()
	fmt.Println("🔧 Dashboard Features:")
	fmt.Println("   • Real-time GPU monitoring")
	fmt.Println("   • Live cost tracking")
	fmt.Println("   • Interactive charts and visualizations")
	fmt.Println("   • WebSocket-based updates")
	fmt.Println("   • Alert management")
	fmt.Println("   • System health monitoring")
	fmt.Println()
	fmt.Println("📝 Demo Data:")
	fmt.Println("   • Simulated GPU metrics every 5 seconds")
	fmt.Println("   • Cost entries every 10 seconds")
	fmt.Println("   • Temperature alerts when > 75°C")
	fmt.Println()
	fmt.Println("Press Ctrl+C to stop the demo...")

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	fmt.Println("\n🛑 Shutting down services...")

	// Stop metrics collection
	metricsCollector.Stop()

	// Stop dashboard
	dashboard.Stop()

	fmt.Println("✅ Demo stopped successfully!")
}
