// =============================================================================
// Monitor Collector 单元测试
// =============================================================================
// 测试 Monitor 采集器的功能，包括:
//   - Describe: 验证注册的指标描述符数量（4 个）
//   - Collect: 验证每个 Monitor 的指标值和标签
//   - CollectError: 验证 Ceph 命令失败时的错误处理
//
// 测试数据说明:
//
//	monDumpJSON 模拟 "ceph quorum_status -f json" 命令的输出，包含:
//	- mon.a: 在 quorum 中，时钟偏移 0.001s，延迟 1.5ms
//	- mon.b: 在 quorum 中，时钟偏移 -0.002s，延迟 2.3ms
//	- mon.c: 不在 quorum 中，时钟偏移 0s，延迟 50ms（异常高）
//
// 重点测试:
//   - InQuorum 布尔值到浮点数的转换（true->1.0, false->0.0）
//   - 延迟从毫秒到秒的转换（除以 1000）
//
// =============================================================================
package collector

import (
	"testing"

	"ceph-exporter/internal/ceph"
)

// monDumpJSON 模拟 "ceph quorum_status -f json" 命令的输出
// 包含三个 Monitor 的统计信息，其中 mon.c 不在 quorum 中
const monDumpJSON = `{
	"mons": [
		{
			"name": "mon.a",
			"rank": 0,
			"addr": "10.0.0.1:6789",
			"store_bytes": 104857600,
			"clock_skew": 0.001,
			"latency": 1.5,
			"in_quorum": true
		},
		{
			"name": "mon.b",
			"rank": 1,
			"addr": "10.0.0.2:6789",
			"store_bytes": 104857600,
			"clock_skew": -0.002,
			"latency": 2.3,
			"in_quorum": true
		},
		{
			"name": "mon.c",
			"rank": 2,
			"addr": "10.0.0.3:6789",
			"store_bytes": 104857600,
			"clock_skew": 0.0,
			"latency": 50.0,
			"in_quorum": false
		}
	]
}`

// TestMonitorCollector_Describe 测试 Monitor 采集器的指标描述符注册
// 验证 MonitorCollector 注册了正确数量的指标描述符（4 个）:
//   - in_quorum: Monitor 是否在仲裁中
//   - store_bytes: Monitor 数据库存储大小
//   - clock_skew_sec: Monitor 时钟偏移
//   - latency_sec: Monitor 响应延迟
func TestMonitorCollector_Describe(t *testing.T) {
	log := newTestLogger()
	client := ceph.NewTestClient(log, nil)
	c := NewMonitorCollector(client, log)

	descs := collectDescs(c)
	if len(descs) != 4 {
		t.Errorf("expected 4 descriptors, got %d", len(descs))
	}
}

// TestMonitorCollector_Collect 测试 Monitor 采集器的指标采集
// 验证:
//   - 采集到的指标总数（3 个 Monitor * 4 个指标 = 12）
//   - mon.a 和 mon.b 的 in_quorum 为 1（在仲裁中）
//   - mon.c 的 in_quorum 为 0（不在仲裁中）
//   - 延迟转换: mon.a 的 1.5ms / 1000 = 0.0015s
func TestMonitorCollector_Collect(t *testing.T) {
	log := newTestLogger()
	client := ceph.NewTestClient(log, func(args []byte) ([]byte, string, error) {
		return []byte(monDumpJSON), "", nil
	})
	c := NewMonitorCollector(client, log)

	metrics := collectMetrics(c)
	// 3 monitors * 4 metrics each = 12
	if len(metrics) != 12 {
		t.Fatalf("expected 12 metrics, got %d", len(metrics))
	}

	// Verify quorum status
	quorumMetrics := findMetricsByName(metrics, "in_quorum")
	if len(quorumMetrics) != 3 {
		t.Fatalf("expected 3 in_quorum metrics, got %d", len(quorumMetrics))
	}
	for _, m := range quorumMetrics {
		labels := getMetricLabels(m)
		v := getMetricValue(m)
		switch labels["monitor"] {
		case "mon.a", "mon.b":
			if v != 1 {
				t.Errorf("%s in_quorum: expected 1, got %v", labels["monitor"], v)
			}
		case "mon.c":
			if v != 0 {
				t.Errorf("mon.c in_quorum: expected 0, got %v", v)
			}
		}
	}

	// Verify latency conversion (ms to sec): 1.5ms / 1000 = 0.0015s
	latencyMetrics := findMetricsByName(metrics, "latency_sec")
	for _, m := range latencyMetrics {
		labels := getMetricLabels(m)
		if labels["monitor"] == "mon.a" {
			expected := 1.5 / 1000.0
			if v := getMetricValue(m); v != expected {
				t.Errorf("mon.a latency_sec: expected %v, got %v", expected, v)
			}
		}
	}
}

// TestMonitorCollector_CollectError 测试 Ceph 命令失败时的错误处理
// 验证采集器在命令失败时不会产生任何指标
func TestMonitorCollector_CollectError(t *testing.T) {
	log := newTestLogger()
	client := ceph.NewTestClient(log, errMonCommand)
	c := NewMonitorCollector(client, log)

	metrics := collectMetrics(c)
	if len(metrics) != 0 {
		t.Errorf("expected 0 metrics on error, got %d", len(metrics))
	}
}
