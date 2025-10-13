package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Finoptimize/agentaflow-sro-community/pkg/gpu"
	"github.com/Finoptimize/agentaflow-sro-community/pkg/observability"
)

func main() {
	fmt.Println("🚀 AgentaFlow Prometheus/Grafana Integration Demo")
	fmt.Println("===============================================")

	// Create monitoring service
	monitoringService := observability.NewMonitoringService(10000)

	// Create Prometheus exporter with custom configuration
	prometheusConfig := observability.PrometheusConfig{
		MetricsPrefix: "agentaflow",
		EnabledMetrics: map[string]bool{
			"gpu_metrics":        true,
			"scheduling_metrics": true,
			"serving_metrics":    true,
			"cost_metrics":       true,
			"system_metrics":     true,
		},
		MetricLabels: map[string]string{
			"instance": "demo",
			"version":  "community",
		},
	}

	prometheusExporter := observability.NewPrometheusExporter(monitoringService, prometheusConfig)

	// Register all metric types
	fmt.Println("📊 Registering Prometheus metrics...")
	prometheusExporter.RegisterGPUMetrics()
	prometheusExporter.RegisterSchedulingMetrics()
	prometheusExporter.RegisterServingMetrics()
	prometheusExporter.RegisterCostMetrics()
	prometheusExporter.RegisterSystemMetrics()

	// Create GPU metrics collector (collect every 5 seconds)
	metricsCollector := gpu.NewMetricsCollector(5 * time.Second)

	// Create GPU metrics integration with Prometheus support
	integration := observability.NewGPUMetricsIntegration(monitoringService, metricsCollector)
	integration.SetPrometheusExporter(prometheusExporter)
	integration.EnablePrometheusExport(true)

	// Set up custom alert thresholds for demo
	customThresholds := observability.GPUAlertThresholds{
		HighTemperature:     70.0,
		CriticalTemperature: 85.0,
		HighMemoryUsage:     80.0,
		CriticalMemoryUsage: 95.0,
		HighPowerUsage:      85.0,
		CriticalPowerUsage:  95.0,
		LowUtilization:      15.0,
		HighUtilization:     90.0,
	}
	integration.SetAlertThresholds(customThresholds)

	// Configure cost settings for demo
	awsCostConfig := observability.GPUCostConfiguration{
		CostPerHour: map[string]float64{
			"a100":    3.06,  // AWS p4d.xlarge
			"v100":    3.06,  // AWS p3.2xlarge
			"t4":      0.526, // AWS g4dn.xlarge
			"rtx":     1.20,  // Custom RTX pricing
			"generic": 1.50,  // Default
		},
		UseUtilizationFactor: true,
		MinUtilizationFactor: 0.15,
		IdleCostReduction:    0.20, // 20% cost reduction for idle time
		CloudProvider:        "aws",
		Region:               "us-west-2",
		Currency:             "USD",
		SpotInstanceDiscount: 0.60, // 60% discount for spot instances
		VolumeDiscounts: []observability.VolumeDiscount{
			{MinHours: 24, DiscountRate: 0.05},  // 5% discount for 24+ hours
			{MinHours: 168, DiscountRate: 0.10}, // 10% discount for 1 week+
		},
	}
	integration.SetCostConfiguration(awsCostConfig)

	// Create metrics aggregation service
	aggregationService := gpu.NewMetricsAggregationService(
		metricsCollector,
		30*time.Second, // Aggregate every 30 seconds for demo
		2*time.Hour,    // Retain 2 hours for demo
	)

	// Set up graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup

	// Start services
	fmt.Println("🔧 Starting services...")

	// Start GPU metrics collection
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := metricsCollector.Start(); err != nil {
			log.Printf("Error starting metrics collector: %v", err)
		}
	}()

	// Start metrics aggregation
	wg.Add(1)
	go func() {
		defer wg.Done()
		aggregationService.Start()
	}()

	// Start Prometheus metrics server
	fmt.Println("🌐 Starting Prometheus metrics server on :8080...")
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := prometheusExporter.StartMetricsServer(8080); err != nil {
			log.Printf("Prometheus metrics server error: %v", err)
		}
	}()

	// Start demo data generation
	wg.Add(1)
	go func() {
		defer wg.Done()
		generateDemoMetrics(ctx, prometheusExporter, integration)
	}()

	// Start periodic metrics sync
	wg.Add(1)
	go func() {
		defer wg.Done()
		syncTicker := time.NewTicker(10 * time.Second)
		defer syncTicker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-syncTicker.C:
				prometheusExporter.SyncFromMonitoringService()
			}
		}
	}()

	fmt.Println("✅ All services started successfully!")
	fmt.Println()
	fmt.Println("🎯 Integration Points:")
	fmt.Println("   • Prometheus metrics: http://localhost:8080/metrics")
	fmt.Println("   • Health endpoint: http://localhost:8080/health")
	fmt.Println()
	fmt.Println("🔧 Setup Instructions:")
	fmt.Println("   1. Deploy Prometheus: kubectl apply -f examples/k8s/monitoring/prometheus.yaml")
	fmt.Println("   2. Deploy Grafana: kubectl apply -f examples/k8s/monitoring/grafana.yaml")
	fmt.Println("   3. Import dashboard: examples/monitoring/grafana-dashboard.json")
	fmt.Println("   4. Access Grafana: kubectl port-forward svc/grafana-service 3000:3000 -n agentaflow-monitoring")
	fmt.Println("   5. Login: admin / agentaflow123")
	fmt.Println()
	fmt.Println("📊 Available Metrics:")
	fmt.Println("   • agentaflow_gpu_utilization_percent")
	fmt.Println("   • agentaflow_gpu_temperature_celsius")
	fmt.Println("   • agentaflow_gpu_memory_used_bytes")
	fmt.Println("   • agentaflow_gpu_health_status")
	fmt.Println("   • agentaflow_workloads_pending")
	fmt.Println("   • agentaflow_cost_total_dollars")
	fmt.Println("   • agentaflow_gpu_efficiency_score")
	fmt.Println()
	fmt.Println("Press Ctrl+C to stop...")

	// Wait for shutdown signal
	<-sigChan
	fmt.Println("\n🛑 Shutting down services...")

	cancel()
	metricsCollector.Stop()

	// Wait for all goroutines to finish
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		fmt.Println("✅ Shutdown complete!")
	case <-time.After(10 * time.Second):
		fmt.Println("⚠️ Shutdown timeout, forcing exit...")
	}
}

