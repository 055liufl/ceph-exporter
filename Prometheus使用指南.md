# Prometheus 使用指南（中文）

## 概述

Prometheus 是一个开源的监控和告警系统。虽然其 Web UI 界面是英文的，但本指南将帮助你理解和使用各项功能。

**重要提示：** 建议主要使用 Grafana 进行日常监控，Prometheus UI 主要用于调试和配置验证。

## 访问地址

- **Prometheus UI：** http://localhost:9090
- **Grafana（中文界面）：** http://localhost:3000

## 主要功能页面

### 1. Graph（图表查询）

**访问路径：** 顶部菜单 → Graph

**用途：** 执行 PromQL 查询并查看结果

**界面说明：**
- **Expression（表达式）：** 输入 PromQL 查询语句
- **Execute（执行）：** 点击执行查询
- **Add Graph（添加图表）：** 添加新的查询面板
- **Table（表格）：** 以表格形式显示结果
- **Graph（图表）：** 以图表形式显示结果

**常用查询示例：**

```promql
# 查看 Ceph 集群健康状态
ceph_health_status

# 查看集群总容量
ceph_cluster_total_bytes

# 查看集群已用容量
ceph_cluster_used_bytes

# 查看集群使用率
ceph_cluster_used_bytes / ceph_cluster_total_bytes * 100

# 查看 OSD 状态
ceph_cluster_osds_up
ceph_cluster_osds_total

# 查看 OSD 延迟
ceph_osd_apply_latency_ms
ceph_osd_commit_latency_ms

# 查看存储池使用情况
ceph_pool_used_bytes
ceph_pool_objects_total
```

### 2. Alerts（告警规则）

**访问路径：** 顶部菜单 → Alerts

**用途：** 查看告警规则和当前触发的告警

**界面说明：**
- **Inactive（未激活）：** 告警条件未满足
- **Pending（待定）：** 告警条件满足，但还在等待时间内
- **Firing（触发中）：** 告警已触发

**告警状态颜色：**
- 🟢 绿色：正常，无告警
- 🟡 黄色：Pending 状态
- 🔴 红色：Firing 状态

**常见告警（中文说明）：**

| 告警名称 | 中文说明 | 严重级别 |
|---------|---------|---------|
| CephClusterWarning | Ceph 集群处于 HEALTH_WARN 状态 | warning |
| CephClusterError | Ceph 集群处于 HEALTH_ERR 状态 | critical |
| CephOSDDown | OSD 节点宕机 | warning |
| CephOSDOut | OSD 节点被踢出集群 | warning |
| CephOSDHighUtilization | OSD 使用率过高 (>85%) | warning |
| CephOSDHighLatency | OSD 延迟过高 (>500ms) | warning |
| CephMultipleOSDDown | 超过 10% 的 OSD 宕机 | critical |
| CephClusterCapacityWarning | 集群容量使用率超过 75% | warning |
| CephClusterCapacityCritical | 集群容量使用率超过 85% | critical |
| CephClusterCapacityEmergency | 集群容量即将耗尽 (>95%) | emergency |
| CephPGNotClean | 存在非 active+clean 状态的 PG | warning |
| CephMonitorOutOfQuorum | Monitor 不在 quorum 中 | critical |
| CephExporterDown | ceph-exporter 服务不可用 | critical |

### 3. Status（状态信息）

**访问路径：** 顶部菜单 → Status

**子菜单说明：**

#### 3.1 Runtime & Build Information（运行时和构建信息）
- **用途：** 查看 Prometheus 版本和配置信息
- **关键信息：**
  - Version（版本）
  - Storage Retention（数据保留时间）
  - TSDB Status（时序数据库状态）

#### 3.2 Command-Line Flags（命令行参数）
- **用途：** 查看 Prometheus 启动参数
- **关键参数：**
  - `--config.file`：配置文件路径
  - `--storage.tsdb.path`：数据存储路径
  - `--storage.tsdb.retention.time`：数据保留时间（默认 30 天）

#### 3.3 Configuration（配置）
- **用途：** 查看当前加载的配置文件内容
- **包含：**
  - 全局配置（global）
  - 告警规则文件（rule_files）
  - 采集目标配置（scrape_configs）

#### 3.4 Rules（规则）
- **用途：** 查看所有告警规则的详细信息
- **显示内容：**
  - 规则名称
  - 规则表达式
  - 持续时间（for）
  - 标签（labels）
  - 注释（annotations）

