// =============================================================================
// Docker 主机地址辅助工具
// =============================================================================
// 提供获取 Docker 主机地址的辅助函数，兼容多种 Docker 环境:
//   - 原生 Docker（Linux/Mac）: 返回 localhost
//   - Docker Toolbox（旧版 Windows/Mac）: 返回 docker-machine IP
//   - 自定义环境: 通过 DOCKER_HOST_IP 环境变量指定
//
// 优先级:
//  1. DOCKER_HOST_IP 环境变量
//  2. docker-machine ip 命令输出
//  3. 默认 localhost
//
// =============================================================================
package integration

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// getDockerHost 获取 Docker 主机地址
// 在 Docker Toolbox 环境下返回 docker-machine IP
// 在原生 Docker 环境下返回 localhost
func getDockerHost() string {
	// 检查是否设置了环境变量
	if host := os.Getenv("DOCKER_HOST_IP"); host != "" {
		return host
	}

	// 尝试获取 docker-machine IP
	cmd := exec.Command("docker-machine", "ip", "default")
	output, err := cmd.Output()
	if err == nil {
		ip := strings.TrimSpace(string(output))
		if ip != "" {
			return ip
		}
	}

	// 默认返回 localhost（原生 Docker）
	return "localhost"
}

// getServiceURL 构建服务 URL
func getServiceURL(port int, path string) string {
	host := getDockerHost()
	return fmt.Sprintf("http://%s:%d%s", host, port, path)
}
