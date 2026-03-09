// =============================================================================
// Pool Collector - 存储池指标采集器
// =============================================================================
// 采集每个 Ceph 存储池的详细指标，包括:
//   - 存储容量（已存储字节数、最大可用、已用字节数、使用率）
//   - 对象数量
//   - IO 速率（读写字节数/秒、读写操作数/秒）
//
// 数据来源:
//
//	通过 "ceph osd pool stats -f json" 命令获取存储池统计 JSON
//
// 标签:
//
//	pool: 存储池名称（如 rbd, cephfs_data, .rgw.root）
//
// 指标列表:
//
//	ceph_pool_stored_bytes       - 存储池已存储数据量（字节）
//	ceph_pool_max_available_bytes - 存储池最大可用容量（字节）
//	ceph_pool_used_bytes         - 存储池已用容量（字节）
//	ceph_pool_percent_used       - 存储池使用率（0.0-1.0）
//	ceph_pool_objects_total      - 存储池中的对象数量
//	ceph_pool_read_bytes_sec     - 存储池读取吞吐量（字节/秒）
//	ceph_pool_write_bytes_sec    - 存储池写入吞吐量（字节/秒）
//	ceph_pool_read_ops_sec       - 存储池读取 IOPS
//	ceph_pool_write_ops_sec      - 存储池写入 IOPS
//
// =============================================================================
package collector

import (
	"ceph-exporter/internal/ceph"
	"ceph-exporter/internal/logger"

	"github.com/prometheus/client_golang/prometheus"
)

// PoolCollector 存储池采集器
type PoolCollector struct {
	client *ceph.Client
	log    *logger.Logger

	// 容量指标
	storedBytes   *prometheus.Desc
	maxAvailBytes *prometheus.Desc
	usedBytes     *prometheus.Desc
	percentUsed   *prometheus.Desc
	objectsTotal  *prometheus.Desc

	// IO 速率指标
	readBytesSec  *prometheus.Desc
	writeBytesSec *prometheus.Desc
	readOpsSec    *prometheus.Desc
	writeOpsSec   *prometheus.Desc
}

// NewPoolCollector 创建存储池采集器实例
// 初始化所有存储池相关的 Prometheus 指标描述符
//
// 参数:
//   - client: Ceph 客户端实例，用于执行命令获取存储池数据
//   - log: 日志实例，用于记录采集过程中的信息和错误
//
// 返回:
//   - *PoolCollector: 初始化完成的存储池采集器实例
func NewPoolCollector(client *ceph.Client, log *logger.Logger) *PoolCollector {
	// 所有存储池指标都带有 pool 标签，用于区分不同的存储池
	poolLabels := []string{"pool"}

	return &PoolCollector{
		client: client,
		log:    log,

		storedBytes: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "pool", "stored_bytes"),
			"存储池已存储数据量（字节）",
			poolLabels, nil,
		),
		maxAvailBytes: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "pool", "max_available_bytes"),
			"存储池最大可用容量（字节）",
			poolLabels, nil,
		),
		usedBytes: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "pool", "used_bytes"),
			"存储池已用容量（字节）",
			poolLabels, nil,
		),
		percentUsed: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "pool", "percent_used"),
			"存储池使用率（0.0-1.0）",
			poolLabels, nil,
		),
		objectsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "pool", "objects_total"),
			"存储池中的对象数量",
			poolLabels, nil,
		),
		readBytesSec: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "pool", "read_bytes_sec"),
			"存储池读取吞吐量（字节/秒）",
			poolLabels, nil,
		),
		writeBytesSec: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "pool", "write_bytes_sec"),
			"存储池写入吞吐量（字节/秒）",
			poolLabels, nil,
		),
		readOpsSec: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "pool", "read_ops_sec"),
			"存储池读取操作数（IOPS）",
			poolLabels, nil,
		),
		writeOpsSec: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "pool", "write_ops_sec"),
			"存储池写入操作数（IOPS）",
			poolLabels, nil,
		),
	}
}

// Describe 向 Prometheus 注册本采集器提供的所有指标描述符
// 实现 prometheus.Collector 接口的 Describe 方法
// Prometheus 在注册采集器时会调用此方法，获取采集器提供的所有指标定义
//
// 参数:
//   - ch: 指标描述符通道，用于发送指标描述符到 Prometheus
func (c *PoolCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.storedBytes
	ch <- c.maxAvailBytes
	ch <- c.usedBytes
	ch <- c.percentUsed
	ch <- c.objectsTotal
	ch <- c.readBytesSec
	ch <- c.writeBytesSec
	ch <- c.readOpsSec
	ch <- c.writeOpsSec
}

// Collect 执行存储池指标采集
// 实现 prometheus.Collector 接口的 Collect 方法
// Prometheus 定期调用此方法采集最新的指标数据
// 遍历所有存储池，为每个池生成一组带 pool 标签的指标
//
// 采集流程:
//  1. 创建带超时的上下文
//  2. 调用 Ceph 客户端获取所有存储池的统计数据
//  3. 遍历每个存储池，生成对应的 Prometheus 指标
//  4. 通过 channel 发送指标到 Prometheus
//
// 参数:
//   - ch: 指标通道，用于发送采集到的指标数据到 Prometheus
func (c *PoolCollector) Collect(ch chan<- prometheus.Metric) {
	// 创建带超时的上下文，防止采集操作阻塞过久
	ctx, cancel := newCollectContext()
	defer cancel()

	// 从 Ceph 获取所有存储池的统计数据
	pools, err := c.client.GetPoolStats(ctx)
	if err != nil {
		c.log.WithComponent("pool-collector").Errorf("获取存储池统计失败: %v", err)
		return
	}

	// 遍历每个存储池，生成指标
	for _, pool := range pools {
		// 每个存储池的指标都带有 pool=<pool_name> 标签，用于区分不同的存储池
		ch <- prometheus.MustNewConstMetric(c.storedBytes, prometheus.GaugeValue,
			float64(pool.Stats.Stored), pool.PoolName)
		ch <- prometheus.MustNewConstMetric(c.maxAvailBytes, prometheus.GaugeValue,
			float64(pool.Stats.MaxAvail), pool.PoolName)
		ch <- prometheus.MustNewConstMetric(c.usedBytes, prometheus.GaugeValue,
			float64(pool.Stats.BytesUsed), pool.PoolName)
		ch <- prometheus.MustNewConstMetric(c.percentUsed, prometheus.GaugeValue,
			pool.Stats.PercentUsed, pool.PoolName)
		ch <- prometheus.MustNewConstMetric(c.objectsTotal, prometheus.GaugeValue,
			float64(pool.Stats.Objects), pool.PoolName)

		ch <- prometheus.MustNewConstMetric(c.readBytesSec, prometheus.GaugeValue,
			float64(pool.ClientIORate.ReadBytesSec), pool.PoolName)
		ch <- prometheus.MustNewConstMetric(c.writeBytesSec, prometheus.GaugeValue,
			float64(pool.ClientIORate.WriteBytesSec), pool.PoolName)
		ch <- prometheus.MustNewConstMetric(c.readOpsSec, prometheus.GaugeValue,
			float64(pool.ClientIORate.ReadOpPerSec), pool.PoolName)
		ch <- prometheus.MustNewConstMetric(c.writeOpsSec, prometheus.GaugeValue,
			float64(pool.ClientIORate.WriteOpPerSec), pool.PoolName)
	}
}
