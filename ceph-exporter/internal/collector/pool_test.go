// =============================================================================
// Pool Collector 单元测试
// =============================================================================
// 测试存储池采集器的功能，包括:
//   - Describe: 验证注册的指标描述符数量（9 个）
//   - Collect: 验证每个存储池的指标值和标签
//   - CollectError: 验证 Ceph 命令失败时的错误处理
//
// 测试数据说明:
//
//	poolStatsJSON 模拟 "ceph osd pool stats -f json" 命令的输出，包含:
//	- rbd 存储池: 1GB 已存储，256 个对象，25% 使用率，有 IO 活动
//	- cephfs_data 存储池: 512MB 已存储，128 个对象，12% 使用率，无 IO 活动
//
// =============================================================================
package collector

import (
	"testing"

	"ceph-exporter/internal/ceph"
)

// poolStatsJSON 模拟 "ceph osd pool stats -f json" 命令的输出
// 包含两个存储池的完整统计信息，用于测试 PoolCollector 的采集逻辑
const poolStatsJSON = `[
	{
		"pool_name": "rbd",
		"pool_id": 1,
		"stats": {
			"stored": 1073741824,
			"objects": 256,
			"max_avail": 3221225472,
			"bytes_used": 2147483648,
			"percent_used": 0.25
		},
		"client_io_rate": {
			"read_bytes_sec": 524288,
			"write_bytes_sec": 1048576,
			"read_op_per_sec": 50,
			"write_op_per_sec": 100
		}
	},
	{
		"pool_name": "cephfs_data",
		"pool_id": 2,
		"stats": {
			"stored": 536870912,
			"objects": 128,
			"max_avail": 3221225472,
			"bytes_used": 1073741824,
			"percent_used": 0.12
		},
		"client_io_rate": {
			"read_bytes_sec": 0,
			"write_bytes_sec": 0,
			"read_op_per_sec": 0,
			"write_op_per_sec": 0
		}
	}
]`

// TestPoolCollector_Describe 测试存储池采集器的指标描述符注册
// 验证 PoolCollector 注册了正确数量的指标描述符（9 个）:
//   - 5 个容量指标: stored_bytes, max_available_bytes, used_bytes, percent_used, objects_total
//   - 4 个 IO 指标: read_bytes_sec, write_bytes_sec, read_ops_sec, write_ops_sec
func TestPoolCollector_Describe(t *testing.T) {
	log := newTestLogger()
	client := ceph.NewTestClient(log, nil)
	c := NewPoolCollector(client, log)

	descs := collectDescs(c)
	if len(descs) != 9 {
		t.Errorf("expected 9 descriptors, got %d", len(descs))
	}
}

// TestPoolCollector_Collect 测试存储池采集器的指标采集
// 验证:
//   - 采集到的指标总数（2 个存储池 * 9 个指标 = 18）
//   - rbd 存储池的 stored_bytes 值是否正确（1073741824 = 1GB）
//   - 所有指标都带有 pool 标签
func TestPoolCollector_Collect(t *testing.T) {
	log := newTestLogger()
	client := ceph.NewTestClient(log, func(args []byte) ([]byte, string, error) {
		return []byte(poolStatsJSON), "", nil
	})
	c := NewPoolCollector(client, log)

	metrics := collectMetrics(c)
	// 2 pools * 9 metrics each = 18
	if len(metrics) != 18 {
		t.Fatalf("expected 18 metrics, got %d", len(metrics))
	}

	// Verify rbd pool stored_bytes
	storedMetrics := findMetricsByName(metrics, "stored_bytes")
	if len(storedMetrics) != 2 {
		t.Fatalf("expected 2 stored_bytes metrics, got %d", len(storedMetrics))
	}
	for _, m := range storedMetrics {
		labels := getMetricLabels(m)
		if labels["pool"] == "rbd" {
			if v := getMetricValue(m); v != 1073741824 {
				t.Errorf("rbd stored_bytes: expected 1073741824, got %v", v)
			}
		}
	}

	// Verify pool labels exist
	for _, m := range metrics {
		labels := getMetricLabels(m)
		if _, ok := labels["pool"]; !ok {
			t.Error("pool label missing from metric")
		}
	}
}

// TestPoolCollector_CollectError 测试 Ceph 命令失败时的错误处理
// 验证采集器在命令失败时不会产生任何指标
func TestPoolCollector_CollectError(t *testing.T) {
	log := newTestLogger()
	client := ceph.NewTestClient(log, errMonCommand)
	c := NewPoolCollector(client, log)

	metrics := collectMetrics(c)
	if len(metrics) != 0 {
		t.Errorf("expected 0 metrics on error, got %d", len(metrics))
	}
}
