// =============================================================================
// 采集器测试辅助工具
// =============================================================================
// 本文件提供所有采集器单元测试共用的辅助函数和工具。
// 这些工具简化了 Prometheus 采集器的测试流程，包括:
//   - 创建测试用日志实例
//   - 收集采集器产生的指标和描述符
//   - 从指标中提取值和标签
//   - 按名称查找指标
//   - 模拟错误的 Ceph 命令
//
// 测试模式说明:
//
//	所有采集器测试都使用 ceph.NewTestClient 创建 Mock 客户端，
//	通过自定义 MonCommandFunc 控制 Ceph 命令的返回值，
//	从而在不依赖真实 Ceph 集群的情况下测试采集器逻辑。
//
// =============================================================================
package collector

import (
	"fmt"
	"strings"

	"ceph-exporter/internal/config"
	"ceph-exporter/internal/logger"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

// newTestLogger 创建测试用日志实例
// 使用 error 级别以减少测试输出噪音，使用 text 格式便于阅读
// 返回:
//   - *logger.Logger: 配置好的测试日志实例
func newTestLogger() *logger.Logger {
	log, _ := logger.NewLogger(&config.LoggerConfig{
		Level:  "error", // 只输出 error 及以上级别，减少测试噪音
		Format: "text",  // 使用文本格式，便于人工阅读
	})
	return log
}

// collectMetrics 收集采集器产生的所有 Prometheus 指标
// 模拟 Prometheus 的抓取流程：创建 channel -> 调用 Collect -> 收集所有指标
//
// 工作原理:
//  1. 创建一个带缓冲的 channel（容量 100，足够大多数测试场景）
//  2. 在 goroutine 中调用采集器的 Collect 方法，将指标发送到 channel
//  3. Collect 完成后关闭 channel
//  4. 从 channel 中读取所有指标并返回
//
// 参数:
//   - c: 要测试的 Prometheus 采集器
//
// 返回:
//   - []prometheus.Metric: 采集器产生的所有指标列表
func collectMetrics(c prometheus.Collector) []prometheus.Metric {
	ch := make(chan prometheus.Metric, 100)
	go func() {
		c.Collect(ch)
		close(ch)
	}()
	var metrics []prometheus.Metric
	for m := range ch {
		metrics = append(metrics, m)
	}
	return metrics
}

// collectDescs 收集采集器注册的所有指标描述符
// 模拟 Prometheus 注册采集器时的流程：调用 Describe 获取所有指标定义
//
// 参数:
//   - c: 要测试的 Prometheus 采集器
//
// 返回:
//   - []*prometheus.Desc: 采集器注册的所有指标描述符列表
func collectDescs(c prometheus.Collector) []*prometheus.Desc {
	ch := make(chan *prometheus.Desc, 100)
	go func() {
		c.Describe(ch)
		close(ch)
	}()
	var descs []*prometheus.Desc
	for d := range ch {
		descs = append(descs, d)
	}
	return descs
}

// getMetricValue 从 Prometheus 指标中提取 Gauge 类型的值
// 通过将指标序列化为 protobuf 格式（dto.Metric），然后读取 Gauge 字段的值
//
// 参数:
//   - m: Prometheus 指标实例
//
// 返回:
//   - float64: 指标的 Gauge 值，如果不是 Gauge 类型则返回 0
func getMetricValue(m prometheus.Metric) float64 {
	pb := &dto.Metric{}
	_ = m.Write(pb)
	if pb.Gauge != nil {
		return pb.Gauge.GetValue()
	}
	return 0
}

// getMetricLabels 从 Prometheus 指标中提取所有标签的键值对
// 通过将指标序列化为 protobuf 格式，然后遍历 Label 字段
//
// 参数:
//   - m: Prometheus 指标实例
//
// 返回:
//   - map[string]string: 标签名到标签值的映射
//     例如: {"osd": "osd.0", "pool": "rbd"}
func getMetricLabels(m prometheus.Metric) map[string]string {
	pb := &dto.Metric{}
	_ = m.Write(pb)
	labels := make(map[string]string)
	for _, lp := range pb.Label {
		labels[lp.GetName()] = lp.GetValue()
	}
	return labels
}

// findMetricsByName 按指标描述符中的关键字查找匹配的指标
// 通过检查指标描述符的字符串表示是否包含指定子串来过滤指标
//
// 使用示例:
//   - findMetricsByName(metrics, "osd_up") 查找所有 OSD up 状态指标
//   - findMetricsByName(metrics, "health_status\"") 精确匹配 health_status 指标
//     （末尾加引号可以避免匹配到 health_status_info）
//
// 参数:
//   - metrics: 要搜索的指标列表
//   - nameSubstr: 要匹配的子串
//
// 返回:
//   - []prometheus.Metric: 匹配的指标列表
func findMetricsByName(metrics []prometheus.Metric, nameSubstr string) []prometheus.Metric {
	var found []prometheus.Metric
	for _, m := range metrics {
		if strings.Contains(m.Desc().String(), nameSubstr) {
			found = append(found, m)
		}
	}
	return found
}

// errMonCommand 模拟 Ceph 命令执行失败的 Mock 函数
// 用于测试采集器在 Ceph 命令失败时的错误处理逻辑
// 无论传入什么参数，始终返回错误
//
// 参数:
//   - _: 命令参数（被忽略）
//
// 返回:
//   - []byte: nil（无响应数据）
//   - string: 空字符串（无状态信息）
//   - error: 固定的 "mock error" 错误
func errMonCommand(_ []byte) ([]byte, string, error) {
	return nil, "", fmt.Errorf("mock error")
}
