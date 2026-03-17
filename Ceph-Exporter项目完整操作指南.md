# Ceph-Exporter 项目完整操作指南

> **版本**: 4.0
> **最后更新**: 2026-03-15
> **适用环境**: CentOS 7 + Docker
> **文档状态**: ✅ 全面更新 - 包含详细操作示例、界面说明、术语对照表

---

## 📖 目录

1. [项目概述](#1-项目概述)
2. [系统架构](#2-系统架构)
3. [快速开始](#3-快速开始)
4. [Grafana 完整操作指南](#4-grafana-完整操作指南)
5. [Prometheus 完整操作指南](#5-prometheus-完整操作指南)
6. [Alertmanager 完整操作指南](#6-alertmanager-完整操作指南)
7. [ELK Stack 完整操作指南](#7-elk-stack-完整操作指南)
8. [Jaeger 分布式追踪指南](#8-jaeger-分布式追踪指南)
9. [Ceph Dashboard 操作指南](#9-ceph-dashboard-操作指南)
10. [界面术语对照表](#10-界面术语对照表)
11. [常用操作速查](#11-常用操作速查)
12. [常见问题与解决方案](#12-常见问题与解决方案)
13. [最佳实践](#13-最佳实践)
14. [附录](#14-附录)

---

## 1. 项目概述

### 1.1 什么是 Ceph-Exporter？

Ceph-Exporter 是一个基于 Go 语言开发的 **Ceph 集群 Prometheus 指标导出器**，提供完整的监控、日志和追踪解决方案。

### 1.2 核心特性

#### 功能特性
- ✅ **7 个 Prometheus 采集器**: Cluster、Pool、OSD、Monitor、Health、MDS、RGW
- ✅ **完全中文化**: Grafana 100% 中文界面，所有配置文件中文注释
- ✅ **完整可观测性**: 指标监控 + 日志分析 + 分布式追踪
- ✅ **一键部署**: Docker Compose 自动化部署，支持多种模式
- ✅ **生产级配置**: 15 个告警规则，自动化诊断和修复脚本

#### 技术特性
- 📊 **50+ 指标**: 覆盖集群、存储池、OSD、Monitor 等
- 🔧 **CGO 集成**: 使用 go-ceph 库直接与 Ceph RADOS 通信
- 🧪 **完整测试**: 81 个单元测试，100% 通过率，覆盖率 68.1%
- 🚀 **高性能**: 并发采集，带超时控制
- 🔒 **安全可靠**: 支持 TLS/HTTPS，优雅关闭

### 1.3 系统要求

| 项目 | 最低要求 | 推荐配置 |
|------|---------|---------|
| 操作系统 | CentOS 7.x | CentOS 7.9 |
| Docker | 19.03+ | 20.10+ |
| Docker Compose | 1.25+ | 1.29+ |
| 内存 | 4GB | 8GB |
| CPU | 2 核 | 4 核 |
| 磁盘 | 30GB | 50GB |
| Ceph 版本 | Luminous (12.x) | Octopus (15.x) / Pacific (16.x) |

### 1.4 服务访问地址

| 服务 | 访问地址 | 账号 | 中文化程度 | 用途 |
|------|---------|------|-----------|------|
| **Grafana** | http://localhost:3000 | admin/admin | ✅ 100% | 可视化仪表板 |
| **Prometheus** | http://localhost:9090 | - | ⚠️ 50% | 指标查询和告警 |
| **Alertmanager** | http://localhost:9093 | - | ⚠️ 40% | 告警管理 |
| **Kibana** | http://localhost:5601 | - | ✅ 90% | 日志查询和分析 |
| **Elasticsearch** | http://localhost:9200 | - | - | 日志存储 |
| **Jaeger UI** | http://localhost:16686 | - | ❌ 0% | 分布式追踪 |
| **Ceph Dashboard** | http://localhost:8080 | - | ⚠️ 60% | Ceph 集群管理 |
| **ceph-exporter** | http://localhost:9128/metrics | - | - | Prometheus 指标 |

---

## 2. 系统架构

### 2.1 整体架构图

```
┌─────────────────────────────────────────────────────────────────┐
│                          用户访问层                               │
│  Grafana (3000)  │  Prometheus (9090)  │  Kibana (5601)         │
│  Jaeger UI (16686)  │  Alertmanager (9093)                      │
└─────────────────────────────────────────────────────────────────┘
                              ▲
                              │ HTTP/API
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                         监控采集层                                │
│  ceph-exporter (9128)  │  Logstash (5044)  │  Jaeger (4318)     │
└─────────────────────────────────────────────────────────────────┘
                              ▲
                              │ librados
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                         存储层                                    │
│                    Ceph Cluster (RADOS)                          │
│  Monitor  │  OSD  │  MDS  │  RGW                                │
└─────────────────────────────────────────────────────────────────┘
```

### 2.2 数据流向

1. **指标采集流程**:
   ```
   Ceph Cluster → ceph-exporter → Prometheus → Grafana
   ```

2. **日志采集流程**:
   ```
   ceph-exporter → Logstash → Elasticsearch → Kibana
   ```

3. **追踪流程**:
   ```
   HTTP Request → ceph-exporter → Jaeger Collector → Jaeger UI
   ```

4. **告警流程**:
   ```
   Prometheus (评估规则) → Alertmanager (分组/路由) → 通知渠道
   ```

---

## 3. 快速开始

### 3.1 一键部署

```bash
# 进入部署目录
cd /home/lfl/ceph-exporter/ceph-exporter/deployments

# 完整监控栈部署（推荐，部署时交互式选择日志方案）
./scripts/deploy.sh full

# 或指定日志方案（跳过交互）
LOGGING_MODE=container ./scripts/deploy.sh full   # 容器日志收集（推荐）
LOGGING_MODE=direct ./scripts/deploy.sh full      # 直接推送到 Logstash

# 等待服务启动（约 2-3 分钟）
./scripts/deploy.sh status

# 验证部署
./scripts/deploy.sh verify
```

### 3.2 访问服务

部署完成后，打开浏览器访问：

1. **Grafana** (推荐首先访问): http://localhost:3000
   - 账号: `admin`
   - 密码: `admin`
   - 首次登录会提示修改密码（可跳过）

2. **Prometheus**: http://localhost:9090
   - 查看指标和告警规则

3. **Alertmanager**: http://localhost:9093
   - 查看和管理告警

4. **Kibana**: http://localhost:5601
   - 查看日志（需要先启用 ELK）

5. **Jaeger**: http://localhost:16686
   - 查看分布式追踪（需要先启用 Jaeger）

### 3.3 验证指标采集

```bash
# 查看 ceph-exporter 指标
curl http://localhost:9128/metrics | grep ceph_

# 查看健康检查
curl http://localhost:9128/health

# 查看 Prometheus targets
curl http://localhost:9090/api/v1/targets
```

---

## 4. Grafana 完整操作指南

### 4.1 Grafana 简介

Grafana 是一个开源的可视化和分析平台，用于展示 Prometheus 采集的 Ceph 集群指标。本项目提供 **100% 中文化** 的 Grafana 界面。

### 4.2 首次登录

1. 打开浏览器访问: http://localhost:3000
2. 输入默认账号:
   - 用户名: `admin`
   - 密码: `admin`
3. 首次登录会提示修改密码，可以选择"跳过"

### 4.3 界面布局说明

#### 主界面结构

```
┌─────────────────────────────────────────────────────────────┐
│  [Grafana Logo]  [搜索]  [创建]  [仪表盘]  [探索]  [告警]  [配置]  [admin▼] │
├─────────────────────────────────────────────────────────────┤
│  侧边栏                    │  主内容区域                      │
│  ├─ 主页                   │                                 │
│  ├─ 仪表盘                 │  [仪表盘内容]                   │
│  ├─ 探索                   │  - 图表面板                     │
│  ├─ 告警                   │  - 数据表格                     │
│  ├─ 配置                   │  - 统计信息                     │
│  └─ 帮助                   │                                 │
└─────────────────────────────────────────────────────────────┘
```

#### 顶部导航栏

| 图标/按钮 | 中文名称 | 功能说明 |
|----------|---------|---------|
| 🏠 | 主页 | 返回 Grafana 主页 |
| 🔍 | 搜索 | 搜索仪表盘、文件夹、标签 |
| ➕ | 创建 | 创建仪表盘、文件夹、导入 |
| 📊 | 仪表盘 | 浏览所有仪表盘 |
| 🔬 | 探索 | 临时查询和探索数据 |
| 🔔 | 告警 | 查看和管理告警规则 |
| ⚙️ | 配置 | 数据源、插件、用户管理 |
| 👤 | 用户菜单 | 个人设置、退出登录 |

### 4.4 Ceph 集群仪表盘详解

#### 4.4.1 访问 Ceph 仪表盘

1. 点击左侧边栏 **"仪表盘"** 图标
2. 在文件夹列表中找到 **"Ceph"** 文件夹
3. 点击进入，选择 **"Ceph 集群监控"** 仪表盘

#### 4.4.2 仪表盘面板说明

**第一行：集群概览**

| 面板名称 | 显示内容 | 说明 |
|---------|---------|------|
| 集群健康状态 | HEALTH_OK / HEALTH_WARN / HEALTH_ERR | 绿色=正常，黄色=警告，红色=错误 |
| 集群总容量 | XX TB | 集群总存储容量 |
| 已用容量 | XX TB (XX%) | 已使用的存储空间和百分比 |
| 可用容量 | XX TB | 剩余可用存储空间 |

**第二行：性能指标**

| 面板名称 | 显示内容 | 说明 |
|---------|---------|------|
| 读取吞吐量 | XX MB/s | 集群当前读取速度 |
| 写入吞吐量 | XX MB/s | 集群当前写入速度 |
| 读取 IOPS | XX ops/s | 每秒读取操作数 |
| 写入 IOPS | XX ops/s | 每秒写入操作数 |

**第三行：组件状态**

| 面板名称 | 显示内容 | 说明 |
|---------|---------|------|
| OSD 状态 | Up: XX / In: XX / Total: XX | OSD 运行状态统计 |
| Monitor 状态 | In Quorum: XX / Total: XX | Monitor 仲裁状态 |
| PG 状态 | Active+Clean: XX / Total: XX | PG 健康状态 |
| 存储池数量 | XX 个 | 集群中的存储池总数 |

**第四行：详细图表**

1. **容量使用趋势图**
   - 显示过去 24 小时的容量变化
   - 可以看到数据增长趋势

2. **IOPS 趋势图**
   - 显示读写 IOPS 的时间序列
   - 可以识别性能峰值和低谷

3. **吞吐量趋势图**
   - 显示读写带宽的时间序列
   - 可以分析 I/O 模式

4. **OSD 利用率分布**
   - 显示各个 OSD 的使用率
   - 可以识别数据倾斜问题

### 4.5 常用操作

#### 4.5.1 调整时间范围

1. 点击右上角的时间选择器（默认显示"Last 6 hours"）
2. 选择预设时间范围：
   - 最近 5 分钟
   - 最近 15 分钟
   - 最近 30 分钟
   - 最近 1 小时
   - 最近 6 小时
   - 最近 12 小时
   - 最近 24 小时
   - 最近 7 天
   - 最近 30 天
3. 或自定义时间范围：
   - 点击"自定义时间范围"
   - 选择开始和结束时间
   - 点击"应用"

#### 4.5.2 刷新仪表盘

1. 点击右上角的刷新按钮 🔄
2. 或设置自动刷新间隔：
   - 点击刷新按钮旁边的下拉箭头
   - 选择刷新间隔（5s, 10s, 30s, 1m, 5m, 15m, 30m, 1h）

#### 4.5.3 查看面板详情

1. 点击面板标题
2. 选择"查看" (View)
3. 可以看到：
   - 完整的图表
   - 查询语句
   - 数据表格
   - 面板 JSON

#### 4.5.4 编辑面板

1. 点击面板标题
2. 选择"编辑" (Edit)
3. 可以修改：
   - 查询语句
   - 可视化类型
   - 面板选项
   - 阈值和告警

#### 4.5.5 导出仪表盘

1. 点击右上角的"分享仪表盘"图标
2. 选择"导出"标签
3. 点击"保存到文件"
4. 仪表盘将以 JSON 格式下载

#### 4.5.6 创建快照

1. 点击右上角的"分享仪表盘"图标
2. 选择"快照"标签
3. 设置快照名称和过期时间
4. 点击"本地快照"
5. 复制快照链接分享给他人

### 4.6 探索功能（Explore）

#### 4.6.1 访问探索页面

1. 点击左侧边栏的"探索"图标 🔬
2. 或在任意面板点击标题 → "探索"

#### 4.6.2 使用探索功能

1. **选择数据源**: 默认已选择 Prometheus
2. **输入查询**: 在查询框中输入 PromQL
3. **执行查询**: 点击"运行查询"按钮
4. **查看结果**:
   - 图表视图：时间序列图
   - 表格视图：原始数据
   - 日志视图：日志流（如果有）

#### 4.6.3 常用 PromQL 查询示例

```promql
# 集群总容量
ceph_cluster_total_bytes

# 集群使用率
(ceph_cluster_used_bytes / ceph_cluster_total_bytes) * 100

# 各存储池使用率
ceph_pool_percent_used

# OSD 状态
ceph_osd_up

# 集群 IOPS
rate(ceph_cluster_read_ops_sec[5m]) + rate(ceph_cluster_write_ops_sec[5m])

# 集群吞吐量
rate(ceph_cluster_read_bytes_sec[5m]) + rate(ceph_cluster_write_bytes_sec[5m])
```

### 4.7 告警配置

#### 4.7.1 查看告警规则

1. 点击左侧边栏的"告警"图标 🔔
2. 选择"告警规则"标签
3. 可以看到所有配置的告警规则

#### 4.7.2 创建告警规则

1. 在仪表盘中选择要添加告警的面板
2. 点击面板标题 → "编辑"
3. 切换到"告警"标签
4. 点击"创建告警"
5. 配置告警条件：
   - 评估间隔
   - 条件表达式
   - 阈值
6. 配置通知渠道
7. 保存

#### 4.7.3 告警通知渠道

支持的通知方式：
- Email (邮件)
- Webhook (自定义 HTTP 回调)
- Slack
- DingTalk (钉钉)
- WeChat Work (企业微信)

配置方法：
1. 点击左侧边栏"配置" → "通知渠道"
2. 点击"添加通知渠道"
3. 选择类型并填写配置
4. 测试通知
5. 保存

### 4.8 用户和权限管理

#### 4.8.1 创建新用户

1. 点击左侧边栏"配置" → "用户"
2. 点击"邀请"按钮
3. 填写用户信息：
   - 邮箱地址
   - 用户名
   - 角色（Admin / Editor / Viewer）
4. 发送邀请

#### 4.8.2 用户角色说明

| 角色 | 权限 | 适用场景 |
|------|------|---------|
| **Admin** | 完全控制权限 | 系统管理员 |
| **Editor** | 可编辑仪表盘和数据源 | 开发人员、运维人员 |
| **Viewer** | 只读权限 | 普通用户、业务人员 |

### 4.9 常见问题

#### 问题 1: 仪表盘显示"No Data"

**原因**:
- Prometheus 未采集到数据
- 时间范围选择不当
- 查询语句错误

**解决方法**:
1. 检查 ceph-exporter 是否正常运行：
   ```bash
   curl http://localhost:9128/metrics
   ```
2. 检查 Prometheus targets 状态：
   访问 http://localhost:9090/targets
3. 调整时间范围到"Last 5 minutes"
4. 检查面板的查询语句

#### 问题 2: 仪表盘加载缓慢

**原因**:
- 时间范围过大
- 查询复杂度高
- 数据量过大

**解决方法**:
1. 缩小时间范围
2. 增加刷新间隔
3. 优化查询语句
4. 增加 Prometheus 资源配置

#### 问题 3: 无法登录

**原因**:
- 密码错误
- Grafana 服务未启动
- 浏览器缓存问题

**解决方法**:
1. 使用默认账号 admin/admin
2. 检查服务状态：
   ```bash
   docker ps | grep grafana
   ```
3. 清除浏览器缓存
4. 重置管理员密码：
   ```bash
   docker exec -it grafana grafana-cli admin reset-admin-password newpassword
   ```

---

## 5. Prometheus 完整操作指南

### 5.1 Prometheus 简介

Prometheus 是一个开源的监控和告警系统，负责采集、存储和查询 Ceph 集群的时序指标数据。

### 5.2 访问 Prometheus

打开浏览器访问: http://localhost:9090

### 5.3 界面布局说明

#### 主界面结构

```
┌─────────────────────────────────────────────────────────────┐
│  [Prometheus Logo]  [Graph] [Alerts] [Status] [Help]        │
├─────────────────────────────────────────────────────────────┤
│  查询输入框: [输入 PromQL 查询]                [Execute]      │
├─────────────────────────────────────────────────────────────┤
│  [Graph] [Console]                                          │
│  ┌───────────────────────────────────────────────────────┐  │
│  │                                                       │  │
│  │              图表显示区域                              │  │
│  │                                                       │  │
│  └───────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

#### 顶部导航栏

| 菜单项 | 英文 | 功能说明 |
|-------|------|---------|
| 图表 | Graph | 查询和可视化指标 |
| 告警 | Alerts | 查看告警规则和状态 |
| 状态 | Status | 查看系统状态和配置 |
| 帮助 | Help | 文档和帮助信息 |

### 5.4 常用操作

#### 5.4.1 查询指标

**基础查询**:
1. 在查询输入框中输入指标名称
2. 点击"Execute"按钮
3. 选择"Graph"或"Console"视图

**示例查询**:

```promql
# 1. 查看集群总容量
ceph_cluster_total_bytes

# 2. 查看集群使用率（百分比）
(ceph_cluster_used_bytes / ceph_cluster_total_bytes) * 100

# 3. 查看各存储池使用率
ceph_pool_percent_used

# 4. 查看 OSD 状态（1=up, 0=down）
ceph_osd_up

# 5. 查看集群健康状态（0=OK, 1=WARN, 2=ERR）
ceph_health_status

# 6. 查看过去 5 分钟的平均 IOPS
rate(ceph_cluster_read_ops_sec[5m]) + rate(ceph_cluster_write_ops_sec[5m])

# 7. 查看过去 5 分钟的平均吞吐量（MB/s）
(rate(ceph_cluster_read_bytes_sec[5m]) + rate(ceph_cluster_write_bytes_sec[5m])) / 1024 / 1024

# 8. 查看 OSD 延迟（毫秒）
ceph_osd_apply_latency_ms

# 9. 查看各存储池的对象数量
ceph_pool_objects_total

# 10. 查看 PG 状态分布
ceph_cluster_pgs_by_state
```

#### 5.4.2 使用函数和聚合

**常用函数**:

```promql
# rate() - 计算速率（每秒变化量）
rate(ceph_cluster_read_bytes_sec[5m])

# irate() - 瞬时速率（更敏感）
irate(ceph_cluster_read_bytes_sec[5m])

# sum() - 求和
sum(ceph_osd_used_bytes)

# avg() - 平均值
avg(ceph_osd_utilization)

# max() - 最大值
max(ceph_osd_utilization)

# min() - 最小值
min(ceph_osd_available_bytes)

# count() - 计数
count(ceph_osd_up == 1)

# topk() - 前 K 个最大值
topk(5, ceph_osd_utilization)

# bottomk() - 前 K 个最小值
bottomk(5, ceph_osd_available_bytes)
```

**聚合示例**:

```promql
# 按存储池聚合读写 IOPS
sum by (pool) (rate(ceph_pool_read_ops_sec[5m]) + rate(ceph_pool_write_ops_sec[5m]))

# 按 OSD 聚合使用率
avg by (osd) (ceph_osd_utilization)

# 统计 Up 状态的 OSD 数量
count(ceph_osd_up == 1)

# 统计各健康状态的数量
count by (status) (ceph_health_status_info)
```

#### 5.4.3 查看告警规则

1. 点击顶部导航栏的"Alerts"
2. 可以看到所有告警规则及其状态：
   - **Inactive** (绿色): 正常，未触发
   - **Pending** (黄色): 等待中，即将触发
   - **Firing** (红色): 已触发，正在告警

**告警规则列表**:

| 告警名称 | 严重程度 | 触发条件 | 说明 |
|---------|---------|---------|------|
| CephClusterHealthWarn | warning | 集群状态 = WARN | 集群健康警告 |
| CephClusterHealthError | critical | 集群状态 = ERR | 集群健康错误 |
| CephOSDDown | critical | OSD down > 5分钟 | OSD 宕机 |
| CephOSDOut | warning | OSD out > 10分钟 | OSD 被踢出集群 |
| CephOSDHighUsage | warning | OSD 使用率 > 80% | OSD 使用率过高 |
| CephOSDCriticalUsage | critical | OSD 使用率 > 90% | OSD 使用率严重 |
| CephOSDHighLatency | warning | OSD 延迟 > 100ms | OSD 延迟过高 |
| CephMultipleOSDsDown | emergency | 多个 OSD down | 多个 OSD 同时宕机 |
| CephClusterCapacity75 | warning | 集群使用率 > 75% | 集群容量警告 |
| CephClusterCapacity85 | critical | 集群使用率 > 85% | 集群容量严重 |
| CephClusterCapacity95 | emergency | 集群使用率 > 95% | 集群容量紧急 |
| CephPGsNotClean | warning | 非 clean PG > 5分钟 | PG 状态异常 |
| CephPGsNone | critical | PG 数量 = 0 | 没有 PG |
| CephMonitorNotInQuorum | critical | Monitor 不在仲裁 | Monitor 仲裁失败 |
| CephMonitorClockSkew | warning | Monitor 时钟偏差 > 0.05s | Monitor 时钟不同步 |

#### 5.4.4 查看 Targets 状态

1. 点击顶部导航栏的"Status" → "Targets"
2. 可以看到所有监控目标的状态：
   - **UP** (绿色): 正常采集
   - **DOWN** (红色): 采集失败
   - **UNKNOWN** (灰色): 未知状态

**Targets 列表**:

| Target | Endpoint | State | Labels |
|--------|----------|-------|--------|
| ceph-exporter | http://ceph-exporter:9128/metrics | UP | job="ceph-exporter" |
| prometheus | http://localhost:9090/metrics | UP | job="prometheus" |
| alertmanager | http://alertmanager:9093/metrics | UP | job="alertmanager" |

#### 5.4.5 查看配置

1. 点击顶部导航栏的"Status" → "Configuration"
2. 可以看到完整的 Prometheus 配置文件
3. 包括：
   - 全局配置
   - 告警规则文件
   - 抓取配置
   - Alertmanager 配置

#### 5.4.6 查看规则

1. 点击顶部导航栏的"Status" → "Rules"
2. 可以看到所有加载的告警规则
3. 包括规则名称、表达式、持续时间等

### 5.5 PromQL 查询语言详解

#### 5.5.1 基础语法

**选择器**:
```promql
# 精确匹配
ceph_osd_up{osd="0"}

# 正则匹配
ceph_osd_up{osd=~"0|1|2"}

# 反向匹配
ceph_osd_up{osd!="0"}

# 反向正则匹配
ceph_osd_up{osd!~"0|1|2"}
```

**时间范围**:
```promql
# 过去 5 分钟的数据
ceph_cluster_read_bytes_sec[5m]

# 过去 1 小时的数据
ceph_cluster_read_bytes_sec[1h]

# 过去 1 天的数据
ceph_cluster_read_bytes_sec[1d]
```

**偏移量**:
```promql
# 1 小时前的值
ceph_cluster_used_bytes offset 1h

# 1 天前的值
ceph_cluster_used_bytes offset 1d
```

#### 5.5.2 运算符

**算术运算符**:
```promql
# 加法
ceph_cluster_read_bytes_sec + ceph_cluster_write_bytes_sec

# 减法
ceph_cluster_total_bytes - ceph_cluster_used_bytes

# 乘法
ceph_cluster_used_bytes * 2

# 除法
ceph_cluster_used_bytes / ceph_cluster_total_bytes

# 取模
ceph_cluster_used_bytes % 1024

# 幂运算
ceph_cluster_used_bytes ^ 2
```

**比较运算符**:
```promql
# 等于
ceph_osd_up == 1

# 不等于
ceph_osd_up != 0

# 大于
ceph_osd_utilization > 80

# 小于
ceph_osd_utilization < 20

# 大于等于
ceph_osd_utilization >= 80

# 小于等于
ceph_osd_utilization <= 20
```

**逻辑运算符**:
```promql
# AND
ceph_osd_up == 1 and ceph_osd_in == 1

# OR
ceph_osd_up == 0 or ceph_osd_in == 0

# UNLESS (排除)
ceph_osd_up unless ceph_osd_in == 0
```

#### 5.5.3 聚合函数

```promql
# sum - 求和
sum(ceph_osd_used_bytes)

# avg - 平均值
avg(ceph_osd_utilization)

# max - 最大值
max(ceph_osd_utilization)

# min - 最小值
min(ceph_osd_available_bytes)

# count - 计数
count(ceph_osd_up)

# stddev - 标准差
stddev(ceph_osd_utilization)

# stdvar - 方差
stdvar(ceph_osd_utilization)

# topk - 前 K 个最大值
topk(5, ceph_osd_utilization)

# bottomk - 前 K 个最小值
bottomk(5, ceph_osd_available_bytes)

# quantile - 分位数
quantile(0.95, ceph_osd_utilization)
```

#### 5.5.4 时间函数

```promql
# rate - 每秒平均增长率
rate(ceph_cluster_read_bytes_sec[5m])

# irate - 瞬时增长率
irate(ceph_cluster_read_bytes_sec[5m])

# increase - 时间范围内的增长量
increase(ceph_cluster_read_bytes_sec[1h])

# delta - 时间范围内的变化量
delta(ceph_cluster_used_bytes[1h])

# idelta - 最后两个样本的变化量
idelta(ceph_cluster_used_bytes[5m])

# deriv - 导数（变化率）
deriv(ceph_cluster_used_bytes[5m])

# predict_linear - 线性预测
predict_linear(ceph_cluster_used_bytes[1h], 3600)
```

### 5.6 常见问题

#### 问题 1: Target 显示 DOWN

**原因**:
- ceph-exporter 服务未启动
- 网络连接问题
- 端口被占用

**解决方法**:
```bash
# 检查 ceph-exporter 状态
docker ps | grep ceph-exporter

# 检查端口
curl http://localhost:9128/metrics

# 查看日志
docker logs ceph-exporter

# 重启服务
docker restart ceph-exporter
```

#### 问题 2: 查询返回空结果

**原因**:
- 指标名称错误
- 时间范围选择不当
- 标签选择器错误

**解决方法**:
1. 检查指标名称是否正确
2. 调整时间范围
3. 简化查询，逐步添加条件
4. 使用"Console"视图查看原始数据

#### 问题 3: 告警规则未触发

**原因**:
- 告警条件未满足
- 持续时间未达到
- 告警规则配置错误

**解决方法**:
1. 检查告警规则表达式
2. 在"Graph"页面测试表达式
3. 检查"for"持续时间设置
4. 查看 Prometheus 日志

---

## 6. Alertmanager 完整操作指南

### 6.1 Alertmanager 简介

Alertmanager 负责接收 Prometheus 发送的告警，进行分组、去重、路由和通知。

### 6.2 访问 Alertmanager

打开浏览器访问: http://localhost:9093

### 6.3 界面布局说明

#### 主界面结构

```
┌─────────────────────────────────────────────────────────────┐
│  [Alertmanager Logo]  [Alerts] [Silences] [Status]          │
├─────────────────────────────────────────────────────────────┤
│  过滤器: [搜索告警]  [按标签过滤]                             │
├─────────────────────────────────────────────────────────────┤
│  告警列表:                                                    │
│  ┌───────────────────────────────────────────────────────┐  │
│  │ [!] CephOSDDown                                       │  │
│  │     OSD osd.0 is down                                 │  │
│  │     Severity: critical                                │  │
│  │     [Silence] [Details]                               │  │
│  └───────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

#### 顶部导航栏

| 菜单项 | 英文 | 功能说明 |
|-------|------|---------|
| 告警 | Alerts | 查看当前活跃的告警 |
| 静默 | Silences | 管理告警静默规则 |
| 状态 | Status | 查看系统状态和配置 |

### 6.4 常用操作

#### 6.4.1 查看告警

1. 点击顶部导航栏的"Alerts"
2. 可以看到所有活跃的告警
3. 告警按严重程度分组显示：
   - **Critical** (红色): 严重告警
   - **Warning** (黄色): 警告告警
   - **Info** (蓝色): 信息告警

**告警信息包括**:
- 告警名称
- 告警描述
- 严重程度
- 标签
- 触发时间
- 状态（Firing / Pending）

#### 6.4.2 静默告警

**什么是静默？**
静默（Silence）可以临时屏蔽某些告警，在维护期间或已知问题期间避免告警骚扰。

**创建静默规则**:
1. 点击顶部导航栏的"Silences"
2. 点击"New Silence"按钮
3. 填写静默规则：
   - **Matchers**: 匹配条件（标签选择器）
   - **Start**: 开始时间
   - **End**: 结束时间
   - **Duration**: 持续时间
   - **Creator**: 创建者
   - **Comment**: 备注说明
4. 点击"Create"创建

**示例：静默特定 OSD 的告警**:
```
Matchers:
  alertname = CephOSDDown
  osd = osd.0

Duration: 2h
Comment: OSD 0 正在维护，预计 2 小时后恢复
```

**示例：静默所有警告级别的告警**:
```
Matchers:
  severity = warning

Duration: 1h
Comment: 系统升级期间，暂时静默警告告警
```

#### 6.4.3 管理静默规则

**查看静默规则**:
1. 点击"Silences"标签
2. 可以看到所有静默规则：
   - **Active**: 正在生效的静默
   - **Pending**: 即将生效的静默
   - **Expired**: 已过期的静默

**编辑静默规则**:
1. 在静默列表中找到要编辑的规则
2. 点击"Edit"按钮
3. 修改配置
4. 点击"Update"保存

**删除静默规则**:
1. 在静默列表中找到要删除的规则
2. 点击"Expire"按钮
3. 静默规则立即失效

#### 6.4.4 查看状态

1. 点击顶部导航栏的"Status"
2. 可以看到：
   - **Cluster Status**: 集群状态
   - **Uptime**: 运行时间
   - **Config**: 配置信息
   - **Version Info**: 版本信息

### 6.5 告警通知配置

#### 6.5.1 配置 Webhook 通知

编辑 `alertmanager.zh-CN.yml`:

```yaml
receivers:
  - name: 'webhook-receiver'
    webhook_configs:
      - url: 'http://your-webhook-url/alert'
        send_resolved: true
```

#### 6.5.2 配置邮件通知

编辑 `alertmanager.zh-CN.yml`:

```yaml
global:
  smtp_smarthost: 'smtp.example.com:587'
  smtp_from: 'alertmanager@example.com'
  smtp_auth_username: 'alertmanager@example.com'
  smtp_auth_password: 'your-password'

receivers:
  - name: 'email-receiver'
    email_configs:
      - to: 'admin@example.com'
        headers:
          Subject: '[Ceph 告警] {{ .GroupLabels.alertname }}'
```

#### 6.5.3 配置企业微信通知

编辑 `alertmanager.zh-CN.yml`:

```yaml
receivers:
  - name: 'wechat-receiver'
    wechat_configs:
      - corp_id: 'your-corp-id'
        agent_id: 'your-agent-id'
        api_secret: 'your-api-secret'
        to_user: '@all'
        message: '{{ template "wechat.default.message" . }}'
```

### 6.6 告警路由规则

#### 6.6.1 基础路由

```yaml
route:
  group_by: ['alertname', 'component']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 1h
  receiver: 'default-receiver'
```

**参数说明**:
- `group_by`: 分组依据（按告警名称和组件分组）
- `group_wait`: 初始等待时间（收到第一个告警后等待 10 秒）
- `group_interval`: 分组间隔（同一组的后续告警等待 10 秒）
- `repeat_interval`: 重复间隔（1 小时后重新发送）
- `receiver`: 默认接收者

#### 6.6.2 高级路由

```yaml
route:
  receiver: 'default-receiver'
  routes:
    # 严重告警立即发送
    - match:
        severity: critical
      receiver: 'critical-receiver'
      group_wait: 0s
      repeat_interval: 5m

    # 警告告警延迟发送
    - match:
        severity: warning
      receiver: 'warning-receiver'
      group_wait: 30s
      repeat_interval: 1h

    # OSD 告警发送给存储团队
    - match_re:
        alertname: ^CephOSD.*
      receiver: 'storage-team'

    # 容量告警发送给容量规划团队
    - match_re:
        alertname: ^CephClusterCapacity.*
      receiver: 'capacity-team'
```

### 6.7 告警抑制规则

#### 6.7.1 什么是告警抑制？

告警抑制（Inhibition）可以在某个告警触发时，自动抑制其他相关的告警，避免告警风暴。

#### 6.7.2 配置抑制规则

```yaml
inhibit_rules:
  # 集群 ERR 时抑制 WARN
  - source_match:
      alertname: 'CephClusterHealthError'
    target_match:
      alertname: 'CephClusterHealthWarn'
    equal: ['cluster']

  # 多个 OSD 宕机时抑制单个 OSD 告警
  - source_match:
      alertname: 'CephMultipleOSDsDown'
    target_match_re:
      alertname: '^CephOSD(Down|Out)$'
    equal: ['cluster']

  # 集群容量 95% 时抑制 85% 和 75% 告警
  - source_match:
      alertname: 'CephClusterCapacity95'
    target_match_re:
      alertname: '^CephClusterCapacity(75|85)$'
    equal: ['cluster']
```

### 6.8 常见问题

#### 问题 1: 收不到告警通知

**原因**:
- 通知渠道配置错误
- 网络连接问题
- 告警路由规则错误

**解决方法**:
1. 检查 Alertmanager 配置
2. 测试通知渠道连接
3. 查看 Alertmanager 日志：
   ```bash
   docker logs alertmanager
   ```
4. 使用"amtool"测试：
   ```bash
   docker exec alertmanager amtool check-config /etc/alertmanager/alertmanager.yml
   ```

#### 问题 2: 告警重复发送

**原因**:
- `repeat_interval` 设置过短
- 告警规则持续触发
- 分组配置不当

**解决方法**:
1. 增加 `repeat_interval` 时间
2. 优化告警规则的触发条件
3. 调整 `group_by` 分组策略

#### 问题 3: 静默规则不生效

**原因**:
- 匹配条件错误
- 时间范围设置错误
- 标签选择器语法错误

**解决方法**:
1. 检查匹配条件是否正确
2. 确认时间范围包含当前时间
3. 使用正则表达式时注意语法
4. 在"Alerts"页面查看告警的实际标签

---

## 7. ELK Stack 完整操作指南

### 7.1 ELK Stack 简介

ELK Stack 由 Elasticsearch、Logstash 和 Kibana 三个组件组成，提供完整的日志收集、存储、搜索和可视化解决方案。

### 7.2 访问 Kibana

打开浏览器访问: http://localhost:5601

### 7.3 Kibana 界面操作

#### 7.3.1 首次访问

1. 打开 Kibana 后，首先需要创建索引模式（Index Pattern）
2. 点击左侧菜单"Management" → "Index Patterns"
3. 点击"Create index pattern"
4. 输入索引模式：`ceph-exporter-*`
5. 选择时间字段：`@timestamp`
6. 点击"Create index pattern"

#### 7.3.2 查看日志

1. 点击左侧菜单"Discover"
2. 选择索引模式：`ceph-exporter-*`
3. 调整时间范围（右上角）
4. 可以看到所有日志记录

#### 7.3.3 搜索日志

**基础搜索**:
```
# 搜索包含"error"的日志
error

# 搜索特定级别的日志
level:error

# 搜索特定组件的日志
component:collector

# 搜索特定时间范围
@timestamp:[2026-03-15 TO 2026-03-16]
```

**高级搜索（KQL）**:
```
# AND 条件
level:error AND component:collector

# OR 条件
level:error OR level:warn

# NOT 条件
NOT level:info

# 通配符
message:*timeout*

# 范围查询
response_time > 100

# 存在性查询
_exists_:trace_id
```

#### 7.3.4 创建可视化

1. 点击左侧菜单"Visualize"
2. 点击"Create visualization"
3. 选择可视化类型：
   - Line (折线图)
   - Area (面积图)
   - Bar (柱状图)
   - Pie (饼图)
   - Data Table (数据表)
   - Metric (指标)
4. 选择数据源
5. 配置聚合和指标
6. 保存可视化

#### 7.3.5 创建仪表盘

1. 点击左侧菜单"Dashboard"
2. 点击"Create dashboard"
3. 点击"Add"添加可视化
4. 调整面板大小和位置
5. 保存仪表盘

### 7.4 日志配置

#### 7.4.1 启用 ELK 日志

编辑 `ceph-exporter.yaml`:

```yaml
logger:
  level: "info"
  format: "json"
  output: "stdout"
  enable_elk: true
  logstash_url: "logstash:5044"
  logstash_protocol: "tcp"
  service_name: "ceph-exporter"
```

#### 7.4.2 日志级别说明

| 级别 | 英文 | 说明 | 使用场景 |
|------|------|------|---------|
| trace | TRACE | 最详细的日志 | 开发调试 |
| debug | DEBUG | 调试信息 | 问题排查 |
| info | INFO | 一般信息 | 生产环境（推荐）|
| warn | WARN | 警告信息 | 潜在问题 |
| error | ERROR | 错误信息 | 错误记录 |
| fatal | FATAL | 致命错误 | 严重错误 |
| panic | PANIC | 恐慌错误 | 程序崩溃 |

### 7.5 常见问题

#### 问题 1: Kibana 无法连接 Elasticsearch

**解决方法**:
```bash
# 检查 Elasticsearch 状态
curl http://localhost:9200

# 检查 Kibana 日志
docker logs kibana

# 重启服务
docker restart elasticsearch kibana
```

#### 问题 2: 没有日志数据

**解决方法**:
1. 检查 ceph-exporter 配置中 `enable_elk` 是否为 `true`
2. 检查 Logstash 是否正常运行
3. 检查网络连接
4. 查看 ceph-exporter 日志

---

## 8. Jaeger 分布式追踪指南

### 8.1 Jaeger 简介

Jaeger 是一个开源的分布式追踪系统，用于监控和排查微服务架构中的性能问题。

### 8.2 访问 Jaeger UI

打开浏览器访问: http://localhost:16686

### 8.3 界面操作

#### 8.3.1 搜索追踪

1. 在左侧选择"Service": `ceph-exporter`
2. 选择"Operation":
   - `HTTP GET /metrics`
   - `HTTP GET /health`
3. 设置时间范围
4. 点击"Find Traces"

#### 8.3.2 查看追踪详情

1. 在搜索结果中点击某个追踪
2. 可以看到：
   - 追踪 ID
   - 开始时间
   - 持续时间
   - Span 列表
   - 时间线视图

#### 8.3.3 分析性能

- 查看各个 Span 的耗时
- 识别性能瓶颈
- 查看错误和异常

### 8.4 启用追踪

编辑 `ceph-exporter.yaml`:

```yaml
tracer:
  enabled: true
  jaeger_url: "jaeger:4318"
  service_name: "ceph-exporter"
  sample_rate: 1.0  # 1.0 = 100% 采样
```

或使用脚本启用：

```bash
cd /home/lfl/ceph-exporter/ceph-exporter/deployments
./scripts/enable-jaeger-tracing.sh
```

---

## 9. Ceph Dashboard 操作指南

### 9.1 访问 Ceph Dashboard

打开浏览器访问: http://localhost:8080

### 9.2 主要功能

- 集群状态概览
- OSD 管理
- 存储池管理
- Monitor 状态
- 性能监控

---

## 10. 界面术语对照表

### 10.1 Grafana 术语

| 英文术语 | 中文翻译 | 说明 |
|---------|---------|------|
| Dashboard | 仪表盘 | 包含多个面板的可视化页面 |
| Panel | 面板 | 单个图表或可视化组件 |
| Data Source | 数据源 | 数据来源（如 Prometheus）|
| Query | 查询 | 数据查询语句 |
| Time Range | 时间范围 | 数据的时间跨度 |
| Refresh | 刷新 | 更新数据 |
| Variables | 变量 | 动态参数 |
| Annotations | 注释 | 事件标记 |
| Alert | 告警 | 告警规则 |
| Threshold | 阈值 | 告警触发条件 |
| Legend | 图例 | 数据系列说明 |
| Tooltip | 提示框 | 鼠标悬停显示的信息 |
| Axis | 坐标轴 | X轴/Y轴 |
| Series | 系列 | 数据序列 |
| Aggregation | 聚合 | 数据聚合方式 |

### 10.2 Prometheus 术语

| 英文术语 | 中文翻译 | 说明 |
|---------|---------|------|
| Metric | 指标 | 监控数据项 |
| Label | 标签 | 指标的维度 |
| Target | 目标 | 监控目标（被监控的服务）|
| Scrape | 抓取 | 采集指标数据 |
| Scrape Interval | 抓取间隔 | 采集频率 |
| Evaluation Interval | 评估间隔 | 告警规则评估频率 |
| Alert Rule | 告警规则 | 告警触发条件 |
| Recording Rule | 记录规则 | 预计算规则 |
| Instant Query | 即时查询 | 查询当前值 |
| Range Query | 范围查询 | 查询时间范围内的值 |
| Selector | 选择器 | 标签匹配条件 |
| Aggregation | 聚合 | 数据聚合操作 |
| Function | 函数 | PromQL 函数 |
| Operator | 运算符 | 算术/比较/逻辑运算符 |
| Time Series | 时间序列 | 带时间戳的数据序列 |

### 10.3 Alertmanager 术语

| 英文术语 | 中文翻译 | 说明 |
|---------|---------|------|
| Alert | 告警 | 告警事件 |
| Silence | 静默 | 临时屏蔽告警 |
| Inhibition | 抑制 | 告警抑制规则 |
| Route | 路由 | 告警路由规则 |
| Receiver | 接收者 | 告警通知目标 |
| Group | 分组 | 告警分组 |
| Firing | 触发中 | 告警正在触发 |
| Pending | 等待中 | 告警即将触发 |
| Resolved | 已解决 | 告警已恢复 |
| Severity | 严重程度 | 告警级别 |
| Matcher | 匹配器 | 标签匹配条件 |
| Notification | 通知 | 告警通知 |

### 10.4 Ceph 术语

| 英文术语 | 中文翻译 | 说明 |
|---------|---------|------|
| Cluster | 集群 | Ceph 存储集群 |
| Pool | 存储池 | 数据存储池 |
| OSD | 对象存储守护进程 | 存储节点 |
| Monitor | 监视器 | 集群监控节点 |
| MDS | 元数据服务器 | CephFS 元数据服务 |
| RGW | 对象网关 | S3/Swift 兼容网关 |
| PG | 归置组 | 数据分布单元 |
| RADOS | 可靠自主分布式对象存储 | Ceph 底层存储系统 |
| CRUSH | 可扩展哈希下的受控复制 | 数据分布算法 |
| Health | 健康状态 | 集群健康状态 |
| Capacity | 容量 | 存储容量 |
| Utilization | 利用率 | 使用率 |
| Latency | 延迟 | 响应延迟 |
| Throughput | 吞吐量 | 数据传输速率 |
| IOPS | 每秒IO操作数 | 性能指标 |

### 10.5 ELK 术语

| 英文术语 | 中文翻译 | 说明 |
|---------|---------|------|
| Index | 索引 | 数据存储单元 |
| Document | 文档 | 单条日志记录 |
| Field | 字段 | 文档属性 |
| Mapping | 映射 | 字段类型定义 |
| Query | 查询 | 搜索语句 |
| Filter | 过滤器 | 数据过滤条件 |
| Aggregation | 聚合 | 数据统计分析 |
| Visualization | 可视化 | 图表展示 |
| Dashboard | 仪表盘 | 可视化集合 |
| Discover | 发现 | 日志浏览页面 |
| Kibana Query Language (KQL) | Kibana 查询语言 | 搜索语法 |

---

## 11. 常用操作速查

### 11.1 部署操作

```bash
# 完整部署
cd /home/lfl/ceph-exporter/ceph-exporter/deployments
./scripts/deploy.sh full

# 查看状态
./scripts/deploy.sh status

# 验证部署
./scripts/deploy.sh verify

# 查看日志
./scripts/deploy.sh logs [service-name]

# 诊断问题
./scripts/deploy.sh diagnose

# 修复部署
sudo ./scripts/deploy.sh fix

# 停止服务
./scripts/deploy.sh stop

# 清理数据
./scripts/deploy.sh clean
```

### 11.2 服务管理

```bash
# 重启单个服务
docker restart ceph-exporter
docker restart prometheus
docker restart grafana
docker restart alertmanager

# 查看服务日志
docker logs -f ceph-exporter
docker logs -f prometheus
docker logs -f grafana

# 进入容器
docker exec -it ceph-exporter /bin/bash
docker exec -it prometheus /bin/sh

# 查看资源使用
docker stats
```

### 11.3 数据备份

```bash
# 备份 Prometheus 数据
tar -czf prometheus-backup.tar.gz data/prometheus/

# 备份 Grafana 数据
tar -czf grafana-backup.tar.gz data/grafana/

# 备份 Elasticsearch 数据
tar -czf elasticsearch-backup.tar.gz data/elasticsearch/

# 备份所有数据
tar -czf ceph-exporter-backup-$(date +%Y%m%d).tar.gz data/
```

### 11.4 常用查询

**Prometheus 查询**:
```promql
# 集群使用率
(ceph_cluster_used_bytes / ceph_cluster_total_bytes) * 100

# 集群 IOPS
rate(ceph_cluster_read_ops_sec[5m]) + rate(ceph_cluster_write_ops_sec[5m])

# OSD 状态统计
count(ceph_osd_up == 1)

# 存储池使用率 Top 5
topk(5, ceph_pool_percent_used)
```

**Kibana 查询**:
```
# 错误日志
level:error

# 特定组件日志
component:collector

# 包含追踪 ID 的日志
_exists_:trace_id

# 响应时间超过 100ms
response_time > 100
```

---

## 12. 常见问题与解决方案

### 12.1 部署问题

#### 问题：Docker 镜像拉取失败

**症状**:
```
Error response from daemon: Get https://registry-1.docker.io/v2/: net/http: TLS handshake timeout
```

**解决方法**:
1. 配置 Docker 镜像加速器
2. 使用国内镜像源
3. 参考 `DOCKER_MIRROR_CONFIGURATION.md`

#### 问题：端口被占用

**症状**:
```
Error starting userland proxy: listen tcp 0.0.0.0:9090: bind: address already in use
```

**解决方法**:
```bash
# 查找占用端口的进程
sudo lsof -i :9090

# 停止占用进程
sudo kill -9 <PID>

# 或修改 docker-compose.yml 中的端口映射
```

#### 问题：权限不足

**症状**:
```
mkdir: cannot create directory '/var/lib/prometheus': Permission denied
```

**解决方法**:
```bash
# 运行修复脚本
sudo ./scripts/deploy.sh fix

# 或手动修复权限
sudo chown -R 65534:65534 data/prometheus
sudo chown -R 472:472 data/grafana
```

### 12.2 监控问题

#### 问题：Grafana 显示 "No Data"

**原因**:
- Prometheus 未采集到数据
- ceph-exporter 未正常运行
- 时间范围选择不当

**解决方法**:
```bash
# 1. 检查 ceph-exporter
curl http://localhost:9128/metrics

# 2. 检查 Prometheus targets
curl http://localhost:9090/api/v1/targets

# 3. 查看日志
docker logs ceph-exporter
docker logs prometheus

# 4. 重启服务
docker restart ceph-exporter prometheus
```

#### 问题：告警未触发

**原因**:
- 告警规则配置错误
- 持续时间未达到
- Alertmanager 未正常运行

**解决方法**:
```bash
# 1. 检查告警规则
curl http://localhost:9090/api/v1/rules

# 2. 测试告警表达式
# 在 Prometheus UI 中执行查询

# 3. 检查 Alertmanager
curl http://localhost:9093/api/v1/status

# 4. 查看日志
docker logs alertmanager
```

### 12.3 性能问题

#### 问题：Prometheus 内存占用过高

**原因**:
- 数据保留时间过长
- 抓取间隔过短
- 指标数量过多

**解决方法**:
1. 调整数据保留时间：
   ```yaml
   # prometheus.yml
   global:
     scrape_interval: 30s  # 增加抓取间隔

   # 启动参数
   --storage.tsdb.retention.time=15d  # 减少保留时间
   ```

2. 增加内存限制：
   ```yaml
   # docker-compose.yml
   services:
     prometheus:
       mem_limit: 2g
   ```

#### 问题：Grafana 仪表盘加载缓慢

**原因**:
- 查询时间范围过大
- 查询复杂度高
- 数据量过大

**解决方法**:
1. 缩小时间范围
2. 增加刷新间隔
3. 优化查询语句
4. 使用记录规则预计算

### 12.4 日志问题

#### 问题：Kibana 无日志数据

**原因**:
- ELK 未启用
- Logstash 未正常运行
- 网络连接问题

**解决方法**:
```bash
# 1. 检查配置
grep enable_elk configs/ceph-exporter.yaml

# 2. 检查 Logstash
docker logs logstash

# 3. 检查 Elasticsearch
curl http://localhost:9200/_cat/indices

# 4. 重启服务
docker restart logstash elasticsearch kibana
```

---

## 13. 最佳实践

### 13.1 监控最佳实践

1. **合理设置告警阈值**
   - 根据实际业务需求调整
   - 避免告警过于敏感或迟钝
   - 定期review和优化

2. **使用分层告警**
   - Warning: 需要关注但不紧急
   - Critical: 需要立即处理
   - Emergency: 严重影响业务

3. **配置告警通知渠道**
   - 工作时间：邮件 + 即时通讯
   - 非工作时间：电话 + 短信
   - 严重告警：多渠道通知

4. **定期检查监控系统**
   - 验证告警规则是否有效
   - 检查数据采集是否正常
   - 清理过期的静默规则

### 13.2 性能优化

1. **Prometheus 优化**
   - 合理设置数据保留时间
   - 使用记录规则预计算
   - 控制指标数量和标签基数

2. **Grafana 优化**
   - 使用变量减少重复查询
   - 设置合理的刷新间隔
   - 避免过于复杂的查询

3. **ELK 优化**
   - 定期清理旧索引
   - 使用索引生命周期管理
   - 控制日志级别和数量

### 13.3 安全建议

1. **修改默认密码**
   - Grafana: admin/admin
   - 其他服务的默认凭据

2. **启用 HTTPS**
   - 配置 TLS 证书
   - 强制使用 HTTPS

3. **访问控制**
   - 配置防火墙规则
   - 使用反向代理
   - 实施 IP 白名单

4. **定期备份**
   - 备份配置文件
   - 备份监控数据
   - 测试恢复流程

---

## 14. 附录

### 14.1 配置文件位置

```
/home/lfl/ceph-exporter/ceph-exporter/
├── configs/
│   ├── ceph-exporter.yaml          # 主配置文件
│   ├── ceph-exporter.zh-CN.yaml    # 中文配置文件
│   └── logger-examples.yaml        # 日志配置示例
├── deployments/
│   ├── prometheus/
│   │   ├── prometheus.yml          # Prometheus 配置
│   │   ├── prometheus.zh-CN.yml    # Prometheus 中文配置
│   │   ├── alert_rules.yml         # 告警规则
│   │   └── alert_rules.zh-CN.yml   # 告警规则（中文）
│   ├── alertmanager/
│   │   ├── alertmanager.yml        # Alertmanager 配置
│   │   └── alertmanager.zh-CN.yml  # Alertmanager 中文配置
│   ├── grafana/
│   │   ├── dashboards/             # 仪表盘 JSON 文件
│   │   └── provisioning/           # 自动配置
│   └── docker-compose-*.yml        # Docker Compose 配置
```

### 14.2 端口列表

| 服务 | 端口 | 协议 | 说明 |
|------|------|------|------|
| ceph-exporter | 9128 | HTTP | Prometheus 指标 |
| Prometheus | 9090 | HTTP | Web UI 和 API |
| Grafana | 3000 | HTTP | Web UI |
| Alertmanager | 9093 | HTTP | Web UI 和 API |
| Elasticsearch | 9200 | HTTP | REST API |
| Elasticsearch | 9300 | TCP | 节点通信 |
| Logstash | 5044 | TCP | Beats 输入 |
| Logstash | 5000 | TCP | TCP 输入 |
| Kibana | 5601 | HTTP | Web UI |
| Jaeger Collector | 4318 | HTTP | OTLP HTTP |
| Jaeger UI | 16686 | HTTP | Web UI |
| Ceph Dashboard | 8080 | HTTP | Web UI |

### 14.3 相关文档

- [快速开始指南](QUICK_START.md)
- [部署指南](DEPLOYMENT_GUIDE.md)
- [Docker 镜像配置](DOCKER_MIRROR_CONFIGURATION.md)
- [故障排查指南](ceph-exporter/deployments/TROUBLESHOOTING.md)
- [ELK 日志指南](ceph-exporter/docs/ELK-LOGGING-GUIDE.md)
- [Jaeger 追踪指南](ceph-exporter/docs/JAEGER-TRACING-GUIDE.md)
- [Prometheus 使用指南](Prometheus使用指南.md)
- [Alertmanager 使用指南](Alertmanager使用指南.md)

### 14.4 版本信息

- **文档版本**: 4.0
- **最后更新**: 2026-03-15
- **适用版本**: ceph-exporter 1.0+
- **Ceph 版本**: Luminous (12.x) 及以上

---

**文档结束**

如有问题或建议，请参考项目 README 或提交 Issue。
