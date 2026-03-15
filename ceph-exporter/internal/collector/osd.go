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

// OSDCollector OSD 采集器结构体
// 负责采集 Ceph 集群中所有 OSD 的运行状态和性能指标
//
// 字段说明:
//   - client: Ceph 客户端实例，用于与 Ceph 集群通信
//   - log: 日志记录器，用于记录采集过程中的信息和错误
//   - up: OSD 运行状态指标（1=运行中, 0=已停止）
//   - in: OSD 集群成员状态指标（1=在集群中, 0=已移出）
//   - totalBytes: OSD 总存储容量指标（字节）
//   - usedBytes: OSD 已使用容量指标（字节）
//   - availableBytes: OSD 可用容量指标（字节）
//   - utilization: OSD 利用率指标（百分比）
//   - pgs: OSD 上的 Placement Group 数量指标
//   - applyLatencyMs: OSD apply 操作延迟指标（毫秒）
//   - commitLatencyMs: OSD commit 操作延迟指标（毫秒）
type OSDCollector struct {
	client *ceph.Client   // Ceph 客户端，用于执行 ceph 命令
	log    *logger.Logger // 日志记录器

	// 状态指标
	// up 和 in 是 OSD 的两个关键状态维度:
	//   - up: 表示 OSD 进程是否正在运行
	//   - in: 表示 OSD 是否参与数据分布（是否在 CRUSH map 中）
	// 四种可能的状态组合:
	//   - up + in: 正常工作状态
	//   - up + out: OSD 运行但不接收新数据（通常是维护状态）
	//   - down + in: OSD 停止但仍在 CRUSH map 中（故障状态）
	//   - down + out: OSD 停止且已从集群移除
	up *prometheus.Desc // OSD 是否运行（1=Up, 0=Down）
	in *prometheus.Desc // OSD 是否在集群中（1=In, 0=Out）

	// 容量指标
	// 这些指标用于监控 OSD 的存储空间使用情况
	// 关系: totalBytes = usedBytes + availableBytes
	totalBytes     *prometheus.Desc // OSD 总容量（字节）
	usedBytes      *prometheus.Desc // OSD 已用容量（字节）
	availableBytes *prometheus.Desc // OSD 可用容量（字节）
	utilization    *prometheus.Desc // OSD 利用率（百分比，0-100）

	// PG 和延迟指标
	// pgs: Placement Group 是 Ceph 数据分布的基本单位
	//      每个 OSD 上的 PG 数量应该相对均衡
	//      PG 数量过多或过少都可能影响性能
	// applyLatencyMs: apply 延迟是数据写入日志后应用到对象存储的时间
	// commitLatencyMs: commit 延迟是数据写入日志并持久化的时间
	//                  这两个延迟指标是 OSD 性能的关键指标
	pgs             *prometheus.Desc // OSD 上的 PG 数量
	applyLatencyMs  *prometheus.Desc // Apply 操作延迟（毫秒）
	commitLatencyMs *prometheus.Desc // Commit 操作延迟（毫秒）
}

