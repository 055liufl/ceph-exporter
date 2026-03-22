// =============================================================================
// OSD Collector 单元测试
// =============================================================================
// 测试 OSD 采集器的功能，包括:
//   - Describe: 验证注册的指标描述符数量（9 个）
//   - Collect: 验证每个 OSD 的指标值、标签和单位转换
//   - CollectError: 验证 Ceph 命令失败时的错误处理
//
// 测试数据说明:
//
//	osdDFJSON 模拟 "ceph osd df -f json" 命令的输出，包含:
//	- osd.0: 正常状态（up+in），1GB 总容量，50% 利用率，64 个 PG
//	- osd.1: 名称为空（测试自动命名），up+out 状态，2GB 总容量，10% 利用率
//
// 重点测试:
//   - KB 到字节的单位转换（乘以 1024）
//   - 空名称 OSD 的自动命名（"osd.<id>" 格式）
//   - Up/In 状态的正确判断
//
// =============================================================================
package collector

import (
	"testing"

	"ceph-exporter/internal/ceph"
)

// osdDFJSON 模拟 "ceph osd df -f json" 命令的输出
// 包含两个 OSD 的统计信息，其中第二个 OSD 名称为空用于测试自动命名
const osdDFJSON = `{
	"nodes": [
		{
			"id": 0,
			"name": "osd.0",
			"up": 1,
			"in": 1,
			"kb": 1048576,
			"kb_used": 524288,
			"kb_avail": 524288,
			"utilization": 50.0,
			"pgs": 64,
			"apply_latency_ms": 1.5,
			"commit_latency_ms": 2.3
		},
		{
			"id": 1,
			"name": "",
			"up": 1,
			"in": 0,
			"kb": 2097152,
			"kb_used": 209715,
			"kb_avail": 1887437,
			"utilization": 10.0,
			"pgs": 32,
			"apply_latency_ms": 0.8,
			"commit_latency_ms": 1.1
		}
	]
}`

// TestOSDCollector_Describe 测试 OSD 采集器的指标描述符注册
// 验证 OSDCollector 注册了正确数量的指标描述符（9 个）:
//   - 2 个状态指标: up, in
//   - 4 个容量指标: total_bytes, used_bytes, available_bytes, utilization
//   - 3 个性能指标: pgs, apply_latency_ms, commit_latency_ms
func TestOSDCollector_Describe(t *testing.T) {
	log := newTestLogger()
	client := ceph.NewTestClient(log, nil)
	c := NewOSDCollector(client, log)

	descs := collectDescs(c)
	if len(descs) != 9 {
		t.Errorf("expected 9 descriptors, got %d", len(descs))
	}
}

// TestOSDCollector_Collect 测试 OSD 采集器的指标采集
// 验证:
//   - 采集到的指标总数（2 个 OSD * 9 个指标 = 18）
//   - osd.0 的 up 状态为 1（运行中）
//   - 空名称 OSD 自动命名为 "osd.1"
//   - KB 到字节的转换: 1048576 KB * 1024 = 1073741824 bytes
//   - 利用率值正确（50.0%）
func TestOSDCollector_Collect(t *testing.T) {
	log := newTestLogger()
	client := ceph.NewTestClient(log, func(args []byte) ([]byte, string, error) {
		return []byte(osdDFJSON), "", nil
	})
	c := NewOSDCollector(client, log)

	metrics := collectMetrics(c)
	// 2 OSDs * 9 metrics each = 18
	if len(metrics) != 18 {
		t.Fatalf("expected 18 metrics, got %d", len(metrics))
	}

	// Verify osd.0 metrics
	upMetrics := findMetricsByName(metrics, "osd_up")
	if len(upMetrics) != 2 {
		t.Fatalf("expected 2 osd_up metrics, got %d", len(upMetrics))
	}

	for _, m := range upMetrics {
		labels := getMetricLabels(m)
		if labels["osd"] == "osd.0" {
			if v := getMetricValue(m); v != 1 {
				t.Errorf("osd.0 up: expected 1, got %v", v)
			}
		}
	}

	// Verify OSD with empty name gets "osd.<id>" format
	found := false
	for _, m := range metrics {
		labels := getMetricLabels(m)
		if labels["osd"] == "osd.1" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected osd.1 label for OSD with empty name")
	}

	// Verify KB to bytes conversion: 1048576 KB * 1024 = 1073741824 bytes
	totalBytesMetrics := findMetricsByName(metrics, "osd_total_bytes")
	for _, m := range totalBytesMetrics {
		labels := getMetricLabels(m)
		if labels["osd"] == "osd.0" {
			expected := float64(1048576) * 1024
			if v := getMetricValue(m); v != expected {
				t.Errorf("osd.0 total_bytes: expected %v, got %v", expected, v)
			}
		}
	}

	// Verify utilization
	utilMetrics := findMetricsByName(metrics, "utilization")
	for _, m := range utilMetrics {
		labels := getMetricLabels(m)
		if labels["osd"] == "osd.0" {
			if v := getMetricValue(m); v != 50.0 {
				t.Errorf("osd.0 utilization: expected 50.0, got %v", v)
			}
		}
	}
}

// TestOSDCollector_CollectError 测试 Ceph 命令失败时的错误处理
// 验证采集器在命令失败时不会产生任何指标
func TestOSDCollector_CollectError(t *testing.T) {
	log := newTestLogger()
	client := ceph.NewTestClient(log, errMonCommand)
	c := NewOSDCollector(client, log)

	metrics := collectMetrics(c)
	if len(metrics) != 0 {
		t.Errorf("expected 0 metrics on error, got %d", len(metrics))
	}
}
