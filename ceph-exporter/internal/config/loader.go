// =============================================================================
// 配置加载器
// =============================================================================
// 本文件负责从 YAML 文件加载配置，并支持环境变量覆盖。
//
// 配置加载流程:
//  1. 检查配置文件是否存在
//  2. 读取并解析 YAML 文件
//  3. 设置默认值（填充未配置的字段）
//  4. 应用环境变量覆盖（优先级高于配置文件）
//  5. 验证配置合法性
//
// 环境变量对照表:
//
//	CEPH_EXPORTER_HOST  -> server.host
//	CEPH_EXPORTER_PORT  -> server.port
//	CEPH_CONFIG         -> ceph.config_file
//	CEPH_USER           -> ceph.user
//	CEPH_KEYRING        -> ceph.keyring
//	CEPH_CLUSTER        -> ceph.cluster
//	LOG_LEVEL           -> logger.level
//	LOG_FORMAT          -> logger.format
//	LOGSTASH_URL        -> logger.logstash_url
//	JAEGER_URL          -> tracer.jaeger_url
//	SERVICE_NAME        -> tracer.service_name
//
// =============================================================================
package config

import (
	"errors"
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

// 配置相关的预定义错误
// 使用 sentinel error 模式，方便调用方通过 errors.Is() 判断具体错误类型
var (
	ErrInvalidPort        = errors.New("无效的端口号，必须在 1-65535 之间")
	ErrMissingCephConfig  = errors.New("缺少 Ceph 配置文件路径")
	ErrInvalidLogLevel    = errors.New("无效的日志级别，支持: trace, debug, info, warn, error, fatal, panic")
	ErrMissingJaegerURL   = errors.New("启用追踪时必须提供 Jaeger URL")
	ErrInvalidSampleRate  = errors.New("采样率必须在 0.0 到 1.0 之间")
	ErrConfigFileNotFound = errors.New("配置文件不存在")
)

// LoadConfig 从指定路径加载配置文件
// 这是配置模块的主入口函数，完成配置的完整加载流程:
// 读取文件 -> 解析 YAML -> 设置默认值 -> 环境变量覆盖 -> 验证
//
// 参数:
//   - path: 配置文件的绝对或相对路径
//
// 返回:
//   - *Config: 解析并验证后的配置对象
//   - error: 加载过程中的错误（文件不存在、解析失败、验证失败等）
func LoadConfig(path string) (*Config, error) {
	// 第一步: 检查配置文件是否存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("%w: %s", ErrConfigFileNotFound, path)
	}

	// 第二步: 读取配置文件内容
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 第三步: 解析 YAML 内容到 Config 结构体
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 第四步: 为未设置的配置项填充默认值
	cfg.SetDefaults()

	// 第五步: 应用环境变量覆盖（环境变量优先级高于配置文件）
	applyEnvOverrides(&cfg)

	// 第六步: 验证配置的合法性
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	return &cfg, nil
}

// applyEnvOverrides 应用环境变量覆盖配置
// 环境变量的优先级高于配置文件，适用于容器化部署场景。
// 只有当环境变量非空时才会覆盖对应的配置项。
//
// 参数:
//   - cfg: 待覆盖的配置对象指针
func applyEnvOverrides(cfg *Config) {
	// ----- Server 配置覆盖 -----
	if host := os.Getenv("CEPH_EXPORTER_HOST"); host != "" {
		cfg.Server.Host = host
	}
	if port := os.Getenv("CEPH_EXPORTER_PORT"); port != "" {
		// 使用 Sscanf 安全地将字符串转换为整数
		if _, err := fmt.Sscanf(port, "%d", &cfg.Server.Port); err != nil {
			// 如果转换失败，保持默认端口
			log.Printf("Warning: invalid port value %q, using default", port)
		}
	}

	// ----- Ceph 配置覆盖 -----
	if configFile := os.Getenv("CEPH_CONFIG"); configFile != "" {
		cfg.Ceph.ConfigFile = configFile
	}
	if user := os.Getenv("CEPH_USER"); user != "" {
		cfg.Ceph.User = user
	}
	if keyring := os.Getenv("CEPH_KEYRING"); keyring != "" {
		cfg.Ceph.Keyring = keyring
	}
	if cluster := os.Getenv("CEPH_CLUSTER"); cluster != "" {
		cfg.Ceph.Cluster = cluster
	}

	// ----- Logger 配置覆盖 -----
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		cfg.Logger.Level = level
	}
	if format := os.Getenv("LOG_FORMAT"); format != "" {
		cfg.Logger.Format = format
	}
	if logstashURL := os.Getenv("LOGSTASH_URL"); logstashURL != "" {
		cfg.Logger.LogstashURL = logstashURL
	}

	// ----- Tracer 配置覆盖 -----
	if jaegerURL := os.Getenv("JAEGER_URL"); jaegerURL != "" {
		cfg.Tracer.JaegerURL = jaegerURL
	}
	if serviceName := os.Getenv("SERVICE_NAME"); serviceName != "" {
		cfg.Tracer.ServiceName = serviceName
	}
}

// SaveConfig 将配置对象序列化并保存到文件
// 主要用于生成默认配置文件或导出当前运行配置
//
// 参数:
//   - cfg: 要保存的配置对象
//   - path: 保存的目标文件路径
//
// 返回:
//   - error: 序列化或写入过程中的错误
func SaveConfig(cfg *Config, path string) error {
	// 将配置结构体序列化为 YAML 格式
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	// 写入文件，权限 0644（所有者读写，其他人只读）
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	return nil
}
