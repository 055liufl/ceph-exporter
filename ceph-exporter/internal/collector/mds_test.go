// =============================================================================
// MDS Collector 单元测试
// =============================================================================
// 测试 MDS（Metadata Server）采集器的功能，包括:
//   - Describe: 验证注册的指标描述符数量（3 个）
//   - Collect: 验证 MDS 守护进程的状态统计和标签
//   - CollectError: 验证 Ceph 命令失败时的错误处理
//
// 测试数据说明:
//
//	mdsStatJSON 模拟 "ceph mds stat -f json" 命令的输出，包含:
//	- mds.a: active 状态，rank 0（正在提供元数据服务）
//	- mds.b: standby 状态（待命，可接管 active MDS）
//	- mds.c: standby 状态（待命）
//
// 重点测试:
//   - active 和 standby MDS 的正确统计
//   - daemon_status 指标的 name 和 state 标签
//   - standby MDS 的状态自动设置为 "up:standby"
//
// =============================================================================
package collector

import (
	"testing"

	"ceph-exporter/internal/ceph"
)

// mdsStatJSON 模拟 "ceph mds stat -f json" 命令的输出
// 包含 1 个 active MDS 和 2 个 standby MDS
const mdsStatJSON = `{
	"fsmap": {
		"up": {"mds_0": 12345},
		"info": {
			"gid_12345": {
				"name": "mds.a",
				"state": "up:active",
				"rank": 0
			}
		},
		"standbys": [
			{"name": "mds.b"},
			{"name": "mds.c"}
		]
	}
}`

// TestMDSCollector_Describe 测试 MDS 采集器的指标描述符注册
// 验证 MDSCollector 注册了正确数量的指标描述符（3 个）:
//   - active_total: active 状态的 MDS 数量
//   - standby_total: standby 状态的 MDS 数量
//   - daemon_status: 每个 MDS 守护进程的状态（带 name 和 state 标签）
func TestMDSCollector_Describe(t *testing.T) {
	log := newTestLogger()
	client := ceph.NewTestClient(log, nil)
	c := NewMDSCollector(client, log)

	descs := collectDescs(c)
	if len(descs) != 3 {
		t.Errorf("expected 3 descriptors, got %d", len(descs))
	}
}

// TestMDSCollector_Collect 测试 MDS 采集器的指标采集
// 验证:
//   - 采集到的指标总数: 3 个 daemon_status + active_total(1) + standby_total(1) = 5
//   - active_total 值为 1（只有 mds.a 是 active）
//   - standby_total 值为 2（mds.b 和 mds.c 是 standby）
//   - mds.a 的 state 标签为 "up:active"
//   - mds.b 的 state 标签为 "up:standby"（自动设置）
func TestMDSCollector_Collect(t *testing.T) {
	log := newTestLogger()
	client := ceph.NewTestClient(log, func(args []byte) ([]byte, string, error) {
		return []byte(mdsStatJSON), "", nil
	})
	c := NewMDSCollector(client, log)

	metrics := collectMetrics(c)
	// 3 daemon_status + active_total(1) + standby_total(1) = 5
	if len(metrics) != 5 {
		t.Fatalf("expected 5 metrics, got %d", len(metrics))
	}

	// Verify active_total = 1
	activeMetrics := findMetricsByName(metrics, "active_total")
	if len(activeMetrics) != 1 {
		t.Fatalf("expected 1 active_total metric, got %d", len(activeMetrics))
	}
	if v := getMetricValue(activeMetrics[0]); v != 1 {
		t.Errorf("active_total: expected 1, got %v", v)
	}

	// Verify standby_total = 2
	standbyMetrics := findMetricsByName(metrics, "standby_total")
	if len(standbyMetrics) != 1 {
		t.Fatalf("expected 1 standby_total metric, got %d", len(standbyMetrics))
	}
	if v := getMetricValue(standbyMetrics[0]); v != 2 {
		t.Errorf("standby_total: expected 2, got %v", v)
	}

	// Verify daemon_status labels
	daemonMetrics := findMetricsByName(metrics, "daemon_status")
	if len(daemonMetrics) != 3 {
		t.Fatalf("expected 3 daemon_status metrics, got %d", len(daemonMetrics))
	}
	states := map[string]string{}
	for _, m := range daemonMetrics {
		labels := getMetricLabels(m)
		states[labels["name"]] = labels["state"]
	}
	if states["mds.a"] != "up:active" {
		t.Errorf("mds.a state: expected up:active, got %s", states["mds.a"])
	}
	if states["mds.b"] != "up:standby" {
		t.Errorf("mds.b state: expected up:standby, got %s", states["mds.b"])
	}
}

// TestMDSCollector_CollectError 测试 Ceph 命令失败时的错误处理
// 验证采集器在命令失败时不会产生任何指标
func TestMDSCollector_CollectError(t *testing.T) {
	log := newTestLogger()
	client := ceph.NewTestClient(log, errMonCommand)
	c := NewMDSCollector(client, log)

	metrics := collectMetrics(c)
	if len(metrics) != 0 {
		t.Errorf("expected 0 metrics on error, got %d", len(metrics))
	}
}
