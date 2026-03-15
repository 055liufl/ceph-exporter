# ceph-exporter 告警问题完整解决方案

## 问题总结

**现象**: Alertmanager 显示 3 个告警（CephMonitorOutOfQuorum, CephOSDDown, CephOSDOut），但 Ceph 集群实际正常运行。

**根本原因**: ceph-exporter 代码中的 OSDStats 结构体字段定义与 `ceph osd df -f json` 实际返回的 JSON 结构不匹配。

### 字段不匹配详情

**代码期望的字段**:
```go
Up   int `json:"up"`   // 期望: 1 或 0
In   int `json:"in"`   // 期望: 1 或 0
```

**实际返回的字段**:
```json
{
  "status": "up",      // 实际: 字符串 "up" 或 "down"
  "reweight": 1.0      // 实际: 浮点数 1.0 (in) 或 0.0 (out)
}
```

由于字段不匹配，JSON 解析时 `Up` 和 `In` 字段使用了默认值 0，导致指标错误。

---

## 解决方案

### 方案 1: 修复 ceph-exporter 代码（推荐）

#### 步骤 1: 备份并修改 client.go

```bash
cd /home/lfl/ceph-exporter/ceph-exporter
cp internal/ceph/client.go internal/ceph/client.go.backup
```

编辑 `internal/ceph/client.go`，找到第 116-129 行的 `OSDStats` 结构体，替换为：

```go
// OSDStats OSD 统计信息
// 对应 "ceph osd df -f json" 命令的输出
type OSDStats struct {
	ID              int     `json:"id"`                  // OSD ID
	Name            string  `json:"name"`                // OSD 名称（如 "osd.0"）
	Status          string  `json:"status"`              // 状态描述 ("up" 或 "down")
	Reweight        float64 `json:"reweight"`            // 权重 (1.0=in, 0.0=out)
	TotalBytes      int64   `json:"kb"`                  // 总容量（KB）
	UsedBytes       int64   `json:"kb_used"`             // 已使用容量（KB）
	AvailBytes      int64   `json:"kb_avail"`            // 可用容量（KB）
	Utilization     float64 `json:"utilization"`         // 使用率（0-100）
	PGs             int     `json:"pgs"`                 // 该 OSD 上的 PG 数量
	ApplyLatencyMs  float64 `json:"apply_latency_ms"`    // 应用延迟（毫秒）
	CommitLatencyMs float64 `json:"commit_latency_ms"`   // 提交延迟（毫秒）
}

// Up 返回 OSD 是否处于 up 状态（1=up, 0=down）
func (o *OSDStats) Up() int {
	if o.Status == "up" {
		return 1
	}
	return 0
}

// In 返回 OSD 是否处于 in 状态（1=in, 0=out）
func (o *OSDStats) In() int {
	if o.Reweight > 0 {
		return 1
	}
	return 0
}
```

#### 步骤 2: 修改 osd.go collector

```bash
cd /home/lfl/ceph-exporter/ceph-exporter
cp internal/collector/osd.go internal/collector/osd.go.backup
```

编辑 `internal/collector/osd.go`，找到第 182-185 行，修改为：

```go
ch <- prometheus.MustNewConstMetric(c.up, prometheus.GaugeValue,
	float64(osd.Up()), osdName)  // 改为调用方法
ch <- prometheus.MustNewConstMetric(c.in, prometheus.GaugeValue,
	float64(osd.In()), osdName)  // 改为调用方法
```

#### 步骤 3: 重新编译和部署

```bash
cd /home/lfl/ceph-exporter/ceph-exporter
make build

cd deployments
sudo docker-compose -f docker-compose-lightweight-full.yml restart ceph-exporter

# 等待 30 秒后检查
sleep 30
curl -s --noproxy "*" http://localhost:9128/metrics | grep -E "ceph_osd_up|ceph_osd_in"
```

预期输出应该是：
```
ceph_osd_up{osd="osd.0"} 1
ceph_osd_in{osd="osd.0"} 1
```

---

### 方案 2: 临时解决方案 - 创建静默规则

如果暂时无法修改代码，可以在 Alertmanager 中创建静默规则：

1. 访问 http://192.168.75.129:9093
2. 点击右上角 **New Silence** 按钮
3. 添加匹配器：
   - `alertname =~ "CephOSDDown|CephOSDOut|CephMonitorOutOfQuorum"`
4. 设置时间范围（如 24 小时）
5. 填写原因："ceph-exporter bug, 实际集群正常"
6. 点击 **Create**

---

## Monitor 问题

还需要检查 Monitor 的问题。请执行：

```bash
sudo docker exec ceph-demo ceph mon dump -f json
sudo docker exec ceph-demo ceph quorum_status -f json
```

如果 `ceph mon dump` 输出无法解析，可能需要使用 `ceph quorum_status` 来获取 Monitor quorum 信息。

---

## 验证修复

修复后，等待 5-10 分钟，告警应该会自动清除。可以通过以下方式验证：

1. **检查指标**:
```bash
curl -s --noproxy "*" http://localhost:9128/metrics | grep -E "ceph_monitor_in_quorum|ceph_osd_up|ceph_osd_in"
```

应该看到：
```
ceph_monitor_in_quorum{monitor="ceph-demo"} 1
ceph_osd_up{osd="osd.0"} 1
ceph_osd_in{osd="osd.0"} 1
```

2. **检查 Prometheus 告警**:
访问 http://192.168.75.129:9090/alerts，告警应该变为绿色（Inactive）

3. **检查 Alertmanager**:
访问 http://192.168.75.129:9093，告警列表应该为空

---

## 文件位置

- 源代码: `/home/lfl/ceph-exporter/ceph-exporter/internal/ceph/client.go`
- 采集器: `/home/lfl/ceph-exporter/ceph-exporter/internal/collector/osd.go`
- 备份文件: `*.backup`
- 修复补丁: `/tmp/fix_osd_stats.patch`
