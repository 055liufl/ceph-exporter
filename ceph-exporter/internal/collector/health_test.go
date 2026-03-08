package collector

import (
	"testing"

	"ceph-exporter/internal/ceph"
)

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

func TestHealthCollector_Describe(t *testing.T) {
	log := newTestLogger()
	client := ceph.NewTestClient(log, nil)
	c := NewHealthCollector(client, log)

	descs := collectDescs(c)
	if len(descs) != 4 {
		t.Errorf("expected 4 descriptors, got %d", len(descs))
	}
}

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

func TestHealthCollector_CollectError(t *testing.T) {
	log := newTestLogger()
	client := ceph.NewTestClient(log, errMonCommand)
	c := NewHealthCollector(client, log)

	metrics := collectMetrics(c)
	if len(metrics) != 0 {
		t.Errorf("expected 0 metrics on error, got %d", len(metrics))
	}
}
