package collector

import (
	"testing"

	"ceph-exporter/internal/ceph"
)

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

func TestMDSCollector_Describe(t *testing.T) {
	log := newTestLogger()
	client := ceph.NewTestClient(log, nil)
	c := NewMDSCollector(client, log)

	descs := collectDescs(c)
	if len(descs) != 3 {
		t.Errorf("expected 3 descriptors, got %d", len(descs))
	}
}

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

func TestMDSCollector_CollectError(t *testing.T) {
	log := newTestLogger()
	client := ceph.NewTestClient(log, errMonCommand)
	c := NewMDSCollector(client, log)

	metrics := collectMetrics(c)
	if len(metrics) != 0 {
		t.Errorf("expected 0 metrics on error, got %d", len(metrics))
	}
}
