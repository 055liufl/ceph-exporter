// =============================================================================
// Ceph 客户端单元测试
// =============================================================================
// 由于单元测试环境中没有真实的 Ceph 集群，本文件主要测试:
//   - 客户端创建和初始化
//   - 连接状态管理（IsConnected、Close）
//   - JSON 数据结构的序列化/反序列化
//   - 命令 JSON 构建
//   - 未连接时的错误处理
//
// 注意: 涉及真实 Ceph 连接的测试需要在集成测试环境中执行
// =============================================================================
package ceph

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"ceph-exporter/internal/config"
	"ceph-exporter/internal/logger"
)

// newTestLogger 创建用于测试的日志实例
func newTestLogger(t *testing.T) *logger.Logger {
	t.Helper()
	cfg := &config.LoggerConfig{
		Level:  "debug",
		Format: "text",
		Output: "stdout",
	}
	log, err := logger.NewLogger(cfg)
	if err != nil {
		t.Fatalf("创建测试日志失败: %v", err)
	}
	return log
}

// newTestCephConfig 创建用于测试的 Ceph 配置
func newTestCephConfig() *config.CephConfig {
	return &config.CephConfig{
		ConfigFile: "/etc/ceph/ceph.conf",
		User:       "admin",
		Keyring:    "/etc/ceph/ceph.client.admin.keyring",
		Cluster:    "ceph",
		Timeout:    10 * time.Second,
	}
}

// TestNewClient 测试客户端创建
func TestNewClient(t *testing.T) {
	log := newTestLogger(t)
	cfg := newTestCephConfig()

	client, err := NewClient(cfg, log)
	if err != nil {
		t.Fatalf("创建客户端失败: %v", err)
	}

	// 验证客户端初始状态
	if client == nil {
		t.Fatal("客户端实例为 nil")
	}
	if client.closed {
		t.Error("新创建的客户端不应该是关闭状态")
	}
	if client.conn != nil {
		t.Error("新创建的客户端不应该有连接")
	}
}

// TestClient_IsConnected_NotConnected 测试未连接时的状态
func TestClient_IsConnected_NotConnected(t *testing.T) {
	log := newTestLogger(t)
	cfg := newTestCephConfig()

	client, _ := NewClient(cfg, log)

	// 未调用 Connect() 时应该返回 false
	if client.IsConnected() {
		t.Error("未连接的客户端 IsConnected() 应该返回 false")
	}
}

// TestClient_Close_NotConnected 测试关闭未连接的客户端
func TestClient_Close_NotConnected(t *testing.T) {
	log := newTestLogger(t)
	cfg := newTestCephConfig()

	client, _ := NewClient(cfg, log)

	// 关闭未连接的客户端不应该 panic
	client.Close()

	// 关闭后状态应该正确
	if client.IsConnected() {
		t.Error("关闭后 IsConnected() 应该返回 false")
	}
}

// TestClient_Close_Multiple 测试多次关闭客户端（幂等性）
func TestClient_Close_Multiple(t *testing.T) {
	log := newTestLogger(t)
	cfg := newTestCephConfig()

	client, _ := NewClient(cfg, log)

	// 多次关闭不应该 panic
	client.Close()
	client.Close()
	client.Close()
}

// TestClient_ExecuteCommand_NotConnected 测试未连接时执行命令
func TestClient_ExecuteCommand_NotConnected(t *testing.T) {
	log := newTestLogger(t)
	cfg := newTestCephConfig()

	client, _ := NewClient(cfg, log)

	// 构建测试命令
	cmd, _ := json.Marshal(map[string]interface{}{
		"prefix": "status",
		"format": "json",
	})

	// 未连接时执行命令应该返回错误
	ctx := context.Background()
	_, err := client.ExecuteCommand(ctx, cmd)
	if err == nil {
		t.Fatal("未连接时执行命令应该返回错误")
	}
}

