package gpu

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestMetricsAggregationServiceBasics(t *testing.T) {
	// Create a mock metrics collector
	collector := NewMetricsCollector(1 * time.Second)
	
	// Create aggregation service
	aggregationService := NewMetricsAggregationService(
		collector,
		1*time.Minute,
		24*time.Hour,
	)
	
	// Test initial state
	if aggregationService.running {
		t.Error("Aggregation service should not be running initially")
	}
	
	// Test getting stats with no data
	stats := aggregationService.GetAllGPUStats()
	if len(stats) != 0 {
		t.Error("Should have no GPU stats initially")
	}
	
	// Test getting cluster metrics with no data
	clusterMetrics := aggregationService.GetClusterMetrics()
	if clusterMetrics != nil {
		t.Error("Should have no cluster metrics initially")
	}
	
	// Test stopping before starting
	aggregationService.Stop() // Should not panic
}

func TestGPUStatsCalculation(t *testing.T) {
	collector := NewMetricsCollector(1 * time.Second)
	aggregationService := NewMetricsAggregationService(
		collector,
		1*time.Minute,
		24*time.Hour,
	)
	
	// Prepare test metrics
	testGPUID := "test-gpu"
	now := time.Now()
	
	testHistory := []GPUMetrics{
		{
			GPUID:          testGPUID,
			UtilizationGPU: 80.0,
			MemoryUsed:     6000,
			Temperature:    70.0,
			PowerDraw:      200.0,
			Timestamp:      now.Add(-3 * time.Hour),
		},
		{
			GPUID:          testGPUID,
			UtilizationGPU: 60.0,
			MemoryUsed:     4000,
			Temperature:    65.0,
			PowerDraw:      150.0,
			Timestamp:      now.Add(-2 * time.Hour),
		},
		{
			GPUID:          testGPUID,
			UtilizationGPU: 90.0,
			MemoryUsed:     8000,
			Temperature:    75.0,
			PowerDraw:      250.0,
			Timestamp:      now.Add(-1 * time.Hour),
		},
		{
			GPUID:          testGPUID,
			UtilizationGPU: 2.0, // Low utilization to create idle time
			MemoryUsed:     2000,
			Temperature:    60.0,
			PowerDraw:      100.0,
			Timestamp:      now,
		},
	}
	
	// Mock the collector to return our test history
	collector.mu.Lock()
	collector.metrics[testGPUID] = testHistory
	collector.mu.Unlock()
	
	// Create and calculate stats
	stats := &GPUStats{GPUID: testGPUID, StartTime: now}
	aggregationService.calculateGPUStatistics(stats, testHistory, now)
	
	// Test calculations
	expectedAvgUtil := (80.0 + 60.0 + 90.0 + 2.0) / 4.0
	if stats.AverageUtilization != expectedAvgUtil {
		t.Errorf("Expected average utilization %.2f, got %.2f", expectedAvgUtil, stats.AverageUtilization)
	}
	
	expectedPeakUtil := 90.0
	if stats.PeakUtilization != expectedPeakUtil {
		t.Errorf("Expected peak utilization %.1f, got %.1f", expectedPeakUtil, stats.PeakUtilization)
	}
	
	expectedMaxTemp := 75.0
	if stats.MaxTemperature != expectedMaxTemp {
		t.Errorf("Expected max temperature %.1f, got %.1f", expectedMaxTemp, stats.MaxTemperature)
	}
	
	expectedPeakMemory := uint64(8000)
	if stats.PeakMemoryUsage != expectedPeakMemory {
		t.Errorf("Expected peak memory usage %d, got %d", expectedPeakMemory, stats.PeakMemoryUsage)
	}
	
	// Test idle time calculation (utilization < 5%)
	if stats.IdleTimePercentage <= 0 {
		t.Error("Idle time percentage should be greater than 0")
	}
	
	// Test efficiency score (utilization per watt)
	expectedEfficiency := expectedAvgUtil / ((200.0 + 150.0 + 250.0 + 100.0) / 4.0)
	if stats.EfficiencyScore != expectedEfficiency {
		t.Errorf("Expected efficiency score %.4f, got %.4f", expectedEfficiency, stats.EfficiencyScore)
	}
	
	// Test energy consumption calculation
	if stats.TotalEnergyConsumed <= 0 {
		t.Error("Total energy consumed should be greater than 0")
	}
	
	// Test uptime calculation
	expectedUptimeHours := 3.0 // 3 hours between first and last metric
	if stats.UptimeHours != expectedUptimeHours {
		t.Errorf("Expected uptime %.1f hours, got %.1f hours", expectedUptimeHours, stats.UptimeHours)
	}
}

