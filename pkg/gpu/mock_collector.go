package gpu

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"
)

// MockMetricsCollector simulates GPU metrics for demo purposes
// This allows the demo to run on any machine without requiring NVIDIA GPUs
type MockMetricsCollector struct {
	gpuIDs          []string
	collectInterval time.Duration
	metrics         map[string][]GPUMetrics
	processes       map[string][]GPUProcess
	mu              sync.RWMutex
	ctx             context.Context
	cancel          context.CancelFunc
	running         bool
	callbacks       []func(GPUMetrics)

	// Mock data configuration
	gpuConfigs      map[string]MockGPUConfig
	startTime       time.Time
	simulationSpeed float64 // Speed multiplier for demo (1.0 = real time)
}

// MockGPUConfig defines parameters for simulating individual GPU behavior
type MockGPUConfig struct {
	Name            string
	MemoryTotal     uint64  // in MB
	BaseUtilization float64 // baseline utilization %
	UtilVariance    float64 // utilization variance
	BaseTemperature float64 // baseline temperature in Celsius
	TempVariance    float64 // temperature variance
	PowerLimit      float64 // power limit in Watts
	BasePowerDraw   float64 // baseline power draw in Watts
	FanSpeedBase    float64 // baseline fan speed %
	ClockGraphics   uint64  // graphics clock in MHz
	ClockMemory     uint64  // memory clock in MHz

	// Workload simulation
	WorkloadPatterns []WorkloadPattern
	CurrentPattern   int
	PatternStartTime time.Time
}

// WorkloadPattern defines different types of GPU workloads for realistic simulation
type WorkloadPattern struct {
	Name           string        // Pattern name (e.g., "Training", "Inference", "Idle")
	Duration       time.Duration // How long this pattern runs
	UtilizationMin float64       // Minimum utilization during pattern
	UtilizationMax float64       // Maximum utilization during pattern
	MemoryUsageMin float64       // Minimum memory usage % during pattern
	MemoryUsageMax float64       // Maximum memory usage % during pattern
	TempIncrease   float64       // Additional temperature during load
	PowerIncrease  float64       // Additional power draw during load
}

// NewMockMetricsCollector creates a new mock metrics collector for demo purposes
func NewMockMetricsCollector(collectInterval time.Duration, numGPUs int) *MockMetricsCollector {
	ctx, cancel := context.WithCancel(context.Background())

	collector := &MockMetricsCollector{
		collectInterval: collectInterval,
		metrics:         make(map[string][]GPUMetrics),
		processes:       make(map[string][]GPUProcess),
		ctx:             ctx,
		cancel:          cancel,
		callbacks:       make([]func(GPUMetrics), 0),
		gpuConfigs:      make(map[string]MockGPUConfig),
		startTime:       time.Now(),
		simulationSpeed: 1.0,
	}

	// Generate GPU IDs and configurations
	gpuNames := []string{
		"NVIDIA GeForce RTX 4090", "NVIDIA GeForce RTX 4080", "NVIDIA GeForce RTX 4070 Ti",
		"NVIDIA Tesla V100", "NVIDIA A100-SXM4-40GB", "NVIDIA H100 80GB HBM3",
		"NVIDIA GeForce RTX 3080", "NVIDIA GeForce RTX 3090", "NVIDIA Quadro RTX 6000",
	}

	memoryConfigs := []uint64{24576, 16384, 12288, 32768, 40960, 81920, 10240, 24576, 24576} // MB

	for i := 0; i < numGPUs; i++ {
		gpuID := fmt.Sprintf("gpu-%d", i)
		collector.gpuIDs = append(collector.gpuIDs, gpuID)

		nameIdx := i % len(gpuNames)
		memIdx := i % len(memoryConfigs)

		config := MockGPUConfig{
			Name:             gpuNames[nameIdx],
			MemoryTotal:      memoryConfigs[memIdx],
			BaseUtilization:  float64(10 + rand.Intn(20)),    // 10-30% base
			UtilVariance:     float64(5 + rand.Intn(15)),     // 5-20% variance
			BaseTemperature:  float64(35 + rand.Intn(15)),    // 35-50°C base
			TempVariance:     float64(5 + rand.Intn(10)),     // 5-15°C variance
			PowerLimit:       float64(250 + rand.Intn(200)),  // 250-450W
			BasePowerDraw:    float64(50 + rand.Intn(100)),   // 50-150W base
			FanSpeedBase:     float64(20 + rand.Intn(20)),    // 20-40% base fan
			ClockGraphics:    uint64(1200 + rand.Intn(800)),  // 1200-2000 MHz
			ClockMemory:      uint64(8000 + rand.Intn(4000)), // 8000-12000 MHz
			WorkloadPatterns: generateWorkloadPatterns(),
			CurrentPattern:   0,
			PatternStartTime: time.Now(),
		}

		collector.gpuConfigs[gpuID] = config
	}

	return collector
}