// TestClusterStatus_JSONParsing 测试 ClusterStatus JSON 反序列化
func TestClusterStatus_JSONParsing(t *testing.T) {
	// 模拟 "ceph status -f json" 的输出
	jsonData := `{
		"health": {
			"status": "HEALTH_OK",
			"checks": {}
		},
		"pgmap": {
			"num_pgs": 128,
			"num_pools": 3,
			"data_bytes": 1073741824,
			"bytes_used": 3221225472,
			"bytes_avail": 107374182400,
			"bytes_total": 110595407872,
			"read_bytes_sec": 1048576,
			"write_bytes_sec": 2097152,
			"read_op_per_sec": 100,
			"write_op_per_sec": 50,
			"num_objects": 1000,
			"pgs_by_state": [
				{"state_name": "active+clean", "count": 128}
			]
		},
		"osdmap": {
			"num_osds": 6,
			"num_up_osds": 6,
			"num_in_osds": 6
		},
		"monmap": {
			"num_mons": 3
		}
	}`

	var status ClusterStatus
	if err := json.Unmarshal([]byte(jsonData), &status); err != nil {
		t.Fatalf("解析 ClusterStatus JSON 失败: %v", err)
	}

	// 验证健康状态
	if status.Health.Status != "HEALTH_OK" {
		t.Errorf("Health.Status 期望 'HEALTH_OK'，实际 '%s'", status.Health.Status)
	}

	// 验证 PG 信息
	if status.PGMap.NumPGs != 128 {
		t.Errorf("PGMap.NumPGs 期望 128，实际 %d", status.PGMap.NumPGs)
	}
	if status.PGMap.NumPools != 3 {
		t.Errorf("PGMap.NumPools 期望 3，实际 %d", status.PGMap.NumPools)
	}
	if status.PGMap.DataBytes != 1073741824 {
		t.Errorf("PGMap.DataBytes 期望 1073741824，实际 %d", status.PGMap.DataBytes)
	}
	if status.PGMap.ReadBytesSec != 1048576 {
		t.Errorf("PGMap.ReadBytesSec 期望 1048576，实际 %d", status.PGMap.ReadBytesSec)
	}
	if status.PGMap.NumObjects != 1000 {
		t.Errorf("PGMap.NumObjects 期望 1000，实际 %d", status.PGMap.NumObjects)
	}

	// 验证 PG 状态分布
	if len(status.PGMap.PGsByState) != 1 {
		t.Fatalf("PGsByState 期望 1 个条目，实际 %d", len(status.PGMap.PGsByState))
	}
	if status.PGMap.PGsByState[0].StateName != "active+clean" {
		t.Errorf("PGsByState[0].StateName 期望 'active+clean'，实际 '%s'", status.PGMap.PGsByState[0].StateName)
	}
	if status.PGMap.PGsByState[0].Count != 128 {
		t.Errorf("PGsByState[0].Count 期望 128，实际 %d", status.PGMap.PGsByState[0].Count)
	}

	// 验证 OSD 信息
	if status.OSDMap.NumOSDs != 6 {
		t.Errorf("OSDMap.NumOSDs 期望 6，实际 %d", status.OSDMap.NumOSDs)
	}
	if status.OSDMap.NumUpOSDs != 6 {
		t.Errorf("OSDMap.NumUpOSDs 期望 6，实际 %d", status.OSDMap.NumUpOSDs)
	}

	// 验证 Monitor 信息
	if status.MonMap.NumMons != 3 {
		t.Errorf("MonMap.NumMons 期望 3，实际 %d", status.MonMap.NumMons)
	}
}

