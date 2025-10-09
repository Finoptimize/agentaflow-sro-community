package gpu

import (
	"fmt"
	"sync"
	"time"
)

// Scheduler manages GPU resources and schedules workloads
type Scheduler struct {
	gpus            map[string]*GPU
	workloadQueue   []*Workload
	strategy        SchedulingStrategy
	mu              sync.RWMutex
	utilizationGoal float64
}

// NewScheduler creates a new GPU scheduler
func NewScheduler(strategy SchedulingStrategy) *Scheduler {
	return &Scheduler{
		gpus:            make(map[string]*GPU),
		workloadQueue:   make([]*Workload, 0),
		strategy:        strategy,
		utilizationGoal: 80.0, // Target 80% utilization
	}
}

// RegisterGPU adds a GPU to the scheduler
func (s *Scheduler) RegisterGPU(gpu *GPU) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.gpus[gpu.ID] = gpu
}

// SubmitWorkload adds a new workload to the queue
func (s *Scheduler) SubmitWorkload(workload *Workload) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	workload.Status = WorkloadPending
	workload.SubmittedAt = time.Now()
	s.workloadQueue = append(s.workloadQueue, workload)

	return nil
}

// Schedule assigns workloads to GPUs based on the scheduling strategy
func (s *Scheduler) Schedule() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.workloadQueue) == 0 {
		return nil
	}

	switch s.strategy {
	case StrategyLeastUtilized:
		return s.scheduleLeastUtilized()
	case StrategyBestFit:
		return s.scheduleBestFit()
	case StrategyPriority:
		return s.schedulePriority()
	case StrategyRoundRobin:
		return s.scheduleRoundRobin()
	default:
		return s.scheduleLeastUtilized()
	}
}

// scheduleLeastUtilized assigns workloads to the least utilized GPU
func (s *Scheduler) scheduleLeastUtilized() error {
	remaining := make([]*Workload, 0)

	for _, workload := range s.workloadQueue {
		gpu := s.findLeastUtilizedGPU(workload.MemoryRequired)
		if gpu != nil {
			s.assignWorkload(gpu, workload)
		} else {
			remaining = append(remaining, workload)
		}
	}

	s.workloadQueue = remaining
	return nil
}

// scheduleBestFit finds the GPU with just enough resources
func (s *Scheduler) scheduleBestFit() error {
	remaining := make([]*Workload, 0)

	for _, workload := range s.workloadQueue {
		gpu := s.findBestFitGPU(workload.MemoryRequired)
		if gpu != nil {
			s.assignWorkload(gpu, workload)
		} else {
			remaining = append(remaining, workload)
		}
	}

	s.workloadQueue = remaining
	return nil
}

// schedulePriority schedules based on workload priority
func (s *Scheduler) schedulePriority() error {
	// Sort by priority (higher priority first)
	for i := 0; i < len(s.workloadQueue)-1; i++ {
		for j := i + 1; j < len(s.workloadQueue); j++ {
			if s.workloadQueue[i].Priority < s.workloadQueue[j].Priority {
				s.workloadQueue[i], s.workloadQueue[j] = s.workloadQueue[j], s.workloadQueue[i]
			}
		}
	}

	return s.scheduleLeastUtilized()
}

// scheduleRoundRobin distributes workloads evenly across GPUs
func (s *Scheduler) scheduleRoundRobin() error {
	gpuList := make([]*GPU, 0, len(s.gpus))
	for _, gpu := range s.gpus {
		if gpu.Available {
			gpuList = append(gpuList, gpu)
		}
	}

	if len(gpuList) == 0 {
		return nil
	}

	remaining := make([]*Workload, 0)
	gpuIndex := 0

	for _, workload := range s.workloadQueue {
		assigned := false
		for i := 0; i < len(gpuList); i++ {
			gpu := gpuList[gpuIndex]
			gpuIndex = (gpuIndex + 1) % len(gpuList)

			if s.canAssign(gpu, workload) {
				s.assignWorkload(gpu, workload)
				assigned = true
				break
			}
		}

		if !assigned {
			remaining = append(remaining, workload)
		}
	}

	s.workloadQueue = remaining
	return nil
}

