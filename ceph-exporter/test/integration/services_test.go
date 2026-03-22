// =============================================================================
// 服务功能集成测试
// =============================================================================
// 测试部署环境中各服务的完整功能，包括:
//   - 数据持久化: Docker 数据卷、Prometheus 数据查询、Grafana 数据源
//   - Prometheus 目标发现: 验证 ceph-exporter 被正确发现
//   - Grafana Dashboard: 验证预配置的 Dashboard 是否加载
//   - 容器健康检查: 验证所有容器的健康状态
//   - Web UI 可访问性: 验证 Prometheus、Grafana、Alertmanager 的 Web 界面
//   - 指标采集: 验证 ceph-exporter 指标格式和 Prometheus 采集
//
// 注意:
//   - 所有测试都需要 Docker 环境已启动
//   - 使用 -short 标志可跳过这些测试
//   - 部分测试需要等待服务初始化（10-30 秒）
//
// =============================================================================
package integration

import (
	"io"
	"net/http"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// TestDataPersistence 测试数据持久化
func TestDataPersistence(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	// 等待服务启动
	time.Sleep(10 * time.Second)

	t.Run("验证数据卷存在", func(t *testing.T) {
		cmd := exec.Command("docker", "volume", "ls")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("获取数据卷列表失败: %v", err)
		}

		// 使用部分匹配，因为数据卷名称可能包含项目前缀
		volumes := []string{
			"prometheus", // 匹配 *prometheus* 或 *prometheus-data*
			"grafana",    // 匹配 *grafana* 或 *grafana-data*
		}

		outputStr := string(output)
		for _, vol := range volumes {
			if !strings.Contains(outputStr, vol) {
				t.Errorf("未找到包含 '%s' 的数据卷", vol)
			}
		}
	})

	t.Run("验证Prometheus数据持久化", func(t *testing.T) {
		// 检查 Prometheus 是否能查询到数据
		url := getServiceURL(9090, "/api/v1/query?query=up")
		resp, err := http.Get(url)
		if err != nil {
			t.Fatalf("查询 Prometheus 失败: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Prometheus 查询失败，状态码: %d", resp.StatusCode)
		}

		body, _ := io.ReadAll(resp.Body)
		if !strings.Contains(string(body), "success") {
			t.Errorf("Prometheus 查询响应异常: %s", string(body))
		}
	})

	t.Run("验证Grafana数据持久化", func(t *testing.T) {
		// 检查 Grafana 数据源配置
		client := &http.Client{}
		url := getServiceURL(3000, "/api/datasources")
		req, _ := http.NewRequest("GET", url, nil)
		req.SetBasicAuth("admin", "admin")

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("查询 Grafana 数据源失败: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Grafana 数据源查询失败，状态码: %d", resp.StatusCode)
		}
	})
}

// TestPrometheusTargets 测试 Prometheus 能否采集 ceph-exporter 指标
func TestPrometheusTargets(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	// 等待 Prometheus 完成服务发现
	time.Sleep(15 * time.Second)

	url := getServiceURL(9090, "/api/v1/targets")
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("查询 Prometheus targets 失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("读取响应失败: %v", err)
	}

	bodyStr := string(body)
	if !strings.Contains(bodyStr, "ceph-exporter") {
		t.Errorf("Prometheus 未发现 ceph-exporter target")
	}

	if !strings.Contains(bodyStr, "9128") {
		t.Errorf("Prometheus target 端口配置错误")
	}

	t.Logf("Prometheus targets 响应: %s", bodyStr)
}