// generateDemoMetrics generates realistic demo metrics for Prometheus/Grafana visualization
func generateDemoMetrics(ctx context.Context, exporter *observability.PrometheusExporter, integration *observability.GPUMetricsIntegration) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	gpuCounter := 0
	workloadCounter := 0

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			generateSystemMetrics(exporter)
			generateWorkloadMetrics(exporter, workloadCounter)
			generateCostMetrics(exporter)
			generateAlertMetrics(exporter, integration)

			gpuCounter++
			workloadCounter++

			if gpuCounter%6 == 0 {
				fmt.Printf("📈 Generated metrics cycle %d (Prometheus: http://localhost:8080/metrics)\n", gpuCounter)
			}
		}
	}
}

// generateSystemMetrics creates realistic GPU system metrics
func generateSystemMetrics(exporter *observability.PrometheusExporter) {
	gpuConfigs := []struct {
		id   string
		name string
	}{
		{"gpu-0", "NVIDIA A100-SXM4-40GB"},
		{"gpu-1", "NVIDIA A100-SXM4-40GB"},
		{"gpu-2", "NVIDIA V100-SXM2-32GB"},
		{"gpu-3", "NVIDIA T4"},
	}

	baseTime := time.Now()

	for i, gpu := range gpuConfigs {
		labels := map[string]string{
			"gpu_id":   gpu.id,
			"gpu_name": gpu.name,
			"node":     fmt.Sprintf("node-%d", i%2+1),
		}

		// Generate realistic utilization patterns
		utilization := 20.0 + 60.0*getWaveValue(baseTime, time.Duration(i+1)*time.Minute)

		// Temperature correlates with utilization
		temperature := 35.0 + (utilization/100.0)*45.0

		// Memory usage varies independently
		memoryUsed := float64(8+i*8) * 1024 * 1024 * 1024 * (0.3 + 0.6*getWaveValue(baseTime, time.Duration(i+2)*2*time.Minute))
		memoryTotal := float64(40-i*8) * 1024 * 1024 * 1024

		// Power draw correlates with utilization
		powerDraw := 150.0 + (utilization/100.0)*250.0
		powerLimit := 400.0

		// Update metrics
		exporter.UpdateMetric("gpu_utilization_percent", utilization, labels)
		exporter.UpdateMetric("gpu_temperature_celsius", temperature, labels)
		exporter.UpdateMetric("gpu_memory_used_bytes", memoryUsed, labels)
		exporter.UpdateMetric("gpu_memory_total_bytes", memoryTotal, labels)
		exporter.UpdateMetric("gpu_power_draw_watts", powerDraw, labels)
		exporter.UpdateMetric("gpu_power_limit_watts", powerLimit, labels)

		// Clock speeds
		exporter.UpdateMetric("gpu_clock_graphics_mhz", 1400.0+200.0*(utilization/100.0), labels)
		exporter.UpdateMetric("gpu_clock_memory_mhz", 1215.0, labels)

		// Process count and efficiency
		processes := float64(1 + i%3)
		exporter.UpdateMetric("gpu_process_count", processes, labels)

		efficiency := utilization / (powerDraw / 100.0) / 100.0
		exporter.UpdateMetric("gpu_efficiency_score", efficiency, labels)

		idleTime := 100.0 - utilization
		exporter.UpdateMetric("gpu_idle_time_percent", idleTime, labels)

		// Health status (0=unhealthy, 1=warning, 2=healthy)
		var healthStatus float64 = 2
		if temperature > 85 {
			healthStatus = 0
		} else if temperature > 75 || utilization > 95 {
			healthStatus = 1
		}

		healthLabels := make(map[string]string)
		for k, v := range labels {
			healthLabels[k] = v
		}
		statusNames := []string{"unhealthy", "warning", "healthy"}
		healthLabels["status"] = statusNames[int(healthStatus)]
		exporter.UpdateMetric("gpu_health_status", healthStatus, healthLabels)
	}

	// System-wide metrics
	exporter.UpdateMetric("gpus_total", 4, map[string]string{"node": "cluster", "gpu_type": "mixed"})
	exporter.UpdateMetric("gpus_available", 3, map[string]string{"node": "cluster", "gpu_type": "mixed"})
	exporter.UpdateMetric("cluster_utilization_percent", 55.0+20.0*getWaveValue(baseTime, 5*time.Minute), map[string]string{})
	exporter.UpdateMetric("cluster_efficiency_score", 0.65+0.15*getWaveValue(baseTime, 7*time.Minute), map[string]string{})
}

