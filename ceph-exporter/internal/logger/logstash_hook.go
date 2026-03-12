// =============================================================================
// Logstash Hook 实现
// =============================================================================
// 提供将日志直接推送到 Logstash 的功能。
// 支持两种推送模式:
//   - TCP: 使用 TCP 连接推送日志（可靠，但需要保持连接）
//   - UDP: 使用 UDP 推送日志（快速，但可能丢失）
//
// 特性:
//   - 异步推送，不阻塞日志调用
//   - 自动重连机制（TCP 模式）
//   - 缓冲队列，防止日志丢失
//   - 优雅关闭，确保缓冲区日志发送完成
//
// 使用示例:
//
//	hook, err := NewLogstashHook("tcp", "logstash:5044", "ceph-exporter")
//	if err != nil {
//	    return err
//	}
//	logger.AddHook(hook)
//	defer hook.Close()
//
// =============================================================================
package logger

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// LogstashHook Logstash Hook 结构
// 实现 logrus.Hook 接口，将日志推送到 Logstash
type LogstashHook struct {
	protocol    string         // 协议类型: tcp 或 udp
	address     string         // Logstash 地址，格式: host:port
	serviceName string         // 服务名称，用于标识日志来源
	conn        net.Conn       // 网络连接（TCP/UDP）
	mu          sync.Mutex     // 保护连接的互斥锁
	buffer      chan []byte    // 异步发送缓冲队列
	closed      bool           // 是否已关闭
	wg          sync.WaitGroup // 等待 goroutine 完成
	reconnect   bool           // 是否启用自动重连（仅 TCP）
}

// NewLogstashHook 创建 Logstash Hook 实例
//
// 参数:
//   - protocol: 协议类型，"tcp" 或 "udp"
//   - address: Logstash 地址，格式: "host:port"（如 "logstash:5044"）
//   - serviceName: 服务名称，用于在日志中标识来源
//
// 返回:
//   - *LogstashHook: Hook 实例
//   - error: 创建过程中的错误（无效协议、连接失败等）
func NewLogstashHook(protocol, address, serviceName string) (*LogstashHook, error) {
	// 验证协议类型
	if protocol != "tcp" && protocol != "udp" {
		return nil, fmt.Errorf("不支持的协议类型: %s，仅支持 tcp 或 udp", protocol)
	}

	hook := &LogstashHook{
		protocol:    protocol,
		address:     address,
		serviceName: serviceName,
		buffer:      make(chan []byte, 1000), // 缓冲 1000 条日志
		reconnect:   protocol == "tcp",       // TCP 模式启用自动重连
	}

	// 建立初始连接
	if err := hook.connect(); err != nil {
		return nil, fmt.Errorf("连接 Logstash 失败: %w", err)
	}

	// 启动异步发送 goroutine
	hook.wg.Add(1)
	go hook.sender()

	return hook, nil
}

// connect 建立到 Logstash 的网络连接
func (h *LogstashHook) connect() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 如果已有连接，先关闭
	if h.conn != nil {
		h.conn.Close()
	}

	// 建立新连接
	conn, err := net.DialTimeout(h.protocol, h.address, 5*time.Second)
	if err != nil {
		return err
	}

	h.conn = conn
	return nil
}

// sender 异步发送日志的 goroutine
// 从缓冲队列中读取日志并发送到 Logstash
func (h *LogstashHook) sender() {
	defer h.wg.Done()

	for data := range h.buffer {
		// 尝试发送，失败时重试
		if err := h.send(data); err != nil {
			// TCP 模式下尝试重连
			if h.reconnect {
				if reconnectErr := h.connect(); reconnectErr == nil {
					// 重连成功，再次尝试发送
					_ = h.send(data)
				}
			}
			// UDP 模式或重连失败，日志丢失（避免阻塞）
		}
	}
}

// send 发送单条日志到 Logstash
func (h *LogstashHook) send(data []byte) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.conn == nil {
		return fmt.Errorf("连接未建立")
	}

	// 设置写入超时（防止阻塞）
	if err := h.conn.SetWriteDeadline(time.Now().Add(2 * time.Second)); err != nil {
		return err
	}

	// 发送数据（JSON 格式 + 换行符）
	_, err := h.conn.Write(append(data, '\n'))
	return err
}

// Levels 返回 Hook 关注的日志级别
// 实现 logrus.Hook 接口
func (h *LogstashHook) Levels() []logrus.Level {
	// 监听所有级别的日志
	return logrus.AllLevels
}

// Fire 处理日志条目
// 实现 logrus.Hook 接口，将日志转换为 JSON 并推送到 Logstash
func (h *LogstashHook) Fire(entry *logrus.Entry) error {
	// 如果已关闭，忽略日志
	if h.closed {
		return nil
	}

	// 构建 Logstash 格式的日志数据
	logData := map[string]interface{}{
		"@timestamp": entry.Time.Format(time.RFC3339Nano),
		"level":      entry.Level.String(),
		"message":    entry.Message,
		"service":    h.serviceName,
	}

	// 添加所有字段（component, trace_id, span_id 等）
	for key, value := range entry.Data {
		logData[key] = value
	}

	// 序列化为 JSON
	data, err := json.Marshal(logData)
	if err != nil {
		return fmt.Errorf("序列化日志失败: %w", err)
	}

	// 非阻塞方式放入缓冲队列
	select {
	case h.buffer <- data:
		// 成功放入队列
	default:
		// 队列已满，丢弃日志（避免阻塞应用）
		return fmt.Errorf("日志缓冲队列已满，日志被丢弃")
	}

	return nil
}

// Close 关闭 Hook，释放资源
// 确保缓冲区中的日志全部发送完成
func (h *LogstashHook) Close() error {
	h.mu.Lock()
	if h.closed {
		h.mu.Unlock()
		return nil
	}
	h.closed = true
	h.mu.Unlock()

	// 关闭缓冲队列，触发 sender goroutine 退出
	close(h.buffer)

	// 等待 sender goroutine 完成
	h.wg.Wait()

	// 关闭网络连接
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.conn != nil {
		return h.conn.Close()
	}

	return nil
}
