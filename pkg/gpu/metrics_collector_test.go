package gpu

import (
	"testing"
	"time"
)

func TestMetricsCollectorBasics(t *testing.T) {
	collector := NewMetricsCollector(1 * time.Second)

	// Test initial state
	if collector.running {
		t.Error("Collector should not be running initially")
	}

	if len(collector.GetLatestMetrics()) != 0 {
		t.Error("Should have no metrics initially")
	}

	// Test stopping before starting
	collector.Stop() // Should not panic
}

func TestMetricsCollectorCallbacks(t *testing.T) {
	collector := NewMetricsCollector(1 * time.Second)

	callbackCalled := false
	collector.RegisterCallback(func(metrics GPUMetrics) {
		callbackCalled = true
	})

	// Simulate metrics collection (without actually starting nvidia-smi)
	testMetrics := GPUMetrics{
		GPUID:          "0",
		Name:           "Test GPU",
		UtilizationGPU: 50.0,
		Temperature:    65.0,
		Timestamp:      time.Now(),
	}

	// Manually trigger callback for testing
	for _, callback := range collector.callbacks {
		callback(testMetrics)
	}

	if !callbackCalled {
		t.Error("Callback should have been called")
	}
}

func TestGPUMetricsHistory(t *testing.T) {
	collector := NewMetricsCollector(1 * time.Second)

	// Test empty history
	history := collector.GetMetricsHistory("test-gpu", time.Now().Add(-1*time.Hour))
	if len(history) != 0 {
		t.Error("History should be empty for non-existent GPU")
	}

	// Manually add test metrics
	testGPUID := "test-gpu"
	now := time.Now()

	testMetrics := []GPUMetrics{
		{
			GPUID:          testGPUID,
			UtilizationGPU: 50.0,
			Temperature:    65.0,
			Timestamp:      now.Add(-30 * time.Minute),
		},
		{
			GPUID:          testGPUID,
			UtilizationGPU: 60.0,
			Temperature:    70.0,
			Timestamp:      now.Add(-15 * time.Minute),
		},
		{
			GPUID:          testGPUID,
			UtilizationGPU: 55.0,
			Temperature:    68.0,
			Timestamp:      now,
		},
	}

	// Manually populate metrics (in real usage, this happens via collectMetrics)
	collector.mu.Lock()
	collector.metrics[testGPUID] = testMetrics
	collector.mu.Unlock()

	// Test history retrieval
	history = collector.GetMetricsHistory(testGPUID, now.Add(-1*time.Hour))
	if len(history) != 3 {
		t.Errorf("Expected 3 metrics, got %d", len(history))
	}

	// Test filtering by time
	history = collector.GetMetricsHistory(testGPUID, now.Add(-20*time.Minute))
	if len(history) != 2 {
		t.Errorf("Expected 2 metrics after filtering, got %d", len(history))
	}
}

func TestGPUEfficiencyMetrics(t *testing.T) {
	collector := NewMetricsCollector(1 * time.Second)

	// Test with no metrics
	efficiency := collector.GetGPUEfficiencyMetrics("nonexistent", time.Hour)
	if _, hasError := efficiency["error"]; !hasError {
		t.Error("Should return error for non-existent GPU")
	}

	// Add test metrics
	testGPUID := "efficiency-test"
	now := time.Now()

	testMetrics := []GPUMetrics{
		{
			GPUID:          testGPUID,
			UtilizationGPU: 80.0,
			PowerDraw:      200.0,
			Temperature:    70.0,
			Timestamp:      now.Add(-2 * time.Hour),
		},
		{
			GPUID:          testGPUID,
			UtilizationGPU: 60.0,
			PowerDraw:      150.0,
			Temperature:    65.0,
			Timestamp:      now.Add(-1 * time.Hour),
		},
		{
			GPUID:          testGPUID,
			UtilizationGPU: 90.0,
			PowerDraw:      250.0,
			Temperature:    75.0,
			Timestamp:      now,
		},
	}

	collector.mu.Lock()
	collector.metrics[testGPUID] = testMetrics
	collector.mu.Unlock()

	// Test efficiency calculation
	efficiency = collector.GetGPUEfficiencyMetrics(testGPUID, 3*time.Hour)

	expectedAvgUtil := (80.0 + 60.0 + 90.0) / 3.0
	if avgUtil, ok := efficiency["avg_utilization"].(float64); !ok || avgUtil != expectedAvgUtil {
		t.Errorf("Expected average utilization %.2f, got %.2f", expectedAvgUtil, avgUtil)
	}

	expectedIdleTime := 100.0 - expectedAvgUtil
	if idleTime, ok := efficiency["idle_time_percent"].(float64); !ok || idleTime != expectedIdleTime {
		t.Errorf("Expected idle time %.2f%%, got %.2f%%", expectedIdleTime, idleTime)
	}

	if maxTemp, ok := efficiency["max_temperature"].(float64); !ok || maxTemp != 75.0 {
		t.Errorf("Expected max temperature 75.0, got %.1f", maxTemp)
	}

	if minTemp, ok := efficiency["min_temperature"].(float64); !ok || minTemp != 65.0 {
		t.Errorf("Expected min temperature 65.0, got %.1f", minTemp)
	}
}

