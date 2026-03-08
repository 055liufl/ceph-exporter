// =============================================================================
// HTTP 服务器
// =============================================================================
// 提供 ceph-exporter 的 HTTP 服务，包括:
//   - /metrics  - Prometheus 指标端点（Phase 2 完整实现采集器）
//   - /health   - 健康检查端点（Kubernetes liveness probe）
//   - /ready    - 就绪检查端点（Kubernetes readiness probe）
//   - /         - 首页，显示可用端点列表
//
// 特性:
//   - 支持 TLS/HTTPS（配置 tls_cert_file 和 tls_key_file）
//   - 支持优雅关闭（Graceful Shutdown）
//   - 集成追踪中间件（Phase 3 完整实现）
//   - 可配置的读写超时
//
// 使用示例:
//
//	srv := server.NewServer(&cfg.Server, registry, cephClient, log, tp)
//	go srv.Start()
//	// ... 等待退出信号 ...
//	srv.Shutdown(ctx)
//
// =============================================================================
package server

import (
	"context"
	"fmt"
	"net/http"

	"ceph-exporter/internal/ceph"
	"ceph-exporter/internal/config"
	"ceph-exporter/internal/logger"
	"ceph-exporter/internal/tracer"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Server HTTP 服务器结构
// 封装了 net/http.Server，集成了 Prometheus Registry 和追踪系统
type Server struct {
	config     *config.ServerConfig   // HTTP 服务器配置
	registry   *prometheus.Registry   // Prometheus 指标注册表
	log        *logger.Logger         // 日志实例
	tp         *tracer.TracerProvider // 追踪提供者（可为 nil）
	cephClient *ceph.Client           // Ceph 客户端，用于就绪检查
	server     *http.Server           // 底层 HTTP 服务器实例
}

// NewServer 创建 HTTP 服务器实例
// 注意: 创建后需要调用 Start() 启动服务
//
// 参数:
//   - cfg: HTTP 服务器配置（地址、端口、超时、TLS 等）
//   - registry: Prometheus 指标注册表，用于 /metrics 端点
//   - cephClient: Ceph 客户端实例，用于就绪检查
//   - log: 日志实例
//   - tp: 追踪提供者（可为 nil，表示不启用追踪）
//
// 返回:
//   - *Server: HTTP 服务器实例
func NewServer(cfg *config.ServerConfig, registry *prometheus.Registry, cephClient *ceph.Client, log *logger.Logger, tp *tracer.TracerProvider) *Server {
	return &Server{
		config:     cfg,
		registry:   registry,
		cephClient: cephClient,
		log:        log,
		tp:         tp,
	}
}

// Start 启动 HTTP 服务器
// 注册所有路由并开始监听请求。此方法会阻塞直到服务器关闭。
// 建议在 goroutine 中调用。
//
// 返回:
//   - error: 启动或运行过程中的错误（端口被占用、TLS 配置错误等）
func (s *Server) Start() error {
	// 创建路由多路复用器
	mux := http.NewServeMux()

	// ----- 注册路由 -----

	// /metrics - Prometheus 指标端点
	// 使用追踪中间件包装，记录每次指标采集的追踪信息
	mux.Handle("/metrics", s.tracingMiddleware(
		promhttp.HandlerFor(s.registry, promhttp.HandlerOpts{
			// 使用默认选项，错误会记录到日志
		}),
	))

	// /health - 健康检查端点（Kubernetes liveness probe）
	// 只要服务进程存活就返回 200 OK
	mux.HandleFunc("/health", s.healthHandler)

	// /ready - 就绪检查端点（Kubernetes readiness probe）
	// 检查服务是否准备好接收请求（如 Ceph 连接是否正常）
	mux.HandleFunc("/ready", s.readyHandler)

	// / - 首页，显示可用端点列表
	mux.HandleFunc("/", s.indexHandler)

	// 构建监听地址
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)

	// 创建 HTTP 服务器，配置超时参数
	s.server = &http.Server{
		Addr:         addr,                  // 监听地址，如 "0.0.0.0:9128"
		Handler:      mux,                   // 路由处理器
		ReadTimeout:  s.config.ReadTimeout,  // 读取请求的超时时间
		WriteTimeout: s.config.WriteTimeout, // 写入响应的超时时间
	}

	s.log.WithComponent("http-server").Infof("HTTP 服务器启动在 %s", addr)

	// 如果配置了 TLS 证书和密钥，使用 HTTPS
	if s.config.TLSCertFile != "" && s.config.TLSKeyFile != "" {
		s.log.WithComponent("http-server").Info("TLS 已启用")
		return s.server.ListenAndServeTLS(s.config.TLSCertFile, s.config.TLSKeyFile)
	}

	// 否则使用 HTTP
	return s.server.ListenAndServe()
}

