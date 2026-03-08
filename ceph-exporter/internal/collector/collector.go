// =============================================================================
// 采集器模块 - 公共定义和基础设施
// =============================================================================
// 本文件提供所有 Prometheus 采集器的公共常量、辅助函数和接口定义。
//
// Phase 2 实现的采集器:
//   - ClusterCollector:  集群整体状态指标（容量、IOPS、PG、OSD 概览）
//   - PoolCollector:     存储池指标（每个池的容量、对象数、IO 速率）
//   - OSDCollector:      OSD 指标（每个 OSD 的容量、利用率、延迟）
//   - MonitorCollector:  Monitor 指标（存储大小、时钟偏移、仲裁状态）
//   - HealthCollector:   健康状态指标（集群健康状态、检查项详情）
//   - MDSCollector:      MDS 指标（元数据服务器状态）
//   - RGWCollector:      RGW 指标（对象网关状态）
//
// 指标命名规范:
//
//	所有指标以 "ceph_" 为前缀，格式为 ceph_<组件>_<指标名>
//	例如: ceph_cluster_total_bytes, ceph_pool_stored_bytes, ceph_osd_utilization
//
// 采集流程:
//  1. Prometheus 定期调用 Collect() 方法
//  2. 采集器通过 Ceph Client 执行命令获取 JSON 数据
//  3. 解析 JSON 并转换为 Prometheus Metric
//  4. 通过 channel 发送给 Prometheus
//
// =============================================================================
package collector

import (
	"context"
	"time"
)

// =============================================================================
// 公共常量
// =============================================================================

const (
	// namespace 是所有 Ceph 指标的命名空间前缀
	// 所有指标名称都以 "ceph_" 开头，符合 Prometheus 命名规范
	namespace = "ceph"

	// defaultCollectTimeout 是单次采集操作的默认超时时间
	// 如果 Ceph 命令在此时间内未返回结果，采集将被取消
	// 设置为 10 秒，足够覆盖大多数 Ceph 命令的执行时间
	defaultCollectTimeout = 10 * time.Second
)

// =============================================================================
// 辅助函数
// =============================================================================

// newCollectContext 创建带超时的采集上下文
// 每次 Collect() 调用时使用，确保采集操作不会无限阻塞
//
// 返回:
//   - context.Context: 带超时的上下文
//   - context.CancelFunc: 取消函数，调用方必须在采集完成后调用以释放资源
func newCollectContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), defaultCollectTimeout)
}

// boolToFloat64 将布尔值转换为 float64
// Prometheus 指标值必须是 float64 类型，布尔值用 1.0（true）和 0.0（false）表示
//
// 参数:
//   - b: 布尔值
//
// 返回:
//   - float64: true 返回 1.0，false 返回 0.0
func boolToFloat64(b bool) float64 {
	if b {
		return 1.0
	}
	return 0.0
}