// generateWorkloadMetrics creates scheduling and workload metrics
func generateWorkloadMetrics(exporter *observability.PrometheusExporter, counter int) {
	// Pending workloads vary over time
	baseTime := time.Now()
	pendingHigh := 3.0 + 5.0*getWaveValue(baseTime, 3*time.Minute)
	pendingLow := 1.0 + 2.0*getWaveValue(baseTime, 4*time.Minute)

	exporter.UpdateMetric("workloads_pending", pendingHigh, map[string]string{"priority": "high"})
	exporter.UpdateMetric("workloads_pending", pendingLow, map[string]string{"priority": "low"})

	// Running workloads
	exporter.UpdateMetric("workloads_running", 2, map[string]string{"gpu_id": "gpu-0", "priority": "high"})
	exporter.UpdateMetric("workloads_running", 1, map[string]string{"gpu_id": "gpu-1", "priority": "low"})
	exporter.UpdateMetric("workloads_running", 1, map[string]string{"gpu_id": "gpu-2", "priority": "high"})

	// Completed workloads (counters)
	if counter%10 == 0 {
		exporter.UpdateMetric("workloads_completed", 1, map[string]string{"status": "success", "priority": "high"})
	}
	if counter%15 == 0 {
		exporter.UpdateMetric("workloads_completed", 1, map[string]string{"status": "success", "priority": "low"})
	}
	if counter%50 == 0 {
		exporter.UpdateMetric("workloads_completed", 1, map[string]string{"status": "failed", "priority": "high"})
	}

	// Scheduling performance
	schedulingDuration := 0.1 + 0.4*getWaveValue(baseTime, 2*time.Minute)
	exporter.UpdateMetric("scheduling_duration_seconds", schedulingDuration, map[string]string{"strategy": "least_utilized"})

	if counter%5 == 0 {
		exporter.UpdateMetric("scheduling_decisions_total", 1, map[string]string{"strategy": "least_utilized", "outcome": "success"})
	}

	efficiency := 75.0 + 15.0*getWaveValue(baseTime, 6*time.Minute)
	exporter.UpdateMetric("gpu_allocation_efficiency", efficiency, map[string]string{"strategy": "least_utilized"})
}

