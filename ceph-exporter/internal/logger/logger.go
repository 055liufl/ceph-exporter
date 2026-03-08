// =============================================================================
// 日志模块
// =============================================================================
// 提供结构化日志功能，支持多种输出方式:
//   - stdout/stderr: 标准输出/标准错误
//   - file: 文件输出（支持日志轮转）
//
// 日志格式:
//   - json: JSON 结构化格式，便于 ELK 等日志系统解析
//   - text: 文本格式，便于人工阅读
//
// 日志级别（从低到高）:
//
//	trace -> debug -> info -> warn -> error -> fatal -> panic
//
// 特性:
//   - 基于 logrus 实现结构化日志
//   - 支持 lumberjack 日志轮转（按大小、数量、天数）
//   - 支持追踪 ID 关联（trace_id, span_id）
//   - 支持组件标识（component 字段）
//   - ELK 集成预留（Phase 3 实现）
//
// =============================================================================
package logger

import (
	"fmt"
	"io"
	"os"

	"ceph-exporter/internal/config"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger 日志实例封装
// 在 logrus.Logger 基础上增加了配置管理和文件写入器的生命周期管理
type Logger struct {
	*logrus.Logger                      // 内嵌 logrus.Logger，继承所有日志方法
	config         *config.LoggerConfig // 日志配置引用
	fileWriter     io.WriteCloser       // 文件写入器（output=file 时使用），需要在 Close 时释放
}

// NewLogger 创建新的日志实例
// 根据配置初始化日志级别、格式和输出目标
//
// 参数:
//   - cfg: 日志配置，包含级别、格式、输出目标等设置
//
// 返回:
//   - *Logger: 初始化完成的日志实例
//   - error: 初始化过程中的错误（无效级别、文件创建失败等）
func NewLogger(cfg *config.LoggerConfig) (*Logger, error) {
	// 创建 logrus 实例
	log := logrus.New()

	// 设置日志级别
	// logrus.ParseLevel 支持: trace, debug, info, warn/warning, error, fatal, panic
	level, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		return nil, fmt.Errorf("无效的日志级别 '%s': %w", cfg.Level, err)
	}
	log.SetLevel(level)

	// 根据配置设置日志格式
	switch cfg.Format {
	case "json":
		// JSON 格式: 适合机器解析，便于 ELK 等日志系统处理
		log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05", // Go 的时间格式化模板
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "timestamp", // 将默认的 "time" 字段重命名为 "timestamp"
				logrus.FieldKeyLevel: "level",     // 保持 "level" 字段名
				logrus.FieldKeyMsg:   "message",   // 将默认的 "msg" 字段重命名为 "message"
			},
		})
	default:
		// Text 格式: 适合人工阅读，开发调试时使用
		log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,                  // 显示完整时间戳
			TimestampFormat: "2006-01-02 15:04:05", // 时间格式
		})
	}

	// 创建 Logger 实例
	logger := &Logger{
		Logger: log,
		config: cfg,
	}

	// 设置输出目标（stdout/stderr/file）
	if err := logger.setOutput(); err != nil {
		return nil, err
	}

	// ELK 集成提示（Phase 3 完整实现）
	if cfg.EnableELK && cfg.LogstashURL != "" {
		// TODO: Phase 3 实现 Logstash Hook
		// hook := NewLogstashHook(cfg.LogstashURL)
		// log.AddHook(hook)
		logger.Info("ELK 集成已配置（将在 Phase 3 完整实现）")
	}

	return logger, nil
}

// setOutput 设置日志输出目标
// 根据配置中的 output 字段选择输出方式:
//   - "stdout": 输出到标准输出（默认）
//   - "stderr": 输出到标准错误
//   - "file": 输出到文件（使用 lumberjack 实现日志轮转）
//
// 返回:
//   - error: 设置过程中的错误（如文件路径为空）
func (l *Logger) setOutput() error {
	switch l.config.Output {
	case "stdout":
		// 标准输出，适合容器化部署（日志由容器运行时收集）
		l.SetOutput(os.Stdout)

	case "stderr":
		// 标准错误输出
		l.SetOutput(os.Stderr)

	case "file":
		// 文件输出，使用 lumberjack 实现自动轮转
		if l.config.FilePath == "" {
			return fmt.Errorf("日志输出目标为 file 时，file_path 不能为空")
		}

		// lumberjack.Logger 实现了 io.WriteCloser 接口
		// 自动处理日志文件的创建、轮转、压缩和清理
		fileWriter := &lumberjack.Logger{
			Filename:   l.config.FilePath,   // 日志文件路径
			MaxSize:    l.config.MaxSize,    // 单个文件最大大小（MB），超过后自动轮转
			MaxBackups: l.config.MaxBackups, // 保留的旧文件数量
			MaxAge:     l.config.MaxAge,     // 保留的最大天数
			Compress:   l.config.Compress,   // 是否压缩旧文件（gzip）
		}

		l.fileWriter = fileWriter
		l.SetOutput(fileWriter)

	default:
		// 未知输出目标，回退到标准输出
		l.SetOutput(os.Stdout)
	}

	return nil
}

// Close 关闭日志实例，释放资源
// 主要用于关闭文件输出的写入器，确保日志数据完整写入磁盘
//
// 返回:
//   - error: 关闭过程中的错误
func (l *Logger) Close() error {
	if l.fileWriter != nil {
		return l.fileWriter.Close()
	}
	return nil
}

// WithTraceID 创建带追踪 ID 的日志条目
// 用于将日志与分布式追踪数据关联，便于在 Jaeger 和 ELK 之间跳转查询
//
// 参数:
//   - traceID: OpenTelemetry 追踪 ID
//
// 返回:
//   - *logrus.Entry: 带有 trace_id 字段的日志条目
func (l *Logger) WithTraceID(traceID string) *logrus.Entry {
	return l.WithField("trace_id", traceID)
}

// WithSpanID 创建带 Span ID 的日志条目
// 用于精确定位日志对应的追踪 Span
//
// 参数:
//   - spanID: OpenTelemetry Span ID
//
// 返回:
//   - *logrus.Entry: 带有 span_id 字段的日志条目
func (l *Logger) WithSpanID(spanID string) *logrus.Entry {
	return l.WithField("span_id", spanID)
}

// WithComponent 创建带组件名称的日志条目
// 用于标识日志来源的模块/组件，便于按组件过滤日志
//
// 参数:
//   - component: 组件名称（如 "ceph-client", "http-server", "collector"）
//
// 返回:
//   - *logrus.Entry: 带有 component 字段的日志条目
func (l *Logger) WithComponent(component string) *logrus.Entry {
	return l.WithField("component", component)
}
