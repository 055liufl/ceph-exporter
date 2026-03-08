// =============================================================================
// HTTP 服务器单元测试
// =============================================================================
// 测试覆盖:
//   - 服务器创建
//   - /health 端点
//   - /ready 端点
//   - / 首页端点
//   - /metrics 端点（空 Registry）
//
// =============================================================================
package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ceph-exporter/internal/config"
	"ceph-exporter/internal/logger"

	"github.com/prometheus/client_golang/prometheus"
)

// newTestServer 创建用于测试的服务器实例
func newTestServer(t *testing.T) *Server {
	t.Helper()

	// 创建日志实例
	logCfg := &config.LoggerConfig{
		Level:  "info",
		Format: "text",
		Output: "stdout",
	}
	log, err := logger.NewLogger(logCfg)
	if err != nil {
		t.Fatalf("创建测试日志失败: %v", err)
	}

	// 创建服务器配置
	srvCfg := &config.ServerConfig{
		Host: "127.0.0.1",
		Port: 9128,
	}

	// 创建空的 Prometheus Registry
	registry := prometheus.NewRegistry()

	return NewServer(srvCfg, registry, nil, log, nil)
}

// TestNewServer 测试服务器创建
func TestNewServer(t *testing.T) {
	srv := newTestServer(t)
	if srv == nil {
		t.Fatal("服务器实例为 nil")
	}
	if srv.config == nil {
		t.Error("服务器配置为 nil")
	}
	if srv.registry == nil {
		t.Error("Prometheus Registry 为 nil")
	}
	if srv.log == nil {
		t.Error("日志实例为 nil")
	}
}

// TestHealthHandler 测试健康检查端点
func TestHealthHandler(t *testing.T) {
	srv := newTestServer(t)

	// 创建测试请求
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	// 调用处理器
	srv.healthHandler(rec, req)

	// 验证响应状态码
	if rec.Code != http.StatusOK {
		t.Errorf("健康检查状态码期望 %d，实际 %d", http.StatusOK, rec.Code)
	}

	// 验证响应内容
	if rec.Body.String() != "OK" {
		t.Errorf("健康检查响应期望 'OK'，实际 '%s'", rec.Body.String())
	}
}

// TestReadyHandler 测试就绪检查端点
func TestReadyHandler(t *testing.T) {
	srv := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec := httptest.NewRecorder()

	srv.readyHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("就绪检查状态码期望 %d，实际 %d", http.StatusOK, rec.Code)
	}

	if rec.Body.String() != "Ready" {
		t.Errorf("就绪检查响应期望 'Ready'，实际 '%s'", rec.Body.String())
	}
}

// TestIndexHandler 测试首页端点
func TestIndexHandler(t *testing.T) {
	srv := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	srv.indexHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("首页状态码期望 %d，实际 %d", http.StatusOK, rec.Code)
	}

	// 验证 Content-Type
	contentType := rec.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		t.Errorf("首页 Content-Type 期望包含 'text/html'，实际 '%s'", contentType)
	}

	// 验证页面包含关键内容
	body := rec.Body.String()
	expectedContents := []string{
		"Ceph Exporter",
		"/metrics",
		"/health",
		"/ready",
	}
	for _, expected := range expectedContents {
		if !strings.Contains(body, expected) {
			t.Errorf("首页内容缺少 '%s'", expected)
		}
	}
}

// TestTracingMiddleware_NoTracer 测试追踪中间件（追踪未启用）
func TestTracingMiddleware_NoTracer(t *testing.T) {
	srv := newTestServer(t)
	// srv.tp 为 nil，追踪未启用

	// 创建一个简单的测试 Handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test"))
	})

	// 用追踪中间件包装
	wrapped := srv.tracingMiddleware(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	// 执行请求
	wrapped.ServeHTTP(rec, req)

	// 验证请求正常通过
	if rec.Code != http.StatusOK {
		t.Errorf("中间件状态码期望 %d，实际 %d", http.StatusOK, rec.Code)
	}
	if rec.Body.String() != "test" {
		t.Errorf("中间件响应期望 'test'，实际 '%s'", rec.Body.String())
	}
}

// TestMetricsHandler 测试 /metrics 端点（空 Registry）
func TestMetricsHandler(t *testing.T) {
	srv := newTestServer(t)

	// 创建路由
	mux := http.NewServeMux()
	mux.Handle("/metrics", srv.tracingMiddleware(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("# HELP test_metric A test metric\n"))
		}),
	))

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("/metrics 状态码期望 %d，实际 %d", http.StatusOK, rec.Code)
	}
}
