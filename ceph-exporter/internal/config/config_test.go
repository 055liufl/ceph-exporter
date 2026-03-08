// =============================================================================
// 配置模块单元测试
// =============================================================================
// 测试覆盖:
//   - 配置文件加载（正常/异常路径）
//   - 默认值设置
//   - 环境变量覆盖
//   - 配置验证（端口、日志级别、采样率等）
//   - 配置保存
//
// =============================================================================
package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// testConfigYAML 用于测试的完整配置文件内容
const testConfigYAML = `
server:
  host: "127.0.0.1"
  port: 9128
  read_timeout: 15s
  write_timeout: 15s

ceph:
  config_file: "/etc/ceph/ceph.conf"
  user: "admin"
  keyring: "/etc/ceph/ceph.client.admin.keyring"
  cluster: "ceph"
  timeout: 10s

prometheus:
  collect_interval: 15s
  timeout: 10s

logger:
  level: "info"
  format: "json"
  output: "stdout"

tracer:
  enabled: false
  service_name: "ceph-exporter"
  sample_rate: 1.0
`

// testMinimalConfigYAML 最小配置文件（仅包含必要字段）
const testMinimalConfigYAML = `
ceph:
  config_file: "/etc/ceph/ceph.conf"
`

// createTempConfigFile 创建临时配置文件用于测试
// 参数:
//   - t: 测试实例
//   - content: 配置文件内容
//
// 返回:
//   - string: 临时文件路径
func createTempConfigFile(t *testing.T, content string) string {
	t.Helper()

	// 创建临时目录
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "config.yaml")

	// 写入配置内容
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("创建临时配置文件失败: %v", err)
	}

	return tmpFile
}

// TestLoadConfig_Success 测试正常加载配置文件
func TestLoadConfig_Success(t *testing.T) {
	// 创建临时配置文件
	configPath := createTempConfigFile(t, testConfigYAML)

	// 加载配置
	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	// 验证 Server 配置
	if cfg.Server.Host != "127.0.0.1" {
		t.Errorf("Server.Host 期望 '127.0.0.1'，实际 '%s'", cfg.Server.Host)
	}
	if cfg.Server.Port != 9128 {
		t.Errorf("Server.Port 期望 9128，实际 %d", cfg.Server.Port)
	}
	if cfg.Server.ReadTimeout != 15*time.Second {
		t.Errorf("Server.ReadTimeout 期望 15s，实际 %v", cfg.Server.ReadTimeout)
	}

	// 验证 Ceph 配置
	if cfg.Ceph.ConfigFile != "/etc/ceph/ceph.conf" {
		t.Errorf("Ceph.ConfigFile 期望 '/etc/ceph/ceph.conf'，实际 '%s'", cfg.Ceph.ConfigFile)
	}
	if cfg.Ceph.User != "admin" {
		t.Errorf("Ceph.User 期望 'admin'，实际 '%s'", cfg.Ceph.User)
	}

	// 验证 Logger 配置
	if cfg.Logger.Level != "info" {
		t.Errorf("Logger.Level 期望 'info'，实际 '%s'", cfg.Logger.Level)
	}
	if cfg.Logger.Format != "json" {
		t.Errorf("Logger.Format 期望 'json'，实际 '%s'", cfg.Logger.Format)
	}
}

// TestLoadConfig_MinimalConfig 测试最小配置文件（验证默认值填充）
func TestLoadConfig_MinimalConfig(t *testing.T) {
	configPath := createTempConfigFile(t, testMinimalConfigYAML)

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("加载最小配置失败: %v", err)
	}

	// 验证默认值是否正确设置
	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("默认 Server.Host 期望 '0.0.0.0'，实际 '%s'", cfg.Server.Host)
	}
	if cfg.Server.Port != 9128 {
		t.Errorf("默认 Server.Port 期望 9128，实际 %d", cfg.Server.Port)
	}
	if cfg.Ceph.User != "admin" {
		t.Errorf("默认 Ceph.User 期望 'admin'，实际 '%s'", cfg.Ceph.User)
	}
	if cfg.Ceph.Cluster != "ceph" {
		t.Errorf("默认 Ceph.Cluster 期望 'ceph'，实际 '%s'", cfg.Ceph.Cluster)
	}
	if cfg.Logger.Level != "info" {
		t.Errorf("默认 Logger.Level 期望 'info'，实际 '%s'", cfg.Logger.Level)
	}
	if cfg.Logger.Format != "json" {
		t.Errorf("默认 Logger.Format 期望 'json'，实际 '%s'", cfg.Logger.Format)
	}
	if cfg.Logger.MaxSize != 100 {
		t.Errorf("默认 Logger.MaxSize 期望 100，实际 %d", cfg.Logger.MaxSize)
	}
	if cfg.Prometheus.CollectInterval != 15*time.Second {
		t.Errorf("默认 Prometheus.CollectInterval 期望 15s，实际 %v", cfg.Prometheus.CollectInterval)
	}
}

