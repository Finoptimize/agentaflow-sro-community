package gpu

import "time"

// MetricsCollectorInterface defines the interface that both MetricsCollector and MockMetricsCollector implement
type MetricsCollectorInterface interface {
	// Start begins collecting GPU metrics
	Start() error

	// Stop stops the metrics collection
	Stop()

	// RegisterCallback registers a callback function to be called when new metrics are collected
	RegisterCallback(callback func(GPUMetrics))

	// GetLatestMetrics returns the most recent metrics for all GPUs
	GetLatestMetrics() map[string]GPUMetrics

	// GetMetricsHistory returns historical metrics for a GPU within a time range
	GetMetricsHistory(gpuID string, since time.Time) []GPUMetrics

	// GetRunningProcesses returns the processes currently running on GPUs
	GetRunningProcesses() map[string][]GPUProcess

	// CollectMetrics provides backward compatibility
	CollectMetrics() (*GPUMetrics, error)

	// GetSystemOverview provides a system-wide GPU overview
	GetSystemOverview() map[string]interface{}

	// GetGPUEfficiencyMetrics calculates efficiency metrics for GPU utilization
	GetGPUEfficiencyMetrics(gpuID string, duration time.Duration) map[string]interface{}
}