// NewOSDCollector 创建 OSD 采集器实例
// 初始化所有 OSD 相关的 Prometheus 指标描述符
//
// 此函数创建并配置一个新的 OSD 采集器，用于监控 Ceph 集群中所有 OSD 的状态和性能。
// 每个 OSD 的指标都会带有 "osd" 标签，标签值为 OSD 名称（如 "osd.0", "osd.1"）。
//
// 指标命名规范:
//   - 所有指标都以 "ceph_osd_" 为前缀
//   - 使用下划线分隔单词（snake_case）
//   - 容量相关指标以 "_bytes" 结尾
//   - 延迟相关指标以 "_ms" 结尾
//
// 参数:
//   - client: Ceph 客户端实例，用于执行 "ceph osd df" 和 "ceph osd perf" 命令获取 OSD 数据
//   - log: 日志实例，用于记录采集过程中的信息和错误
//
// 返回:
//   - *OSDCollector: 初始化完成的 OSD 采集器实例，可直接注册到 Prometheus
//
// 使用示例:
//
//	collector := NewOSDCollector(cephClient, logger)
//	prometheus.MustRegister(collector)
func NewOSDCollector(client *ceph.Client, log *logger.Logger) *OSDCollector {
	// 所有 OSD 指标都带有 osd 标签，用于区分不同的 OSD
	// 标签值格式: "osd.0", "osd.1", "osd.2" 等
	osdLabels := []string{"osd"}

	return &OSDCollector{
		client: client,
		log:    log,

		// 状态指标定义
		// up 和 in 是布尔值指标，使用 GaugeValue 类型
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

		// 容量指标定义
		// 所有容量指标的单位都是字节（bytes）
		// 注意: Ceph 命令返回的单位是 KB，需要在采集时转换为字节
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

		// PG 和性能指标定义
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
// 采集流程:
//  1. 创建带超时的上下文（默认 10 秒超时）
//  2. 调用 Ceph 客户端的 GetOSDStats 方法获取所有 OSD 的统计数据
//     - 执行 "ceph osd df -f json" 命令获取容量和利用率数据
//     - 执行 "ceph osd perf -f json" 命令获取延迟数据
//  3. 遍历每个 OSD，生成对应的 Prometheus 指标
//  4. 通过 channel 发送指标到 Prometheus
//
// 数据转换说明:
//   - OSD 容量数据从 Ceph 返回的单位是 KB，需要乘以 1024 转换为字节
//   - Up/In 状态是整数（1 或 0），需要转换为 float64 类型
//   - OSD 名称如果为空，会自动生成为 "osd.<id>" 格式
//
// 错误处理:
//   - 如果获取 OSD 统计失败，记录错误日志并直接返回
//   - 不会抛出 panic，确保采集器的稳定性
//   - 单个 OSD 的数据问题不会影响其他 OSD 的采集
//
// 性能考虑:
//   - 使用带超时的上下文，防止 Ceph 命令执行时间过长
//   - 一次性获取所有 OSD 数据，避免多次调用 Ceph 命令
//   - 使用 MustNewConstMetric 创建常量指标，性能优于可变指标
//
// 参数:
//   - ch: 指标通道，用于发送采集到的指标数据到 Prometheus
//
// 使用的 Ceph 命令:
//   - ceph osd df -f json: 获取 OSD 容量和利用率
//   - ceph osd perf -f json: 获取 OSD 性能延迟数据
func (c *OSDCollector) Collect(ch chan<- prometheus.Metric) {
	// 创建带超时的上下文，防止采集操作阻塞过久
	// 超时时间由 newCollectContext() 函数定义（通常为 10 秒）
	ctx, cancel := newCollectContext()
	defer cancel() // 确保上下文资源被释放

	// 从 Ceph 获取所有 OSD 的统计数据
	// 此调用会执行 Ceph 命令并解析 JSON 响应
	osds, err := c.client.GetOSDStats(ctx)
	if err != nil {
		// 记录错误但不中断程序，确保其他采集器仍能正常工作
		c.log.WithComponent("osd-collector").Errorf("获取 OSD 统计失败: %v", err)
		return
	}

	// 遍历每个 OSD，生成指标
	// 每个 OSD 会生成 9 个指标（2 个状态 + 4 个容量 + 1 个利用率 + 1 个 PG + 2 个延迟）
	for _, osd := range osds {
		// 使用 OSD 名称作为标签值（如 "osd.0"）
		// 如果名称为空，使用 "osd.<id>" 格式
		// 这确保每个 OSD 都有唯一的标识符
		osdName := osd.Name
		if osdName == "" {
			osdName = fmt.Sprintf("osd.%d", osd.ID)
		}

		// 调试日志：输出 OSD 的关键状态信息
		// 这有助于排查 OSD 状态判断的问题
		// Status: OSD 的状态字符串（如 "up", "down"）
		// Reweight: OSD 的权重值（0.0-1.0，0 表示 out）
		// Up(): 计算得出的 Up 状态（1 或 0）
		// In(): 计算得出的 In 状态（1 或 0）
		c.log.WithComponent("osd-collector").Debugf("OSD %s: Status=%q, Reweight=%f, Up()=%d, In()=%d",
			osdName, osd.Status, osd.Reweight, osd.Up(), osd.In())

		// 发送状态指标
		// Up 状态: 1 表示 OSD 进程正在运行，0 表示已停止
		// In 状态: 1 表示 OSD 在 CRUSH map 中参与数据分布，0 表示已移出
		ch <- prometheus.MustNewConstMetric(c.up, prometheus.GaugeValue,
			float64(osd.Up()), osdName)
		ch <- prometheus.MustNewConstMetric(c.in, prometheus.GaugeValue,
			float64(osd.In()), osdName)

		// 发送容量指标
		// 注意: Ceph osd df 返回的容量单位是 KB，需要乘以 1024 转换为字节
		// 这样可以与其他 Prometheus 指标保持一致（通常使用字节作为基本单位）
		ch <- prometheus.MustNewConstMetric(c.totalBytes, prometheus.GaugeValue,
			float64(osd.TotalBytes)*1024, osdName)
		ch <- prometheus.MustNewConstMetric(c.usedBytes, prometheus.GaugeValue,
			float64(osd.UsedBytes)*1024, osdName)
		ch <- prometheus.MustNewConstMetric(c.availableBytes, prometheus.GaugeValue,
			float64(osd.AvailBytes)*1024, osdName)

		// 发送利用率指标
		// 利用率是百分比值（0-100），表示 OSD 已用空间占总空间的比例
		ch <- prometheus.MustNewConstMetric(c.utilization, prometheus.GaugeValue,
			osd.Utilization, osdName)

		// 发送 PG 数量指标
		// PG (Placement Group) 是 Ceph 数据分布的基本单位
		// 每个 OSD 上的 PG 数量应该相对均衡，过多或过少都可能影响性能
		ch <- prometheus.MustNewConstMetric(c.pgs, prometheus.GaugeValue,
			float64(osd.PGs), osdName)

		// 发送延迟指标
		// ApplyLatencyMs: 数据写入日志后应用到对象存储的延迟（毫秒）
		// CommitLatencyMs: 数据写入日志并持久化到磁盘的延迟（毫秒）
		// 这两个指标是 OSD 性能的关键指标，延迟过高可能表示磁盘性能问题
		ch <- prometheus.MustNewConstMetric(c.applyLatencyMs, prometheus.GaugeValue,
			osd.ApplyLatencyMs, osdName)
		ch <- prometheus.MustNewConstMetric(c.commitLatencyMs, prometheus.GaugeValue,
			osd.CommitLatencyMs, osdName)
	}
}