// TestPoolStats_JSONParsing 测试 PoolStats JSON 反序列化
func TestPoolStats_JSONParsing(t *testing.T) {
	jsonData := `[
		{
			"pool_name": "rbd",
			"pool_id": 1,
			"stats": {
				"stored": 536870912,
				"objects": 500,
				"max_avail": 53687091200,
				"bytes_used": 1610612736,
				"percent_used": 0.03
			},
			"client_io_rate": {
				"read_bytes_sec": 524288,
				"write_bytes_sec": 1048576,
				"read_op_per_sec": 50,
				"write_op_per_sec": 25
			}
		}
	]`

	var stats []PoolStats
	if err := json.Unmarshal([]byte(jsonData), &stats); err != nil {
		t.Fatalf("解析 PoolStats JSON 失败: %v", err)
	}

	if len(stats) != 1 {
		t.Fatalf("期望 1 个存储池，实际 %d", len(stats))
	}

	pool := stats[0]
	if pool.PoolName != "rbd" {
		t.Errorf("PoolName 期望 'rbd'，实际 '%s'", pool.PoolName)
	}
	if pool.PoolID != 1 {
		t.Errorf("PoolID 期望 1，实际 %d", pool.PoolID)
	}
	if pool.Stats.Stored != 536870912 {
		t.Errorf("Stats.Stored 期望 536870912，实际 %d", pool.Stats.Stored)
	}
	if pool.Stats.Objects != 500 {
		t.Errorf("Stats.Objects 期望 500，实际 %d", pool.Stats.Objects)
	}
	if pool.ClientIORate.ReadBytesSec != 524288 {
		t.Errorf("ClientIORate.ReadBytesSec 期望 524288，实际 %d", pool.ClientIORate.ReadBytesSec)
	}
}

// TestOSDStats_JSONParsing 测试 OSDStats JSON 反序列化
func TestOSDStats_JSONParsing(t *testing.T) {
	// 模拟 "ceph osd df -f json" 输出中的 nodes 部分
	jsonData := `{
		"nodes": [
			{
				"id": 0,
				"name": "osd.0",
				"up": 1,
				"in": 1,
				"kb": 104857600,
				"kb_used": 31457280,
				"kb_avail": 73400320,
				"utilization": 30.0,
				"pgs": 42,
				"status": "up",
				"apply_latency_ms": 1.5,
				"commit_latency_ms": 0.8
			},
			{
				"id": 1,
				"name": "osd.1",
				"up": 1,
				"in": 1,
				"kb": 104857600,
				"kb_used": 20971520,
				"kb_avail": 83886080,
				"utilization": 20.0,
				"pgs": 43,
				"status": "up",
				"apply_latency_ms": 1.2,
				"commit_latency_ms": 0.6
			}
		]
	}`

	var osdDF struct {
		Nodes []OSDStats `json:"nodes"`
	}
	if err := json.Unmarshal([]byte(jsonData), &osdDF); err != nil {
		t.Fatalf("解析 OSD 统计 JSON 失败: %v", err)
	}

	if len(osdDF.Nodes) != 2 {
		t.Fatalf("期望 2 个 OSD，实际 %d", len(osdDF.Nodes))
	}

	osd0 := osdDF.Nodes[0]
	if osd0.ID != 0 {
		t.Errorf("OSD[0].ID 期望 0，实际 %d", osd0.ID)
	}
	if osd0.Name != "osd.0" {
		t.Errorf("OSD[0].Name 期望 'osd.0'，实际 '%s'", osd0.Name)
	}
	if osd0.Up() != 1 {
		t.Errorf("OSD[0].Up() 期望 1，实际 %d", osd0.Up())
	}
	if osd0.Utilization != 30.0 {
		t.Errorf("OSD[0].Utilization 期望 30.0，实际 %f", osd0.Utilization)
	}
	if osd0.ApplyLatencyMs != 1.5 {
		t.Errorf("OSD[0].ApplyLatencyMs 期望 1.5，实际 %f", osd0.ApplyLatencyMs)
	}
}

