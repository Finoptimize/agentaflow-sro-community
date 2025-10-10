package k8s

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// GPU monitoring constants
const (
	// Memory thresholds for health monitoring
	MemoryUsageAlertThresholdMB = 8000  // Alert if using more than 8GB
	DefaultGPUMemoryTotalMB     = 40960 // Default assumption: 40GB GPU (e.g., A100)
	MemoryUsageWarningPercent   = 90.0  // Warn if memory usage exceeds 90%

	// Temperature thresholds
	TemperatureWarningC  = 85.0 // Warning temperature in Celsius
	TemperatureCriticalC = 95.0 // Critical temperature in Celsius

	// Utilization thresholds
	UtilizationHighPercent      = 95.0 // High utilization threshold
	UtilizationActivePercent    = 10.0 // Minimum utilization to consider GPU active
	UtilizationAvailablePercent = 90.0 // Maximum utilization to consider GPU available
)

// GPUMonitor monitors GPU resources on a single node
type GPUMonitor struct {
	clientset kubernetes.Interface
	nodeName  string
	namespace string
	stopCh    chan struct{}
	logger    *log.Logger
}

// NewGPUMonitor creates a new GPU monitor for a node
func NewGPUMonitor(clientset kubernetes.Interface, nodeName, namespace string) *GPUMonitor {
	// Create structured logger with node context
	logger := log.New(os.Stderr, fmt.Sprintf("[GPU-Monitor-%s] ", nodeName), log.LstdFlags|log.Lshortfile)

	return &GPUMonitor{
		clientset: clientset,
		nodeName:  nodeName,
		namespace: namespace,
		stopCh:    make(chan struct{}),
		logger:    logger,
	}
}

// Start begins monitoring GPU resources on this node
func (gm *GPUMonitor) Start(ctx context.Context) error {
	gm.logger.Printf("INFO: Starting GPU monitor for node %s", gm.nodeName)

	// Initialize node with GPU information
	err := gm.initializeNode()
	if err != nil {
		gm.logger.Printf("ERROR: Failed to initialize node: %v", err)
		return fmt.Errorf("failed to initialize node: %v", err)
	}

	gm.logger.Printf("INFO: Node initialization complete, starting monitoring loop")

	// Start monitoring loop
	go gm.monitoringLoop(ctx)

	return nil
}

// Stop gracefully stops the GPU monitor
func (gm *GPUMonitor) Stop() {
	gm.logger.Printf("INFO: Stopping GPU monitor for node %s", gm.nodeName)
	close(gm.stopCh)
}

// initializeNode discovers and registers GPU devices on this node
func (gm *GPUMonitor) initializeNode() error {
	gpuDevices, err := gm.discoverGPUDevices()
	if err != nil {
		return fmt.Errorf("failed to discover GPU devices: %v", err)
	}

	if len(gpuDevices) == 0 {
		gm.logger.Printf("WARNING: No GPU devices found on node %s", gm.nodeName)
		return fmt.Errorf("no GPU devices found on node %s", gm.nodeName)
	}

	gm.logger.Printf("INFO: Discovered %d GPU device(s) on node %s", len(gpuDevices), gm.nodeName)

	// Update node annotations with GPU information
	return gm.updateNodeAnnotations(gpuDevices)
}

// discoverGPUDevices discovers GPU devices using nvidia-smi
func (gm *GPUMonitor) discoverGPUDevices() ([]GPUDevice, error) {
	// Validate nvidia-smi is available and secure
	nvidiaSmiPath, err := exec.LookPath("nvidia-smi")
	if err != nil {
		return nil, fmt.Errorf("nvidia-smi not found in PATH: %v", err)
	}

	// Ensure nvidia-smi exists and is executable
	if _, err := os.Stat(nvidiaSmiPath); err != nil {
		return nil, fmt.Errorf("nvidia-smi not accessible: %v", err)
	}

	// Query GPU information using nvidia-smi with validated path
	cmd := exec.Command(nvidiaSmiPath,
		"--query-gpu=index,name,memory.total,pci.bus_id,driver_version",
		"--format=csv,noheader,nounits")

	// Set environment variables to prevent injection
	cmd.Env = []string{
		"PATH=/usr/bin:/bin:/usr/local/bin",
		"LC_ALL=C",
	}

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("nvidia-smi command failed: %v", err)
	}

	return gm.parseNvidiaSmiOutput(string(output))
}

