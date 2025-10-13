package gpu

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

// GPUMetrics represents detailed metrics for a single GPU
type GPUMetrics struct {
	GPUID              string    `json:"gpu_id"`
	Name               string    `json:"name"`
	UtilizationGPU     float64   `json:"utilization_gpu"`     // GPU utilization percentage
	UtilizationMemory  float64   `json:"utilization_memory"`  // Memory utilization percentage
	MemoryTotal        uint64    `json:"memory_total"`        // Total memory in MB
	MemoryUsed         uint64    `json:"memory_used"`         // Used memory in MB
	MemoryFree         uint64    `json:"memory_free"`         // Free memory in MB
	Temperature        float64   `json:"temperature"`         // Temperature in Celsius
	PowerDraw          float64   `json:"power_draw"`          // Power draw in Watts
	PowerLimit         float64   `json:"power_limit"`         // Power limit in Watts
	FanSpeed           float64   `json:"fan_speed"`           // Fan speed percentage
	ClockGraphics      uint64    `json:"clock_graphics"`      // Graphics clock in MHz
	ClockMemory        uint64    `json:"clock_memory"`        // Memory clock in MHz
	ProcessCount       int       `json:"process_count"`       // Number of running processes
	EncoderUtilization float64   `json:"encoder_utilization"` // Encoder utilization percentage
	DecoderUtilization float64   `json:"decoder_utilization"` // Decoder utilization percentage
	Timestamp          time.Time `json:"timestamp"`
}

// GPUProcess represents a process running on the GPU
type GPUProcess struct {
	PID         int    `json:"pid"`
	ProcessName string `json:"process_name"`
	MemoryUsed  uint64 `json:"memory_used"` // Memory used by process in MB
	Type        string `json:"type"`        // Process type (C for compute, G for graphics)
}

// MetricsCollector collects real-time GPU metrics
type MetricsCollector struct {
	gpuIDs          []string
	collectInterval time.Duration
	metrics         map[string][]GPUMetrics // GPU ID -> historical metrics
	processes       map[string][]GPUProcess // GPU ID -> running processes
	mu              sync.RWMutex
	ctx             context.Context
	cancel          context.CancelFunc
	running         bool
	callbacks       []func(GPUMetrics)
}

// NewMetricsCollector creates a new GPU metrics collector
func NewMetricsCollector(collectInterval time.Duration) *MetricsCollector {
	ctx, cancel := context.WithCancel(context.Background())
	return &MetricsCollector{
		collectInterval: collectInterval,
		metrics:         make(map[string][]GPUMetrics),
		processes:       make(map[string][]GPUProcess),
		ctx:             ctx,
		cancel:          cancel,
		callbacks:       make([]func(GPUMetrics), 0),
	}
}

// Start begins collecting GPU metrics
func (mc *MetricsCollector) Start() error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if mc.running {
		return fmt.Errorf("metrics collector is already running")
	}

	// Discover available GPUs
	gpus, err := mc.discoverGPUs()
	if err != nil {
		return fmt.Errorf("failed to discover GPUs: %w", err)
	}

	mc.gpuIDs = gpus
	mc.running = true

	// Start collection goroutine
	go mc.collectLoop()

	return nil
}

// Stop stops the metrics collection
func (mc *MetricsCollector) Stop() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if mc.running {
		mc.cancel()
		mc.running = false
	}
}

// RegisterCallback registers a callback function to be called when new metrics are collected
func (mc *MetricsCollector) RegisterCallback(callback func(GPUMetrics)) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.callbacks = append(mc.callbacks, callback)
}

// GetLatestMetrics returns the most recent metrics for all GPUs
func (mc *MetricsCollector) GetLatestMetrics() map[string]GPUMetrics {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	latest := make(map[string]GPUMetrics)
	for gpuID, metricsHistory := range mc.metrics {
		if len(metricsHistory) > 0 {
			latest[gpuID] = metricsHistory[len(metricsHistory)-1]
		}
	}

	return latest
}

// GetMetricsHistory returns historical metrics for a GPU within a time range
func (mc *MetricsCollector) GetMetricsHistory(gpuID string, since time.Time) []GPUMetrics {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	history, exists := mc.metrics[gpuID]
	if !exists {
		return []GPUMetrics{}
	}

	result := make([]GPUMetrics, 0)
	for _, metric := range history {
		if metric.Timestamp.After(since) {
			result = append(result, metric)
		}
	}

	return result
}