// Shutdown 优雅关闭 HTTP 服务器
// 停止接受新连接，等待已有请求处理完成后关闭。
// 如果超过 context 的截止时间，强制关闭。
//
// 参数:
//   - ctx: 上下文，用于控制关闭的超时时间
//
// 返回:
//   - error: 关闭过程中的错误
func (s *Server) Shutdown(ctx context.Context) error {
	s.log.WithComponent("http-server").Info("正在优雅关闭 HTTP 服务器...")
	return s.server.Shutdown(ctx)
}

// tracingMiddleware 追踪中间件
// 为每个 HTTP 请求创建追踪 Span，记录请求和响应信息。
// Phase 1 中追踪为占位实现，不会产生实际追踪数据。
//
// 参数:
//   - next: 下一个 HTTP 处理器
//
// 返回:
//   - http.Handler: 包装了追踪逻辑的 HTTP 处理器
func (s *Server) tracingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 如果追踪未启用，直接调用下一个 Handler
		if s.tp == nil {
			next.ServeHTTP(w, r)
			return
		}

		// 创建追踪 Span（Phase 1 为空操作）
		ctx, span := tracer.StartSpan(r.Context(), "http.request")
		defer span.End()

		// 将带有 Span 信息的上下文注入到请求中
		r = r.WithContext(ctx)

		// 记录追踪 ID 到日志（Phase 1 中 traceID 为空字符串）
		traceID := tracer.GetTraceID(ctx)
		if traceID != "" {
			s.log.WithTraceID(traceID).Debugf("处理请求: %s %s", r.Method, r.URL.Path)
		}

		// 调用下一个 Handler 处理实际请求
		next.ServeHTTP(w, r)
	})
}

// healthHandler 健康检查接口
// 用于 Kubernetes liveness probe，只要服务进程存活就返回 200 OK。
// 不检查外部依赖（如 Ceph 连接），因为 liveness 只关心进程是否存活。
func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("OK")); err != nil {
		s.log.WithComponent("http-server").Warnf("Failed to write health response: %v", err)
	}
}

// readyHandler 就绪检查接口
// 用于 Kubernetes readiness probe，检查服务是否准备好接收请求。
// 通过执行 Ceph 健康检查验证集群连接是否正常。
func (s *Server) readyHandler(w http.ResponseWriter, r *http.Request) {
	if s.cephClient == nil {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("Ready")); err != nil {
			s.log.WithComponent("http-server").Warnf("Failed to write ready response: %v", err)
		}
		return
	}
	if err := s.cephClient.HealthCheck(); err != nil {
		s.log.WithComponent("http-server").Warnf("就绪检查失败: %v", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		if _, err := w.Write([]byte("Not Ready")); err != nil {
			s.log.WithComponent("http-server").Warnf("Failed to write not ready response: %v", err)
		}
		return
	}
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("Ready")); err != nil {
		s.log.WithComponent("http-server").Warnf("Failed to write ready response: %v", err)
	}
}

// indexHandler 首页处理器
// 显示 ceph-exporter 的基本信息和可用端点列表。
// 提供简单的 HTML 页面，方便用户通过浏览器访问时了解服务状态。
func (s *Server) indexHandler(w http.ResponseWriter, r *http.Request) {
	// 简洁的 HTML 页面，列出所有可用端点
	html := `<!DOCTYPE html>
<html>
<head>
    <title>Ceph Exporter</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        h1 { color: #333; }
        ul { list-style-type: none; padding: 0; }
        li { margin: 10px 0; }
        a { color: #0066cc; text-decoration: none; }
        a:hover { text-decoration: underline; }
    </style>
</head>
<body>
    <h1>Ceph Exporter</h1>
    <p>Prometheus exporter for Ceph cluster metrics</p>
    <h2>可用端点:</h2>
    <ul>
        <li><a href="/metrics">/metrics</a> - Prometheus 指标</li>
        <li><a href="/health">/health</a> - 健康检查</li>
        <li><a href="/ready">/ready</a> - 就绪检查</li>
    </ul>
</body>
</html>`
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := w.Write([]byte(html)); err != nil {
		s.log.WithComponent("http-server").Warnf("Failed to write index response: %v", err)
	}
}
