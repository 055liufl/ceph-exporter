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
func (c *RGWCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.total
	ch <- c.activeTotal
	ch <- c.daemonStatus
}

// Collect 执行 RGW 指标采集
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
