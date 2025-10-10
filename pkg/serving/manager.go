package serving

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// Model represents an AI model being served
type Model struct {
	ID         string
	Name       string
	Version    string
	Framework  string
	MemorySize uint64
	LoadedAt   time.Time
}

// InferenceRequest represents a request for model inference
type InferenceRequest struct {
	ID        string
	ModelID   string
	Input     []byte
	BatchID   string
	Priority  int
	CreatedAt time.Time
}

// InferenceResponse represents the result of an inference
type InferenceResponse struct {
	RequestID   string
	Output      []byte
	Latency     time.Duration
	CacheHit    bool
	BatchSize   int
	CompletedAt time.Time
}

// BatchConfig defines batching behavior
type BatchConfig struct {
	MaxBatchSize int
	MaxWaitTime  time.Duration
	MinBatchSize int
}

// CacheEntry stores cached inference results
type CacheEntry struct {
	Key       string
	Response  *InferenceResponse
	ExpiresAt time.Time
	HitCount  int
}

// ServingManager manages AI model serving with optimization
type ServingManager struct {
	models       map[string]*Model
	requestQueue []*InferenceRequest
	cache        map[string]*CacheEntry
	batchConfig  *BatchConfig
	mu           sync.RWMutex
	cacheTTL     time.Duration
}

// NewServingManager creates a new serving manager
func NewServingManager(batchConfig *BatchConfig, cacheTTL time.Duration) *ServingManager {
	if batchConfig == nil {
		batchConfig = &BatchConfig{
			MaxBatchSize: 32,
			MaxWaitTime:  100 * time.Millisecond,
			MinBatchSize: 1,
		}
	}

	return &ServingManager{
		models:       make(map[string]*Model),
		requestQueue: make([]*InferenceRequest, 0),
		cache:        make(map[string]*CacheEntry),
		batchConfig:  batchConfig,
		cacheTTL:     cacheTTL,
	}
}

// RegisterModel adds a model to the serving manager
func (sm *ServingManager) RegisterModel(model *Model) error {
	if model == nil {
		return fmt.Errorf("model cannot be nil")
	}
	if model.ID == "" {
		return fmt.Errorf("model ID cannot be empty")
	}
	if model.Name == "" {
		return fmt.Errorf("model name cannot be empty")
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	model.LoadedAt = time.Now()
	sm.models[model.ID] = model
	return nil
}

// SubmitInferenceRequest submits a new inference request
func (sm *ServingManager) SubmitInferenceRequest(req *InferenceRequest) (*InferenceResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("inference request cannot be nil")
	}
	if req.ID == "" {
		return nil, fmt.Errorf("request ID cannot be empty")
	}
	if req.ModelID == "" {
		return nil, fmt.Errorf("model ID cannot be empty")
	}
	if len(req.Input) == 0 {
		return nil, fmt.Errorf("request input cannot be empty")
	}

	req.CreatedAt = time.Now()

	// Check cache first
	cacheKey := sm.generateCacheKey(req.ModelID, req.Input)
	if cached := sm.checkCache(cacheKey); cached != nil {
		cached.CacheHit = true
		sm.incrementCacheHit(cacheKey)
		return cached, nil
	}

	sm.mu.Lock()
	sm.requestQueue = append(sm.requestQueue, req)
	sm.mu.Unlock()

	// In a real implementation, this would process asynchronously
	// For now, simulate processing
	response := &InferenceResponse{
		RequestID:   req.ID,
		Output:      []byte(fmt.Sprintf("processed_%s", req.ID)),
		Latency:     50 * time.Millisecond,
		CacheHit:    false,
		BatchSize:   1,
		CompletedAt: time.Now(),
	}

	// Store in cache
	sm.storeInCache(cacheKey, response)

	return response, nil
}

// generateCacheKey creates a unique key for caching
func (sm *ServingManager) generateCacheKey(modelID string, input []byte) string {
	hash := sha256.Sum256(append([]byte(modelID), input...))
	return hex.EncodeToString(hash[:])
}

// checkCache looks up a cached response
func (sm *ServingManager) checkCache(key string) *InferenceResponse {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	entry, exists := sm.cache[key]
	if !exists {
		return nil
	}

	if time.Now().After(entry.ExpiresAt) {
		delete(sm.cache, key)
		return nil
	}

	return entry.Response
}

// incrementCacheHit updates cache statistics
func (sm *ServingManager) incrementCacheHit(key string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if entry, exists := sm.cache[key]; exists {
		entry.HitCount++
	}
}

// storeInCache stores a response in the cache
func (sm *ServingManager) storeInCache(key string, response *InferenceResponse) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.cache[key] = &CacheEntry{
		Key:       key,
		Response:  response,
		ExpiresAt: time.Now().Add(sm.cacheTTL),
		HitCount:  0,
	}
}

// ProcessBatch processes queued requests in batches
func (sm *ServingManager) ProcessBatch() ([]*InferenceResponse, error) {
	sm.mu.Lock()

	if len(sm.requestQueue) == 0 {
		sm.mu.Unlock()
		return nil, nil
	}

	batchSize := min(len(sm.requestQueue), sm.batchConfig.MaxBatchSize)
	batch := sm.requestQueue[:batchSize]
	sm.requestQueue = sm.requestQueue[batchSize:]

	sm.mu.Unlock()

	// Process batch
	responses := make([]*InferenceResponse, len(batch))
	for i, req := range batch {
		responses[i] = &InferenceResponse{
			RequestID:   req.ID,
			Output:      []byte(fmt.Sprintf("batch_processed_%s", req.ID)),
			Latency:     30 * time.Millisecond,
			CacheHit:    false,
			BatchSize:   batchSize,
			CompletedAt: time.Now(),
		}
	}

	return responses, nil
}

// GetCacheMetrics returns cache performance statistics
func (sm *ServingManager) GetCacheMetrics() map[string]interface{} {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	totalEntries := len(sm.cache)
	totalHits := 0
	expiredEntries := 0

	now := time.Now()
	for _, entry := range sm.cache {
		totalHits += entry.HitCount
		if now.After(entry.ExpiresAt) {
			expiredEntries++
		}
	}

	return map[string]interface{}{
		"total_entries":   totalEntries,
		"total_hits":      totalHits,
		"expired_entries": expiredEntries,
		"cache_ttl_sec":   sm.cacheTTL.Seconds(),
	}
}

// GetServingMetrics returns overall serving statistics
func (sm *ServingManager) GetServingMetrics() map[string]interface{} {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return map[string]interface{}{
		"total_models":     len(sm.models),
		"pending_requests": len(sm.requestQueue),
		"max_batch_size":   sm.batchConfig.MaxBatchSize,
		"min_batch_size":   sm.batchConfig.MinBatchSize,
		"max_wait_time_ms": sm.batchConfig.MaxWaitTime.Milliseconds(),
	}
}

// CleanExpiredCache removes expired cache entries
func (sm *ServingManager) CleanExpiredCache() int {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	removed := 0
	now := time.Now()

	for key, entry := range sm.cache {
		if now.After(entry.ExpiresAt) {
			delete(sm.cache, key)
			removed++
		}
	}

	return removed
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