// parseNvidiaSmiOutput parses nvidia-smi output into GPU devices
func (gm *GPUMonitor) parseNvidiaSmiOutput(output string) ([]GPUDevice, error) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	devices := make([]GPUDevice, 0, len(lines))

	for _, line := range lines {
		fields := strings.Split(line, ", ")
		if len(fields) != 5 {
			continue
		}

		index := strings.TrimSpace(fields[0])
		name := strings.TrimSpace(fields[1])
		memoryStr := strings.TrimSpace(fields[2])
		pciBusID := strings.TrimSpace(fields[3])
		driverVersion := strings.TrimSpace(fields[4])

		memory, err := strconv.ParseInt(memoryStr, 10, 64)
		if err != nil {
			continue
		}

		device := GPUDevice{
			ID:            fmt.Sprintf("gpu-%s", index),
			Name:          name,
			MemoryTotal:   memory,
			PCIBusID:      pciBusID,
			DriverVersion: driverVersion,
		}

		devices = append(devices, device)
	}

	return devices, nil
}

// updateNodeAnnotations updates the node with GPU device information
func (gm *GPUMonitor) updateNodeAnnotations(devices []GPUDevice) error {
	node, err := gm.clientset.CoreV1().Nodes().Get(context.TODO(), gm.nodeName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get node: %v", err)
	}

	if node.Annotations == nil {
		node.Annotations = make(map[string]string)
	}

	// Add GPU annotations
	node.Annotations["agentaflow.gpu/enabled"] = "true"
	node.Annotations["agentaflow.gpu/count"] = strconv.Itoa(len(devices))

	devicesJSON, err := json.Marshal(devices)
	if err != nil {
		return fmt.Errorf("failed to marshal GPU devices: %v", err)
	}
	node.Annotations["agentaflow.gpu/devices"] = string(devicesJSON)

	// Add labels for scheduling
	if node.Labels == nil {
		node.Labels = make(map[string]string)
	}
	node.Labels["agentaflow.gpu/enabled"] = "true"
	node.Labels["agentaflow.gpu/count"] = strconv.Itoa(len(devices))

	// Update the node
	_, err = gm.clientset.CoreV1().Nodes().Update(context.TODO(), node, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update node: %v", err)
	}

	return nil
}

// monitoringLoop continuously monitors GPU status
func (gm *GPUMonitor) monitoringLoop(ctx context.Context) {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-gm.stopCh:
			return
		case <-ticker.C:
			gm.updateGPUStatus()
		}
	}
}

// updateGPUStatus updates the current GPU status information
func (gm *GPUMonitor) updateGPUStatus() {
	gpuStatuses, err := gm.getGPUStatuses()
	if err != nil {
		gm.logger.Printf("ERROR: Failed to get GPU statuses: %v", err)
		return
	}

	err = gm.updateNodeStatus(gpuStatuses)
	if err != nil {
		gm.logger.Printf("ERROR: Failed to update node status: %v", err)
	}
}

// getGPUStatuses retrieves current GPU utilization and memory usage
func (gm *GPUMonitor) getGPUStatuses() ([]GPUStatus, error) {
	// Validate nvidia-smi path for security
	nvidiaSmiPath, err := exec.LookPath("nvidia-smi")
	if err != nil {
		return nil, fmt.Errorf("nvidia-smi not found: %v", err)
	}

	// Query current GPU status
	cmd := exec.Command(nvidiaSmiPath,
		"--query-gpu=index,utilization.gpu,memory.used,memory.total,temperature.gpu,power.draw",
		"--format=csv,noheader,nounits")

	// Set secure environment
	cmd.Env = []string{
		"PATH=/usr/bin:/bin:/usr/local/bin",
		"LC_ALL=C",
	}

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("nvidia-smi status query failed: %v", err)
	}

	return gm.parseGPUStatusOutput(string(output))
}

