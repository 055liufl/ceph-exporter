package collector

import (
	"testing"

	"ceph-exporter/internal/ceph"
)

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

func TestClusterCollector_Describe(t *testing.T) {
	log := newTestLogger()
	client := ceph.NewTestClient(log, nil)
	c := NewClusterCollector(client, log)

	descs := collectDescs(c)
	if len(descs) != 15 {
		t.Errorf("expected 15 descriptors, got %d", len(descs))
	}
}

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

func TestClusterCollector_CollectError(t *testing.T) {
	log := newTestLogger()
	client := ceph.NewTestClient(log, errMonCommand)
	c := NewClusterCollector(client, log)

	metrics := collectMetrics(c)
	if len(metrics) != 0 {
		t.Errorf("expected 0 metrics on error, got %d", len(metrics))
	}
}
