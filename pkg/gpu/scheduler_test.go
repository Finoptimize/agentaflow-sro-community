package gpu

import (
	"testing"
	"time"
)

func TestGPUScheduler(t *testing.T) {
	// Test least-utilized strategy
	scheduler := NewScheduler(StrategyLeastUtilized)

	// Register GPUs
	gpu1 := &GPU{
		ID:          "gpu-0",
		Name:        "NVIDIA A100",
		MemoryTotal: 40960,
		MemoryUsed:  0,
		Utilization: 0,
		Available:   true,
	}
	scheduler.RegisterGPU(gpu1)

	gpu2 := &GPU{
		ID:          "gpu-1",
		Name:        "NVIDIA A100",
		MemoryTotal: 40960,
		MemoryUsed:  10240,
		Utilization: 25.0,
		Available:   true,
	}
	scheduler.RegisterGPU(gpu2)

	// Submit workload
	workload := &Workload{
		ID:             "test-workload",
		Name:           "Test Job",
		Priority:       1,
		MemoryRequired: 8192,
		EstimatedTime:  1 * time.Hour,
	}

	err := scheduler.SubmitWorkload(workload)
	if err != nil {
		t.Fatalf("Failed to submit workload: %v", err)
	}

	// Schedule workloads
	err = scheduler.Schedule()
	if err != nil {
		t.Fatalf("Failed to schedule: %v", err)
	}

	// Verify assignment - should be assigned to gpu-0 (least utilized)
	metrics := scheduler.GetUtilizationMetrics()
	activeGPUs := metrics["active_gpus"].(int)
	if activeGPUs != 1 {
		t.Errorf("Expected 1 active GPU, got %d", activeGPUs)
	}

	// Verify GPU status
	gpus := scheduler.GetGPUStatus()
	var assignedGPU *GPU
	for _, g := range gpus {
		if g.CurrentWorkload != nil {
			assignedGPU = g
			break
		}
	}

	if assignedGPU == nil {
		t.Fatal("Workload was not assigned to any GPU")
	}

	if assignedGPU.ID != "gpu-0" {
		t.Errorf("Expected workload assigned to gpu-0, got %s", assignedGPU.ID)
	}

	if assignedGPU.CurrentWorkload.ID != "test-workload" {
		t.Errorf("Expected workload test-workload, got %s", assignedGPU.CurrentWorkload.ID)
	}
}

func TestWorkloadCompletion(t *testing.T) {
	scheduler := NewScheduler(StrategyLeastUtilized)

	gpu1 := &GPU{
		ID:          "gpu-0",
		Name:        "Test GPU",
		MemoryTotal: 40960,
		MemoryUsed:  0,
		Available:   true,
	}
	scheduler.RegisterGPU(gpu1)

	workload := &Workload{
		ID:             "workload-1",
		MemoryRequired: 8192,
		Priority:       1,
	}

	scheduler.SubmitWorkload(workload)
	scheduler.Schedule()

	// Complete the workload
	err := scheduler.CompleteWorkload("workload-1")
	if err != nil {
		t.Fatalf("Failed to complete workload: %v", err)
	}

	// Verify GPU is freed
	gpus := scheduler.GetGPUStatus()
	if gpus[0].CurrentWorkload != nil {
		t.Error("GPU should be free after workload completion")
	}

	if gpus[0].MemoryUsed != 0 {
		t.Errorf("GPU memory should be freed, got %d", gpus[0].MemoryUsed)
	}
}

func TestBestFitScheduling(t *testing.T) {
	scheduler := NewScheduler(StrategyBestFit)

	// Register GPUs with different free memory
	gpu1 := &GPU{
		ID:          "gpu-0",
		MemoryTotal: 40960,
		MemoryUsed:  30000,
		Available:   true,
	}
	scheduler.RegisterGPU(gpu1)

	gpu2 := &GPU{
		ID:          "gpu-1",
		MemoryTotal: 40960,
		MemoryUsed:  0,
		Available:   true,
	}
	scheduler.RegisterGPU(gpu2)

	// Submit workload that fits best in gpu-0
	workload := &Workload{
		ID:             "test-workload",
		MemoryRequired: 10000,
		Priority:       1,
	}

	scheduler.SubmitWorkload(workload)
	scheduler.Schedule()

	// Should be assigned to gpu-0 (best fit)
	gpus := scheduler.GetGPUStatus()
	var assigned bool
	for _, g := range gpus {
		if g.ID == "gpu-0" && g.CurrentWorkload != nil {
			assigned = true
			break
		}
	}

	if !assigned {
		t.Error("Workload should be assigned to gpu-0 with best fit strategy")
	}
}

func TestPriorityScheduling(t *testing.T) {
	scheduler := NewScheduler(StrategyPriority)

	gpu1 := &GPU{
		ID:          "gpu-0",
		MemoryTotal: 40960,
		Available:   true,
	}
	scheduler.RegisterGPU(gpu1)

	// Submit workloads with different priorities
	lowPriority := &Workload{
		ID:             "low-priority",
		MemoryRequired: 8192,
		Priority:       1,
	}

	highPriority := &Workload{
		ID:             "high-priority",
		MemoryRequired: 8192,
		Priority:       10,
	}

	scheduler.SubmitWorkload(lowPriority)
	scheduler.SubmitWorkload(highPriority)
	scheduler.Schedule()

	// High priority should be scheduled first
	gpus := scheduler.GetGPUStatus()
	if gpus[0].CurrentWorkload == nil {
		t.Fatal("No workload assigned")
	}

	if gpus[0].CurrentWorkload.ID != "high-priority" {
		t.Errorf("Expected high-priority workload, got %s", gpus[0].CurrentWorkload.ID)
	}
}
