package k8s

import (
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// GPUWorkload represents a Kubernetes workload that requires GPU resources
type GPUWorkload struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GPUWorkloadSpec   `json:"spec,omitempty"`
	Status GPUWorkloadStatus `json:"status,omitempty"`
}

// GPUWorkloadSpec defines the desired state of GPUWorkload
type GPUWorkloadSpec struct {
	// Priority of the workload (0-10, higher is more important)
	Priority int32 `json:"priority,omitempty"`

	// GPU memory required in MB
	GPUMemoryRequired int64 `json:"gpuMemoryRequired"`

	// Estimated execution time
	EstimatedDuration *metav1.Duration `json:"estimatedDuration,omitempty"`

	// Pod template for the workload
	PodTemplate v1.PodTemplateSpec `json:"podTemplate"`

	// GPU requirements
	GPURequirements GPURequirements `json:"gpuRequirements"`

	// Scheduling strategy preference
	SchedulingStrategy string `json:"schedulingStrategy,omitempty"`
}

// GPURequirements specifies GPU resource requirements
type GPURequirements struct {
	// Minimum GPU memory in MB
	MinGPUMemory int64 `json:"minGPUMemory"`

	// Preferred GPU type (optional)
	PreferredGPUType string `json:"preferredGPUType,omitempty"`

	// Number of GPUs required
	GPUCount int32 `json:"gpuCount,omitempty"`

	// Exclusive access requirement
	ExclusiveAccess bool `json:"exclusiveAccess,omitempty"`
}

