package main

import (
	"fmt"
	"time"

	"github.com/Finoptimize/agentaflow-sro-community/pkg/serving"
)

func runModelServingExample() {
	fmt.Println("=== Model Serving Example ===")

	// Create serving manager with custom configuration
	fmt.Println("1. Creating serving manager with batching and caching...")
	batchConfig := &serving.BatchConfig{
		MaxBatchSize: 16,
		MaxWaitTime:  50 * time.Millisecond,
		MinBatchSize: 1,
	}
	cacheTTL := 10 * time.Minute
	servingMgr := serving.NewServingManager(batchConfig, cacheTTL)

	// Register models
	fmt.Println("2. Registering AI models...")
	models := []*serving.Model{
		{
			ID:         "gpt-3.5-turbo",
			Name:       "GPT-3.5 Turbo",
			Version:    "1.0.0",
			Framework:  "PyTorch",
			MemorySize: 12288, // 12GB
		},
		{
			ID:         "bert-large",
			Name:       "BERT Large",
			Version:    "2.1.0",
			Framework:  "TensorFlow",
			MemorySize: 4096, // 4GB
		},
		{
			ID:         "llama-2-7b",
			Name:       "LLaMA 2 7B",
			Version:    "2.0.0",
			Framework:  "PyTorch",
			MemorySize: 16384, // 16GB
		},
	}

	for _, m := range models {
		servingMgr.RegisterModel(m)
		fmt.Printf("   Registered: %s (%s, %dMB)\n", m.Name, m.Framework, m.MemorySize)
	}

	// Create router for load balancing
	fmt.Println("\n3. Setting up request routing...")
	router := serving.NewRouter(serving.RouteLeastLatency)

	// Register model instances
	instances := []*serving.ModelInstance{
		{
			ID:             "gpt-instance-1",
			ModelID:        "gpt-3.5-turbo",
			Endpoint:       "http://gpu-node-1:8000",
			MaxLoad:        100,
			CurrentLoad:    0,
			AverageLatency: 45 * time.Millisecond,
			Available:      true,
		},
		{
			ID:             "gpt-instance-2",
			ModelID:        "gpt-3.5-turbo",
			Endpoint:       "http://gpu-node-2:8000",
			MaxLoad:        100,
			CurrentLoad:    0,
			AverageLatency: 52 * time.Millisecond,
			Available:      true,
		},
		{
			ID:             "bert-instance-1",
			ModelID:        "bert-large",
			Endpoint:       "http://gpu-node-3:8000",
			MaxLoad:        200,
			CurrentLoad:    0,
			AverageLatency: 20 * time.Millisecond,
			Available:      true,
		},
	}

	for _, inst := range instances {
		router.RegisterInstance(inst)
		fmt.Printf("   Registered instance: %s -> %s (latency: %vms)\n",
			inst.ID, inst.Endpoint, inst.AverageLatency.Milliseconds())
	}

	// Submit inference requests
	fmt.Println("\n4. Processing inference requests...")
	requests := []*serving.InferenceRequest{
		{
			ID:       "req-1",
			ModelID:  "gpt-3.5-turbo",
			Input:    []byte("What is artificial intelligence?"),
			Priority: 1,
		},
		{
			ID:       "req-2",
			ModelID:  "gpt-3.5-turbo",
			Input:    []byte("What is artificial intelligence?"), // Duplicate - should hit cache
			Priority: 1,
		},
		{
			ID:       "req-3",
			ModelID:  "bert-large",
			Input:    []byte("Classify: This is a positive review"),
			Priority: 2,
		},
		{
			ID:       "req-4",
			ModelID:  "gpt-3.5-turbo",
			Input:    []byte("Explain machine learning"),
			Priority: 1,
		},
		{
			ID:       "req-5",
			ModelID:  "bert-large",
			Input:    []byte("Classify: This is a positive review"), // Duplicate - should hit cache
			Priority: 1,
		},
	}

	var totalLatency time.Duration
	cacheHits := 0
	cacheMisses := 0

	for i, req := range requests {
		// Route request
		instance, err := router.RouteRequest(req.ModelID)
		if err != nil {
			fmt.Printf("   Error routing request %s: %v\n", req.ID, err)
			continue
		}

		// Process request
		response, err := servingMgr.SubmitInferenceRequest(req)
		if err != nil {
			fmt.Printf("   Error processing request %s: %v\n", req.ID, err)
			continue
		}

		totalLatency += response.Latency

		cacheStatus := "MISS"
		if response.CacheHit {
			cacheStatus = "HIT"
			cacheHits++
		} else {
			cacheMisses++
		}

		fmt.Printf("   Request %d (%s):\n", i+1, req.ID)
		fmt.Printf("      Model: %s\n", req.ModelID)
		fmt.Printf("      Routed to: %s\n", instance.Endpoint)
		fmt.Printf("      Latency: %vms\n", response.Latency.Milliseconds())
		fmt.Printf("      Cache: %s\n", cacheStatus)
	}

	// Display performance metrics
	fmt.Println("\n5. Performance Metrics:")
	avgLatency := totalLatency / time.Duration(len(requests))
	fmt.Printf("   Total Requests: %d\n", len(requests))
	fmt.Printf("   Average Latency: %vms\n", avgLatency.Milliseconds())
	fmt.Printf("   Cache Hits: %d\n", cacheHits)
	fmt.Printf("   Cache Misses: %d\n", cacheMisses)
	if len(requests) > 0 {
		cacheHitRate := float64(cacheHits) / float64(len(requests)) * 100
		fmt.Printf("   Cache Hit Rate: %.1f%%\n", cacheHitRate)
	}

	// Display serving metrics
	fmt.Println("\n6. Serving Manager Metrics:")
	servingMetrics := servingMgr.GetServingMetrics()
	for key, value := range servingMetrics {
		fmt.Printf("   %s: %v\n", key, value)
	}

	// Display cache metrics
	fmt.Println("\n7. Cache Metrics:")
	cacheMetrics := servingMgr.GetCacheMetrics()
	for key, value := range cacheMetrics {
		fmt.Printf("   %s: %v\n", key, value)
	}

	// Display routing metrics
	fmt.Println("\n8. Routing Metrics:")
	routingMetrics := router.GetRoutingMetrics()
	for key, value := range routingMetrics {
		fmt.Printf("   %s: %v\n", key, value)
	}

	// Demonstrate batch processing
	fmt.Println("\n9. Demonstrating batch processing...")
	batchRequests := []*serving.InferenceRequest{
		{ID: "batch-1", ModelID: "gpt-3.5-turbo", Input: []byte("Request 1")},
		{ID: "batch-2", ModelID: "gpt-3.5-turbo", Input: []byte("Request 2")},
		{ID: "batch-3", ModelID: "gpt-3.5-turbo", Input: []byte("Request 3")},
		{ID: "batch-4", ModelID: "gpt-3.5-turbo", Input: []byte("Request 4")},
	}

	for _, req := range batchRequests {
		servingMgr.SubmitInferenceRequest(req)
	}

	responses, err := servingMgr.ProcessBatch()
	if err != nil {
		fmt.Printf("Error processing batch: %v\n", err)
	} else if responses != nil {
		fmt.Printf("   Processed batch of %d requests\n", len(responses))
		for _, resp := range responses {
			fmt.Printf("      %s: %vms (batch size: %d)\n",
				resp.RequestID, resp.Latency.Milliseconds(), resp.BatchSize)
		}
	}

	// Test cache cleanup
	fmt.Println("\n10. Cache Management:")
	removed := servingMgr.CleanExpiredCache()
	fmt.Printf("   Cleaned up %d expired cache entries\n", removed)

	// Compare routing strategies
	fmt.Println("\n11. Comparing routing strategies...")
	strategies := []serving.RoutingStrategy{
		serving.RouteLeastLatency,
		serving.RouteLeastLoad,
		serving.RouteRoundRobin,
	}

	for _, strategy := range strategies {
		testRouter := serving.NewRouter(strategy)
		for _, inst := range instances {
			testRouter.RegisterInstance(inst)
		}

		fmt.Printf("\n   Strategy: %s\n", strategy)
		for i := 0; i < 3; i++ {
			instance, err := testRouter.RouteRequest("gpt-3.5-turbo")
			if err != nil {
				fmt.Printf("      Error: %v\n", err)
			} else {
				fmt.Printf("      Request %d routed to: %s\n", i+1, instance.ID)
			}
		}
	}

	fmt.Println("\n=== Example Complete ===")
}
