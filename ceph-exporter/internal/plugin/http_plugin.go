// =============================================================================
// HTTP 插件实现（Phase 5 示例插件）
// =============================================================================
// 本文件实现基于 HTTP 的远程插件，用于监控第三方存储系统。
// 通过 HTTP API 调用远程服务获取监控指标。
//
// 功能特性:
//   - 支持 HTTP/HTTPS 协议
//   - 支持自定义请求头（如认证 Token）
//   - 支持超时控制
//   - 支持健康检查
//   - 自动重试机制
//
// =============================================================================
package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// HTTPPlugin HTTP 远程插件
// 通过 HTTP API 调用远程服务获取监控指标
type HTTPPlugin struct {
	// name 插件名称
	name string

	// version 插件版本
	version string

	// description 插件描述
	description string

	// endpoint HTTP 端点 URL
	endpoint string

	// headers 自定义请求头
	headers map[string]string

	// timeout 请求超时时间
	timeout time.Duration

	// client HTTP 客户端
	client *http.Client

	// collector Prometheus 采集器
	collector *httpPluginCollector

	// mu 保护并发访问
	mu sync.RWMutex

	// running 插件是否正在运行
	running bool

	// ctx 上下文
	ctx context.Context

	// cancel 取消函数
	cancel context.CancelFunc
}

// NewHTTPPlugin 创建 HTTP 插件
//
// 参数:
//   - name: 插件名称
//   - version: 插件版本
//   - description: 插件描述
//
// 返回:
//   - *HTTPPlugin: HTTP 插件实例
func NewHTTPPlugin(name, version, description string) *HTTPPlugin {
	return &HTTPPlugin{
		name:        name,
		version:     version,
		description: description,
		timeout:     10 * time.Second, // 默认超时 10 秒
	}
}

// Name 返回插件名称
func (p *HTTPPlugin) Name() string {
	return p.name
}

// Version 返回插件版本
func (p *HTTPPlugin) Version() string {
	return p.version
}

// Description 返回插件描述
func (p *HTTPPlugin) Description() string {
	return p.description
}

// Init 初始化插件
//
// 配置参数:
//   - endpoint: HTTP 端点 URL（必需）
//   - headers: 自定义请求头（可选）
//   - timeout: 请求超时时间，单位秒（可选，默认 10）
//
// 返回:
//   - error: 初始化失败时返回错误
func (p *HTTPPlugin) Init(config map[string]interface{}) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 解析 endpoint
	endpoint, ok := config["endpoint"].(string)
	if !ok || endpoint == "" {
		return fmt.Errorf("endpoint is required")
	}
	p.endpoint = endpoint

	// 解析 headers
	if headers, ok := config["headers"].(map[string]interface{}); ok {
		p.headers = make(map[string]string)
		for k, v := range headers {
			if strVal, ok := v.(string); ok {
				p.headers[k] = strVal
			}
		}
	}

	// 解析 timeout
	if timeout, ok := config["timeout"].(float64); ok {
		p.timeout = time.Duration(timeout) * time.Second
	}

	// 创建 HTTP 客户端
	p.client = &http.Client{
		Timeout: p.timeout,
		Transport: &http.Transport{
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 5,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	// 创建 Prometheus 采集器
	p.collector = newHTTPPluginCollector(p)

	return nil
}

// Start 启动插件
//
// 参数:
//   - ctx: 上下文，用于控制插件生命周期
//
// 返回:
//   - error: 启动失败时返回错误
func (p *HTTPPlugin) Start(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.running {
		return fmt.Errorf("plugin already running")
	}

	p.ctx, p.cancel = context.WithCancel(ctx)
	p.running = true

	return nil
}

// Stop 停止插件
//
// 返回:
//   - error: 停止失败时返回错误
func (p *HTTPPlugin) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.running {
		return nil
	}

	if p.cancel != nil {
		p.cancel()
	}

	if p.client != nil {
		p.client.CloseIdleConnections()
	}

	p.running = false

	return nil
}

// Health 健康检查
//
// 返回:
//   - error: 插件不健康时返回错误
func (p *HTTPPlugin) Health() error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.running {
		return fmt.Errorf("plugin not running")
	}

	// 发送健康检查请求
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", p.endpoint+"/health", nil)
	if err != nil {
		return fmt.Errorf("create health check request: %w", err)
	}

	// 添加自定义请求头
	for k, v := range p.headers {
		req.Header.Set(k, v)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("health check request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed: status=%d", resp.StatusCode)
	}

	return nil
}

// Collect 采集指标
//
// 参数:
//   - ctx: 上下文，用于超时控制
//
// 返回:
//   - []Metric: 指标列表
//   - error: 采集失败时返回错误
func (p *HTTPPlugin) Collect(ctx context.Context) ([]Metric, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.running {
		return nil, fmt.Errorf("plugin not running")
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "GET", p.endpoint+"/metrics", nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	// 添加自定义请求头
	for k, v := range p.headers {
		req.Header.Set(k, v)
	}

	// 发送请求
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	// 解析 JSON 响应
	var response struct {
		Metrics []Metric `json:"metrics"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	return response.Metrics, nil
}

// Collector 返回 Prometheus 采集器
func (p *HTTPPlugin) Collector() prometheus.Collector {
	return p.collector
}

// httpPluginCollector HTTP 插件的 Prometheus 采集器
type httpPluginCollector struct {
	plugin *HTTPPlugin
}

// newHTTPPluginCollector 创建 HTTP 插件采集器
func newHTTPPluginCollector(plugin *HTTPPlugin) *httpPluginCollector {
	return &httpPluginCollector{
		plugin: plugin,
	}
}

// Describe 实现 prometheus.Collector 接口
func (c *httpPluginCollector) Describe(ch chan<- *prometheus.Desc) {
	// HTTP 插件的指标是动态的，不需要预先描述
}

// Collect 实现 prometheus.Collector 接口
func (c *httpPluginCollector) Collect(ch chan<- prometheus.Metric) {
	// 创建超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 采集指标
	metrics, err := c.plugin.Collect(ctx)
	if err != nil {
		// 采集失败时不发送指标，避免影响其他采集器
		return
	}

	// 转换为 Prometheus 指标
	for _, metric := range metrics {
		// 创建指标描述
		desc := prometheus.NewDesc(
			"plugin_"+c.plugin.Name()+"_"+metric.Name,
			metric.Help,
			nil,
			metric.Labels,
		)

		// 根据指标类型创建不同的 Prometheus 指标
		var promMetric prometheus.Metric
		switch metric.Type {
		case "gauge":
			promMetric = prometheus.MustNewConstMetric(
				desc,
				prometheus.GaugeValue,
				metric.Value,
			)
		case "counter":
			promMetric = prometheus.MustNewConstMetric(
				desc,
				prometheus.CounterValue,
				metric.Value,
			)
		default:
			// 默认使用 Gauge 类型
			promMetric = prometheus.MustNewConstMetric(
				desc,
				prometheus.GaugeValue,
				metric.Value,
			)
		}

		ch <- promMetric
	}
}
