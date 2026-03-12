package tracer

import (
	"context"
	"testing"

	"ceph-exporter/internal/config"
	"ceph-exporter/internal/logger"

	"go.opentelemetry.io/otel/attribute"
)

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

func TestGetTraceID(t *testing.T) {
	ctx := context.Background()
	newCtx, span := StartSpan(ctx, "test-span")
	defer span.End()

	traceID := GetTraceID(newCtx)
	// 当没有启用追踪时，可能返回空字符串或有效的 trace ID
	t.Logf("Trace ID: %s", traceID)
}

func TestGetSpanID(t *testing.T) {
	ctx := context.Background()
	newCtx, span := StartSpan(ctx, "test-span")
	defer span.End()

	spanID := GetSpanID(newCtx)
	t.Logf("Span ID: %s", spanID)
}

func TestSetAttributes(t *testing.T) {
	ctx := context.Background()
	newCtx, span := StartSpan(ctx, "test-span")
	defer span.End()

	SetAttributes(newCtx,
		attribute.String("key1", "value1"),
		attribute.Int("key2", 42),
	)
}

func TestSetSpanStatus(t *testing.T) {
	ctx := context.Background()
	newCtx, span := StartSpan(ctx, "test-span")
	defer span.End()

	// 测试成功状态
	SetSpanStatus(newCtx, StatusOK, "")

	// 测试错误状态
	SetSpanStatus(newCtx, StatusError, "test error")
}

func TestStringAttr(t *testing.T) {
	attr := StringAttr("key", "value")
	if attr.Key != "key" {
		t.Errorf("Expected key 'key', got '%s'", attr.Key)
	}
}

func TestIntAttr(t *testing.T) {
	attr := IntAttr("count", 42)
	if attr.Key != "count" {
		t.Errorf("Expected key 'count', got '%s'", attr.Key)
	}
}

func TestBoolAttr(t *testing.T) {
	attr := BoolAttr("enabled", true)
	if attr.Key != "enabled" {
		t.Errorf("Expected key 'enabled', got '%s'", attr.Key)
	}
}