func TestClusterMetricsCalculation(t *testing.T) {
	collector := NewMetricsCollector(1 * time.Second)
	aggregationService := NewMetricsAggregationService(
		collector,
		1*time.Minute,
		24*time.Hour,
	)
	
	now := time.Now()
	
	// Create test metrics for multiple GPUs
	latestMetrics := map[string]GPUMetrics{
		"gpu0": {
			GPUID:          "gpu0",
			UtilizationGPU: 80.0,
			MemoryUsed:     6000,
			MemoryTotal:    8000,
			Temperature:    70.0,
			PowerDraw:      200.0,
			ProcessCount:   2,
			Timestamp:      now,
		},
		"gpu1": {
			GPUID:          "gpu1",
			UtilizationGPU: 40.0,
			MemoryUsed:     3000,
			MemoryTotal:    8000,
			Temperature:    65.0,
			PowerDraw:      150.0,
			ProcessCount:   1,
			Timestamp:      now,
		},
		"gpu2": {
			GPUID:          "gpu2",
			UtilizationGPU: 10.0, // Low utilization (not active)
			MemoryUsed:     1000,
			MemoryTotal:    8000,
			Temperature:    60.0,
			PowerDraw:      100.0,
			ProcessCount:   0,
			Timestamp:      now,
		},
	}
	
	// Update cluster metrics
	aggregationService.updateClusterMetrics(latestMetrics, now)
	
	clusterMetrics := aggregationService.GetClusterMetrics()
	if clusterMetrics == nil {
		t.Fatal("Cluster metrics should not be nil")
	}
	
	// Test cluster calculations
	if clusterMetrics.TotalGPUs != 3 {
		t.Errorf("Expected 3 total GPUs, got %d", clusterMetrics.TotalGPUs)
	}
	
	if clusterMetrics.ActiveGPUs != 3 {
		t.Errorf("Expected 3 active GPUs (>5%% util), got %d", clusterMetrics.ActiveGPUs)
	}
	
	expectedAvgUtil := (80.0 + 40.0 + 10.0) / 3.0
	if clusterMetrics.AverageUtilization != expectedAvgUtil {
		t.Errorf("Expected average utilization %.2f, got %.2f", 
			expectedAvgUtil, clusterMetrics.AverageUtilization)
	}
	
	expectedTotalMemory := uint64(24000) // 3 * 8000
	if clusterMetrics.TotalMemoryMB != expectedTotalMemory {
		t.Errorf("Expected total memory %d MB, got %d MB", 
			expectedTotalMemory, clusterMetrics.TotalMemoryMB)
	}
	
	expectedUsedMemory := uint64(10000) // 6000 + 3000 + 1000
	if clusterMetrics.UsedMemoryMB != expectedUsedMemory {
		t.Errorf("Expected used memory %d MB, got %d MB", 
			expectedUsedMemory, clusterMetrics.UsedMemoryMB)
	}
	
	expectedTotalPower := 450.0 // 200 + 150 + 100
	if clusterMetrics.TotalPowerDraw != expectedTotalPower {
		t.Errorf("Expected total power draw %.1f W, got %.1f W", 
			expectedTotalPower, clusterMetrics.TotalPowerDraw)
	}
	
	expectedTotalProcesses := 3 // 2 + 1 + 0
	if clusterMetrics.TotalProcesses != expectedTotalProcesses {
		t.Errorf("Expected total processes %d, got %d", 
			expectedTotalProcesses, clusterMetrics.TotalProcesses)
	}
}

