# Ceph-Exporter 项目完整操作指南

> **版本**: 5.0
> **最后更新**: 2026-03-22
> **适用环境**: Ubuntu 20.04 + Docker
> **文档状态**: 全面更新 - 包含详细操作示例、界面说明、术语对照表

---

## 目录

1. [项目概述](#1-项目概述)
2. [系统架构](#2-系统架构)
3. [快速开始](#3-快速开始)
4. [Ceph-Exporter 服务操作指南](#4-ceph-exporter-服务操作指南)
5. [Grafana 完整操作指南](#5-grafana-完整操作指南)
6. [Prometheus 完整操作指南](#6-prometheus-完整操作指南)
7. [Alertmanager 完整操作指南](#7-alertmanager-完整操作指南)
8. [ELK Stack 完整操作指南](#8-elk-stack-完整操作指南)
9. [Jaeger 分布式追踪指南](#9-jaeger-分布式追踪指南)
10. [Ceph Dashboard 操作指南](#10-ceph-dashboard-操作指南)
11. [部署脚本工具集](#11-部署脚本工具集)
12. [界面术语对照表](#12-界面术语对照表)
13. [常用操作速查](#13-常用操作速查)
14. [常见问题与解决方案](#14-常见问题与解决方案)
15. [最佳实践](#15-最佳实践)
16. [附录](#16-附录)

---

## 1. 项目概述

### 1.1 什么是 Ceph-Exporter？

Ceph-Exporter 是一个基于 Go 语言开发的 Ceph 集群 Prometheus 指标导出器，提供完整的监控、日志和追踪解决方案。通过 CGO 集成 go-ceph 库，直接与 Ceph RADOS 通信获取集群状态数据。

### 1.2 核心特性

**功能特性**:
- 7 个 Prometheus 采集器: Cluster、Pool、OSD、Monitor、Health、MDS、RGW
- 完全中文化: Grafana 100% 中文界面，所有配置文件中文注释
- 完整可观测性: 指标监控 + 日志分析 + 分布式追踪
- 一键部署: Docker Compose 自动化部署，支持多种模式
- 生产级配置: 20+ 条告警规则，自动化诊断和修复脚本

**技术特性**:
- 70+ 指标: 覆盖集群、存储池、OSD、Monitor、MDS、RGW 等
- CGO 集成: 使用 go-ceph 库直接与 Ceph RADOS 通信
- 完整测试: 90 个单元测试，100% 通过率，覆盖率 68.1%
- 高性能: 并发采集，带超时控制（默认 10 秒）
- 安全可靠: 支持 TLS/HTTPS，优雅关闭

### 1.3 系统要求

| 项目 | 最低要求 | 推荐配置 |
|------|---------|---------|
| 操作系统 | Ubuntu 20.04 | Ubuntu 20.04 |
| Docker | 19.03+ | 20.10+ |
| Docker Compose | V2 (插件) | V2 (插件) |
| 内存 | 4GB | 8GB |
| CPU | 2 核 | 4 核 |
| 磁盘 | 30GB | 50GB |
| Ceph 版本 | Octopus (15.x) | Octopus (15.x) / Pacific (16.x) |

### 1.4 服务访问地址

| 服务 | 访问地址 | 账号 | 用途 |
|------|---------|------|------|
| **Grafana** | http://localhost:3000 | admin/admin | 可视化仪表板 |
| **Prometheus** | http://localhost:9090 | - | 指标查询和告警 |
| **Alertmanager** | http://localhost:9093 | - | 告警管理 |
| **Kibana** | http://localhost:5601 | - | 日志查询和分析 |
| **Elasticsearch** | http://localhost:9200 | - | 日志存储 |
| **Jaeger UI** | http://localhost:16686 | - | 分布式追踪 |
| **Ceph Dashboard** | http://localhost:8080 | - | Ceph 集群管理 |
| **Ceph RGW (S3)** | http://localhost:5000 | - | 对象存储网关 |
| **ceph-exporter** | http://localhost:9128/metrics | - | Prometheus 指标端点 |

### 1.5 部署模式

本项目提供 4 种部署模式，适用于不同场景：

| 模式 | 命令 | 包含服务 | 适用场景 |
|------|------|---------|---------|
| **完整栈** | `deploy.sh full` | 全部 10 个服务 | 生产环境、完整测试 |
| **最小栈** | `deploy.sh minimal` | ceph-exporter + Prometheus + Grafana + Alertmanager | 轻量监控 |
| **集成测试** | `deploy.sh integration` | Ceph Demo + 监控栈 | 自动化测试 |
| **仅 Ceph** | `deploy.sh ceph-demo` | Ceph Demo 集群 | 开发调试 |

---

## 2. 系统架构

### 2.1 整体架构图

```
┌─────────────────────────────────────────────────────────────────┐
│                          用户访问层                               │
│  Grafana (3000)  │  Prometheus (9090)  │  Kibana (5601)         │
│  Jaeger UI (16686)  │  Alertmanager (9093)  │ Ceph Dashboard    │
└─────────────────────────────────────────────────────────────────┘
                              ▲
                              │ HTTP/API
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                         监控采集层                                │
│  ceph-exporter (9128)  │  Logstash (5000)  │  Jaeger (4318)     │
│  Filebeat (sidecar)                                              │
└─────────────────────────────────────────────────────────────────┘
                              ▲
                              │ librados / Docker API
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
   Ceph Cluster → ceph-exporter (librados) → Prometheus (HTTP 抓取) → Grafana (可视化)
   ```

2. **日志采集流程（3 种方案）**:
   ```
   方案1: ceph-exporter → TCP/UDP → Logstash → Elasticsearch → Kibana
   方案2: ceph-exporter → stdout → Docker → Filebeat sidecar → Logstash → Elasticsearch → Kibana
   方案3: ceph-exporter → 日志文件 → Filebeat → Logstash → Elasticsearch → Kibana
   ```

3. **追踪流程**:
   ```
   HTTP Request → ceph-exporter (OpenTelemetry SDK) → Jaeger Collector (OTLP HTTP 4318) → Jaeger UI
   ```

4. **告警流程**:
   ```
   Prometheus (评估规则 15s) → Alertmanager (分组/路由/抑制) → Webhook/邮件/企业微信
   ```

### 2.3 Docker 网络架构

完整栈部署使用 4 个隔离网络：

| 网络名称 | 用途 | 连接的服务 |
|---------|------|-----------|
| ceph-network | Ceph 集群通信 | ceph-demo, ceph-exporter |
| monitor-network | 监控栈通信 | ceph-exporter, prometheus, grafana, alertmanager |
| logging-network | 日志栈通信 | ceph-exporter, logstash, elasticsearch, kibana, filebeat |
| tracing-network | 追踪栈通信 | ceph-exporter, jaeger |

### 2.4 服务资源限制

完整栈部署中各服务的资源配置：

| 服务 | 内存限制 | 镜像版本 |
|------|---------|---------|
| ceph-demo | 1024MB | quay.io/ceph/daemon:latest-octopus |
| ceph-exporter | 128MB | 本地构建 (ceph-exporter:dev) |
| prometheus | 512MB | prom/prometheus:v2.51.0 |
| grafana | 256MB | grafana/grafana:10.4.0 |
| alertmanager | 128MB | prom/alertmanager:v0.25.0 |
| elasticsearch | 512MB | elasticsearch:7.17.0 |
| logstash | 768MB | logstash:7.17.0 |
| kibana | 1024MB | kibana:7.17.0 |
| filebeat-sidecar | 128MB | filebeat:7.17.0 |
| jaeger | 256MB | jaegertracing/all-in-one:1.35 |

---

## 3. 快速开始

### 3.1 环境准备

```bash
# 1. 检查 Docker 是否安装
docker --version
docker compose version

# 2. 如果未安装 Docker，执行以下命令
sudo apt-get update
sudo apt-get install -y ca-certificates curl gnupg
sudo install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

# 3. 将当前用户添加到 docker 组（避免每次使用 sudo）
sudo usermod -aG docker $USER
newgrp docker

# 4. 设置 Elasticsearch 所需的系统参数
sudo sysctl -w vm.max_map_count=262144
```

### 3.2 一键部署

```bash
# 进入部署目录
cd ceph-exporter/deployments

# 完整监控栈部署（推荐，部署时交互式选择日志方案）
./scripts/deploy.sh full

# 或指定日志方案（跳过交互）
LOGGING_MODE=container ./scripts/deploy.sh full   # 容器日志收集（推荐）
LOGGING_MODE=direct ./scripts/deploy.sh full      # 直接推送到 Logstash (TCP)
LOGGING_MODE=direct-udp ./scripts/deploy.sh full  # 直接推送到 Logstash (UDP)
LOGGING_MODE=file ./scripts/deploy.sh full         # 文件日志 + Filebeat
LOGGING_MODE=dev ./scripts/deploy.sh full          # 开发模式 (stdout + text)

# 等待服务启动（约 2-3 分钟）
./scripts/deploy.sh status

# 验证部署
./scripts/deploy.sh verify
```

部署脚本会自动执行以下步骤：
1. 检查操作系统和 Docker 环境
2. 检查系统资源（内存、CPU、磁盘）
3. 配置 Docker 镜像加速器（国内源）
4. 配置防火墙规则
5. 预拉取所有镜像
6. 初始化数据目录并设置权限
7. 构建 ceph-exporter:dev 镜像
8. 选择并应用日志方案
9. 分阶段启动所有服务
10. 修复 Ceph keyring 权限
11. 生成测试数据

### 3.3 访问服务

部署完成后，打开浏览器访问：

1. **Grafana** (推荐首先访问): http://localhost:3000
   - 账号: `admin`，密码: `admin`
   - 首次登录会提示修改密码（可跳过）
   - 已预配置 Ceph 集群监控仪表盘

2. **Prometheus**: http://localhost:9090
   - 查看指标和告警规则
   - 验证 targets 状态

3. **Alertmanager**: http://localhost:9093
   - 查看和管理告警

4. **Kibana**: http://localhost:5601
   - 首次需要创建索引模式 `ceph-exporter-*`

5. **Jaeger**: http://localhost:16686
   - Service 选择 `ceph-exporter`

### 3.4 验证指标采集

```bash
# 查看 ceph-exporter 指标
curl http://localhost:9128/metrics | grep ceph_

# 查看健康检查
curl http://localhost:9128/health

# 查看 Prometheus targets 状态
curl http://localhost:9090/api/v1/targets | python3 -m json.tool

# 查看告警规则
curl http://localhost:9090/api/v1/rules | python3 -m json.tool

# 查看 Elasticsearch 索引
curl http://localhost:9200/_cat/indices?v

# 查看 Alertmanager 状态
curl http://localhost:9093/api/v1/status | python3 -m json.tool
```

---

## 4. Ceph-Exporter 服务操作指南

### 4.1 服务简介

ceph-exporter 是本项目的核心服务，负责连接 Ceph 集群并将监控指标以 Prometheus 格式暴露。

### 4.2 配置文件详解

主配置文件位于 `ceph-exporter/configs/ceph-exporter.yaml`，包含以下配置段：

#### 4.2.1 HTTP 服务器配置

```yaml
server:
  host: "0.0.0.0"              # 监听地址，默认监听所有接口
  port: 9128                    # 监听端口，Prometheus 拉取指标的端口
  read_timeout: 30s             # HTTP 读取超时
  write_timeout: 30s            # HTTP 写入超时
  tls_cert_file: ""             # TLS 证书文件路径（留空禁用 HTTPS）
  tls_key_file: ""              # TLS 密钥文件路径（留空禁用 HTTPS）
```

#### 4.2.2 Ceph 连接配置

```yaml
ceph:
  config_file: "/etc/ceph/ceph.conf"                  # Ceph 配置文件路径
  user: "admin"                                        # Ceph 认证用户名
  keyring: "/etc/ceph/ceph.client.admin.keyring"       # Keyring 认证文件路径
  cluster: "ceph"                                      # Ceph 集群名称
  timeout: 10s                                         # 命令执行超时
```

环境变量覆盖：`CEPH_CONFIG`, `CEPH_USER`, `CEPH_KEYRING`, `CEPH_CLUSTER`

#### 4.2.3 Prometheus 采集配置

```yaml
prometheus:
  collect_interval: 15s         # 指标采集间隔
  timeout: 10s                  # 单次采集超时
```

#### 4.2.4 日志配置

```yaml
logger:
  level: "debug"                # 日志级别: trace, debug, info, warn, error, fatal, panic
  format: "json"                # 日志格式: json (结构化), text (纯文本)
  output: "stdout"              # 输出目标: stdout, stderr, file
  file_path: "/var/log/ceph-exporter/ceph-exporter.log"  # 日志文件路径
  max_size: 100                 # 单个日志文件最大大小 (MB)
  max_backups: 3                # 保留旧日志文件数量
  max_age: 28                   # 日志文件最大保留天数
  compress: true                # 是否压缩归档旧日志文件

  # ELK 集成（方案 1: 直接推送到 Logstash）
  enable_elk: false             # 是否启用直接推送到 Logstash
  logstash_url: "logstash:5000" # Logstash 地址 (host:port)
  logstash_protocol: "tcp"      # 协议: tcp (可靠) 或 udp (高性能)
  service_name: "ceph-exporter" # 服务名称，用于 ELK 中的日志标识
```

环境变量覆盖：`LOG_LEVEL`, `LOG_FORMAT`, `LOGSTASH_URL`

#### 4.2.5 追踪配置

```yaml
tracer:
  enabled: true                 # 是否启用追踪
  jaeger_url: "jaeger:4318"     # Jaeger OTLP HTTP 端点 (host:port，不含 http://)
  service_name: "ceph-exporter" # 服务名称
  sample_rate: 1.0              # 采样率 (0.0-1.0，1.0 = 100% 采样)
```

环境变量覆盖：`JAEGER_URL`, `SERVICE_NAME`

#### 4.2.6 插件配置（预留）

```yaml
plugins:
  - name: "example-storage"
    enabled: false
    type: "http"
    path: "http://example-storage:8080"
    config:
      endpoint: "http://example-storage:8080"
      timeout: "10s"
```

### 4.3 日志方案选择

本项目支持 3 种日志推送方案，可通过脚本快速切换：

| 方案 | 配置 | 优点 | 缺点 | 适用场景 |
|------|------|------|------|---------|
| **方案1: 直接推送** | enable_elk: true | 实时推送，无需额外组件 | Logstash 故障影响日志 | 生产环境（直连） |
| **方案2: 容器日志收集** | enable_elk: false + Filebeat sidecar | 解耦，故障不影响应用 | 需要部署 Filebeat | 生产环境（推荐） |
| **方案3: 文件日志** | output: file + Filebeat | 日志持久化，可离线分析 | 需要管理磁盘空间 | 需要日志持久化 |

切换日志方案：

```bash
cd ceph-exporter/deployments

# 切换到容器日志收集模式（推荐）
./scripts/switch-logging-mode.sh container

# 切换到直接推送模式 (TCP)
./scripts/switch-logging-mode.sh direct

# 切换到直接推送模式 (UDP，高性能)
./scripts/switch-logging-mode.sh direct-udp

# 切换到文件日志模式
./scripts/switch-logging-mode.sh file

# 切换到开发模式
./scripts/switch-logging-mode.sh dev

# 查看当前配置
./scripts/switch-logging-mode.sh show
```

### 4.4 采集器详解

ceph-exporter 包含 7 个采集器，每个采集器负责采集特定类型的 Ceph 指标：

#### 4.4.1 集群采集器 (Cluster Collector)

采集集群级别的整体状态指标。

| 指标名称 | 类型 | 说明 |
|---------|------|------|
| `ceph_cluster_total_bytes` | Gauge | 集群总容量（字节） |
| `ceph_cluster_used_bytes` | Gauge | 集群已用容量（字节） |
| `ceph_cluster_available_bytes` | Gauge | 集群可用容量（字节） |
| `ceph_cluster_objects_total` | Gauge | 集群对象总数 |
| `ceph_cluster_read_bytes_sec` | Gauge | 集群读取吞吐量（字节/秒） |
| `ceph_cluster_write_bytes_sec` | Gauge | 集群写入吞吐量（字节/秒） |
| `ceph_cluster_read_ops_sec` | Gauge | 集群读取 IOPS |
| `ceph_cluster_write_ops_sec` | Gauge | 集群写入 IOPS |
| `ceph_cluster_pgs_total` | Gauge | PG 总数 |
| `ceph_cluster_pgs_by_state` | Gauge | 各状态 PG 数量（标签: state） |
| `ceph_cluster_pools_total` | Gauge | 存储池总数 |
| `ceph_cluster_osds_total` | Gauge | OSD 总数 |
| `ceph_cluster_osds_up` | Gauge | Up 状态 OSD 数量 |
| `ceph_cluster_osds_in` | Gauge | In 状态 OSD 数量 |
| `ceph_cluster_mons_total` | Gauge | Monitor 总数 |

#### 4.4.2 OSD 采集器 (OSD Collector)

采集每个 OSD 的详细状态指标。标签: `osd`（OSD 编号）

| 指标名称 | 类型 | 说明 |
|---------|------|------|
| `ceph_osd_up` | Gauge | OSD 是否 Up (1=是, 0=否) |
| `ceph_osd_in` | Gauge | OSD 是否 In (1=是, 0=否) |
| `ceph_osd_total_bytes` | Gauge | OSD 总容量（字节） |
| `ceph_osd_used_bytes` | Gauge | OSD 已用容量（字节） |
| `ceph_osd_available_bytes` | Gauge | OSD 可用容量（字节） |
| `ceph_osd_utilization` | Gauge | OSD 利用率（百分比） |
| `ceph_osd_pgs` | Gauge | OSD 上的 PG 数量 |
| `ceph_osd_apply_latency_ms` | Gauge | Apply 延迟（毫秒） |
| `ceph_osd_commit_latency_ms` | Gauge | Commit 延迟（毫秒） |

#### 4.4.3 存储池采集器 (Pool Collector)

采集每个存储池的状态指标。标签: `pool`（存储池名称）

| 指标名称 | 类型 | 说明 |
|---------|------|------|
| `ceph_pool_stored_bytes` | Gauge | 已存储数据量（字节） |
| `ceph_pool_max_available_bytes` | Gauge | 最大可用容量（字节） |
| `ceph_pool_used_bytes` | Gauge | 已用容量（字节） |
| `ceph_pool_percent_used` | Gauge | 使用率 (0.0-1.0) |
| `ceph_pool_objects_total` | Gauge | 对象数量 |
| `ceph_pool_read_bytes_sec` | Gauge | 读取吞吐量（字节/秒） |
| `ceph_pool_write_bytes_sec` | Gauge | 写入吞吐量（字节/秒） |
| `ceph_pool_read_ops_sec` | Gauge | 读取 IOPS |
| `ceph_pool_write_ops_sec` | Gauge | 写入 IOPS |

#### 4.4.4 Monitor 采集器 (Monitor Collector)

采集每个 Monitor 节点的状态指标。标签: `monitor`（Monitor 名称）

| 指标名称 | 类型 | 说明 |
|---------|------|------|
| `ceph_monitor_in_quorum` | Gauge | 是否在仲裁中 (1=是, 0=否) |
| `ceph_monitor_store_bytes` | Gauge | 数据库存储大小（字节） |
| `ceph_monitor_clock_skew_sec` | Gauge | 时钟偏移（秒） |
| `ceph_monitor_latency_sec` | Gauge | 响应延迟（秒） |

#### 4.4.5 健康状态采集器 (Health Collector)

采集集群整体健康状态。

| 指标名称 | 类型 | 说明 |
|---------|------|------|
| `ceph_health_status` | Gauge | 健康状态码 (0=OK, 1=WARN, 2=ERR) |
| `ceph_health_status_info` | Gauge | 健康状态信息（值恒为 1，标签: status） |
| `ceph_health_checks_total` | Gauge | 健康检查项总数 |
| `ceph_health_check` | Gauge | 健康检查项（值恒为 1，标签: name, severity） |

#### 4.4.6 MDS 采集器 (MDS Collector)

采集 CephFS 元数据服务器状态。

| 指标名称 | 类型 | 说明 |
|---------|------|------|
| `ceph_mds_active_total` | Gauge | Active 状态 MDS 数量 |
| `ceph_mds_standby_total` | Gauge | Standby 状态 MDS 数量 |
| `ceph_mds_daemon_status` | Gauge | MDS 守护进程状态（标签: name, state） |

#### 4.4.7 RGW 采集器 (RGW Collector)

采集对象网关状态。

| 指标名称 | 类型 | 说明 |
|---------|------|------|
| `ceph_rgw_total` | Gauge | RGW 守护进程总数 |
| `ceph_rgw_active_total` | Gauge | Active 状态 RGW 数量 |
| `ceph_rgw_daemon_status` | Gauge | RGW 守护进程状态（标签: name） |

### 4.5 HTTP 端点

| 端点 | 方法 | 说明 |
|------|------|------|
| `/metrics` | GET | Prometheus 指标端点 |
| `/health` | GET | 健康检查端点 |

### 4.6 常用操作

```bash
# 查看所有指标
curl http://localhost:9128/metrics

# 查看特定指标
curl http://localhost:9128/metrics | grep ceph_cluster_
curl http://localhost:9128/metrics | grep ceph_osd_
curl http://localhost:9128/metrics | grep ceph_pool_

# 健康检查
curl http://localhost:9128/health

# 查看容器日志
docker logs -f ceph-exporter

# 重启服务
docker restart ceph-exporter

# 进入容器调试
docker exec -it ceph-exporter /bin/bash

# 查看 Ceph 集群状态（在容器内）
docker exec ceph-exporter ceph -s
docker exec ceph-exporter ceph osd tree
docker exec ceph-exporter ceph df
```

---

## 5. Grafana 完整操作指南

### 5.1 Grafana 简介

Grafana 是一个开源的可视化和分析平台，用于展示 Prometheus 采集的 Ceph 集群指标。本项目提供 100% 中文化的 Grafana 界面，并预配置了 Ceph 集群监控仪表盘。

### 5.2 首次登录

1. 打开浏览器访问: http://localhost:3000
2. 输入默认账号: 用户名 `admin`，密码 `admin`
3. 首次登录会提示修改密码，可以选择"跳过"

### 5.3 界面布局说明

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
| 主页图标 | 主页 | 返回 Grafana 主页 |
| 搜索图标 | 搜索 | 搜索仪表盘、文件夹、标签 |
| 加号图标 | 创建 | 创建仪表盘、文件夹、导入 |
| 仪表盘图标 | 仪表盘 | 浏览所有仪表盘 |
| 探索图标 | 探索 | 临时查询和探索数据 |
| 告警图标 | 告警 | 查看和管理告警规则 |
| 配置图标 | 配置 | 数据源、插件、用户管理 |
| 用户图标 | 用户菜单 | 个人设置、退出登录 |

### 5.4 预配置仪表盘

本项目预配置了 3 个仪表盘，自动加载到 Grafana 的 "Ceph" 文件夹中：

| 仪表盘 | 文件 | 说明 |
|--------|------|------|
| Ceph 集群监控 | ceph-cluster.json | 集群核心指标监控 |
| Grafana 自身监控 | grafana-metrics-zh.json | Grafana 服务状态 |
| Prometheus 监控 | prometheus-stats-zh.json | Prometheus 服务状态 |

#### Ceph 集群仪表盘面板详解

**第一行：集群概览**

| 面板名称 | 显示内容 | 说明 |
|---------|---------|------|
| 集群健康状态 | HEALTH_OK / HEALTH_WARN / HEALTH_ERR | 绿色=正常，黄色=警告，红色=错误 |
| 集群总容量 | XX TB | 集群总存储容量 |
| 已用容量 | XX TB (XX%) | 已使用的存储空间和百分比 |
| 可用容量 | XX TB | 剩余可用存储空间 |

容量使用率阈值：75% 显示黄色警告，85% 显示红色告警。

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

1. **容量使用趋势图** - 显示过去 24 小时的容量变化，可以看到数据增长趋势
2. **IOPS 趋势图** - 显示读写 IOPS 的时间序列，可以识别性能峰值和低谷
3. **吞吐量趋势图** - 显示读写带宽的时间序列，可以分析 I/O 模式
4. **OSD 利用率分布** - 显示各个 OSD 的使用率，可以识别数据倾斜问题

### 5.5 常用操作

#### 5.5.1 调整时间范围

1. 点击右上角的时间选择器（默认显示"Last 6 hours"）
2. 选择预设时间范围：最近 5 分钟 / 15 分钟 / 30 分钟 / 1 小时 / 6 小时 / 12 小时 / 24 小时 / 7 天 / 30 天
3. 或自定义时间范围：点击"自定义时间范围" → 选择开始和结束时间 → 点击"应用"

#### 5.5.2 刷新仪表盘

1. 点击右上角的刷新按钮
2. 或设置自动刷新间隔：点击刷新按钮旁边的下拉箭头，选择间隔（5s, 10s, 30s, 1m, 5m, 15m, 30m, 1h）

#### 5.5.3 查看面板详情

1. 点击面板标题 → 选择"查看" (View)
2. 可以看到完整的图表、查询语句、数据表格、面板 JSON

#### 5.5.4 编辑面板

1. 点击面板标题 → 选择"编辑" (Edit)
2. 可以修改查询语句、可视化类型、面板选项、阈值和告警

#### 5.5.5 导出仪表盘

1. 点击右上角的"分享仪表盘"图标
2. 选择"导出"标签 → 点击"保存到文件"
3. 仪表盘将以 JSON 格式下载

#### 5.5.6 创建快照

1. 点击右上角的"分享仪表盘"图标
2. 选择"快照"标签
3. 设置快照名称和过期时间
4. 点击"本地快照" → 复制快照链接分享给他人

### 5.6 探索功能（Explore）

#### 访问探索页面

点击左侧边栏的"探索"图标，或在任意面板点击标题 → "探索"

#### 使用探索功能

1. **选择数据源**: 默认已选择 Prometheus
2. **输入查询**: 在查询框中输入 PromQL
3. **执行查询**: 点击"运行查询"按钮
4. **查看结果**: 图表视图（时间序列图）、表格视图（原始数据）

#### 常用 PromQL 查询示例

```promql
# 集群总容量
ceph_cluster_total_bytes

# 集群使用率（百分比）
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

### 5.7 告警配置

#### 查看告警规则

点击左侧边栏的"告警"图标 → 选择"告警规则"标签

#### 创建告警规则

1. 在仪表盘中选择要添加告警的面板
2. 点击面板标题 → "编辑" → 切换到"告警"标签
3. 点击"创建告警" → 配置评估间隔、条件表达式、阈值
4. 配置通知渠道 → 保存

#### 告警通知渠道

支持的通知方式：Email、Webhook、Slack、DingTalk (钉钉)、WeChat Work (企业微信)

配置方法：点击左侧边栏"配置" → "通知渠道" → "添加通知渠道" → 选择类型并填写配置 → 测试通知 → 保存

### 5.8 用户和权限管理

#### 创建新用户

点击左侧边栏"配置" → "用户" → "邀请"按钮 → 填写邮箱、用户名、角色

#### 用户角色说明

| 角色 | 权限 | 适用场景 |
|------|------|---------|
| **Admin** | 完全控制权限 | 系统管理员 |
| **Editor** | 可编辑仪表盘和数据源 | 开发人员、运维人员 |
| **Viewer** | 只读权限 | 普通用户、业务人员 |

### 5.9 数据源配置

本项目已预配置 Prometheus 数据源：

```yaml
# 预配置参数
名称: Prometheus
类型: prometheus
访问模式: proxy
URL: http://prometheus:9090
默认数据源: true
时间间隔: 15s
HTTP 方法: POST
```

如需添加其他数据源：点击左侧边栏"配置" → "数据源" → "添加数据源"

### 5.10 常见问题

**问题 1: 仪表盘显示"No Data"**

原因: Prometheus 未采集到数据 / 时间范围选择不当 / 查询语句错误

解决方法:
```bash
# 检查 ceph-exporter 是否正常运行
curl http://localhost:9128/metrics
# 检查 Prometheus targets 状态
# 访问 http://localhost:9090/targets
# 调整时间范围到"Last 5 minutes"
```

**问题 2: 仪表盘加载缓慢**

解决方法: 缩小时间范围 / 增加刷新间隔 / 优化查询语句

**问题 3: 无法登录**

解决方法:
```bash
# 检查服务状态
docker ps | grep grafana
# 重置管理员密码
docker exec -it grafana grafana-cli admin reset-admin-password newpassword
```

---

## 6. Prometheus 完整操作指南

### 6.1 Prometheus 简介

Prometheus 是一个开源的监控和告警系统，负责采集、存储和查询 Ceph 集群的时序指标数据。本项目配置了 15 秒的抓取间隔和 30 天的数据保留。

### 6.2 访问 Prometheus

打开浏览器访问: http://localhost:9090

### 6.3 界面布局说明

```
┌─────────────────────────────────────────────────────────────┐
│  [Prometheus Logo]  [Graph] [Alerts] [Status] [Help]        │
├─────────────────────────────────────────────────────────────┤
│  查询输入框: [输入 PromQL 查询]                [Execute]      │
├─────────────────────────────────────────────────────────────┤
│  [Graph] [Console]                                          │
│  ┌───────────────────────────────────────────────────────┐  │
│  │              图表/控制台显示区域                        │  │
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

### 6.4 Prometheus 配置详解

#### 全局配置

```yaml
global:
  scrape_interval: 15s          # 默认抓取间隔
  evaluation_interval: 15s      # 告警规则评估间隔
  external_labels:
    cluster: ceph-monitor       # 集群标签
    environment: production     # 环境标签
```

#### 抓取目标

| 目标 | 端点 | 抓取间隔 | 标签 |
|------|------|---------|------|
| prometheus | localhost:9090 | 15s | job="prometheus" |
| ceph-exporter | ceph-exporter:9128 | 15s | job="ceph-exporter", service="ceph-exporter", component="ceph" |
| alertmanager | alertmanager:9093 | 30s | job="alertmanager" |
| grafana | grafana:3000 | 30s | job="grafana" |

### 6.5 常用操作

#### 6.5.1 查询指标

在查询输入框中输入指标名称，点击"Execute"按钮，选择"Graph"或"Console"视图。

**基础查询示例**:

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

# 10. 查看 Monitor 时钟偏移
ceph_monitor_clock_skew_sec
```

#### 6.5.2 使用函数和聚合

**常用函数**:

```promql
# rate() - 计算速率（每秒变化量，适合缓慢变化的计数器）
rate(ceph_cluster_read_bytes_sec[5m])

# irate() - 瞬时速率（更敏感，适合快速变化的指标）
irate(ceph_cluster_read_bytes_sec[5m])

# increase() - 时间范围内的增长量
increase(ceph_cluster_read_bytes_sec[1h])

# delta() - 时间范围内的变化量（适合 Gauge 类型）
delta(ceph_cluster_used_bytes[1h])

# predict_linear() - 线性预测（预测未来值）
predict_linear(ceph_cluster_used_bytes[1h], 3600)

# sum() - 求和
sum(ceph_osd_used_bytes)

# avg() - 平均值
avg(ceph_osd_utilization)

# max() / min() - 最大值 / 最小值
max(ceph_osd_utilization)

# count() - 计数
count(ceph_osd_up == 1)

# topk() / bottomk() - 前 K 个最大/最小值
topk(5, ceph_osd_utilization)
```

**聚合示例**:

```promql
# 按存储池聚合读写 IOPS
sum by (pool) (rate(ceph_pool_read_ops_sec[5m]) + rate(ceph_pool_write_ops_sec[5m]))

# 按 OSD 聚合使用率
avg by (osd) (ceph_osd_utilization)

# 统计 Up 状态的 OSD 数量
count(ceph_osd_up == 1)
```

### 6.6 PromQL 查询语言详解

#### 选择器

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

#### 时间范围

```promql
# 过去 5 分钟的数据
ceph_cluster_read_bytes_sec[5m]

# 过去 1 小时的数据
ceph_cluster_read_bytes_sec[1h]

# 1 小时前的值（偏移量）
ceph_cluster_used_bytes offset 1h
```

#### 运算符

```promql
# 算术: +, -, *, /, %, ^
ceph_cluster_used_bytes / ceph_cluster_total_bytes

# 比较: ==, !=, >, <, >=, <=
ceph_osd_utilization > 80

# 逻辑: and, or, unless
ceph_osd_up == 1 and ceph_osd_in == 1
```

### 6.7 告警规则详解

本项目配置了 6 组共 16 条告警规则：

#### 集群健康告警 (ceph_cluster_health)

| 告警名称 | 条件 | 持续时间 | 严重程度 |
|---------|------|---------|---------|
| CephClusterWarning | health_status == 1 | 5 分钟 | warning |
| CephClusterError | health_status == 2 | 1 分钟 | critical |

#### OSD 告警 (ceph_osd_alerts)

| 告警名称 | 条件 | 持续时间 | 严重程度 |
|---------|------|---------|---------|
| CephOSDDown | osd_up == 0 | 3 分钟 | warning |
| CephOSDOut | osd_in == 0 | 5 分钟 | warning |
| CephOSDHighUtilization | utilization > 85% | 10 分钟 | warning |
| CephOSDHighLatency | apply_latency_ms > 500 | 5 分钟 | warning |
| CephMultipleOSDDown | > 10% OSDs down | 2 分钟 | critical |

#### 容量告警 (ceph_capacity_alerts)

| 告警名称 | 条件 | 持续时间 | 严重程度 |
|---------|------|---------|---------|
| CephClusterCapacityWarning | 使用率 > 75% | 10 分钟 | warning |
| CephClusterCapacityCritical | 使用率 > 85% | 5 分钟 | critical |
| CephClusterCapacityEmergency | 使用率 > 95% | 1 分钟 | emergency |

#### PG 告警 (ceph_pg_alerts)

| 告警名称 | 条件 | 持续时间 | 严重程度 |
|---------|------|---------|---------|
| CephPGNotClean | PGs 不在 active+clean | 15 分钟 | warning |
| CephNoPGs | PG 总数为 0 | 5 分钟 | critical |

#### Monitor 告警 (ceph_monitor_alerts)

| 告警名称 | 条件 | 持续时间 | 严重程度 |
|---------|------|---------|---------|
| CephMonitorOutOfQuorum | in_quorum == 0 | 2 分钟 | critical |
| CephMonitorClockSkew | clock_skew > 0.5s | 5 分钟 | warning |

#### 服务可用性告警 (ceph_exporter_alerts)

| 告警名称 | 条件 | 持续时间 | 严重程度 |
|---------|------|---------|---------|
| CephExporterDown | up == 0 | 2 分钟 | critical |
| CephExporterSlowScrape | scrape_duration > 10s | 5 分钟 | warning |

### 6.8 查看 Targets 状态

1. 点击顶部导航栏的"Status" → "Targets"
2. 状态说明：
   - **UP** (绿色): 正常采集
   - **DOWN** (红色): 采集失败
   - **UNKNOWN** (灰色): 未知状态

### 6.9 常见问题

**问题 1: Target 显示 DOWN**

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

**问题 2: 查询返回空结果**

1. 检查指标名称是否正确（在查询框输入 `ceph_` 会自动补全）
2. 调整时间范围
3. 使用"Console"视图查看原始数据

**问题 3: 告警规则未触发**

```bash
# 检查告警规则
curl http://localhost:9090/api/v1/rules
# 在 Graph 页面测试告警表达式
# 检查 for 持续时间设置
```

---

## 7. Alertmanager 完整操作指南

### 7.1 Alertmanager 简介

Alertmanager 负责接收 Prometheus 发送的告警，进行分组、去重、路由、抑制和通知。

### 7.2 访问 Alertmanager

打开浏览器访问: http://localhost:9093

### 7.3 界面布局说明

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

### 7.4 配置详解

#### 路由规则

```yaml
route:
  group_by: ['alertname', 'component']  # 分组依据
  group_wait: 30s                        # 初始等待时间（收到第一个告警后等待）
  group_interval: 5m                     # 同组告警间隔
  repeat_interval: 4h                    # 重复通知间隔
  receiver: 'default-webhook'            # 默认接收器

  routes:
    # 紧急告警立即发送
    - match:
        severity: critical|emergency
      receiver: 'critical-webhook'
      group_wait: 10s
      repeat_interval: 1h

    # OSD 告警按 OSD 分组
    - match_re:
        alertname: ^CephOSD.*
      receiver: 'default-webhook'
      group_by: ['alertname', 'osd']

    # 容量告警延长重复间隔
    - match_re:
        alertname: ^CephClusterCapacity.*
      receiver: 'default-webhook'
      repeat_interval: 2h
```

**参数说明**:
- `group_by`: 分组依据，相同标签值的告警会被合并
- `group_wait`: 收到第一个告警后等待的时间，用于收集同组的其他告警
- `group_interval`: 同一组的后续告警等待时间
- `repeat_interval`: 已发送的告警重复发送的间隔
- `receiver`: 告警接收器名称

#### 抑制规则

```yaml
inhibit_rules:
  # critical 告警抑制同组件的 warning 告警
  - source_match:
      severity: critical
    target_match:
      severity: warning
    equal: ['alertname', 'component']

  # emergency 告警抑制同组件的 critical 和 warning 告警
  - source_match:
      severity: emergency
    target_match_re:
      severity: critical|warning
    equal: ['alertname', 'component']
```

### 7.5 常用操作

#### 7.5.1 查看告警

1. 点击顶部导航栏的"Alerts"
2. 告警按严重程度分组显示：
   - **Critical** (红色): 严重告警，需要立即处理
   - **Warning** (黄色): 警告告警，需要关注
   - **Info** (蓝色): 信息告警

告警信息包括：告警名称、描述、严重程度、标签、触发时间、状态（Firing / Pending）

#### 7.5.2 静默告警

静默（Silence）可以临时屏蔽某些告警，在维护期间或已知问题期间避免告警骚扰。

**创建静默规则**:
1. 点击"Silences" → "New Silence"
2. 填写匹配条件（Matchers）、开始/结束时间、创建者、备注
3. 点击"Create"

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

#### 7.5.3 管理静默规则

- **Active**: 正在生效的静默
- **Pending**: 即将生效的静默
- **Expired**: 已过期的静默

编辑：点击"Edit" → 修改配置 → "Update"
删除：点击"Expire" → 静默规则立即失效

#### 7.5.4 查看状态

点击"Status"可以看到：Cluster Status、Uptime、Config、Version Info

### 7.6 告警通知配置

#### 配置 Webhook 通知

```yaml
receivers:
  - name: 'webhook-receiver'
    webhook_configs:
      - url: 'http://your-webhook-url/alert'
        send_resolved: true
```

#### 配置邮件通知

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

#### 配置企业微信通知

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

### 7.7 常见问题

**问题 1: 收不到告警通知**

```bash
# 检查 Alertmanager 配置
docker exec alertmanager amtool check-config /etc/alertmanager/alertmanager.yml
# 查看日志
docker logs alertmanager
```

**问题 2: 告警重复发送**

增加 `repeat_interval` 时间，优化告警规则的触发条件，调整 `group_by` 分组策略。

**问题 3: 静默规则不生效**

检查匹配条件是否正确，确认时间范围包含当前时间，在"Alerts"页面查看告警的实际标签。

---

## 8. ELK Stack 完整操作指南

### 8.1 ELK Stack 简介

ELK Stack 由 Elasticsearch、Logstash 和 Kibana 三个组件组成，配合 Filebeat 提供完整的日志收集、存储、搜索和可视化解决方案。

| 组件 | 版本 | 端口 | 功能 |
|------|------|------|------|
| Elasticsearch | 7.17.0 | 9200 (HTTP), 9300 (TCP) | 日志存储和搜索引擎 |
| Logstash | 7.17.0 | 5044 (Beats), 5000 (TCP) | 日志处理管道 |
| Kibana | 7.17.0 | 5601 | 日志可视化界面 |
| Filebeat | 7.17.0 | - | 日志收集代理 (sidecar) |

### 8.2 数据流架构

```
方案1 (直接推送):
  ceph-exporter → TCP/UDP 5000 → Logstash → Elasticsearch → Kibana

方案2 (容器日志收集，推荐):
  ceph-exporter → stdout → Docker JSON log → Filebeat sidecar → Logstash 5044 → Elasticsearch → Kibana

方案3 (文件日志):
  ceph-exporter → 日志文件 → Filebeat → Logstash 5044 → Elasticsearch → Kibana
```

### 8.3 Logstash 管道配置

Logstash 管道配置文件 (`configs/logstash.conf`) 定义了日志的输入、过滤和输出：

**输入**:
- Beats 输入 (端口 5044): 接收 Filebeat 发送的日志，使用 JSON codec
- TCP 输入 (端口 5000): 接收 ceph-exporter 直接推送的日志，使用 json_lines codec

**过滤处理**:
1. JSON 解析日志内容
2. 时间戳处理（ISO8601 格式）
3. 服务标识添加
4. 组件信息提取
5. 追踪 ID 关联（与 Jaeger 联动）
6. 字段清理

**输出**:
- Elasticsearch: 索引名称 `ceph-exporter-YYYY.MM.dd`（按日期分割）

### 8.4 Filebeat 配置

Filebeat 以 sidecar 模式运行，采集 ceph-exporter 容器的标准输出日志：

```yaml
# 输入配置
filebeat.inputs:
  - type: container
    paths: /var/lib/docker/containers/*/*.log
    json.keys_under_root: true
    json.add_error_key: true
    fields:
      service: ceph-exporter

# 处理器
processors:
  - add_docker_metadata          # 从 Docker socket 获取容器元数据
  - drop_event                   # 仅保留 ceph-exporter 容器日志

# 输出
output.logstash:
  hosts: ["logstash:5044"]
  loadbalance: true
  compression_level: 3
  timeout: 30s
  bulk_max_size: 2048
```

### 8.5 访问 Kibana

打开浏览器访问: http://localhost:5601

### 8.6 Kibana 界面操作

#### 8.6.1 首次访问 - 创建索引模式

1. 点击左侧菜单"Management" → "Index Patterns"（或 "Stack Management" → "索引模式"）
2. 点击"Create index pattern"
3. 输入索引模式：`ceph-exporter-*`
4. 选择时间字段：`@timestamp`
5. 点击"Create index pattern"

#### 8.6.2 查看日志 (Discover)

1. 点击左侧菜单"Discover"
2. 选择索引模式：`ceph-exporter-*`
3. 调整时间范围（右上角）
4. 可以看到所有日志记录

#### 8.6.3 搜索日志

**基础搜索**:
```
# 搜索包含"error"的日志
error

# 搜索特定级别的日志
level:error

# 搜索特定组件的日志
component:collector
```

**高级搜索（KQL - Kibana Query Language）**:
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

# 存在性查询（查找包含追踪 ID 的日志）
_exists_:trace_id
```

#### 8.6.4 创建可视化

1. 点击左侧菜单"Visualize" → "Create visualization"
2. 选择可视化类型：
   - Line (折线图): 适合展示日志量趋势
   - Bar (柱状图): 适合展示日志级别分布
   - Pie (饼图): 适合展示组件日志占比
   - Data Table (数据表): 适合展示详细日志统计
   - Metric (指标): 适合展示错误日志总数
3. 选择数据源 → 配置聚合和指标 → 保存

#### 8.6.5 创建仪表盘

1. 点击左侧菜单"Dashboard" → "Create dashboard"
2. 点击"Add"添加已保存的可视化
3. 调整面板大小和位置
4. 保存仪表盘

### 8.7 日志级别说明

| 级别 | 英文 | 说明 | 使用场景 |
|------|------|------|---------|
| trace | TRACE | 最详细的日志 | 开发调试 |
| debug | DEBUG | 调试信息 | 问题排查 |
| info | INFO | 一般信息 | 生产环境（推荐） |
| warn | WARN | 警告信息 | 潜在问题 |
| error | ERROR | 错误信息 | 错误记录 |
| fatal | FATAL | 致命错误 | 严重错误 |
| panic | PANIC | 恐慌错误 | 程序崩溃 |

### 8.8 常见问题

**问题 1: Kibana 无法连接 Elasticsearch**

```bash
# 检查 Elasticsearch 状态
curl http://localhost:9200
# 检查 Kibana 日志
docker logs kibana
# 重启服务
docker restart elasticsearch kibana
```

**问题 2: 没有日志数据**

```bash
# 1. 检查 ceph-exporter 日志配置
grep enable_elk configs/ceph-exporter.yaml
# 2. 检查 Logstash 状态
docker logs logstash
# 3. 检查 Elasticsearch 索引
curl http://localhost:9200/_cat/indices?v
# 4. 检查 Filebeat 状态（容器日志收集模式）
docker logs filebeat-sidecar
```

**问题 3: Logstash OOM（内存不足）**

```bash
# 检查 Logstash 内存使用
docker stats logstash
# 调整 JVM 堆大小（在 docker-compose 中设置）
# LS_JAVA_OPTS: "-Xmx512m -Xms512m"
```

**问题 4: Elasticsearch 启动失败**

```bash
# 检查系统参数
sysctl vm.max_map_count
# 如果小于 262144，设置：
sudo sysctl -w vm.max_map_count=262144
# 检查目录权限
sudo chown -R 1000:1000 data/elasticsearch
```

---

## 9. Jaeger 分布式追踪指南

### 9.1 Jaeger 简介

Jaeger 是一个开源的分布式追踪系统，用于监控和排查微服务架构中的性能问题。本项目使用 OpenTelemetry SDK 将追踪数据发送到 Jaeger。

### 9.2 架构

```
HTTP Request → ceph-exporter (OpenTelemetry SDK) → OTLP HTTP (4318) → Jaeger Collector → Jaeger UI (16686)
```

Jaeger 使用 all-in-one 模式部署，包含 Collector、Query 和 UI 组件。

### 9.3 端口说明

| 端口 | 协议 | 用途 |
|------|------|------|
| 16686 | HTTP | Jaeger UI 界面 |
| 4318 | HTTP | OTLP HTTP 接收端点 |
| 14268 | HTTP | Jaeger HTTP Thrift 接收端点 |
| 6831 | UDP | Jaeger Compact Thrift 接收端点 |

### 9.4 访问 Jaeger UI

打开浏览器访问: http://localhost:16686

### 9.5 界面操作

#### 9.5.1 搜索追踪

1. 在左侧面板选择"Service": `ceph-exporter`
2. 选择"Operation"（可选）:
   - `HTTP GET /metrics` - 指标采集请求
   - `HTTP GET /health` - 健康检查请求
3. 设置时间范围（Lookback）
4. 可选设置：
   - **Min Duration**: 最小持续时间（过滤慢请求）
   - **Max Duration**: 最大持续时间
   - **Limit Results**: 结果数量限制
   - **Tags**: 按标签过滤（如 `http.status_code=200`）
5. 点击"Find Traces"

#### 9.5.2 查看追踪详情

点击搜索结果中的某个追踪，可以看到：

- **追踪 ID**: 唯一标识一次请求
- **开始时间**: 请求开始的时间
- **持续时间**: 请求总耗时
- **Span 数量**: 追踪中包含的 Span 数量
- **服务数量**: 涉及的服务数量

#### 9.5.3 时间线视图

时间线视图展示了请求的完整调用链：

```
[HTTP GET /metrics] ─────────────────────────── 150ms
  ├─ [ceph.connect] ──────── 10ms
  ├─ [cluster.collect] ───────────── 50ms
  ├─ [osd.collect] ──────────────── 40ms
  ├─ [pool.collect] ─────────── 30ms
  ├─ [monitor.collect] ──── 10ms
  └─ [health.collect] ──── 10ms
```

每个 Span 显示：
- 操作名称
- 持续时间
- 标签（Tags）: 如 HTTP 状态码、错误信息
- 日志（Logs）: 事件记录

#### 9.5.4 分析性能

- 查看各个 Span 的耗时，识别性能瓶颈
- 查看错误和异常（红色标记的 Span）
- 比较不同追踪的耗时分布
- 使用"Compare"功能对比两个追踪

#### 9.5.5 依赖关系图

点击顶部导航栏的"Dependencies"可以查看服务间的依赖关系图。

### 9.6 启用/禁用追踪

#### 使用脚本启用

```bash
cd ceph-exporter/deployments
./scripts/enable-jaeger-tracing.sh
```

脚本会自动执行：
1. 修改配置文件启用追踪 (`tracer.enabled: true`)
2. 启动 Jaeger 容器
3. 重新构建并启动 ceph-exporter
4. 等待服务就绪并验证

#### 手动配置

编辑 `ceph-exporter/configs/ceph-exporter.yaml`:

```yaml
tracer:
  enabled: true                 # true=启用, false=禁用
  jaeger_url: "jaeger:4318"     # Jaeger OTLP HTTP 端点
  service_name: "ceph-exporter" # 服务名称
  sample_rate: 1.0              # 采样率 (0.0-1.0)
```

采样率说明：
- `1.0`: 100% 采样（开发/测试环境推荐）
- `0.1`: 10% 采样（生产环境推荐，减少性能开销）
- `0.01`: 1% 采样（高流量生产环境）

### 9.7 与 ELK 联动

追踪 ID 会自动注入到日志中，可以在 Kibana 中通过追踪 ID 关联日志：

```
# 在 Kibana 中搜索特定追踪的日志
trace_id: "abc123def456"

# 搜索包含追踪 ID 的所有日志
_exists_:trace_id
```

### 9.8 常见问题

**问题 1: Jaeger UI 无追踪数据**

```bash
# 检查 Jaeger 状态
docker ps | grep jaeger
# 检查 ceph-exporter 追踪配置
grep -A4 "tracer:" configs/ceph-exporter.yaml
# 生成测试数据
curl http://localhost:9128/metrics
# 检查 Jaeger 日志
docker logs jaeger
```

**问题 2: 追踪数据不完整**

检查采样率设置，确保 `sample_rate` 不为 0。

---

## 10. Ceph Dashboard 操作指南

### 10.1 访问 Ceph Dashboard

打开浏览器访问: http://localhost:8080

注意：Ceph Dashboard 由 ceph-demo 容器提供，仅在包含 ceph-demo 的部署模式中可用。

### 10.2 主要功能

- **集群状态概览**: 查看集群健康状态、容量使用情况
- **OSD 管理**: 查看 OSD 状态、标记 OSD in/out/up/down
- **存储池管理**: 创建、删除、修改存储池
- **Monitor 状态**: 查看 Monitor 仲裁状态
- **性能监控**: 查看实时性能指标

### 10.3 Ceph RGW (S3 对象存储)

Ceph RGW 提供 S3 兼容的对象存储接口，端口 5000。

```bash
# 使用 curl 测试 RGW
curl http://localhost:5000

# 使用 s3cmd 或 aws-cli 操作（需要配置访问密钥）
```

---

## 11. 部署脚本工具集

所有脚本位于 `ceph-exporter/deployments/scripts/` 目录。

### 11.1 deploy.sh - 主部署脚本

自动检查环境、配置镜像加速、分阶段部署所有组件。

```bash
cd ceph-exporter/deployments

# 完整栈部署（包含所有服务）
./scripts/deploy.sh full

# 最小栈部署（仅监控组件）
./scripts/deploy.sh minimal

# 集成测试环境
./scripts/deploy.sh integration

# 仅 Ceph Demo
./scripts/deploy.sh ceph-demo

# 初始化数据目录
./scripts/deploy.sh init

# 查看服务状态
./scripts/deploy.sh status

# 验证部署
./scripts/deploy.sh verify

# 查看日志
./scripts/deploy.sh logs [service-name]

# 诊断问题
./scripts/deploy.sh diagnose

# 修复部署（需要 sudo）
sudo ./scripts/deploy.sh fix

# 停止服务
./scripts/deploy.sh stop

# 清理数据
./scripts/deploy.sh clean
```

部署脚本自动执行的检查：
- 操作系统版本检查（Ubuntu 20.04）
- Docker 和 Docker Compose 安装检查
- 系统资源检查（内存 >= 2GB、CPU >= 2 核、磁盘 >= 20GB）
- 防火墙规则配置（自动开放所需端口）
- Docker 镜像加速器配置（国内源：USTC、163、腾讯云）

### 11.2 diagnose.sh - 统一诊断脚本

全面诊断所有服务状态，收集日志和配置信息。

```bash
# 完整诊断（默认）
sudo ./scripts/diagnose.sh

# 诊断特定服务
sudo ./scripts/diagnose.sh ceph-exporter
sudo ./scripts/diagnose.sh ceph-demo
sudo ./scripts/diagnose.sh prometheus
sudo ./scripts/diagnose.sh grafana
sudo ./scripts/diagnose.sh kibana
sudo ./scripts/diagnose.sh elasticsearch
sudo ./scripts/diagnose.sh jaeger
```

诊断内容包括：
1. 容器状态总览
2. 各服务详细诊断（状态、重启次数、日志、端口连通性）
3. 网络连接检查
4. 配置文件验证

### 11.3 verify-deployment.sh - 部署验证脚本

验证所有服务是否正常运行并可访问。

```bash
./scripts/verify-deployment.sh
```

验证内容：
- 中文配置文件是否存在（Grafana Dashboard、Prometheus、Alertmanager）
- Docker Compose 配置文件是否存在
- 各服务 HTTP 端点是否可访问

### 11.4 switch-logging-mode.sh - 日志方案切换脚本

快速切换日志输出模式。

```bash
# 查看当前配置
./scripts/switch-logging-mode.sh show

# 切换模式
./scripts/switch-logging-mode.sh direct      # 直接推送到 Logstash (TCP)
./scripts/switch-logging-mode.sh direct-udp  # 直接推送到 Logstash (UDP)
./scripts/switch-logging-mode.sh container   # 容器日志收集（推荐）
./scripts/switch-logging-mode.sh file        # 文件日志 + Filebeat
./scripts/switch-logging-mode.sh dev         # 开发模式 (stdout + text)
```

### 11.5 enable-jaeger-tracing.sh - Jaeger 追踪启用脚本

快速启用 Jaeger 分布式追踪。

```bash
./scripts/enable-jaeger-tracing.sh
```

执行步骤：
1. 修改配置文件启用追踪
2. 启动 Jaeger 容器
3. 重新构建并启动 ceph-exporter
4. 等待服务就绪并验证

### 11.6 fix-deployment.sh - 部署修复脚本

修复常见的部署问题（权限、配置路径等）。需要 root 权限。

```bash
sudo ./scripts/fix-deployment.sh
```

修复内容：
- Prometheus 数据目录权限 (65534:65534)
- Grafana 数据目录权限 (472:472)
- Elasticsearch 数据目录权限 (1000:1000)
- Ceph keyring 文件权限
- configs 目录软链接

### 11.7 clean-volumes.sh - 数据清理脚本

彻底清理所有持久化数据。

```bash
./scripts/clean-volumes.sh
```

执行步骤：
1. 停止所有服务
2. 删除 `data/` 目录中的所有数据
3. 重新初始化数据目录
4. 重新启动 Ceph Demo

**警告**: 此操作会删除所有持久化数据，不可恢复！

### 11.8 其他脚本

| 脚本 | 功能 | 用法 |
|------|------|------|
| diagnose-logstash.sh | 诊断 Logstash 连接问题 | `./scripts/diagnose-logstash.sh` |
| diagnose-elk-full.sh | 完整 ELK 栈诊断 | `./scripts/diagnose-elk-full.sh` |
| diagnose-and-test.sh | 诊断并运行集成测试 | `./scripts/diagnose-and-test.sh` |
| list-temp-files.sh | 列出临时文件 | `./scripts/list-temp-files.sh` |
| cleanup-temp-files.sh | 清理临时文件 | `./scripts/cleanup-temp-files.sh` |

---

## 12. 界面术语对照表

### 12.1 Grafana 术语

| 英文术语 | 中文翻译 | 说明 |
|---------|---------|------|
| Dashboard | 仪表盘 | 包含多个面板的可视化页面 |
| Panel | 面板 | 单个图表或可视化组件 |
| Data Source | 数据源 | 数据来源（如 Prometheus） |
| Query | 查询 | 数据查询语句 |
| Time Range | 时间范围 | 数据的时间跨度 |
| Refresh | 刷新 | 更新数据 |
| Variables | 变量 | 动态参数，用于仪表盘模板化 |
| Annotations | 注释 | 事件标记，在图表上显示事件 |
| Alert | 告警 | 告警规则 |
| Threshold | 阈值 | 告警触发条件值 |
| Legend | 图例 | 数据系列说明 |
| Tooltip | 提示框 | 鼠标悬停显示的信息 |
| Axis | 坐标轴 | X轴（时间）/Y轴（数值） |
| Series | 系列 | 数据序列 |
| Aggregation | 聚合 | 数据聚合方式 |
| Provisioning | 自动配置 | 通过文件自动配置数据源和仪表盘 |
| Folder | 文件夹 | 仪表盘分组 |
| Snapshot | 快照 | 仪表盘的静态副本 |
| Explore | 探索 | 临时查询和数据探索功能 |
| Stat | 统计 | 单值统计面板类型 |
| Gauge | 仪表 | 仪表盘面板类型 |
| Graph | 图表 | 时间序列图表面板类型 |
| Table | 表格 | 数据表格面板类型 |

### 12.2 Prometheus 术语

| 英文术语 | 中文翻译 | 说明 |
|---------|---------|------|
| Metric | 指标 | 监控数据项 |
| Label | 标签 | 指标的维度标识 |
| Target | 目标 | 被监控的服务端点 |
| Scrape | 抓取 | 从目标采集指标数据 |
| Scrape Interval | 抓取间隔 | 采集频率（本项目默认 15s） |
| Evaluation Interval | 评估间隔 | 告警规则评估频率 |
| Alert Rule | 告警规则 | 告警触发条件定义 |
| Recording Rule | 记录规则 | 预计算规则，将复杂查询结果存储为新指标 |
| Instant Query | 即时查询 | 查询当前时刻的值 |
| Range Query | 范围查询 | 查询时间范围内的值 |
| Selector | 选择器 | 标签匹配条件 |
| Aggregation | 聚合 | 数据聚合操作（sum, avg, max 等） |
| Function | 函数 | PromQL 内置函数（rate, irate 等） |
| Operator | 运算符 | 算术/比较/逻辑运算符 |
| Time Series | 时间序列 | 带时间戳的数据序列 |
| Counter | 计数器 | 只增不减的指标类型 |
| Gauge | 仪表 | 可增可减的指标类型 |
| Histogram | 直方图 | 数据分布统计指标类型 |
| Summary | 摘要 | 分位数统计指标类型 |
| External Labels | 外部标签 | 附加到所有指标的全局标签 |

### 12.3 Alertmanager 术语

| 英文术语 | 中文翻译 | 说明 |
|---------|---------|------|
| Alert | 告警 | 告警事件 |
| Silence | 静默 | 临时屏蔽告警 |
| Inhibition | 抑制 | 高优先级告警自动抑制低优先级告警 |
| Route | 路由 | 告警路由规则，决定告警发送到哪个接收器 |
| Receiver | 接收者 | 告警通知目标（Webhook、邮件等） |
| Group | 分组 | 将相关告警合并为一组发送 |
| Group Wait | 分组等待 | 收到第一个告警后等待的时间 |
| Group Interval | 分组间隔 | 同一组后续告警的等待时间 |
| Repeat Interval | 重复间隔 | 已发送告警的重复发送间隔 |
| Firing | 触发中 | 告警正在触发 |
| Pending | 等待中 | 告警条件满足但未达到持续时间 |
| Resolved | 已解决 | 告警条件不再满足，已恢复 |
| Severity | 严重程度 | 告警级别（warning, critical, emergency） |
| Matcher | 匹配器 | 标签匹配条件 |
| Notification | 通知 | 告警通知消息 |
| Send Resolved | 发送恢复通知 | 告警恢复时是否发送通知 |

### 12.4 Ceph 术语

| 英文术语 | 中文翻译 | 说明 |
|---------|---------|------|
| Cluster | 集群 | Ceph 存储集群 |
| Pool | 存储池 | 数据存储池，逻辑分区 |
| OSD (Object Storage Daemon) | 对象存储守护进程 | 管理物理磁盘的存储节点 |
| Monitor (MON) | 监视器 | 维护集群状态映射的节点 |
| MDS (Metadata Server) | 元数据服务器 | CephFS 文件系统元数据服务 |
| RGW (RADOS Gateway) | 对象网关 | S3/Swift 兼容的对象存储网关 |
| PG (Placement Group) | 归置组 | 数据分布的基本单元 |
| RADOS | 可靠自主分布式对象存储 | Ceph 底层存储系统 |
| CRUSH | 可扩展哈希下的受控复制 | 数据分布算法 |
| Health | 健康状态 | 集群健康状态 (OK/WARN/ERR) |
| Capacity | 容量 | 存储容量 |
| Utilization | 利用率 | 使用率百分比 |
| Latency | 延迟 | 响应延迟（毫秒） |
| Throughput | 吞吐量 | 数据传输速率（字节/秒） |
| IOPS | 每秒IO操作数 | 性能指标 |
| Quorum | 仲裁 | Monitor 节点的多数派共识 |
| Keyring | 密钥环 | Ceph 认证密钥文件 |
| OSD Up | OSD 运行中 | OSD 进程正在运行 |
| OSD In | OSD 在集群中 | OSD 参与数据分布 |
| OSD Down | OSD 宕机 | OSD 进程未运行 |
| OSD Out | OSD 踢出集群 | OSD 不参与数据分布 |
| Active+Clean | 活跃+干净 | PG 的正常健康状态 |
| Clock Skew | 时钟偏移 | Monitor 节点间的时钟差异 |

### 12.5 ELK 术语

| 英文术语 | 中文翻译 | 说明 |
|---------|---------|------|
| Index | 索引 | 数据存储单元（类似数据库表） |
| Document | 文档 | 单条日志记录 |
| Field | 字段 | 文档属性 |
| Mapping | 映射 | 字段类型定义 |
| Query | 查询 | 搜索语句 |
| Filter | 过滤器 | 数据过滤条件 |
| Aggregation | 聚合 | 数据统计分析 |
| Visualization | 可视化 | 图表展示 |
| Dashboard | 仪表盘 | 可视化集合 |
| Discover | 发现 | 日志浏览页面 |
| KQL (Kibana Query Language) | Kibana 查询语言 | Kibana 搜索语法 |
| Index Pattern | 索引模式 | 匹配多个索引的通配符模式 |
| Pipeline | 管道 | Logstash 数据处理管道 |
| Beats | 轻量级数据采集器 | Filebeat 等数据采集工具 |
| Sidecar | 边车 | 与主容器一起运行的辅助容器 |
| Codec | 编解码器 | 数据格式编解码（json, json_lines 等） |

### 12.6 Jaeger 术语

| 英文术语 | 中文翻译 | 说明 |
|---------|---------|------|
| Trace | 追踪 | 一次完整请求的调用链 |
| Span | 跨度 | 追踪中的一个操作单元 |
| Service | 服务 | 产生追踪数据的服务 |
| Operation | 操作 | Span 的操作名称 |
| Duration | 持续时间 | 操作耗时 |
| Tags | 标签 | Span 的键值对属性 |
| Logs | 日志 | Span 中的事件记录 |
| Trace ID | 追踪 ID | 追踪的唯一标识符 |
| Span ID | 跨度 ID | Span 的唯一标识符 |
| Parent Span | 父跨度 | 当前 Span 的调用者 |
| Child Span | 子跨度 | 当前 Span 调用的操作 |
| Sampling | 采样 | 控制追踪数据的采集比例 |
| Sample Rate | 采样率 | 采样比例 (0.0-1.0) |
| Collector | 收集器 | 接收追踪数据的组件 |
| OTLP | OpenTelemetry 协议 | 追踪数据传输协议 |
| Dependencies | 依赖关系 | 服务间的调用依赖图 |

### 12.7 Docker 术语

| 英文术语 | 中文翻译 | 说明 |
|---------|---------|------|
| Container | 容器 | 运行中的 Docker 实例 |
| Image | 镜像 | 容器的模板 |
| Volume | 卷 | 持久化数据存储 |
| Bind Mount | 绑定挂载 | 将宿主机目录挂载到容器 |
| Network | 网络 | 容器间通信网络 |
| Compose | 编排 | 多容器应用编排工具 |
| Service | 服务 | Compose 中定义的容器 |
| Registry | 镜像仓库 | Docker 镜像存储服务 |
| Registry Mirror | 镜像加速器 | 镜像下载加速代理 |
| Resource Limit | 资源限制 | 容器的 CPU/内存限制 |

---

## 13. 常用操作速查

### 13.1 部署操作

```bash
cd ceph-exporter/deployments

# 完整部署
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

### 13.2 服务管理

```bash
# 重启单个服务
docker restart ceph-exporter
docker restart prometheus
docker restart grafana
docker restart alertmanager
docker restart elasticsearch
docker restart logstash
docker restart kibana
docker restart jaeger

# 查看服务日志（实时跟踪）
docker logs -f ceph-exporter
docker logs -f prometheus
docker logs -f grafana
docker logs -f logstash

# 查看最近 100 行日志
docker logs --tail 100 ceph-exporter

# 进入容器
docker exec -it ceph-exporter /bin/bash
docker exec -it prometheus /bin/sh

# 查看资源使用
docker stats

# 查看所有容器状态
docker ps -a --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
```

### 13.3 数据备份

```bash
cd ceph-exporter/deployments

# 备份 Prometheus 数据
tar -czf prometheus-backup-$(date +%Y%m%d).tar.gz data/prometheus/

# 备份 Grafana 数据
tar -czf grafana-backup-$(date +%Y%m%d).tar.gz data/grafana/

# 备份 Elasticsearch 数据
tar -czf elasticsearch-backup-$(date +%Y%m%d).tar.gz data/elasticsearch/

# 备份所有数据
tar -czf ceph-exporter-backup-$(date +%Y%m%d).tar.gz data/

# 备份配置文件
tar -czf configs-backup-$(date +%Y%m%d).tar.gz ../configs/ prometheus/ alertmanager/ grafana/
```

### 13.4 常用 Prometheus 查询

```promql
# 集群使用率
(ceph_cluster_used_bytes / ceph_cluster_total_bytes) * 100

# 集群 IOPS
rate(ceph_cluster_read_ops_sec[5m]) + rate(ceph_cluster_write_ops_sec[5m])

# OSD 状态统计
count(ceph_osd_up == 1)

# 存储池使用率 Top 5
topk(5, ceph_pool_percent_used)

# OSD 延迟 Top 5
topk(5, ceph_osd_apply_latency_ms)

# 集群容量预测（1 小时后）
predict_linear(ceph_cluster_used_bytes[1h], 3600)

# 健康状态
ceph_health_status
```

### 13.5 常用 Kibana 查询

```
# 错误日志
level:error

# 特定组件日志
component:collector

# 包含追踪 ID 的日志
_exists_:trace_id

# 响应时间超过 100ms
response_time > 100

# 特定时间范围的错误
level:error AND @timestamp:[2026-03-22 TO 2026-03-23]

# 排除 info 级别
NOT level:info
```

### 13.6 日志方案切换

```bash
cd ceph-exporter/deployments

# 查看当前配置
./scripts/switch-logging-mode.sh show

# 切换到容器日志收集（推荐）
./scripts/switch-logging-mode.sh container

# 切换到直接推送 (TCP)
./scripts/switch-logging-mode.sh direct

# 切换到开发模式
./scripts/switch-logging-mode.sh dev
```

---

## 14. 常见问题与解决方案

### 14.1 部署问题

#### 问题：Docker 镜像拉取失败

**症状**:
```
Error response from daemon: Get https://registry-1.docker.io/v2/: net/http: TLS handshake timeout
```

**解决方法**:
1. 配置 Docker 镜像加速器（deploy.sh 会自动配置）
2. 手动配置：
```bash
sudo tee /etc/docker/daemon.json > /dev/null <<EOF
{
  "registry-mirrors": [
    "https://docker.mirrors.ustc.edu.cn",
    "https://hub-mirror.c.163.com",
    "https://mirror.ccs.tencentyun.com"
  ]
}
EOF
sudo systemctl daemon-reload
sudo systemctl restart docker
```
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
sudo ./scripts/fix-deployment.sh

# 或手动修复权限
sudo chown -R 65534:65534 data/prometheus    # Prometheus (nobody 用户)
sudo chown -R 472:472 data/grafana           # Grafana
sudo chown -R 1000:1000 data/elasticsearch   # Elasticsearch
```

#### 问题：Elasticsearch 启动失败 (vm.max_map_count)

**症状**:
```
max virtual memory areas vm.max_map_count [65530] is too low, increase to at least [262144]
```

**解决方法**:
```bash
# 临时设置
sudo sysctl -w vm.max_map_count=262144

# 永久设置
echo "vm.max_map_count=262144" | sudo tee -a /etc/sysctl.conf
sudo sysctl -p
```

#### 问题：Ceph keyring 权限错误

**症状**: ceph-exporter 无法连接 Ceph 集群

**解决方法**:
```bash
cd ceph-exporter/deployments
chmod 644 data/ceph-demo/config/ceph.client.admin.keyring
chmod 644 data/ceph-demo/config/ceph.mon.keyring
docker restart ceph-exporter
```

### 14.2 监控问题

#### 问题：Grafana 显示 "No Data"

**原因**: Prometheus 未采集到数据 / ceph-exporter 未正常运行 / 时间范围选择不当

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

```bash
# 1. 检查告警规则
curl http://localhost:9090/api/v1/rules
# 2. 在 Prometheus UI 中测试告警表达式
# 3. 检查 Alertmanager
curl http://localhost:9093/api/v1/status
# 4. 查看日志
docker logs alertmanager
```

#### 问题：Prometheus 内存占用过高

**解决方法**:
1. 增加抓取间隔：修改 `prometheus.yml` 中 `scrape_interval: 30s`
2. 减少数据保留时间：启动参数 `--storage.tsdb.retention.time=15d`
3. 增加内存限制：修改 docker-compose 中 `memory: 2g`

### 14.3 日志问题

#### 问题：Kibana 无日志数据

```bash
# 1. 检查配置
grep enable_elk configs/ceph-exporter.yaml
# 2. 检查 Logstash
docker logs logstash
# 3. 检查 Elasticsearch 索引
curl http://localhost:9200/_cat/indices?v
# 4. 检查 Filebeat（容器日志收集模式）
docker logs filebeat-sidecar
# 5. 重启服务
docker restart logstash elasticsearch kibana
```

#### 问题：Logstash OOM

```bash
# 检查内存使用
docker stats logstash
# 重启 Logstash
docker restart logstash
# 如果持续 OOM，增加内存限制或减少日志量
```

### 14.4 追踪问题

#### 问题：Jaeger UI 无追踪数据

```bash
# 1. 检查追踪配置
grep -A4 "tracer:" configs/ceph-exporter.yaml
# 2. 确保 enabled: true
# 3. 生成测试数据
curl http://localhost:9128/metrics
# 4. 检查 Jaeger
docker logs jaeger
# 5. 重启
docker restart ceph-exporter jaeger
```

### 14.5 性能问题

#### 问题：Grafana 仪表盘加载缓慢

**解决方法**:
1. 缩小时间范围（从 7 天改为 1 小时）
2. 增加刷新间隔（从 5s 改为 30s）
3. 优化查询语句（避免使用 `.*` 正则）
4. 使用记录规则预计算复杂查询

#### 问题：整体服务响应慢

```bash
# 检查资源使用
docker stats
# 检查磁盘空间
df -h
# 检查系统负载
uptime
# 清理旧数据
curl -X DELETE "http://localhost:9200/ceph-exporter-2026.03.01"
```

---

## 15. 最佳实践

### 15.1 监控最佳实践

1. **合理设置告警阈值**: 根据实际业务需求调整，避免告警过于敏感或迟钝，定期 review 和优化
2. **使用分层告警**: Warning（需要关注）→ Critical（需要立即处理）→ Emergency（严重影响业务）
3. **配置告警通知渠道**: 工作时间用邮件+即时通讯，非工作时间用电话+短信，严重告警多渠道通知
4. **定期检查监控系统**: 验证告警规则有效性，检查数据采集正常性，清理过期静默规则

### 15.2 性能优化

1. **Prometheus 优化**: 合理设置数据保留时间（默认 30 天），使用记录规则预计算，控制指标数量和标签基数
2. **Grafana 优化**: 使用变量减少重复查询，设置合理的刷新间隔，避免过于复杂的查询
3. **ELK 优化**: 定期清理旧索引，使用索引生命周期管理，控制日志级别（生产环境用 info）
4. **Jaeger 优化**: 生产环境降低采样率（0.1），定期清理旧追踪数据

### 15.3 安全建议

1. **修改默认密码**: Grafana 默认 admin/admin，首次登录后立即修改
2. **启用 HTTPS**: 配置 TLS 证书，ceph-exporter 支持 `tls_cert_file` 和 `tls_key_file`
3. **访问控制**: 配置防火墙规则，使用反向代理，实施 IP 白名单
4. **定期备份**: 备份配置文件和监控数据，测试恢复流程

### 15.4 日志方案选择建议

| 场景 | 推荐方案 | 原因 |
|------|---------|------|
| 开发环境 | dev 模式 | stdout + text 格式，方便查看 |
| 生产环境（容器） | container 模式 | 解耦，Logstash 故障不影响应用 |
| 生产环境（直连） | direct 模式 | 实时推送，无需额外组件 |
| 需要日志持久化 | file 模式 | 日志文件可离线分析 |
| 高性能要求 | direct-udp 模式 | UDP 协议开销小 |

---

## 16. 附录

### 16.1 配置文件位置

```
ceph-exporter/
├── configs/
│   ├── ceph-exporter.yaml          # 主配置文件（英文）
│   ├── ceph-exporter.zh-CN.yaml    # 中文配置文件
│   ├── logger-examples.yaml        # 日志配置示例（6 种场景）
│   └── logstash.conf               # Logstash 管道配置
├── deployments/
│   ├── .env                        # 环境变量参考文件
│   ├── docker-compose.yml          # 标准监控栈配置
│   ├── docker-compose-lightweight-full.yml  # 完整轻量级栈（推荐）
│   ├── docker-compose-integration-test.yml  # 集成测试配置
│   ├── docker-compose-ceph-demo.yml         # 仅 Ceph Demo
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
│   │   │   ├── ceph-cluster.json   # Ceph 集群监控仪表盘
│   │   │   ├── grafana-metrics-zh.json    # Grafana 自身监控
│   │   │   └── prometheus-stats-zh.json   # Prometheus 监控
│   │   └── provisioning/
│   │       ├── datasources/datasource.yml  # 数据源自动配置
│   │       └── dashboards/dashboard.yml    # 仪表盘自动配置
│   ├── filebeat/
│   │   └── filebeat.yml            # Filebeat 配置
│   └── scripts/                    # 部署和运维脚本
│       ├── deploy.sh               # 主部署脚本
│       ├── diagnose.sh             # 统一诊断脚本
│       ├── verify-deployment.sh    # 部署验证脚本
│       ├── switch-logging-mode.sh  # 日志方案切换
│       ├── enable-jaeger-tracing.sh # Jaeger 启用脚本
│       ├── fix-deployment.sh       # 部署修复脚本
│       ├── clean-volumes.sh        # 数据清理脚本
│       ├── diagnose-logstash.sh    # Logstash 诊断
│       ├── diagnose-elk-full.sh    # ELK 完整诊断
│       └── ...
```

### 16.2 端口列表

| 服务 | 端口 | 协议 | 说明 |
|------|------|------|------|
| ceph-exporter | 9128 | HTTP | Prometheus 指标端点 |
| Prometheus | 9090 | HTTP | Web UI 和 API |
| Grafana | 3000 | HTTP | Web UI |
| Alertmanager | 9093 | HTTP | Web UI 和 API |
| Elasticsearch | 9200 | HTTP | REST API |
| Elasticsearch | 9300 | TCP | 节点间通信 |
| Logstash | 5044 | TCP | Beats 输入 (Filebeat) |
| Logstash | 5000 | TCP | TCP 直接输入 |
| Kibana | 5601 | HTTP | Web UI |
| Jaeger Collector | 4318 | HTTP | OTLP HTTP 接收 |
| Jaeger Collector | 14268 | HTTP | Thrift HTTP 接收 |
| Jaeger Collector | 6831 | UDP | Compact Thrift 接收 |
| Jaeger UI | 16686 | HTTP | Web UI |
| Ceph Dashboard | 8080 | HTTP | Web UI |
| Ceph RGW | 5000 | HTTP | S3 兼容对象存储 |

### 16.3 环境变量参考

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| VERSION | dev | 版本标识 |
| CEPH_EXPORTER_PORT | 9128 | ceph-exporter 端口 |
| CEPH_CONFIG | /etc/ceph/ceph.conf | Ceph 配置文件路径 |
| CEPH_USER | admin | Ceph 用户名 |
| LOG_LEVEL | info | 日志级别 |
| LOG_FORMAT | json | 日志格式 |
| LOGSTASH_URL | logstash:5000 | Logstash 地址 |
| JAEGER_URL | jaeger:4318 | Jaeger 地址 |
| SERVICE_NAME | ceph-exporter | 服务名称 |
| PROMETHEUS_PORT | 9090 | Prometheus 端口 |
| GRAFANA_PORT | 3000 | Grafana 端口 |
| GRAFANA_ADMIN_USER | admin | Grafana 管理员用户名 |
| GRAFANA_ADMIN_PASSWORD | admin | Grafana 管理员密码 |
| ALERTMANAGER_PORT | 9093 | Alertmanager 端口 |

### 16.4 相关文档

- [Docker 镜像配置](DOCKER_MIRROR_CONFIGURATION.md)
- [故障排查指南](ceph-exporter/deployments/TROUBLESHOOTING.md)
- [ELK 日志指南](ceph-exporter/docs/ELK-LOGGING-GUIDE.md)
- [Jaeger 追踪指南](ceph-exporter/docs/JAEGER-TRACING-GUIDE.md)
- [Alertmanager 使用指南](Alertmanager使用指南.md)
- [YAML 配置文件说明](YAML配置文件说明.md)
- [PRE_COMMIT 使用指南](PRE_COMMIT_使用指南.md)
- [数据存储说明](ceph-exporter/deployments/DATA_STORAGE.md)
- [时区配置说明](ceph-exporter/deployments/TIMEZONE_CONFIGURATION.md)

### 16.5 版本信息

- **文档版本**: 5.0
- **最后更新**: 2026-03-22
- **适用版本**: ceph-exporter 1.0+
- **Ceph 版本**: Octopus (15.x) 及以上

---

**文档结束**

如有问题或建议，请参考项目 README 或提交 Issue。