// generateCostMetrics creates cost tracking metrics
func generateCostMetrics(exporter *observability.PrometheusExporter) {
	baseTime := time.Now()

	// Hourly cost variations
	inferenceCost := 5.0 + 3.0*getWaveValue(baseTime, 4*time.Minute)
	trainingCost := 15.0 + 10.0*getWaveValue(baseTime, 8*time.Minute)

	exporter.UpdateMetric("cost_total_dollars", inferenceCost, map[string]string{
		"operation": "inference",
		"model_id":  "gpt-model",
		"currency":  "USD",
	})

	exporter.UpdateMetric("cost_total_dollars", trainingCost, map[string]string{
		"operation": "training",
		"model_id":  "llama-model",
		"currency":  "USD",
	})

	// Cost per hour rates
	exporter.UpdateMetric("cost_per_hour_dollars", 3.06, map[string]string{
		"resource_type": "a100",
		"currency":      "USD",
	})

	exporter.UpdateMetric("cost_per_hour_dollars", 0.526, map[string]string{
		"resource_type": "t4",
		"currency":      "USD",
	})

	// GPU hours consumed
	a100Hours := 2.5 + 1.0*getWaveValue(baseTime, 5*time.Minute)
	t4Hours := 1.2 + 0.8*getWaveValue(baseTime, 6*time.Minute)

	exporter.UpdateMetric("gpu_hours_consumed", a100Hours, map[string]string{
		"gpu_type":      "a100",
		"workload_type": "training",
	})

	exporter.UpdateMetric("gpu_hours_consumed", t4Hours, map[string]string{
		"gpu_type":      "t4",
		"workload_type": "inference",
	})

	// Monthly cost estimates
	exporter.UpdateMetric("estimated_monthly_cost_dollars", 2200.0+300.0*getWaveValue(baseTime, 10*time.Minute),
		map[string]string{"resource_type": "gpu_cluster"})
}

// generateAlertMetrics creates alert and health metrics
func generateAlertMetrics(exporter *observability.PrometheusExporter, integration *observability.GPUMetricsIntegration) {
	baseTime := time.Now()

	// Generate periodic alerts
	alertCounter := int(baseTime.Unix()/30) % 10

	if alertCounter == 0 {
		exporter.UpdateMetric("alerts_total", 1, map[string]string{
			"severity": "warning",
			"type":     "temperature",
			"source":   "gpu_monitor",
		})
	}

	if alertCounter == 5 {
		exporter.UpdateMetric("alerts_total", 1, map[string]string{
			"severity": "info",
			"type":     "utilization",
			"source":   "scheduler",
		})
	}

	// Active alerts
	activeWarnings := float64(alertCounter % 3)
	activeCritical := float64(0)
	if alertCounter == 8 {
		activeCritical = 1
	}

	exporter.UpdateMetric("active_alerts", activeWarnings, map[string]string{
		"severity": "warning",
		"type":     "temperature",
	})

	exporter.UpdateMetric("active_alerts", activeCritical, map[string]string{
		"severity": "critical",
		"type":     "memory",
	})

	// System health
	components := []string{"gpu_monitor", "scheduler", "cost_tracker"}
	for i, component := range components {
		health := 2.0 // healthy
		if component == "cost_tracker" && alertCounter > 7 {
			health = 1.0 // degraded
		}

		exporter.UpdateMetric("component_health_status", health, map[string]string{
			"component": component,
		})

		uptime := float64(time.Now().Unix() - 1000*int64(i))
		exporter.UpdateMetric("system_uptime_seconds", uptime, map[string]string{
			"component": component,
		})
	}
}

// getWaveValue returns a value between 0 and 1 based on a sine wave
func getWaveValue(baseTime time.Time, period time.Duration) float64 {
	elapsed := time.Since(baseTime)
	radians := (float64(elapsed) / float64(period)) * 2.0 * math.Pi
	return (1.0 + 0.8*time.Duration(radians).Seconds()) / 2.0
}