// parseGPUStatusOutput parses nvidia-smi status output
func (gm *GPUMonitor) parseGPUStatusOutput(output string) ([]GPUStatus, error) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	statuses := make([]GPUStatus, 0, len(lines))

	for _, line := range lines {
		fields := strings.Split(line, ", ")
		if len(fields) != 6 {
			continue
		}

		index := strings.TrimSpace(fields[0])
		utilizationStr := strings.TrimSpace(fields[1])
		memoryUsedStr := strings.TrimSpace(fields[2])
		// Skip fields[3] (memoryTotalStr) - total memory is already tracked in GPUDevice
		temperatureStr := strings.TrimSpace(fields[4])
		powerStr := strings.TrimSpace(fields[5])

		utilization, _ := strconv.ParseFloat(utilizationStr, 64)
		memoryUsed, _ := strconv.ParseInt(memoryUsedStr, 10, 64)
		temperature, _ := strconv.ParseFloat(temperatureStr, 64)
		power, _ := strconv.ParseFloat(powerStr, 64)

		status := GPUStatus{
			ID:          fmt.Sprintf("gpu-%s", index),
			Available:   utilization < UtilizationAvailablePercent,
			MemoryUsed:  memoryUsed,
			Utilization: utilization,
			Temperature: temperature,
			PowerUsage:  power,
		}

		// Check if GPU is being used by a workload
		if utilization > UtilizationActivePercent {
			status.CurrentWorkload = gm.findWorkloadUsingGPU(status.ID)
		}

		statuses = append(statuses, status)
	}

	return statuses, nil
}

// findWorkloadUsingGPU finds which workload is currently using a GPU
func (gm *GPUMonitor) findWorkloadUsingGPU(gpuID string) string {
	// Query pods on this node that might be using GPUs
	pods, err := gm.clientset.CoreV1().Pods(gm.namespace).List(context.TODO(), metav1.ListOptions{
		FieldSelector: fmt.Sprintf("spec.nodeName=%s", gm.nodeName),
		LabelSelector: "agentaflow.gpu/managed=true",
	})
	if err != nil {
		return ""
	}

	for _, pod := range pods.Items {
		if assignedGPU, exists := pod.Annotations["agentaflow.gpu/assigned-gpu"]; exists {
			if strings.Contains(assignedGPU, gpuID) {
				return pod.Annotations["agentaflow.gpu/workload"]
			}
		}
	}

	return ""
}

// updateNodeStatus updates the node with current GPU status
func (gm *GPUMonitor) updateNodeStatus(statuses []GPUStatus) error {
	node, err := gm.clientset.CoreV1().Nodes().Get(context.TODO(), gm.nodeName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get node: %v", err)
	}

	if node.Annotations == nil {
		node.Annotations = make(map[string]string)
	}

	// Update GPU status annotations
	statusJSON, err := json.Marshal(statuses)
	if err != nil {
		return fmt.Errorf("failed to marshal GPU statuses: %v", err)
	}
	node.Annotations["agentaflow.gpu/status"] = string(statusJSON)
	node.Annotations["agentaflow.gpu/last-update"] = time.Now().Format(time.RFC3339)

	// Calculate overall node metrics
	totalUtilization := 0.0
	availableGPUs := 0
	for _, status := range statuses {
		totalUtilization += status.Utilization
		if status.Available {
			availableGPUs++
		}
	}

	avgUtilization := 0.0
	if len(statuses) > 0 {
		avgUtilization = totalUtilization / float64(len(statuses))
	}

	node.Annotations["agentaflow.gpu/average-utilization"] = fmt.Sprintf("%.2f", avgUtilization)
	node.Annotations["agentaflow.gpu/available-count"] = strconv.Itoa(availableGPUs)

	// Update node labels for scheduling decisions
	if node.Labels == nil {
		node.Labels = make(map[string]string)
	}

	// Mark node as schedulable if any GPUs are available
	if availableGPUs > 0 {
		node.Labels["agentaflow.gpu/schedulable"] = "true"
	} else {
		node.Labels["agentaflow.gpu/schedulable"] = "false"
	}

	// Set utilization tier labels for scheduling preferences
	if avgUtilization < 25.0 {
		node.Labels["agentaflow.gpu/utilization-tier"] = "low"
	} else if avgUtilization < 75.0 {
		node.Labels["agentaflow.gpu/utilization-tier"] = "medium"
	} else {
		node.Labels["agentaflow.gpu/utilization-tier"] = "high"
	}

	// Update the node
	_, err = gm.clientset.CoreV1().Nodes().Update(context.TODO(), node, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update node: %v", err)
	}

	return nil
}

