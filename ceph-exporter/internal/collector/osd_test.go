package collector

import (
	"testing"

	"ceph-exporter/internal/ceph"
)

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

func TestOSDCollector_Describe(t *testing.T) {
	log := newTestLogger()
	client := ceph.NewTestClient(log, nil)
	c := NewOSDCollector(client, log)

	descs := collectDescs(c)
	if len(descs) != 9 {
		t.Errorf("expected 9 descriptors, got %d", len(descs))
	}
}

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

func TestOSDCollector_CollectError(t *testing.T) {
	log := newTestLogger()
	client := ceph.NewTestClient(log, errMonCommand)
	c := NewOSDCollector(client, log)

	metrics := collectMetrics(c)
	if len(metrics) != 0 {
		t.Errorf("expected 0 metrics on error, got %d", len(metrics))
	}
}
