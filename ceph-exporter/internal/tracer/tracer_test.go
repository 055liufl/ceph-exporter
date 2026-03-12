// =============================================================================
// 追踪模块单元测试
// =============================================================================
// 测试 Phase 1 占位实现的正确性:
//   - TracerProvider 创建和关闭
//   - StartSpan 返回有效的上下文和 Span
//   - GetTraceID / GetSpanID 返回空字符串
//   - Span.End() 和 SetAttributes() 不 panic
//
// =============================================================================
package tracer

import (
	"context"
	"testing"

	"ceph-exporter/internal/config"
	"ceph-exporter/internal/logger"

	"go.opentelemetry.io/otel/attribute"
)

// newTestLogger 创建用于测试的日志实例
func newTestLogger(t *testing.T) *logger.Logger {
	t.Helper()
	cfg := &config.LoggerConfig{
		Level:  "info",
		Format: "text",
		Output: "stdout",
	}
	log, err := logger.NewLogger(cfg)
	if err != nil {
		t.Fatalf("创建测试日志失败: %v", err)
	}
	return log
}

// TestNewTracerProvider 测试追踪提供者创建
func TestNewTracerProvider(t *testing.T) {
	log := newTestLogger(t)
	cfg := &config.TracerConfig{
		Enabled:     true,
		JaegerURL:   "http://jaeger:14268/api/traces",
		ServiceName: "ceph-exporter-test",
		SampleRate:  1.0,
	}

	tp, err := NewTracerProvider(cfg, log)
	if err != nil {
		t.Fatalf("创建追踪提供者失败: %v", err)
	}
	if tp == nil {
		t.Fatal("追踪提供者为 nil")
	}
}

// TestTracerProvider_Shutdown 测试追踪提供者关闭
func TestTracerProvider_Shutdown(t *testing.T) {
	log := newTestLogger(t)
	cfg := &config.TracerConfig{
		Enabled:     true,
		JaegerURL:   "http://jaeger:14268/api/traces",
		ServiceName: "ceph-exporter-test",
		SampleRate:  1.0,
	}

	tp, _ := NewTracerProvider(cfg, log)

	// 关闭不应该返回错误
	ctx := context.Background()
	if err := tp.Shutdown(ctx); err != nil {
		t.Errorf("关闭追踪提供者失败: %v", err)
	}
}

// TestStartSpan 测试创建 Span
func TestStartSpan(t *testing.T) {
	ctx := context.Background()

	// 创建 Span
	newCtx, span := StartSpan(ctx, "test.operation")

	// 验证返回值不为 nil
	if newCtx == nil {
		t.Fatal("StartSpan 返回的上下文为 nil")
	}
	if span == nil {
		t.Fatal("StartSpan 返回的 Span 为 nil")
	}

	// End 不应该 panic
	span.End()
}

// TestSpan_SetAttributes 测试设置 Span 属性（不应 panic）
func TestSpan_SetAttributes(t *testing.T) {
	_, span := StartSpan(context.Background(), "test")

	// 各种参数调用都不应该 panic
	span.SetAttributes()
	span.SetAttributes(attribute.String("key", "value"))
	span.SetAttributes(
		attribute.String("key1", "value1"),
		attribute.String("key2", "value2"),
	)

	span.End()
}

// TestGetTraceID 测试获取追踪 ID
func TestGetTraceID(t *testing.T) {
	ctx := context.Background()

	// Phase 1 应该返回空字符串
	traceID := GetTraceID(ctx)
	if traceID != "" {
		t.Errorf("Phase 1 GetTraceID 期望空字符串，实际 '%s'", traceID)
	}
}

// TestGetSpanID 测试获取 Span ID
func TestGetSpanID(t *testing.T) {
	ctx := context.Background()

	// Phase 1 应该返回空字符串
	spanID := GetSpanID(ctx)
	if spanID != "" {
		t.Errorf("Phase 1 GetSpanID 期望空字符串，实际 '%s'", spanID)
	}
}

// TestStartSpan_ContextPropagation 测试上下文传播
func TestStartSpan_ContextPropagation(t *testing.T) {
	// 创建带值的上下文
	type ctxKey string
	parentCtx := context.WithValue(context.Background(), ctxKey("test"), "value")

	// StartSpan 应该保留父上下文中的值
	childCtx, span := StartSpan(parentCtx, "child.operation")
	defer span.End()

	// 验证父上下文的值在子上下文中仍然可用
	if val := childCtx.Value(ctxKey("test")); val != "value" {
		t.Errorf("子上下文丢失了父上下文的值: 期望 'value'，实际 '%v'", val)
	}
}