// GetNodeGPUStatus returns the current GPU status for this node
func (gm *GPUMonitor) GetNodeGPUStatus() ([]GPUStatus, error) {
	return gm.getGPUStatuses()
}

// CheckGPUHealth performs health checks on GPU devices
func (gm *GPUMonitor) CheckGPUHealth() (*GPUHealthReport, error) {
	statuses, err := gm.getGPUStatuses()
	if err != nil {
		return nil, err
	}

	report := &GPUHealthReport{
		NodeName:    gm.nodeName,
		CheckTime:   time.Now(),
		GPUCount:    len(statuses),
		HealthyGPUs: 0,
		Issues:      make([]GPUHealthIssue, 0),
	}

	for _, status := range statuses {
		healthy := true

		// Check temperature
		if status.Temperature > TemperatureWarningC {
			report.Issues = append(report.Issues, GPUHealthIssue{
				GPUID:    status.ID,
				Severity: "warning",
				Issue:    "High temperature",
				Value:    fmt.Sprintf("%.1f°C", status.Temperature),
			})
			if status.Temperature > TemperatureCriticalC {
				healthy = false
			}
		}

		// Check utilization
		if status.Utilization > UtilizationHighPercent {
			report.Issues = append(report.Issues, GPUHealthIssue{
				GPUID:    status.ID,
				Severity: "info",
				Issue:    "High utilization",
				Value:    fmt.Sprintf("%.1f%%", status.Utilization),
			})
		}

		// Check memory usage - use configurable thresholds
		if status.MemoryUsed > MemoryUsageAlertThresholdMB {
			memoryUsagePercent := float64(status.MemoryUsed) / DefaultGPUMemoryTotalMB * 100
			if memoryUsagePercent > MemoryUsageWarningPercent {
				report.Issues = append(report.Issues, GPUHealthIssue{
					GPUID:    status.ID,
					Severity: "warning",
					Issue:    "High memory usage",
					Value:    fmt.Sprintf("%.1f%%", memoryUsagePercent),
				})
			}
		}

		if healthy {
			report.HealthyGPUs++
		}
	}

	report.OverallHealth = "healthy"
	if report.HealthyGPUs < len(statuses) {
		if report.HealthyGPUs == 0 {
			report.OverallHealth = "critical"
		} else {
			report.OverallHealth = "degraded"
		}
	}

	return report, nil
}

// GPUHealthReport represents the health status of GPUs on a node
type GPUHealthReport struct {
	NodeName      string           `json:"nodeName"`
	CheckTime     time.Time        `json:"checkTime"`
	GPUCount      int              `json:"gpuCount"`
	HealthyGPUs   int              `json:"healthyGpus"`
	OverallHealth string           `json:"overallHealth"`
	Issues        []GPUHealthIssue `json:"issues"`
}

// GPUHealthIssue represents a health issue with a GPU
type GPUHealthIssue struct {
	GPUID    string `json:"gpuId"`
	Severity string `json:"severity"`
	Issue    string `json:"issue"`
	Value    string `json:"value"`
}