func TestSystemOverview(t *testing.T) {
	collector := NewMetricsCollector(1 * time.Second)

	// Test with no GPUs
	overview := collector.GetSystemOverview()
	if totalGPUs, ok := overview["total_gpus"].(int); !ok || totalGPUs != 0 {
		t.Errorf("Expected 0 total GPUs, got %v", totalGPUs)
	}

	// Add test data for multiple GPUs
	now := time.Now()

	collector.mu.Lock()
	collector.gpuIDs = []string{"0", "1"}
	collector.metrics["0"] = []GPUMetrics{
		{
			GPUID:          "0",
			UtilizationGPU: 80.0,
			MemoryUsed:     8000,
			MemoryTotal:    16000,
			Timestamp:      now,
		},
	}
	collector.metrics["1"] = []GPUMetrics{
		{
			GPUID:          "1",
			UtilizationGPU: 20.0,
			MemoryUsed:     2000,
			MemoryTotal:    16000,
			Timestamp:      now,
		},
	}
	collector.processes["0"] = []GPUProcess{{PID: 1234, ProcessName: "test1"}}
	collector.processes["1"] = []GPUProcess{{PID: 5678, ProcessName: "test2"}}
	collector.mu.Unlock()

	overview = collector.GetSystemOverview()

	if totalGPUs, ok := overview["total_gpus"].(int); !ok || totalGPUs != 2 {
		t.Errorf("Expected 2 total GPUs, got %v", totalGPUs)
	}

	if activeGPUs, ok := overview["active_gpus"].(int); !ok || activeGPUs != 2 {
		t.Errorf("Expected 2 active GPUs (>5%% util), got %v", activeGPUs)
	}

	expectedAvgUtil := (80.0 + 20.0) / 2.0
	if avgUtil, ok := overview["avg_utilization"].(float64); !ok || avgUtil != expectedAvgUtil {
		t.Errorf("Expected average utilization %.1f, got %.1f", expectedAvgUtil, avgUtil)
	}

	if totalProc, ok := overview["total_processes"].(int); !ok || totalProc != 2 {
		t.Errorf("Expected 2 total processes, got %v", totalProc)
	}
}

func TestParseHelperFunctions(t *testing.T) {
	// Test parseFloat
	tests := []struct {
		input    string
		expected float64
		hasError bool
	}{
		{"42.5", 42.5, false},
		{"[Not Supported]", 0.0, false},
		{"", 0.0, false},
		{"  25.0  ", 25.0, false},
		{"invalid", 0.0, true},
	}

	for _, test := range tests {
		result, err := parseFloat(test.input)
		if test.hasError && err == nil {
			t.Errorf("Expected error for input '%s'", test.input)
		}
		if !test.hasError && err != nil {
			t.Errorf("Unexpected error for input '%s': %v", test.input, err)
		}
		if result != test.expected {
			t.Errorf("For input '%s', expected %.1f, got %.1f", test.input, test.expected, result)
		}
	}

	// Test parseUint64
	uintTests := []struct {
		input    string
		expected uint64
		hasError bool
	}{
		{"1024", 1024, false},
		{"[Not Supported]", 0, false},
		{"", 0, false},
		{"  2048  ", 2048, false},
		{"invalid", 0, true},
		{"-123", 0, true}, // Negative numbers should error for uint64
	}

	for _, test := range uintTests {
		result, err := parseUint64(test.input)
		if test.hasError && err == nil {
			t.Errorf("Expected error for input '%s'", test.input)
		}
		if !test.hasError && err != nil {
			t.Errorf("Unexpected error for input '%s': %v", test.input, err)
		}
		if result != test.expected {
			t.Errorf("For input '%s', expected %d, got %d", test.input, test.expected, result)
		}
	}
}

