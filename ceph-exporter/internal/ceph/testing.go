// =============================================================================
// Ceph 客户端测试辅助工具
// =============================================================================
// 本文件提供用于单元测试的 Mock 对象和辅助函数。
// 通过 Mock RADOS 连接，可以在不依赖真实 Ceph 集群的情况下测试客户端逻辑。
//
// 主要组件:
//   - MockConn: Mock RADOS 连接，实现 radosConn 接口
//   - NewTestClient: 创建用于测试的 Ceph 客户端实例
//
// 使用示例:
//
//	// 创建返回固定 JSON 的 mock 函数
//	mockFunc := func(args []byte) ([]byte, string, error) {
//	    return []byte(`{"health":{"status":"HEALTH_OK"}}`), "", nil
//	}
//
//	// 创建测试客户端
//	client := NewTestClient(log, mockFunc)
//
//	// 执行测试
//	status, err := client.GetClusterStatus(ctx)
//
// =============================================================================
package ceph

import (
	"time"

	"ceph-exporter/internal/config"
	"ceph-exporter/internal/logger"
)

// MockConn Mock RADOS 连接
// 实现 radosConn 接口，用于单元测试中替代真实的 Ceph 连接。
// 通过设置 MonCommandFunc 可以控制 MonCommand 方法的返回值，模拟各种场景。
//
// 特性:
//   - 所有配置方法（ReadConfigFile, SetConfigOption）都是空操作
//   - Connect 和 Shutdown 方法都是空操作
//   - MonCommand 方法通过 MonCommandFunc 委托给测试代码控制
//
// 使用场景:
//   - 测试正常的 Ceph 命令响应
//   - 测试错误处理逻辑（返回错误）
//   - 测试超时场景（在 MonCommandFunc 中 sleep）
//   - 测试各种 JSON 响应格式
type MockConn struct {
	// MonCommandFunc 自定义的命令执行函数
	// 参数: args - 命令参数（JSON 格式）
	// 返回: (响应数据, 状态字符串, 错误)
	// 如果为 nil，MonCommand 将返回空响应
	MonCommandFunc func(args []byte) ([]byte, string, error)
}

// ReadConfigFile Mock 实现 - 读取配置文件（空操作）
// 参数:
//   - path: 配置文件路径（被忽略）
//
// 返回:
//   - error: 始终返回 nil
func (m *MockConn) ReadConfigFile(path string) error { return nil }

// SetConfigOption Mock 实现 - 设置配置选项（空操作）
// 参数:
//   - option: 配置选项名称（被忽略）
//   - value: 配置选项值（被忽略）
//
// 返回:
//   - error: 始终返回 nil
func (m *MockConn) SetConfigOption(option, value string) error { return nil }

// Connect Mock 实现 - 连接到 Ceph 集群（空操作）
// 返回:
//   - error: 始终返回 nil，表示连接成功
func (m *MockConn) Connect() error { return nil }

// Shutdown Mock 实现 - 关闭连接（空操作）
func (m *MockConn) Shutdown() {}

// MonCommand Mock 实现 - 执行 Monitor 命令
// 如果设置了 MonCommandFunc，则委托给该函数处理；
// 否则返回空响应。
//
// 参数:
//   - args: 命令参数（JSON 格式）
//
// 返回:
//   - []byte: 命令响应数据
//   - string: 状态字符串
//   - error: 执行错误
func (m *MockConn) MonCommand(args []byte) ([]byte, string, error) {
	if m.MonCommandFunc != nil {
		return m.MonCommandFunc(args)
	}
	return nil, "", nil
}

// NewTestClient 创建用于测试的 Ceph 客户端
// 返回一个使用 MockConn 的 Client 实例，可以在单元测试中使用。
//
// 参数:
//   - log: 日志实例（可以使用 logger.NewLogger 创建测试日志）
//   - monCommandFunc: Mock MonCommand 函数，用于控制命令响应
//
// 返回:
//   - *Client: 配置好的测试客户端实例
//
// 使用示例:
//
//	// 模拟返回集群状态
//	mockFunc := func(args []byte) ([]byte, string, error) {
//	    statusJSON := `{
//	        "health": {"status": "HEALTH_OK"},
//	        "pgmap": {"num_pgs": 128, "bytes_total": 1000000000}
//	    }`
//	    return []byte(statusJSON), "", nil
//	}
//	client := NewTestClient(log, mockFunc)
//
//	// 模拟返回错误
//	mockFunc := func(args []byte) ([]byte, string, error) {
//	    return nil, "", fmt.Errorf("connection timeout")
//	}
//	client := NewTestClient(log, mockFunc)
func NewTestClient(log *logger.Logger, monCommandFunc func([]byte) ([]byte, string, error)) *Client {
	return &Client{
		conn: &MockConn{MonCommandFunc: monCommandFunc},
		config: &config.CephConfig{
			Timeout: 10 * time.Second, // 默认超时 10 秒
		},
		log:    log,
		closed: false,
	}
}