// generateWorkloadPatterns creates realistic workload patterns for simulation
func generateWorkloadPatterns() []WorkloadPattern {
	return []WorkloadPattern{
		{
			Name:           "Idle",
			Duration:       time.Duration(30+rand.Intn(60)) * time.Second,
			UtilizationMin: 0,
			UtilizationMax: 15,
			MemoryUsageMin: 5,
			MemoryUsageMax: 20,
			TempIncrease:   0,
			PowerIncrease:  0,
		},
		{
			Name:           "Light Inference",
			Duration:       time.Duration(45+rand.Intn(90)) * time.Second,
			UtilizationMin: 20,
			UtilizationMax: 45,
			MemoryUsageMin: 30,
			MemoryUsageMax: 55,
			TempIncrease:   8,
			PowerIncrease:  50,
		},
		{
			Name:           "Training",
			Duration:       time.Duration(120+rand.Intn(300)) * time.Second,
			UtilizationMin: 70,
			UtilizationMax: 98,
			MemoryUsageMin: 75,
			MemoryUsageMax: 95,
			TempIncrease:   20,
			PowerIncrease:  120,
		},
		{
			Name:           "Heavy Inference",
			Duration:       time.Duration(60+rand.Intn(120)) * time.Second,
			UtilizationMin: 45,
			UtilizationMax: 75,
			MemoryUsageMin: 50,
			MemoryUsageMax: 70,
			TempIncrease:   12,
			PowerIncrease:  80,
		},
		{
			Name:           "Batch Processing",
			Duration:       time.Duration(180+rand.Intn(240)) * time.Second,
			UtilizationMin: 85,
			UtilizationMax: 100,
			MemoryUsageMin: 80,
			MemoryUsageMax: 98,
			TempIncrease:   25,
			PowerIncrease:  150,
		},
	}
}

// Start begins collecting mock GPU metrics
func (mc *MockMetricsCollector) Start() error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if mc.running {
		return fmt.Errorf("mock metrics collector is already running")
	}

	mc.running = true

	// Initialize metrics history for each GPU
	for _, gpuID := range mc.gpuIDs {
		mc.metrics[gpuID] = make([]GPUMetrics, 0)
		mc.processes[gpuID] = generateMockProcesses(gpuID)
	}

	// Start collection goroutine
	go mc.collectLoop()

	return nil
}

// Stop stops the mock metrics collection
func (mc *MockMetricsCollector) Stop() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if mc.running {
		mc.cancel()
		mc.running = false
	}
}

// RegisterCallback registers a callback function to be called when new metrics are collected
func (mc *MockMetricsCollector) RegisterCallback(callback func(GPUMetrics)) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.callbacks = append(mc.callbacks, callback)
}

