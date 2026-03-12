// =============================================================================
// 配置结构定义
// =============================================================================
// 本文件定义了 ceph-exporter 的所有配置结构体。
// 每个结构体对应配置文件（ceph-exporter.yaml）中的一个配置段。
// 配置通过 YAML 标签与配置文件字段进行映射。
//
// 配置层级:
//
//	Config (根配置)
//	├── ServerConfig     - HTTP 服务器配置
//	├── CephConfig       - Ceph 集群连接配置
//	├── PrometheusConfig - Prometheus 采集配置
//	├── LoggerConfig     - 日志系统配置
//	├── TracerConfig     - 追踪系统配置（Phase 3）
//	└── PluginConfig[]   - 插件配置列表（Phase 5）
//
// =============================================================================
package config

import (
	"time"
)

// Config 主配置结构
// 对应配置文件的根级别，包含所有子模块的配置
type Config struct {
	Server     ServerConfig     `yaml:"server"`     // HTTP 服务器配置
	Ceph       CephConfig       `yaml:"ceph"`       // Ceph 集群连接配置
	Prometheus PrometheusConfig `yaml:"prometheus"` // Prometheus 采集配置
	Logger     LoggerConfig     `yaml:"logger"`     // 日志系统配置
	Tracer     TracerConfig     `yaml:"tracer"`     // 追踪系统配置
	Plugins    []PluginConfig   `yaml:"plugins"`    // 插件配置列表
}

// ServerConfig HTTP 服务器配置
// 控制 ceph-exporter 对外暴露的 HTTP 服务行为，包括监听地址、端口、超时和 TLS
type ServerConfig struct {
	Host         string        `yaml:"host"`          // 监听地址，默认 "0.0.0.0"（监听所有网卡）
	Port         int           `yaml:"port"`          // 监听端口，默认 9128
	ReadTimeout  time.Duration `yaml:"read_timeout"`  // HTTP 读取超时时间，默认 30s
	WriteTimeout time.Duration `yaml:"write_timeout"` // HTTP 写入超时时间，默认 30s
	TLSCertFile  string        `yaml:"tls_cert_file"` // TLS 证书文件路径（留空不启用 HTTPS）
	TLSKeyFile   string        `yaml:"tls_key_file"`  // TLS 密钥文件路径（留空不启用 HTTPS）
}

// CephConfig Ceph 集群连接配置
// 定义如何连接到 Ceph 集群，包括配置文件路径、认证信息和超时设置
type CephConfig struct {
	ConfigFile string        `yaml:"config_file"` // Ceph 配置文件路径，默认 /etc/ceph/ceph.conf
	User       string        `yaml:"user"`        // Ceph 认证用户名，默认 admin
	Keyring    string        `yaml:"keyring"`     // Keyring 认证文件路径
	Cluster    string        `yaml:"cluster"`     // Ceph 集群名称，默认 ceph
	Timeout    time.Duration `yaml:"timeout"`     // 命令执行超时时间，默认 10s
}

// PrometheusConfig Prometheus 采集相关配置
// 控制指标采集的频率和超时
type PrometheusConfig struct {
	CollectInterval time.Duration `yaml:"collect_interval"` // 采集间隔，默认 15s
	Timeout         time.Duration `yaml:"timeout"`          // 单次采集超时时间，默认 10s
}

// LoggerConfig 日志系统配置
// 支持多级别日志、多种输出格式、文件轮转和 ELK 集成
type LoggerConfig struct {
	Level            string `yaml:"level"`             // 日志级别: trace, debug, info, warn, error, fatal, panic
	Format           string `yaml:"format"`            // 日志格式: json（结构化）, text（文本）
	Output           string `yaml:"output"`            // 输出目标: stdout, stderr, file
	FilePath         string `yaml:"file_path"`         // 日志文件路径（output=file 时生效）
	MaxSize          int    `yaml:"max_size"`          // 单个日志文件最大大小（MB），默认 100
	MaxBackups       int    `yaml:"max_backups"`       // 保留的旧日志文件数量，默认 3
	MaxAge           int    `yaml:"max_age"`           // 保留日志文件的最大天数，默认 28
	Compress         bool   `yaml:"compress"`          // 是否压缩归档的旧日志文件
	EnableELK        bool   `yaml:"enable_elk"`        // 是否启用 ELK 日志集成
	LogstashURL      string `yaml:"logstash_url"`      // Logstash 地址（enable_elk=true 时生效）
	LogstashProtocol string `yaml:"logstash_protocol"` // Logstash 协议: tcp（默认）, udp
	ServiceName      string `yaml:"service_name"`      // 服务名称，用于在 ELK 中标识日志来源
}

