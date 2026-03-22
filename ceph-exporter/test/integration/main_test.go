// =============================================================================
// 集成测试入口和环境管理
// =============================================================================
// 本文件提供集成测试的入口点和环境管理函数。
//
// TestMain 是 Go 测试框架的入口函数，在所有测试之前执行。
// 当前实现中，环境的启动和清理需要手动完成（docker-compose up/down），
// 因为自动化启动在某些平台上存在编码问题。
//
// 辅助函数:
//   - setupTestEnvironment: 启动 docker-compose 环境（清理旧环境 -> 构建并启动 -> 等待就绪）
//   - teardownTestEnvironment: 清理 docker-compose 环境
//
// =============================================================================
package integration

import (
	"context"
	"os"
	"os/exec"
	"testing"
	"time"
)

// TestMain 是集成测试的入口点
// 负责启动和清理测试环境
func TestMain(m *testing.M) {
	// Skip automatic setup for now due to docker-compose encoding issues on Windows
	// Services should be started manually before running tests

	// 运行测试
	code := m.Run()

	os.Exit(code)
}

// setupTestEnvironment 启动 docker-compose 环境
func setupTestEnvironment() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// 先清理可能存在的旧环境
	cmd := exec.Command("docker-compose", "-f", "../../deployments/docker-compose.yml", "down", "-v")
	_ = cmd.Run()

	// 启动服务
	cmd = exec.CommandContext(ctx, "docker-compose", "-f", "../../deployments/docker-compose.yml", "up", "-d", "--build")
	_, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}

	// 等待服务启动
	time.Sleep(45 * time.Second)

	return nil
}

// teardownTestEnvironment 清理 docker-compose 环境
func teardownTestEnvironment() error {
	cmd := exec.Command("docker-compose", "-f", "../../deployments/docker-compose.yml", "down")
	_, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}

	return nil
}
