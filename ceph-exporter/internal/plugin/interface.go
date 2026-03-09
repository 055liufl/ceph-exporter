// =============================================================================
// 插件接口定义
// =============================================================================
// 本文件定义了 ceph-exporter 的插件系统接口。
// 插件系统允许第三方开发者扩展 ceph-exporter 的功能，支持:
//   - 自定义 Prometheus 采集器
//   - 第三方存储系统监控
//   - 通用功能扩展
//
// 插件类型:
//  1. CollectorPlugin: 采集器插件，用于扩展 Prometheus 指标采集
//  2. StoragePlugin: 存储插件，用于监控第三方存储系统
//  3. Plugin: 通用插件接口，所有插件的基础
//
// 插件加载方式:
//   - .so 动态库: 通过 Go plugin 包加载（需要 CGO）
//   - HTTP 远程调用: 通过 HTTP API 与远程插件通信
//
// 使用示例:
//
//	// 实现一个采集器插件
//	type MyCollectorPlugin struct {
//	    collector prometheus.Collector
//	}
//
//	func (p *MyCollectorPlugin) Name() string { return "my-collector" }
//	func (p *MyCollectorPlugin) Init(config map[string]interface{}) error { ... }
//	func (p *MyCollectorPlugin) Collector() prometheus.Collector { return p.collector }
//
// =============================================================================
package plugin

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
)

// Plugin 定义了插件的基本接口
// 所有插件必须实现此接口才能被插件管理器加载和管理。
// 这是所有插件类型的基础接口，定义了插件的生命周期方法。
type Plugin interface {
	// Name 返回插件名称
	// 名称必须唯一，用于标识插件和日志记录。
	// 建议使用小写字母和连字符，如 "my-storage-plugin"。
	// 返回:
	//   - string: 插件名称
	Name() string

	// Version 返回插件版本
	// 遵循语义化版本规范（Semantic Versioning），如 "v1.0.0"。
	// 版本号用于兼容性检查和问题排查。
	// 返回:
	//   - string: 插件版本号
	Version() string

	// Description 返回插件描述
	// 简要说明插件的功能和用途，用于文档和日志。
	// 返回:
	//   - string: 插件描述信息
	Description() string

	// Init 初始化插件
	// 在插件加载后、启动前调用，用于读取配置、初始化资源等。
	// 参数:
	//   - config: 插件配置参数（从配置文件的 plugins[].config 加载）
	// 返回:
	//   - error: 初始化失败时返回错误，插件将不会被启动
	Init(config map[string]interface{}) error

	// Start 启动插件
	// 在初始化成功后调用，开始执行插件的主要功能。
	// 对于长期运行的插件，应该在 goroutine 中执行，并监听 ctx 的取消信号。
	// 参数:
	//   - ctx: 上下文，用于控制插件生命周期。当 ctx 被取消时，插件应该停止运行。
	// 返回:
	//   - error: 启动失败时返回错误
	Start(ctx context.Context) error

	// Stop 停止插件
	// 在程序退出或插件被卸载时调用，执行清理工作、释放资源。
	// 应该确保所有 goroutine 都已停止，所有资源都已释放。
	// 返回:
	//   - error: 停止失败时返回错误（通常记录日志但不影响程序退出）
	Stop() error

	// Health 健康检查
	// 定期被调用以检查插件是否正常运行。
	// 可以检查连接状态、资源可用性等。
	// 返回:
	//   - error: nil 表示插件健康，否则返回错误信息描述问题
	Health() error
}

// CollectorPlugin 定义了采集器插件接口
// 用于扩展 Prometheus 指标采集功能，允许第三方开发者添加自定义指标。
// 实现此接口的插件会被自动注册到 Prometheus Registry。
//
// 使用场景:
//   - 监控自定义应用或服务
//   - 采集 Ceph 之外的存储系统指标
//   - 实现特定业务指标
type CollectorPlugin interface {
	Plugin // 继承基础插件接口

	// Collector 返回 Prometheus 采集器
	// 返回的采集器将被注册到 Prometheus Registry，
	// Prometheus 会定期调用其 Collect() 方法采集指标。
	// 返回:
	//   - prometheus.Collector: Prometheus 采集器实例
	Collector() prometheus.Collector
}

// StoragePlugin 定义了存储插件接口
// 用于支持第三方存储系统的监控，如 MinIO、GlusterFS、Lustre 等。
// 与 CollectorPlugin 不同，StoragePlugin 使用自定义的 Metric 结构，
// 由插件管理器负责转换为 Prometheus 指标。
//
// 使用场景:
//   - 监控对象存储系统（MinIO、S3）
//   - 监控分布式文件系统（GlusterFS、Lustre）
//   - 监控块存储系统（iSCSI、FC）
type StoragePlugin interface {
	Plugin // 继承基础插件接口

	// Collect 采集存储指标
	// 由插件管理器定期调用，采集存储系统的指标数据。
	// 参数:
	//   - ctx: 上下文，用于超时控制。插件应该在 ctx 超时前返回。
	// 返回:
	//   - []Metric: 采集到的指标列表
	//   - error: 采集失败时返回错误
	Collect(ctx context.Context) ([]Metric, error)
}

// Metric 定义了插件返回的指标结构
// 用于 StoragePlugin 返回指标数据，由插件管理器转换为 Prometheus 指标。
type Metric struct {
	// Name 指标名称（不含前缀）
	// 插件管理器会自动添加 "ceph_plugin_<plugin_name>_" 前缀
	// 例如: "storage_used_bytes" -> "ceph_plugin_minio_storage_used_bytes"
	Name string

	// Help 指标帮助信息
	// 描述指标的含义和用途，会显示在 Prometheus 的 /metrics 端点
	Help string

	// Type 指标类型
	// 支持的类型: "gauge", "counter", "histogram", "summary"
	// 大多数存储指标使用 "gauge" 类型
	Type string

	// Value 指标值
	// 必须是有效的浮点数，NaN 和 Inf 会被拒绝
	Value float64

	// Labels 指标标签
	// 用于区分同一指标的不同维度，如 {"pool": "data", "type": "replicated"}
	// 标签名应该使用小写字母和下划线
	Labels map[string]string
}

// PluginType 定义插件类型
// 用于标识插件的功能类别
type PluginType string

const (
	// PluginTypeCollector 采集器插件
	// 实现 CollectorPlugin 接口，提供 Prometheus 采集器
	PluginTypeCollector PluginType = "collector"

	// PluginTypeStorage 存储插件
	// 实现 StoragePlugin 接口，监控第三方存储系统
	PluginTypeStorage PluginType = "storage"

	// PluginTypeGeneric 通用插件
	// 只实现 Plugin 基础接口，用于通用功能扩展
	PluginTypeGeneric PluginType = "generic"
)

// PluginInfo 插件信息
// 用于注册和管理插件的元数据
type PluginInfo struct {
	// Name 插件名称
	// 必须唯一，用于标识插件
	Name string

	// Version 插件版本
	// 遵循语义化版本规范
	Version string

	// Description 插件描述
	// 简要说明插件功能
	Description string

	// Type 插件类型
	// 决定插件的加载和管理方式
	Type PluginType

	// Enabled 是否启用
	// false 表示插件已注册但未启动
	Enabled bool

	// Config 插件配置
	// 从配置文件加载的插件特定配置
	Config map[string]interface{}
}