// GetLatestMetrics returns the most recent metrics for all GPUs
func (mc *MockMetricsCollector) GetLatestMetrics() map[string]GPUMetrics {
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
func (mc *MockMetricsCollector) GetMetricsHistory(gpuID string, since time.Time) []GPUMetrics {
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
func (mc *MockMetricsCollector) GetRunningProcesses() map[string][]GPUProcess {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	result := make(map[string][]GPUProcess)
	for gpuID, processes := range mc.processes {
		result[gpuID] = append([]GPUProcess{}, processes...)
	}

	return result
}

// collectLoop is the main collection loop for mock data
func (mc *MockMetricsCollector) collectLoop() {
	ticker := time.NewTicker(mc.collectInterval)
	defer ticker.Stop()

	for {
		select {
		case <-mc.ctx.Done():
			return
		case <-ticker.C:
			mc.collectMockMetrics()
		}
	}
}

// collectMockMetrics generates and stores mock metrics for all GPUs
func (mc *MockMetricsCollector) collectMockMetrics() {
	currentTime := time.Now()

	for _, gpuID := range mc.gpuIDs {
		config := mc.gpuConfigs[gpuID]

		// Update workload pattern if needed
		if currentTime.Sub(config.PatternStartTime) >= config.WorkloadPatterns[config.CurrentPattern].Duration {
			config.CurrentPattern = (config.CurrentPattern + 1) % len(config.WorkloadPatterns)
			config.PatternStartTime = currentTime
			mc.gpuConfigs[gpuID] = config
		}

		metrics := mc.generateGPUMetrics(gpuID, config, currentTime)

		mc.mu.Lock()

		// Store metrics (keep last 1000 entries per GPU)
		mc.metrics[gpuID] = append(mc.metrics[gpuID], metrics)
		if len(mc.metrics[gpuID]) > 1000 {
			mc.metrics[gpuID] = mc.metrics[gpuID][len(mc.metrics[gpuID])-1000:]
		}

		// Update processes periodically
		if rand.Float64() < 0.1 { // 10% chance to update processes
			mc.processes[gpuID] = generateMockProcesses(gpuID)
		}

		// Call callbacks
		for _, callback := range mc.callbacks {
			go callback(metrics)
		}

		mc.mu.Unlock()
	}
}

// generateGPUMetrics creates realistic GPU metrics based on configuration and workload pattern
func (mc *MockMetricsCollector) generateGPUMetrics(gpuID string, config MockGPUConfig, timestamp time.Time) GPUMetrics {
	pattern := config.WorkloadPatterns[config.CurrentPattern]

	// Calculate time-based variations
	elapsed := timestamp.Sub(mc.startTime).Seconds() * mc.simulationSpeed

	// Generate utilization based on pattern and some randomness
	utilization := config.BaseUtilization
	utilization += pattern.UtilizationMin + rand.Float64()*(pattern.UtilizationMax-pattern.UtilizationMin)

	// Add some sinusoidal variation for realism
	utilization += math.Sin(elapsed/60.0) * config.UtilVariance * 0.5

	// Add small random noise
	utilization += (rand.Float64() - 0.5) * 5.0

	// Clamp utilization
	if utilization < 0 {
		utilization = 0
	}
	if utilization > 100 {
		utilization = 100
	}

	// Memory utilization correlates with GPU utilization but has its own pattern
	memoryUtilization := pattern.MemoryUsageMin + rand.Float64()*(pattern.MemoryUsageMax-pattern.MemoryUsageMin)
	memoryUtilization += math.Cos(elapsed/45.0) * 5.0
	memoryUsed := uint64(float64(config.MemoryTotal) * memoryUtilization / 100.0)

	// Temperature based on utilization and pattern
	temperature := config.BaseTemperature + pattern.TempIncrease
	temperature += (utilization / 100.0) * 25.0 // Up to 25°C increase at full utilization
	temperature += math.Sin(elapsed/30.0) * config.TempVariance * 0.3
	temperature += (rand.Float64() - 0.5) * 3.0

	// Power draw correlates with utilization
	powerDraw := config.BasePowerDraw + pattern.PowerIncrease
	powerDraw += (utilization / 100.0) * (config.PowerLimit - config.BasePowerDraw) * 0.8
	powerDraw += (rand.Float64() - 0.5) * 15.0

	if powerDraw > config.PowerLimit {
		powerDraw = config.PowerLimit
	}

	// Fan speed based on temperature
	fanSpeed := config.FanSpeedBase
	if temperature > 60 {
		fanSpeed += (temperature - 60) * 2.0 // 2% per degree above 60°C
	}
	if fanSpeed > 100 {
		fanSpeed = 100
	}

	// Clock speeds vary slightly based on temperature and power
	graphicsClock := config.ClockGraphics
	memoryClock := config.ClockMemory

	if temperature > 80 {
		// Thermal throttling
		throttleFactor := 0.95 - (temperature-80)*0.01
		graphicsClock = uint64(float64(graphicsClock) * throttleFactor)
	}

	return GPUMetrics{
		GPUID:              gpuID,
		Name:               config.Name,
		UtilizationGPU:     utilization,
		UtilizationMemory:  memoryUtilization,
		MemoryTotal:        config.MemoryTotal,
		MemoryUsed:         memoryUsed,
		MemoryFree:         config.MemoryTotal - memoryUsed,
		Temperature:        temperature,
		PowerDraw:          powerDraw,
		PowerLimit:         config.PowerLimit,
		FanSpeed:           fanSpeed,
		ClockGraphics:      graphicsClock,
		ClockMemory:        memoryClock,
		ProcessCount:       len(mc.processes[gpuID]),
		EncoderUtilization: utilization * 0.3, // Encoder typically lower
		DecoderUtilization: utilization * 0.2, // Decoder typically lower
		Timestamp:          timestamp,
	}
}

// generateMockProcesses creates mock GPU processes for demo purposes
func generateMockProcesses(gpuID string) []GPUProcess {
	processNames := []string{
		"python", "pytorch_training", "tensorflow_serve", "cuda_inference",
		"blender", "obs-studio", "chrome", "firefox", "nvenc_encoder",
		"stable_diffusion", "llama_inference", "whisper_transcribe",
	}

	processCount := rand.Intn(6) // 0-5 processes
	processes := make([]GPUProcess, processCount)

	for i := 0; i < processCount; i++ {
		nameIdx := rand.Intn(len(processNames))
		processType := "C" // Compute by default
		if rand.Float64() < 0.3 {
			processType = "G" // Graphics
		}

		processes[i] = GPUProcess{
			PID:         1000 + rand.Intn(30000),
			ProcessName: processNames[nameIdx],
			MemoryUsed:  uint64(rand.Intn(8192)), // 0-8GB
			Type:        processType,
		}
	}

	return processes
}

// CollectMetrics provides backward compatibility
func (mc *MockMetricsCollector) CollectMetrics() (*GPUMetrics, error) {
	latest := mc.GetLatestMetrics()

	if len(latest) == 0 {
		return nil, fmt.Errorf("no GPU metrics available")
	}

	// Return the first available metric
	for _, metrics := range latest {
		return &metrics, nil
	}

	return nil, fmt.Errorf("no GPU metrics available")
}

// GetSystemOverview provides a system-wide GPU overview for mock data
func (mc *MockMetricsCollector) GetSystemOverview() map[string]interface{} {
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

			if latest.UtilizationGPU > 5.0 {
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

// GetGPUEfficiencyMetrics calculates efficiency metrics for mock GPU utilization
func (mc *MockMetricsCollector) GetGPUEfficiencyMetrics(gpuID string, duration time.Duration) map[string]interface{} {
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

	idleTime := 100.0 - avgUtilization

	return map[string]interface{}{
		"gpu_id":               gpuID,
		"avg_utilization":      avgUtilization,
		"avg_memory_util":      avgMemoryUtil,
		"idle_time_percent":    idleTime,
		"avg_power_efficiency": avgPowerEfficiency,
		"max_temperature":      maxTemp,
		"min_temperature":      minTemp,
		"sample_count":         len(history),
		"duration_minutes":     duration.Minutes(),
	}
}

// SetSimulationSpeed adjusts the speed of the simulation (for demo purposes)
func (mc *MockMetricsCollector) SetSimulationSpeed(speed float64) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.simulationSpeed = speed
}

// TriggerWorkloadChange forces a specific workload pattern on a GPU (for demo)
func (mc *MockMetricsCollector) TriggerWorkloadChange(gpuID string, patternName string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	config, exists := mc.gpuConfigs[gpuID]
	if !exists {
		return
	}

	for i, pattern := range config.WorkloadPatterns {
		if pattern.Name == patternName {
			config.CurrentPattern = i
			config.PatternStartTime = time.Now()
			mc.gpuConfigs[gpuID] = config
			break
		}
	}
}

// GetCurrentWorkloadPattern returns the current workload pattern for a GPU
func (mc *MockMetricsCollector) GetCurrentWorkloadPattern(gpuID string) string {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	config, exists := mc.gpuConfigs[gpuID]
	if !exists {
		return "Unknown"
	}

	return config.WorkloadPatterns[config.CurrentPattern].Name
}
