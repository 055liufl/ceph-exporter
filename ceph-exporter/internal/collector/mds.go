// =============================================================================
// MDS Collector - MDS 指标采集器
// =============================================================================
// 采集 Ceph MDS（Metadata Server）的状态指标，包括:
//   - MDS 守护进程状态（active, standby, standby-replay 等）
//   - MDS 排名信息
//
// 数据来源:
//
//	通过 "ceph mds stat -f json" 命令获取 MDS 状态 JSON
//
// 标签:
//
//	name: MDS 守护进程名称
//	rank: MDS 排名（active MDS 的排名编号）
//
// 指标列表:
//
//	ceph_mds_active_total        - 处于 active 状态的 MDS 数量
//	ceph_mds_standby_total       - 处于 standby 状态的 MDS 数量
//	ceph_mds_daemon_status       - MDS 守护进程状态（带 name, state 标签，值恒为 1）
//
// =============================================================================
package collector

import (
	"ceph-exporter/internal/ceph"
	"ceph-exporter/internal/logger"

	"github.com/prometheus/client_golang/prometheus"
)

// MDSCollector MDS 采集器
type MDSCollector struct {
	client *ceph.Client
	log    *logger.Logger

	activeTotal  *prometheus.Desc
	standbyTotal *prometheus.Desc
	daemonStatus *prometheus.Desc
}

// NewMDSCollector 创建 MDS 采集器实例
func NewMDSCollector(client *ceph.Client, log *logger.Logger) *MDSCollector {
	return &MDSCollector{
		client: client,
		log:    log,

		activeTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "mds", "active_total"),
			"处于 active 状态的 MDS 数量",
			nil, nil,
		),
		standbyTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "mds", "standby_total"),
			"处于 standby 状态的 MDS 数量",
			nil, nil,
		),
		daemonStatus: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "mds", "daemon_status"),
			"MDS 守护进程状态（值恒为 1，通过 name 和 state 标签区分）",
			[]string{"name", "state"}, nil,
		),
	}
}

// Describe 向 Prometheus 注册本采集器提供的所有指标描述符
func (c *MDSCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.activeTotal
	ch <- c.standbyTotal
	ch <- c.daemonStatus
}

// Collect 执行 MDS 指标采集
func (c *MDSCollector) Collect(ch chan<- prometheus.Metric) {
	ctx, cancel := newCollectContext()
	defer cancel()

	mdsStats, err := c.client.GetMDSStats(ctx)
	if err != nil {
		c.log.WithComponent("mds-collector").Errorf("获取 MDS 统计失败: %v", err)
		return
	}

	// 统计 active 和 standby 数量
	var activeCount, standbyCount int
	for _, daemon := range mdsStats.Daemons {
		switch daemon.State {
		case "up:active":
			activeCount++
		case "up:standby", "up:standby-replay":
			standbyCount++
		}

		// 每个 MDS 守护进程生成一个带 name 和 state 标签的指标
		ch <- prometheus.MustNewConstMetric(c.daemonStatus, prometheus.GaugeValue,
			1, daemon.Name, daemon.State)
	}

	ch <- prometheus.MustNewConstMetric(c.activeTotal, prometheus.GaugeValue,
		float64(activeCount))
	ch <- prometheus.MustNewConstMetric(c.standbyTotal, prometheus.GaugeValue,
		float64(standbyCount))
}
