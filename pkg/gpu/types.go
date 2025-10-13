package gpu

import "time"

// GPU represents a single GPU resource
type GPU struct {
	ID              string
	Name            string
	MemoryTotal     uint64  // in MB
	MemoryUsed      uint64  // in MB
	Utilization     float64 // 0-100%
	Temperature     float64
	PowerUsage      float64
	Available       bool
	CurrentWorkload *Workload

	// Real-time metrics integration
	LastMetricsUpdate time.Time
	MetricsHistory    []GPUMetrics
	ProcessCount      int
	PowerLimit        float64
	FanSpeed          float64
	ClockGraphics     uint64
	ClockMemory       uint64
}

// Workload represents a task that requires GPU resources
type Workload struct {
	ID             string
	Name           string
	Priority       int
	MemoryRequired uint64
	EstimatedTime  time.Duration
	Status         WorkloadStatus
	AssignedGPU    string
	SubmittedAt    time.Time
	StartedAt      *time.Time
	CompletedAt    *time.Time
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
	StrategyRoundRobin    SchedulingStrategy = "round_robin"
	StrategyLeastUtilized SchedulingStrategy = "least_utilized"
	StrategyBestFit       SchedulingStrategy = "best_fit"
	StrategyPriority      SchedulingStrategy = "priority"
)

// GPUStats represents aggregated statistics for a GPU over time
type GPUStats struct {
	GPUID               string        `json:"gpu_id"`
	Period              time.Duration `json:"period"`
	AverageUtilization  float64       `json:"average_utilization"`
	PeakUtilization     float64       `json:"peak_utilization"`
	AverageMemoryUsage  float64       `json:"average_memory_usage"`
	PeakMemoryUsage     uint64        `json:"peak_memory_usage"`
	AverageTemperature  float64       `json:"average_temperature"`
	MaxTemperature      float64       `json:"max_temperature"`
	AveragePowerDraw    float64       `json:"average_power_draw"`
	MaxPowerDraw        float64       `json:"max_power_draw"`
	TotalEnergyConsumed float64       `json:"total_energy_consumed"` // in kWh
	IdleTimePercentage  float64       `json:"idle_time_percentage"`
	EfficiencyScore     float64       `json:"efficiency_score"` // Utilization per watt
	ProcessSwitches     int           `json:"process_switches"`
	UptimeHours         float64       `json:"uptime_hours"`
	StartTime           time.Time     `json:"start_time"`
	EndTime             time.Time     `json:"end_time"`
}

// GPUHealthStatus represents the health status of a GPU
type GPUHealthStatus struct {
	GPUID             string     `json:"gpu_id"`
	Status            string     `json:"status"` // healthy, warning, critical
	Timestamp         time.Time  `json:"timestamp"`
	TemperatureStatus string     `json:"temperature_status"`
	MemoryStatus      string     `json:"memory_status"`
	PowerStatus       string     `json:"power_status"`
	UtilizationStatus string     `json:"utilization_status"`
	Issues            []string   `json:"issues"`
	Recommendations   []string   `json:"recommendations"`
	Alerts            []GPUAlert `json:"alerts"`
}

// GPUAlert represents an alert condition for a GPU
type GPUAlert struct {
	Type         string    `json:"type"`     // temperature, memory, power, utilization
	Severity     string    `json:"severity"` // info, warning, critical
	Message      string    `json:"message"`
	Value        float64   `json:"value"`
	Threshold    float64   `json:"threshold"`
	Timestamp    time.Time `json:"timestamp"`
	Acknowledged bool      `json:"acknowledged"`
}

// ClusterMetrics represents metrics for the entire GPU cluster
type ClusterMetrics struct {
	TotalGPUs          int                        `json:"total_gpus"`
	AvailableGPUs      int                        `json:"available_gpus"`
	ActiveGPUs         int                        `json:"active_gpus"`
	TotalMemoryMB      uint64                     `json:"total_memory_mb"`
	UsedMemoryMB       uint64                     `json:"used_memory_mb"`
	AverageUtilization float64                    `json:"average_utilization"`
	AverageTemperature float64                    `json:"average_temperature"`
	TotalPowerDraw     float64                    `json:"total_power_draw"`
	TotalProcesses     int                        `json:"total_processes"`
	HealthyGPUs        int                        `json:"healthy_gpus"`
	GPUStats           map[string]GPUStats        `json:"gpu_stats"`
	GPUHealth          map[string]GPUHealthStatus `json:"gpu_health"`
	Timestamp          time.Time                  `json:"timestamp"`
}
