package serving

import (
	"testing"
	"time"
)

func TestServingManagerCaching(t *testing.T) {
	batchConfig := &BatchConfig{
		MaxBatchSize: 32,
		MaxWaitTime:  100 * time.Millisecond,
	}
	manager := NewServingManager(batchConfig, 5*time.Minute)

	// Register model
	model := &Model{
		ID:      "test-model",
		Name:    "Test Model",
		Version: "1.0",
	}
	manager.RegisterModel(model)

	// First request
	req1 := &InferenceRequest{
		ID:      "req-1",
		ModelID: "test-model",
		Input:   []byte("test input"),
	}

	resp1, err := manager.SubmitInferenceRequest(req1)
	if err != nil {
		t.Fatalf("Failed to submit request: %v", err)
	}

	if resp1.CacheHit {
		t.Error("First request should not be a cache hit")
	}

	// Second request with same input - should hit cache
	req2 := &InferenceRequest{
		ID:      "req-2",
		ModelID: "test-model",
		Input:   []byte("test input"),
	}

	resp2, err := manager.SubmitInferenceRequest(req2)
	if err != nil {
		t.Fatalf("Failed to submit request: %v", err)
	}

	if !resp2.CacheHit {
		t.Error("Second request should be a cache hit")
	}

	// Verify cache metrics
	metrics := manager.GetCacheMetrics()
	totalHits := metrics["total_hits"].(int)
	if totalHits < 1 {
		t.Errorf("Expected at least 1 cache hit, got %d", totalHits)
	}
}

func TestBatchProcessing(t *testing.T) {
	batchConfig := &BatchConfig{
		MaxBatchSize: 4,
		MaxWaitTime:  50 * time.Millisecond,
	}
	manager := NewServingManager(batchConfig, 5*time.Minute)

	model := &Model{
		ID:   "test-model",
		Name: "Test Model",
	}
	manager.RegisterModel(model)

	// Note: In the current implementation, SubmitInferenceRequest processes
	// requests immediately. To test batch processing, we need to use the
	// ProcessBatch method directly after manually adding to the queue.
	// For this test, we'll verify that ProcessBatch works correctly.

	// Add requests to queue manually for batch processing
	manager.mu.Lock()
	for i := 0; i < 5; i++ {
		req := &InferenceRequest{
			ID:      string(rune('a' + i)),
			ModelID: "test-model",
			Input:   []byte("test"),
		}
		manager.requestQueue = append(manager.requestQueue, req)
	}
	manager.mu.Unlock()

	// Process batch
	responses, err := manager.ProcessBatch()
	if err != nil {
		t.Fatalf("Failed to process batch: %v", err)
	}

	// Should process up to MaxBatchSize
	if len(responses) != 4 {
		t.Errorf("Expected 4 responses in batch, got %d", len(responses))
	}

	// Verify batch size is recorded
	if responses[0].BatchSize != 4 {
		t.Errorf("Expected batch size 4, got %d", responses[0].BatchSize)
	}

	// Verify one request remains in queue
	manager.mu.RLock()
	remaining := len(manager.requestQueue)
	manager.mu.RUnlock()

	if remaining != 1 {
		t.Errorf("Expected 1 request remaining in queue, got %d", remaining)
	}
}

func TestRouterLeastLatency(t *testing.T) {
	router := NewRouter(RouteLeastLatency)

	// Register instances with different latencies
	inst1 := &ModelInstance{
		ID:             "inst-1",
		ModelID:        "model-1",
		Available:      true,
		MaxLoad:        100,
		CurrentLoad:    0,
		AverageLatency: 100 * time.Millisecond,
	}

	inst2 := &ModelInstance{
		ID:             "inst-2",
		ModelID:        "model-1",
		Available:      true,
		MaxLoad:        100,
		CurrentLoad:    0,
		AverageLatency: 50 * time.Millisecond,
	}

	router.RegisterInstance(inst1)
	router.RegisterInstance(inst2)

	// Route request - should go to inst-2 (lower latency)
	instance, err := router.RouteRequest("model-1")
	if err != nil {
		t.Fatalf("Failed to route request: %v", err)
	}

	if instance.ID != "inst-2" {
		t.Errorf("Expected inst-2 (lowest latency), got %s", instance.ID)
	}
}

func TestRouterLeastLoad(t *testing.T) {
	router := NewRouter(RouteLeastLoad)

	inst1 := &ModelInstance{
		ID:          "inst-1",
		ModelID:     "model-1",
		Available:   true,
		MaxLoad:     100,
		CurrentLoad: 50,
	}

	inst2 := &ModelInstance{
		ID:          "inst-2",
		ModelID:     "model-1",
		Available:   true,
		MaxLoad:     100,
		CurrentLoad: 10,
	}

	router.RegisterInstance(inst1)
	router.RegisterInstance(inst2)

	// Route request - should go to inst-2 (lower load)
	instance, err := router.RouteRequest("model-1")
	if err != nil {
		t.Fatalf("Failed to route request: %v", err)
	}

	if instance.ID != "inst-2" {
		t.Errorf("Expected inst-2 (lowest load), got %s", instance.ID)
	}
}

func TestCacheExpiration(t *testing.T) {
	batchConfig := &BatchConfig{
		MaxBatchSize: 32,
		MaxWaitTime:  100 * time.Millisecond,
	}
	// Short TTL for testing
	manager := NewServingManager(batchConfig, 100*time.Millisecond)

	model := &Model{
		ID: "test-model",
	}
	manager.RegisterModel(model)

	// Submit request
	req := &InferenceRequest{
		ID:      "req-1",
		ModelID: "test-model",
		Input:   []byte("test"),
	}
	manager.SubmitInferenceRequest(req)

	// Wait for cache to expire
	time.Sleep(200 * time.Millisecond)

	// Clean expired entries
	removed := manager.CleanExpiredCache()
	if removed < 1 {
		t.Error("Expected at least 1 expired cache entry to be cleaned")
	}
}

func TestServingMetrics(t *testing.T) {
	batchConfig := &BatchConfig{
		MaxBatchSize: 16,
		MaxWaitTime:  50 * time.Millisecond,
		MinBatchSize: 2,
	}
	manager := NewServingManager(batchConfig, 5*time.Minute)

	metrics := manager.GetServingMetrics()

	if metrics["max_batch_size"].(int) != 16 {
		t.Errorf("Expected max_batch_size 16, got %v", metrics["max_batch_size"])
	}

	if metrics["min_batch_size"].(int) != 2 {
		t.Errorf("Expected min_batch_size 2, got %v", metrics["min_batch_size"])
	}

	if metrics["max_wait_time_ms"].(int64) != 50 {
		t.Errorf("Expected max_wait_time_ms 50, got %v", metrics["max_wait_time_ms"])
	}
}
