package plugin

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
)

// Plugin 定义了插件的基本接口
// 所有插件必须实现此接口才能被插件管理器加载和管理
type Plugin interface {
	// Name 返回插件名称
	// 名称必须唯一，用于标识插件
	Name() string

	// Version 返回插件版本
	// 遵循语义化版本规范（如 v1.0.0）
	Version() string

	// Description 返回插件描述
	// 简要说明插件的功能和用途
	Description() string

	// Init 初始化插件
	// config: 插件配置参数（从配置文件加载）
	// 返回错误表示初始化失败
	Init(config map[string]interface{}) error

	// Start 启动插件
	// ctx: 上下文，用于控制插件生命周期
	// 返回错误表示启动失败
	Start(ctx context.Context) error

	// Stop 停止插件
	// 执行清理工作，释放资源
	// 返回错误表示停止失败
	Stop() error

	// Health 健康检查
	// 返回 nil 表示插件健康，否则返回错误信息
	Health() error
}

// CollectorPlugin 定义了采集器插件接口
// 用于扩展 Prometheus 指标采集功能
type CollectorPlugin interface {
	Plugin

	// Collector 返回 Prometheus 采集器
	// 返回的采集器将被注册到 Prometheus Registry
	Collector() prometheus.Collector
}

// StoragePlugin 定义了存储插件接口
// 用于支持第三方存储系统的监控
type StoragePlugin interface {
	Plugin

	// Collect 采集存储指标
	// ctx: 上下文，用于超时控制
	// 返回指标列表和可能的错误
	Collect(ctx context.Context) ([]Metric, error)
}

// Metric 定义了插件返回的指标结构
type Metric struct {
	// Name 指标名称（不含前缀）
	Name string

	// Help 指标帮助信息
	Help string

	// Type 指标类型（gauge, counter, histogram, summary）
	Type string

	// Value 指标值
	Value float64

	// Labels 指标标签
	Labels map[string]string
}

// PluginType 定义插件类型
type PluginType string

const (
	// PluginTypeCollector 采集器插件
	PluginTypeCollector PluginType = "collector"

	// PluginTypeStorage 存储插件
	PluginTypeStorage PluginType = "storage"

	// PluginTypeGeneric 通用插件
	PluginTypeGeneric PluginType = "generic"
)

// PluginInfo 插件信息
type PluginInfo struct {
	// Name 插件名称
	Name string

	// Version 插件版本
	Version string

	// Description 插件描述
	Description string

	// Type 插件类型
	Type PluginType

	// Enabled 是否启用
	Enabled bool

	// Config 插件配置
	Config map[string]interface{}
}