// GetRunningProcesses returns the processes currently running on GPUs
func (mc *MetricsCollector) GetRunningProcesses() map[string][]GPUProcess {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	// Create a deep copy to avoid race conditions
	result := make(map[string][]GPUProcess)
	for gpuID, processes := range mc.processes {
		result[gpuID] = append([]GPUProcess{}, processes...)
	}

	return result
}

// collectLoop is the main collection loop
func (mc *MetricsCollector) collectLoop() {
	ticker := time.NewTicker(mc.collectInterval)
	defer ticker.Stop()

	for {
		select {
		case <-mc.ctx.Done():
			return
		case <-ticker.C:
			mc.collectMetrics()
		}
	}
}

// collectMetrics collects metrics for all GPUs
func (mc *MetricsCollector) collectMetrics() {
	for _, gpuID := range mc.gpuIDs {
		metrics, err := mc.collectGPUMetrics(gpuID)
		if err != nil {
			// Log error but continue collecting other GPUs
			continue
		}

		processes, err := mc.collectGPUProcesses(gpuID)
		if err != nil {
			// Processes collection is optional, continue anyway
			processes = []GPUProcess{}
		}

		mc.mu.Lock()

		// Store metrics (keep last 1000 entries per GPU)
		if _, exists := mc.metrics[gpuID]; !exists {
			mc.metrics[gpuID] = make([]GPUMetrics, 0)
		}
		mc.metrics[gpuID] = append(mc.metrics[gpuID], metrics)
		if len(mc.metrics[gpuID]) > 1000 {
			mc.metrics[gpuID] = mc.metrics[gpuID][len(mc.metrics[gpuID])-1000:]
		}
		// Store metrics (keep last MaxMetricsHistory entries per GPU)
		const MaxMetricsHistory = 1000
		if _, exists := mc.metrics[gpuID]; !exists {
			mc.metrics[gpuID] = make([]GPUMetrics, 0)
		}
		mc.metrics[gpuID] = append(mc.metrics[gpuID], metrics)
		if len(mc.metrics[gpuID]) > MaxMetricsHistory {
			mc.metrics[gpuID] = mc.metrics[gpuID][len(mc.metrics[gpuID])-MaxMetricsHistory:]
		}
		// Store processes
		mc.processes[gpuID] = processes

		// Call callbacks
		for _, callback := range mc.callbacks {
			go callback(metrics)
		}

		mc.mu.Unlock()
	}
}

// discoverGPUs discovers available NVIDIA GPUs
func (mc *MetricsCollector) discoverGPUs() ([]string, error) {
	cmd := exec.Command("nvidia-smi", "--query-gpu=index", "--format=csv,noheader,nounits")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("nvidia-smi not available or no GPUs found: %w", err)
	}

	var gpuIDs []string
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		gpuID := strings.TrimSpace(scanner.Text())
		if gpuID != "" {
			gpuIDs = append(gpuIDs, gpuID)
		}
	}

	if len(gpuIDs) == 0 {
		return nil, fmt.Errorf("no GPUs discovered")
	}

	return gpuIDs, nil
}