func TestEfficiencyReport(t *testing.T) {
	collector := NewMetricsCollector(1 * time.Second)
	aggregationService := NewMetricsAggregationService(
		collector,
		1*time.Minute,
		24*time.Hour,
	)
	
	// Add test GPU stats
	aggregationService.mu.Lock()
	aggregationService.gpuStats = map[string]*GPUStats{
		"gpu0": {
			GPUID:                "gpu0",
			AverageUtilization:   80.0,
			IdleTimePercentage:   20.0,
			EfficiencyScore:      0.4, // 80/200 watts
		},
		"gpu1": {
			GPUID:                "gpu1",
			AverageUtilization:   60.0,
			IdleTimePercentage:   40.0,
			EfficiencyScore:      0.3, // 60/200 watts
		},
		"gpu2": {
			GPUID:                "gpu2",
			AverageUtilization:   90.0,
			IdleTimePercentage:   10.0,
			EfficiencyScore:      0.45, // 90/200 watts
		},
	}
	aggregationService.mu.Unlock()
	
	report := aggregationService.GetEfficiencyReport()
	
	// Test cluster efficiency
	clusterEff, ok := report["cluster_efficiency"].(map[string]interface{})
	if !ok {
		t.Fatal("Cluster efficiency should be present in report")
	}
	
	expectedAvgIdleTime := (20.0 + 40.0 + 10.0) / 3.0
	if avgIdleTime, ok := clusterEff["average_idle_time_percent"].(float64); !ok || 
		avgIdleTime != expectedAvgIdleTime {
		t.Errorf("Expected average idle time %.2f, got %.2f", expectedAvgIdleTime, avgIdleTime)
	}
	
	expectedAvgEfficiency := (0.4 + 0.3 + 0.45) / 3.0
	if avgEff, ok := clusterEff["average_efficiency_score"].(float64); !ok {
		t.Error("Expected average efficiency score to be present")
	} else {
		// Use approximate comparison for floating point numbers
		diff := avgEff - expectedAvgEfficiency
		if diff < 0 {
			diff = -diff
		}
		if diff > 0.001 {
			t.Errorf("Expected average efficiency %.3f, got %.3f", expectedAvgEfficiency, avgEff)
		}
	}
	
	// Test GPU rankings
	if rankings, ok := report["gpu_rankings"]; ok {
		// Use reflection to check slice length and content
		if reflect.ValueOf(rankings).Kind() == reflect.Slice {
			rankingsValue := reflect.ValueOf(rankings)
			if rankingsValue.Len() != 3 {
				t.Errorf("Expected 3 GPUs in rankings, got %d", rankingsValue.Len())
			}
		} else {
			t.Error("GPU rankings should be a slice")
		}
	} else {
		t.Fatal("GPU rankings should be present in report")
	}
}

