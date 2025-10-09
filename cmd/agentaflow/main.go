package main

import (
	"fmt"
	"log"
	"time"

	"github.com/Finoptimize/agentaflow-sro-community/pkg/gpu"
	"github.com/Finoptimize/agentaflow-sro-community/pkg/observability"
	"github.com/Finoptimize/agentaflow-sro-community/pkg/serving"
)

func main() {
	fmt.Println("=== AgentaFlow SRO - AI Infrastructure Tooling ===")
	fmt.Println()

	// Initialize monitoring
	monitor := observability.NewMonitoringService(10000)
	debugger := observability.NewDebugger(observability.DebugLevelInfo)

	// Demo 1: GPU Orchestration and Scheduling
	fmt.Println("1. GPU Orchestration & Scheduling Demo")
	fmt.Println("----------------------------------------")
	demoGPUScheduling(monitor, debugger)
	fmt.Println()

	// Demo 2: AI Model Serving Optimization
	fmt.Println("2. AI Model Serving Optimization Demo")
	fmt.Println("----------------------------------------")
	demoModelServing(monitor, debugger)
	fmt.Println()

	// Demo 3: Observability Tools
	fmt.Println("3. Observability & Cost Tracking Demo")
	fmt.Println("----------------------------------------")
	demoObservability(monitor, debugger)
	fmt.Println()

	fmt.Println("=== Demo Complete ===")
}

func demoGPUScheduling(monitor *observability.MonitoringService, debugger *observability.Debugger) {
	// Create GPU scheduler with least-utilized strategy
	scheduler := gpu.NewScheduler(gpu.StrategyLeastUtilized)

	// Register GPUs
	gpus := []*gpu.GPU{
		{
			ID:          "gpu-0",
			Name:        "NVIDIA A100",
			MemoryTotal: 40960, // 40GB
			MemoryUsed:  0,
			Utilization: 0,
			Available:   true,
		},
		{
			ID:          "gpu-1",
			Name:        "NVIDIA A100",
			MemoryTotal: 40960,
			MemoryUsed:  0,
			Utilization: 0,
			Available:   true,
		},
		{
			ID:          "gpu-2",
			Name:        "NVIDIA A100",
			MemoryTotal: 40960,
			MemoryUsed:  0,
			Utilization: 0,
			Available:   true,
		},
	}

	for _, g := range gpus {
		scheduler.RegisterGPU(g)
		debugger.Log(observability.DebugLevelInfo, "gpu_scheduler",
			fmt.Sprintf("Registered GPU: %s", g.ID), nil)
	}

	// Submit workloads
	workloads := []*gpu.Workload{
		{
			ID:             "workload-1",
			Name:           "Training Job - BERT",
			Priority:       1,
			MemoryRequired: 32768, // 32GB
			EstimatedTime:  2 * time.Hour,
		},
		{
			ID:             "workload-2",
			Name:           "Inference - GPT",
			Priority:       2,
			MemoryRequired: 16384, // 16GB
			EstimatedTime:  30 * time.Minute,
		},
		{
			ID:             "workload-3",
			Name:           "Fine-tuning - LLaMA",
			Priority:       1,
			MemoryRequired: 24576, // 24GB
			EstimatedTime:  1 * time.Hour,
		},
	}

	for _, w := range workloads {
		scheduler.SubmitWorkload(w)
		monitor.RecordEvent(observability.Event{
			ID:       fmt.Sprintf("event-%s", w.ID),
			Type:     "workload_submitted",
			Severity: "info",
			Message:  fmt.Sprintf("Workload %s submitted", w.Name),
			Source:   "gpu_scheduler",
		})
	}

	// Schedule workloads
	scheduler.Schedule()

	// Get and display metrics
	metrics := scheduler.GetUtilizationMetrics()
	fmt.Printf("GPU Utilization Metrics:\n")
	for key, value := range metrics {
		fmt.Printf("  %s: %v\n", key, value)
	}

	// Record metrics
	monitor.RecordMetric(observability.Metric{
		Name:   "gpu_utilization",
		Type:   observability.MetricGauge,
		Value:  metrics["average_utilization"].(float64),
		Labels: map[string]string{"cluster": "main"},
	})

	// Show GPU status
	fmt.Printf("\nGPU Status:\n")
	for _, g := range scheduler.GetGPUStatus() {
		fmt.Printf("  %s: Memory Used: %d/%d MB, ", g.ID, g.MemoryUsed, g.MemoryTotal)
		if g.CurrentWorkload != nil {
			fmt.Printf("Running: %s\n", g.CurrentWorkload.Name)
		} else {
			fmt.Printf("Idle\n")
		}
	}
}