// collectGPUMetrics collects detailed metrics for a specific GPU
func (mc *MetricsCollector) collectGPUMetrics(gpuID string) (GPUMetrics, error) {
	// Use nvidia-smi to collect comprehensive metrics
	cmd := exec.Command("nvidia-smi",
		fmt.Sprintf("--id=%s", gpuID),
		"--query-gpu=name,utilization.gpu,utilization.memory,memory.total,memory.used,memory.free,temperature.gpu,power.draw,power.limit,fan.speed,clocks.current.graphics,clocks.current.memory,encoder.stats.sessionCount,decoder.stats.sessionCount",
		"--format=csv,noheader,nounits")

	output, err := cmd.Output()
	if err != nil {
		return GPUMetrics{}, fmt.Errorf("failed to collect GPU metrics: %w", err)
	}

	line := strings.TrimSpace(string(output))
	fields := strings.Split(line, ", ")

	if len(fields) < 14 {
		return GPUMetrics{}, fmt.Errorf("unexpected nvidia-smi output format")
	}

	// Parse metrics
	metrics := GPUMetrics{
		GPUID:     gpuID,
		Timestamp: time.Now(),
	}

	// Parse each field with error handling
	metrics.Name = strings.TrimSpace(fields[0])

	if val, err := parseFloat(fields[1]); err == nil {
		metrics.UtilizationGPU = val
	}

	if val, err := parseFloat(fields[2]); err == nil {
		metrics.UtilizationMemory = val
	}

	if val, err := parseUint64(fields[3]); err == nil {
		metrics.MemoryTotal = val
	}

	if val, err := parseUint64(fields[4]); err == nil {
		metrics.MemoryUsed = val
	}

	if val, err := parseUint64(fields[5]); err == nil {
		metrics.MemoryFree = val
	}

	if val, err := parseFloat(fields[6]); err == nil {
		metrics.Temperature = val
	}

	if val, err := parseFloat(fields[7]); err == nil {
		metrics.PowerDraw = val
	}

	if val, err := parseFloat(fields[8]); err == nil {
		metrics.PowerLimit = val
	}

	if val, err := parseFloat(fields[9]); err == nil {
		metrics.FanSpeed = val
	}

	if val, err := parseUint64(fields[10]); err == nil {
		metrics.ClockGraphics = val
	}

	if val, err := parseUint64(fields[11]); err == nil {
		metrics.ClockMemory = val
	}

	if val, err := parseFloat(fields[12]); err == nil {
		metrics.EncoderUtilization = val
	}

	if val, err := parseFloat(fields[13]); err == nil {
		metrics.DecoderUtilization = val
	}

	return metrics, nil
}

// collectGPUProcesses collects information about processes running on a GPU
func (mc *MetricsCollector) collectGPUProcesses(gpuID string) ([]GPUProcess, error) {
	cmd := exec.Command("nvidia-smi",
		fmt.Sprintf("--id=%s", gpuID),
		"--query-compute-apps=pid,name,used_memory",
		"--format=csv,noheader,nounits")

	output, err := cmd.Output()
	if err != nil {
		return []GPUProcess{}, fmt.Errorf("failed to collect GPU processes: %w", err)
	}

	var processes []GPUProcess
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || line == "[Not Supported]" {
			continue
		}

		fields := strings.Split(line, ", ")
		if len(fields) >= 3 {
			process := GPUProcess{
				Type: "C", // Compute process
			}

			if pid, err := strconv.Atoi(strings.TrimSpace(fields[0])); err == nil {
				process.PID = pid
			}

			process.ProcessName = strings.TrimSpace(fields[1])

			if memStr := strings.TrimSpace(fields[2]); memStr != "[Not Supported]" {
				if mem, err := parseUint64(memStr); err == nil {
					process.MemoryUsed = mem
				}
			}

			processes = append(processes, process)
		}
	}

	// Also collect graphics processes
	cmd = exec.Command("nvidia-smi",
		fmt.Sprintf("--id=%s", gpuID),
		"--query-graphics-apps=pid,name,used_memory",
		"--format=csv,noheader,nounits")

	output, err = cmd.Output()
	if err == nil {
		scanner = bufio.NewScanner(strings.NewReader(string(output)))
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || line == "[Not Supported]" {
				continue
			}

			fields := strings.Split(line, ", ")
			if len(fields) >= 3 {
				process := GPUProcess{
					Type: "G", // Graphics process
				}

				if pid, err := strconv.Atoi(strings.TrimSpace(fields[0])); err == nil {
					process.PID = pid
				}

				process.ProcessName = strings.TrimSpace(fields[1])

				if memStr := strings.TrimSpace(fields[2]); memStr != "[Not Supported]" {
					if mem, err := parseUint64(memStr); err == nil {
						process.MemoryUsed = mem
					}
				}

				processes = append(processes, process)
			}
		}
	}

	// Update process count for latest metrics
	if metricsHistory, exists := mc.metrics[gpuID]; exists && len(metricsHistory) > 0 {
		lastMetrics := &mc.metrics[gpuID][len(mc.metrics[gpuID])-1]
		lastMetrics.ProcessCount = len(processes)
	}

	return processes, nil
}