// TestLoadConfig_FileNotFound 测试配置文件不存在的情况
func TestLoadConfig_FileNotFound(t *testing.T) {
	_, err := LoadConfig("/nonexistent/path/config.yaml")
	if err == nil {
		t.Fatal("期望返回错误，但返回了 nil")
	}
}

// TestLoadConfig_InvalidYAML 测试无效 YAML 格式
func TestLoadConfig_InvalidYAML(t *testing.T) {
	configPath := createTempConfigFile(t, "invalid: yaml: [content")

	_, err := LoadConfig(configPath)
	if err == nil {
		t.Fatal("期望返回解析错误，但返回了 nil")
	}
}

// TestConfig_Validate_InvalidPort 测试无效端口号验证
func TestConfig_Validate_InvalidPort(t *testing.T) {
	// 使用 table-driven tests 测试多种无效端口
	tests := []struct {
		name string // 测试用例名称
		port int    // 测试端口值
	}{
		{"端口为0", 0},
		{"端口为负数", -1},
		{"端口超过最大值", 65536},
		{"端口远超最大值", 100000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Server: ServerConfig{Port: tt.port},
				Ceph:   CephConfig{ConfigFile: "/etc/ceph/ceph.conf"},
				Logger: LoggerConfig{Level: "info"},
			}
			if err := cfg.Validate(); err != ErrInvalidPort {
				t.Errorf("端口 %d: 期望 ErrInvalidPort，实际 %v", tt.port, err)
			}
		})
	}
}

// TestConfig_Validate_InvalidLogLevel 测试无效日志级别验证
func TestConfig_Validate_InvalidLogLevel(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{Port: 9128},
		Ceph:   CephConfig{ConfigFile: "/etc/ceph/ceph.conf"},
		Logger: LoggerConfig{Level: "invalid_level"},
	}

	if err := cfg.Validate(); err != ErrInvalidLogLevel {
		t.Errorf("期望 ErrInvalidLogLevel，实际 %v", err)
	}
}

// TestConfig_Validate_InvalidSampleRate 测试无效采样率验证
func TestConfig_Validate_InvalidSampleRate(t *testing.T) {
	tests := []struct {
		name       string
		sampleRate float64
	}{
		{"采样率为负数", -0.1},
		{"采样率超过1", 1.1},
		{"采样率远超1", 5.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Server: ServerConfig{Port: 9128},
				Ceph:   CephConfig{ConfigFile: "/etc/ceph/ceph.conf"},
				Logger: LoggerConfig{Level: "info"},
				Tracer: TracerConfig{SampleRate: tt.sampleRate},
			}
			if err := cfg.Validate(); err != ErrInvalidSampleRate {
				t.Errorf("采样率 %f: 期望 ErrInvalidSampleRate，实际 %v", tt.sampleRate, err)
			}
		})
	}
}

// TestConfig_Validate_TracerRequiresJaegerURL 测试启用追踪时必须提供 Jaeger URL
func TestConfig_Validate_TracerRequiresJaegerURL(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{Port: 9128},
		Ceph:   CephConfig{ConfigFile: "/etc/ceph/ceph.conf"},
		Logger: LoggerConfig{Level: "info"},
		Tracer: TracerConfig{
			Enabled:    true,
			JaegerURL:  "", // 未提供 Jaeger URL
			SampleRate: 1.0,
		},
	}

	if err := cfg.Validate(); err != ErrMissingJaegerURL {
		t.Errorf("期望 ErrMissingJaegerURL，实际 %v", err)
	}
}

// TestConfig_Validate_Success 测试合法配置通过验证
func TestConfig_Validate_Success(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{Port: 9128},
		Ceph:   CephConfig{ConfigFile: "/etc/ceph/ceph.conf"},
		Logger: LoggerConfig{Level: "info"},
		Tracer: TracerConfig{
			Enabled:    true,
			JaegerURL:  "http://jaeger:14268/api/traces",
			SampleRate: 0.5,
		},
	}

	if err := cfg.Validate(); err != nil {
		t.Errorf("合法配置验证失败: %v", err)
	}
}

// TestApplyEnvOverrides 测试环境变量覆盖功能
func TestApplyEnvOverrides(t *testing.T) {
	// 设置测试环境变量
	envVars := map[string]string{
		"CEPH_EXPORTER_HOST": "192.168.1.100",
		"CEPH_EXPORTER_PORT": "9999",
		"CEPH_CONFIG":        "/custom/ceph.conf",
		"CEPH_USER":          "monitor",
		"LOG_LEVEL":          "debug",
	}

	// 设置环境变量并在测试结束后清理
	for k, v := range envVars {
		os.Setenv(k, v)
		defer os.Unsetenv(k)
	}

	// 创建初始配置
	cfg := &Config{}
	cfg.SetDefaults()

	// 应用环境变量覆盖
	applyEnvOverrides(cfg)

	// 验证覆盖结果
	if cfg.Server.Host != "192.168.1.100" {
		t.Errorf("环境变量覆盖 Host 失败: 期望 '192.168.1.100'，实际 '%s'", cfg.Server.Host)
	}
	if cfg.Server.Port != 9999 {
		t.Errorf("环境变量覆盖 Port 失败: 期望 9999，实际 %d", cfg.Server.Port)
	}
	if cfg.Ceph.ConfigFile != "/custom/ceph.conf" {
		t.Errorf("环境变量覆盖 ConfigFile 失败: 期望 '/custom/ceph.conf'，实际 '%s'", cfg.Ceph.ConfigFile)
	}
	if cfg.Ceph.User != "monitor" {
		t.Errorf("环境变量覆盖 User 失败: 期望 'monitor'，实际 '%s'", cfg.Ceph.User)
	}
	if cfg.Logger.Level != "debug" {
		t.Errorf("环境变量覆盖 Level 失败: 期望 'debug'，实际 '%s'", cfg.Logger.Level)
	}
}

