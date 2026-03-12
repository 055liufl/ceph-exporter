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

type TracerProvider struct {
	tp  *sdktrace.TracerProvider
	log *logger.Logger
}

func NewTracerProvider(cfg *config.TracerConfig, log *logger.Logger) (*TracerProvider, error) {
	if !cfg.Enabled {
		log.WithComponent("tracer").Info("追踪系统未启用")
		return &TracerProvider{log: log}, nil
	}

	ctx := context.Background()
	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(cfg.JaegerURL),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(semconv.ServiceName(cfg.ServiceName)),
	)
	if err != nil {
		log.WithComponent("tracer").Warnf("创建资源失败，使用默认资源: %v", err)
		res = resource.Default()
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(cfg.SampleRate)),
	)

	otel.SetTracerProvider(tp)
	log.WithComponent("tracer").Info("追踪系统已启用")
	return &TracerProvider{tp: tp, log: log}, nil
}

func (tp *TracerProvider) Shutdown(ctx context.Context) error {
	if tp.tp != nil {
		return tp.tp.Shutdown(ctx)
	}
	return nil
}

func StartSpan(ctx context.Context, name string) (context.Context, trace.Span) {
	return otel.Tracer("ceph-exporter").Start(ctx, name)
}

func GetTraceID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		return span.SpanContext().TraceID().String()
	}
	return ""
}

func GetSpanID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		return span.SpanContext().SpanID().String()
	}
	return ""
}

func SetAttributes(ctx context.Context, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attrs...)
}

// SpanStatus Span 状态类型
type SpanStatus int

const (
	StatusOK    SpanStatus = 0 // 成功
	StatusError SpanStatus = 1 // 错误
)

// SetSpanStatus 设置 Span 状态
func SetSpanStatus(ctx context.Context, status SpanStatus, description string) {
	span := trace.SpanFromContext(ctx)
	if status == StatusError {
		span.SetStatus(codes.Error, description)
	} else {
		span.SetStatus(codes.Ok, description)
	}
}

// StringAttr 创建字符串属性
func StringAttr(key, value string) attribute.KeyValue {
	return attribute.String(key, value)
}

// IntAttr 创建整数属性
func IntAttr(key string, value int) attribute.KeyValue {
	return attribute.Int(key, value)
}

// BoolAttr 创建布尔属性
func BoolAttr(key string, value bool) attribute.KeyValue {
	return attribute.Bool(key, value)
}
