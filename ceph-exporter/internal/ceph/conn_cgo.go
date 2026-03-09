//go:build cgo
// +build cgo

// =============================================================================
// RADOS 连接 - CGO 实现（使用 go-ceph 库）
// =============================================================================
// 本文件实现了与 Ceph RADOS 集群的底层连接逻辑。
// 使用 go-ceph 库（github.com/ceph/go-ceph/rados）通过 CGO 调用 Ceph C 库。
//
// 构建标签说明:
//   - //go:build cgo: 只有在启用 CGO 时才编译此文件
//   - CGO_ENABLED=1 必须在编译时设置
//
// RADOS (Reliable Autonomic Distributed Object Store):
//   - Ceph 的底层对象存储系统
//   - 提供了与 Ceph 集群通信的基础接口
//   - 支持执行 Monitor 命令、读写对象等操作
//
// =============================================================================
package ceph

import (
	cephRados "github.com/ceph/go-ceph/rados"
)

// radosConn 是 RADOS 连接的接口抽象
// 定义了与 Ceph 集群交互所需的核心方法
// 使用接口而不是具体类型，便于测试和模拟
type radosConn interface {
	// ReadConfigFile 读取 Ceph 配置文件
	// 参数:
	//   - path: 配置文件路径（通常是 /etc/ceph/ceph.conf）
	// 返回:
	//   - error: 读取失败时返回错误
	ReadConfigFile(path string) error

	// SetConfigOption 设置配置选项
	// 参数:
	//   - option: 配置项名称（如 "keyring"）
	//   - value: 配置项的值
	// 返回:
	//   - error: 设置失败时返回错误
	SetConfigOption(option, value string) error

	// Connect 建立到 Ceph 集群的连接
	// 必须在调用其他方法之前先调用此方法
	// 返回:
	//   - error: 连接失败时返回错误
	Connect() error

	// Shutdown 关闭连接并释放资源
	// 应该在程序退出前调用，确保资源正确释放
	Shutdown()

	// MonCommand 执行 Monitor 命令
	// Monitor 命令用于查询和管理 Ceph 集群状态
	// 参数:
	//   - args: JSON 格式的命令参数（如 {"prefix": "status", "format": "json"}）
	// 返回:
	//   - []byte: 命令执行结果（通常是 JSON 格式）
	//   - string: 命令执行的状态信息
	//   - error: 执行失败时返回错误
	MonCommand(args []byte) ([]byte, string, error)
}

// newRadosConn 创建真实的 RADOS 连接（CGO 模式）
// 使用 go-ceph 库创建到 Ceph 集群的连接对象
// 参数:
//   - cluster: 集群名称（通常是 "ceph"）
//   - user: 认证用户名（通常是 "admin"）
//
// 返回:
//   - radosConn: RADOS 连接接口
//   - error: 创建失败时返回错误
//
// 注意:
//   - 此函数只创建连接对象，不建立实际连接
//   - 需要调用 ReadConfigFile、Connect 等方法完成连接建立
func newRadosConn(cluster, user string) (radosConn, error) {
	// 使用 NewConn() 创建连接对象
	// 注意: 不使用 NewConnWithClusterAndUser()，因为集群名和用户名
	// 会从配置文件中读取，这里传入的参数主要用于日志记录
	conn, err := cephRados.NewConn()
	if err != nil {
		return nil, err
	}
	return conn, nil
}