// TestSetDefaults 测试默认值设置
func TestSetDefaults(t *testing.T) {
	cfg := &Config{}
	cfg.SetDefaults()

	// 验证所有默认值
	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("默认 Host 期望 '0.0.0.0'，实际 '%s'", cfg.Server.Host)
	}
	if cfg.Server.Port != 9128 {
		t.Errorf("默认 Port 期望 9128，实际 %d", cfg.Server.Port)
	}
	if cfg.Server.ReadTimeout != 30*time.Second {
		t.Errorf("默认 ReadTimeout 期望 30s，实际 %v", cfg.Server.ReadTimeout)
	}
	if cfg.Ceph.ConfigFile != "/etc/ceph/ceph.conf" {
		t.Errorf("默认 ConfigFile 期望 '/etc/ceph/ceph.conf'，实际 '%s'", cfg.Ceph.ConfigFile)
	}
	if cfg.Ceph.User != "admin" {
		t.Errorf("默认 User 期望 'admin'，实际 '%s'", cfg.Ceph.User)
	}
	if cfg.Ceph.Cluster != "ceph" {
		t.Errorf("默认 Cluster 期望 'ceph'，实际 '%s'", cfg.Ceph.Cluster)
	}
	if cfg.Tracer.ServiceName != "ceph-exporter" {
		t.Errorf("默认 ServiceName 期望 'ceph-exporter'，实际 '%s'", cfg.Tracer.ServiceName)
	}
	if cfg.Tracer.SampleRate != 1.0 {
		t.Errorf("默认 SampleRate 期望 1.0，实际 %f", cfg.Tracer.SampleRate)
	}
}

// TestSaveConfig 测试配置保存功能
func TestSaveConfig(t *testing.T) {
	// 创建测试配置
	cfg := &Config{
		Server: ServerConfig{
			Host: "127.0.0.1",
			Port: 9128,
		},
		Ceph: CephConfig{
			ConfigFile: "/etc/ceph/ceph.conf",
			User:       "admin",
		},
		Logger: LoggerConfig{
			Level:  "info",
			Format: "json",
		},
	}

	// 保存到临时文件
	tmpDir := t.TempDir()
	savePath := filepath.Join(tmpDir, "saved_config.yaml")

	if err := SaveConfig(cfg, savePath); err != nil {
		t.Fatalf("保存配置失败: %v", err)
	}

	// 验证文件已创建
	if _, err := os.Stat(savePath); os.IsNotExist(err) {
		t.Fatal("配置文件未创建")
	}

	// 重新加载并验证内容
	// 注意: 重新加载时会设置默认值，所以需要验证关键字段
	loaded, err := LoadConfig(savePath)
	if err != nil {
		t.Fatalf("重新加载保存的配置失败: %v", err)
	}

	if loaded.Server.Host != cfg.Server.Host {
		t.Errorf("保存后 Host 不匹配: 期望 '%s'，实际 '%s'", cfg.Server.Host, loaded.Server.Host)
	}
	if loaded.Server.Port != cfg.Server.Port {
		t.Errorf("保存后 Port 不匹配: 期望 %d，实际 %d", cfg.Server.Port, loaded.Server.Port)
	}
}

// TestSetDefaults_DoesNotOverrideExisting 测试默认值不会覆盖已有值
func TestSetDefaults_DoesNotOverrideExisting(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Host: "192.168.1.1",
			Port: 8080,
		},
		Ceph: CephConfig{
			User: "monitor",
		},
		Logger: LoggerConfig{
			Level: "debug",
		},
	}

	cfg.SetDefaults()

	// 验证已有值未被覆盖
	if cfg.Server.Host != "192.168.1.1" {
		t.Errorf("已有 Host 被覆盖: 期望 '192.168.1.1'，实际 '%s'", cfg.Server.Host)
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("已有 Port 被覆盖: 期望 8080，实际 %d", cfg.Server.Port)
	}
	if cfg.Ceph.User != "monitor" {
		t.Errorf("已有 User 被覆盖: 期望 'monitor'，实际 '%s'", cfg.Ceph.User)
	}
	if cfg.Logger.Level != "debug" {
		t.Errorf("已有 Level 被覆盖: 期望 'debug'，实际 '%s'", cfg.Logger.Level)
	}
}
