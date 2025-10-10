package main

import (
	"fmt"
	"time"

	"github.com/Finoptimize/agentaflow-sro-community/pkg/gpu"
)

func runGPUSchedulingExample() {
	fmt.Println("=== GPU Scheduling Example ===")

	// Create scheduler with different strategies
	fmt.Println("1. Creating scheduler with Least Utilized strategy...")
	scheduler := gpu.NewScheduler(gpu.StrategyLeastUtilized)

	// Register multiple GPUs
	fmt.Println("2. Registering GPUs...")
	gpus := []*gpu.GPU{
		{
			ID:          "gpu-0",
			Name:        "NVIDIA A100",
			MemoryTotal: 40960, // 40GB
			MemoryUsed:  0,
			Utilization: 0,
			Temperature: 45.0,
			PowerUsage:  150.0,
			Available:   true,
		},
		{
			ID:          "gpu-1",
			Name:        "NVIDIA A100",
			MemoryTotal: 40960,
			MemoryUsed:  10240, // 10GB already in use
			Utilization: 25.0,
			Temperature: 50.0,
			PowerUsage:  200.0,
			Available:   true,
		},
		{
			ID:          "gpu-2",
			Name:        "NVIDIA V100",
			MemoryTotal: 32768, // 32GB
			MemoryUsed:  0,
			Utilization: 0,
			Temperature: 42.0,
			PowerUsage:  140.0,
			Available:   true,
		},
	}

	for _, g := range gpus {
		scheduler.RegisterGPU(g)
		fmt.Printf("   Registered %s (%s): %dMB total, %dMB used\n",
			g.ID, g.Name, g.MemoryTotal, g.MemoryUsed)
	}

	// Create and submit workloads
	fmt.Println("\n3. Submitting workloads...")
	workloads := []*gpu.Workload{
		{
			ID:             "training-bert",
			Name:           "BERT Model Training",
			Priority:       2,
			MemoryRequired: 24576, // 24GB
			EstimatedTime:  2 * time.Hour,
		},
		{
			ID:             "inference-gpt",
			Name:           "GPT-3 Inference",
			Priority:       3,
			MemoryRequired: 16384, // 16GB
			EstimatedTime:  30 * time.Minute,
		},
		{
			ID:             "finetuning-llama",
			Name:           "LLaMA Fine-tuning",
			Priority:       1,
			MemoryRequired: 20480, // 20GB
			EstimatedTime:  90 * time.Minute,
		},
		{
			ID:             "training-resnet",
			Name:           "ResNet Training",
			Priority:       1,
			MemoryRequired: 8192, // 8GB
			EstimatedTime:  1 * time.Hour,
		},
	}

	for _, w := range workloads {
		scheduler.SubmitWorkload(w)
		fmt.Printf("   Submitted: %s (Priority: %d, Memory: %dMB)\n",
			w.Name, w.Priority, w.MemoryRequired)
	}

	// Schedule workloads
	fmt.Println("\n4. Scheduling workloads...")
	err := scheduler.Schedule()
	if err != nil {
		fmt.Printf("Error scheduling: %v\n", err)
		return
	}

	// Display results
	fmt.Println("\n5. GPU Assignment Results:")
	for _, g := range scheduler.GetGPUStatus() {
		fmt.Printf("\n   %s (%s):\n", g.ID, g.Name)
		fmt.Printf("      Memory: %dMB / %dMB (%.1f%% used)\n",
			g.MemoryUsed, g.MemoryTotal,
			float64(g.MemoryUsed)/float64(g.MemoryTotal)*100)
		fmt.Printf("      Utilization: %.1f%%\n", g.Utilization)
		fmt.Printf("      Temperature: %.1f°C\n", g.Temperature)
		if g.CurrentWorkload != nil {
			fmt.Printf("      Current Workload: %s\n", g.CurrentWorkload.Name)
			fmt.Printf("      Workload Priority: %d\n", g.CurrentWorkload.Priority)
		} else {
			fmt.Printf("      Status: Idle\n")
		}
	}

	// Display utilization metrics
	fmt.Println("\n6. Cluster Utilization Metrics:")
	metrics := scheduler.GetUtilizationMetrics()
	fmt.Printf("   Total GPUs: %v\n", metrics["total_gpus"])
	fmt.Printf("   Active GPUs: %v\n", metrics["active_gpus"])
	fmt.Printf("   Average Utilization: %.2f%%\n", metrics["average_utilization"])
	fmt.Printf("   Memory Usage: %vMB / %vMB\n",
		metrics["memory_used_mb"], metrics["memory_available_mb"])
	fmt.Printf("   Memory Utilization: %.2f%%\n", metrics["memory_utilization"])
	fmt.Printf("   Pending Workloads: %v\n", metrics["pending_workloads"])
	fmt.Printf("   Utilization Goal: %.2f%%\n", metrics["utilization_goal"])

	// Demonstrate different scheduling strategies
	fmt.Println("\n7. Testing different scheduling strategies...")

	strategies := []gpu.SchedulingStrategy{
		gpu.StrategyBestFit,
		gpu.StrategyPriority,
		gpu.StrategyRoundRobin,
	}

	for _, strategy := range strategies {
		fmt.Printf("\n   Testing %s strategy:\n", strategy)
		testScheduler := gpu.NewScheduler(strategy)

		// Register fresh GPUs
		for _, g := range gpus {
			newGPU := &gpu.GPU{
				ID:          g.ID,
				Name:        g.Name,
				MemoryTotal: g.MemoryTotal,
				MemoryUsed:  0,
				Utilization: 0,
				Available:   true,
			}
			testScheduler.RegisterGPU(newGPU)
		}

		// Submit same workloads
		for _, w := range workloads {
			newWorkload := &gpu.Workload{
				ID:             w.ID,
				Name:           w.Name,
				Priority:       w.Priority,
				MemoryRequired: w.MemoryRequired,
				EstimatedTime:  w.EstimatedTime,
			}
			testScheduler.SubmitWorkload(newWorkload)
		}

		// Schedule and show results
		testScheduler.Schedule()
		testMetrics := testScheduler.GetUtilizationMetrics()

		assigned := testMetrics["total_gpus"].(int) - testMetrics["pending_workloads"].(int)
		fmt.Printf("      Workloads Assigned: %d/%d\n", assigned, len(workloads))
		fmt.Printf("      Memory Utilization: %.2f%%\n", testMetrics["memory_utilization"])
	}

	fmt.Println("\n=== Example Complete ===")
}
