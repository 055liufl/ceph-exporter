// =============================================================================
// Docker Compose 集成测试
// =============================================================================
// 测试 Docker Compose 环境的启动和容器状态。
// 使用 docker-compose-integration-test.yml 配置文件，包含:
//   - ceph-demo: Ceph 演示集群（单节点）
//   - ceph-exporter: Ceph 指标导出器
//   - prometheus: 指标采集和存储
//   - grafana: 可视化面板
//
// 注意:
//   - 此测试需要 Docker 和 docker-compose 环境
//   - Ceph Demo 初始化需要约 90 秒
//   - 使用 -short 标志可跳过此测试
//   - 测试完成后会自动清理环境（docker-compose down）
//
// =============================================================================
package integration

import (
	"context"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// TestDockerComposeUp 测试所有容器能否成功启动
// 流程:
//  1. 使用 docker-compose up -d 启动所有服务
//  2. 等待 90 秒让 Ceph Demo 完成初始化
//  3. 使用 docker-compose ps 验证所有容器都在运行
//  4. 测试结束后自动执行 docker-compose down 清理环境
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
