package collector

import (
	"testing"

	"ceph-exporter/internal/ceph"
)

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

func TestMonitorCollector_Describe(t *testing.T) {
	log := newTestLogger()
	client := ceph.NewTestClient(log, nil)
	c := NewMonitorCollector(client, log)

	descs := collectDescs(c)
	if len(descs) != 4 {
		t.Errorf("expected 4 descriptors, got %d", len(descs))
	}
}

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

func TestMonitorCollector_CollectError(t *testing.T) {
	log := newTestLogger()
	client := ceph.NewTestClient(log, errMonCommand)
	c := NewMonitorCollector(client, log)

	metrics := collectMetrics(c)
	if len(metrics) != 0 {
		t.Errorf("expected 0 metrics on error, got %d", len(metrics))
	}
}
