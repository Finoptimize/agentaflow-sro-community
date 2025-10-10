package k8s

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Finoptimize/agentaflow-sro-community/pkg/gpu"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// KubernetesGPUScheduler manages GPU scheduling across Kubernetes clusters
type KubernetesGPUScheduler struct {
	clientset         kubernetes.Interface
	gpuScheduler      *gpu.Scheduler
	namespace         string
	nodeMap           map[string]*GPUNode
	workloadMap       map[string]*GPUWorkload
	mu                sync.RWMutex
	stopCh            chan struct{}
	metricsUpdateTime time.Time
}

// NewKubernetesGPUScheduler creates a new Kubernetes GPU scheduler
func NewKubernetesGPUScheduler(namespace string, strategy gpu.SchedulingStrategy) (*KubernetesGPUScheduler, error) {
	// Try in-cluster config first, then fallback to kubeconfig
	var config *rest.Config
	var err error

	config, err = rest.InClusterConfig()
	if err != nil {
		// Fallback to kubeconfig
		config, err = clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
		if err != nil {
			return nil, fmt.Errorf("failed to create Kubernetes config: %v", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %v", err)
	}

	return &KubernetesGPUScheduler{
		clientset:    clientset,
		gpuScheduler: gpu.NewScheduler(strategy),
		namespace:    namespace,
		nodeMap:      make(map[string]*GPUNode),
		workloadMap:  make(map[string]*GPUWorkload),
		stopCh:       make(chan struct{}),
	}, nil
}

// Start begins the GPU scheduler and monitoring loops
func (ks *KubernetesGPUScheduler) Start(ctx context.Context) error {
	// Start node discovery and monitoring
	go ks.nodeDiscoveryLoop(ctx)

	// Start workload scheduling loop
	go ks.schedulingLoop(ctx)

	// Start metrics collection
	go ks.metricsLoop(ctx)

	return nil
}

// Stop gracefully shuts down the scheduler
func (ks *KubernetesGPUScheduler) Stop() {
	close(ks.stopCh)
}

// nodeDiscoveryLoop continuously discovers and monitors GPU-enabled nodes
func (ks *KubernetesGPUScheduler) nodeDiscoveryLoop(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Initial discovery
	ks.discoverNodes(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ks.stopCh:
			return
		case <-ticker.C:
			ks.discoverNodes(ctx)
		}
	}
}

// discoverNodes finds GPU-enabled nodes and registers them
func (ks *KubernetesGPUScheduler) discoverNodes(ctx context.Context) {
	nodes, err := ks.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{
		LabelSelector: "agentaflow.gpu/enabled=true",
	})
	if err != nil {
		fmt.Printf("Failed to list nodes: %v\n", err)
		return
	}

	ks.mu.Lock()
	defer ks.mu.Unlock()

	for _, node := range nodes.Items {
		ks.processNode(&node)
	}
}

// processNode processes a Kubernetes node and extracts GPU information
func (ks *KubernetesGPUScheduler) processNode(node *v1.Node) {
	// Look for GPU annotations
	gpuCountStr, hasGPUs := node.Annotations["agentaflow.gpu/count"]
	if !hasGPUs {
		return
	}

	gpuDevicesStr, hasDevices := node.Annotations["agentaflow.gpu/devices"]
	if !hasDevices {
		return
	}

	// Parse GPU information from annotations
	gpuDevices := ks.parseGPUDevices(gpuDevicesStr)

	gpuNode := &GPUNode{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "agentaflow.io/v1",
			Kind:       "GPUNode",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      node.Name,
			Namespace: ks.namespace,
		},
		Spec: GPUNodeSpec{
			NodeName:   node.Name,
			GPUDevices: gpuDevices,
		},
		Status: GPUNodeStatus{
			Phase:       GPUNodeActive,
			LastUpdated: metav1.Now(),
		},
	}

	ks.nodeMap[node.Name] = gpuNode

	// Register GPUs with the scheduler
	for _, device := range gpuDevices {
		gpuResource := &gpu.GPU{
			ID:          fmt.Sprintf("%s/%s", node.Name, device.ID),
			Name:        device.Name,
			MemoryTotal: uint64(device.MemoryTotal),
			Available:   true,
		}
		ks.gpuScheduler.RegisterGPU(gpuResource)
	}
}

