// =============================================================================
// 服务可用性简单测试
// =============================================================================
// 测试各个服务的 HTTP 端点是否可访问。
// 这是一个轻量级测试，不需要 docker-compose 启动环境，
// 适用于验证已经运行的服务是否正常。
//
// 测试的服务:
//   - ceph-exporter: 端口 9128，/health 端点
//   - Prometheus: 端口 9090，/-/healthy 端点
//   - Grafana: 端口 3000，/api/health 端点
//   - Alertmanager: 端口 9093，/-/healthy 端点
//
// 注意:
//
//	如果服务未运行，测试会被跳过（Skip）而不是失败
//
// =============================================================================
package integration

import (
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

// TestServicesRunning 测试各服务是否可访问
// 使用 table-driven tests 遍历所有服务端点，
// 验证 HTTP 响应中包含预期的字符串
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
