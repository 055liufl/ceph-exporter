// =============================================================================
// OSD Collector - OSD 指标采集器
// =============================================================================
// 采集每个 Ceph OSD（Object Storage Daemon）的详细指标，包括:
//   - 状态信息（Up/Down、In/Out）
//   - 容量信息（总容量、已用、可用，单位 KB）
//   - 利用率百分比
//   - PG 数量
//   - 延迟信息（apply 延迟、commit 延迟）
//
// 数据来源:
//
//	通过 "ceph osd df -f json" 命令获取 OSD 统计 JSON
//
// 标签:
//
//	osd: OSD 名称（如 osd.0, osd.1）
//
// 指标列表:
//
//	ceph_osd_up                 - OSD 是否处于 Up 状态（1=Up, 0=Down）
//	ceph_osd_in                 - OSD 是否处于 In 状态（1=In, 0=Out）
//	ceph_osd_total_bytes        - OSD 总容量（字节）
//	ceph_osd_used_bytes         - OSD 已用容量（字节）
//	ceph_osd_available_bytes    - OSD 可用容量（字节）
//	ceph_osd_utilization        - OSD 利用率百分比
//	ceph_osd_pgs                - OSD 上的 PG 数量
//	ceph_osd_apply_latency_ms   - OSD apply 延迟（毫秒）
//	ceph_osd_commit_latency_ms  - OSD commit 延迟（毫秒）
//
// =============================================================================
package collector

import (
	"fmt"

	"ceph-exporter/internal/ceph"
	"ceph-exporter/internal/logger"

	"github.com/prometheus/client_golang/prometheus"
)

// OSDCollector OSD 采集器
type OSDCollector struct {
	client *ceph.Client
	log    *logger.Logger

	// 状态指标
	up *prometheus.Desc
	in *prometheus.Desc

	// 容量指标
	totalBytes     *prometheus.Desc
	usedBytes      *prometheus.Desc
	availableBytes *prometheus.Desc
	utilization    *prometheus.Desc

	// PG 和延迟指标
	pgs             *prometheus.Desc
	applyLatencyMs  *prometheus.Desc
	commitLatencyMs *prometheus.Desc
}

// NewOSDCollector 创建 OSD 采集器实例
// 初始化所有 OSD 相关的 Prometheus 指标描述符
//
// 参数:
//   - client: Ceph 客户端实例，用于执行命令获取 OSD 数据
//   - log: 日志实例，用于记录采集过程中的信息和错误
//
// 返回:
//   - *OSDCollector: 初始化完成的 OSD 采集器实例
func NewOSDCollector(client *ceph.Client, log *logger.Logger) *OSDCollector {
	// 所有 OSD 指标都带有 osd 标签，用于区分不同的 OSD
	osdLabels := []string{"osd"}

	return &OSDCollector{
		client: client,
		log:    log,

		up: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "osd", "up"),
			"OSD 是否处于 Up 状态（1=Up, 0=Down）",
			osdLabels, nil,
		),
		in: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "osd", "in"),
			"OSD 是否处于 In 状态（1=In, 0=Out）",
			osdLabels, nil,
		),
		totalBytes: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "osd", "total_bytes"),
			"OSD 总容量（字节）",
			osdLabels, nil,
		),
		usedBytes: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "osd", "used_bytes"),
			"OSD 已用容量（字节）",
			osdLabels, nil,
		),
		availableBytes: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "osd", "available_bytes"),
			"OSD 可用容量（字节）",
			osdLabels, nil,
		),
		utilization: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "osd", "utilization"),
			"OSD 利用率百分比",
			osdLabels, nil,
		),
		pgs: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "osd", "pgs"),
			"OSD 上的 Placement Group 数量",
			osdLabels, nil,
		),
		applyLatencyMs: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "osd", "apply_latency_ms"),
			"OSD apply 延迟（毫秒）",
			osdLabels, nil,
		),
		commitLatencyMs: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "osd", "commit_latency_ms"),
			"OSD commit 延迟（毫秒）",
			osdLabels, nil,
		),
	}
}

// Describe 向 Prometheus 注册本采集器提供的所有指标描述符
// 实现 prometheus.Collector 接口的 Describe 方法
// Prometheus 在注册采集器时会调用此方法，获取采集器提供的所有指标定义
//
// 参数:
//   - ch: 指标描述符通道，用于发送指标描述符到 Prometheus
func (c *OSDCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.up
	ch <- c.in
	ch <- c.totalBytes
	ch <- c.usedBytes
	ch <- c.availableBytes
	ch <- c.utilization
	ch <- c.pgs
	ch <- c.applyLatencyMs
	ch <- c.commitLatencyMs
}

// Collect 执行 OSD 指标采集
// 实现 prometheus.Collector 接口的 Collect 方法
// Prometheus 定期调用此方法采集最新的指标数据
// 遍历所有 OSD，为每个 OSD 生成一组带 osd 标签的指标
//
// 注意事项:
//   - OSD 容量数据从 Ceph 返回的单位是 KB，需要转换为字节（乘以 1024）
//   - Up/In 状态是整数（1 或 0），需要转换为浮点数
//
// 采集流程:
//  1. 创建带超时的上下文
//  2. 调用 Ceph 客户端获取所有 OSD 的统计数据
//  3. 遍历每个 OSD，生成对应的 Prometheus 指标
//  4. 通过 channel 发送指标到 Prometheus
//
// 参数:
//   - ch: 指标通道，用于发送采集到的指标数据到 Prometheus
func (c *OSDCollector) Collect(ch chan<- prometheus.Metric) {
	ctx, cancel := newCollectContext()
	defer cancel()

	osds, err := c.client.GetOSDStats(ctx)
	if err != nil {
		c.log.WithComponent("osd-collector").Errorf("获取 OSD 统计失败: %v", err)
		return
	}

	for _, osd := range osds {
		// 使用 OSD 名称作为标签值（如 "osd.0"）
		// 如果名称为空，使用 "osd.<id>" 格式
		osdName := osd.Name
		if osdName == "" {
			osdName = fmt.Sprintf("osd.%d", osd.ID)
		}

		ch <- prometheus.MustNewConstMetric(c.up, prometheus.GaugeValue,
			float64(osd.Up), osdName)
		ch <- prometheus.MustNewConstMetric(c.in, prometheus.GaugeValue,
			float64(osd.In), osdName)

		// Ceph osd df 返回的容量单位是 KB，转换为字节
		ch <- prometheus.MustNewConstMetric(c.totalBytes, prometheus.GaugeValue,
			float64(osd.TotalBytes)*1024, osdName)
		ch <- prometheus.MustNewConstMetric(c.usedBytes, prometheus.GaugeValue,
			float64(osd.UsedBytes)*1024, osdName)
		ch <- prometheus.MustNewConstMetric(c.availableBytes, prometheus.GaugeValue,
			float64(osd.AvailBytes)*1024, osdName)

		ch <- prometheus.MustNewConstMetric(c.utilization, prometheus.GaugeValue,
			osd.Utilization, osdName)
		ch <- prometheus.MustNewConstMetric(c.pgs, prometheus.GaugeValue,
			float64(osd.PGs), osdName)
		ch <- prometheus.MustNewConstMetric(c.applyLatencyMs, prometheus.GaugeValue,
			osd.ApplyLatencyMs, osdName)
		ch <- prometheus.MustNewConstMetric(c.commitLatencyMs, prometheus.GaugeValue,
			osd.CommitLatencyMs, osdName)
	}
}