// parseGPUDevices parses GPU device information from node annotations
func (ks *KubernetesGPUScheduler) parseGPUDevices(devicesStr string) []GPUDevice {
	// This is a simplified parser - in reality, you'd parse JSON or YAML
	// For demo purposes, we'll create a mock GPU device
	return []GPUDevice{
		{
			ID:            "gpu-0",
			Name:          "NVIDIA A100",
			MemoryTotal:   40960, // 40GB
			PCIBusID:      "0000:00:1e.0",
			DriverVersion: "470.129.06",
		},
	}
}

// SubmitGPUWorkload submits a new GPU workload for scheduling
func (ks *KubernetesGPUScheduler) SubmitGPUWorkload(workload *GPUWorkload) error {
	ks.mu.Lock()
	defer ks.mu.Unlock()

	// Convert to internal workload format
	internalWorkload := &gpu.Workload{
		ID:             workload.Name,
		Name:           workload.Name,
		Priority:       int(workload.Spec.Priority),
		MemoryRequired: uint64(workload.Spec.GPUMemoryRequired),
	}

	if workload.Spec.EstimatedDuration != nil {
		internalWorkload.EstimatedTime = workload.Spec.EstimatedDuration.Duration
	}

	// Submit to internal scheduler
	err := ks.gpuScheduler.SubmitWorkload(internalWorkload)
	if err != nil {
		return err
	}

	// Update workload status
	workload.Status.Phase = GPUWorkloadPending
	workload.Status.Conditions = []GPUWorkloadCondition{
		{
			Type:               GPUWorkloadSchedulable,
			Status:             v1.ConditionTrue,
			LastTransitionTime: metav1.Now(),
			Reason:             "WorkloadSubmitted",
			Message:            "Workload has been submitted for scheduling",
		},
	}

	ks.workloadMap[workload.Name] = workload
	return nil
}

// schedulingLoop continuously runs the scheduling algorithm
func (ks *KubernetesGPUScheduler) schedulingLoop(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ks.stopCh:
			return
		case <-ticker.C:
			ks.runSchedulingCycle()
		}
	}
}

// runSchedulingCycle executes one scheduling cycle
func (ks *KubernetesGPUScheduler) runSchedulingCycle() {
	ks.mu.Lock()
	defer ks.mu.Unlock()

	// Run the internal scheduler
	err := ks.gpuScheduler.Schedule()
	if err != nil {
		fmt.Printf("Scheduling error: %v\n", err)
		return
	}

	// Update workload statuses based on scheduling results
	ks.updateWorkloadStatuses()
}

// updateWorkloadStatuses updates the status of workloads based on scheduling results
func (ks *KubernetesGPUScheduler) updateWorkloadStatuses() {
	gpuStatuses := ks.gpuScheduler.GetGPUStatus()

	for _, gpuStatus := range gpuStatuses {
		if gpuStatus.CurrentWorkload != nil {
			workload, exists := ks.workloadMap[gpuStatus.CurrentWorkload.ID]
			if !exists {
				continue
			}

			// Update workload status based on GPU assignment
			if workload.Status.Phase == GPUWorkloadPending {
				workload.Status.Phase = GPUWorkloadScheduled
				workload.Status.AssignedGPU = gpuStatus.ID
				workload.Status.AssignedNode = ks.extractNodeName(gpuStatus.ID)
				workload.Status.StartTime = &metav1.Time{Time: time.Now()}

				workload.Status.Conditions = append(workload.Status.Conditions, GPUWorkloadCondition{
					Type:               GPUWorkloadAssigned,
					Status:             v1.ConditionTrue,
					LastTransitionTime: metav1.Now(),
					Reason:             "GPUAssigned",
					Message:            fmt.Sprintf("Workload assigned to GPU %s", gpuStatus.ID),
				})

				// In a real implementation, you would create the actual pod here
				ks.createWorkloadPod(workload)
			}
		}
	}
}