func demoModelServing(monitor *observability.MonitoringService, debugger *observability.Debugger) {
	// Create serving manager with batching and caching
	batchConfig := &serving.BatchConfig{
		MaxBatchSize: 32,
		MaxWaitTime:  100 * time.Millisecond,
		MinBatchSize: 1,
	}
	servingMgr := serving.NewServingManager(batchConfig, 5*time.Minute)

	// Register models
	models := []*serving.Model{
		{
			ID:         "model-gpt",
			Name:       "GPT-3.5-Turbo",
			Version:    "v1.0",
			Framework:  "PyTorch",
			MemorySize: 12288,
		},
		{
			ID:         "model-bert",
			Name:       "BERT-Large",
			Version:    "v2.1",
			Framework:  "TensorFlow",
			MemorySize: 4096,
		},
	}

	for _, m := range models {
		servingMgr.RegisterModel(m)
		debugger.Log(observability.DebugLevelInfo, "serving_manager",
			fmt.Sprintf("Registered model: %s", m.Name), nil)
	}

	// Create router
	router := serving.NewRouter(serving.RouteLeastLatency)

	// Register model instances
	instances := []*serving.ModelInstance{
		{
			ID:             "instance-1",
			ModelID:        "model-gpt",
			Endpoint:       "http://localhost:8001",
			MaxLoad:        100,
			CurrentLoad:    0,
			AverageLatency: 50 * time.Millisecond,
			Available:      true,
		},
		{
			ID:             "instance-2",
			ModelID:        "model-gpt",
			Endpoint:       "http://localhost:8002",
			MaxLoad:        100,
			CurrentLoad:    0,
			AverageLatency: 60 * time.Millisecond,
			Available:      true,
		},
	}

	for _, inst := range instances {
		router.RegisterInstance(inst)
	}

	// Submit inference requests
	requests := []*serving.InferenceRequest{
		{
			ID:       "req-1",
			ModelID:  "model-gpt",
			Input:    []byte("What is AI?"),
			Priority: 1,
		},
		{
			ID:       "req-2",
			ModelID:  "model-gpt",
			Input:    []byte("What is AI?"), // Same as req-1, should hit cache
			Priority: 1,
		},
		{
			ID:       "req-3",
			ModelID:  "model-bert",
			Input:    []byte("Classify this text"),
			Priority: 2,
		},
	}

	fmt.Println("Processing inference requests...")
	for _, req := range requests {
		resp, err := servingMgr.SubmitInferenceRequest(req)
		if err != nil {
			log.Printf("Error processing request %s: %v", req.ID, err)
			continue
		}

		cacheStatus := "miss"
		if resp.CacheHit {
			cacheStatus = "hit"
		}

		fmt.Printf("  Request %s: Latency=%vms, Cache=%s\n",
			req.ID, resp.Latency.Milliseconds(), cacheStatus)

		// Record metrics
		monitor.RecordMetric(observability.Metric{
			Name:  "inference_latency",
			Type:  observability.MetricHistogram,
			Value: float64(resp.Latency.Milliseconds()),
			Labels: map[string]string{
				"model": req.ModelID,
				"cache": cacheStatus,
			},
		})
	}

	// Display serving metrics
	fmt.Printf("\nServing Metrics:\n")
	servingMetrics := servingMgr.GetServingMetrics()
	for key, value := range servingMetrics {
		fmt.Printf("  %s: %v\n", key, value)
	}

	fmt.Printf("\nCache Metrics:\n")
	cacheMetrics := servingMgr.GetCacheMetrics()
	for key, value := range cacheMetrics {
		fmt.Printf("  %s: %v\n", key, value)
	}

	fmt.Printf("\nRouting Metrics:\n")
	routingMetrics := router.GetRoutingMetrics()
	for key, value := range routingMetrics {
		fmt.Printf("  %s: %v\n", key, value)
	}
}

