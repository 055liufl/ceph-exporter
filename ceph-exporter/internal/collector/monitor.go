// =============================================================================
// Monitor Collector - Monitor 指标采集器
// =============================================================================
// 采集每个 Ceph Monitor 的详细指标，包括:
//   - 仲裁状态（是否在 quorum 中）
//   - 存储大小（Monitor 数据库占用空间）
//   - 时钟偏移（与集群时钟的偏差）
//   - 延迟（Monitor 响应延迟）
//
// 数据来源:
//
//	通过 "ceph mon dump -f json" 命令获取 Monitor 信息 JSON
//
// 标签:
//
//	monitor: Monitor 名称（如 mon.a, mon.b）
//
// 指标列表:
//
//	ceph_monitor_in_quorum       - Monitor 是否在仲裁中（1=是, 0=否）
//	ceph_monitor_store_bytes     - Monitor 数据库存储大小（字节）
//	ceph_monitor_clock_skew_sec  - Monitor 时钟偏移（秒）
//	ceph_monitor_latency_sec     - Monitor 响应延迟（秒）
//
// =============================================================================
package collector

import (
	"ceph-exporter/internal/ceph"
	"ceph-exporter/internal/logger"

	"github.com/prometheus/client_golang/prometheus"
)

// MonitorCollector Monitor 采集器
type MonitorCollector struct {
	client *ceph.Client
	log    *logger.Logger

	inQuorum   *prometheus.Desc
	storeBytes *prometheus.Desc
	clockSkew  *prometheus.Desc
	latency    *prometheus.Desc
}

// NewMonitorCollector 创建 Monitor 采集器实例
// 初始化所有 Monitor 相关的 Prometheus 指标描述符
//
// 参数:
//   - client: Ceph 客户端实例，用于执行命令获取 Monitor 数据
//   - log: 日志实例，用于记录采集过程中的信息和错误
//
// 返回:
//   - *MonitorCollector: 初始化完成的 Monitor 采集器实例
func NewMonitorCollector(client *ceph.Client, log *logger.Logger) *MonitorCollector {
	// 所有 Monitor 指标都带有 monitor 标签，用于区分不同的 Monitor
	monLabels := []string{"monitor"}

	return &MonitorCollector{
		client: client,
		log:    log,

		inQuorum: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "monitor", "in_quorum"),
			"Monitor 是否在仲裁中（1=是, 0=否）",
			monLabels, nil,
		),
		storeBytes: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "monitor", "store_bytes"),
			"Monitor 数据库存储大小（字节）",
			monLabels, nil,
		),
		clockSkew: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "monitor", "clock_skew_sec"),
			"Monitor 时钟偏移（秒）",
			monLabels, nil,
		),
		latency: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "monitor", "latency_sec"),
			"Monitor 响应延迟（秒）",
			monLabels, nil,
		),
	}
}

// Describe 向 Prometheus 注册本采集器提供的所有指标描述符
// 实现 prometheus.Collector 接口的 Describe 方法
// Prometheus 在注册采集器时会调用此方法，获取采集器提供的所有指标定义
//
// 参数:
//   - ch: 指标描述符通道，用于发送指标描述符到 Prometheus
func (c *MonitorCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.inQuorum
	ch <- c.storeBytes
	ch <- c.clockSkew
	ch <- c.latency
}

// Collect 执行 Monitor 指标采集
// 实现 prometheus.Collector 接口的 Collect 方法
// Prometheus 定期调用此方法采集最新的指标数据
// 遍历所有 Monitor，为每个 Monitor 生成一组带 monitor 标签的指标
//
// 注意事项:
//   - InQuorum 是布尔值，需要转换为浮点数（1.0 或 0.0）
//   - LatencyMs 从毫秒转换为秒，符合 Prometheus 时间单位惯例
//
// 采集流程:
//  1. 创建带超时的上下文
//  2. 调用 Ceph 客户端获取所有 Monitor 的统计数据
//  3. 遍历每个 Monitor，生成对应的 Prometheus 指标
//  4. 通过 channel 发送指标到 Prometheus
//
// 参数:
//   - ch: 指标通道，用于发送采集到的指标数据到 Prometheus
func (c *MonitorCollector) Collect(ch chan<- prometheus.Metric) {
	ctx, cancel := newCollectContext()
	defer cancel()

	monitors, err := c.client.GetMonitorStats(ctx)
	if err != nil {
		c.log.WithComponent("monitor-collector").Errorf("获取 Monitor 统计失败: %v", err)
		return
	}

	for _, mon := range monitors {
		// 调试日志
		c.log.WithComponent("monitor-collector").Debugf("Monitor %s: InQuorum=%v, boolToFloat64=%f",
			mon.Name, mon.InQuorum, boolToFloat64(mon.InQuorum))

		ch <- prometheus.MustNewConstMetric(c.inQuorum, prometheus.GaugeValue,
			boolToFloat64(mon.InQuorum), mon.Name)
		ch <- prometheus.MustNewConstMetric(c.storeBytes, prometheus.GaugeValue,
			float64(mon.StoreBytes), mon.Name)
		ch <- prometheus.MustNewConstMetric(c.clockSkew, prometheus.GaugeValue,
			mon.ClockSkew, mon.Name)
		// LatencyMs 从毫秒转换为秒，与 Prometheus 惯例一致
		ch <- prometheus.MustNewConstMetric(c.latency, prometheus.GaugeValue,
			mon.LatencyMs/1000.0, mon.Name)
	}
}