// extractNodeName extracts the node name from a GPU ID
func (ks *KubernetesGPUScheduler) extractNodeName(gpuID string) string {
	// GPU ID format: "node-name/gpu-id"
	for i, char := range gpuID {
		if char == '/' {
			return gpuID[:i]
		}
	}
	return gpuID
}

// createWorkloadPod creates a Kubernetes pod for the scheduled workload
func (ks *KubernetesGPUScheduler) createWorkloadPod(workload *GPUWorkload) error {
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      workload.Name,
			Namespace: ks.namespace,
			Labels: map[string]string{
				"agentaflow.gpu/workload": workload.Name,
				"agentaflow.gpu/managed":  "true",
			},
			Annotations: map[string]string{
				"agentaflow.gpu/assigned-gpu":  workload.Status.AssignedGPU,
				"agentaflow.gpu/assigned-node": workload.Status.AssignedNode,
			},
		},
		Spec: workload.Spec.PodTemplate.Spec,
	}

	// Add GPU resource requirements and node selector
	if pod.Spec.NodeSelector == nil {
		pod.Spec.NodeSelector = make(map[string]string)
	}
	pod.Spec.NodeSelector["kubernetes.io/hostname"] = workload.Status.AssignedNode

	// Add GPU resource limits
	for i := range pod.Spec.Containers {
		if pod.Spec.Containers[i].Resources.Limits == nil {
			pod.Spec.Containers[i].Resources.Limits = make(v1.ResourceList)
		}
		pod.Spec.Containers[i].Resources.Limits["nvidia.com/gpu"] = *metav1.NewQuantity(int64(workload.Spec.GPURequirements.GPUCount), metav1.DecimalSI)
	}

	_, err := ks.clientset.CoreV1().Pods(ks.namespace).Create(context.TODO(), pod, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create pod: %v", err)
	}

	workload.Status.Phase = GPUWorkloadRunning
	workload.Status.Conditions = append(workload.Status.Conditions, GPUWorkloadCondition{
		Type:               GPUWorkloadReady,
		Status:             v1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Reason:             "PodCreated",
		Message:            "Pod has been created and is starting",
	})

	return nil
}

// metricsLoop continuously updates scheduling metrics
func (ks *KubernetesGPUScheduler) metricsLoop(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ks.stopCh:
			return
		case <-ticker.C:
			ks.updateMetrics()
		}
	}
}

// updateMetrics updates the scheduling metrics
func (ks *KubernetesGPUScheduler) updateMetrics() {
	ks.mu.Lock()
	defer ks.mu.Unlock()

	ks.metricsUpdateTime = time.Now()
}

