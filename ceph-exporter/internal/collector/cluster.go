// =============================================================================
// Cluster Collector - 集群整体状态指标采集器
// =============================================================================
// 采集 Ceph 集群的全局状态指标，包括:
//   - 集群容量信息（总容量、已用、可用）
//   - 集群 IO 吞吐量（读写字节数/秒、读写操作数/秒）
//   - PG（Placement Group）统计（总数、各状态分布）
//   - OSD 概览（总数、Up 数、In 数）
//   - Monitor 数量
//   - 对象总数
//
// 数据来源:
//
//	通过 "ceph status -f json" 命令获取集群状态 JSON
//
// 指标列表:
//
//	ceph_cluster_total_bytes        - 集群总容量（字节）
//	ceph_cluster_used_bytes         - 集群已用容量（字节）
//	ceph_cluster_available_bytes    - 集群可用容量（字节）
//	ceph_cluster_objects_total      - 集群对象总数
//	ceph_cluster_read_bytes_sec     - 集群读取吞吐量（字节/秒）
//	ceph_cluster_write_bytes_sec    - 集群写入吞吐量（字节/秒）
//	ceph_cluster_read_ops_sec       - 集群读取 IOPS
//	ceph_cluster_write_ops_sec      - 集群写入 IOPS
//	ceph_cluster_pgs_total          - PG 总数
//	ceph_cluster_pgs_by_state       - 各状态 PG 数量（按 state 标签区分）
//	ceph_cluster_pools_total        - 存储池总数
//	ceph_cluster_osds_total         - OSD 总数
//	ceph_cluster_osds_up            - 处于 Up 状态的 OSD 数量
//	ceph_cluster_osds_in            - 处于 In 状态的 OSD 数量
//	ceph_cluster_mons_total         - Monitor 总数
//
// =============================================================================
package collector

import (
	"ceph-exporter/internal/ceph"
	"ceph-exporter/internal/logger"

	"github.com/prometheus/client_golang/prometheus"
)

// ClusterCollector 集群状态采集器
// 实现 prometheus.Collector 接口，负责采集 Ceph 集群的全局状态指标
type ClusterCollector struct {
	client *ceph.Client   // Ceph 客户端，用于执行命令获取数据
	log    *logger.Logger // 日志实例

	// ----- 容量指标描述符 -----
	totalBytes     *prometheus.Desc // 集群总容量
	usedBytes      *prometheus.Desc // 集群已用容量
	availableBytes *prometheus.Desc // 集群可用容量
	objectsTotal   *prometheus.Desc // 对象总数

	// ----- IO 吞吐量指标描述符 -----
	readBytesSec  *prometheus.Desc // 读取吞吐量（字节/秒）
	writeBytesSec *prometheus.Desc // 写入吞吐量（字节/秒）
	readOpsSec    *prometheus.Desc // 读取 IOPS
	writeOpsSec   *prometheus.Desc // 写入 IOPS

	// ----- PG 指标描述符 -----
	pgsTotal   *prometheus.Desc // PG 总数
	pgsByState *prometheus.Desc // 各状态 PG 数量

	// ----- 组件数量指标描述符 -----
	poolsTotal *prometheus.Desc // 存储池总数
	osdsTotal  *prometheus.Desc // OSD 总数
	osdsUp     *prometheus.Desc // Up 状态 OSD 数量
	osdsIn     *prometheus.Desc // In 状态 OSD 数量
	monsTotal  *prometheus.Desc // Monitor 总数
}

// NewClusterCollector 创建集群状态采集器实例
// 初始化所有 Prometheus 指标描述符（prometheus.Desc）
//
// 参数:
//   - client: Ceph 客户端实例
//   - log: 日志实例
//
// 返回:
//   - *ClusterCollector: 初始化完成的采集器实例
func NewClusterCollector(client *ceph.Client, log *logger.Logger) *ClusterCollector {
	return &ClusterCollector{
		client: client,
		log:    log,

		// ----- 容量指标 -----
		totalBytes: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "cluster", "total_bytes"),
			"Ceph 集群总容量（字节）",
			nil, nil, // 无标签，集群级别指标
		),
		usedBytes: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "cluster", "used_bytes"),
			"Ceph 集群已用容量（字节）",
			nil, nil,
		),
		availableBytes: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "cluster", "available_bytes"),
			"Ceph 集群可用容量（字节）",
			nil, nil,
		),
		objectsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "cluster", "objects_total"),
			"Ceph 集群中的对象总数",
			nil, nil,
		),

		// ----- IO 吞吐量指标 -----
		readBytesSec: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "cluster", "read_bytes_sec"),
			"集群读取吞吐量（字节/秒）",
			nil, nil,
		),
		writeBytesSec: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "cluster", "write_bytes_sec"),
			"集群写入吞吐量（字节/秒）",
			nil, nil,
		),
		readOpsSec: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "cluster", "read_ops_sec"),
			"集群读取操作数（IOPS）",
			nil, nil,
		),
		writeOpsSec: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "cluster", "write_ops_sec"),
			"集群写入操作数（IOPS）",
			nil, nil,
		),

		// ----- PG 指标 -----
		pgsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "cluster", "pgs_total"),
			"Placement Group 总数",
			nil, nil,
		),
		pgsByState: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "cluster", "pgs_by_state"),
			"各状态的 Placement Group 数量",
			[]string{"state"}, nil, // 按 PG 状态（如 active+clean）区分
		),

		// ----- 组件数量指标 -----
		poolsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "cluster", "pools_total"),
			"存储池总数",
			nil, nil,
		),
		osdsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "cluster", "osds_total"),
			"OSD 总数",
			nil, nil,
		),
		osdsUp: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "cluster", "osds_up"),
			"处于 Up 状态的 OSD 数量",
			nil, nil,
		),
		osdsIn: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "cluster", "osds_in"),
			"处于 In 状态的 OSD 数量",
			nil, nil,
		),
		monsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "cluster", "mons_total"),
			"Monitor 总数",
			nil, nil,
		),
	}
}