// TestGrafanaDashboards 测试 Grafana Dashboard 是否正确加载
func TestGrafanaDashboards(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	// 等待 Grafana 完成初始化
	time.Sleep(20 * time.Second)

	client := &http.Client{}
	url := getServiceURL(3000, "/api/search?type=dash-db")
	req, _ := http.NewRequest("GET", url, nil)
	req.SetBasicAuth("admin", "admin")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("查询 Grafana dashboards 失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Grafana dashboards 查询失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	// 检查是否有 dashboard（如果配置了的话）
	if bodyStr == "[]" || bodyStr == "null" {
		t.Log("警告: 未找到预配置的 Grafana dashboards")
	} else {
		t.Logf("找到 Grafana dashboards: %s", bodyStr)
	}
}

// TestContainerHealthChecks 测试容器健康检查
func TestContainerHealthChecks(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	// 等待健康检查生效
	time.Sleep(30 * time.Second)

	cmd := exec.Command("docker-compose", "-f", "../../deployments/docker-compose.yml", "ps")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("获取容器状态失败: %v", err)
	}

	outputStr := string(output)
	t.Logf("容器状态:\n%s", outputStr)

	// 检查是否有 unhealthy 的容器
	if strings.Contains(strings.ToLower(outputStr), "unhealthy") {
		t.Error("发现不健康的容器")
	}

	// 使用 docker inspect 检查详细健康状态
	services := []string{"ceph-exporter", "prometheus", "grafana", "alertmanager"}
	for _, service := range services {
		cmd := exec.Command("docker", "inspect", "--format", "{{.State.Health.Status}}", service)
		output, err := cmd.CombinedOutput()
		if err != nil {
			// 某些容器可能没有配置健康检查
			t.Logf("警告: 无法获取 %s 的健康状态: %v", service, err)
			continue
		}

		status := strings.TrimSpace(string(output))
		if status == "unhealthy" {
			t.Errorf("容器 %s 状态不健康: %s", service, status)
		} else {
			t.Logf("容器 %s 健康状态: %s", service, status)
		}
	}
}

// TestWebUIAccess 测试所有 Web UI 可访问性
func TestWebUIAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	// 等待服务启动
	time.Sleep(10 * time.Second)

	tests := []struct {
		name string
		port int
		path string
	}{
		{"Prometheus UI", 9090, "/graph"},
		{"Grafana UI", 3000, "/login"},
		{"Alertmanager UI", 9093, "/#/alerts"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := getServiceURL(tt.port, tt.path)
			client := &http.Client{
				Timeout: 10 * time.Second,
				CheckRedirect: func(req *http.Request, via []*http.Request) error {
					// 允许重定向
					return nil
				},
			}

			resp, err := client.Get(url)
			if err != nil {
				t.Fatalf("访问 %s 失败: %v", tt.name, err)
			}
			defer resp.Body.Close()

			if resp.StatusCode >= 500 {
				t.Errorf("%s 返回服务器错误: %d", tt.name, resp.StatusCode)
			}

			t.Logf("%s 可访问，状态码: %d", tt.name, resp.StatusCode)
		})
	}
}

// TestMetricsCollection 测试指标采集功能
func TestMetricsCollection(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	// 等待指标采集
	time.Sleep(20 * time.Second)

	t.Run("验证ceph-exporter指标格式", func(t *testing.T) {
		url := getServiceURL(9128, "/metrics")
		resp, err := http.Get(url)
		if err != nil {
			t.Fatalf("获取指标失败: %v", err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		bodyStr := string(body)

		// 检查 Prometheus 指标格式
		if !strings.Contains(bodyStr, "# HELP") {
			t.Error("指标缺少 HELP 注释")
		}

		if !strings.Contains(bodyStr, "# TYPE") {
			t.Error("指标缺少 TYPE 注释")
		}

		// 检查是否有实际的指标数据
		lines := strings.Split(bodyStr, "\n")
		hasMetrics := false
		for _, line := range lines {
			if len(line) > 0 && !strings.HasPrefix(line, "#") {
				hasMetrics = true
				break
			}
		}

		if !hasMetrics {
			t.Error("未找到实际的指标数据")
		}

		t.Logf("指标采集正常，共 %d 行", len(lines))
	})

	t.Run("验证Prometheus采集到指标", func(t *testing.T) {
		// 查询 Prometheus 是否采集到 ceph-exporter 的指标
		url := getServiceURL(9090, "/api/v1/query?query=up{job=\"ceph-exporter\"}")
		resp, err := http.Get(url)
		if err != nil {
			t.Fatalf("查询 Prometheus 失败: %v", err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		bodyStr := string(body)

		if !strings.Contains(bodyStr, "success") {
			t.Errorf("Prometheus 查询失败: %s", bodyStr)
		}

		if strings.Contains(bodyStr, "\"result\":[]") {
			t.Error("Prometheus 未采集到 ceph-exporter 指标")
		}

		t.Logf("Prometheus 查询结果: %s", bodyStr)
	})
}
