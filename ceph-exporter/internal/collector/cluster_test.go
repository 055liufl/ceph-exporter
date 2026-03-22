// =============================================================================
// Cluster Collector 单元测试
// =============================================================================
// 测试集群状态采集器的功能，包括:
//   - Describe: 验证注册的指标描述符数量（15 个）
//   - Collect: 验证采集到的指标值是否与 Mock 数据一致
//   - CollectError: 验证 Ceph 命令失败时的错误处理
//
// 测试数据说明:
//
//	clusterStatusJSON 模拟 "ceph status -f json" 命令的输出，包含:
//	- 健康状态: HEALTH_OK
//	- PG 信息: 128 个 PG，分为 active+clean(120) 和 active+undersized(8)
//	- 容量信息: 总容量 3GB，已用 512MB，可用 2GB
//	- IO 信息: 读 1MB/s，写 2MB/s，读 100 IOPS，写 200 IOPS
//	- OSD 信息: 6 个 OSD，5 个 Up，6 个 In
//	- Monitor 信息: 3 个 Monitor
//
// =============================================================================
package collector

import (
	"testing"

	"ceph-exporter/internal/ceph"
)

// clusterStatusJSON 模拟 "ceph status -f json" 命令的完整输出
// 用于测试 ClusterCollector 的指标采集逻辑
const clusterStatusJSON = `{
	"health": {
		"status": "HEALTH_OK",
		"checks": {}
	},
	"pgmap": {
		"num_pgs": 128,
		"num_pools": 3,
		"data_bytes": 1073741824,
		"bytes_used": 536870912,
		"bytes_avail": 2147483648,
		"bytes_total": 3221225472,
		"read_bytes_sec": 1048576,
		"write_bytes_sec": 2097152,
		"read_op_per_sec": 100,
		"write_op_per_sec": 200,
		"num_objects": 5000,
		"pgs_by_state": [
			{"state_name": "active+clean", "count": 120},
			{"state_name": "active+undersized", "count": 8}
		]
	},
	"osdmap": {
		"num_osds": 6,
		"num_up_osds": 5,
		"num_in_osds": 6
	},
	"monmap": {
		"num_mons": 3
	}
}`

// TestClusterCollector_Describe 测试集群采集器的指标描述符注册
// 验证 ClusterCollector 注册了正确数量的指标描述符（15 个）:
//   - 4 个容量指标: total_bytes, used_bytes, available_bytes, objects_total
//   - 4 个 IO 指标: read_bytes_sec, write_bytes_sec, read_ops_sec, write_ops_sec
//   - 2 个 PG 指标: pgs_total, pgs_by_state
//   - 5 个组件指标: pools_total, osds_total, osds_up, osds_in, mons_total
func TestClusterCollector_Describe(t *testing.T) {
	log := newTestLogger()
	client := ceph.NewTestClient(log, nil)
	c := NewClusterCollector(client, log)

	descs := collectDescs(c)
	if len(descs) != 15 {
		t.Errorf("expected 15 descriptors, got %d", len(descs))
	}
}

// TestClusterCollector_Collect 测试集群采集器的指标采集
// 使用 Mock 客户端返回预定义的 JSON 数据，验证:
//   - 采集到的指标总数（14 个固定指标 + 2 个 pgs_by_state = 16）
//   - 每个指标的值是否与 Mock 数据一致
//   - pgs_by_state 指标的标签和值是否正确
func TestClusterCollector_Collect(t *testing.T) {
	log := newTestLogger()
	client := ceph.NewTestClient(log, func(args []byte) ([]byte, string, error) {
		return []byte(clusterStatusJSON), "", nil
	})
	c := NewClusterCollector(client, log)

	metrics := collectMetrics(c)
	// 14 固定指标 + 2 个 pgs_by_state = 16
	if len(metrics) != 16 {
		t.Fatalf("expected 16 metrics, got %d", len(metrics))
	}

	tests := []struct {
		name  string
		value float64
	}{
		{"total_bytes", 3221225472},
		{"used_bytes", 536870912},
		{"available_bytes", 2147483648},
		{"objects_total", 5000},
		{"read_bytes_sec", 1048576},
		{"write_bytes_sec", 2097152},
		{"read_ops_sec", 100},
		{"write_ops_sec", 200},
		{"pgs_total", 128},
		{"pools_total", 3},
		{"osds_total", 6},
		{"osds_up", 5},
		{"osds_in", 6},
		{"mons_total", 3},
	}

	for _, tt := range tests {
		found := findMetricsByName(metrics, tt.name)
		if len(found) == 0 {
			t.Errorf("metric %s not found", tt.name)
			continue
		}
		if v := getMetricValue(found[0]); v != tt.value {
			t.Errorf("metric %s: expected %v, got %v", tt.name, tt.value, v)
		}
	}

	// 验证 pgs_by_state 标签
	pgsByState := findMetricsByName(metrics, "pgs_by_state")
	if len(pgsByState) != 2 {
		t.Fatalf("expected 2 pgs_by_state metrics, got %d", len(pgsByState))
	}
	stateValues := map[string]float64{}
	for _, m := range pgsByState {
		labels := getMetricLabels(m)
		stateValues[labels["state"]] = getMetricValue(m)
	}
	if stateValues["active+clean"] != 120 {
		t.Errorf("pgs_by_state active+clean: expected 120, got %v", stateValues["active+clean"])
	}
	if stateValues["active+undersized"] != 8 {
		t.Errorf("pgs_by_state active+undersized: expected 8, got %v", stateValues["active+undersized"])
	}
}

// TestClusterCollector_CollectError 测试 Ceph 命令失败时的错误处理
// 使用 errMonCommand 模拟命令执行失败，验证:
//   - 采集器不会 panic
//   - 不会产生任何指标（返回空列表）
//   - 错误会被记录到日志中（不在此测试中验证）
func TestClusterCollector_CollectError(t *testing.T) {
	log := newTestLogger()
	client := ceph.NewTestClient(log, errMonCommand)
	c := NewClusterCollector(client, log)

	metrics := collectMetrics(c)
	if len(metrics) != 0 {
		t.Errorf("expected 0 metrics on error, got %d", len(metrics))
	}
}
