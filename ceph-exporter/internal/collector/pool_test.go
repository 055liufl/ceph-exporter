package collector

import (
	"testing"

	"ceph-exporter/internal/ceph"
)

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

func TestPoolCollector_Describe(t *testing.T) {
	log := newTestLogger()
	client := ceph.NewTestClient(log, nil)
	c := NewPoolCollector(client, log)

	descs := collectDescs(c)
	if len(descs) != 9 {
		t.Errorf("expected 9 descriptors, got %d", len(descs))
	}
}

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

func TestPoolCollector_CollectError(t *testing.T) {
	log := newTestLogger()
	client := ceph.NewTestClient(log, errMonCommand)
	c := NewPoolCollector(client, log)

	metrics := collectMetrics(c)
	if len(metrics) != 0 {
		t.Errorf("expected 0 metrics on error, got %d", len(metrics))
	}
}
