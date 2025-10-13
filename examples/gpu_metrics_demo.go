package main

import (
	"context"
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
	fmt.Println("🚀 AgentaFlow Real-Time GPU Metrics Collection Demo")
	fmt.Println("================================================")

	// Create monitoring service
	monitoringService := observability.NewMonitoringService(10000)

	// Create GPU metrics collector (collect every 5 seconds)
	metricsCollector := gpu.NewMetricsCollector(5 * time.Second)

	// Create GPU metrics integration
	integration := observability.NewGPUMetricsIntegration(monitoringService, metricsCollector)

	// Create metrics aggregation service (aggregate every minute, retain 24 hours)
	aggregationService := gpu.NewMetricsAggregationService(
		metricsCollector,
		1*time.Minute,  // Aggregation interval
		24*time.Hour,   // Retention period
	)

	// Set up custom alert thresholds
	customThresholds := observability.GPUAlertThresholds{
		HighTemperature:     70.0,
		CriticalTemperature: 80.0,
		HighMemoryUsage:     75.0,
		CriticalMemoryUsage: 90.0,
		HighPowerUsage:      85.0,
		CriticalPowerUsage:  95.0,
		LowUtilization:      15.0,
		HighUtilization:     90.0,
	}
	integration.SetAlertThresholds(customThresholds)

	// Register custom callback for real-time monitoring
	metricsCollector.RegisterCallback(func(metrics gpu.GPUMetrics) {
		fmt.Printf("📊 GPU %s: %.1f%% util, %.1f°C, %.0fMB used, %.1fW\n",
			metrics.GPUID,
			metrics.UtilizationGPU,
			metrics.Temperature,
			float64(metrics.MemoryUsed),
			metrics.PowerDraw,
		)
	})

	// Start services
	fmt.Println("Starting GPU metrics collection...")
	
	if err := metricsCollector.Start(); err != nil {
		log.Fatalf("Failed to start metrics collector: %v", err)
	}

	if err := aggregationService.Start(); err != nil {
		log.Fatalf("Failed to start aggregation service: %v", err)
	}

	fmt.Println("✅ Services started successfully!")
	fmt.Println("\n🔍 Real-time GPU metrics (updates every 5 seconds):")

	// Set up graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start demo monitoring goroutines
	go demoReporting(metricsCollector, aggregationService, integration, ctx)
	go demoAlertMonitoring(integration, ctx)

	// Wait for shutdown signal
	<-sigChan
	fmt.Println("\n🛑 Shutting down...")

	// Stop services
	cancel()
	metricsCollector.Stop()
	aggregationService.Stop()

	fmt.Println("✅ Shutdown complete!")
}

// demoReporting demonstrates various reporting capabilities
func demoReporting(
	collector *gpu.MetricsCollector,
	aggregation *gpu.MetricsAggregationService,
	integration *observability.GPUMetricsIntegration,
	ctx context.Context,
) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			displaySystemOverview(collector)
			displayGPUStats(aggregation)
			displayHealthStatus(integration)
			displayEfficiencyReport(aggregation)
		}
	}
}

// demoAlertMonitoring demonstrates alert monitoring
func demoAlertMonitoring(integration *observability.GPUMetricsIntegration, ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			health := integration.GetGPUHealth()
			for gpuID, status := range health {
				if status.Status != "healthy" {
					fmt.Printf("⚠️  GPU %s Health Alert: %s - %v\n", 
						gpuID, status.Status, status.Issues)
				}
				
				if len(status.Alerts) > 0 {
					for _, alert := range status.Alerts {
						fmt.Printf("🚨 GPU %s Alert: %s (%s) - %s\n",
							gpuID, alert.Type, alert.Severity, alert.Message)
					}
				}
			}
		}
	}
}

