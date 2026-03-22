// =============================================================================
// RGW Collector 单元测试
// =============================================================================
// 测试 RGW（RADOS Gateway）采集器的功能，包括:
//   - Describe: 验证注册的指标描述符数量（3 个）
//   - Collect: 验证 RGW 守护进程的状态统计
//   - NoRGWService: 验证没有 RGW 服务时的处理
//   - CollectError: 验证 Ceph 命令失败时的错误处理
//
// 测试数据说明:
//
//	serviceDumpJSON 模拟 "ceph service dump -f json" 命令的输出，包含:
//	- rgw.store1: 地址 10.0.0.1:7480
//	- rgw.store2: 地址 10.0.0.2:7480
//	两个 RGW 守护进程都被视为 active 状态
//
// 重点测试:
//   - 从 service dump 中正确解析 RGW 守护进程信息
//   - 没有 RGW 服务时返回 total=0
//   - daemon_status 指标的 name 标签
//
// =============================================================================
package collector

import (
	"testing"

	"ceph-exporter/internal/ceph"
)

// serviceDumpJSON 模拟 "ceph service dump -f json" 命令的输出
// 包含两个 RGW 守护进程的信息
const serviceDumpJSON = `{
	"services": {
		"rgw": {
			"daemons": {
				"rgw.store1": {
					"start_epoch": 100,
					"addr": "10.0.0.1:7480"
				},
				"rgw.store2": {
					"start_epoch": 101,
					"addr": "10.0.0.2:7480"
				}
			}
		}
	}
}`

// TestRGWCollector_Describe 测试 RGW 采集器的指标描述符注册
// 验证 RGWCollector 注册了正确数量的指标描述符（3 个）:
//   - total: RGW 守护进程总数
//   - active_total: active 状态的 RGW 数量
//   - daemon_status: 每个 RGW 守护进程的状态（带 name 标签）
func TestRGWCollector_Describe(t *testing.T) {
	log := newTestLogger()
	client := ceph.NewTestClient(log, nil)
	c := NewRGWCollector(client, log)

	descs := collectDescs(c)
	if len(descs) != 3 {
		t.Errorf("expected 3 descriptors, got %d", len(descs))
	}
}

// TestRGWCollector_Collect 测试 RGW 采集器的指标采集
// 验证:
//   - 采集到的指标总数: total(1) + active_total(1) + 2 个 daemon_status = 4
//   - total 值为 2（两个 RGW 守护进程）
//   - active_total 值为 2（所有守护进程都是 active）
func TestRGWCollector_Collect(t *testing.T) {
	log := newTestLogger()
	client := ceph.NewTestClient(log, func(args []byte) ([]byte, string, error) {
		return []byte(serviceDumpJSON), "", nil
	})
	c := NewRGWCollector(client, log)

	metrics := collectMetrics(c)
	// total(1) + active_total(1) + 2 daemon_status = 4
	if len(metrics) != 4 {
		t.Fatalf("expected 4 metrics, got %d", len(metrics))
	}

	// Verify total = 2
	totalMetrics := findMetricsByName(metrics, "rgw_total\"")
	if len(totalMetrics) != 1 {
		t.Fatalf("expected 1 rgw_total metric, got %d", len(totalMetrics))
	}
	if v := getMetricValue(totalMetrics[0]); v != 2 {
		t.Errorf("rgw_total: expected 2, got %v", v)
	}

	// Verify active_total = 2 (all daemons from service dump are active)
	activeMetrics := findMetricsByName(metrics, "active_total")
	if len(activeMetrics) != 1 {
		t.Fatalf("expected 1 active_total metric, got %d", len(activeMetrics))
	}
	if v := getMetricValue(activeMetrics[0]); v != 2 {
		t.Errorf("active_total: expected 2, got %v", v)
	}
}

// TestRGWCollector_NoRGWService 测试没有 RGW 服务时的处理
// 当 service dump 中没有 rgw 服务时，验证:
//   - 采集到的指标总数: total(1) + active_total(1) = 2（无 daemon_status）
//   - total 值为 0
func TestRGWCollector_NoRGWService(t *testing.T) {
	noRGW := `{"services": {}}`
	log := newTestLogger()
	client := ceph.NewTestClient(log, func(args []byte) ([]byte, string, error) {
		return []byte(noRGW), "", nil
	})
	c := NewRGWCollector(client, log)

	metrics := collectMetrics(c)
	// total(1) + active_total(1) + 0 daemons = 2
	if len(metrics) != 2 {
		t.Fatalf("expected 2 metrics, got %d", len(metrics))
	}

	totalMetrics := findMetricsByName(metrics, "rgw_total\"")
	if v := getMetricValue(totalMetrics[0]); v != 0 {
		t.Errorf("rgw_total with no service: expected 0, got %v", v)
	}
}

// TestRGWCollector_CollectError 测试 Ceph 命令失败时的错误处理
// 验证采集器在命令失败时不会产生任何指标
func TestRGWCollector_CollectError(t *testing.T) {
	log := newTestLogger()
	client := ceph.NewTestClient(log, errMonCommand)
	c := NewRGWCollector(client, log)

	metrics := collectMetrics(c)
	if len(metrics) != 0 {
		t.Errorf("expected 0 metrics on error, got %d", len(metrics))
	}
}
