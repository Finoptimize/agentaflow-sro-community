package serving

import (
	"fmt"
	"sync"
	"time"
)

// RoutingStrategy defines how requests are routed to model instances
type RoutingStrategy string

const (
	RouteRoundRobin    RoutingStrategy = "round_robin"
	RouteLeastLatency  RoutingStrategy = "least_latency"
	RouteLeastLoad     RoutingStrategy = "least_load"
)

// ModelInstance represents a running instance of a model
type ModelInstance struct {
	ID              string
	ModelID         string
	Endpoint        string
	CurrentLoad     int
	MaxLoad         int
	AverageLatency  time.Duration
	Available       bool
}

// Router manages request routing across model instances
type Router struct {
	instances map[string][]*ModelInstance
	strategy  RoutingStrategy
	mu        sync.RWMutex
}

// NewRouter creates a new request router
func NewRouter(strategy RoutingStrategy) *Router {
	return &Router{
		instances: make(map[string][]*ModelInstance),
		strategy:  strategy,
	}
}

// RegisterInstance adds a model instance to the router
func (r *Router) RegisterInstance(instance *ModelInstance) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, exists := r.instances[instance.ModelID]; !exists {
		r.instances[instance.ModelID] = make([]*ModelInstance, 0)
	}
	
	r.instances[instance.ModelID] = append(r.instances[instance.ModelID], instance)
}

// RouteRequest selects the best instance for a request
func (r *Router) RouteRequest(modelID string) (*ModelInstance, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	instances, exists := r.instances[modelID]
	if !exists || len(instances) == 0 {
		return nil, fmt.Errorf("no instances available for model %s", modelID)
	}
	
	switch r.strategy {
	case RouteLeastLatency:
		return r.routeByLatency(instances)
	case RouteLeastLoad:
		return r.routeByLoad(instances)
	case RouteRoundRobin:
		fallthrough
	default:
		return r.routeRoundRobin(instances)
	}
}

// routeByLatency selects the instance with lowest average latency
func (r *Router) routeByLatency(instances []*ModelInstance) (*ModelInstance, error) {
	var best *ModelInstance
	minLatency := time.Duration(1<<63 - 1)
	
	for _, instance := range instances {
		if instance.Available && instance.CurrentLoad < instance.MaxLoad {
			if instance.AverageLatency < minLatency {
				minLatency = instance.AverageLatency
				best = instance
			}
		}
	}
	
	if best == nil {
		return nil, fmt.Errorf("no available instances")
	}
	
	return best, nil
}

// routeByLoad selects the instance with lowest current load
func (r *Router) routeByLoad(instances []*ModelInstance) (*ModelInstance, error) {
	var best *ModelInstance
	minLoad := int(^uint(0) >> 1)
	
	for _, instance := range instances {
		if instance.Available && instance.CurrentLoad < instance.MaxLoad {
			if instance.CurrentLoad < minLoad {
				minLoad = instance.CurrentLoad
				best = instance
			}
		}
	}
	
	if best == nil {
		return nil, fmt.Errorf("no available instances")
	}
	
	return best, nil
}

// routeRoundRobin distributes requests evenly
func (r *Router) routeRoundRobin(instances []*ModelInstance) (*ModelInstance, error) {
	for _, instance := range instances {
		if instance.Available && instance.CurrentLoad < instance.MaxLoad {
			return instance, nil
		}
	}
	
	return nil, fmt.Errorf("no available instances")
}

// GetRoutingMetrics returns routing statistics
func (r *Router) GetRoutingMetrics() map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	totalInstances := 0
	availableInstances := 0
	
	for _, instances := range r.instances {
		totalInstances += len(instances)
		for _, instance := range instances {
			if instance.Available {
				availableInstances++
			}
		}
	}
	
	return map[string]interface{}{
		"total_instances":     totalInstances,
		"available_instances": availableInstances,
		"routing_strategy":    string(r.strategy),
		"models_registered":   len(r.instances),
	}
}
