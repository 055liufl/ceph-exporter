//go:build !cgo
// +build !cgo

// =============================================================================
// RADOS 连接 - 无 CGO stub 实现
// =============================================================================
// 当 CGO_ENABLED=0 时使用此文件。
// 提供编译通过的 stub，实际调用会返回错误。
// 用于 Windows 开发环境和纯 Go 测试场景。
// =============================================================================
package ceph

import "fmt"

// radosConn 是 RADOS 连接的接口抽象
type radosConn interface {
	ReadConfigFile(path string) error
	SetConfigOption(option, value string) error
	Connect() error
	Shutdown()
	MonCommand(args []byte) ([]byte, string, error)
}

// stubConn 无 CGO 环境下的 stub 连接
type stubConn struct{}

func (s *stubConn) ReadConfigFile(path string) error {
	return fmt.Errorf("CGO 未启用，无法读取 Ceph 配置")
}
func (s *stubConn) SetConfigOption(option, value string) error {
	return fmt.Errorf("CGO 未启用，无法设置配置选项")
}
func (s *stubConn) Connect() error { return fmt.Errorf("CGO 未启用，无法连接 Ceph 集群") }
func (s *stubConn) Shutdown()      {}
func (s *stubConn) MonCommand(args []byte) ([]byte, string, error) {
	return nil, "", fmt.Errorf("CGO 未启用，无法执行 Ceph 命令")
}

// newRadosConn 创建 stub 连接（无 CGO 模式）
func newRadosConn(cluster, user string) (radosConn, error) {
	return &stubConn{}, nil
}
