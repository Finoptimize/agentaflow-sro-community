package k8s

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Finoptimize/agentaflow-sro-community/pkg/gpu"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GPUSchedulerCLI provides command-line interface for the Kubernetes GPU scheduler
type GPUSchedulerCLI struct {
	scheduler *KubernetesGPUScheduler
}

// NewGPUSchedulerCLI creates a new CLI interface
func NewGPUSchedulerCLI(scheduler *KubernetesGPUScheduler) *GPUSchedulerCLI {
	return &GPUSchedulerCLI{
		scheduler: scheduler,
	}
}

// ExecuteCommand executes a CLI command
func (cli *GPUSchedulerCLI) ExecuteCommand(args []string) error {
	if len(args) == 0 {
		return cli.showHelp()
	}

	command := args[0]
	switch command {
	case "status":
		return cli.showStatus()
	case "nodes":
		return cli.listNodes()
	case "workloads":
		return cli.listWorkloads()
	case "submit":
		if len(args) < 2 {
			return fmt.Errorf("submit command requires a workload file")
		}
		return cli.submitWorkload(args[1])
	case "complete":
		if len(args) < 2 {
			return fmt.Errorf("complete command requires a workload name")
		}
		return cli.completeWorkload(args[1])
	case "metrics":
		return cli.showMetrics()
	case "strategy":
		if len(args) < 2 {
			return cli.showCurrentStrategy()
		}
		return cli.setStrategy(args[1])
	case "watch":
		return cli.watchStatus()
	case "health":
		return cli.showHealthStatus()
	case "help":
		return cli.showHelp()
	default:
		return fmt.Errorf("unknown command: %s", command)
	}
}

// showHelp displays help information
func (cli *GPUSchedulerCLI) showHelp() error {
	help := `AgentaFlow GPU Scheduler CLI

COMMANDS:
  status               Show overall scheduler status
  nodes                List GPU-enabled nodes
  workloads            List all GPU workloads
  submit <file>        Submit a workload from YAML file
  complete <name>      Mark a workload as completed
  metrics              Show detailed scheduling metrics
  strategy [name]      Show or set scheduling strategy
  watch                Watch status updates in real-time
  health               Show GPU health status
  help                 Show this help message

SCHEDULING STRATEGIES:
  least_utilized       Schedule on least utilized GPUs (default)
  best_fit            Schedule on GPU with just enough free memory
  priority            Schedule based on workload priority
  round_robin         Distribute workloads evenly

EXAMPLES:
  agentaflow-k8s status
  agentaflow-k8s submit workload.yaml
  agentaflow-k8s strategy least_utilized
  agentaflow-k8s complete training-job-1
`
	fmt.Print(help)
	return nil
}

// showStatus displays the current scheduler status
func (cli *GPUSchedulerCLI) showStatus() error {
	metrics := cli.scheduler.GetSchedulingMetrics()

	fmt.Println("=== AgentaFlow GPU Scheduler Status ===")
	fmt.Printf("Cluster Overview:\n")
	fmt.Printf("  Total Nodes:     %d\n", metrics.TotalNodes)
	fmt.Printf("  Active Nodes:    %d\n", metrics.ActiveNodes)
	fmt.Printf("  Total GPUs:      %d\n", metrics.TotalGPUs)
	fmt.Printf("  Available GPUs:  %d\n", metrics.AvailableGPUs)
	fmt.Printf("  Utilized GPUs:   %d\n", metrics.UtilizedGPUs)
	fmt.Printf("\n")

	fmt.Printf("Resource Utilization:\n")
	fmt.Printf("  Average GPU Utilization: %.1f%%\n", metrics.AverageUtilization)
	fmt.Printf("  Memory Utilization:      %.1f%%\n", metrics.MemoryUtilization)
	fmt.Printf("\n")

	fmt.Printf("Workload Status:\n")
	fmt.Printf("  Pending:     %d\n", metrics.PendingWorkloads)
	fmt.Printf("  Running:     %d\n", metrics.RunningWorkloads)
	fmt.Printf("  Completed:   %d\n", metrics.CompletedWorkloads)
	fmt.Printf("\n")

	fmt.Printf("Last Updated: %s\n", metrics.LastUpdateTime.Format(time.RFC3339))

	return nil
}

