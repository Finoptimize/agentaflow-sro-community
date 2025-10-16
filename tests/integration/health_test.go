package integration

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

const (
	dashboardURL    = "http://localhost:9000"
	metricsURL      = "http://localhost:9001/metrics"
	prometheusURL   = "http://localhost:9090"
	healthTimeout   = 30 * time.Second
	healthRetryWait = 2 * time.Second
)

// TestHealthEndpoints verifies all services are healthy
func TestHealthEndpoints(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected int
	}{
		{"Dashboard Health", dashboardURL + "/health", http.StatusOK},
		{"Dashboard Root", dashboardURL + "/", http.StatusOK},
		{"Metrics Endpoint", metricsURL, http.StatusOK},
		{"Prometheus Health", prometheusURL + "/-/healthy", http.StatusOK},
		{"Prometheus Ready", prometheusURL + "/-/ready", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := waitForEndpoint(tt.url, healthTimeout)
			if err != nil {
				t.Fatalf("Failed to reach %s: %v", tt.url, err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.expected {
				body, _ := io.ReadAll(resp.Body)
				t.Errorf("Expected status %d, got %d. Response: %s", tt.expected, resp.StatusCode, string(body))
			}
		})
	}
}

// TestPrometheusMetrics verifies Prometheus is scraping metrics
func TestPrometheusMetrics(t *testing.T) {
	// Wait a bit for Prometheus to scrape at least once
	resp, err := waitForEndpoint(prometheusURL+"/-/ready", healthTimeout)
	if err != nil {
		t.Fatalf("Prometheus not ready: %v", err)
	}
	resp.Body.Close()

	// Give Prometheus a moment to scrape at least once after being ready
	time.Sleep(5 * time.Second)

	resp, err := http.Get(metricsURL)
	if err != nil {
		t.Fatalf("Failed to get metrics: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read metrics response: %v", err)
	}

	metrics := string(body)

	// Check for key metrics
	expectedMetrics := []string{
		"gpu_temperature_celsius",
		"gpu_utilization_percent",
		"gpu_memory_used_bytes",
		"gpu_power_usage_watts",
	}

	for _, metric := range expectedMetrics {
		if !strings.Contains(metrics, metric) {
			t.Errorf("Expected metric %s not found in metrics output", metric)
		}
	}
}

// TestDashboardContent verifies dashboard serves HTML content
func TestDashboardContent(t *testing.T) {
	resp, err := http.Get(dashboardURL + "/")
	if err != nil {
		t.Fatalf("Failed to get dashboard: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read dashboard response: %v", err)
	}

	content := string(body)

	// Check for key dashboard elements
	expectedContent := []string{
		"AgentaFlow GPU Monitoring",
		"GPU Performance Metrics",
		"Cost Analytics",
		"<!DOCTYPE html>",
	}

	for _, expected := range expectedContent {
		if !strings.Contains(content, expected) {
	targetFound, targets := waitForAgentaFlowTarget(prometheusURL+"/api/v1/targets", healthTimeout, healthRetryWait)
	if !targetFound {
		t.Fatalf("AgentaFlow target not found in Prometheus targets after %v. Last response: %s", healthTimeout, targets)
	}
}

<<<<<<< Updated upstream
<<<<<<< Updated upstream
// TestPrometheusTargets verifies Prometheus is scraping AgentaFlow
func TestPrometheusTargets(t *testing.T) {
	targetFound, targets := waitForAgentaFlowTarget(prometheusURL+"/api/v1/targets", healthTimeout, healthRetryWait)
	if !targetFound {
		t.Fatalf("AgentaFlow target not found in Prometheus targets after %v. Last response: %s", healthTimeout, targets)
	}
}

=======
>>>>>>> Stashed changes
=======
>>>>>>> Stashed changes
// waitForAgentaFlowTarget polls the Prometheus targets endpoint until AgentaFlow appears or timeout is reached.
func waitForAgentaFlowTarget(url string, timeout, retryWait time.Duration) (bool, string) {
	deadline := time.Now().Add(timeout)
	var lastBody string
	for time.Now().Before(deadline) {
		resp, err := http.Get(url)
		if err != nil {
			lastBody = fmt.Sprintf("Error: %v", err)
			time.Sleep(retryWait)
			continue
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastBody = fmt.Sprintf("Error reading body: %v", err)
			time.Sleep(retryWait)
			continue
		}
		lastBody = string(body)
		if resp.StatusCode == http.StatusOK &&
			(strings.Contains(lastBody, "agentaflow") || strings.Contains(lastBody, "9001")) {
			return true, lastBody
		}
		time.Sleep(retryWait)
<<<<<<< Updated upstream
<<<<<<< Updated upstream
=======
=======
>>>>>>> Stashed changes
	}
	return false, lastBody

	targets := string(body)

	// Check that AgentaFlow is being scraped
	if !strings.Contains(targets, "agentaflow") && !strings.Contains(targets, "9001") {
		t.Errorf("AgentaFlow target not found in Prometheus targets. Response: %s", targets)
>>>>>>> Stashed changes
	}
	return false, lastBody
}

// TestWebSocketConnection verifies WebSocket endpoint is accessible
func TestWebSocketConnection(t *testing.T) {
	// Just verify the endpoint exists (actual WebSocket upgrade would require ws:// client)
	resp, err := http.Get(dashboardURL + "/ws")
	if err != nil {
		t.Fatalf("Failed to reach WebSocket endpoint: %v", err)
	}
	defer resp.Body.Close()

	// WebSocket upgrade should fail with regular HTTP, but endpoint should be reachable
	// We expect either 400 (bad request) or similar, not 404
	if resp.StatusCode == http.StatusNotFound {
		t.Errorf("WebSocket endpoint returned 404, it should exist")
	}
}

// TestConcurrentRequests verifies dashboard handles concurrent load
func TestConcurrentRequests(t *testing.T) {
	const numRequests = 50
	errors := make(chan error, numRequests)

	for i := 0; i < numRequests; i++ {
		go func() {
			resp, err := http.Get(dashboardURL + "/health")
			if err != nil {
				errors <- err
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				errors <- fmt.Errorf("got status %d", resp.StatusCode)
				return
			}
			errors <- nil
		}()
	}

	// Collect results
	failCount := 0
	for i := 0; i < numRequests; i++ {
		if err := <-errors; err != nil {
			t.Logf("Request failed: %v", err)
			failCount++
		}
	}

	if failCount > numRequests/10 { // Allow 10% failure rate
		t.Errorf("Too many concurrent requests failed: %d/%d", failCount, numRequests)
	}
}

// waitForEndpoint retries HTTP GET until success or timeout
func waitForEndpoint(url string, timeout time.Duration) (*http.Response, error) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		resp, err := client.Get(url)
		if err == nil && resp.StatusCode == http.StatusOK {
			return resp, nil
		}

		if resp != nil {
			resp.Body.Close()
		}

		time.Sleep(healthRetryWait)
	}

	return nil, fmt.Errorf("endpoint %s not ready after %v", url, timeout)
}
