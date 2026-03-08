package ceph

import (
	"time"

	"ceph-exporter/internal/config"
	"ceph-exporter/internal/logger"
)

// MockConn 用于测试的 mock RADOS 连接
// 通过 MonCommandFunc 控制 MonCommand 的返回值
type MockConn struct {
	MonCommandFunc func(args []byte) ([]byte, string, error)
}

func (m *MockConn) ReadConfigFile(path string) error           { return nil }
func (m *MockConn) SetConfigOption(option, value string) error { return nil }
func (m *MockConn) Connect() error                             { return nil }
func (m *MockConn) Shutdown()                                  {}
func (m *MockConn) MonCommand(args []byte) ([]byte, string, error) {
	if m.MonCommandFunc != nil {
		return m.MonCommandFunc(args)
	}
	return nil, "", nil
}

// NewTestClient 创建用于测试的 Client，使用 mock 连接
func NewTestClient(log *logger.Logger, monCommandFunc func([]byte) ([]byte, string, error)) *Client {
	return &Client{
		conn: &MockConn{MonCommandFunc: monCommandFunc},
		config: &config.CephConfig{
			Timeout: 10 * time.Second,
		},
		log:    log,
		closed: false,
	}
}
