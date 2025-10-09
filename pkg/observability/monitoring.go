package observability

import (
	"sync"
	"time"
)

// MetricType represents different types of metrics
type MetricType string

const (
	MetricCounter   MetricType = "counter"
	MetricGauge     MetricType = "gauge"
	MetricHistogram MetricType = "histogram"
)

// Metric represents a single metric measurement
type Metric struct {
	Name      string
	Type      MetricType
	Value     float64
	Labels    map[string]string
	Timestamp time.Time
}

// Event represents a significant occurrence in the system
type Event struct {
	ID        string
	Type      string
	Severity  string
	Message   string
	Source    string
	Metadata  map[string]interface{}
	Timestamp time.Time
}

// CostEntry tracks costs for AI operations
type CostEntry struct {
	ID          string
	Operation   string // "inference" or "training"
	ModelID     string
	Duration    time.Duration
	TokensUsed  int64
	GPUHours    float64
	Cost        float64
	Currency    string
	Timestamp   time.Time
}

// MonitoringService provides observability for AI systems
type MonitoringService struct {
	metrics      []Metric
	events       []Event
	costs        []CostEntry
	mu           sync.RWMutex
	maxHistorySize int
}

// NewMonitoringService creates a new monitoring service
func NewMonitoringService(maxHistorySize int) *MonitoringService {
	if maxHistorySize <= 0 {
		maxHistorySize = 10000
	}
	
	return &MonitoringService{
		metrics:        make([]Metric, 0),
		events:         make([]Event, 0),
		costs:          make([]CostEntry, 0),
		maxHistorySize: maxHistorySize,
	}
}

// RecordMetric records a new metric
func (ms *MonitoringService) RecordMetric(metric Metric) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	
	metric.Timestamp = time.Now()
	ms.metrics = append(ms.metrics, metric)
	
	// Trim old metrics if we exceed max size
	if len(ms.metrics) > ms.maxHistorySize {
		ms.metrics = ms.metrics[len(ms.metrics)-ms.maxHistorySize:]
	}
}

// RecordEvent records a new event
func (ms *MonitoringService) RecordEvent(event Event) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	
	event.Timestamp = time.Now()
	ms.events = append(ms.events, event)
	
	// Trim old events if we exceed max size
	if len(ms.events) > ms.maxHistorySize {
		ms.events = ms.events[len(ms.events)-ms.maxHistorySize:]
	}
}

// RecordCost records a cost entry
func (ms *MonitoringService) RecordCost(cost CostEntry) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	
	cost.Timestamp = time.Now()
	ms.costs = append(ms.costs, cost)
	
	// Trim old cost entries if we exceed max size
	if len(ms.costs) > ms.maxHistorySize {
		ms.costs = ms.costs[len(ms.costs)-ms.maxHistorySize:]
	}
}

// GetMetrics returns metrics within a time range
func (ms *MonitoringService) GetMetrics(start, end time.Time, metricName string) []Metric {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	
	result := make([]Metric, 0)
	for _, metric := range ms.metrics {
		if metric.Timestamp.After(start) && metric.Timestamp.Before(end) {
			if metricName == "" || metric.Name == metricName {
				result = append(result, metric)
			}
		}
	}
	
	return result
}

// GetEvents returns events within a time range
func (ms *MonitoringService) GetEvents(start, end time.Time, severity string) []Event {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	
	result := make([]Event, 0)
	for _, event := range ms.events {
		if event.Timestamp.After(start) && event.Timestamp.Before(end) {
			if severity == "" || event.Severity == severity {
				result = append(result, event)
			}
		}
	}
	
	return result
}

// GetCostSummary calculates cost summary for a time period
func (ms *MonitoringService) GetCostSummary(start, end time.Time) map[string]interface{} {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	
	totalCost := 0.0
	inferenceCost := 0.0
	trainingCost := 0.0
	totalTokens := int64(0)
	totalGPUHours := 0.0
	operationCounts := make(map[string]int)
	
	for _, cost := range ms.costs {
		if cost.Timestamp.After(start) && cost.Timestamp.Before(end) {
			totalCost += cost.Cost
			totalTokens += cost.TokensUsed
			totalGPUHours += cost.GPUHours
			operationCounts[cost.Operation]++
			
			if cost.Operation == "inference" {
				inferenceCost += cost.Cost
			} else if cost.Operation == "training" {
				trainingCost += cost.Cost
			}
		}
	}
	
	return map[string]interface{}{
		"total_cost":         totalCost,
		"inference_cost":     inferenceCost,
		"training_cost":      trainingCost,
		"total_tokens":       totalTokens,
		"total_gpu_hours":    totalGPUHours,
		"operation_counts":   operationCounts,
		"period_start":       start,
		"period_end":         end,
	}
}

// GetSystemHealth returns current system health metrics
func (ms *MonitoringService) GetSystemHealth() map[string]interface{} {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	
	recentEvents := 0
	criticalEvents := 0
	now := time.Now()
	fiveMinutesAgo := now.Add(-5 * time.Minute)
	
	for _, event := range ms.events {
		if event.Timestamp.After(fiveMinutesAgo) {
			recentEvents++
			if event.Severity == "critical" || event.Severity == "error" {
				criticalEvents++
			}
		}
	}
	
	return map[string]interface{}{
		"total_metrics":     len(ms.metrics),
		"total_events":      len(ms.events),
		"total_costs":       len(ms.costs),
		"recent_events":     recentEvents,
		"critical_events":   criticalEvents,
		"max_history_size":  ms.maxHistorySize,
	}
}

// GetLatencyStats calculates latency statistics from metrics
func (ms *MonitoringService) GetLatencyStats(metricName string, duration time.Duration) map[string]interface{} {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	
	now := time.Now()
	start := now.Add(-duration)
	
	latencies := make([]float64, 0)
	for _, metric := range ms.metrics {
		if metric.Name == metricName && metric.Timestamp.After(start) {
			latencies = append(latencies, metric.Value)
		}
	}
	
	if len(latencies) == 0 {
		return map[string]interface{}{
			"count": 0,
		}
	}
	
	// Calculate statistics
	sum := 0.0
	min := latencies[0]
	max := latencies[0]
	
	for _, val := range latencies {
		sum += val
		if val < min {
			min = val
		}
		if val > max {
			max = val
		}
	}
	
	avg := sum / float64(len(latencies))
	
	return map[string]interface{}{
		"count":   len(latencies),
		"average": avg,
		"min":     min,
		"max":     max,
		"sum":     sum,
	}
}
