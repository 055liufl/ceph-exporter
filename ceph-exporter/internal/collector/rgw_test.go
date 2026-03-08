package collector

import (
	"testing"

	"ceph-exporter/internal/ceph"
)

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

func TestRGWCollector_Describe(t *testing.T) {
	log := newTestLogger()
	client := ceph.NewTestClient(log, nil)
	c := NewRGWCollector(client, log)

	descs := collectDescs(c)
	if len(descs) != 3 {
		t.Errorf("expected 3 descriptors, got %d", len(descs))
	}
}

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

func TestRGWCollector_CollectError(t *testing.T) {
	log := newTestLogger()
	client := ceph.NewTestClient(log, errMonCommand)
	c := NewRGWCollector(client, log)

	metrics := collectMetrics(c)
	if len(metrics) != 0 {
		t.Errorf("expected 0 metrics on error, got %d", len(metrics))
	}
}
