// =============================================================================
// Tracer 模块单元测试
// =============================================================================
// 测试分布式追踪模块的功能，包括:
//   - TracerProvider 创建（启用/禁用状态）
//   - Span 创建和管理
//   - TraceID 和 SpanID 提取
//   - 属性设置（字符串、整数、布尔值）
//   - Span 状态设置（成功/错误）
//
// 注意:
//
//	这些测试不依赖真实的 Jaeger 服务，主要验证 API 的正确性。
//	当追踪未启用时，所有操作都应该是安全的空操作（no-op）。
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

// TestNewTracerProvider_Disabled 测试追踪禁用时创建 TracerProvider
// 验证:
//   - 不会返回错误
//   - 返回的 TracerProvider 不为 nil（但内部 tp 为 nil）
//   - Shutdown 调用不会报错
func TestNewTracerProvider_Disabled(t *testing.T) {
	cfg := &config.TracerConfig{
		Enabled: false,
	}
	log, _ := logger.NewLogger(&config.LoggerConfig{
		Level:  "info",
		Format: "json",
	})

	tp, err := NewTracerProvider(cfg, log)
	if err != nil {
		t.Fatalf("创建 TracerProvider 失败: %v", err)
	}
	if tp == nil {
		t.Fatal("TracerProvider 不应为 nil")
	}

	ctx := context.Background()
	if err := tp.Shutdown(ctx); err != nil {
		t.Fatalf("关闭 TracerProvider 失败: %v", err)
	}
}

// TestStartSpan 测试创建 Span
// 验证:
//   - 返回的 context 不为 nil
//   - 返回的 span 不为 nil
//   - span.End() 调用不会 panic
func TestStartSpan(t *testing.T) {
	ctx := context.Background()
	newCtx, span := StartSpan(ctx, "test-span")

	if newCtx == nil {
		t.Fatal("返回的 context 不应为 nil")
	}
	if span == nil {
		t.Fatal("返回的 span 不应为 nil")
	}

	span.End()
}

// TestGetTraceID 测试从上下文中提取 TraceID
// 当没有启用追踪后端时，可能返回空字符串或全零的 TraceID
func TestGetTraceID(t *testing.T) {
	ctx := context.Background()
	newCtx, span := StartSpan(ctx, "test-span")
	defer span.End()

	traceID := GetTraceID(newCtx)
	// 当没有启用追踪时，可能返回空字符串或有效的 trace ID
	t.Logf("Trace ID: %s", traceID)
}

// TestGetSpanID 测试从上下文中提取 SpanID
func TestGetSpanID(t *testing.T) {
	ctx := context.Background()
	newCtx, span := StartSpan(ctx, "test-span")
	defer span.End()

	spanID := GetSpanID(newCtx)
	t.Logf("Span ID: %s", spanID)
}

// TestSetAttributes 测试为 Span 设置属性
// 验证设置属性不会 panic（即使追踪未启用）
func TestSetAttributes(t *testing.T) {
	ctx := context.Background()
	newCtx, span := StartSpan(ctx, "test-span")
	defer span.End()

	SetAttributes(newCtx,
		attribute.String("key1", "value1"),
		attribute.Int("key2", 42),
	)
}

// TestSetSpanStatus 测试设置 Span 状态
// 验证成功和错误状态的设置都不会 panic
func TestSetSpanStatus(t *testing.T) {
	ctx := context.Background()
	newCtx, span := StartSpan(ctx, "test-span")
	defer span.End()

	// 测试成功状态
	SetSpanStatus(newCtx, StatusOK, "")

	// 测试错误状态
	SetSpanStatus(newCtx, StatusError, "test error")
}

// TestStringAttr 测试创建字符串类型属性
func TestStringAttr(t *testing.T) {
	attr := StringAttr("key", "value")
	if attr.Key != "key" {
		t.Errorf("Expected key 'key', got '%s'", attr.Key)
	}
}

// TestIntAttr 测试创建整数类型属性
func TestIntAttr(t *testing.T) {
	attr := IntAttr("count", 42)
	if attr.Key != "count" {
		t.Errorf("Expected key 'count', got '%s'", attr.Key)
	}
}

// TestBoolAttr 测试创建布尔类型属性
func TestBoolAttr(t *testing.T) {
	attr := BoolAttr("enabled", true)
	if attr.Key != "enabled" {
		t.Errorf("Expected key 'enabled', got '%s'", attr.Key)
	}
}