func demoObservability(monitor *observability.MonitoringService, debugger *observability.Debugger) {
	// Record cost entries
	costs := []observability.CostEntry{
		{
			ID:         "cost-1",
			Operation:  "inference",
			ModelID:    "model-gpt",
			Duration:   30 * time.Minute,
			TokensUsed: 150000,
			GPUHours:   0.5,
			Cost:       2.50,
			Currency:   "USD",
		},
		{
			ID:         "cost-2",
			Operation:  "training",
			ModelID:    "model-bert",
			Duration:   2 * time.Hour,
			TokensUsed: 0,
			GPUHours:   6.0,
			Cost:       18.00,
			Currency:   "USD",
		},
		{
			ID:         "cost-3",
			Operation:  "inference",
			ModelID:    "model-gpt",
			Duration:   1 * time.Hour,
			TokensUsed: 300000,
			GPUHours:   1.0,
			Cost:       5.00,
			Currency:   "USD",
		},
	}

	for _, cost := range costs {
		monitor.RecordCost(cost)
	}

	// Get cost summary
	now := time.Now()
	summary := monitor.GetCostSummary(now.Add(-24*time.Hour), now)

	fmt.Println("Cost Summary (Last 24 Hours):")
	fmt.Printf("  Total Cost: $%.2f\n", summary["total_cost"])
	fmt.Printf("  Inference Cost: $%.2f\n", summary["inference_cost"])
	fmt.Printf("  Training Cost: $%.2f\n", summary["training_cost"])
	fmt.Printf("  Total Tokens: %d\n", summary["total_tokens"])
	fmt.Printf("  Total GPU Hours: %.2f\n", summary["total_gpu_hours"])

	// Start a trace
	traceID := "trace-inference-001"
	debugger.StartTrace(traceID, "model_inference", map[string]string{
		"model": "gpt-3.5",
		"user":  "demo",
	})

	debugger.AddTraceLog(traceID, observability.DebugLevelInfo,
		"Starting model inference", map[string]interface{}{
			"batch_size": 1,
		})

	// Simulate some processing
	time.Sleep(50 * time.Millisecond)

	debugger.AddTraceLog(traceID, observability.DebugLevelDebug,
		"Model loaded into memory", nil)

	time.Sleep(30 * time.Millisecond)

	debugger.AddTraceLog(traceID, observability.DebugLevelInfo,
		"Inference completed", map[string]interface{}{
			"latency_ms": 80,
		})

	debugger.EndTrace(traceID, "success")

	// Get trace
	trace, _ := debugger.GetTrace(traceID)
	fmt.Printf("\nTrace Analysis:\n")
	fmt.Printf("  Trace ID: %s\n", trace.ID)
	fmt.Printf("  Operation: %s\n", trace.Operation)
	fmt.Printf("  Duration: %vms\n", trace.Duration.Milliseconds())
	fmt.Printf("  Status: %s\n", trace.Status)
	fmt.Printf("  Log Entries: %d\n", len(trace.Logs))

	// Performance analysis
	fmt.Printf("\nPerformance Analysis:\n")
	perfStats := debugger.AnalyzePerformance()
	for op, stats := range perfStats {
		fmt.Printf("  %s: %v\n", op, stats)
	}

	// System health
	fmt.Printf("\nSystem Health:\n")
	health := monitor.GetSystemHealth()
	for key, value := range health {
		fmt.Printf("  %s: %v\n", key, value)
	}

	// Debug stats
	fmt.Printf("\nDebug Statistics:\n")
	debugStats := debugger.GetDebugStats()
	for key, value := range debugStats {
		fmt.Printf("  %s: %v\n", key, value)
	}
}