// Describe 向 Prometheus 注册本采集器提供的所有指标描述符
// 实现 prometheus.Collector 接口的 Describe 方法
// Prometheus 在启动时调用此方法，了解采集器能提供哪些指标
//
// 参数:
//   - ch: Prometheus 描述符通道，将所有指标描述符发送到此通道
func (c *ClusterCollector) Describe(ch chan<- *prometheus.Desc) {
	// 容量指标
	ch <- c.totalBytes
	ch <- c.usedBytes
	ch <- c.availableBytes
	ch <- c.objectsTotal

	// IO 吞吐量指标
	ch <- c.readBytesSec
	ch <- c.writeBytesSec
	ch <- c.readOpsSec
	ch <- c.writeOpsSec

	// PG 指标
	ch <- c.pgsTotal
	ch <- c.pgsByState

	// 组件数量指标
	ch <- c.poolsTotal
	ch <- c.osdsTotal
	ch <- c.osdsUp
	ch <- c.osdsIn
	ch <- c.monsTotal
}

// Collect 执行实际的指标采集
// 实现 prometheus.Collector 接口的 Collect 方法
// Prometheus 每次抓取（scrape）时调用此方法获取最新指标值
//
// 采集流程:
//  1. 创建带超时的上下文
//  2. 调用 Ceph Client 获取集群状态 JSON
//  3. 将 JSON 数据转换为 Prometheus Metric
//  4. 通过 channel 发送给 Prometheus
//
// 错误处理:
//
//	如果采集失败，记录错误日志但不 panic，确保其他采集器不受影响
//
// 参数:
//   - ch: Prometheus 指标通道，将采集到的指标值发送到此通道
func (c *ClusterCollector) Collect(ch chan<- prometheus.Metric) {
	// 创建带超时的采集上下文，防止 Ceph 命令长时间阻塞
	ctx, cancel := newCollectContext()
	defer cancel()

	// 通过 Ceph Client 获取集群状态
	status, err := c.client.GetClusterStatus(ctx)
	if err != nil {
		// 采集失败时记录错误日志，但不中断其他采集器的工作
		c.log.WithComponent("cluster-collector").Errorf("获取集群状态失败: %v", err)
		return
	}

	// ----- 发送容量指标 -----
	ch <- prometheus.MustNewConstMetric(c.totalBytes, prometheus.GaugeValue,
		float64(status.PGMap.BytesTotal))
	ch <- prometheus.MustNewConstMetric(c.usedBytes, prometheus.GaugeValue,
		float64(status.PGMap.BytesUsed))
	ch <- prometheus.MustNewConstMetric(c.availableBytes, prometheus.GaugeValue,
		float64(status.PGMap.BytesAvail))
	ch <- prometheus.MustNewConstMetric(c.objectsTotal, prometheus.GaugeValue,
		float64(status.PGMap.NumObjects))

	// ----- 发送 IO 吞吐量指标 -----
	ch <- prometheus.MustNewConstMetric(c.readBytesSec, prometheus.GaugeValue,
		float64(status.PGMap.ReadBytesSec))
	ch <- prometheus.MustNewConstMetric(c.writeBytesSec, prometheus.GaugeValue,
		float64(status.PGMap.WriteBytesSec))
	ch <- prometheus.MustNewConstMetric(c.readOpsSec, prometheus.GaugeValue,
		float64(status.PGMap.ReadOpPerSec))
	ch <- prometheus.MustNewConstMetric(c.writeOpsSec, prometheus.GaugeValue,
		float64(status.PGMap.WriteOpPerSec))

	// ----- 发送 PG 指标 -----
	ch <- prometheus.MustNewConstMetric(c.pgsTotal, prometheus.GaugeValue,
		float64(status.PGMap.NumPGs))

	// 遍历各状态的 PG 数量，每个状态生成一个带 state 标签的指标
	// 例如: ceph_cluster_pgs_by_state{state="active+clean"} 128
	for _, pgState := range status.PGMap.PGsByState {
		ch <- prometheus.MustNewConstMetric(c.pgsByState, prometheus.GaugeValue,
			float64(pgState.Count), pgState.StateName)
	}

	// ----- 发送组件数量指标 -----
	ch <- prometheus.MustNewConstMetric(c.poolsTotal, prometheus.GaugeValue,
		float64(status.PGMap.NumPools))
	ch <- prometheus.MustNewConstMetric(c.osdsTotal, prometheus.GaugeValue,
		float64(status.OSDMap.NumOSDs))
	ch <- prometheus.MustNewConstMetric(c.osdsUp, prometheus.GaugeValue,
		float64(status.OSDMap.NumUpOSDs))
	ch <- prometheus.MustNewConstMetric(c.osdsIn, prometheus.GaugeValue,
		float64(status.OSDMap.NumInOSDs))
	ch <- prometheus.MustNewConstMetric(c.monsTotal, prometheus.GaugeValue,
		float64(status.MonMap.NumMons))
}