// GetGPUEfficiencyMetrics calculates efficiency metrics for GPU utilization
func (mc *MetricsCollector) GetGPUEfficiencyMetrics(gpuID string, duration time.Duration) map[string]interface{} {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	history := mc.GetMetricsHistory(gpuID, time.Now().Add(-duration))
	if len(history) == 0 {
		return map[string]interface{}{
			"error": "no metrics available",
		}
	}

	// Calculate averages and efficiency metrics
	totalUtilization := 0.0
	totalMemoryUtil := 0.0
	totalPowerEfficiency := 0.0
	maxTemp := 0.0
	minTemp := 1000.0

	for _, metric := range history {
		totalUtilization += metric.UtilizationGPU
		totalMemoryUtil += metric.UtilizationMemory

		// Power efficiency: utilization per watt
		if metric.PowerDraw > 0 {
			totalPowerEfficiency += metric.UtilizationGPU / metric.PowerDraw
		}

		if metric.Temperature > maxTemp {
			maxTemp = metric.Temperature
		}
		if metric.Temperature < minTemp {
			minTemp = metric.Temperature
		}
	}

	count := float64(len(history))
	avgUtilization := totalUtilization / count
	avgMemoryUtil := totalMemoryUtil / count
	avgPowerEfficiency := totalPowerEfficiency / count

	// Calculate idle time percentage
	idleTime := 100.0 - avgUtilization

	return map[string]interface{}{
		"gpu_id":               gpuID,
		"avg_utilization":      avgUtilization,
		"avg_memory_util":      avgMemoryUtil,
		"idle_time_percent":    idleTime,
		"avg_power_efficiency": avgPowerEfficiency, // Utilization per watt
		"max_temperature":      maxTemp,
		"min_temperature":      minTemp,
		"sample_count":         len(history),
		"duration_minutes":     duration.Minutes(),
	}
}

// Helper functions for parsing
func parseFloat(s string) (float64, error) {
	s = strings.TrimSpace(s)
	if s == "[Not Supported]" || s == "" {
		return 0, nil
	}
	return strconv.ParseFloat(s, 64)
}

func parseUint64(s string) (uint64, error) {
	s = strings.TrimSpace(s)
	if s == "[Not Supported]" || s == "" {
		return 0, nil
	}
	return strconv.ParseUint(s, 10, 64)
}

// ExportMetricsJSON exports metrics to JSON format
func (mc *MetricsCollector) ExportMetricsJSON(gpuID string, since time.Time) ([]byte, error) {
	history := mc.GetMetricsHistory(gpuID, since)
	return json.MarshalIndent(history, "", "  ")
}

// GetSystemOverview provides a system-wide GPU overview
func (mc *MetricsCollector) GetSystemOverview() map[string]interface{} {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	totalGPUs := len(mc.gpuIDs)
	activeGPUs := 0
	totalUtilization := 0.0
	totalMemoryUsed := uint64(0)
	totalMemoryAvailable := uint64(0)
	totalProcesses := 0

	for _, gpuID := range mc.gpuIDs {
		if metricsHistory, exists := mc.metrics[gpuID]; exists && len(metricsHistory) > 0 {
			latest := metricsHistory[len(metricsHistory)-1]

			if latest.UtilizationGPU > 5.0 { // Consider >5% as active
				activeGPUs++
			}

			totalUtilization += latest.UtilizationGPU
			totalMemoryUsed += latest.MemoryUsed
			totalMemoryAvailable += latest.MemoryTotal
		}

		if processes, exists := mc.processes[gpuID]; exists {
			totalProcesses += len(processes)
		}
	}

	avgUtilization := 0.0
	if totalGPUs > 0 {
		avgUtilization = totalUtilization / float64(totalGPUs)
	}

	memoryUtilization := 0.0
	if totalMemoryAvailable > 0 {
		memoryUtilization = float64(totalMemoryUsed) / float64(totalMemoryAvailable) * 100
	}

	return map[string]interface{}{
		"total_gpus":          totalGPUs,
		"active_gpus":         activeGPUs,
		"avg_utilization":     avgUtilization,
		"memory_used_mb":      totalMemoryUsed,
		"memory_available_mb": totalMemoryAvailable,
		"memory_utilization":  memoryUtilization,
		"total_processes":     totalProcesses,
		"collection_interval": mc.collectInterval.String(),
		"timestamp":           time.Now(),
	}
}
