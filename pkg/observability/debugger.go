package observability

import (
	"fmt"
	"sync"
	"time"
)

// DebugLevel represents the level of debug information
type DebugLevel string

const (
	DebugLevelTrace DebugLevel = "trace"
	DebugLevelDebug DebugLevel = "debug"
	DebugLevelInfo  DebugLevel = "info"
	DebugLevelWarn  DebugLevel = "warn"
	DebugLevelError DebugLevel = "error"
)

// DebugEntry represents a debug log entry
type DebugEntry struct {
	Level     DebugLevel
	Message   string
	Source    string
	Data      map[string]interface{}
	Timestamp time.Time
}

// Trace represents a distributed trace
type Trace struct {
	ID        string
	ParentID  string
	Operation string
	StartTime time.Time
	EndTime   *time.Time
	Duration  time.Duration
	Status    string
	Tags      map[string]string
	Logs      []DebugEntry
}

// Debugger provides debugging utilities for AI systems
type Debugger struct {
	logs   []DebugEntry
	traces map[string]*Trace
	mu     sync.RWMutex
	level  DebugLevel
}

// NewDebugger creates a new debugger
func NewDebugger(level DebugLevel) *Debugger {
	return &Debugger{
		logs:   make([]DebugEntry, 0),
		traces: make(map[string]*Trace),
		level:  level,
	}
}

// Log records a debug entry
func (d *Debugger) Log(level DebugLevel, source, message string, data map[string]interface{}) {
	d.mu.Lock()
	defer d.mu.Unlock()

	entry := DebugEntry{
		Level:     level,
		Message:   message,
		Source:    source,
		Data:      data,
		Timestamp: time.Now(),
	}

	d.logs = append(d.logs, entry)
}

// StartTrace begins a new trace
func (d *Debugger) StartTrace(traceID, operation string, tags map[string]string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	trace := &Trace{
		ID:        traceID,
		Operation: operation,
		StartTime: time.Now(),
		Status:    "running",
		Tags:      tags,
		Logs:      make([]DebugEntry, 0),
	}

	d.traces[traceID] = trace
}

// EndTrace completes a trace
func (d *Debugger) EndTrace(traceID string, status string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	trace, exists := d.traces[traceID]
	if !exists {
		return fmt.Errorf("trace %s not found", traceID)
	}

	now := time.Now()
	trace.EndTime = &now
	trace.Duration = now.Sub(trace.StartTime)
	trace.Status = status

	return nil
}

// AddTraceLog adds a log entry to a trace
func (d *Debugger) AddTraceLog(traceID string, level DebugLevel, message string, data map[string]interface{}) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	trace, exists := d.traces[traceID]
	if !exists {
		return fmt.Errorf("trace %s not found", traceID)
	}

	entry := DebugEntry{
		Level:     level,
		Message:   message,
		Data:      data,
		Timestamp: time.Now(),
	}

	trace.Logs = append(trace.Logs, entry)
	return nil
}

// GetTrace retrieves a trace by ID
func (d *Debugger) GetTrace(traceID string) (*Trace, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	trace, exists := d.traces[traceID]
	if !exists {
		return nil, fmt.Errorf("trace %s not found", traceID)
	}

	return trace, nil
}

// GetLogs returns debug logs filtered by level and time range
func (d *Debugger) GetLogs(level DebugLevel, start, end time.Time) []DebugEntry {
	d.mu.RLock()
	defer d.mu.RUnlock()

	result := make([]DebugEntry, 0)
	for _, log := range d.logs {
		if log.Timestamp.After(start) && log.Timestamp.Before(end) {
			if level == "" || log.Level == level {
				result = append(result, log)
			}
		}
	}

	return result
}

// GetTraces returns all traces
func (d *Debugger) GetTraces() []*Trace {
	d.mu.RLock()
	defer d.mu.RUnlock()

	traces := make([]*Trace, 0, len(d.traces))
	for _, trace := range d.traces {
		traces = append(traces, trace)
	}

	return traces
}

// GetDebugStats returns debugging statistics
func (d *Debugger) GetDebugStats() map[string]interface{} {
	d.mu.RLock()
	defer d.mu.RUnlock()

	totalLogs := len(d.logs)
	totalTraces := len(d.traces)
	activeTraces := 0
	errorLogs := 0

	for _, trace := range d.traces {
		if trace.EndTime == nil {
			activeTraces++
		}
	}

	for _, log := range d.logs {
		if log.Level == DebugLevelError {
			errorLogs++
		}
	}

	return map[string]interface{}{
		"total_logs":    totalLogs,
		"total_traces":  totalTraces,
		"active_traces": activeTraces,
		"error_logs":    errorLogs,
		"debug_level":   string(d.level),
	}
}

// AnalyzePerformance analyzes performance bottlenecks from traces
func (d *Debugger) AnalyzePerformance() map[string]interface{} {
	d.mu.RLock()
	defer d.mu.RUnlock()

	operationStats := make(map[string]struct {
		count     int
		totalTime time.Duration
		avgTime   time.Duration
		maxTime   time.Duration
	})

	for _, trace := range d.traces {
		if trace.EndTime == nil {
			continue
		}

		stats := operationStats[trace.Operation]
		stats.count++
		stats.totalTime += trace.Duration

		if trace.Duration > stats.maxTime {
			stats.maxTime = trace.Duration
		}

		stats.avgTime = stats.totalTime / time.Duration(stats.count)
		operationStats[trace.Operation] = stats
	}

	result := make(map[string]interface{})
	for op, stats := range operationStats {
		result[op] = map[string]interface{}{
			"count":        stats.count,
			"avg_duration": stats.avgTime.Milliseconds(),
			"max_duration": stats.maxTime.Milliseconds(),
		}
	}

	return result
}
