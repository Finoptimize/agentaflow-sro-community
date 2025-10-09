package main

import (
	"fmt"
	"time"

	"github.com/Finoptimize/agentaflow-sro-community/pkg/observability"
)

func main() {
	fmt.Println("=== Observability & Cost Tracking Example ===\n")

	// Create monitoring service
	fmt.Println("1. Initializing monitoring service...")
	monitor := observability.NewMonitoringService(10000)

	// Create debugger
	debugger := observability.NewDebugger(observability.DebugLevelInfo)

	// Record various metrics
	fmt.Println("\n2. Recording performance metrics...")
	metrics := []observability.Metric{
		{
			Name:  "inference_latency_ms",
			Type:  observability.MetricHistogram,
			Value: 45.5,
			Labels: map[string]string{
				"model":  "gpt-3.5",
				"region": "us-east-1",
			},
		},
		{
			Name:  "gpu_utilization_percent",
			Type:  observability.MetricGauge,
			Value: 78.5,
			Labels: map[string]string{
				"gpu_id": "gpu-0",
				"node":   "node-1",
			},
		},
		{
			Name:  "requests_processed",
			Type:  observability.MetricCounter,
			Value: 150,
			Labels: map[string]string{
				"model":  "bert-large",
				"status": "success",
			},
		},
		{
			Name:  "cache_hit_rate",
			Type:  observability.MetricGauge,
			Value: 65.2,
			Labels: map[string]string{
				"model": "gpt-3.5",
			},
		},
	}

	for _, metric := range metrics {
		monitor.RecordMetric(metric)
		fmt.Printf("   Recorded: %s = %.2f (%s)\n", metric.Name, metric.Value, metric.Type)
	}

	// Record events
	fmt.Println("\n3. Recording system events...")
	events := []observability.Event{
		{
			ID:       "evt-1",
			Type:     "deployment",
			Severity: "info",
			Message:  "Model deployed successfully",
			Source:   "deployment_service",
			Metadata: map[string]interface{}{
				"model":   "gpt-3.5-turbo",
				"version": "1.0.0",
			},
		},
		{
			ID:       "evt-2",
			Type:     "performance",
			Severity: "warn",
			Message:  "High GPU temperature detected",
			Source:   "monitoring_agent",
			Metadata: map[string]interface{}{
				"gpu_id":      "gpu-2",
				"temperature": 85.0,
				"threshold":   80.0,
			},
		},
		{
			ID:       "evt-3",
			Type:     "error",
			Severity: "error",
			Message:  "Model inference timeout",
			Source:   "inference_service",
			Metadata: map[string]interface{}{
				"model":      "llama-2-7b",
				"timeout_ms": 30000,
			},
		},
	}

	for _, event := range events {
		monitor.RecordEvent(event)
		fmt.Printf("   Event [%s]: %s - %s\n", event.Severity, event.Type, event.Message)
	}

	// Record cost entries
	fmt.Println("\n4. Recording cost data...")
	costs := []observability.CostEntry{
		{
			ID:         "cost-inference-1",
			Operation:  "inference",
			ModelID:    "gpt-3.5-turbo",
			Duration:   30 * time.Minute,
			TokensUsed: 150000,
			GPUHours:   0.5,
			Cost:       2.50,
			Currency:   "USD",
		},
		{
			ID:         "cost-training-1",
			Operation:  "training",
			ModelID:    "bert-large",
			Duration:   4 * time.Hour,
			TokensUsed: 0,
			GPUHours:   16.0, // 4 GPUs for 4 hours
			Cost:       48.00,
			Currency:   "USD",
		},
		{
			ID:         "cost-inference-2",
			Operation:  "inference",
			ModelID:    "llama-2-7b",
			Duration:   1 * time.Hour,
			TokensUsed: 500000,
			GPUHours:   2.0,
			Cost:       10.00,
			Currency:   "USD",
		},
		{
			ID:         "cost-training-2",
			Operation:  "training",
			ModelID:    "gpt-3.5-turbo",
			Duration:   8 * time.Hour,
			TokensUsed: 0,
			GPUHours:   64.0, // 8 GPUs for 8 hours
			Cost:       192.00,
			Currency:   "USD",
		},
	}

	for _, cost := range costs {
		monitor.RecordCost(cost)
		fmt.Printf("   Cost: %s - %s ($%.2f, %.1f GPU hours)\n",
			cost.Operation, cost.ModelID, cost.Cost, cost.GPUHours)
	}

	// Get cost summary
	fmt.Println("\n5. Cost Summary Analysis:")
	now := time.Now()
	summary := monitor.GetCostSummary(now.Add(-24*time.Hour), now)

	fmt.Printf("   Period: Last 24 hours\n")
	fmt.Printf("   Total Cost: $%.2f\n", summary["total_cost"])
	fmt.Printf("   Inference Cost: $%.2f\n", summary["inference_cost"])
	fmt.Printf("   Training Cost: $%.2f\n", summary["training_cost"])
	fmt.Printf("   Total Tokens: %d\n", summary["total_tokens"])
	fmt.Printf("   Total GPU Hours: %.2f\n", summary["total_gpu_hours"])

	opCounts := summary["operation_counts"].(map[string]int)
	fmt.Printf("   Operation Breakdown:\n")
	for op, count := range opCounts {
		fmt.Printf("      %s: %d operations\n", op, count)
	}

	// Calculate cost per operation type
	if summary["total_cost"].(float64) > 0 {
		inferencePct := summary["inference_cost"].(float64) / summary["total_cost"].(float64) * 100
		trainingPct := summary["training_cost"].(float64) / summary["total_cost"].(float64) * 100
		fmt.Printf("   Cost Distribution:\n")
		fmt.Printf("      Inference: %.1f%%\n", inferencePct)
		fmt.Printf("      Training: %.1f%%\n", trainingPct)
	}

	// Distributed tracing example
	fmt.Println("\n6. Distributed Tracing Demo:")
	traces := []struct {
		id        string
		operation string
		duration  time.Duration
	}{
		{"trace-req-1", "inference_pipeline", 120 * time.Millisecond},
		{"trace-req-2", "model_loading", 500 * time.Millisecond},
		{"trace-req-3", "batch_processing", 80 * time.Millisecond},
	}

	for _, t := range traces {
		// Start trace
		debugger.StartTrace(t.id, t.operation, map[string]string{
			"user":    "demo-user",
			"service": "ai-platform",
		})

		// Simulate operations with logs
		debugger.AddTraceLog(t.id, observability.DebugLevelInfo,
			fmt.Sprintf("Started %s", t.operation), nil)

		time.Sleep(t.duration / 4)

		debugger.AddTraceLog(t.id, observability.DebugLevelDebug,
			"Processing step 1", map[string]interface{}{
				"progress": "25%",
			})

		time.Sleep(t.duration / 4)

		debugger.AddTraceLog(t.id, observability.DebugLevelDebug,
			"Processing step 2", map[string]interface{}{
				"progress": "50%",
			})

		time.Sleep(t.duration / 4)

		debugger.AddTraceLog(t.id, observability.DebugLevelInfo,
			fmt.Sprintf("Completed %s", t.operation), map[string]interface{}{
				"duration_ms": t.duration.Milliseconds(),
			})

		time.Sleep(t.duration / 4)

		// End trace
		debugger.EndTrace(t.id, "success")

		fmt.Printf("   Trace %s: %s (duration: %vms)\n",
			t.id, t.operation, t.duration.Milliseconds())
	}

	// Analyze traces
	fmt.Println("\n7. Trace Analysis:")
	allTraces := debugger.GetTraces()
	fmt.Printf("   Total Traces: %d\n", len(allTraces))

	for _, trace := range allTraces {
		fmt.Printf("\n   Trace: %s\n", trace.ID)
		fmt.Printf("      Operation: %s\n", trace.Operation)
		fmt.Printf("      Duration: %vms\n", trace.Duration.Milliseconds())
		fmt.Printf("      Status: %s\n", trace.Status)
		fmt.Printf("      Log Entries: %d\n", len(trace.Logs))
		if len(trace.Tags) > 0 {
			fmt.Printf("      Tags:\n")
			for k, v := range trace.Tags {
				fmt.Printf("         %s: %s\n", k, v)
			}
		}
	}

	// Performance analysis
	fmt.Println("\n8. Performance Analysis:")
	perfStats := debugger.AnalyzePerformance()
	for operation, stats := range perfStats {
		statsMap := stats.(map[string]interface{})
		fmt.Printf("   %s:\n", operation)
		fmt.Printf("      Count: %v\n", statsMap["count"])
		fmt.Printf("      Avg Duration: %vms\n", statsMap["avg_duration"])
		fmt.Printf("      Max Duration: %vms\n", statsMap["max_duration"])
	}

	// Get latency statistics
	fmt.Println("\n9. Latency Statistics:")
	latencyStats := monitor.GetLatencyStats("inference_latency_ms", 1*time.Hour)
	for key, value := range latencyStats {
		fmt.Printf("   %s: %v\n", key, value)
	}

	// System health check
	fmt.Println("\n10. System Health Check:")
	health := monitor.GetSystemHealth()
	for key, value := range health {
		fmt.Printf("   %s: %v\n", key, value)
	}

	// Debug statistics
	fmt.Println("\n11. Debug Statistics:")
	debugStats := debugger.GetDebugStats()
	for key, value := range debugStats {
		fmt.Printf("   %s: %v\n", key, value)
	}

	// Query specific events
	fmt.Println("\n12. Querying Events by Severity:")
	severities := []string{"error", "warn", "info"}
	for _, severity := range severities {
		filteredEvents := monitor.GetEvents(now.Add(-24*time.Hour), now, severity)
		fmt.Printf("   %s events: %d\n", severity, len(filteredEvents))
		for _, event := range filteredEvents {
			fmt.Printf("      - %s: %s\n", event.Type, event.Message)
		}
	}

	// Query specific metrics
	fmt.Println("\n13. Querying Metrics:")
	metricNames := []string{"inference_latency_ms", "gpu_utilization_percent"}
	for _, name := range metricNames {
		filteredMetrics := monitor.GetMetrics(now.Add(-1*time.Hour), now, name)
		if len(filteredMetrics) > 0 {
			fmt.Printf("   %s: %d measurements\n", name, len(filteredMetrics))
			total := 0.0
			for _, m := range filteredMetrics {
				total += m.Value
			}
			avg := total / float64(len(filteredMetrics))
			fmt.Printf("      Average: %.2f\n", avg)
		}
	}

	// Demonstrate debug logging
	fmt.Println("\n14. Debug Logging Example:")
	logLevels := []observability.DebugLevel{
		observability.DebugLevelTrace,
		observability.DebugLevelDebug,
		observability.DebugLevelInfo,
		observability.DebugLevelWarn,
		observability.DebugLevelError,
	}

	for _, level := range logLevels {
		debugger.Log(level, "example_service",
			fmt.Sprintf("This is a %s level message", level),
			map[string]interface{}{
				"level": string(level),
			})
		fmt.Printf("   Logged %s message\n", level)
	}

	// Get filtered logs
	fmt.Println("\n15. Retrieving Debug Logs:")
	allLogs := debugger.GetLogs("", now.Add(-1*time.Hour), now)
	fmt.Printf("   Total logs: %d\n", len(allLogs))

	errorLogs := debugger.GetLogs(observability.DebugLevelError, now.Add(-1*time.Hour), now)
	fmt.Printf("   Error logs: %d\n", len(errorLogs))

	fmt.Println("\n=== Example Complete ===")
	fmt.Println("\nKey Takeaways:")
	fmt.Println("  - Comprehensive cost tracking across operations")
	fmt.Println("  - Distributed tracing for performance analysis")
	fmt.Println("  - Multi-level debugging and logging")
	fmt.Println("  - Real-time metrics collection and analysis")
	fmt.Println("  - Event tracking with severity filtering")
}