func TestPerformanceTrends(t *testing.T) {
	collector := NewMetricsCollector(1 * time.Second)
	aggregationService := NewMetricsAggregationService(
		collector,
		1*time.Minute,
		24*time.Hour,
	)
	
	// Test with no data
	trends := aggregationService.GetPerformanceTrends("nonexistent", time.Hour)
	if _, hasError := trends["error"]; !hasError {
		t.Error("Should return error for non-existent GPU")
	}
	
	// Add test data with clear trends
	testGPUID := "trend-test"
	now := time.Now()
	
	// Create metrics with increasing utilization trend
	testMetrics := make([]GPUMetrics, 10)
	for i := 0; i < 10; i++ {
		testMetrics[i] = GPUMetrics{
			GPUID:          testGPUID,
			UtilizationGPU: float64(i * 10), // 0, 10, 20, ..., 90
			Temperature:    60.0 + float64(i), // Increasing temperature
			MemoryUsed:     uint64(i * 1000),  // Increasing memory usage
			PowerDraw:      100.0 + float64(i*10), // Increasing power
			Timestamp:      now.Add(time.Duration(i) * time.Minute),
		}
	}
	
	collector.mu.Lock()
	collector.metrics[testGPUID] = testMetrics
	collector.mu.Unlock()
	
	trends = aggregationService.GetPerformanceTrends(testGPUID, 2*time.Hour)
	
	// Test that trends are calculated
	if sampleCount, ok := trends["sample_count"].(int); !ok || sampleCount != 10 {
		t.Errorf("Expected 10 samples, got %v", sampleCount)
	}
	
	if utilTrend, ok := trends["utilization_trend"].(map[string]float64); ok {
		// Should have positive slope (increasing trend)
		if slope := utilTrend["slope"]; slope <= 0 {
			t.Errorf("Expected positive utilization slope, got %.3f", slope)
		}
		// Should have high R-squared (good linear fit)
		if rSquared := utilTrend["r_squared"]; rSquared < 0.8 {
			t.Errorf("Expected high R-squared (>0.8), got %.3f", rSquared)
		}
	} else {
		t.Error("Utilization trend should be present")
	}
	
	if tempTrend, ok := trends["temperature_trend"].(map[string]float64); ok {
		// Should have positive slope (increasing trend)
		if slope := tempTrend["slope"]; slope <= 0 {
			t.Errorf("Expected positive temperature slope, got %.3f", slope)
		}
	} else {
		t.Error("Temperature trend should be present")
	}
}

func TestCostAnalysis(t *testing.T) {
	collector := NewMetricsCollector(1 * time.Second)
	aggregationService := NewMetricsAggregationService(
		collector,
		1*time.Minute,
		24*time.Hour,
	)
	
	// Add test GPU stats
	aggregationService.mu.Lock()
	aggregationService.gpuStats = map[string]*GPUStats{
		"gpu0": {
			GPUID:                "gpu0",
			AverageUtilization:   85.0, // High utilization -> high-performance pricing
			UptimeHours:          10.0,
			EfficiencyScore:      0.42,
		},
		"gpu1": {
			GPUID:                "gpu1",
			AverageUtilization:   50.0, // Medium utilization -> medium pricing
			UptimeHours:          8.0,
			EfficiencyScore:      0.25,
		},
		"gpu2": {
			GPUID:                "gpu2",
			AverageUtilization:   15.0, // Low utilization -> basic pricing
			UptimeHours:          12.0,
			EfficiencyScore:      0.15,
		},
	}
	aggregationService.mu.Unlock()
	
	analysis := aggregationService.GetCostAnalysis()
	
	// Test that analysis contains expected fields
	if _, ok := analysis["total_estimated_cost"]; !ok {
		t.Error("Analysis should contain total estimated cost")
	}
	
	if _, ok := analysis["total_potential_savings"]; !ok {
		t.Error("Analysis should contain total potential savings")
	}
	
	if _, ok := analysis["savings_percentage"]; !ok {
		t.Error("Analysis should contain savings percentage")
	}
	
	if gpuCosts, ok := analysis["gpu_costs"].(map[string]interface{}); ok {
		if len(gpuCosts) != 3 {
			t.Errorf("Expected costs for 3 GPUs, got %d", len(gpuCosts))
		}
		
		// Test that each GPU has cost information
		for _, gpuID := range []string{"gpu0", "gpu1", "gpu2"} {
			if costs, exists := gpuCosts[gpuID]; exists {
				if costMap, ok := costs.(map[string]interface{}); ok {
					requiredFields := []string{"actual_cost", "optimized_cost", "potential_savings", "uptime_hours"}
					for _, field := range requiredFields {
						if _, hasField := costMap[field]; !hasField {
							t.Errorf("GPU %s costs should contain field '%s'", gpuID, field)
						}
					}
				}
			}
		}
	} else {
		t.Error("Analysis should contain GPU costs breakdown")
	}
}

