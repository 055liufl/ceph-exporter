// =============================================================================
// Tracer - 分布式追踪模块
// =============================================================================
// 基于 OpenTelemetry 实现的分布式追踪系统，用于监控和分析请求链路。
//
// 功能特性:
//   - 支持 OTLP HTTP 协议导出追踪数据到 Jaeger
//   - 可配置采样率，控制追踪数据量
//   - 提供 Span 创建、属性设置、状态管理等便捷方法
//   - 支持从上下文中提取 TraceID 和 SpanID
//
// 配置项:
//   - Enabled: 是否启用追踪（默认 false）
//   - JaegerURL: Jaeger 收集器地址（如 localhost:4318）
//   - ServiceName: 服务名称（用于标识追踪来源）
//   - SampleRate: 采样率（0.0-1.0，1.0 表示全量采样）
//
// 使用示例:
//
//	// 创建追踪提供者
//	tp, err := NewTracerProvider(cfg, log)
//	defer tp.Shutdown(context.Background())
//
//	// 创建 Span
//	ctx, span := StartSpan(ctx, "operation-name")
//	defer span.End()
//
//	// 设置属性
//	SetAttributes(ctx, StringAttr("key", "value"))
//
// =============================================================================
package tracer

import (
	"context"

	"ceph-exporter/internal/config"
	"ceph-exporter/internal/logger"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

// TracerProvider 追踪提供者
// 封装 OpenTelemetry 的 TracerProvider，提供追踪系统的生命周期管理
//
// 字段说明:
//   - tp: OpenTelemetry SDK 的 TracerProvider 实例，负责创建和管理 Tracer
//   - log: 日志记录器，用于记录追踪系统的初始化和运行状态
type TracerProvider struct {
	tp  *sdktrace.TracerProvider // OpenTelemetry TracerProvider 实例
	log *logger.Logger           // 日志记录器
}

// NewTracerProvider 创建追踪提供者实例
// 根据配置初始化 OpenTelemetry 追踪系统，设置 OTLP HTTP 导出器和采样策略
//
// 初始化流程:
//  1. 检查追踪是否启用，未启用则返回空实例
//  2. 创建 OTLP HTTP 导出器，连接到 Jaeger 收集器
//  3. 创建资源描述符，标识服务名称
//  4. 创建 TracerProvider，配置批量导出器和采样率
//  5. 设置为全局 TracerProvider
//
// 参数:
//   - cfg: 追踪配置，包含 Jaeger 地址、服务名称、采样率等
//   - log: 日志实例，用于记录初始化过程
//
// 返回:
//   - *TracerProvider: 追踪提供者实例
//   - error: 初始化失败时返回错误（如无法连接 Jaeger）
//
// 注意事项:
//   - 如果追踪未启用，返回的实例不包含 TracerProvider，但不会报错
//   - 使用 WithInsecure() 选项，适用于开发环境，生产环境应使用 TLS
//   - 采样率为 1.0 表示全量采样，0.0 表示不采样
func NewTracerProvider(cfg *config.TracerConfig, log *logger.Logger) (*TracerProvider, error) {
	// 检查追踪是否启用
	if !cfg.Enabled {
		log.WithComponent("tracer").Info("追踪系统未启用")
		return &TracerProvider{log: log}, nil
	}

	ctx := context.Background()

	// 创建 OTLP HTTP 导出器，用于将追踪数据发送到 Jaeger
	// WithInsecure() 表示使用 HTTP 而非 HTTPS（适用于开发环境）
	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(cfg.JaegerURL),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	// 创建资源描述符，用于标识追踪数据的来源服务
	// 使用语义化约定（semconv）设置服务名称
	res, err := resource.New(ctx,
		resource.WithAttributes(semconv.ServiceName(cfg.ServiceName)),
	)
	if err != nil {
		// 资源创建失败不影响追踪功能，使用默认资源
		log.WithComponent("tracer").Warnf("创建资源失败，使用默认资源: %v", err)
		res = resource.Default()
	}

	// 创建 TracerProvider，配置导出器、资源和采样策略
	// WithBatcher: 批量导出追踪数据，提高性能
	// WithResource: 设置服务标识信息
	// WithSampler: 设置采样率，控制追踪数据量
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(cfg.SampleRate)),
	)

	// 设置为全局 TracerProvider，使得 otel.Tracer() 可以获取到此实例
	otel.SetTracerProvider(tp)
	log.WithComponent("tracer").Info("追踪系统已启用")
	return &TracerProvider{tp: tp, log: log}, nil
}

// Shutdown 关闭追踪提供者
// 优雅地关闭 TracerProvider，确保所有待发送的追踪数据都被导出
//
// 参数:
//   - ctx: 上下文，用于控制关闭超时
//
// 返回:
//   - error: 关闭失败时返回错误
//
// 注意事项:
//   - 应在程序退出前调用此方法，确保追踪数据不丢失
//   - 如果追踪未启用（tp 为 nil），直接返回 nil
func (tp *TracerProvider) Shutdown(ctx context.Context) error {
	if tp.tp != nil {
		return tp.tp.Shutdown(ctx)
	}
	return nil
}

