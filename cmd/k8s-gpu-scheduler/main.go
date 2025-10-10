package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Finoptimize/agentaflow-sro-community/pkg/gpu"
	"github.com/Finoptimize/agentaflow-sro-community/pkg/k8s"
)

func main() {
	var (
		namespace = flag.String("namespace", "agentaflow", "Kubernetes namespace to use")
		strategy  = flag.String("strategy", "least_utilized", "Scheduling strategy to use")
		mode      = flag.String("mode", "scheduler", "Mode to run in: scheduler, monitor, cli")
		nodeName  = flag.String("node", "", "Node name for monitor mode")
	)
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown gracefully
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		log.Println("Received shutdown signal, stopping...")
		cancel()
	}()

	switch *mode {
	case "scheduler":
		err := runScheduler(ctx, *namespace, *strategy)
		if err != nil {
			log.Fatalf("Scheduler failed: %v", err)
		}
	case "monitor":
		if *nodeName == "" {
			log.Fatal("Node name is required for monitor mode")
		}
		err := runMonitor(ctx, *nodeName, *namespace)
		if err != nil {
			log.Fatalf("Monitor failed: %v", err)
		}
	case "cli":
		err := runCLI(*namespace, *strategy, flag.Args())
		if err != nil {
			log.Fatalf("CLI command failed: %v", err)
		}
	default:
		log.Fatalf("Unknown mode: %s", *mode)
	}
}

// runScheduler runs the Kubernetes GPU scheduler
func runScheduler(ctx context.Context, namespace, strategyName string) error {
	log.Printf("Starting AgentaFlow GPU Scheduler in namespace '%s'", namespace)

	// Parse strategy
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
		return fmt.Errorf("unknown scheduling strategy: %s", strategyName)
	}

	// Create scheduler
	scheduler, err := k8s.NewKubernetesGPUScheduler(namespace, strategy)
	if err != nil {
		return fmt.Errorf("failed to create scheduler: %v", err)
	}

	log.Printf("Using scheduling strategy: %s", strategyName)

	// Start scheduler
	err = scheduler.Start(ctx)
	if err != nil {
		return fmt.Errorf("failed to start scheduler: %v", err)
	}

	log.Println("GPU Scheduler started successfully")

	// Print status periodically
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Shutting down scheduler...")
			scheduler.Stop()
			return nil
		case <-ticker.C:
			printSchedulerStatus(scheduler)
		}
	}
}

// runMonitor runs the GPU monitor on a specific node
func runMonitor(ctx context.Context, nodeName, namespace string) error {
	log.Printf("Starting GPU Monitor for node '%s'", nodeName)

	// Create Kubernetes client
	scheduler, err := k8s.NewKubernetesGPUScheduler(namespace, gpu.StrategyLeastUtilized)
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %v", err)
	}

	// Create monitor
	clientset := scheduler.GetClientset()
	monitor := k8s.NewGPUMonitor(clientset, nodeName, namespace)

	// Start monitor
	err = monitor.Start(ctx)
	if err != nil {
		return fmt.Errorf("failed to start monitor: %v", err)
	}

	log.Println("GPU Monitor started successfully")

	// Print health status periodically
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Shutting down monitor...")
			monitor.Stop()
			return nil
		case <-ticker.C:
			printHealthStatus(monitor, nodeName)
		}
	}
}

// runCLI runs CLI commands
func runCLI(namespace, strategyName string, args []string) error {
	// Parse strategy
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
		return fmt.Errorf("unknown scheduling strategy: %s", strategyName)
	}

	// Create scheduler
	scheduler, err := k8s.NewKubernetesGPUScheduler(namespace, strategy)
	if err != nil {
		return fmt.Errorf("failed to create scheduler: %v", err)
	}

	// Create CLI
	cli := k8s.NewGPUSchedulerCLI(scheduler)

	// Handle special CLI commands
	if len(args) > 0 && args[0] == "generate-template" {
		filename := "workload-template.yaml"
		if len(args) > 1 {
			filename = args[1]
		}
		return cli.GenerateWorkloadTemplate(filename)
	}

	// Execute CLI command
	return cli.ExecuteCommand(args)
}

// printSchedulerStatus prints current scheduler status
func printSchedulerStatus(scheduler *k8s.KubernetesGPUScheduler) {
	metrics := scheduler.GetSchedulingMetrics()
	log.Printf("Status - Nodes: %d/%d, GPUs: %d/%d, Workloads: %d pending, %d running",
		metrics.ActiveNodes, metrics.TotalNodes,
		metrics.AvailableGPUs, metrics.TotalGPUs,
		metrics.PendingWorkloads, metrics.RunningWorkloads)
}

// printHealthStatus prints GPU health status for a node
func printHealthStatus(monitor *k8s.GPUMonitor, nodeName string) {
	report, err := monitor.CheckGPUHealth()
	if err != nil {
		log.Printf("Failed to get health report for node %s: %v", nodeName, err)
		return
	}

	log.Printf("Health - Node: %s, GPUs: %d/%d healthy, Status: %s",
		report.NodeName, report.HealthyGPUs, report.GPUCount, report.OverallHealth)

	if len(report.Issues) > 0 {
		log.Printf("Issues detected: %d", len(report.Issues))
		for _, issue := range report.Issues {
			log.Printf("  - %s on %s: %s (%s)", issue.Issue, issue.GPUID, issue.Value, issue.Severity)
		}
	}
}