func TestTrendCalculation(t *testing.T) {
	collector := NewMetricsCollector(1 * time.Second)
	aggregationService := NewMetricsAggregationService(
		collector,
		1*time.Minute,
		24*time.Hour,
	)
	
	// Test with empty history
	emptyHistory := []GPUMetrics{}
	trend := aggregationService.calculateTrend(emptyHistory, func(m GPUMetrics) float64 {
		return m.UtilizationGPU
	})
	
	if slope := trend["slope"]; slope != 0 {
		t.Errorf("Expected slope 0 for empty history, got %.3f", slope)
	}
	
	// Test with single data point
	singleHistory := []GPUMetrics{
		{UtilizationGPU: 50.0, Timestamp: time.Now()},
	}
	trend = aggregationService.calculateTrend(singleHistory, func(m GPUMetrics) float64 {
		return m.UtilizationGPU
	})
	
	if slope := trend["slope"]; slope != 0 {
		t.Errorf("Expected slope 0 for single data point, got %.3f", slope)
	}
	
	// Test with perfect linear trend
	now := time.Now()
	linearHistory := []GPUMetrics{
		{UtilizationGPU: 10.0, Timestamp: now},
		{UtilizationGPU: 20.0, Timestamp: now.Add(1 * time.Hour)},
		{UtilizationGPU: 30.0, Timestamp: now.Add(2 * time.Hour)},
		{UtilizationGPU: 40.0, Timestamp: now.Add(3 * time.Hour)},
	}
	
	trend = aggregationService.calculateTrend(linearHistory, func(m GPUMetrics) float64 {
		return m.UtilizationGPU
	})
	
	expectedSlope := 10.0 // 10% increase per hour
	if slope := trend["slope"]; slope != expectedSlope {
		t.Errorf("Expected slope %.1f, got %.3f", expectedSlope, slope)
	}
	
	// Perfect linear relationship should have R² = 1.0
	if rSquared := trend["r_squared"]; rSquared < 0.99 {
		t.Errorf("Expected R-squared close to 1.0, got %.3f", rSquared)
	}
}

// Benchmark tests for aggregation performance
func BenchmarkAggregationCalculation(b *testing.B) {
	collector := NewMetricsCollector(1 * time.Second)
	aggregationService := NewMetricsAggregationService(
		collector,
		1*time.Minute,
		24*time.Hour,
	)
	
	// Generate large dataset
	testGPUID := "benchmark-gpu"
	now := time.Now()
	history := make([]GPUMetrics, 1000)
	
	for i := 0; i < 1000; i++ {
		history[i] = GPUMetrics{
			GPUID:          testGPUID,
			UtilizationGPU: float64(i%100) + 10.0,
			MemoryUsed:     uint64(i%8000) + 1000,
			Temperature:    65.0 + float64(i%20),
			PowerDraw:      150.0 + float64(i%100),
			Timestamp:      now.Add(time.Duration(i) * time.Minute),
		}
	}
	
	stats := &GPUStats{GPUID: testGPUID, StartTime: now}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		aggregationService.calculateGPUStatistics(stats, history, now)
	}
}

func BenchmarkEfficiencyReport(b *testing.B) {
	collector := NewMetricsCollector(1 * time.Second)
	aggregationService := NewMetricsAggregationService(
		collector,
		1*time.Minute,
		24*time.Hour,
	)
	
	// Set up test data with many GPUs
	aggregationService.mu.Lock()
	for i := 0; i < 100; i++ {
		gpuID := fmt.Sprintf("gpu%d", i)
		aggregationService.gpuStats[gpuID] = &GPUStats{
			GPUID:                gpuID,
			AverageUtilization:   float64(i%100) + 10.0,
			IdleTimePercentage:   float64(100 - (i%100) - 10),
			EfficiencyScore:      (float64(i%100) + 10.0) / 200.0,
		}
	}
	aggregationService.mu.Unlock()
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		aggregationService.GetEfficiencyReport()
	}
}