// displaySystemOverview shows cluster-wide GPU overview
func displaySystemOverview(collector *gpu.MetricsCollector) {
	overview := collector.GetSystemOverview()
	
	fmt.Printf("\n📈 System Overview:\n")
	fmt.Printf("   Total GPUs: %v\n", overview["total_gpus"])
	fmt.Printf("   Active GPUs: %v\n", overview["active_gpus"])
	fmt.Printf("   Average Utilization: %.1f%%\n", overview["avg_utilization"])
	fmt.Printf("   Memory Usage: %vMB / %vMB (%.1f%%)\n", 
		overview["memory_used_mb"], 
		overview["memory_available_mb"],
		overview["memory_utilization"])
	fmt.Printf("   Total Processes: %v\n", overview["total_processes"])
}

// displayGPUStats shows aggregated GPU statistics
func displayGPUStats(aggregation *gpu.MetricsAggregationService) {
	stats := aggregation.GetAllGPUStats()
	
	if len(stats) == 0 {
		return
	}
	
	fmt.Printf("\n📊 GPU Statistics:\n")
	for gpuID, stat := range stats {
		fmt.Printf("   GPU %s:\n", gpuID)
		fmt.Printf("     Avg Utilization: %.1f%% (Peak: %.1f%%)\n", 
			stat.AverageUtilization, stat.PeakUtilization)
		fmt.Printf("     Avg Memory Usage: %.0fMB (Peak: %vMB)\n", 
			stat.AverageMemoryUsage, stat.PeakMemoryUsage)
		fmt.Printf("     Avg Temperature: %.1f°C (Max: %.1f°C)\n", 
			stat.AverageTemperature, stat.MaxTemperature)
		fmt.Printf("     Idle Time: %.1f%%\n", stat.IdleTimePercentage)
		fmt.Printf("     Efficiency Score: %.3f\n", stat.EfficiencyScore)
		fmt.Printf("     Energy Consumed: %.3f kWh\n", stat.TotalEnergyConsumed)
	}
}

// displayHealthStatus shows GPU health information
func displayHealthStatus(integration *observability.GPUMetricsIntegration) {
	health := integration.GetGPUHealth()
	
	if len(health) == 0 {
		return
	}
	
	fmt.Printf("\n🏥 GPU Health Status:\n")
	for gpuID, status := range health {
		statusIcon := "✅"
		if status.Status == "warning" {
			statusIcon = "⚠️"
		} else if status.Status == "critical" {
			statusIcon = "🚨"
		}
		
		fmt.Printf("   %s GPU %s: %s\n", statusIcon, gpuID, status.Status)
		fmt.Printf("     Temperature: %s, Memory: %s, Power: %s\n",
			status.TemperatureStatus, status.MemoryStatus, status.PowerStatus)
		
		if len(status.Issues) > 0 {
			fmt.Printf("     Issues: %v\n", status.Issues)
		}
		
		if len(status.Recommendations) > 0 {
			fmt.Printf("     Recommendations: %v\n", status.Recommendations)
		}
	}
}

// displayEfficiencyReport shows efficiency analysis
func displayEfficiencyReport(aggregation *gpu.MetricsAggregationService) {
	report := aggregation.GetEfficiencyReport()
	
	if clusterEff, ok := report["cluster_efficiency"].(map[string]interface{}); ok {
		fmt.Printf("\n⚡ Efficiency Report:\n")
		fmt.Printf("   Cluster Average Idle Time: %.1f%%\n", clusterEff["average_idle_time_percent"])
		fmt.Printf("   Cluster Average Efficiency: %.3f\n", clusterEff["average_efficiency_score"])
		fmt.Printf("   Utilization Potential: %.1f%%\n", clusterEff["utilization_potential"])
	}
	
	if rankings, ok := report["gpu_rankings"].([]interface{}); ok && len(rankings) > 0 {
		fmt.Printf("   Top Efficient GPU: %v\n", rankings[0])
		if len(rankings) > 1 {
			fmt.Printf("   Least Efficient GPU: %v\n", rankings[len(rankings)-1])
		}
	}
}

