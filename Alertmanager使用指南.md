# Alertmanager 使用指南

> **版本**: 1.0
> **最后更新**: 2026-03-15
> **适用项目**: Ceph-Exporter 监控系统

---

## 📖 目录

1. [什么是 Alertmanager](#什么是-alertmanager)
2. [快速开始](#快速开始)
3. [界面功能详解](#界面功能详解)
4. [告警规则说明](#告警规则说明)
5. [配置告警通知](#配置告警通知)
6. [告警管理操作](#告警管理操作)
7. [常见问题排查](#常见问题排查)
8. [最佳实践](#最佳实践)

---

## 什么是 Alertmanager

### 核心概念

Alertmanager 是 Prometheus 生态系统中的告警管理组件，负责处理 Prometheus 发送的告警，并通过各种渠道（邮件、Webhook、企业微信等）通知运维人员。

### 工作流程

```
┌─────────────┐      ┌──────────────┐      ┌─────────────┐
│  Prometheus │ ───> │ Alertmanager │ ───> │ 通知渠道     │
│  (告警规则)  │      │  (告警管理)   │      │ (邮件/Webhook)│
└─────────────┘      └──────────────┘      └─────────────┘
       │                     │
       │                     ├─ 告警分组
       │                     ├─ 告警抑制
       │                     ├─ 告警静默
       │                     └─ 告警路由
       │
   评估指标数据
```

### 核心功能

- **告警分组**: 将相似告警聚合，避免告警风暴
- **告警抑制**: 高级别告警自动抑制低级别告警
- **告警静默**: 临时屏蔽特定告警（维护期间）
- **告警路由**: 根据标签将告警路由到不同通知渠道
- **告警去重**: 避免重复发送相同告警

---

## 快速开始

### 访问 Alertmanager


**访问地址**: http://192.168.75.129:9093 (根据你的截图)

**默认端口**: 9093

**容器名称**: alertmanager

### 启动服务

```bash
# 启动完整监控栈（包含 Alertmanager）
cd /home/lfl/ceph-exporter/ceph-exporter/deployments
sudo docker-compose -f docker-compose-lightweight-full.yml up -d

# 查看 Alertmanager 状态
sudo docker-compose -f docker-compose-lightweight-full.yml ps alertmanager

# 查看 Alertmanager 日志
sudo docker-compose -f docker-compose-lightweight-full.yml logs -f alertmanager
```

### 验证服务

```bash
# 检查 Alertmanager 是否正常运行
curl http://localhost:9093/-/healthy

# 查看当前告警
curl http://localhost:9093/api/v2/alerts

# 查看 Alertmanager 配置
curl http://localhost:9093/api/v2/status
```

---

## 界面功能详解

### 主界面布局

根据你的截图，Alertmanager 界面包含以下部分：

```
┌────────────────────────────────────────────────────────┐
│  Alertmanager  [Alerts] [Silences] [Status] [Settings] │
├────────────────────────────────────────────────────────┤
│  [Filter] [Group]  Receiver: All  □Silenced  □Inhibited│
├────────────────────────────────────────────────────────┤
│  Custom matcher: env="production"                       │
│  [+] [Silence]                                          │
├────────────────────────────────────────────────────────┤
│  ⊕ Expand all groups                                    │
│                                                         │
│  + alertname="CephMonitorOutOfQuorum" component="monitor" │
│  + alertname="CephOSDDown" osd="osd.0"                  │
│  + alertname="CephOSDOut" osd="osd.0"                   │
└────────────────────────────────────────────────────────┘
```

### 1. Alerts 页面（告警列表）

**功能**: 显示当前所有活跃的告警

**界面元素**:
- **Filter**: 过滤告警（支持标签匹配）
- **Group**: 按标签分组显示
- **Receiver**: 选择接收者过滤
- **Silenced**: 显示已静默的告警
- **Inhibited**: 显示被抑制的告警

**告警卡片信息**:
- 告警名称（如 `CephMonitorOutOfQuorum`）
- 标签（如 `component="monitor"`）
- 告警数量（如 `1 alert`）
- 告警状态（Firing/Pending）

### 2. Silences 页面（静默管理）

**功能**: 创建和管理告警静默规则

**使用场景**:
- 系统维护期间临时屏蔽告警
- 已知问题暂时不需要通知
- 测试环境屏蔽特定告警

**操作步骤**:
1. 点击右上角 **New Silence** 按钮
2. 填写静默规则：
   - **Matchers**: 匹配条件（如 `alertname="CephOSDDown"`）
   - **Start**: 开始时间
   - **End**: 结束时间
   - **Creator**: 创建者
   - **Comment**: 静默原因
3. 点击 **Create** 创建静默

### 3. Status 页面（状态信息）

**功能**: 显示 Alertmanager 运行状态

**信息内容**:
- 版本信息
- 配置文件状态
- 集群状态（如果启用 HA）
- 上游 Prometheus 连接状态

### 4. Settings 页面（设置）

**功能**: 界面显示设置

**可配置项**:
- 时区设置
- 刷新间隔
- 告警分组方式

---

## 告警规则说明

### 当前配置的告警规则

本项目配置了 **15 个告警规则**，覆盖 Ceph 集群的各个方面：

#### 1. 集群健康状态告警

| 告警名称 | 触发条件 | 持续时间 | 级别 | 说明 |
|---------|---------|---------|------|------|
| CephClusterWarning | `ceph_health_status == 1` | 5分钟 | warning | 集群处于 HEALTH_WARN 状态 |
| CephClusterError | `ceph_health_status == 2` | 1分钟 | critical | 集群处于 HEALTH_ERR 状态 |

**示例**: 你的截图中没有显示集群健康告警，说明集群状态正常。

#### 2. OSD 告警

| 告警名称 | 触发条件 | 持续时间 | 级别 | 说明 |
|---------|---------|---------|------|------|
| CephOSDDown | `ceph_osd_up == 0` | 3分钟 | warning | OSD 宕机 |
| CephOSDOut | `ceph_osd_in == 0` | 5分钟 | warning | OSD 被踢出集群 |
| CephOSDHighUtilization | `ceph_osd_utilization > 85` | 10分钟 | warning | OSD 使用率超过 85% |
| CephOSDHighLatency | `ceph_osd_apply_latency_ms > 500` | 5分钟 | warning | OSD 写入延迟超过 500ms |
| CephMultipleOSDDown | `(down/total) > 0.1` | 2分钟 | critical | 超过 10% OSD 宕机 |

**示例**: 你的截图显示：
- `alertname="CephOSDDown" osd="osd.0"` - OSD 0 已宕机
- `alertname="CephOSDOut" osd="osd.0"` - OSD 0 被踢出集群

#### 3. 容量告警

| 告警名称 | 触发条件 | 持续时间 | 级别 | 说明 |
|---------|---------|---------|------|------|
| CephClusterCapacityWarning | 使用率 > 75% | 10分钟 | warning | 容量告警 |
| CephClusterCapacityCritical | 使用率 > 85% | 5分钟 | critical | 容量严重 |
| CephClusterCapacityEmergency | 使用率 > 95% | 1分钟 | emergency | 容量紧急 |

#### 4. PG 告警

| 告警名称 | 触发条件 | 持续时间 | 级别 | 说明 |
|---------|---------|---------|------|------|
| CephPGNotClean | 存在非 active+clean 的 PG | 15分钟 | warning | PG 状态异常 |
| CephNoPGs | `ceph_cluster_pgs_total == 0` | 5分钟 | critical | 没有 PG |

#### 5. Monitor 告警

| 告警名称 | 触发条件 | 持续时间 | 级别 | 说明 |
|---------|---------|---------|------|------|
| CephMonitorOutOfQuorum | `ceph_monitor_in_quorum == 0` | 2分钟 | critical | Monitor 脱离 quorum |
| CephMonitorClockSkew | 时钟偏差 > 0.5秒 | 5分钟 | warning | Monitor 时钟不同步 |

**示例**: 你的截图显示：
- `alertname="CephMonitorOutOfQuorum" component="monitor"` - 有 Monitor 脱离了 quorum

#### 6. 服务可用性告警

| 告警名称 | 触发条件 | 持续时间 | 级别 | 说明 |
|---------|---------|---------|------|------|
| CephExporterDown | `up{job="ceph-exporter"} == 0` | 2分钟 | critical | Exporter 不可用 |
| CephExporterSlowScrape | 采集耗时 > 10秒 | 5分钟 | warning | 采集延迟过高 |

---

## 配置告警通知

### 配置文件位置

```bash
# Alertmanager 配置文件
/home/lfl/ceph-exporter/ceph-exporter/deployments/alertmanager/alertmanager.zh-CN.yml

# 告警规则文件
/home/lfl/ceph-exporter/ceph-exporter/deployments/prometheus/alert_rules.zh-CN.yml
```

### 1. 配置 Webhook 通知（推荐）

Webhook 是最灵活的通知方式，可以对接企业微信、钉钉、Slack 等。

#### 企业微信机器人配置

```yaml
receivers:
  - name: "wechat-webhook"
    webhook_configs:
      - url: "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=YOUR_KEY"
        send_resolved: true
```

**Webhook 消息格式**:
```json
{
  "msgtype": "markdown",
  "markdown": {
    "content": "## Ceph 告警\n\n**告警名称**: CephOSDDown\n**OSD**: osd.0\n**级别**: warning\n**描述**: OSD 0 已宕机超过 3 分钟"
  }
}
```

#### 钉钉机器人配置

```yaml
receivers:
  - name: "dingtalk-webhook"
    webhook_configs:
      - url: "https://oapi.dingtalk.com/robot/send?access_token=YOUR_TOKEN"
        send_resolved: true
```

### 2. 配置邮件通知

编辑 `alertmanager.zh-CN.yml`:

```yaml
global:
  # SMTP 配置
  smtp_smarthost: "smtp.example.com:587"
  smtp_from: "alertmanager@example.com"
  smtp_auth_username: "alertmanager@example.com"
  smtp_auth_password: "your_password"
  smtp_require_tls: true

receivers:
  - name: "email-alerts"
    email_configs:
      - to: "ops-team@example.com"
        send_resolved: true
        headers:
          Subject: "[Ceph Alert] {{ .GroupLabels.alertname }}"
```

### 3. 配置告警路由

根据告警级别和组件路由到不同接收者：

```yaml
route:
  receiver: "default-webhook"
  group_by: ["alertname", "component"]
  group_wait: 30s
  group_interval: 5m
  repeat_interval: 4h

  routes:
    # 紧急告警 -> 电话通知
    - match_re:
        severity: "critical|emergency"
      receiver: "critical-webhook"
      group_wait: 10s
      repeat_interval: 1h

    # OSD 告警 -> OSD 运维组
    - match:
        component: "osd"
      receiver: "osd-team-webhook"

    # 容量告警 -> 存储运维组
    - match:
        component: "capacity"
      receiver: "storage-team-email"
      repeat_interval: 2h
```

### 4. 配置告警抑制

避免告警风暴，高级别告警自动抑制低级别告警：

```yaml
inhibit_rules:
  # critical 抑制 warning
  - source_match:
      severity: "critical"
    target_match:
      severity: "warning"
    equal: ["component"]

  # emergency 抑制 critical 和 warning
  - source_match:
      severity: "emergency"
    target_match_re:
      severity: "critical|warning"
    equal: ["component"]
```

### 5. 重新加载配置

```bash
# 方法 1: 热重载（推荐）
curl -X POST http://localhost:9093/-/reload

# 方法 2: 重启容器
cd /home/lfl/ceph-exporter/ceph-exporter/deployments
sudo docker-compose -f docker-compose-lightweight-full.yml restart alertmanager

# 验证配置
curl http://localhost:9093/api/v2/status | jq .
```

---

## 告警管理操作

### 1. 查看当前告警

**Web 界面**:
1. 访问 http://192.168.75.129:9093
2. 点击 **Alerts** 标签
3. 查看告警列表

**命令行**:
```bash
# 查看所有告警
curl -s http://localhost:9093/api/v2/alerts | jq .

# 查看特定告警
curl -s http://localhost:9093/api/v2/alerts | jq '.[] | select(.labels.alertname=="CephOSDDown")'

# 统计告警数量
curl -s http://localhost:9093/api/v2/alerts | jq 'length'
```

### 2. 过滤告警

**使用 Filter 功能**:

在界面顶部的搜索框输入过滤条件：

```
# 按告警名称过滤
alertname="CephOSDDown"

# 按组件过滤
component="osd"

# 按级别过滤
severity="critical"

# 组合条件
alertname="CephOSDDown",osd="osd.0"

# 正则匹配
alertname=~"CephOSD.*"
```

### 3. 创建静默规则

**场景**: OSD 0 正在维护，临时屏蔽相关告警

**操作步骤**:

1. 点击右上角 **New Silence** 按钮
2. 填写静默规则：
   ```
   Matchers:
     osd = "osd.0"

   Start: 2026-03-15 16:00:00
   End:   2026-03-15 18:00:00

   Creator: admin
   Comment: OSD 0 维护，预计 2 小时完成
   ```
3. 点击 **Create** 创建

**命令行创建静默**:
```bash
# 创建静默规则
curl -X POST http://localhost:9093/api/v2/silences \
  -H "Content-Type: application/json" \
  -d '{
    "matchers": [
      {
        "name": "osd",
        "value": "osd.0",
        "isRegex": false
      }
    ],
    "startsAt": "2026-03-15T16:00:00Z",
    "endsAt": "2026-03-15T18:00:00Z",
    "createdBy": "admin",
    "comment": "OSD 0 维护"
  }'
```

### 4. 管理静默规则

**查看静默列表**:
```bash
# Web 界面: 点击 Silences 标签
# 命令行:
curl -s http://localhost:9093/api/v2/silences | jq .
```

**删除静默规则**:
```bash
# 获取静默 ID
SILENCE_ID=$(curl -s http://localhost:9093/api/v2/silences | jq -r '.[0].id')

# 删除静默
curl -X DELETE http://localhost:9093/api/v2/silence/$SILENCE_ID
```

### 5. 告警分组查看

**按组件分组**:
1. 点击 **Group** 按钮
2. 选择分组字段（如 `component`）
3. 告警会按组件分组显示

**展开所有分组**:
- 点击 **Expand all groups** 链接

---

## 常见问题排查

### 问题 1: Alertmanager 没有收到告警

**排查步骤**:

```bash
# 1. 检查 Prometheus 是否配置了 Alertmanager
curl -s http://localhost:9090/api/v1/alertmanagers | jq .

# 2. 检查 Prometheus 告警规则是否加载
curl -s http://localhost:9090/api/v1/rules | jq .

# 3. 检查 Prometheus 是否有活跃告警
curl -s http://localhost:9090/api/v1/alerts | jq .

# 4. 检查 Alertmanager 日志
sudo docker-compose -f docker-compose-lightweight-full.yml logs alertmanager
```

**解决方案**:

检查 Prometheus 配置文件 `prometheus.zh-CN.yml`:
```yaml
alerting:
  alertmanagers:
    - static_configs:
        - targets:
            - alertmanager:9093  # 确保配置正确
```

### 问题 2: 告警通知没有发送

**排查步骤**:

```bash
# 1. 检查 Alertmanager 配置是否正确
curl -s http://localhost:9093/api/v2/status | jq .config

# 2. 检查接收者配置
curl -s http://localhost:9093/api/v2/receivers

# 3. 测试 Webhook 是否可达
curl -X POST http://localhost:5001/webhook/alerts \
  -H "Content-Type: application/json" \
  -d '{"test": "message"}'

# 4. 查看 Alertmanager 日志
sudo docker logs alertmanager 2>&1 | grep -i error
```

**常见原因**:
- Webhook URL 配置错误
- SMTP 配置错误（邮件通知）
- 网络不通
- 接收者名称拼写错误

### 问题 3: 告警一直处于 Pending 状态

**原因**: 告警规则的 `for` 持续时间未满足

**示例**:
```yaml
- alert: CephOSDDown
  expr: ceph_osd_up == 0
  for: 3m  # 需要持续 3 分钟才触发
```

**解决方案**:
- 等待持续时间满足
- 或修改 `for` 参数缩短等待时间

### 问题 4: 告警被抑制（Inhibited）

**原因**: 存在更高级别的同类告警

**示例**:
- `CephClusterError` (critical) 存在时
- `CephClusterWarning` (warning) 会被抑制

**查看抑制规则**:
```bash
curl -s http://localhost:9093/api/v2/status | jq .config.inhibit_rules
```

**解决方案**:
- 这是正常行为，先处理高级别告警
- 如需查看被抑制的告警，勾选界面上的 **Inhibited** 复选框

### 问题 5: 告警风暴（大量重复告警）

**原因**: 告警分组配置不当

**解决方案**:

调整 `alertmanager.zh-CN.yml` 的分组配置：
```yaml
route:
  group_by: ["alertname", "component"]  # 按告警名称和组件分组
  group_wait: 30s       # 等待 30 秒收集同组告警
  group_interval: 5m    # 同组告警间隔 5 分钟
  repeat_interval: 4h   # 重复通知间隔 4 小时
```

---

## 最佳实践

### 1. 告警级别规范

| 级别 | 使用场景 | 响应时间 | 通知方式 |
|------|---------|---------|---------|
| **warning** | 需要关注但不紧急 | 工作时间内处理 | 邮件/企业微信 |
| **critical** | 严重问题，需尽快处理 | 30 分钟内响应 | 电话/短信 |
| **emergency** | 紧急问题，立即处理 | 5 分钟内响应 | 电话 |

### 2. 告警命名规范

```
<组件><问题类型>
```

**示例**:
- `CephOSDDown` - Ceph OSD 宕机
- `CephClusterCapacityWarning` - Ceph 集群容量告警
- `CephMonitorOutOfQuorum` - Ceph Monitor 脱离 quorum

### 3. 告警描述规范

告警描述应包含：
- **问题现象**: 发生了什么
- **影响范围**: 影响哪些服务
- **排查命令**: 如何排查
- **处理建议**: 如何处理

**示例**:
```yaml
annotations:
  summary: "OSD {{ $labels.osd }} 已宕机"
  description: |
    OSD {{ $labels.osd }} 处于 down 状态已超过 3 分钟。

    影响: 数据冗余度降低，可能影响集群性能

    排查命令:
      ceph osd tree
      ceph osd status
      systemctl status ceph-osd@{{ $labels.osd }}

    处理建议:
      1. 检查 OSD 所在节点是否正常
      2. 检查磁盘是否故障
      3. 尝试重启 OSD 服务
```

### 4. 静默规则使用规范

**何时使用静默**:
- ✅ 计划内维护
- ✅ 已知问题暂时无法修复
- ✅ 测试环境屏蔽告警

**何时不使用静默**:
- ❌ 长期屏蔽告警（应修改告警规则）
- ❌ 不清楚原因的告警（应先排查）

**静默规则命名**:
```
<原因>-<组件>-<日期>
```

**示例**:
- `maintenance-osd0-20260315` - OSD 0 维护
- `known-issue-monitor-20260315` - Monitor 已知问题

### 5. 告警通知分级

```yaml
route:
  routes:
    # 紧急告警 -> 电话 + 短信 + 企业微信
    - match_re:
        severity: "emergency"
      receiver: "emergency-all"
      group_wait: 10s
      repeat_interval: 30m

    # 严重告警 -> 电话 + 企业微信
    - match_re:
        severity: "critical"
      receiver: "critical-phone-wechat"
      group_wait: 30s
      repeat_interval: 1h

    # 警告告警 -> 企业微信 + 邮件
    - match_re:
        severity: "warning"
      receiver: "warning-wechat-email"
      group_wait: 5m
      repeat_interval: 4h
```

### 6. 告警测试

**定期测试告警流程**:

```bash
# 1. 触发测试告警
curl -X POST http://localhost:9093/api/v2/alerts \
  -H "Content-Type: application/json" \
  -d '[{
    "labels": {
      "alertname": "TestAlert",
      "severity": "warning",
      "component": "test"
    },
    "annotations": {
      "summary": "这是一个测试告警",
      "description": "用于测试告警通知是否正常"
    },
    "startsAt": "'$(date -u +%Y-%m-%dT%H:%M:%SZ)'"
  }]'

# 2. 验证是否收到通知

# 3. 清理测试告警（等待 resolve_timeout 自动清理）
```

### 7. 监控 Alertmanager 自身

在 Prometheus 中添加 Alertmanager 监控：

```yaml
# alert_rules.yml
- alert: AlertmanagerDown
  expr: up{job="alertmanager"} == 0
  for: 2m
  labels:
    severity: critical
  annotations:
    summary: "Alertmanager 服务不可用"
    description: "Alertmanager 已宕机超过 2 分钟，告警通知将无法发送"

- alert: AlertmanagerConfigReloadFailed
  expr: alertmanager_config_last_reload_successful == 0
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "Alertmanager 配置重载失败"
    description: "Alertmanager 配置文件可能存在语法错误"
```

---

## 附录

### A. 快速参考

**常用 API 端点**:
```bash
# 健康检查
GET  http://localhost:9093/-/healthy

# 查看告警
GET  http://localhost:9093/api/v2/alerts

# 查看静默
GET  http://localhost:9093/api/v2/silences

# 创建静默
POST http://localhost:9093/api/v2/silences

# 删除静默
DELETE http://localhost:9093/api/v2/silence/{id}

# 查看状态
GET  http://localhost:9093/api/v2/status

# 重载配置
POST http://localhost:9093/-/reload
```

**常用命令**:
```bash
# 启动服务
sudo docker-compose -f docker-compose-lightweight-full.yml up -d alertmanager

# 停止服务
sudo docker-compose -f docker-compose-lightweight-full.yml stop alertmanager

# 重启服务
sudo docker-compose -f docker-compose-lightweight-full.yml restart alertmanager

# 查看日志
sudo docker-compose -f docker-compose-lightweight-full.yml logs -f alertmanager

# 进入容器
sudo docker exec -it alertmanager sh

# 验证配置
docker exec alertmanager amtool check-config /etc/alertmanager/alertmanager.yml
```

### B. 相关文档

- **Prometheus 使用指南**: `/home/lfl/ceph-exporter/Prometheus使用指南.md`
- **Ceph-Exporter 完整操作指南**: `/home/lfl/ceph-exporter/Ceph-Exporter项目完整操作指南.md`
- **官方文档**: https://prometheus.io/docs/alerting/latest/alertmanager/

---

**文档维护**: 如有问题或建议，请联系运维团队。