// TracerConfig 追踪系统配置
// 使用 OpenTelemetry + Jaeger 实现分布式追踪（Phase 3 完整实现）
type TracerConfig struct {
	Enabled     bool    `yaml:"enabled"`      // 是否启用追踪
	JaegerURL   string  `yaml:"jaeger_url"`   // Jaeger Collector URL
	ServiceName string  `yaml:"service_name"` // 服务名称，用于在 Jaeger 中标识
	SampleRate  float64 `yaml:"sample_rate"`  // 采样率 (0.0-1.0)，1.0 表示全量采集
}

// PluginConfig 插件配置
// 定义第三方存储系统监控插件的配置（Phase 5 完整实现）
type PluginConfig struct {
	Name    string                 `yaml:"name"`    // 插件名称，用于标识和日志
	Enabled bool                   `yaml:"enabled"` // 是否启用该插件
	Type    string                 `yaml:"type"`    // 插件类型: so（动态库）, http（远程调用）
	Path    string                 `yaml:"path"`    // 插件路径（.so 文件路径或 HTTP 地址）
	Config  map[string]interface{} `yaml:"config"`  // 插件自定义配置（键值对）
}

// SetDefaults 为所有配置项设置默认值
// 在配置文件解析完成后调用，确保未设置的配置项有合理的默认值
func (c *Config) SetDefaults() {
	// ----- Server 默认值 -----
	if c.Server.Host == "" {
		c.Server.Host = "0.0.0.0" // 默认监听所有网卡
	}
	if c.Server.Port == 0 {
		c.Server.Port = 9128 // Ceph Exporter 的标准端口
	}
	if c.Server.ReadTimeout == 0 {
		c.Server.ReadTimeout = 30 * time.Second
	}
	if c.Server.WriteTimeout == 0 {
		c.Server.WriteTimeout = 30 * time.Second
	}

	// ----- Ceph 默认值 -----
	if c.Ceph.ConfigFile == "" {
		c.Ceph.ConfigFile = "/etc/ceph/ceph.conf" // Ceph 标准配置文件路径
	}
	if c.Ceph.User == "" {
		c.Ceph.User = "admin" // 默认使用 admin 用户
	}
	if c.Ceph.Cluster == "" {
		c.Ceph.Cluster = "ceph" // 默认集群名称
	}
	if c.Ceph.Timeout == 0 {
		c.Ceph.Timeout = 10 * time.Second
	}

	// ----- Prometheus 默认值 -----
	if c.Prometheus.CollectInterval == 0 {
		c.Prometheus.CollectInterval = 15 * time.Second
	}
	if c.Prometheus.Timeout == 0 {
		c.Prometheus.Timeout = 10 * time.Second
	}

	// ----- Logger 默认值 -----
	if c.Logger.Level == "" {
		c.Logger.Level = "info" // 默认 info 级别
	}
	if c.Logger.Format == "" {
		c.Logger.Format = "json" // 默认 JSON 格式，便于 ELK 解析
	}
	if c.Logger.Output == "" {
		c.Logger.Output = "stdout" // 默认输出到标准输出
	}
	if c.Logger.MaxSize == 0 {
		c.Logger.MaxSize = 100 // 默认单文件最大 100MB
	}
	if c.Logger.MaxBackups == 0 {
		c.Logger.MaxBackups = 3 // 默认保留 3 个备份
	}
	if c.Logger.MaxAge == 0 {
		c.Logger.MaxAge = 28 // 默认保留 28 天
	}
	if c.Logger.LogstashProtocol == "" {
		c.Logger.LogstashProtocol = "tcp" // 默认使用 TCP 协议
	}
	if c.Logger.ServiceName == "" {
		c.Logger.ServiceName = "ceph-exporter" // 默认服务名称
	}

	// ----- Tracer 默认值 -----
	if c.Tracer.ServiceName == "" {
		c.Tracer.ServiceName = "ceph-exporter"
	}
	if c.Tracer.SampleRate == 0 {
		c.Tracer.SampleRate = 1.0 // 默认全量采样
	}
}

// Validate 验证配置的合法性
// 在设置默认值之后调用，确保所有配置项的值在合理范围内
// 返回:
//   - error: 如果配置不合法，返回对应的错误；合法则返回 nil
func (c *Config) Validate() error {
	// 验证端口号范围 (1-65535)
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return ErrInvalidPort
	}

	// 验证 Ceph 配置文件路径不能为空
	if c.Ceph.ConfigFile == "" {
		return ErrMissingCephConfig
	}

	// 验证日志级别是否为合法值
	validLevels := map[string]bool{
		"trace": true, "debug": true, "info": true,
		"warn": true, "error": true, "fatal": true, "panic": true,
	}
	if !validLevels[c.Logger.Level] {
		return ErrInvalidLogLevel
	}

	// 验证追踪配置：启用追踪时必须提供 Jaeger URL
	if c.Tracer.Enabled && c.Tracer.JaegerURL == "" {
		return ErrMissingJaegerURL
	}

	// 验证采样率范围 (0.0-1.0)
	if c.Tracer.SampleRate < 0 || c.Tracer.SampleRate > 1 {
		return ErrInvalidSampleRate
	}

	return nil
}