// GPUWorkloadStatus defines the observed state of GPUWorkload
type GPUWorkloadStatus struct {
	// Phase represents the current phase of the workload
	Phase GPUWorkloadPhase `json:"phase,omitempty"`

	// AssignedNode is the node where the workload is scheduled
	AssignedNode string `json:"assignedNode,omitempty"`

	// AssignedGPU is the GPU ID assigned to this workload
	AssignedGPU string `json:"assignedGPU,omitempty"`

	// StartTime is when the workload started executing
	StartTime *metav1.Time `json:"startTime,omitempty"`

	// CompletionTime is when the workload completed
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`

	// Conditions represent the current conditions
	Conditions []GPUWorkloadCondition `json:"conditions,omitempty"`

	// Message provides additional information about the workload status
	Message string `json:"message,omitempty"`
}

// GPUWorkloadPhase represents the phase of a GPU workload
type GPUWorkloadPhase string

const (
	GPUWorkloadPending   GPUWorkloadPhase = "Pending"
	GPUWorkloadScheduled GPUWorkloadPhase = "Scheduled"
	GPUWorkloadRunning   GPUWorkloadPhase = "Running"
	GPUWorkloadSucceeded GPUWorkloadPhase = "Succeeded"
	GPUWorkloadFailed    GPUWorkloadPhase = "Failed"
)

// GPUWorkloadCondition represents a condition of a GPU workload
type GPUWorkloadCondition struct {
	Type               GPUWorkloadConditionType `json:"type"`
	Status             v1.ConditionStatus       `json:"status"`
	LastTransitionTime metav1.Time              `json:"lastTransitionTime,omitempty"`
	Reason             string                   `json:"reason,omitempty"`
	Message            string                   `json:"message,omitempty"`
}

// GPUWorkloadConditionType represents the type of condition
type GPUWorkloadConditionType string

const (
	GPUWorkloadSchedulable GPUWorkloadConditionType = "Schedulable"
	GPUWorkloadAssigned    GPUWorkloadConditionType = "Assigned"
	GPUWorkloadReady       GPUWorkloadConditionType = "Ready"
)

// GPUWorkloadList contains a list of GPUWorkload
type GPUWorkloadList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GPUWorkload `json:"items"`
}

// GPUNode represents a Kubernetes node with GPU resources
type GPUNode struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GPUNodeSpec   `json:"spec,omitempty"`
	Status GPUNodeStatus `json:"status,omitempty"`
}

// GPUNodeSpec defines the desired state of GPUNode
type GPUNodeSpec struct {
	// NodeName is the Kubernetes node name
	NodeName string `json:"nodeName"`

	// GPUDevices lists all GPU devices on this node
	GPUDevices []GPUDevice `json:"gpuDevices"`
}

// GPUDevice represents a GPU device on a node
type GPUDevice struct {
	// ID is the unique identifier for this GPU
	ID string `json:"id"`

	// Name is the GPU model name
	Name string `json:"name"`

	// MemoryTotal is total GPU memory in MB
	MemoryTotal int64 `json:"memoryTotal"`

	// PCIBusID is the PCI bus ID for the GPU
	PCIBusID string `json:"pciBusID,omitempty"`

	// Driver version
	DriverVersion string `json:"driverVersion,omitempty"`
}

// GPUNodeStatus defines the observed state of GPUNode
type GPUNodeStatus struct {
	// Phase represents the current phase of the node
	Phase GPUNodePhase `json:"phase,omitempty"`

	// GPUStatus contains the status of each GPU
	GPUStatus []GPUStatus `json:"gpuStatus,omitempty"`

	// LastUpdated is when the status was last updated
	LastUpdated metav1.Time `json:"lastUpdated,omitempty"`

	// Conditions represent the current conditions
	Conditions []GPUNodeCondition `json:"conditions,omitempty"`
}

// GPUNodePhase represents the phase of a GPU node
type GPUNodePhase string

const (
	GPUNodeActive      GPUNodePhase = "Active"
	GPUNodeUnavailable GPUNodePhase = "Unavailable"
	GPUNodeMaintenance GPUNodePhase = "Maintenance"
)

// GPUStatus represents the current status of a GPU
type GPUStatus struct {
	// ID is the GPU identifier
	ID string `json:"id"`

	// Available indicates if the GPU is available for scheduling
	Available bool `json:"available"`

	// MemoryUsed is current memory usage in MB
	MemoryUsed int64 `json:"memoryUsed"`

	// Utilization is current GPU utilization (0-100%)
	Utilization float64 `json:"utilization"`

	// Temperature in Celsius
	Temperature float64 `json:"temperature,omitempty"`

	// PowerUsage in watts
	PowerUsage float64 `json:"powerUsage,omitempty"`

	// CurrentWorkload is the workload currently using this GPU
	CurrentWorkload string `json:"currentWorkload,omitempty"`
}

// GPUNodeCondition represents a condition of a GPU node
type GPUNodeCondition struct {
	Type               GPUNodeConditionType `json:"type"`
	Status             v1.ConditionStatus   `json:"status"`
	LastTransitionTime metav1.Time          `json:"lastTransitionTime,omitempty"`
	Reason             string               `json:"reason,omitempty"`
	Message            string               `json:"message,omitempty"`
}

// GPUNodeConditionType represents the type of node condition
type GPUNodeConditionType string

const (
	GPUNodeReady       GPUNodeConditionType = "Ready"
	GPUNodeHealthy     GPUNodeConditionType = "Healthy"
	GPUNodeSchedulable GPUNodeConditionType = "Schedulable"
)

// GPUNodeList contains a list of GPUNode
type GPUNodeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GPUNode `json:"items"`
}

// SchedulingMetrics contains metrics about GPU scheduling
type SchedulingMetrics struct {
	TotalNodes         int       `json:"totalNodes"`
	ActiveNodes        int       `json:"activeNodes"`
	TotalGPUs          int       `json:"totalGPUs"`
	AvailableGPUs      int       `json:"availableGPUs"`
	UtilizedGPUs       int       `json:"utilizedGPUs"`
	AverageUtilization float64   `json:"averageUtilization"`
	PendingWorkloads   int       `json:"pendingWorkloads"`
	RunningWorkloads   int       `json:"runningWorkloads"`
	CompletedWorkloads int       `json:"completedWorkloads"`
	MemoryUtilization  float64   `json:"memoryUtilization"`
	LastUpdateTime     time.Time `json:"lastUpdateTime"`
}

// DeepCopyObject implements runtime.Object interface
func (w *GPUWorkload) DeepCopyObject() runtime.Object {
	return w.DeepCopy()
}

// DeepCopy creates a deep copy of GPUWorkload
func (w *GPUWorkload) DeepCopy() *GPUWorkload {
	if w == nil {
		return nil
	}
	out := new(GPUWorkload)
	w.DeepCopyInto(out)
	return out
}

// DeepCopyInto copies the receiver into out
func (w *GPUWorkload) DeepCopyInto(out *GPUWorkload) {
	*out = *w
	out.TypeMeta = w.TypeMeta
	w.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	w.Spec.DeepCopyInto(&out.Spec)
	w.Status.DeepCopyInto(&out.Status)
}

// DeepCopyInto copies the receiver into out
func (spec *GPUWorkloadSpec) DeepCopyInto(out *GPUWorkloadSpec) {
	*out = *spec
	if spec.EstimatedDuration != nil {
		out.EstimatedDuration = new(metav1.Duration)
		*out.EstimatedDuration = *spec.EstimatedDuration
	}
	spec.PodTemplate.DeepCopyInto(&out.PodTemplate)
	out.GPURequirements = spec.GPURequirements
}

// DeepCopyInto copies the receiver into out
func (status *GPUWorkloadStatus) DeepCopyInto(out *GPUWorkloadStatus) {
	*out = *status
	if status.StartTime != nil {
		out.StartTime = new(metav1.Time)
		*out.StartTime = *status.StartTime
	}
	if status.CompletionTime != nil {
		out.CompletionTime = new(metav1.Time)
		*out.CompletionTime = *status.CompletionTime
	}
	if status.Conditions != nil {
		out.Conditions = make([]GPUWorkloadCondition, len(status.Conditions))
		for i := range status.Conditions {
			status.Conditions[i].DeepCopyInto(&out.Conditions[i])
		}
	}
}

// DeepCopyInto copies the receiver into out
func (cond *GPUWorkloadCondition) DeepCopyInto(out *GPUWorkloadCondition) {
	*out = *cond
	cond.LastTransitionTime.DeepCopyInto(&out.LastTransitionTime)
}

// DeepCopyObject implements runtime.Object interface
func (n *GPUNode) DeepCopyObject() runtime.Object {
	return n.DeepCopy()
}

// DeepCopy creates a deep copy of GPUNode
func (n *GPUNode) DeepCopy() *GPUNode {
	if n == nil {
		return nil
	}
	out := new(GPUNode)
	n.DeepCopyInto(out)
	return out
}

// DeepCopyInto copies the receiver into out
func (n *GPUNode) DeepCopyInto(out *GPUNode) {
	*out = *n
	out.TypeMeta = n.TypeMeta
	n.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	n.Spec.DeepCopyInto(&out.Spec)
	n.Status.DeepCopyInto(&out.Status)
}

// DeepCopyInto copies the receiver into out
func (spec *GPUNodeSpec) DeepCopyInto(out *GPUNodeSpec) {
	*out = *spec
	if spec.GPUDevices != nil {
		out.GPUDevices = make([]GPUDevice, len(spec.GPUDevices))
		copy(out.GPUDevices, spec.GPUDevices)
	}
}

// DeepCopyInto copies the receiver into out
func (status *GPUNodeStatus) DeepCopyInto(out *GPUNodeStatus) {
	*out = *status
	if status.GPUStatus != nil {
		out.GPUStatus = make([]GPUStatus, len(status.GPUStatus))
		copy(out.GPUStatus, status.GPUStatus)
	}
	status.LastUpdated.DeepCopyInto(&out.LastUpdated)
	if status.Conditions != nil {
		out.Conditions = make([]GPUNodeCondition, len(status.Conditions))
		for i := range status.Conditions {
			status.Conditions[i].DeepCopyInto(&out.Conditions[i])
		}
	}
}

// DeepCopyInto copies the receiver into out
func (cond *GPUNodeCondition) DeepCopyInto(out *GPUNodeCondition) {
	*out = *cond
	cond.LastTransitionTime.DeepCopyInto(&out.LastTransitionTime)
}
