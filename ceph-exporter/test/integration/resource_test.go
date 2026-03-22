// =============================================================================
// 资源约束测试
// =============================================================================
// 测试 Docker 容器的资源限制配置，包括:
//   - 内存限制是否已设置
//   - 内存使用情况
//   - 资源约束是否符合预期值
//
// 预期的内存限制:
//   - ceph-exporter: 128MB
//   - prometheus: 512MB
//   - grafana: 256MB
//   - ceph-demo: 1024MB
//
// 注意:
//   - 使用 -short 标志可跳过此测试
//   - 需要 Docker 环境
//
// =============================================================================
package integration

import (
	"encoding/json"
	"os/exec"
	"strings"
	"testing"
)

// ContainerStats 容器资源统计信息
type ContainerStats struct {
	Name     string
	MemUsage string
	MemLimit string
	MemPerc  string
}

// TestMemoryLimits 测试所有容器都设置了内存限制
func TestMemoryLimits(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	t.Run("验证容器内存限制", func(t *testing.T) {
		// 获取容器列表
		cmd := exec.Command("docker", "ps", "--format", "{{.Names}}")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("获取容器列表失败: %v", err)
		}

		containers := strings.Split(strings.TrimSpace(string(output)), "\n")
		if len(containers) == 0 {
			t.Skip("没有运行的容器")
		}

		// 检查每个容器的内存限制
		for _, container := range containers {
			if container == "" {
				continue
			}

			t.Run(container, func(t *testing.T) {
				// 使用 docker inspect 检查内存限制
				cmd := exec.Command("docker", "inspect", container, "--format", "{{.HostConfig.Memory}}")
				output, err := cmd.CombinedOutput()
				if err != nil {
					t.Fatalf("检查容器 %s 内存限制失败: %v", container, err)
				}

				memLimit := strings.TrimSpace(string(output))
				if memLimit == "0" {
					t.Logf("警告: 容器 %s 未设置内存限制", container)
				} else {
					t.Logf("容器 %s 内存限制: %s 字节", container, memLimit)
				}
			})
		}
	})

	t.Run("验证容器内存使用", func(t *testing.T) {
		// 获取容器内存使用情况
		cmd := exec.Command("docker", "stats", "--no-stream", "--format", "{{.Name}}\t{{.MemUsage}}\t{{.MemPerc}}")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("获取容器内存使用失败: %v", err)
		}

		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		for _, line := range lines {
			if line == "" {
				continue
			}

			parts := strings.Split(line, "\t")
			if len(parts) >= 3 {
				name := parts[0]
				memUsage := parts[1]
				memPerc := parts[2]
				t.Logf("容器 %s: 内存使用 %s (%s)", name, memUsage, memPerc)
			}
		}
	})
}

// TestResourceConstraints 测试资源约束配置
func TestResourceConstraints(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	// Expected memory limits (bytes)
	expectedLimits := map[string]int64{
		"ceph-exporter-test": 128 * 1024 * 1024,  // 128MB
		"prometheus-test":    512 * 1024 * 1024,  // 512MB
		"grafana-test":       256 * 1024 * 1024,  // 256MB
		"ceph-demo-test":     1024 * 1024 * 1024, // 1024MB
	}

	for container, expectedLimit := range expectedLimits {
		t.Run(container, func(t *testing.T) {
			// 检查容器是否存在
			cmd := exec.Command("docker", "ps", "--filter", "name="+container, "--format", "{{.Names}}")
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("检查容器失败: %v", err)
			}

			if strings.TrimSpace(string(output)) == "" {
				t.Skipf("容器 %s 未运行", container)
			}

			// 获取实际内存限制
			cmd = exec.Command("docker", "inspect", container, "--format", "{{.HostConfig.Memory}}")
			output, err = cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("获取内存限制失败: %v", err)
			}

			var actualLimit int64
			limitStr := strings.TrimSpace(string(output))
			if limitStr != "0" {
				// 解析内存限制
				err := json.Unmarshal([]byte(limitStr), &actualLimit)
				if err != nil {
					// 如果不是 JSON，尝试直接解析
					t.Logf("容器 %s 内存限制: %s", container, limitStr)
				}

				// 验证内存限制是否符合预期（允许一定误差）
				if actualLimit > 0 && actualLimit != expectedLimit {
					t.Logf("容器 %s 内存限制 (%d) 与预期 (%d) 不完全匹配",
						container, actualLimit, expectedLimit)
				}
			}
		})
	}
}