// listNodes displays all GPU-enabled nodes
func (cli *GPUSchedulerCLI) listNodes() error {
	nodes := cli.scheduler.ListGPUNodes()

	fmt.Println("=== GPU-Enabled Nodes ===")
	if len(nodes) == 0 {
		fmt.Println("No GPU nodes found")
		return nil
	}

	fmt.Printf("%-20s %-10s %-8s %-15s %-15s\n", "NODE", "STATUS", "GPUS", "AVG UTIL", "LAST UPDATED")
	fmt.Printf("%-20s %-10s %-8s %-15s %-15s\n",
		strings.Repeat("-", 20),
		strings.Repeat("-", 10),
		strings.Repeat("-", 8),
		strings.Repeat("-", 15),
		strings.Repeat("-", 15))

	for _, node := range nodes {
		gpuCount := len(node.Spec.GPUDevices)
		avgUtil := "N/A"

		if len(node.Status.GPUStatus) > 0 {
			totalUtil := 0.0
			for _, gpu := range node.Status.GPUStatus {
				totalUtil += gpu.Utilization
			}
			avgUtil = fmt.Sprintf("%.1f%%", totalUtil/float64(len(node.Status.GPUStatus)))
		}

		fmt.Printf("%-20s %-10s %-8d %-15s %-15s\n",
			node.Name,
			node.Status.Phase,
			gpuCount,
			avgUtil,
			node.Status.LastUpdated.Format("15:04:05"))
	}

	return nil
}

// listWorkloads displays all GPU workloads
func (cli *GPUSchedulerCLI) listWorkloads() error {
	workloads := cli.scheduler.ListGPUWorkloads()

	fmt.Println("=== GPU Workloads ===")
	if len(workloads) == 0 {
		fmt.Println("No workloads found")
		return nil
	}

	fmt.Printf("%-25s %-12s %-8s %-15s %-20s %-15s\n",
		"NAME", "STATUS", "PRIORITY", "GPU MEMORY", "ASSIGNED GPU", "ASSIGNED NODE")
	fmt.Printf("%-25s %-12s %-8s %-15s %-20s %-15s\n",
		strings.Repeat("-", 25),
		strings.Repeat("-", 12),
		strings.Repeat("-", 8),
		strings.Repeat("-", 15),
		strings.Repeat("-", 20),
		strings.Repeat("-", 15))

	for _, workload := range workloads {
		gpuMemory := fmt.Sprintf("%d MB", workload.Spec.GPUMemoryRequired)
		assignedGPU := workload.Status.AssignedGPU
		if assignedGPU == "" {
			assignedGPU = "None"
		}
		assignedNode := workload.Status.AssignedNode
		if assignedNode == "" {
			assignedNode = "None"
		}

		fmt.Printf("%-25s %-12s %-8d %-15s %-20s %-15s\n",
			workload.Name,
			workload.Status.Phase,
			workload.Spec.Priority,
			gpuMemory,
			assignedGPU,
			assignedNode)
	}

	return nil
}

// submitWorkload submits a workload from a YAML file
func (cli *GPUSchedulerCLI) submitWorkload(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read workload file: %v", err)
	}

	var workload GPUWorkload
	err = yaml.Unmarshal(data, &workload)
	if err != nil {
		return fmt.Errorf("failed to parse workload YAML: %v", err)
	}

	// Set default values if not specified
	if workload.Spec.GPURequirements.GPUCount == 0 {
		workload.Spec.GPURequirements.GPUCount = 1
	}

	err = cli.scheduler.SubmitGPUWorkload(&workload)
	if err != nil {
		return fmt.Errorf("failed to submit workload: %v", err)
	}

	fmt.Printf("Workload '%s' submitted successfully\n", workload.Name)
	return nil
}

// completeWorkload marks a workload as completed
func (cli *GPUSchedulerCLI) completeWorkload(workloadName string) error {
	err := cli.scheduler.CompleteWorkload(workloadName)
	if err != nil {
		return fmt.Errorf("failed to complete workload: %v", err)
	}

	fmt.Printf("Workload '%s' marked as completed\n", workloadName)
	return nil
}

// showMetrics displays detailed scheduling metrics
func (cli *GPUSchedulerCLI) showMetrics() error {
	metrics := cli.scheduler.GetSchedulingMetrics()

	metricsJSON, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to format metrics: %v", err)
	}

	fmt.Println("=== Detailed Scheduling Metrics ===")
	fmt.Println(string(metricsJSON))
	return nil
}

// showCurrentStrategy displays the current scheduling strategy
func (cli *GPUSchedulerCLI) showCurrentStrategy() error {
	// This would need to be implemented in the scheduler to expose the current strategy
	fmt.Println("Current scheduling strategy: least_utilized")
	fmt.Println("\nAvailable strategies:")
	fmt.Println("  - least_utilized: Schedule on least utilized GPUs")
	fmt.Println("  - best_fit: Schedule on GPU with just enough free memory")
	fmt.Println("  - priority: Schedule based on workload priority")
	fmt.Println("  - round_robin: Distribute workloads evenly")
	return nil
}

