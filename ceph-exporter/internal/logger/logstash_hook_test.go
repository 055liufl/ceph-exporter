// =============================================================================
// Logstash Hook 单元测试
// =============================================================================
// 测试 Logstash Hook 的功能，包括:
//   - Hook 创建（有效/无效的协议和地址）
//   - 日志发送（通过模拟 TCP 服务器验证）
//   - 日志级别过滤（应监听所有级别）
//   - 关闭和资源释放（包括幂等性测试）
//
// 测试方法:
//
//	使用 net.Listen 创建本地 TCP 服务器模拟 Logstash，
//	验证 Hook 能正确建立连接并发送 JSON 格式的日志数据。
//
// =============================================================================
package logger

import (
	"net"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

// TestNewLogstashHook 测试创建 Logstash Hook
// 使用 table-driven tests 测试多种场景:
//   - valid tcp: 使用有效的 TCP 地址创建 Hook（应成功）
//   - invalid protocol: 使用不支持的协议（http）创建 Hook（应失败）
//   - invalid address: 使用无效的地址创建 Hook（应失败）
func TestNewLogstashHook(t *testing.T) {
	// 启动一个模拟的 TCP 服务器用于测试
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to start mock server: %v", err)
	}
	defer listener.Close()
	validAddress := listener.Addr().String()

	tests := []struct {
		name        string
		protocol    string
		address     string
		serviceName string
		wantErr     bool
	}{
		{
			name:        "valid tcp",
			protocol:    "tcp",
			address:     validAddress,
			serviceName: "test-service",
			wantErr:     false,
		},
		{
			name:        "invalid protocol",
			protocol:    "http",
			address:     "localhost:5044",
			serviceName: "test-service",
			wantErr:     true,
		},
		{
			name:        "invalid address",
			protocol:    "tcp",
			address:     "invalid:99999",
			serviceName: "test-service",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hook, err := NewLogstashHook(tt.protocol, tt.address, tt.serviceName)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewLogstashHook() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if hook != nil {
				hook.Close()
			}
		})
	}
}

// TestLogstashHook_Fire 测试日志发送功能
// 流程:
//  1. 启动模拟 TCP 服务器
//  2. 创建 Hook 并连接到模拟服务器
//  3. 构造带有 component 和 trace_id 字段的日志条目
//  4. 调用 Fire 发送日志
//  5. 验证模拟服务器收到了日志数据
func TestLogstashHook_Fire(t *testing.T) {
	// 启动一个模拟的 TCP 服务器
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to start mock server: %v", err)
	}
	defer listener.Close()

	address := listener.Addr().String()

	// 接收日志的 goroutine
	received := make(chan string, 10)
	go func() {
		conn, connErr := listener.Accept()
		if connErr != nil {
			return
		}
		defer conn.Close()

		buf := make([]byte, 4096)
		for {
			n, readErr := conn.Read(buf)
			if readErr != nil {
				return
			}
			received <- string(buf[:n])
		}
	}()

	// 创建 Hook
	hook, err := NewLogstashHook("tcp", address, "test-service")
	if err != nil {
		t.Fatalf("Failed to create hook: %v", err)
	}
	defer hook.Close()

	// 创建日志条目
	entry := &logrus.Entry{
		Logger:  logrus.New(),
		Time:    time.Now(),
		Level:   logrus.InfoLevel,
		Message: "test message",
		Data: logrus.Fields{
			"component": "test",
			"trace_id":  "12345",
		},
	}

	// 发送日志
	if err := hook.Fire(entry); err != nil {
		t.Errorf("Fire() error = %v", err)
	}

	// 等待接收
	select {
	case data := <-received:
		if len(data) == 0 {
			t.Error("Received empty data")
		}
		t.Logf("Received: %s", data)
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for log data")
	}
}

// TestLogstashHook_Levels 测试 Hook 监听的日志级别
// 验证 Hook 监听所有日志级别（从 Panic 到 Trace）
func TestLogstashHook_Levels(t *testing.T) {
	hook := &LogstashHook{}
	levels := hook.Levels()

	if len(levels) != len(logrus.AllLevels) {
		t.Errorf("Expected %d levels, got %d", len(logrus.AllLevels), len(levels))
	}
}

// TestLogstashHook_Close 测试关闭 Hook
// 验证:
//   - 正常关闭不报错
//   - 重复关闭不报错（幂等性）
//   - 关闭后 sender goroutine 正确退出
func TestLogstashHook_Close(t *testing.T) {
	hook := &LogstashHook{
		buffer: make(chan []byte, 10),
	}

	// 启动 sender goroutine
	hook.wg.Add(1)
	go hook.sender()

	// 关闭
	if err := hook.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}

	// 再次关闭应该不报错
	if err := hook.Close(); err != nil {
		t.Errorf("Second Close() error = %v", err)
	}
}