// GetSchedulingMetrics returns current scheduling metrics
func (ks *KubernetesGPUScheduler) GetSchedulingMetrics() *SchedulingMetrics {
	ks.mu.RLock()
	defer ks.mu.RUnlock()

	utilizationMetrics := ks.gpuScheduler.GetUtilizationMetrics()

	totalNodes := len(ks.nodeMap)
	activeNodes := 0
	for _, node := range ks.nodeMap {
		if node.Status.Phase == GPUNodeActive {
			activeNodes++
		}
	}

	pendingWorkloads := 0
	runningWorkloads := 0
	completedWorkloads := 0

	for _, workload := range ks.workloadMap {
		switch workload.Status.Phase {
		case GPUWorkloadPending:
			pendingWorkloads++
		case GPUWorkloadRunning, GPUWorkloadScheduled:
			runningWorkloads++
		case GPUWorkloadSucceeded:
			completedWorkloads++
		}
	}

	// Safely extract metrics with type checks and defaults
	totalGPUs, ok := utilizationMetrics["total_gpus"].(int)
	if !ok {
		totalGPUs = 0
	}
	activeGPUs, ok := utilizationMetrics["active_gpus"].(int)
	if !ok {
		activeGPUs = 0
	}
	averageUtilization, ok := utilizationMetrics["average_utilization"].(float64)
	if !ok {
		averageUtilization = 0.0
	}
	memoryUtilization, ok := utilizationMetrics["memory_utilization"].(float64)
	if !ok {
		memoryUtilization = 0.0
	}

	return &SchedulingMetrics{
		TotalNodes:         totalNodes,
		ActiveNodes:        activeNodes,
		TotalGPUs:          totalGPUs,
		AvailableGPUs:      totalGPUs - activeGPUs,
		UtilizedGPUs:       activeGPUs,
		AverageUtilization: averageUtilization,
		PendingWorkloads:   pendingWorkloads,
		RunningWorkloads:   runningWorkloads,
		CompletedWorkloads: completedWorkloads,
		MemoryUtilization:  memoryUtilization,
		LastUpdateTime:     ks.metricsUpdateTime,
	}
}

// ListGPUNodes returns all GPU-enabled nodes
func (ks *KubernetesGPUScheduler) ListGPUNodes() []*GPUNode {
	ks.mu.RLock()
	defer ks.mu.RUnlock()

	nodes := make([]*GPUNode, 0, len(ks.nodeMap))
	for _, node := range ks.nodeMap {
		nodes = append(nodes, node)
	}
	return nodes
}

// ListGPUWorkloads returns all GPU workloads
func (ks *KubernetesGPUScheduler) ListGPUWorkloads() []*GPUWorkload {
	ks.mu.RLock()
	defer ks.mu.RUnlock()

	workloads := make([]*GPUWorkload, 0, len(ks.workloadMap))
	for _, workload := range ks.workloadMap {
		workloads = append(workloads, workload)
	}
	return workloads
}

// CompleteWorkload marks a workload as completed
func (ks *KubernetesGPUScheduler) CompleteWorkload(workloadName string) error {
	ks.mu.Lock()
	defer ks.mu.Unlock()

	workload, exists := ks.workloadMap[workloadName]
	if !exists {
		return fmt.Errorf("workload %s not found", workloadName)
	}

	// Complete in internal scheduler
	err := ks.gpuScheduler.CompleteWorkload(workloadName)
	if err != nil {
		return err
	}

	// Update workload status
	workload.Status.Phase = GPUWorkloadSucceeded
	workload.Status.CompletionTime = &metav1.Time{Time: time.Now()}
	workload.Status.Message = "Workload completed successfully"

	return nil
}

// GetWorkloadStatus returns the status of a specific workload
func (ks *KubernetesGPUScheduler) GetWorkloadStatus(workloadName string) (*GPUWorkloadStatus, error) {
	ks.mu.RLock()
	defer ks.mu.RUnlock()

	workload, exists := ks.workloadMap[workloadName]
	if !exists {
		return nil, fmt.Errorf("workload %s not found", workloadName)
	}

	return &workload.Status, nil
}

// SetSchedulingStrategy changes the scheduling strategy
func (ks *KubernetesGPUScheduler) SetSchedulingStrategy(strategy gpu.SchedulingStrategy) {
	ks.mu.Lock()
	defer ks.mu.Unlock()

	// Create new scheduler with the new strategy
	newScheduler := gpu.NewScheduler(strategy)

	// Transfer GPU registrations
	for _, gpuStatus := range ks.gpuScheduler.GetGPUStatus() {
		newScheduler.RegisterGPU(gpuStatus)
	}

	ks.gpuScheduler = newScheduler
}