// setStrategy changes the scheduling strategy
func (cli *GPUSchedulerCLI) setStrategy(strategyName string) error {
	var strategy gpu.SchedulingStrategy

	switch strategyName {
	case "least_utilized":
		strategy = gpu.StrategyLeastUtilized
	case "best_fit":
		strategy = gpu.StrategyBestFit
	case "priority":
		strategy = gpu.StrategyPriority
	case "round_robin":
		strategy = gpu.StrategyRoundRobin
	default:
		return fmt.Errorf("unknown strategy: %s", strategyName)
	}

	cli.scheduler.SetSchedulingStrategy(strategy)
	fmt.Printf("Scheduling strategy changed to: %s\n", strategyName)
	return nil
}

// watchStatus continuously displays status updates
func (cli *GPUSchedulerCLI) watchStatus() error {
	fmt.Println("=== Watching GPU Scheduler Status (Press Ctrl+C to stop) ===")

	for {
		// Clear screen
		fmt.Print("\033[2J\033[H")

		err := cli.showStatus()
		if err != nil {
			return err
		}

		fmt.Println("\n--- Nodes ---")
		err = cli.listNodes()
		if err != nil {
			return err
		}

		fmt.Println("\n--- Active Workloads ---")
		err = cli.listWorkloads()
		if err != nil {
			return err
		}

		time.Sleep(5 * time.Second)
	}
}

// showHealthStatus displays GPU health information
func (cli *GPUSchedulerCLI) showHealthStatus() error {
	nodes := cli.scheduler.ListGPUNodes()

	fmt.Println("=== GPU Health Status ===")
	if len(nodes) == 0 {
		fmt.Println("No GPU nodes found")
		return nil
	}

	for _, node := range nodes {
		fmt.Printf("\nNode: %s\n", node.Name)
		fmt.Printf("Status: %s\n", node.Status.Phase)

		if len(node.Status.GPUStatus) == 0 {
			fmt.Printf("  No GPU status available\n")
			continue
		}

		fmt.Printf("GPUs:\n")
		for _, gpuStatus := range node.Status.GPUStatus {
			status := "Healthy"
			if gpuStatus.Temperature > 85.0 {
				status = "Warning (High Temperature)"
			}
			if gpuStatus.Utilization > 95.0 {
				status = "High Utilization"
			}

			fmt.Printf("  %s:\n", gpuStatus.ID)
			fmt.Printf("    Status: %s\n", status)
			fmt.Printf("    Utilization: %.1f%%\n", gpuStatus.Utilization)
			fmt.Printf("    Memory Used: %d MB\n", gpuStatus.MemoryUsed)
			fmt.Printf("    Temperature: %.1f°C\n", gpuStatus.Temperature)
			fmt.Printf("    Power Usage: %.1fW\n", gpuStatus.PowerUsage)

			if gpuStatus.CurrentWorkload != "" {
				fmt.Printf("    Current Workload: %s\n", gpuStatus.CurrentWorkload)
			}
		}
	}

	return nil
}

// GenerateWorkloadTemplate generates a sample workload YAML file
func (cli *GPUSchedulerCLI) GenerateWorkloadTemplate(filename string) error {
	template := &GPUWorkload{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "agentaflow.io/v1",
			Kind:       "GPUWorkload",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "example-training-job",
		},
		Spec: GPUWorkloadSpec{
			Priority:          1,
			GPUMemoryRequired: 8192, // 8GB
			EstimatedDuration: &metav1.Duration{Duration: 2 * time.Hour},
			GPURequirements: GPURequirements{
				MinGPUMemory:    8192,
				GPUCount:        1,
				ExclusiveAccess: true,
			},
			SchedulingStrategy: "least_utilized",
			PodTemplate: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  "training-container",
							Image: "tensorflow/tensorflow:latest-gpu",
							Command: []string{
								"python",
								"/app/train.py",
							},
							Resources: v1.ResourceRequirements{
								Requests: v1.ResourceList{
									"nvidia.com/gpu": *resource.NewQuantity(1, resource.DecimalSI),
								},
								Limits: v1.ResourceList{
									"nvidia.com/gpu": *resource.NewQuantity(1, resource.DecimalSI),
								},
							},
						},
					},
					RestartPolicy: v1.RestartPolicyNever,
				},
			},
		},
	}

	data, err := yaml.Marshal(template)
	if err != nil {
		return fmt.Errorf("failed to marshal template: %v", err)
	}

	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write template file: %v", err)
	}

	fmt.Printf("Workload template written to %s\n", filename)
	return nil
}
