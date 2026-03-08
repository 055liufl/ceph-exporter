package integration

import (
	"context"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// TestDockerComposeUp tests that all containers can start successfully
func TestDockerComposeUp(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Start all services
	// Use docker-compose-integration-test.yml which includes Ceph Demo
	t.Log("Starting docker-compose services with Ceph Demo...")
	cmd := exec.CommandContext(ctx, "docker-compose", "-f", "../../deployments/docker-compose-integration-test.yml", "up", "-d")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to start services: %v\nOutput: %s", err, output)
	}

	// Wait for services to start (Ceph Demo needs more time)
	t.Log("Waiting for services to start (90 seconds for Ceph Demo initialization)...")
	time.Sleep(90 * time.Second)

	// Verify all containers are running
	t.Run("Verify container status", func(t *testing.T) {
		cmd := exec.Command("docker-compose", "-f", "../../deployments/docker-compose-integration-test.yml", "ps")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to get container status: %v", err)
		}

		outputStr := string(output)
		// Integration test includes: ceph-demo, ceph-exporter, prometheus, grafana
		services := []string{"ceph-demo", "ceph-exporter", "prometheus", "grafana"}
		for _, service := range services {
			if !strings.Contains(outputStr, service) {
				t.Errorf("Service %s is not running", service)
			}
		}
	})

	// Cleanup: stop services after test
	t.Cleanup(func() {
		t.Log("Cleaning up test environment...")
		cmd := exec.Command("docker-compose", "-f", "../../deployments/docker-compose-integration-test.yml", "down")
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Logf("Cleanup failed: %v\nOutput: %s", err, output)
		}
	})
}