// TestMonitorStats_JSONParsing 测试 MonitorStats JSON 反序列化
func TestMonitorStats_JSONParsing(t *testing.T) {
	jsonData := `{
		"mons": [
			{
				"name": "mon.a",
				"rank": 0,
				"addr": "192.168.1.10:6789/0",
				"store_bytes": 104857600,
				"clock_skew": 0.001,
				"latency": 0.5,
				"in_quorum": true
			},
			{
				"name": "mon.b",
				"rank": 1,
				"addr": "192.168.1.11:6789/0",
				"store_bytes": 104857600,
				"clock_skew": 0.002,
				"latency": 0.8,
				"in_quorum": true
			}
		]
	}`

	var monDump struct {
		Mons []MonitorStats `json:"mons"`
	}
	if err := json.Unmarshal([]byte(jsonData), &monDump); err != nil {
		t.Fatalf("解析 Monitor 统计 JSON 失败: %v", err)
	}

	if len(monDump.Mons) != 2 {
		t.Fatalf("期望 2 个 Monitor，实际 %d", len(monDump.Mons))
	}

	mon0 := monDump.Mons[0]
	if mon0.Name != "mon.a" {
		t.Errorf("Mon[0].Name 期望 'mon.a'，实际 '%s'", mon0.Name)
	}
	if mon0.Rank != 0 {
		t.Errorf("Mon[0].Rank 期望 0，实际 %d", mon0.Rank)
	}
	if !mon0.InQuorum {
		t.Error("Mon[0].InQuorum 期望 true")
	}
	if mon0.ClockSkew != 0.001 {
		t.Errorf("Mon[0].ClockSkew 期望 0.001，实际 %f", mon0.ClockSkew)
	}
}

// TestCommandJSON_Build 测试命令 JSON 构建
func TestCommandJSON_Build(t *testing.T) {
	// 测试各种 Ceph 命令的 JSON 构建
	tests := []struct {
		name     string                 // 测试用例名称
		command  map[string]interface{} // 命令参数
		expected string                 // 期望的 prefix 值
	}{
		{
			name:     "status 命令",
			command:  map[string]interface{}{"prefix": "status", "format": "json"},
			expected: "status",
		},
		{
			name:     "osd df 命令",
			command:  map[string]interface{}{"prefix": "osd df", "format": "json"},
			expected: "osd df",
		},
		{
			name:     "mon dump 命令",
			command:  map[string]interface{}{"prefix": "mon dump", "format": "json"},
			expected: "mon dump",
		},
		{
			name:     "pg stat 命令",
			command:  map[string]interface{}{"prefix": "pg stat", "format": "json"},
			expected: "pg stat",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 序列化命令
			data, err := json.Marshal(tt.command)
			if err != nil {
				t.Fatalf("序列化命令失败: %v", err)
			}

			// 反序列化验证
			var parsed map[string]interface{}
			if err := json.Unmarshal(data, &parsed); err != nil {
				t.Fatalf("反序列化命令失败: %v", err)
			}

			if parsed["prefix"] != tt.expected {
				t.Errorf("prefix 期望 '%s'，实际 '%v'", tt.expected, parsed["prefix"])
			}
			if parsed["format"] != "json" {
				t.Errorf("format 期望 'json'，实际 '%v'", parsed["format"])
			}
		})
	}
}

// TestClusterStatus_HealthWarn 测试 HEALTH_WARN 状态解析
func TestClusterStatus_HealthWarn(t *testing.T) {
	jsonData := `{
		"health": {
			"status": "HEALTH_WARN",
			"checks": {
				"TOO_FEW_OSDS": {
					"severity": "HEALTH_WARN",
					"summary": {
						"message": "OSD count 1 < osd_pool_default_size 3"
					}
				}
			}
		},
		"pgmap": {"num_pgs": 64, "num_pools": 1},
		"osdmap": {"num_osds": 1, "num_up_osds": 1, "num_in_osds": 1},
		"monmap": {"num_mons": 1}
	}`

	var status ClusterStatus
	if err := json.Unmarshal([]byte(jsonData), &status); err != nil {
		t.Fatalf("解析 HEALTH_WARN 状态失败: %v", err)
	}

	if status.Health.Status != "HEALTH_WARN" {
		t.Errorf("Health.Status 期望 'HEALTH_WARN'，实际 '%s'", status.Health.Status)
	}

	// 验证健康检查项
	check, ok := status.Health.Checks["TOO_FEW_OSDS"]
	if !ok {
		t.Fatal("缺少 TOO_FEW_OSDS 健康检查项")
	}
	if check.Severity != "HEALTH_WARN" {
		t.Errorf("检查项严重程度期望 'HEALTH_WARN'，实际 '%s'", check.Severity)
	}
}