// Example function to demonstrate performance trends analysis
func demonstratePerformanceTrends(aggregation *gpu.MetricsAggregationService, gpuID string) {
	fmt.Printf("\n📈 Performance Trends for GPU %s (Last 4 hours):\n", gpuID)
	
	trends := aggregation.GetPerformanceTrends(gpuID, 4*time.Hour)
	
	if errMsg, hasError := trends["error"]; hasError {
		fmt.Printf("   Error: %s\n", errMsg)
		return
	}
	
	fmt.Printf("   Sample Count: %.0f\n", trends["sample_count"])
	
	if utilTrend, ok := trends["utilization_trend"].(map[string]float64); ok {
		fmt.Printf("   Utilization Trend: slope=%.3f, r²=%.3f\n", 
			utilTrend["slope"], utilTrend["r_squared"])
	}
	
	if tempTrend, ok := trends["temperature_trend"].(map[string]float64); ok {
		fmt.Printf("   Temperature Trend: slope=%.3f, r²=%.3f\n", 
			tempTrend["slope"], tempTrend["r_squared"])
	}
	
	if memTrend, ok := trends["memory_trend"].(map[string]float64); ok {
		fmt.Printf("   Memory Trend: slope=%.3f, r²=%.3f\n", 
			memTrend["slope"], memTrend["r_squared"])
	}
}

// Example function to demonstrate cost analysis
func demonstrateCostAnalysis(aggregation *gpu.MetricsAggregationService) {
	fmt.Printf("\n💰 Cost Analysis:\n")
	
	analysis := aggregation.GetCostAnalysis()
	
	fmt.Printf("   Total Estimated Cost: $%.2f\n", analysis["total_estimated_cost"])
	fmt.Printf("   Potential Savings: $%.2f (%.1f%%)\n", 
		analysis["total_potential_savings"], analysis["savings_percentage"])
	
	if gpuCosts, ok := analysis["gpu_costs"].(map[string]interface{}); ok {
		for gpuID, costs := range gpuCosts {
			if costMap, ok := costs.(map[string]interface{}); ok {
				fmt.Printf("   GPU %s: $%.2f actual, $%.2f potential savings\n",
					gpuID, costMap["actual_cost"], costMap["potential_savings"])
			}
		}
	}
}

// Example function to demonstrate metrics export
func demonstrateMetricsExport(collector *gpu.MetricsCollector, gpuID string) {
	fmt.Printf("\n📄 Exporting metrics for GPU %s...\n", gpuID)
	
	// Export last hour of metrics as JSON
	jsonData, err := collector.ExportMetricsJSON(gpuID, time.Now().Add(-1*time.Hour))
	if err != nil {
		fmt.Printf("   Error exporting metrics: %v\n", err)
		return
	}
	
	// Save to file
	filename := fmt.Sprintf("gpu_%s_metrics_%d.json", gpuID, time.Now().Unix())
	err = os.WriteFile(filename, jsonData, 0644)
	if err != nil {
		fmt.Printf("   Error saving file: %v\n", err)
		return
	}
	
	fmt.Printf("   ✅ Metrics exported to %s\n", filename)
}

// Example function to demonstrate process monitoring
func demonstrateProcessMonitoring(collector *gpu.MetricsCollector) {
	fmt.Printf("\n🔍 GPU Processes:\n")
	
	processes := collector.GetRunningProcesses()
	for gpuID, procList := range processes {
		if len(procList) > 0 {
			fmt.Printf("   GPU %s (%d processes):\n", gpuID, len(procList))
			for _, proc := range procList {
				fmt.Printf("     PID %d (%s): %s, %dMB, type=%s\n",
					proc.PID, proc.ProcessName, proc.ProcessName, 
					proc.MemoryUsed, proc.Type)
			}
		} else {
			fmt.Printf("   GPU %s: No active processes\n", gpuID)
		}
	}
}