// =============================================================================
// Health Collector 单元测试
// =============================================================================
// 测试健康状态采集器的功能，包括:
//   - Describe: 验证注册的指标描述符数量（4 个）
//   - Collect: 验证 HEALTH_WARN 状态下的指标值
//   - HealthOK: 验证 HEALTH_OK 状态下的指标值
//   - CollectError: 验证 Ceph 命令失败时的错误处理
//
// 测试数据说明:
//
//	healthStatusJSON 模拟 HEALTH_WARN 状态的集群，包含:
//	- 健康状态: HEALTH_WARN（对应状态码 1）
//	- 2 个健康检查项: TOO_FEW_OSDS 和 POOL_NO_REDUNDANCY
//	- 两个检查项的严重程度都是 HEALTH_WARN
//
// 重点测试:
//   - 健康状态字符串到数值的映射（HEALTH_OK=0, HEALTH_WARN=1, HEALTH_ERR=2）
//   - 健康检查项的 name 和 severity 标签
//   - HEALTH_OK 状态下检查项为空的情况
//
// =============================================================================
package collector

import (
	"testing"

	"ceph-exporter/internal/ceph"
)

// healthStatusJSON 模拟 HEALTH_WARN 状态的集群状态 JSON
// 包含两个健康检查项，用于测试 HealthCollector 的采集逻辑
const healthStatusJSON = `{
	"health": {
		"status": "HEALTH_WARN",
		"checks": {
			"TOO_FEW_OSDS": {
				"severity": "HEALTH_WARN",
				"summary": {"message": "OSD count 1 < osd_pool_default_size 3"}
			},
			"POOL_NO_REDUNDANCY": {
				"severity": "HEALTH_WARN",
				"summary": {"message": "pool has no redundancy"}
			}
		}
	},
	"pgmap": {"num_pgs": 0, "pgs_by_state": []},
	"osdmap": {},
	"monmap": {}
}`

// TestHealthCollector_Describe 测试健康状态采集器的指标描述符注册
// 验证 HealthCollector 注册了正确数量的指标描述符（4 个）:
//   - status: 健康状态码（数值型）
//   - status_info: 健康状态信息（带 status 标签）
//   - checks_total: 健康检查项总数
//   - check: 各健康检查项（带 name 和 severity 标签）
func TestHealthCollector_Describe(t *testing.T) {
	log := newTestLogger()
	client := ceph.NewTestClient(log, nil)
	c := NewHealthCollector(client, log)

	descs := collectDescs(c)
	if len(descs) != 4 {
		t.Errorf("expected 4 descriptors, got %d", len(descs))
	}
}

// TestHealthCollector_Collect 测试 HEALTH_WARN 状态下的指标采集
// 验证:
//   - 采集到的指标总数: status(1) + status_info(1) + checks_total(1) + check(2) = 5
//   - health_status 值为 1（HEALTH_WARN 对应的数值）
//   - checks_total 值为 2（两个检查项）
//   - 每个 check 指标的值为 1，severity 标签为 "HEALTH_WARN"
func TestHealthCollector_Collect(t *testing.T) {
	log := newTestLogger()
	client := ceph.NewTestClient(log, func(args []byte) ([]byte, string, error) {
		return []byte(healthStatusJSON), "", nil
	})
	c := NewHealthCollector(client, log)

	metrics := collectMetrics(c)
	// status(1) + status_info(1) + checks_total(1) + check(2) = 5
	if len(metrics) != 5 {
		t.Fatalf("expected 5 metrics, got %d", len(metrics))
	}

	// health_status should be 1 (HEALTH_WARN)
	statusMetrics := findMetricsByName(metrics, "health_status\"")
	if len(statusMetrics) == 0 {
		t.Fatal("health_status metric not found")
	}
	if v := getMetricValue(statusMetrics[0]); v != 1 {
		t.Errorf("health_status: expected 1, got %v", v)
	}

	// checks_total should be 2
	checksTotal := findMetricsByName(metrics, "checks_total")
	if len(checksTotal) == 0 {
		t.Fatal("checks_total metric not found")
	}
	if v := getMetricValue(checksTotal[0]); v != 2 {
		t.Errorf("checks_total: expected 2, got %v", v)
	}

	// Each check metric should have value 1
	checkMetrics := findMetricsByName(metrics, "health_check\"")
	if len(checkMetrics) != 2 {
		t.Fatalf("expected 2 health_check metrics, got %d", len(checkMetrics))
	}
	for _, m := range checkMetrics {
		if v := getMetricValue(m); v != 1 {
			t.Errorf("health_check value: expected 1, got %v", v)
		}
		labels := getMetricLabels(m)
		if labels["severity"] != "HEALTH_WARN" {
			t.Errorf("health_check severity: expected HEALTH_WARN, got %s", labels["severity"])
		}
	}
}

// TestHealthCollector_HealthOK 测试 HEALTH_OK 状态下的指标采集
// 验证:
//   - 采集到的指标总数: status(1) + status_info(1) + checks_total(1) = 3（无检查项）
//   - health_status 值为 0（HEALTH_OK 对应的数值）
func TestHealthCollector_HealthOK(t *testing.T) {
	json := `{
		"health": {"status": "HEALTH_OK", "checks": {}},
		"pgmap": {"num_pgs": 0, "pgs_by_state": []},
		"osdmap": {},
		"monmap": {}
	}`
	log := newTestLogger()
	client := ceph.NewTestClient(log, func(args []byte) ([]byte, string, error) {
		return []byte(json), "", nil
	})
	c := NewHealthCollector(client, log)

	metrics := collectMetrics(c)
	// status(1) + status_info(1) + checks_total(1) + 0 checks = 3
	if len(metrics) != 3 {
		t.Fatalf("expected 3 metrics, got %d", len(metrics))
	}

	statusMetrics := findMetricsByName(metrics, "health_status\"")
	if v := getMetricValue(statusMetrics[0]); v != 0 {
		t.Errorf("health_status for HEALTH_OK: expected 0, got %v", v)
	}
}

// TestHealthCollector_CollectError 测试 Ceph 命令失败时的错误处理
// 验证采集器在命令失败时不会产生任何指标
func TestHealthCollector_CollectError(t *testing.T) {
	log := newTestLogger()
	client := ceph.NewTestClient(log, errMonCommand)
	c := NewHealthCollector(client, log)

	metrics := collectMetrics(c)
	if len(metrics) != 0 {
		t.Errorf("expected 0 metrics on error, got %d", len(metrics))
	}
}
