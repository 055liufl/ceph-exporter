package integration

import (
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

// TestServicesRunning tests if services are accessible
// This is a simplified test that doesn't require docker-compose
func TestServicesRunning(t *testing.T) {
	// Get Docker host
	host := getDockerHost()
	t.Logf("Using Docker host: %s", host)

	tests := []struct {
		name     string
		port     int
		path     string
		expected string
	}{
		{
			name:     "ceph-exporter health",
			port:     9128,
			path:     "/health",
			expected: "ok",
		},
		{
			name:     "Prometheus healthy",
			port:     9090,
			path:     "/-/healthy",
			expected: "Prometheus",
		},
		{
			name:     "Grafana health",
			port:     3000,
			path:     "/api/health",
			expected: "ok",
		},
		{
			name:     "Alertmanager healthy",
			port:     9093,
			path:     "/-/healthy",
			expected: "OK",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := getServiceURL(tt.port, tt.path)
			t.Logf("Testing URL: %s", url)

			client := &http.Client{
				Timeout: 10 * time.Second,
			}

			resp, err := client.Get(url)
			if err != nil {
				t.Skipf("Service not accessible (may not be running): %v", err)
				return
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("Failed to read response: %v", err)
			}

			bodyStr := string(body)
			if !strings.Contains(bodyStr, tt.expected) {
				t.Errorf("Response doesn't contain expected string '%s', got: %s", tt.expected, bodyStr)
			} else {
				t.Logf("✓ Service is healthy")
			}
		})
	}
}
