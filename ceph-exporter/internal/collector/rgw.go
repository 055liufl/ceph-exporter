// =============================================================================
// RGW Collector - RGW 指标采集器
// =============================================================================
// 采集 Ceph RGW（RADOS Gateway）的状态指标，包括:
//   - RGW 守护进程数量和状态
//
// 数据来源:
//
//	通过 "ceph rgw stat -f json" 或解析 "ceph status" 中的 rgw 信息
//
// 标签:
//
//	name: RGW 守护进程名称
//
// 指标列表:
//
//	ceph_rgw_total               - RGW 守护进程总数
//	ceph_rgw_active_total        - 处于 active 状态的 RGW 数量
//	ceph_rgw_daemon_status       - RGW 守护进程状态（带 name 标签，值恒为 1）
//
// =============================================================================
package collector

import (
	"ceph-exporter/internal/ceph"
	"ceph-exporter/internal/logger"

	"github.com/prometheus/client_golang/prometheus"
)

// RGWCollector RGW 采集器
type RGWCollector struct {
	client *ceph.Client
	log    *logger.Logger

	total        *prometheus.Desc
	activeTotal  *prometheus.Desc
	daemonStatus *prometheus.Desc
}

// NewRGWCollector 创建 RGW 采集器实例
// 初始化所有 RGW 相关的 Prometheus 指标描述符
//
// 参数:
//   - client: Ceph 客户端实例，用于执行命令获取 RGW 数据
//   - log: 日志实例，用于记录采集过程中的信息和错误
//
// 返回:
//   - *RGWCollector: 初始化完成的 RGW 采集器实例
func NewRGWCollector(client *ceph.Client, log *logger.Logger) *RGWCollector {
	return &RGWCollector{
		client: client,
		log:    log,

		total: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "rgw", "total"),
			"RGW 守护进程总数",
			nil, nil,
		),
		activeTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "rgw", "active_total"),
			"处于 active 状态的 RGW 数量",
			nil, nil,
		),
		daemonStatus: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "rgw", "daemon_status"),
			"RGW 守护进程状态（值恒为 1，通过 name 标签区分）",
			[]string{"name"}, nil,
		),
	}
}

// Describe 向 Prometheus 注册本采集器提供的所有指标描述符
// 实现 prometheus.Collector 接口的 Describe 方法
// Prometheus 在注册采集器时会调用此方法，获取采集器提供的所有指标定义
//
// 参数:
//   - ch: 指标描述符通道，用于发送指标描述符到 Prometheus
func (c *RGWCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.total
	ch <- c.activeTotal
	ch <- c.daemonStatus
}

// Collect 执行 RGW 指标采集
// 实现 prometheus.Collector 接口的 Collect 方法
// Prometheus 定期调用此方法采集最新的指标数据
//
// RGW (RADOS Gateway) 说明:
//   - RGW 提供对象存储服务，兼容 S3 和 Swift API
//   - 每个 RGW 守护进程独立运行，可以水平扩展
//   - 通常所有运行中的 RGW 都处于 active 状态
//
// 采集流程:
//  1. 创建带超时的上下文
//  2. 调用 Ceph 客户端获取所有 RGW 的状态数据
//  3. 统计 RGW 总数和 active 状态的数量
//  4. 为每个 RGW 守护进程生成带 name 标签的指标
//  5. 生成汇总指标（total 和 active_total）
//  6. 通过 channel 发送指标到 Prometheus
//
// 参数:
//   - ch: 指标通道，用于发送采集到的指标数据到 Prometheus
func (c *RGWCollector) Collect(ch chan<- prometheus.Metric) {
	ctx, cancel := newCollectContext()
	defer cancel()

	rgwStats, err := c.client.GetRGWStats(ctx)
	if err != nil {
		c.log.WithComponent("rgw-collector").Errorf("获取 RGW 统计失败: %v", err)
		return
	}

	ch <- prometheus.MustNewConstMetric(c.total, prometheus.GaugeValue,
		float64(len(rgwStats.Daemons)))

	var activeCount int
	for _, daemon := range rgwStats.Daemons {
		if daemon.Status == "active" {
			activeCount++
		}
		ch <- prometheus.MustNewConstMetric(c.daemonStatus, prometheus.GaugeValue,
			1, daemon.Name)
	}

	ch <- prometheus.MustNewConstMetric(c.activeTotal, prometheus.GaugeValue,
		float64(activeCount))
}