// StartSpan 创建并启动一个新的 Span
// Span 代表一个操作的执行过程，用于追踪请求链路中的单个操作
//
// 参数:
//   - ctx: 父上下文，如果包含 Span 则创建子 Span
//   - name: Span 名称，描述操作类型（如 "GetOSDStats", "HTTP GET /metrics"）
//
// 返回:
//   - context.Context: 包含新 Span 的上下文，应传递给后续操作
//   - trace.Span: Span 实例，调用者应在操作完成后调用 span.End()
//
// 使用示例:
//
//	ctx, span := StartSpan(ctx, "database-query")
//	defer span.End()
//	// 执行数据库查询...
func StartSpan(ctx context.Context, name string) (context.Context, trace.Span) {
	return otel.Tracer("ceph-exporter").Start(ctx, name)
}

// GetTraceID 从上下文中提取 TraceID
// TraceID 是追踪链路的唯一标识符，用于关联同一请求的所有 Span
//
// 参数:
//   - ctx: 包含 Span 的上下文
//
// 返回:
//   - string: TraceID 的十六进制字符串表示，如果上下文中没有有效 Span 则返回空字符串
//
// 使用场景:
//   - 在日志中记录 TraceID，便于关联日志和追踪数据
//   - 在 HTTP 响应头中返回 TraceID，便于客户端追踪
func GetTraceID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		return span.SpanContext().TraceID().String()
	}
	return ""
}

// GetSpanID 从上下文中提取 SpanID
// SpanID 是当前 Span 的唯一标识符，用于标识追踪链路中的单个操作
//
// 参数:
//   - ctx: 包含 Span 的上下文
//
// 返回:
//   - string: SpanID 的十六进制字符串表示，如果上下文中没有有效 Span 则返回空字符串
//
// 使用场景:
//   - 在日志中记录 SpanID，便于定位具体操作
//   - 在错误报告中包含 SpanID，便于追踪问题
func GetSpanID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		return span.SpanContext().SpanID().String()
	}
	return ""
}

// SetAttributes 为当前 Span 设置属性
// 属性用于为 Span 添加额外的元数据，便于在追踪系统中过滤和分析
//
// 参数:
//   - ctx: 包含 Span 的上下文
//   - attrs: 一个或多个属性键值对
//
// 使用示例:
//
//	SetAttributes(ctx,
//	    StringAttr("http.method", "GET"),
//	    IntAttr("http.status_code", 200),
//	)
//
// 注意事项:
//   - 如果上下文中没有 Span，此操作不会报错但也不会生效
//   - 属性会被发送到追踪后端，避免添加敏感信息
func SetAttributes(ctx context.Context, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attrs...)
}

// SpanStatus Span 状态类型
// 用于标识 Span 执行的结果状态
type SpanStatus int

const (
	StatusOK    SpanStatus = 0 // 成功 - 操作正常完成
	StatusError SpanStatus = 1 // 错误 - 操作执行失败
)

// SetSpanStatus 设置 Span 状态
// 用于标记 Span 的执行结果，便于在追踪系统中识别成功和失败的操作
//
// 参数:
//   - ctx: 包含 Span 的上下文
//   - status: Span 状态（StatusOK 或 StatusError）
//   - description: 状态描述，通常在错误时包含错误信息
//
// 使用示例:
//
//	if err != nil {
//	    SetSpanStatus(ctx, StatusError, err.Error())
//	} else {
//	    SetSpanStatus(ctx, StatusOK, "操作成功")
//	}
//
// 注意事项:
//   - 应在 Span 结束前调用此方法
//   - 错误状态会在追踪系统中高亮显示，便于快速定位问题
func SetSpanStatus(ctx context.Context, status SpanStatus, description string) {
	span := trace.SpanFromContext(ctx)
	if status == StatusError {
		// 设置为错误状态，追踪系统会将此 Span 标记为失败
		span.SetStatus(codes.Error, description)
	} else {
		// 设置为成功状态
		span.SetStatus(codes.Ok, description)
	}
}

// StringAttr 创建字符串类型的属性
// 用于为 Span 添加字符串类型的元数据
//
// 参数:
//   - key: 属性键（如 "http.method", "db.statement"）
//   - value: 属性值
//
// 返回:
//   - attribute.KeyValue: 可传递给 SetAttributes 的属性对象
func StringAttr(key, value string) attribute.KeyValue {
	return attribute.String(key, value)
}

// IntAttr 创建整数类型的属性
// 用于为 Span 添加整数类型的元数据
//
// 参数:
//   - key: 属性键（如 "http.status_code", "db.rows_affected"）
//   - value: 属性值
//
// 返回:
//   - attribute.KeyValue: 可传递给 SetAttributes 的属性对象
func IntAttr(key string, value int) attribute.KeyValue {
	return attribute.Int(key, value)
}

// BoolAttr 创建布尔类型的属性
// 用于为 Span 添加布尔类型的元数据
//
// 参数:
//   - key: 属性键（如 "cache.hit", "retry.enabled"）
//   - value: 属性值
//
// 返回:
//   - attribute.KeyValue: 可传递给 SetAttributes 的属性对象
func BoolAttr(key string, value bool) attribute.KeyValue {
	return attribute.Bool(key, value)
}