// findLeastUtilizedGPU finds the GPU with lowest utilization
func (s *Scheduler) findLeastUtilizedGPU(memoryRequired uint64) *GPU {
	var bestGPU *GPU
	minUtilization := 101.0

	for _, gpu := range s.gpus {
		if s.canAssign(gpu, &Workload{MemoryRequired: memoryRequired}) {
			if gpu.Utilization < minUtilization {
				minUtilization = gpu.Utilization
				bestGPU = gpu
			}
		}
	}

	return bestGPU
}

// findBestFitGPU finds the GPU with just enough free memory
func (s *Scheduler) findBestFitGPU(memoryRequired uint64) *GPU {
	var bestGPU *GPU
	minFreeMemory := uint64(^uint64(0))

	for _, gpu := range s.gpus {
		freeMemory := gpu.MemoryTotal - gpu.MemoryUsed
		if freeMemory >= memoryRequired && freeMemory < minFreeMemory {
			minFreeMemory = freeMemory
			bestGPU = gpu
		}
	}

	return bestGPU
}

// canAssign checks if a workload can be assigned to a GPU
func (s *Scheduler) canAssign(gpu *GPU, workload *Workload) bool {
	if !gpu.Available || gpu.CurrentWorkload != nil {
		return false
	}

	freeMemory := gpu.MemoryTotal - gpu.MemoryUsed
	return freeMemory >= workload.MemoryRequired
}

// assignWorkload assigns a workload to a GPU
func (s *Scheduler) assignWorkload(gpu *GPU, workload *Workload) {
	now := time.Now()
	workload.Status = WorkloadRunning
	workload.AssignedGPU = gpu.ID
	workload.StartedAt = &now

	gpu.CurrentWorkload = workload
	gpu.MemoryUsed += workload.MemoryRequired
}

// GetUtilizationMetrics returns overall GPU utilization statistics
func (s *Scheduler) GetUtilizationMetrics() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	totalGPUs := len(s.gpus)
	activeGPUs := 0
	totalUtilization := 0.0
	totalMemoryUsed := uint64(0)
	totalMemoryAvailable := uint64(0)

	for _, gpu := range s.gpus {
		if gpu.CurrentWorkload != nil {
			activeGPUs++
		}
		totalUtilization += gpu.Utilization
		totalMemoryUsed += gpu.MemoryUsed
		totalMemoryAvailable += gpu.MemoryTotal
	}

	avgUtilization := 0.0
	if totalGPUs > 0 {
		avgUtilization = totalUtilization / float64(totalGPUs)
	}

	return map[string]interface{}{
		"total_gpus":          totalGPUs,
		"active_gpus":         activeGPUs,
		"average_utilization": avgUtilization,
		"memory_used_mb":      totalMemoryUsed,
		"memory_available_mb": totalMemoryAvailable,
		"memory_utilization":  float64(totalMemoryUsed) / float64(totalMemoryAvailable) * 100,
		"pending_workloads":   len(s.workloadQueue),
		"utilization_goal":    s.utilizationGoal,
	}
}

// CompleteWorkload marks a workload as completed and frees GPU resources
func (s *Scheduler) CompleteWorkload(workloadID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, gpu := range s.gpus {
		if gpu.CurrentWorkload != nil && gpu.CurrentWorkload.ID == workloadID {
			now := time.Now()
			gpu.CurrentWorkload.CompletedAt = &now
			gpu.CurrentWorkload.Status = WorkloadCompleted
			gpu.MemoryUsed -= gpu.CurrentWorkload.MemoryRequired
			gpu.CurrentWorkload = nil
			return nil
		}
	}

	return fmt.Errorf("workload %s not found", workloadID)
}

// GetGPUStatus returns the current status of all GPUs
func (s *Scheduler) GetGPUStatus() []*GPU {
	s.mu.RLock()
	defer s.mu.RUnlock()

	gpus := make([]*GPU, 0, len(s.gpus))
	for _, gpu := range s.gpus {
		gpus = append(gpus, gpu)
	}
	return gpus
}