func TestMetricsExportJSON(t *testing.T) {
	collector := NewMetricsCollector(1 * time.Second)

	// Test export with no data
	jsonData, err := collector.ExportMetricsJSON("nonexistent", time.Now().Add(-1*time.Hour))
	if err != nil {
		t.Errorf("Export should not error with empty data: %v", err)
	}

	// Should return empty JSON array
	if string(jsonData) != "[]" {
		t.Errorf("Expected empty JSON array, got: %s", string(jsonData))
	}

	// Add test data
	testGPUID := "export-test"
	now := time.Now()

	testMetrics := []GPUMetrics{
		{
			GPUID:          testGPUID,
			Name:           "Test GPU",
			UtilizationGPU: 75.0,
			Temperature:    68.0,
			MemoryUsed:     4000,
			MemoryTotal:    8000,
			Timestamp:      now,
		},
	}

	collector.mu.Lock()
	collector.metrics[testGPUID] = testMetrics
	collector.mu.Unlock()

	// Test export with data
	jsonData, err = collector.ExportMetricsJSON(testGPUID, now.Add(-1*time.Minute))
	if err != nil {
		t.Errorf("Export failed: %v", err)
	}

	// Should contain valid JSON
	if len(jsonData) == 0 {
		t.Error("Exported JSON should not be empty")
	}

	// Check that it contains expected fields
	jsonStr := string(jsonData)
	expectedFields := []string{"gpu_id", "name", "utilization_gpu", "temperature", "timestamp"}
	for _, field := range expectedFields {
		if !contains(jsonStr, field) {
			t.Errorf("Exported JSON should contain field '%s'", field)
		}
	}
}

// Helper function for string containment check
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[0:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			containsInMiddle(s, substr))))
}

func containsInMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Benchmark tests for performance
func BenchmarkMetricsCollection(b *testing.B) {
	collector := NewMetricsCollector(1 * time.Second)

	// Simulate metrics collection
	for i := 0; i < b.N; i++ {
		metrics := GPUMetrics{
			GPUID:          "benchmark-gpu",
			UtilizationGPU: float64(i % 100),
			Temperature:    65.0 + float64(i%20),
			Timestamp:      time.Now(),
		}

		// Simulate callback processing
		for _, callback := range collector.callbacks {
			callback(metrics)
		}
	}
}

func BenchmarkSystemOverview(b *testing.B) {
	collector := NewMetricsCollector(1 * time.Second)

	// Set up test data
	collector.mu.Lock()
	collector.gpuIDs = []string{"0", "1", "2", "3"}
	now := time.Now()

	for i := 0; i < 4; i++ {
		gpuID := string(rune('0' + i))
		collector.metrics[gpuID] = []GPUMetrics{
			{
				GPUID:          gpuID,
				UtilizationGPU: float64(i * 25),
				MemoryUsed:     uint64(i * 2000),
				MemoryTotal:    8000,
				Timestamp:      now,
			},
		}
	}
	collector.mu.Unlock()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		collector.GetSystemOverview()
	}
}

func BenchmarkEfficiencyMetrics(b *testing.B) {
	collector := NewMetricsCollector(1 * time.Second)

	// Set up test data with 100 metrics points
	testGPUID := "benchmark-gpu"
	now := time.Now()
	metrics := make([]GPUMetrics, 100)

	for i := 0; i < 100; i++ {
		metrics[i] = GPUMetrics{
			GPUID:          testGPUID,
			UtilizationGPU: float64(i%100) + 10.0,
			PowerDraw:      200.0 + float64(i%50),
			Temperature:    65.0 + float64(i%15),
			Timestamp:      now.Add(time.Duration(i) * time.Minute),
		}
	}

	collector.mu.Lock()
	collector.metrics[testGPUID] = metrics
	collector.mu.Unlock()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		collector.GetGPUEfficiencyMetrics(testGPUID, 2*time.Hour)
	}
}
