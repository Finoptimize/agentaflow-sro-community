package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	var example = flag.String("example", "all", "Example to run: gpu, serving, observability, or all")
	flag.Parse()

	fmt.Println("=== AgentaFlow SRO Examples ===")
	fmt.Println()

	switch *example {
	case "gpu":
		fmt.Println("Running GPU Scheduling Example...")
		runGPUSchedulingExample()
	case "serving":
		fmt.Println("Running Model Serving Example...")
		runModelServingExample()
	case "observability":
		fmt.Println("Running Observability Example...")
		runObservabilityExample()
	case "all":
		fmt.Println("Running All Examples...")
		fmt.Println()
		fmt.Println("--- GPU Scheduling ---")
		runGPUSchedulingExample()
		fmt.Println("\n--- Model Serving ---")
		runModelServingExample()
		fmt.Println("\n--- Observability ---")
		runObservabilityExample()
	default:
		fmt.Printf("Unknown example: %s\n", *example)
		fmt.Println("Available examples: gpu, serving, observability, all")
		os.Exit(1)
	}

	fmt.Println("\n=== Examples Complete ===")
}