#### 3.5 Targets（采集目标）⭐ 最常用
- **用途：** 查看所有监控目标的状态
- **界面说明：**
  - **All scrape pools（所有采集池）：** 下拉选择特定的采集任务
  - **All（全部）：** 显示所有目标
  - **Unhealthy（不健康）：** 只显示采集失败的目标
  - **Collapse All（全部折叠）：** 折叠所有采集池

**目标状态说明：**

| 状态 | 图标 | 说明 |
|------|------|------|
| UP | 🟢 | 采集成功，目标正常 |
| DOWN | 🔴 | 采集失败，目标不可达 |
| UNKNOWN | 🟡 | 状态未知 |

**表格列说明：**
- **Endpoint（端点）：** 采集目标的 URL
- **State（状态）：** UP/DOWN
- **Labels（标签）：** 目标的标签信息
- **Last Scrape（最后采集）：** 最后一次采集的时间
- **Scrape Duration（采集耗时）：** 采集花费的时间
- **Error（错误）：** 如果采集失败，显示错误信息

**常见采集目标：**
- `ceph-exporter`：Ceph 集群指标采集器
- `prometheus`：Prometheus 自身指标
- `alertmanager`：Alertmanager 指标

#### 3.6 Service Discovery（服务发现）
- **用途：** 查看服务发现的目标
- **说明：** 本项目使用静态配置，此页面通常为空

### 4. Help（帮助）

**访问路径：** 顶部菜单 → Help

**内容：**
- Prometheus 官方文档链接
- PromQL 查询语言文档
- API 文档

## 常用操作指南

### 操作 1：检查采集目标状态

1. 访问 http://localhost:9090
2. 点击顶部菜单 **Status** → **Targets**
3. 查看所有目标的状态
4. 确认 `ceph-exporter` 显示为 **UP**（绿色）

**如果显示 DOWN（红色）：**
- 检查 ceph-exporter 容器是否运行：`./deploy.sh status`
- 查看 ceph-exporter 日志：`./deploy.sh logs ceph-exporter`
- 验证端点可访问：`curl http://localhost:9128/metrics`

### 操作 2：查看告警规则

1. 访问 http://localhost:9090
2. 点击顶部菜单 **Alerts**
3. 查看所有告警规则的状态
4. 点击告警名称查看详细信息

**告警状态说明：**
- **Inactive（绿色）：** 正常，无需处理
- **Pending（黄色）：** 告警条件满足，等待触发
- **Firing（红色）：** 告警已触发，需要处理

### 操作 3：执行 PromQL 查询

1. 访问 http://localhost:9090
2. 点击顶部菜单 **Graph**
3. 在 **Expression** 输入框中输入查询语句
4. 点击 **Execute** 执行查询
5. 切换 **Table** 或 **Graph** 查看结果

**示例查询：**

```promql
# 查看集群容量使用率（百分比）
(ceph_cluster_used_bytes / ceph_cluster_total_bytes) * 100

# 查看最近 5 分钟的 OSD 延迟
rate(ceph_osd_apply_latency_ms[5m])

# 查看存储池使用情况（按存储池分组）
sum by (pool) (ceph_pool_used_bytes)
```

### 操作 4：验证配置文件

1. 访问 http://localhost:9090
2. 点击顶部菜单 **Status** → **Configuration**
3. 查看当前加载的配置文件
4. 确认采集目标和告警规则配置正确

### 操作 5：查看 Prometheus 性能

1. 访问 http://localhost:9090
2. 点击顶部菜单 **Status** → **Runtime & Build Information**
3. 查看以下信息：
   - **Storage Retention（数据保留时间）：** 默认 30 天
   - **TSDB Status（数据库状态）：**
     - Head Chunks：内存中的数据块数量
     - WAL Size：预写日志大小
     - Series：时间序列数量

## 界面术语对照表

