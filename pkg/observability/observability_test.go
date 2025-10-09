package observability

import (
	"testing"
	"time"
)

func TestMonitoringService(t *testing.T) {
	monitor := NewMonitoringService(1000)

	// Record metrics
	metric := Metric{
		Name:  "test_metric",
		Type:  MetricCounter,
		Value: 42.0,
		Labels: map[string]string{
			"env": "test",
		},
	}

	monitor.RecordMetric(metric)

	// Verify metrics are stored
	now := time.Now()
	metrics := monitor.GetMetrics(now.Add(-1*time.Minute), now.Add(1*time.Minute), "test_metric")

	if len(metrics) != 1 {
		t.Errorf("Expected 1 metric, got %d", len(metrics))
	}

	if metrics[0].Name != "test_metric" {
		t.Errorf("Expected metric name 'test_metric', got %s", metrics[0].Name)
	}

	if metrics[0].Value != 42.0 {
		t.Errorf("Expected value 42.0, got %f", metrics[0].Value)
	}
}

func TestEventTracking(t *testing.T) {
	monitor := NewMonitoringService(1000)

	// Record events
	event := Event{
		ID:       "evt-1",
		Type:     "test",
		Severity: "info",
		Message:  "Test event",
		Source:   "test_source",
	}

	monitor.RecordEvent(event)

	// Verify events are stored
	now := time.Now()
	events := monitor.GetEvents(now.Add(-1*time.Minute), now.Add(1*time.Minute), "info")

	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}

	if events[0].Message != "Test event" {
		t.Errorf("Expected message 'Test event', got %s", events[0].Message)
	}
}

func TestCostTracking(t *testing.T) {
	monitor := NewMonitoringService(1000)

	// Record costs
	cost1 := CostEntry{
		ID:         "cost-1",
		Operation:  "inference",
		ModelID:    "model-1",
		TokensUsed: 1000,
		GPUHours:   0.5,
		Cost:       2.50,
		Currency:   "USD",
	}

	cost2 := CostEntry{
		ID:        "cost-2",
		Operation: "training",
		ModelID:   "model-2",
		GPUHours:  4.0,
		Cost:      12.00,
		Currency:  "USD",
	}

	monitor.RecordCost(cost1)
	monitor.RecordCost(cost2)

	// Get cost summary
	now := time.Now()
	summary := monitor.GetCostSummary(now.Add(-1*time.Hour), now.Add(1*time.Hour))

	totalCost := summary["total_cost"].(float64)
	if totalCost != 14.50 {
		t.Errorf("Expected total cost 14.50, got %f", totalCost)
	}

	inferenceCost := summary["inference_cost"].(float64)
	if inferenceCost != 2.50 {
		t.Errorf("Expected inference cost 2.50, got %f", inferenceCost)
	}

	trainingCost := summary["training_cost"].(float64)
	if trainingCost != 12.00 {
		t.Errorf("Expected training cost 12.00, got %f", trainingCost)
	}

	totalTokens := summary["total_tokens"].(int64)
	if totalTokens != 1000 {
		t.Errorf("Expected 1000 tokens, got %d", totalTokens)
	}
}

func TestLatencyStats(t *testing.T) {
	monitor := NewMonitoringService(1000)

	// Record latency metrics
	for i := 0; i < 5; i++ {
		metric := Metric{
			Name:  "latency",
			Type:  MetricHistogram,
			Value: float64(10 + i*10),
		}
		monitor.RecordMetric(metric)
	}

	// Get latency stats
	stats := monitor.GetLatencyStats("latency", 1*time.Hour)

	count := stats["count"].(int)
	if count != 5 {
		t.Errorf("Expected count 5, got %d", count)
	}

	avg := stats["average"].(float64)
	if avg != 30.0 {
		t.Errorf("Expected average 30.0, got %f", avg)
	}

	min := stats["min"].(float64)
	if min != 10.0 {
		t.Errorf("Expected min 10.0, got %f", min)
	}

	max := stats["max"].(float64)
	if max != 50.0 {
		t.Errorf("Expected max 50.0, got %f", max)
	}
}

func TestDebugger(t *testing.T) {
	debugger := NewDebugger(DebugLevelInfo)

	// Log messages
	debugger.Log(DebugLevelInfo, "test", "Test message", nil)

	// Verify logs
	now := time.Now()
	logs := debugger.GetLogs(DebugLevelInfo, now.Add(-1*time.Minute), now.Add(1*time.Minute))

	if len(logs) != 1 {
		t.Errorf("Expected 1 log entry, got %d", len(logs))
	}

	if logs[0].Message != "Test message" {
		t.Errorf("Expected message 'Test message', got %s", logs[0].Message)
	}
}

func TestDistributedTracing(t *testing.T) {
	debugger := NewDebugger(DebugLevelInfo)

	// Start trace
	traceID := "trace-1"
	debugger.StartTrace(traceID, "test_operation", map[string]string{
		"tag1": "value1",
	})

	// Add log to trace
	err := debugger.AddTraceLog(traceID, DebugLevelInfo, "Test log", nil)
	if err != nil {
		t.Fatalf("Failed to add trace log: %v", err)
	}

	// End trace
	time.Sleep(10 * time.Millisecond)
	err = debugger.EndTrace(traceID, "success")
	if err != nil {
		t.Fatalf("Failed to end trace: %v", err)
	}

	// Verify trace
	trace, err := debugger.GetTrace(traceID)
	if err != nil {
		t.Fatalf("Failed to get trace: %v", err)
	}

	if trace.Operation != "test_operation" {
		t.Errorf("Expected operation 'test_operation', got %s", trace.Operation)
	}

	if trace.Status != "success" {
		t.Errorf("Expected status 'success', got %s", trace.Status)
	}

	if len(trace.Logs) != 1 {
		t.Errorf("Expected 1 log entry, got %d", len(trace.Logs))
	}

	if trace.Duration <= 0 {
		t.Error("Expected positive duration")
	}
}

func TestPerformanceAnalysis(t *testing.T) {
	debugger := NewDebugger(DebugLevelInfo)

	// Create multiple traces
	for i := 0; i < 3; i++ {
		traceID := "trace-" + string(rune('a'+i))
		debugger.StartTrace(traceID, "operation_1", nil)
		time.Sleep(10 * time.Millisecond)
		debugger.EndTrace(traceID, "success")
	}

	// Analyze performance
	analysis := debugger.AnalyzePerformance()

	stats, exists := analysis["operation_1"]
	if !exists {
		t.Fatal("Expected operation_1 in performance analysis")
	}

	statsMap := stats.(map[string]interface{})
	count := statsMap["count"].(int)
	if count != 3 {
		t.Errorf("Expected count 3, got %d", count)
	}
}

func TestSystemHealth(t *testing.T) {
	monitor := NewMonitoringService(1000)

	// Record some data
	monitor.RecordMetric(Metric{Name: "test", Value: 1.0})
	monitor.RecordEvent(Event{Type: "test", Severity: "info"})
	monitor.RecordCost(CostEntry{Operation: "test", Cost: 1.0})

	// Check health
	health := monitor.GetSystemHealth()

	totalMetrics := health["total_metrics"].(int)
	if totalMetrics != 1 {
		t.Errorf("Expected 1 metric, got %d", totalMetrics)
	}

	totalEvents := health["total_events"].(int)
	if totalEvents != 1 {
		t.Errorf("Expected 1 event, got %d", totalEvents)
	}

	totalCosts := health["total_costs"].(int)
	if totalCosts != 1 {
		t.Errorf("Expected 1 cost entry, got %d", totalCosts)
	}
}
