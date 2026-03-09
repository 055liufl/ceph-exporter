// =============================================================================
// Health Collector - 健康状态指标采集器
// =============================================================================
// 采集 Ceph 集群的健康状态指标，包括:
//   - 集群健康状态码（HEALTH_OK=0, HEALTH_WARN=1, HEALTH_ERR=2）
//   - 健康检查项数量（按严重程度分类）
//
// 数据来源:
//
//	通过 "ceph status -f json" 命令获取集群状态中的 health 部分
//
// 指标列表:
//
//	ceph_health_status           - 集群健康状态码（0=OK, 1=WARN, 2=ERR）
//	ceph_health_status_info      - 集群健康状态信息（带 status 标签，值恒为 1）
//	ceph_health_checks_total     - 健康检查项总数
//	ceph_health_check            - 各健康检查项（带 name, severity 标签，值恒为 1）
//
// =============================================================================
package collector

import (
	"ceph-exporter/internal/ceph"
	"ceph-exporter/internal/logger"

	"github.com/prometheus/client_golang/prometheus"
)

// healthStatusMap 将 Ceph 健康状态字符串映射为数值
// 数值越大表示状态越严重，便于 Prometheus 告警规则编写
var healthStatusMap = map[string]float64{
	"HEALTH_OK":   0,
	"HEALTH_WARN": 1,
	"HEALTH_ERR":  2,
}

// HealthCollector 健康状态采集器
type HealthCollector struct {
	client *ceph.Client
	log    *logger.Logger

	status      *prometheus.Desc
	statusInfo  *prometheus.Desc
	checksTotal *prometheus.Desc
	check       *prometheus.Desc
}

// NewHealthCollector 创建健康状态采集器实例
// 初始化所有健康状态相关的 Prometheus 指标描述符
//
// 参数:
//   - client: Ceph 客户端实例，用于执行命令获取集群健康状态数据
//   - log: 日志实例，用于记录采集过程中的信息和错误
//
// 返回:
//   - *HealthCollector: 初始化完成的健康状态采集器实例
func NewHealthCollector(client *ceph.Client, log *logger.Logger) *HealthCollector {
	return &HealthCollector{
		client: client,
		log:    log,

		status: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "health", "status"),
			"集群健康状态码（0=HEALTH_OK, 1=HEALTH_WARN, 2=HEALTH_ERR）",
			nil, nil,
		),
		statusInfo: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "health", "status_info"),
			"集群健康状态信息（值恒为 1，通过 status 标签区分状态）",
			[]string{"status"}, nil,
		),
		checksTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "health", "checks_total"),
			"健康检查项总数",
			nil, nil,
		),
		check: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "health", "check"),
			"健康检查项（值恒为 1，通过 name 和 severity 标签区分）",
			[]string{"name", "severity"}, nil,
		),
	}
}

// Describe 向 Prometheus 注册本采集器提供的所有指标描述符
// 实现 prometheus.Collector 接口的 Describe 方法
// Prometheus 在注册采集器时会调用此方法，获取采集器提供的所有指标定义
//
// 参数:
//   - ch: 指标描述符通道，用于发送指标描述符到 Prometheus
func (c *HealthCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.status
	ch <- c.statusInfo
	ch <- c.checksTotal
	ch <- c.check
}

// Collect 执行健康状态指标采集
// 实现 prometheus.Collector 接口的 Collect 方法
// Prometheus 定期调用此方法采集最新的指标数据
//
// 健康状态说明:
//   - HEALTH_OK (0): 集群完全健康，无任何问题
//   - HEALTH_WARN (1): 集群有警告，但仍可正常工作
//   - HEALTH_ERR (2): 集群有错误，可能影响数据可用性
//
// 健康检查项示例:
//   - TOO_FEW_OSDS: OSD 数量不足
//   - OSD_DOWN: 有 OSD 处于 down 状态
//   - PG_DEGRADED: 有 PG 处于降级状态
//   - MON_CLOCK_SKEW: Monitor 时钟偏移过大
//
// 采集流程:
//  1. 创建带超时的上下文
//  2. 调用 Ceph 客户端获取集群状态（包含健康信息）
//  3. 生成健康状态码指标（数值型，便于告警规则）
//  4. 生成健康状态信息指标（带标签，便于 Grafana 展示）
//  5. 生成健康检查项总数指标
//  6. 为每个健康检查项生成带 name 和 severity 标签的指标
//  7. 通过 channel 发送指标到 Prometheus
//
// 参数:
//   - ch: 指标通道，用于发送采集到的指标数据到 Prometheus
func (c *HealthCollector) Collect(ch chan<- prometheus.Metric) {
	ctx, cancel := newCollectContext()
	defer cancel()

	status, err := c.client.GetClusterStatus(ctx)
	if err != nil {
		c.log.WithComponent("health-collector").Errorf("获取集群健康状态失败: %v", err)
		return
	}

	// 发送健康状态码（数值型，便于告警）
	statusCode, ok := healthStatusMap[status.Health.Status]
	if !ok {
		// 未知状态视为最严重
		statusCode = 2
	}
	ch <- prometheus.MustNewConstMetric(c.status, prometheus.GaugeValue, statusCode)

	// 发送健康状态信息（带标签，便于 Grafana 展示）
	ch <- prometheus.MustNewConstMetric(c.statusInfo, prometheus.GaugeValue,
		1, status.Health.Status)

	// 发送健康检查项总数
	ch <- prometheus.MustNewConstMetric(c.checksTotal, prometheus.GaugeValue,
		float64(len(status.Health.Checks)))

	// 遍历每个健康检查项，生成带 name 和 severity 标签的指标
	// 例如: ceph_health_check{name="TOO_FEW_OSDS",severity="HEALTH_WARN"} 1
	for name, checkInfo := range status.Health.Checks {
		ch <- prometheus.MustNewConstMetric(c.check, prometheus.GaugeValue,
			1, name, checkInfo.Severity)
	}
}