| 英文术语 | 中文翻译 | 说明 |
|---------|---------|------|
| Graph | 图表 | 查询和可视化页面 |
| Alerts | 告警 | 告警规则和状态 |
| Status | 状态 | 系统状态信息 |
| Targets | 采集目标 | 监控目标列表 |
| Rules | 规则 | 告警规则详情 |
| Configuration | 配置 | 配置文件内容 |
| Service Discovery | 服务发现 | 自动发现的目标 |
| Expression | 表达式 | PromQL 查询语句 |
| Execute | 执行 | 执行查询 |
| Table | 表格 | 表格视图 |
| Endpoint | 端点 | 采集目标 URL |
| State | 状态 | UP/DOWN 状态 |
| Labels | 标签 | 目标标签 |
| Last Scrape | 最后采集 | 最后采集时间 |
| Scrape Duration | 采集耗时 | 采集花费时间 |
| Error | 错误 | 错误信息 |
| Inactive | 未激活 | 告警未触发 |
| Pending | 待定 | 告警等待中 |
| Firing | 触发中 | 告警已触发 |
| Unhealthy | 不健康 | 采集失败的目标 |
| Collapse All | 全部折叠 | 折叠所有面板 |
| Show less | 显示更少 | 折叠详情 |
| Filter | 过滤 | 过滤条件 |

## 推荐使用方式

### 日常监控：使用 Grafana（中文界面）

**推荐理由：**
- ✅ 完整的中文界面
- ✅ 更好的可视化效果
- ✅ 自定义 Dashboard
- ✅ 更友好的用户体验

**访问地址：** http://localhost:3000

### 调试和验证：使用 Prometheus

**使用场景：**
- 验证采集目标状态
- 调试 PromQL 查询
- 查看告警规则详情
- 检查配置文件

**访问地址：** http://localhost:9090

## 常见问题

### Q1: 为什么 Prometheus 界面是英文的？

**A:** Prometheus 是一个 Go 语言编写的开源项目，官方不提供多语言支持。界面文本是硬编码在源代码中的，无法通过配置更改。

### Q2: 可以汉化 Prometheus 吗？

**A:** 理论上可以，但需要：
1. 修改 Prometheus 源代码
2. 重新编译
3. 维护自定义版本

这样做成本很高，不推荐。建议主要使用 Grafana（已完全中文化）。

### Q3: 如何快速找到需要的功能？

**A:** 参考本指南的"主要功能页面"部分，了解各个菜单的用途。最常用的是：
- **Targets**：查看采集目标状态
- **Alerts**：查看告警
- **Graph**：执行查询

### Q4: 告警规则的中文说明在哪里？

**A:** 虽然 Prometheus UI 是英文的，但告警规则的描述（annotations）已经配置为中文。你可以：
1. 在 Prometheus Alerts 页面点击告警名称查看详情
2. 在 Grafana 中查看告警（中文界面）
3. 查看配置文件：`ceph-exporter/deployments/prometheus/alert_rules.zh-CN.yml`

### Q5: 如何学习 PromQL？

**A:** 推荐资源：
1. Prometheus 官方文档：https://prometheus.io/docs/prometheus/latest/querying/basics/
2. 本项目的查询示例（见上文"常用查询示例"）
3. 在 Grafana 中查看现有 Dashboard 的查询语句

## 快速参考

### 常用 URL

| 功能 | URL |
|------|-----|
| Grafana（中文） | http://localhost:3000 |
| Prometheus 主页 | http://localhost:9090 |
| 采集目标状态 | http://localhost:9090/targets |
| 告警规则 | http://localhost:9090/alerts |
| 配置文件 | http://localhost:9090/config |
| Prometheus API | http://localhost:9090/api/v1/ |
| Ceph Exporter 指标 | http://localhost:9128/metrics |

### 常用命令

```bash
# 查看服务状态
./deploy.sh status

# 查看 Prometheus 日志
./deploy.sh logs prometheus

# 验证 Prometheus 配置
docker exec prometheus promtool check config /etc/prometheus/prometheus.yml

# 验证告警规则
docker exec prometheus promtool check rules /etc/prometheus/alert_rules.yml

# 重新加载配置（无需重启）
curl -X POST http://localhost:9090/-/reload

# 查看 Prometheus 指标
curl http://localhost:9090/metrics
```

## 总结

虽然 Prometheus Web UI 是英文的，但通过本指南，你可以：
- ✅ 理解各个功能页面的用途
- ✅ 掌握常用操作方法
- ✅ 查看中文的告警规则说明
- ✅ 主要使用 Grafana 进行日常监控

**建议：** 将 Grafana 作为主要监控界面（完全中文），Prometheus 仅用于调试和配置验证。
