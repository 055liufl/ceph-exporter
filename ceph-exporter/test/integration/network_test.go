// =============================================================================
// 容器间网络通信测试
// =============================================================================
// 测试 Docker 容器之间的网络连通性，验证各服务的 HTTP 端点是否可达。
// 包含重试机制（最多 5 次，每次间隔 2 秒），以应对服务启动延迟。
//
// 测试的端点:
//   - ceph-exporter: /health, /ready, /metrics
//   - Prometheus: /-/healthy, /-/ready
//   - Grafana: /api/health
//
// 注意:
//   - 使用 -short 标志可跳过此测试
//   - 测试前会等待 10 秒让服务完全启动
//
// =============================================================================
package integration

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

// TestNetworkConnectivity 测试容器间网络通信
func TestNetworkConnectivity(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	// 等待服务完全启动
	time.Sleep(10 * time.Second)

	tests := []struct {
		name     string
		port     int
		path     string
		expected string
	}{
		{
			name:     "ceph-exporter健康检查",
			port:     9128,
			path:     "/health",
			expected: "ok",
		},
		{
			name:     "ceph-exporter就绪检查",
			port:     9128,
			path:     "/ready",
			expected: "ready",
		},
		{
			name:     "ceph-exporter指标端点",
			port:     9128,
			path:     "/metrics",
			expected: "# HELP",
		},
		{
			name:     "Prometheus健康检查",
			port:     9090,
			path:     "/-/healthy",
			expected: "Prometheus is Healthy",
		},
		{
			name:     "Prometheus就绪检查",
			port:     9090,
			path:     "/-/ready",
			expected: "Prometheus is Ready",
		},
		{
			name:     "Grafana健康检查",
			port:     3000,
			path:     "/api/health",
			expected: "ok",
		},
		// 注意: minimal-test 配置不包含 alertmanager
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := getServiceURL(tt.port, tt.path)

			// 重试机制，最多重试5次
			var lastErr error
			for i := 0; i < 5; i++ {
				resp, err := http.Get(url)
				if err != nil {
					lastErr = fmt.Errorf("请求失败: %v", err)
					time.Sleep(2 * time.Second)
					continue
				}
				defer resp.Body.Close()

				body, err := io.ReadAll(resp.Body)
				if err != nil {
					lastErr = fmt.Errorf("读取响应失败: %v", err)
					time.Sleep(2 * time.Second)
					continue
				}

				if resp.StatusCode != http.StatusOK {
					lastErr = fmt.Errorf("状态码错误: %d, 响应: %s", resp.StatusCode, string(body))
					time.Sleep(2 * time.Second)
					continue
				}

				if !strings.Contains(string(body), tt.expected) {
					lastErr = fmt.Errorf("响应内容不符合预期，期望包含: %s, 实际: %s", tt.expected, string(body))
					time.Sleep(2 * time.Second)
					continue
				}

				// 测试通过
				lastErr = nil
				break
			}

			if lastErr != nil {
				t.Errorf("%s: %v", tt.name, lastErr)
			}
		})
	}
}
