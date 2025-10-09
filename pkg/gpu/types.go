package gpu

import "time"

// GPU represents a single GPU resource
type GPU struct {
	ID              string
	Name            string
	MemoryTotal     uint64 // in MB
	MemoryUsed      uint64 // in MB
	Utilization     float64 // 0-100%
	Temperature     float64
	PowerUsage      float64
	Available       bool
	CurrentWorkload *Workload
}

// Workload represents a task that requires GPU resources
type Workload struct {
	ID              string
	Name            string
	Priority        int
	MemoryRequired  uint64
	EstimatedTime   time.Duration
	Status          WorkloadStatus
	AssignedGPU     string
	SubmittedAt     time.Time
	StartedAt       *time.Time
	CompletedAt     *time.Time
}

// WorkloadStatus represents the current state of a workload
type WorkloadStatus string

const (
	WorkloadPending   WorkloadStatus = "pending"
	WorkloadRunning   WorkloadStatus = "running"
	WorkloadCompleted WorkloadStatus = "completed"
	WorkloadFailed    WorkloadStatus = "failed"
)

// SchedulingStrategy defines how workloads are scheduled
type SchedulingStrategy string

const (
	StrategyRoundRobin     SchedulingStrategy = "round_robin"
	StrategyLeastUtilized  SchedulingStrategy = "least_utilized"
	StrategyBestFit        SchedulingStrategy = "best_fit"
	StrategyPriority       SchedulingStrategy = "priority"
)
