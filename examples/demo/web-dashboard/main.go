package main

import (
	"fmt"
	"log"
	"math/rand"
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
	fmt.Println("🎯 Purpose: Comprehensive GPU monitoring demo for local development")
	fmt.Println("💻 Compatible: Runs on any laptop without requiring NVIDIA GPUs")
	fmt.Println()

	// Determine number of GPUs to simulate based on demo preference
	numGPUs := 4 // Default to 4 GPUs for a good demo
	if gpuCount := os.Getenv("DEMO_GPU_COUNT"); gpuCount != "" {
		fmt.Printf("🎮 GPU count from environment: %s\n", gpuCount)
		// In a real app, you'd parse this, but for demo we'll stick with 4
	}

	fmt.Printf("🔧 Simulating %d GPUs with realistic workload patterns\n", numGPUs)

	// Create monitoring service with larger buffer for demo
	fmt.Println("📊 Setting up monitoring service...")
	monitoringService := observability.NewMonitoringService(50000)

	// Create MOCK GPU metrics collector for demo (works without real GPUs)
	fmt.Println("🎮 Initializing MOCK GPU metrics collector for demo...")
	mockCollector := gpu.NewMockMetricsCollector(3*time.Second, numGPUs)

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
		fmt.Println("🌐 Starting Prometheus metrics server on :9001...")
		if err := prometheusExporter.StartMetricsServer(9001); err != nil {
			log.Printf("Error starting Prometheus server: %v", err)
		}
	}()

	// Create GPU integration with mock collector
	fmt.Println("🔌 Setting up GPU metrics integration...")
	integration := observability.NewGPUMetricsIntegration(monitoringService, mockCollector)
	integration.SetPrometheusExporter(prometheusExporter)
	integration.EnablePrometheusExport(true)

	// Configure web dashboard with enhanced settings
	dashboardConfig := observability.WebDashboardConfig{
		Port:                  9000,
		Title:                 "AgentaFlow SRO Community Edition - GPU Monitoring Dashboard",
		RefreshInterval:       2000, // 2 seconds for smooth demo
		EnableRealTimeUpdates: true,
		Theme:                 "dark",
	}

	// Create web dashboard
	fmt.Println("🌐 Setting up web dashboard...")
	dashboard := observability.NewWebDashboard(monitoringService, mockCollector, prometheusExporter, dashboardConfig)

	// Start mock metrics collection
	fmt.Println("📡 Starting MOCK GPU metrics collection...")
	if err := mockCollector.Start(); err != nil {
		log.Fatalf("Failed to start mock collector: %v", err)
	}

	// Register callback for real-time monitoring and alerts
	mockCollector.RegisterCallback(func(metrics gpu.GPUMetrics) {
		// Print periodic status updates
		if rand.Float64() < 0.1 { // 10% chance to print status
			fmt.Printf("📊 %s [%s]: %.1f%% util, %.1f°C, %.1fGB/%.1fGB memory\n",
				metrics.GPUID,
				metrics.Name,
				metrics.UtilizationGPU,
				metrics.Temperature,
				float64(metrics.MemoryUsed)/1024,
				float64(metrics.MemoryTotal)/1024,
			)
		}

		// Generate realistic alerts for demonstration
		if metrics.Temperature > 80 {
			fmt.Printf("🔥 HIGH TEMPERATURE ALERT: GPU %s reached %.1f°C (threshold: 80°C)\n",
				metrics.GPUID, metrics.Temperature)

			// Broadcast alert to dashboard (if alert broadcasting was implemented)
			alert := observability.Alert{
				ID:        fmt.Sprintf("temp-%s-%d", metrics.GPUID, time.Now().Unix()),
				Level:     "critical",
				Message:   fmt.Sprintf("High temperature on GPU %s: %.1f°C", metrics.GPUID, metrics.Temperature),
				Source:    metrics.GPUID,
				Timestamp: time.Now(),
			}
			dashboard.BroadcastAlert(alert)
		}

		if metrics.UtilizationGPU > 95 {
			fmt.Printf("⚡ HIGH UTILIZATION: GPU %s at %.1f%% utilization\n",
				metrics.GPUID, metrics.UtilizationGPU)

			alert := observability.Alert{
				ID:        fmt.Sprintf("util-%s-%d", metrics.GPUID, time.Now().Unix()),
				Level:     "warning",
				Message:   fmt.Sprintf("High utilization on GPU %s: %.1f%%", metrics.GPUID, metrics.UtilizationGPU),
				Source:    metrics.GPUID,
				Timestamp: time.Now(),
			}
			dashboard.BroadcastAlert(alert)
		}

		// Memory usage alerts
		if metrics.MemoryTotal > 0 {
			memUsagePercent := float64(metrics.MemoryUsed) / float64(metrics.MemoryTotal) * 100
			if memUsagePercent > 90 {
				fmt.Printf("💾 MEMORY WARNING: GPU %s memory at %.1f%%\n",
					metrics.GPUID, memUsagePercent)

				alert := observability.Alert{
					ID:        fmt.Sprintf("mem-%s-%d", metrics.GPUID, time.Now().Unix()),
					Level:     "warning",
					Message:   fmt.Sprintf("High memory usage on GPU %s: %.1f%%", metrics.GPUID, memUsagePercent),
					Source:    metrics.GPUID,
					Timestamp: time.Now(),
				}
				dashboard.BroadcastAlert(alert)
			}
		}
	})

	// Start the web dashboard server
	go func() {
		fmt.Println("🌐 Starting web dashboard server...")
		if err := dashboard.Start(); err != nil {
			log.Printf("Error starting dashboard server: %v", err)
		}
	}()

	// Enhanced demo cost tracking with realistic patterns
	go func() {
		ticker := time.NewTicker(8 * time.Second)
		defer ticker.Stop()

		operations := []string{
			"gpu_training", "gpu_inference", "model_serving",
			"batch_processing", "image_generation", "llm_inference",
		}

		for range ticker.C {
			// Random operation
			operation := operations[rand.Intn(len(operations))]

			// Realistic cost calculation based on operation type
			var gpuHours, tokensUsed float64
			var cost float64

			switch operation {
			case "gpu_training":
				gpuHours = 0.005 + rand.Float64()*0.01       // 5-15 minutes worth
				tokensUsed = float64(rand.Intn(5000) + 1000) // 1K-6K tokens
				cost = gpuHours * 2.50                       // $2.50/hour for training
			case "gpu_inference":
				gpuHours = 0.001 + rand.Float64()*0.002     // 1-3 minutes worth
				tokensUsed = float64(rand.Intn(1500) + 100) // 100-1600 tokens
				cost = gpuHours * 1.80                      // $1.80/hour for inference
			case "model_serving":
				gpuHours = 0.002 + rand.Float64()*0.003 // 2-5 minutes worth
				tokensUsed = float64(rand.Intn(2000) + 200)
				cost = gpuHours * 2.00
			default:
				gpuHours = 0.002 + rand.Float64()*0.004
				tokensUsed = float64(rand.Intn(3000) + 500)
				cost = gpuHours * 2.20
			}

			monitoringService.RecordCost(observability.CostEntry{
				Operation:  operation,
				GPUHours:   gpuHours,
				TokensUsed: int64(tokensUsed),
				Cost:       cost,
				Timestamp:  time.Now(),
			})

			// Occasionally print cost updates
			if rand.Float64() < 0.2 { // 20% chance
				fmt.Printf("💰 Cost recorded: %s - $%.3f (%.4f GPU hours, %.0f tokens)\n",
					operation, cost, gpuHours, tokensUsed)
			}
		}
	}()

	// Demo workload pattern automation
	go func() {
		ticker := time.NewTicker(45 * time.Second)
		defer ticker.Stop()

		patterns := []string{"Idle", "Light Inference", "Training", "Heavy Inference", "Batch Processing"}

		for range ticker.C {
			// Randomly change workload patterns on GPUs for demo
			gpuToChange := fmt.Sprintf("gpu-%d", rand.Intn(numGPUs))
			newPattern := patterns[rand.Intn(len(patterns))]

			fmt.Printf("🎮 DEMO: Triggering %s workload on %s\n", newPattern, gpuToChange)
			mockCollector.TriggerWorkloadChange(gpuToChange, newPattern)

			// Send notification to dashboard
			dashboard.SendNotification(
				"Workload Pattern Changed",
				fmt.Sprintf("GPU %s switched to %s workload pattern", gpuToChange, newPattern),
				"info",
			)
		}
	}()

	// Print comprehensive startup information
	time.Sleep(3 * time.Second)
	fmt.Println()
	fmt.Println("✅ All services started successfully!")
	fmt.Println()
	fmt.Println("📊 Access Points:")
	fmt.Println("   🌐 Web Dashboard:      http://localhost:9000")
	fmt.Println("   📈 Prometheus Metrics:  http://localhost:9001/metrics")
	fmt.Println("   🔍 API Endpoints:")
	fmt.Println("      • GET /api/v1/metrics        - Complete metrics data")
	fmt.Println("      • GET /api/v1/system/stats   - System statistics")
	fmt.Println("      • GET /api/v1/alerts         - Active alerts")
	fmt.Println("      • GET /api/v1/costs          - Cost information")
	fmt.Println("      • GET /api/v1/gpus           - GPU list")
	fmt.Println("      • GET /health                - Health check")
	fmt.Println("      • GET /ws                    - WebSocket connection")
	fmt.Println()
	fmt.Println("🎯 Dashboard Features:")
	fmt.Println("   • 📊 Real-time GPU monitoring with live charts")
	fmt.Println("   • 💰 Live cost tracking and forecasting")
	fmt.Println("   • 📈 Interactive performance visualizations")
	fmt.Println("   • 🔄 WebSocket-based real-time updates (2sec)")
	fmt.Println("   • 🚨 Alert management with notifications")
	fmt.Println("   • 💡 System health monitoring and efficiency scoring")
	fmt.Println("   • 📱 Responsive design (desktop, tablet, mobile)")
	fmt.Println("   • 🌙 Modern dark theme optimized for monitoring")
	fmt.Println()
	fmt.Println("🎮 Demo Data Generation:")
	fmt.Printf("   • %d simulated GPUs with realistic hardware specs\n", numGPUs)
	fmt.Println("   • Dynamic workload patterns (Idle → Training → Inference)")
	fmt.Println("   • Realistic temperature, utilization, and memory metrics")
	fmt.Println("   • Automated pattern changes every 45 seconds")
	fmt.Println("   • Cost tracking with multiple operation types")
	fmt.Println("   • Alert generation based on thresholds:")
	fmt.Println("     - Temperature > 80°C (critical)")
	fmt.Println("     - GPU utilization > 95% (warning)")
	fmt.Println("     - Memory usage > 90% (warning)")
	fmt.Println()
	fmt.Println("🔧 Hardware Simulation:")
	fmt.Println("   • RTX 4090, RTX 4080, A100, H100 GPU types")
	fmt.Println("   • 8GB to 80GB memory configurations")
	fmt.Println("   • Realistic power consumption (50W to 400W)")
	fmt.Println("   • Temperature modeling with thermal throttling")
	fmt.Println("   • Fan speed curves and clock speed variations")
	fmt.Println()
	fmt.Println("🚀 Production Features Demonstrated:")
	fmt.Println("   • Multi-GPU cluster monitoring")
	fmt.Println("   • Cost optimization recommendations")
	fmt.Println("   • Performance trend analysis")
	fmt.Println("   • Real-time alerting system")
	fmt.Println("   • RESTful API with comprehensive endpoints")
	fmt.Println("   • Prometheus metrics integration")
	fmt.Println("   • WebSocket real-time communication")
	fmt.Println()
	fmt.Println("📝 Demo Tips:")
	fmt.Println("   • Watch for automatic workload pattern changes")
	fmt.Println("   • Observe real-time chart updates in the dashboard")
	fmt.Println("   • Check alert notifications for threshold violations")
	fmt.Println("   • Explore different API endpoints for integration examples")
	fmt.Println("   • Monitor cost accumulation over time")
	fmt.Println()
	fmt.Println("⚡ Press Ctrl+C to stop the demo...")

	// Enhanced system monitoring
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			overview := mockCollector.GetSystemOverview()
			fmt.Printf("📊 SYSTEM STATUS: %d GPUs, %.1f%% avg utilization, %.1f%% memory usage\n",
				overview["total_gpus"],
				overview["avg_utilization"],
				overview["memory_utilization"],
			)

			fmt.Printf("   💻 Active connections to dashboard: %d\n", dashboard.GetActiveConnections())
		}
	}()

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	fmt.Println("\n🛑 Shutting down AgentaFlow demo services...")

	// Stop metrics collection
	mockCollector.Stop()
	fmt.Println("✅ Stopped mock GPU metrics collection")

	// Stop dashboard
	dashboard.Stop()
	fmt.Println("✅ Stopped web dashboard server")

	fmt.Println("✅ Demo stopped successfully!")
	fmt.Println()
	fmt.Println("🙏 Thank you for trying AgentaFlow SRO Community Edition!")
	fmt.Println("   For more information: https://github.com/Finoptimize/agentaflow-sro-community")
}
