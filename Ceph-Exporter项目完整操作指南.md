# Ceph-Exporter 项目完整操作指南

> **版本**: 1.0
> **最后更新**: 2026-03-11
> **适用环境**: CentOS 7 + Docker

---

## 📖 目录

1. [项目概述](#项目概述)
2. [服务架构](#服务架构)
3. [快速开始](#快速开始)
4. [核心服务详解](#核心服务详解)
   - [Ceph-Exporter 服务](#ceph-exporter-服务)
   - [Prometheus 监控服务](#prometheus-监控服务)
   - [Grafana 可视化服务](#grafana-可视化服务)
   - [Alertmanager 告警服务](#alertmanager-告警服务)
   - [Ceph Demo 测试集群](#ceph-demo-测试集群)
   - [ELK 日志系统](#elk-日志系统)
   - [Jaeger 追踪系统](#jaeger-追踪系统)
5. [部署脚本详解](#部署脚本详解)
6. [常用操作指南](#常用操作指南)
7. [界面术语对照表](#界面术语对照表)
8. [配置文件说明](#配置文件说明)
9. [数据管理](#数据管理)
10. [故障排查](#故障排查)
11. [最佳实践](#最佳实践)

---

## 项目概述

### 什么是 Ceph-Exporter？

Ceph-Exporter 是一个基于 Go 语言开发的 Ceph 集群 Prometheus 指标导出器。它通过 go-ceph 库与 Ceph 集群通信，采集集群状态、存储池、OSD、Monitor 等核心指标，并以 Prometheus 标准格式暴露，配合 Grafana 实现可视化监控。

### 核心特性

- ✅ **7 个 Prometheus 采集器**: Cluster、Pool、OSD、Monitor、Health、MDS、RGW
- ✅ **CGO 集成**: 使用 go-ceph 库直接与 Ceph 通信
- ✅ **完整测试**: 81 个单元测试，100% 通过率，覆盖率 68.1%
- ✅ **容器化部署**: Docker Compose 一键部署
- ✅ **可观测性**: 支持 OpenTelemetry 分布式追踪
- ✅ **插件系统**: 支持自定义插件扩展
- ✅ **中文界面**: Grafana 完全中文化，Prometheus 告警规则中文化

### 系统要求

| 项目 | 要求 |
|------|------|
| 操作系统 | CentOS 7.x |
| Docker | 19.03+ |
| Docker Compose | 1.25+ |
| 内存 | 4GB（推荐 8GB） |
| CPU | 2 核（推荐 4 核） |
| 磁盘 | 30GB |

---

## 服务架构

### 整体架构图

```
┌─────────────────────────────────────────────────────────────┐
│                        用户访问层                              │
│  Grafana (3000)  │  Prometheus (9090)  │  Kibana (5601)     │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│                        监控采集层                              │
│  Ceph-Exporter (9128)  │  Alertmanager (9093)               │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│                        数据存储层                              │
│  Prometheus TSDB  │  Elasticsearch  │  Jaeger                │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│                        数据源层                                │
│              Ceph Demo 集群 (6789, 8080)                      │
└─────────────────────────────────────────────────────────────┘
```

### 服务端口映射

| 服务 | 端口 | 用途 | 访问地址 |
|------|------|------|----------|
| Ceph-Exporter | 9128 | 指标导出 | http://localhost:9128/metrics |
| Prometheus | 9090 | 监控查询 | http://localhost:9090 |
| Grafana | 3000 | 可视化 | http://localhost:3000 |
| Alertmanager | 9093 | 告警管理 | http://localhost:9093 |
| Ceph Demo (Mon) | 6789 | Ceph Monitor | - |
| Ceph Demo (RGW) | 8080 | Ceph 对象网关 | http://localhost:8080 |
| Elasticsearch | 9200 | 日志存储 | http://localhost:9200 |
| Kibana | 5601 | 日志查询 | http://localhost:5601 |
| Jaeger | 16686 | 追踪查询 | http://localhost:16686 |

### 部署模式对比

| 模式 | 命令 | 包含组件 | 资源需求 | 适用场景 |
|------|------|----------|----------|----------|
| **完整监控栈** | `./scripts/deploy.sh full` | Ceph Demo + 监控 + ELK + Jaeger | 4-6GB | 演示、功能测试 ⭐ |
| **集成测试** | `./scripts/deploy.sh integration` | Ceph Demo + 监控 | 2-3GB | 开发测试 |
| **最小监控栈** | `./scripts/deploy.sh minimal` | 监控组件 | 1GB | 生产环境 |

---

## 快速开始

### 环境准备（首次部署必需）

#### 1. 安装 Docker

```bash
sudo yum install -y yum-utils
sudo yum-config-manager --add-repo https://mirrors.aliyun.com/docker-ce/linux/centos/docker-ce.repo
sudo yum install -y docker-ce docker-ce-cli containerd.io
sudo systemctl start docker
sudo systemctl enable docker
```

#### 2. 安装 Docker Compose

```bash
sudo curl -L "https://get.daocloud.io/docker/compose/releases/download/v2.24.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose
docker-compose --version
```

#### 3. 配置 Docker 镜像加速（国内必需）

```bash
sudo mkdir -p /etc/docker
sudo tee /etc/docker/daemon.json <<EOF
{
  "registry-mirrors": [
    "https://docker.mirrors.ustc.edu.cn",
    "https://hub-mirror.c.163.com"
  ]
}
EOF
sudo systemctl restart docker
```

#### 4. 配置防火墙（可选）

```bash
# 方式 1: 开放端口
sudo firewall-cmd --permanent --add-port=9128/tcp
sudo firewall-cmd --permanent --add-port=9090/tcp
sudo firewall-cmd --permanent --add-port=3000/tcp
sudo firewall-cmd --reload

# 方式 2: 临时关闭防火墙（仅测试环境）
sudo systemctl stop firewalld
```

### 一键部署

```bash
# 1. 进入部署目录
cd ceph-exporter/deployments

# 2. 赋予脚本执行权限
chmod +x scripts/deploy.sh

# 3. 完整部署（推荐）
./scripts/deploy.sh full

# 4. 等待服务启动（约 2-3 分钟）
# 脚本会自动：
# - 检查环境依赖
# - 创建数据目录
# - 设置正确的权限
# - 拉取 Docker 镜像
# - 启动所有服务
# - 验证服务健康状态
```

### 验证部署

```bash
# 查看服务状态
./scripts/deploy.sh status

# 验证所有服务
./scripts/deploy.sh verify

# 访问服务
curl http://localhost:9128/metrics  # Ceph-Exporter
curl http://localhost:9090          # Prometheus
curl http://localhost:3000          # Grafana
```

### 访问界面

| 服务 | 地址 | 默认凭据 |
|------|------|----------|
| Grafana | http://localhost:3000 | admin / admin |
| Prometheus | http://localhost:9090 | 无需认证 |
| Alertmanager | http://localhost:9093 | 无需认证 |
| Kibana | http://localhost:5601 | 无需认证 |
| Jaeger | http://localhost:16686 | 无需认证 |

---


## 核心服务详解（续）

### Alertmanager 告警服务

#### 服务说明

Alertmanager 负责接收 Prometheus 发送的告警，进行分组、抑制、静默和路由，然后发送通知。

#### 主要功能

1. **告警接收**: 接收 Prometheus 发送的告警
2. **告警分组**: 将相似的告警分组
3. **告警抑制**: 抑制重复或相关的告警
4. **告警静默**: 临时静默特定告警
5. **告警路由**: 根据规则路由告警到不同的接收器

#### 访问信息

- **地址**: http://localhost:9093
- **界面语言**: 英文（配置文件支持中文注释）

#### 界面导航

**主要页面**:

1. **Alerts（告警列表）**
   - 显示当前所有活跃的告警
   - 可以按标签过滤
   - 可以创建静默规则

2. **Silences（静默规则）**
   - 查看和管理静默规则
   - 创建新的静默规则
   - 过期的静默规则

3. **Status（状态信息）**
   - 查看 Alertmanager 配置
   - 查看集群状态（如果配置了高可用）

#### 常用操作

**操作 1: 查看活跃告警**
```bash
# 方式 1: 通过 Web UI
# 1. 访问 http://localhost:9093
# 2. 查看 Alerts 页面
# 3. 可以看到所有触发的告警

# 方式 2: 通过 API
curl -s http://localhost:9093/api/v2/alerts | jq '.[] | {labels: .labels, status: .status.state}'
```

**操作 2: 创建静默规则**
```bash
# 通过 Web UI:
# 1. 访问 http://localhost:9093
# 2. 点击 "Silences" 标签
# 3. 点击 "New Silence"
# 4. 填写匹配器（Matchers）:
#    - Name: alertname
#    - Value: CephOSDDown
# 5. 设置持续时间（Duration）
# 6. 填写创建者和注释
# 7. 点击 "Create"

# 通过 API:
curl -X POST http://localhost:9093/api/v2/silences \
  -H "Content-Type: application/json" \
  -d '{
    "matchers": [
      {
        "name": "alertname",
        "value": "CephOSDDown",
        "isRegex": false
      }
    ],
    "startsAt": "2026-03-11T00:00:00Z",
    "endsAt": "2026-03-11T23:59:59Z",
    "createdBy": "admin",
    "comment": "维护期间静默 OSD 告警"
  }'
```

**操作 3: 删除静默规则**
```bash
# 1. 访问 http://localhost:9093
# 2. 点击 "Silences" 标签
# 3. 找到要删除的静默规则
# 4. 点击 "Expire" 按钮
```

#### 配置文件

配置文件位置: `alertmanager/alertmanager.yml`

**主要配置项**:
```yaml
global:
  resolve_timeout: 5m  # 告警解决超时时间

route:
  group_by: ['alertname', 'cluster']  # 分组依据
  group_wait: 10s       # 分组等待时间
  group_interval: 10s   # 分组间隔
  repeat_interval: 1h   # 重复发送间隔
  receiver: 'default'   # 默认接收器

receivers:
  - name: 'default'
    # 可以配置多种通知方式:
    # - email
    # - webhook
    # - slack
    # - wechat
    # - etc.
```

---

### Ceph Demo 测试集群

#### 服务说明

Ceph Demo 是一个容器化的 Ceph 集群，用于开发和测试环境，包含 Monitor、OSD、MDS 和 RGW 组件。

#### 主要功能

1. **完整的 Ceph 集群**: 提供完整的 Ceph 功能
2. **快速部署**: 容器化部署，几分钟内启动
3. **测试环境**: 用于开发和测试 ceph-exporter
4. **学习工具**: 学习 Ceph 的理想环境

#### 访问信息

- **Monitor 端口**: 6789
- **RGW 端口**: 8080
- **容器名称**: ceph-demo

#### 常用操作

**操作 1: 查看集群状态**
```bash
# 查看集群健康状态
docker exec ceph-demo ceph -s

# 预期输出:
#   cluster:
#     id:     xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
#     health: HEALTH_OK
#   services:
#     mon: 1 daemons
#     mgr: 1 daemons
#     osd: 1 osds: 1 up, 1 in
#     rgw: 1 daemon active
```

**操作 2: 查看 OSD 状态**
```bash
# 查看 OSD 树
docker exec ceph-demo ceph osd tree

# 查看 OSD 详细信息
docker exec ceph-demo ceph osd stat

# 查看 OSD 使用情况
docker exec ceph-demo ceph osd df
```

**操作 3: 查看存储池**
```bash
# 列出所有存储池
docker exec ceph-demo ceph osd lspools

# 查看存储池详细信息
docker exec ceph-demo ceph df

# 查看特定存储池状态
docker exec ceph-demo ceph osd pool stats
```

**操作 4: 查看 Monitor 状态**
```bash
# 查看 Monitor 状态
docker exec ceph-demo ceph mon stat

# 查看 Quorum 状态
docker exec ceph-demo ceph quorum_status

# 查看 Monitor 详细信息
docker exec ceph-demo ceph mon dump
```

**操作 5: 查看 PG 状态**
```bash
# 查看 PG 统计
docker exec ceph-demo ceph pg stat

# 查看 PG 详细信息
docker exec ceph-demo ceph pg dump

# 查看非正常 PG
docker exec ceph-demo ceph pg dump_stuck
```

**操作 6: 测试 RGW**
```bash
# 测试 RGW 端口（返回 404 是正常的）
curl -v http://localhost:8080

# 创建 RGW 用户
docker exec ceph-demo radosgw-admin user create --uid=testuser --display-name="Test User"

# 列出 RGW 用户
docker exec ceph-demo radosgw-admin user list
```

#### 数据存储

Ceph Demo 的数据存储在 `data/ceph-demo/` 目录：

```
data/ceph-demo/
├── data/              # Ceph 集群数据（OSD、Mon 数据）
└── config/            # Ceph 配置文件
    ├── ceph.conf                      # Ceph 配置文件
    └── ceph.client.admin.keyring      # 管理员密钥环
```

#### 常见问题

**问题 1: Ceph 集群状态为 HEALTH_WARN**
```bash
# 查看详细健康信息
docker exec ceph-demo ceph health detail

# 常见原因:
# 1. OSD 数量不足（需要至少 3 个 OSD 才能达到 HEALTH_OK）
# 2. PG 数量不合适
# 3. 时钟偏差

# 解决方案: Demo 环境中 HEALTH_WARN 是正常的
```

**问题 2: 无法连接到 Ceph 集群**
```bash
# 检查容器状态
docker ps | grep ceph-demo

# 检查容器日志
docker logs ceph-demo --tail 100

# 检查网络连接
docker exec ceph-exporter ping ceph-demo

# 检查配置文件
ls -la data/ceph-demo/config/
```

---

### ELK 日志系统

#### 服务说明

ELK 是 Elasticsearch、Logstash、Kibana 的组合，用于日志收集、存储、分析和可视化。

#### 组件说明

1. **Elasticsearch**: 日志存储和搜索引擎
2. **Logstash**: 日志收集和处理管道
3. **Kibana**: 日志可视化和分析界面

#### 访问信息

- **Elasticsearch**: http://localhost:9200
- **Kibana**: http://localhost:5601
- **界面语言**: 简体中文

#### Kibana 常用操作

**首次访问**
```bash
# 1. 访问 http://localhost:5601
# 2. 等待 Kibana 初始化（首次启动需要几分钟）
# 3. 进入主界面
```

**创建索引模式**
```bash
# 1. 点击左上角菜单 → Stack Management
# 2. 点击 "索引模式"
# 3. 点击 "创建索引模式"
# 4. 输入索引模式: logstash-*
# 5. 选择时间字段: @timestamp
# 6. 点击 "创建索引模式"
```

**查看日志**
```bash
# 1. 点击左上角菜单 → Discover
# 2. 选择索引模式: logstash-*
# 3. 选择时间范围
# 4. 可以看到所有日志
# 5. 使用 KQL 查询语言过滤日志
```

**KQL 查询示例**
```
# 查询特定服务的日志
service: "ceph-exporter"

# 查询错误日志
level: "error"

# 查询包含特定关键词的日志
message: "connection failed"

# 组合查询
service: "ceph-exporter" AND level: "error"
```

---

### Jaeger 追踪系统

#### 服务说明

Jaeger 是一个开源的分布式追踪系统，用于监控和排查微服务架构中的性能问题。

#### 主要功能

1. **分布式追踪**: 追踪请求在系统中的完整路径
2. **性能分析**: 分析每个操作的耗时
3. **依赖分析**: 分析服务之间的依赖关系
4. **根因分析**: 快速定位性能瓶颈

#### 访问信息

- **地址**: http://localhost:16686
- **界面语言**: 英文

#### 常用操作

**查看追踪**
```bash
# 1. 访问 http://localhost:16686
# 2. 在 "Service" 下拉菜单中选择 "ceph-exporter"
# 3. 点击 "Find Traces"
# 4. 可以看到所有追踪记录
# 5. 点击任意追踪查看详情
```

**分析性能**
```bash
# 1. 选择一个追踪记录
# 2. 查看 Span 列表
# 3. 每个 Span 显示:
#    - 操作名称
#    - 开始时间
#    - 持续时间
#    - 标签和日志
# 4. 找出耗时最长的操作
```

---


## 部署脚本详解

### deploy.sh - 主部署脚本

#### 脚本位置

`ceph-exporter/deployments/scripts/deploy.sh`

#### 主要功能

这是项目的核心部署脚本，提供了完整的部署、管理和诊断功能。

#### 命令列表

| 命令 | 说明 | 示例 |
|------|------|------|
| `check` | 检查环境依赖 | `./scripts/deploy.sh check` |
| `init` | 初始化数据目录和权限 | `./scripts/deploy.sh init` |
| `full` | 完整部署（推荐） | `./scripts/deploy.sh full` |
| `integration` | 集成测试环境部署 | `./scripts/deploy.sh integration` |
| `minimal` | 最小监控栈部署 | `./scripts/deploy.sh minimal` |
| `status` | 查看服务状态 | `./scripts/deploy.sh status` |
| `logs` | 查看服务日志 | `./scripts/deploy.sh logs [service]` |
| `verify` | 验证部署 | `./scripts/deploy.sh verify` |
| `diagnose` | 诊断问题 | `./scripts/deploy.sh diagnose [service]` |
| `fix` | 修复部署问题 | `./scripts/deploy.sh fix` |
| `stop` | 停止服务 | `./scripts/deploy.sh stop` |
| `clean` | 清理数据 | `./scripts/deploy.sh clean` |
| `help` | 显示帮助信息 | `./scripts/deploy.sh help` |

#### 详细说明

**1. check - 环境检查**
```bash
./scripts/deploy.sh check

# 检查内容:
# - Docker 是否安装
# - Docker Compose 是否安装
# - Docker 服务是否运行
# - 系统资源（内存、磁盘）
# - 端口占用情况
```

**2. init - 初始化**
```bash
./scripts/deploy.sh init

# 执行操作:
# - 创建数据目录结构
# - 设置正确的目录权限
#   - Prometheus: 65534:65534
#   - Grafana: 472:472
#   - Elasticsearch: 1000:1000
# - 创建 configs 软链接
# - 设置 vm.max_map_count（Elasticsearch 需要）
```

**3. full - 完整部署**
```bash
./scripts/deploy.sh full

# 部署内容:
# - Ceph Demo 集群
# - Ceph-Exporter
# - Prometheus
# - Grafana
# - Alertmanager
# - Elasticsearch
# - Logstash
# - Kibana
# - Jaeger

# 执行流程:
# 1. 环境检查
# 2. 初始化数据目录
# 3. 拉取 Docker 镜像
# 4. 启动所有服务
# 5. 等待服务就绪
# 6. 验证部署
```

**4. integration - 集成测试**
```bash
./scripts/deploy.sh integration

# 部署内容:
# - Ceph Demo 集群
# - Ceph-Exporter
# - Prometheus
# - Grafana

# 适用场景:
# - 开发测试
# - CI/CD 集成测试
# - 功能验证
```

**5. minimal - 最小部署**
```bash
./scripts/deploy.sh minimal

# 部署内容:
# - Ceph-Exporter（连接现有 Ceph 集群）
# - Prometheus
# - Grafana
# - Alertmanager

# 适用场景:
# - 生产环境
# - 已有 Ceph 集群
# - 资源受限环境
```

**6. status - 查看状态**
```bash
./scripts/deploy.sh status

# 显示信息:
# - 容器运行状态
# - 容器健康状态
# - 端口映射
# - 资源使用情况
```

**7. logs - 查看日志**
```bash
# 查看所有服务日志
./scripts/deploy.sh logs

# 查看特定服务日志
./scripts/deploy.sh logs ceph-exporter
./scripts/deploy.sh logs prometheus
./scripts/deploy.sh logs grafana

# 实时跟踪日志
./scripts/deploy.sh logs ceph-exporter -f
```

**8. verify - 验证部署**
```bash
./scripts/deploy.sh verify

# 验证内容:
# - 容器运行状态
# - 服务端点可访问性
# - Prometheus 采集目标状态
# - Grafana 数据源配置
# - 时区配置
# - 资源使用情况

# 输出示例:
# ✓ ceph-demo 运行正常
# ✓ ceph-exporter 运行正常
# ✓ prometheus 运行正常
# ✓ grafana 运行正常
```

**9. diagnose - 诊断问题**
```bash
# 诊断所有服务
./scripts/deploy.sh diagnose

# 诊断特定服务
./scripts/deploy.sh diagnose ceph-exporter

# 诊断内容:
# - 容器状态和日志
# - 配置文件检查
# - 网络连接测试
# - 权限检查
# - 资源使用分析
```

**10. fix - 修复问题**
```bash
sudo ./scripts/deploy.sh fix

# 修复内容:
# - 数据目录权限
# - configs 软链接
# - Ceph keyring 权限
# - vm.max_map_count 设置
# - 重启失败的服务

# 注意: 需要 root 权限
```

**11. stop - 停止服务**
```bash
./scripts/deploy.sh stop

# 执行操作:
# - 停止所有容器
# - 保留数据（不删除 data/ 目录）
```

**12. clean - 清理数据**
```bash
./scripts/deploy.sh clean

# 执行操作:
# - 停止所有容器
# - 删除容器
# - 删除 data/ 目录
# - 删除网络

# 警告: 此操作会删除所有数据，无法恢复！
```

#### 使用示例

**完整部署流程**
```bash
# 1. 进入部署目录
cd ceph-exporter/deployments

# 2. 检查环境
./scripts/deploy.sh check

# 3. 完整部署
./scripts/deploy.sh full

# 4. 等待服务启动（约 2-3 分钟）
sleep 120

# 5. 验证部署
./scripts/deploy.sh verify

# 6. 查看服务状态
./scripts/deploy.sh status
```

**故障排查流程**
```bash
# 1. 查看服务状态
./scripts/deploy.sh status

# 2. 查看问题服务的日志
./scripts/deploy.sh logs ceph-exporter

# 3. 运行诊断
./scripts/deploy.sh diagnose ceph-exporter

# 4. 尝试修复
sudo ./scripts/deploy.sh fix

# 5. 重新验证
./scripts/deploy.sh verify
```

---

## 常用操作指南

### 日常监控操作

#### 1. 查看集群健康状态

**通过 Grafana（推荐）**
```bash
# 1. 访问 http://localhost:3000
# 2. 登录（admin/admin）
# 3. 打开 "Ceph 集群监控" 仪表板
# 4. 查看 "集群概览" 面板
# 5. 健康状态显示:
#    - 绿色: HEALTH_OK（健康）
#    - 黄色: HEALTH_WARN（警告）
#    - 红色: HEALTH_ERR（错误）
```

**通过 Prometheus**
```bash
# 查询健康状态
curl -s 'http://localhost:9090/api/v1/query?query=ceph_health_status' | jq '.data.result[0].value[1]'

# 返回值:
# "0" = HEALTH_OK
# "1" = HEALTH_WARN
# "2" = HEALTH_ERR
```

**通过 Ceph 命令**
```bash
# 查看详细健康信息
docker exec ceph-demo ceph health detail
```

#### 2. 查看集群容量

**通过 Grafana**
```bash
# 在 "Ceph 集群监控" 仪表板中查看:
# - 总容量
# - 已用容量
# - 可用容量
# - 使用率百分比
```

**通过 Prometheus**
```bash
# 查询总容量（字节）
curl -s 'http://localhost:9090/api/v1/query?query=ceph_cluster_total_bytes' | jq '.data.result[0].value[1]'

# 查询已用容量（字节）
curl -s 'http://localhost:9090/api/v1/query?query=ceph_cluster_used_bytes' | jq '.data.result[0].value[1]'

# 查询使用率（百分比）
curl -s 'http://localhost:9090/api/v1/query?query=(ceph_cluster_used_bytes/ceph_cluster_total_bytes)*100' | jq '.data.result[0].value[1]'
```

**通过 Ceph 命令**
```bash
# 查看集群容量
docker exec ceph-demo ceph df
```

#### 3. 查看 OSD 状态

**通过 Grafana**
```bash
# 在 "Ceph 集群监控" 仪表板中查看:
# - OSD 总数
# - 在线 OSD 数量
# - OSD 延迟
# - OSD 使用率
```

**通过 Prometheus**
```bash
# 查询 OSD 总数
curl -s 'http://localhost:9090/api/v1/query?query=ceph_cluster_osds_total' | jq '.data.result[0].value[1]'

# 查询在线 OSD 数量
curl -s 'http://localhost:9090/api/v1/query?query=ceph_cluster_osds_up' | jq '.data.result[0].value[1]'

# 查询 OSD 延迟
curl -s 'http://localhost:9090/api/v1/query?query=ceph_osd_apply_latency_ms' | jq '.data.result'
```

**通过 Ceph 命令**
```bash
# 查看 OSD 状态
docker exec ceph-demo ceph osd stat

# 查看 OSD 树
docker exec ceph-demo ceph osd tree

# 查看 OSD 使用情况
docker exec ceph-demo ceph osd df
```

#### 4. 查看告警

**通过 Grafana**
```bash
# 1. 访问 http://localhost:3000
# 2. 点击左侧菜单 "告警"
# 3. 查看所有活跃的告警
```

**通过 Prometheus**
```bash
# 1. 访问 http://localhost:9090
# 2. 点击顶部菜单 "Alerts"
# 3. 查看告警规则状态
```

**通过 Alertmanager**
```bash
# 1. 访问 http://localhost:9093
# 2. 查看活跃告警列表

# 或通过 API
curl -s http://localhost:9093/api/v2/alerts | jq '.[] | {labels: .labels, status: .status.state}'
```

#### 5. 查看日志

**通过 Kibana**
```bash
# 1. 访问 http://localhost:5601
# 2. 点击左上角菜单 → Discover
# 3. 选择索引模式: logstash-*
# 4. 查看所有日志
# 5. 使用 KQL 过滤日志
```

**通过 Docker**
```bash
# 查看 ceph-exporter 日志
docker logs ceph-exporter --tail 100

# 实时跟踪日志
docker logs -f ceph-exporter

# 查看特定时间范围的日志
docker logs ceph-exporter --since 1h

# 查看所有服务日志
docker-compose logs --tail 100
```

### 数据管理操作

#### 1. 备份数据

**备份所有数据**
```bash
cd ceph-exporter/deployments

# 创建备份
tar -czf ceph-exporter-backup-$(date +%Y%m%d-%H%M%S).tar.gz data/

# 查看备份文件
ls -lh ceph-exporter-backup-*.tar.gz
```

**备份特定服务数据**
```bash
# 备份 Prometheus 数据
tar -czf prometheus-backup-$(date +%Y%m%d).tar.gz data/prometheus/

# 备份 Grafana 数据
tar -czf grafana-backup-$(date +%Y%m%d).tar.gz data/grafana/

# 备份 Ceph Demo 数据
tar -czf ceph-demo-backup-$(date +%Y%m%d).tar.gz data/ceph-demo/
```

#### 2. 恢复数据

```bash
# 停止服务
docker-compose down

# 恢复数据
tar -xzf ceph-exporter-backup-20260311-120000.tar.gz

# 启动服务
docker-compose up -d

# 验证恢复
./scripts/deploy.sh verify
```

#### 3. 清理数据

**清理所有数据**
```bash
# 使用部署脚本（推荐）
./scripts/deploy.sh clean

# 或手动清理
docker-compose down
rm -rf data/
```

**清理特定服务数据**
```bash
# 停止服务
docker-compose stop prometheus

# 清理数据
rm -rf data/prometheus/*

# 重启服务
docker-compose start prometheus
```

#### 4. 查看数据占用

```bash
# 查看总体占用
du -sh data/

# 查看各服务占用
du -sh data/*

# 详细查看
du -h --max-depth=2 data/

# 输出示例:
# 2.5G    data/prometheus
# 150M    data/grafana
# 1.2G    data/elasticsearch
# 500M    data/ceph-demo
```

### 服务管理操作

#### 1. 启动服务

```bash
# 启动所有服务
docker-compose up -d

# 启动特定服务
docker-compose up -d ceph-exporter
docker-compose up -d prometheus
docker-compose up -d grafana

# 查看启动日志
docker-compose logs -f
```

#### 2. 停止服务

```bash
# 停止所有服务
docker-compose stop

# 停止特定服务
docker-compose stop ceph-exporter
docker-compose stop prometheus

# 停止并删除容器
docker-compose down
```

#### 3. 重启服务

```bash
# 重启所有服务
docker-compose restart

# 重启特定服务
docker-compose restart ceph-exporter
docker-compose restart prometheus

# 重新创建并启动服务
docker-compose up -d --force-recreate
```

#### 4. 更新服务

```bash
# 拉取最新镜像
docker-compose pull

# 重新创建并启动服务
docker-compose up -d

# 查看更新后的版本
docker-compose images
```

### 配置管理操作

#### 1. 修改配置文件

**修改 Prometheus 配置**
```bash
# 编辑配置文件
vi prometheus/prometheus.yml

# 验证配置
docker exec prometheus promtool check config /etc/prometheus/prometheus.yml

# 重新加载配置（无需重启）
curl -X POST http://localhost:9090/-/reload
```

**修改 Grafana 配置**
```bash
# 编辑环境变量
vi docker-compose.yml

# 重启 Grafana
docker-compose restart grafana
```

**修改告警规则**
```bash
# 编辑告警规则
vi prometheus/alert_rules.yml

# 验证规则
docker exec prometheus promtool check rules /etc/prometheus/alert_rules.yml

# 重新加载配置
curl -X POST http://localhost:9090/-/reload
```

#### 2. 导出配置

**导出 Grafana 仪表板**
```bash
# 通过 API 导出
curl -s http://admin:admin@localhost:3000/api/dashboards/uid/ceph-cluster | jq '.dashboard' > ceph-dashboard-backup.json
```

**导出 Prometheus 配置**
```bash
# 复制配置文件
cp prometheus/prometheus.yml prometheus-backup.yml
cp prometheus/alert_rules.yml alert-rules-backup.yml
```

#### 3. 导入配置

**导入 Grafana 仪表板**
```bash
# 通过 Web UI:
# 1. 访问 http://localhost:3000
# 2. 点击左侧菜单 "仪表盘" → "导入"
# 3. 上传 JSON 文件或粘贴 JSON 内容
# 4. 选择数据源
# 5. 点击 "导入"
```

---


## 界面术语对照表

### Prometheus 界面术语

| 英文术语 | 中文翻译 | 说明 |
|---------|---------|------|
| Graph | 图表 | 查询和可视化页面 |
| Alerts | 告警 | 告警规则和状态 |
| Status | 状态 | 系统状态信息 |
| Targets | 采集目标 | 监控目标列表 |
| Rules | 规则 | 告警规则详情 |
| Configuration | 配置 | 配置文件内容 |
| Service Discovery | 服务发现 | 自动发现的目标 |
| Runtime & Build Information | 运行时和构建信息 | 版本和配置信息 |
| Command-Line Flags | 命令行参数 | 启动参数 |
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
| Add Graph | 添加图表 | 添加新的查询面板 |

### Grafana 界面术语（中文界面）

Grafana 已完全中文化，以下是常用术语：

| 功能 | 中文名称 | 说明 |
|------|---------|------|
| Dashboard | 仪表盘 | 可视化面板集合 |
| Panel | 面板 | 单个可视化组件 |
| Data Source | 数据源 | 数据来源配置 |
| Query | 查询 | 数据查询语句 |
| Alert | 告警 | 告警规则 |
| Explore | 探索 | 临时查询界面 |
| Playlist | 播放列表 | 自动轮播仪表盘 |
| Snapshot | 快照 | 仪表盘快照 |
| Folder | 文件夹 | 仪表盘分组 |
| Organization | 组织 | 多租户管理 |
| User | 用户 | 用户管理 |
| Team | 团队 | 团队管理 |
| Plugin | 插件 | 扩展功能 |
| Provisioning | 自动配置 | 自动化配置 |

### Alertmanager 界面术语

| 英文术语 | 中文翻译 | 说明 |
|---------|---------|------|
| Alerts | 告警 | 告警列表 |
| Silences | 静默 | 静默规则 |
| Status | 状态 | 系统状态 |
| New Silence | 新建静默 | 创建静默规则 |
| Matchers | 匹配器 | 告警匹配条件 |
| Duration | 持续时间 | 静默持续时间 |
| Creator | 创建者 | 创建人 |
| Comment | 注释 | 说明信息 |
| Expire | 过期 | 使静默规则失效 |
| Active | 活跃 | 活跃的告警 |
| Suppressed | 已抑制 | 被抑制的告警 |
| Unprocessed | 未处理 | 未处理的告警 |

### Kibana 界面术语（中文界面）

Kibana 已完全中文化，以下是常用术语：

| 功能 | 中文名称 | 说明 |
|------|---------|------|
| Discover | 发现 | 日志查询界面 |
| Visualize | 可视化 | 创建可视化图表 |
| Dashboard | 仪表板 | 可视化面板集合 |
| Canvas | 画布 | 自定义报告 |
| Maps | 地图 | 地理数据可视化 |
| Machine Learning | 机器学习 | 异常检测 |
| Index Pattern | 索引模式 | 索引匹配规则 |
| Saved Objects | 已保存对象 | 保存的查询和可视化 |
| Stack Management | 堆栈管理 | 系统管理 |
| Dev Tools | 开发工具 | API 调试工具 |

### Ceph 命令术语

| 英文术语 | 中文翻译 | 说明 |
|---------|---------|------|
| Health | 健康状态 | 集群健康状态 |
| Status | 状态 | 集群状态 |
| OSD | 对象存储守护进程 | 存储节点 |
| Monitor | 监视器 | 集群监控节点 |
| MDS | 元数据服务器 | CephFS 元数据服务 |
| RGW | RADOS 网关 | 对象存储网关 |
| Pool | 存储池 | 数据存储池 |
| PG | 归置组 | 数据分布单元 |
| Quorum | 法定人数 | Monitor 集群状态 |
| Up | 在线 | 服务在线 |
| Down | 离线 | 服务离线 |
| In | 在集群中 | OSD 在集群中 |
| Out | 不在集群中 | OSD 被踢出 |
| Active | 活跃 | PG 活跃状态 |
| Clean | 干净 | PG 正常状态 |
| Degraded | 降级 | PG 降级状态 |
| Peering | 对等 | PG 同步状态 |
| Backfill | 回填 | 数据恢复 |
| Recovery | 恢复 | 数据恢复 |

---

## 配置文件说明

### 项目配置文件结构

```
ceph-exporter/
├── configs/
│   └── ceph-exporter.yaml              # Ceph-Exporter 配置
├── deployments/
│   ├── prometheus/
│   │   ├── prometheus.yml              # Prometheus 主配置
│   │   └── alert_rules.yml             # 告警规则（中文）
│   ├── alertmanager/
│   │   └── alertmanager.yml            # Alertmanager 配置
│   ├── grafana/
│   │   ├── provisioning/
│   │   │   ├── datasources/            # 数据源自动配置
│   │   │   └── dashboards/             # 仪表板自动配置
│   │   └── dashboards/
│   │       └── ceph-cluster.json       # Ceph 集群监控仪表板
│   └── logstash/
│       └── logstash.conf               # Logstash 配置
└── docker-compose*.yml                 # Docker Compose 配置
```

### Ceph-Exporter 配置

**文件位置**: `configs/ceph-exporter.yaml`

**完整配置示例**:
```yaml
# 服务器配置
server:
  address: ":9128"                      # 监听地址和端口
  read_timeout: 30s                     # 读取超时
  write_timeout: 30s                    # 写入超时
  idle_timeout: 60s                     # 空闲超时

# Ceph 连接配置
ceph:
  config_file: "/etc/ceph/ceph.conf"    # Ceph 配置文件路径
  keyring_file: "/etc/ceph/ceph.client.admin.keyring"  # 密钥环文件路径
  user: "admin"                         # Ceph 用户名
  cluster: "ceph"                       # 集群名称

# 日志配置
logger:
  level: "info"                         # 日志级别: debug, info, warn, error
  format: "json"                        # 日志格式: json, text
  output: "stdout"                      # 输出: stdout, stderr, file
  file: "/var/log/ceph-exporter.log"    # 日志文件路径（当 output=file 时）

# 追踪配置
tracer:
  enabled: true                         # 是否启用追踪
  endpoint: "jaeger:4318"               # Jaeger 端点
  service_name: "ceph-exporter"         # 服务名称
  sample_rate: 1.0                      # 采样率 (0.0-1.0)

# 插件配置
plugins:
  enabled: true                         # 是否启用插件系统
  directory: "/etc/ceph-exporter/plugins"  # 插件目录
```

**配置说明**:

- **server.address**: 服务监听地址，格式为 `host:port`，`:9128` 表示监听所有网卡的 9128 端口
- **ceph.config_file**: Ceph 配置文件路径，必须可读
- **ceph.keyring_file**: Ceph 认证密钥环文件，必须可读
- **logger.level**: 日志级别，生产环境建议使用 `info`，调试时使用 `debug`
- **tracer.enabled**: 是否启用分布式追踪，需要配合 Jaeger 使用

### Prometheus 配置

**文件位置**: `deployments/prometheus/prometheus.yml`

**完整配置示例**:
```yaml
# 全局配置
global:
  scrape_interval: 15s                  # 采集间隔
  evaluation_interval: 15s              # 告警评估间隔
  scrape_timeout: 10s                   # 采集超时

# 告警规则文件
rule_files:
  - /etc/prometheus/alert_rules.yml     # 告警规则文件路径

# 告警管理器配置
alerting:
  alertmanagers:
    - static_configs:
        - targets:
            - alertmanager:9093         # Alertmanager 地址

# 采集配置
scrape_configs:
  # Ceph Exporter
  - job_name: 'ceph-exporter'
    static_configs:
      - targets: ['ceph-exporter:9128']
        labels:
          cluster: 'ceph-demo'
          environment: 'test'

  # Prometheus 自身
  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']

  # Alertmanager
  - job_name: 'alertmanager'
    static_configs:
      - targets: ['alertmanager:9093']
```

**配置说明**:

- **scrape_interval**: 采集间隔，建议 15s-60s
- **scrape_timeout**: 采集超时，必须小于 scrape_interval
- **job_name**: 采集任务名称，会作为 `job` 标签
- **targets**: 采集目标列表，格式为 `host:port`
- **labels**: 自定义标签，会添加到所有指标上

### Prometheus 告警规则

**文件位置**: `deployments/prometheus/alert_rules.yml`

**告警规则示例**:
```yaml
groups:
  - name: ceph_cluster_alerts
    interval: 30s
    rules:
      # 集群健康状态告警
      - alert: CephClusterWarning
        expr: ceph_health_status == 1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Ceph 集群处于 HEALTH_WARN 状态"
          description: "集群健康状态为 WARN，请检查集群状态。当前值: {{ $value }}"

      - alert: CephClusterError
        expr: ceph_health_status == 2
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Ceph 集群处于 HEALTH_ERR 状态"
          description: "集群健康状态为 ERR，需要立即处理！当前值: {{ $value }}"

      # OSD 告警
      - alert: CephOSDDown
        expr: ceph_cluster_osds_up < ceph_cluster_osds_total
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "OSD 节点宕机"
          description: "有 {{ $value }} 个 OSD 节点宕机"

      # 容量告警
      - alert: CephClusterCapacityWarning
        expr: (ceph_cluster_used_bytes / ceph_cluster_total_bytes) * 100 > 75
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "集群容量使用率超过 75%"
          description: "当前使用率: {{ $value | humanizePercentage }}"
```

**告警规则说明**:

- **alert**: 告警名称
- **expr**: PromQL 表达式，返回非空结果时触发告警
- **for**: 持续时间，表达式持续满足多久后触发告警
- **labels**: 告警标签，用于路由和分组
- **annotations**: 告警注释，包含摘要和详细描述

### Grafana 配置

**环境变量配置** (在 docker-compose.yml 中):
```yaml
grafana:
  environment:
    - GF_DEFAULT_LOCALE=zh-CN           # 默认语言
    - GF_SECURITY_ADMIN_PASSWORD=admin  # 管理员密码
    - GF_USERS_ALLOW_SIGN_UP=false      # 禁止注册
    - GF_AUTH_ANONYMOUS_ENABLED=false   # 禁止匿名访问
```

**数据源自动配置** (`grafana/provisioning/datasources/datasource.yml`):
```yaml
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true
    editable: false
```

**仪表板自动配置** (`grafana/provisioning/dashboards/dashboard.yml`):
```yaml
apiVersion: 1

providers:
  - name: 'Ceph Monitoring'
    orgId: 1
    folder: ''
    type: file
    disableDeletion: false
    updateIntervalSeconds: 10
    allowUiUpdates: true
    options:
      path: /etc/grafana/dashboards
```

### Docker Compose 配置

**主要配置文件**:

1. **docker-compose.yml** - 标准监控栈
2. **docker-compose-integration-test.yml** - 集成测试环境
3. **docker-compose-lightweight-full.yml** - 完整监控栈

**配置示例** (docker-compose.yml):
```yaml
version: '3.8'

services:
  ceph-exporter:
    image: ceph-exporter:latest
    container_name: ceph-exporter
    ports:
      - "9128:9128"
    volumes:
      - ./configs:/etc/ceph-exporter:ro
      - ./data/ceph-demo/config:/etc/ceph:ro
      - /etc/localtime:/etc/localtime:ro
      - /etc/timezone:/etc/timezone:ro
    networks:
      - ceph-network
    depends_on:
      - ceph-demo
    restart: unless-stopped

  prometheus:
    image: prom/prometheus:latest
    container_name: prometheus
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus:/etc/prometheus:ro
      - ./data/prometheus:/prometheus
      - /etc/localtime:/etc/localtime:ro
      - /etc/timezone:/etc/timezone:ro
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--storage.tsdb.retention.time=30d'
    networks:
      - ceph-network
    restart: unless-stopped

  grafana:
    image: grafana/grafana:latest
    container_name: grafana
    ports:
      - "3000:3000"
    volumes:
      - ./data/grafana:/var/lib/grafana
      - ./grafana/provisioning:/etc/grafana/provisioning:ro
      - ./grafana/dashboards:/etc/grafana/dashboards:ro
      - /etc/localtime:/etc/localtime:ro
      - /etc/timezone:/etc/timezone:ro
    environment:
      - GF_DEFAULT_LOCALE=zh-CN
      - GF_SECURITY_ADMIN_PASSWORD=admin
    networks:
      - ceph-network
    restart: unless-stopped

networks:
  ceph-network:
    driver: bridge
```

**配置说明**:

- **volumes**: 数据卷挂载，`:ro` 表示只读
- **ports**: 端口映射，格式为 `host:container`
- **networks**: 网络配置，所有服务在同一网络中
- **restart**: 重启策略，`unless-stopped` 表示除非手动停止，否则总是重启
- **depends_on**: 依赖关系，确保服务启动顺序

---

## 数据管理

### 数据目录结构

```
deployments/data/
├── ceph-demo/                  # Ceph Demo 数据
│   ├── data/                   # Ceph 集群数据
│   │   ├── mon/                # Monitor 数据
│   │   ├── osd/                # OSD 数据
│   │   └── mds/                # MDS 数据
│   └── config/                 # Ceph 配置文件
│       ├── ceph.conf           # Ceph 配置
│       └── ceph.client.admin.keyring  # 管理员密钥
├── prometheus/                 # Prometheus 数据
│   ├── wal/                    # 预写日志
│   ├── chunks_head/            # 内存数据块
│   └── 01XXXXX/                # 时序数据块
├── grafana/                    # Grafana 数据
│   ├── grafana.db              # SQLite 数据库
│   ├── plugins/                # 插件
│   └── png/                    # 图片缓存
├── alertmanager/               # Alertmanager 数据
│   └── nflog                   # 通知日志
├── elasticsearch/              # Elasticsearch 数据
│   └── nodes/                  # 节点数据
└── test/                       # 测试环境数据
    ├── ceph-demo/
    ├── prometheus/
    └── grafana/
```

### 数据权限要求

| 服务 | 运行用户 | UID | 数据目录权限 |
|------|---------|-----|-------------|
| Prometheus | nobody | 65534 | 65534:65534 |
| Grafana | grafana | 472 | 472:472 |
| Elasticsearch | elasticsearch | 1000 | 1000:1000 |
| Alertmanager | 当前用户 | $UID | $USER:$USER |
| Ceph Demo | root | 0 | root:root |

### 数据备份策略

**1. 定期备份**
```bash
# 每日备份脚本
#!/bin/bash
BACKUP_DIR="/backup/ceph-exporter"
DATE=$(date +%Y%m%d)

cd /home/lfl/ceph-exporter/deployments
tar -czf ${BACKUP_DIR}/ceph-exporter-${DATE}.tar.gz data/

# 保留最近 7 天的备份
find ${BACKUP_DIR} -name "ceph-exporter-*.tar.gz" -mtime +7 -delete
```

**2. 增量备份**
```bash
# 使用 rsync 进行增量备份
rsync -av --delete data/ /backup/ceph-exporter/data/
```

**3. 远程备份**
```bash
# 备份到远程服务器
tar -czf - data/ | ssh backup-server "cat > /backup/ceph-exporter-$(date +%Y%m%d).tar.gz"
```

### 数据恢复流程

**1. 完全恢复**
```bash
# 停止服务
docker-compose down

# 删除现有数据
rm -rf data/

# 恢复备份
tar -xzf ceph-exporter-backup-20260311.tar.gz

# 修复权限
sudo chown -R 65534:65534 data/prometheus
sudo chown -R 472:472 data/grafana
sudo chown -R 1000:1000 data/elasticsearch

# 启动服务
docker-compose up -d
```

**2. 部分恢复**
```bash
# 只恢复 Grafana 数据
docker-compose stop grafana
rm -rf data/grafana/*
tar -xzf grafana-backup-20260311.tar.gz
sudo chown -R 472:472 data/grafana
docker-compose start grafana
```

### 数据清理策略

**1. Prometheus 数据清理**
```bash
# Prometheus 自动清理（通过配置）
# 在 prometheus.yml 中设置:
# --storage.tsdb.retention.time=30d

# 手动清理旧数据
docker-compose stop prometheus
rm -rf data/prometheus/*
docker-compose start prometheus
```

**2. Elasticsearch 数据清理**
```bash
# 删除旧索引（保留最近 7 天）
curl -X DELETE "localhost:9200/logstash-$(date -d '7 days ago' +%Y.%m.%d)"
```

**3. 日志清理**
```bash
# 清理 Docker 日志
docker-compose logs --tail 0

# 清理系统日志
sudo journalctl --vacuum-time=7d
```

---


## 故障排查

### 常见问题分类

#### 1. 服务启动问题

**问题 1.1: Prometheus 不断重启**

**症状**:
```bash
$ docker ps
CONTAINER ID   NAME         STATUS
abc123         prometheus   Restarting (1) 10 seconds ago
```

**原因**: 数据目录权限不正确

**解决方案**:
```bash
# 方式 1: 使用部署脚本自动修复
sudo ./scripts/deploy.sh init

# 方式 2: 手动修复权限
sudo chown -R 65534:65534 data/prometheus
docker-compose restart prometheus

# 验证
docker logs prometheus --tail 20
```

**问题 1.2: Ceph-Exporter 连接失败**

**症状**:
```bash
$ docker logs ceph-exporter
{"level":"error","message":"连接 Ceph 集群失败: rados: ret=-13, Permission denied"}
```

**原因**:
1. configs 目录软链接不存在
2. Ceph keyring 文件权限不正确
3. Ceph 集群尚未完全启动

**解决方案**:
```bash
# 1. 检查并创建 configs 软链接
cd deployments
ls -la configs || ln -s ../configs configs

# 2. 等待 ceph-demo 完全启动
docker logs ceph-demo --tail 50

# 3. 修复 keyring 权限
sudo chmod 644 data/ceph-demo/config/ceph.client.admin.keyring

# 4. 重启 ceph-exporter
docker-compose restart ceph-exporter

# 5. 验证连接
curl http://localhost:9128/metrics | head -20
```

**问题 1.3: Grafana 无法启动**

**症状**:
```bash
$ docker logs grafana
mkdir: cannot create directory '/var/lib/grafana/plugins': Permission denied
```

**原因**: Grafana 数据目录权限不正确

**解决方案**:
```bash
# 修复权限
sudo chown -R 472:472 data/grafana
docker-compose restart grafana

# 验证
curl http://localhost:3000/api/health
```

**问题 1.4: Elasticsearch 启动失败**

**症状**:
```bash
$ docker logs elasticsearch
max virtual memory areas vm.max_map_count [65530] is too low
```

**原因**: 系统 vm.max_map_count 设置过低

**解决方案**:
```bash
# 临时设置
sudo sysctl -w vm.max_map_count=262144

# 永久生效
echo "vm.max_map_count=262144" | sudo tee -a /etc/sysctl.conf
sudo sysctl -p

# 重启 Elasticsearch
docker-compose restart elasticsearch
```

#### 2. 网络连接问题

**问题 2.1: 端口被占用**

**症状**:
```bash
Error starting userland proxy: listen tcp 0.0.0.0:9090: bind: address already in use
```

**解决方案**:
```bash
# 查找占用端口的进程
sudo netstat -tlnp | grep 9090
# 或
sudo ss -tlnp | grep 9090

# 停止占用端口的进程
sudo kill <PID>

# 或修改 docker-compose.yml 中的端口映射
# 例如: "9091:9090" 改为使用 9091 端口
```

**问题 2.2: 容器间无法通信**

**症状**: ceph-exporter 无法连接到 ceph-demo

**解决方案**:
```bash
# 检查网络配置
docker network ls
docker network inspect deployments_ceph-network

# 测试容器间连接
docker exec ceph-exporter ping ceph-demo

# 重建网络
docker-compose down
docker-compose up -d
```

**问题 2.3: 防火墙阻止访问**

**症状**: 无法从外部访问服务

**解决方案**:
```bash
# 方式 1: 开放端口
sudo firewall-cmd --permanent --add-port=9128/tcp
sudo firewall-cmd --permanent --add-port=9090/tcp
sudo firewall-cmd --permanent --add-port=3000/tcp
sudo firewall-cmd --reload

# 方式 2: 临时关闭防火墙（仅测试环境）
sudo systemctl stop firewalld

# 验证
curl http://localhost:9128/metrics
```

#### 3. 数据采集问题

**问题 3.1: Prometheus 无法采集数据**

**症状**: Prometheus Targets 页面显示 ceph-exporter 为 DOWN

**解决方案**:
```bash
# 1. 检查 ceph-exporter 是否运行
docker ps | grep ceph-exporter

# 2. 检查 ceph-exporter 端点
curl http://localhost:9128/metrics

# 3. 检查网络连接
docker exec prometheus curl http://ceph-exporter:9128/metrics

# 4. 查看 Prometheus 日志
docker logs prometheus --tail 50

# 5. 重启服务
docker-compose restart ceph-exporter prometheus
```

**问题 3.2: Grafana 显示 "No Data"**

**症状**: Grafana 仪表板显示 "No Data"

**解决方案**:
```bash
# 1. 检查 Prometheus 采集状态
curl http://localhost:9090/api/v1/targets

# 2. 检查数据源配置
# 访问 Grafana → 配置 → 数据源 → Prometheus
# 点击 "保存并测试"

# 3. 检查查询语句
# 在 Grafana 中打开面板编辑模式
# 查看查询语句是否正确

# 4. 手动执行查询
curl -s 'http://localhost:9090/api/v1/query?query=ceph_health_status' | jq '.'
```

**问题 3.3: 指标数据不准确**

**症状**: 指标值异常或不更新

**解决方案**:
```bash
# 1. 检查 Ceph 集群状态
docker exec ceph-demo ceph -s

# 2. 检查 ceph-exporter 日志
docker logs ceph-exporter --tail 100

# 3. 重启 ceph-exporter
docker-compose restart ceph-exporter

# 4. 清理 Prometheus 缓存
docker-compose stop prometheus
rm -rf data/prometheus/wal/*
docker-compose start prometheus
```

#### 4. 告警问题

**问题 4.1: 告警未触发**

**症状**: 满足告警条件但未触发告警

**解决方案**:
```bash
# 1. 检查告警规则语法
docker exec prometheus promtool check rules /etc/prometheus/alert_rules.yml

# 2. 查看告警规则状态
# 访问 http://localhost:9090/alerts

# 3. 检查 Alertmanager 连接
curl http://localhost:9093/api/v2/status

# 4. 重新加载配置
curl -X POST http://localhost:9090/-/reload
```

**问题 4.2: 告警通知未发送**

**症状**: 告警触发但未收到通知

**解决方案**:
```bash
# 1. 检查 Alertmanager 配置
docker exec alertmanager cat /etc/alertmanager/alertmanager.yml

# 2. 查看 Alertmanager 日志
docker logs alertmanager --tail 100

# 3. 测试通知接收器
# 根据配置的接收器类型（email、webhook 等）进行测试

# 4. 检查静默规则
# 访问 http://localhost:9093/#/silences
```

#### 5. 性能问题

**问题 5.1: 内存不足**

**症状**:
```bash
$ docker inspect <container> | grep OOMKilled
"OOMKilled": true
```

**解决方案**:
```bash
# 1. 查看内存使用
docker stats

# 2. 检查系统可用内存
free -h

# 3. 解决方法:
# - 增加系统内存
# - 减少运行的服务数量
# - 调整 docker-compose.yml 中的 mem_limit
# - 使用最小部署模式
./scripts/deploy.sh minimal
```

**问题 5.2: 磁盘空间不足**

**症状**:
```bash
no space left on device
```

**解决方案**:
```bash
# 1. 检查磁盘使用
df -h
du -sh data/*

# 2. 清理 Docker 资源
docker system prune -a

# 3. 清理旧数据
./scripts/deploy.sh clean

# 4. 调整数据保留策略
# 编辑 prometheus.yml:
# --storage.tsdb.retention.time=15d  # 从 30d 改为 15d
```

**问题 5.3: CPU 使用率过高**

**症状**: 服务响应缓慢，CPU 使用率持续高于 80%

**解决方案**:
```bash
# 1. 查看 CPU 使用情况
docker stats

# 2. 识别高 CPU 使用的容器
top -c

# 3. 优化措施:
# - 增加采集间隔（prometheus.yml 中的 scrape_interval）
# - 减少查询频率
# - 优化 PromQL 查询语句
# - 限制容器 CPU 使用（docker-compose.yml 中添加 cpus 限制）
```

### 诊断工具和命令

#### 1. 容器诊断

```bash
# 查看所有容器状态
docker ps -a

# 查看容器详细信息
docker inspect <container-name>

# 查看容器资源使用
docker stats <container-name>

# 进入容器
docker exec -it <container-name> /bin/bash

# 查看容器日志
docker logs <container-name> --tail 100 -f
```

#### 2. 网络诊断

```bash
# 查看网络列表
docker network ls

# 查看网络详情
docker network inspect deployments_ceph-network

# 测试容器间连接
docker exec <container-name> ping <target-container>

# 测试端口连接
docker exec <container-name> telnet <target-container> <port>

# 查看端口监听
docker exec <container-name> netstat -tlnp
```

#### 3. 服务诊断

```bash
# 使用部署脚本诊断
./scripts/deploy.sh diagnose

# 诊断特定服务
./scripts/deploy.sh diagnose ceph-exporter

# 验证部署
./scripts/deploy.sh verify

# 查看服务状态
./scripts/deploy.sh status
```

#### 4. Ceph 诊断

```bash
# 查看集群状态
docker exec ceph-demo ceph -s

# 查看详细健康信息
docker exec ceph-demo ceph health detail

# 查看 OSD 状态
docker exec ceph-demo ceph osd stat
docker exec ceph-demo ceph osd tree

# 查看 Monitor 状态
docker exec ceph-demo ceph mon stat

# 查看 PG 状态
docker exec ceph-demo ceph pg stat
```

### 完整诊断流程

当遇到问题时，按以下顺序进行诊断：

**步骤 1: 检查环境**
```bash
# 检查 Docker
docker --version
sudo systemctl status docker

# 检查 Docker Compose
docker-compose --version

# 检查系统资源
free -h
df -h
```

**步骤 2: 检查容器状态**
```bash
# 查看所有容器
docker ps -a

# 查看失败容器的日志
docker logs <container-name> --tail 100
```

**步骤 3: 检查权限**
```bash
# 检查数据目录权限
ls -la data/

# 修复权限
sudo ./scripts/deploy.sh init
```

**步骤 4: 检查配置**
```bash
# 检查 configs 软链接
ls -la configs

# 检查 Ceph 配置
ls -la data/ceph-demo/config/
```

**步骤 5: 检查网络**
```bash
# 检查网络配置
docker network ls
docker network inspect deployments_ceph-network

# 测试容器间连接
docker exec ceph-exporter ping ceph-demo
```

**步骤 6: 重启服务**
```bash
# 重启单个服务
docker-compose restart <service-name>

# 重启所有服务
docker-compose restart

# 完全重新部署
docker-compose down
sudo ./scripts/deploy.sh full
```

---

## 最佳实践

### 部署最佳实践

#### 1. 首次部署

```bash
# 推荐流程
cd ceph-exporter/deployments

# 1. 检查环境
./scripts/deploy.sh check

# 2. 使用部署脚本（自动处理所有配置）
sudo ./scripts/deploy.sh full

# 3. 等待服务完全启动
sleep 120

# 4. 验证部署
sudo ./scripts/deploy.sh verify

# 5. 检查所有服务状态
docker ps
docker-compose ps
```

#### 2. 生产环境部署

```bash
# 1. 使用最小部署模式
./scripts/deploy.sh minimal

# 2. 配置外部 Ceph 集群连接
# 编辑 configs/ceph-exporter.yaml
# 修改 ceph.config_file 和 ceph.keyring_file

# 3. 配置数据保留策略
# 编辑 prometheus/prometheus.yml
# 设置合适的 retention.time

# 4. 配置告警通知
# 编辑 alertmanager/alertmanager.yml
# 配置 email、webhook 等接收器

# 5. 启用 HTTPS
# 配置反向代理（Nginx、Traefik 等）
```

#### 3. 开发测试环境

```bash
# 使用集成测试模式
./scripts/deploy.sh integration

# 启用调试日志
# 编辑 configs/ceph-exporter.yaml
# 设置 logger.level: "debug"
```

### 监控最佳实践

#### 1. 日常监控

- **主要使用 Grafana**: 完整的中文界面，更好的可视化效果
- **定期检查告警**: 每天查看 Alertmanager 的活跃告警
- **关注关键指标**:
  - 集群健康状态
  - 集群容量使用率
  - OSD 状态和延迟
  - PG 状态

#### 2. 告警配置

- **合理设置告警阈值**: 避免告警疲劳
- **配置告警分组**: 相关告警分组处理
- **设置告警抑制**: 避免重复告警
- **配置多种通知方式**: Email、Webhook、Slack 等

#### 3. 性能优化

- **调整采集间隔**: 根据需求调整 scrape_interval
- **优化查询语句**: 使用高效的 PromQL 查询
- **定期清理数据**: 设置合理的数据保留时间
- **监控资源使用**: 定期检查内存、磁盘使用情况

### 数据管理最佳实践

#### 1. 备份策略

- **定期备份**: 每天自动备份数据
- **多地备份**: 备份到本地和远程服务器
- **测试恢复**: 定期测试备份恢复流程
- **保留策略**: 保留最近 7-30 天的备份

#### 2. 数据清理

- **自动清理**: 配置 Prometheus 自动清理旧数据
- **日志轮转**: 配置 Docker 日志轮转
- **定期检查**: 每周检查磁盘使用情况

#### 3. 数据迁移

```bash
# 迁移流程
# 1. 在源服务器备份数据
tar -czf ceph-exporter-data.tar.gz data/

# 2. 传输到目标服务器
scp ceph-exporter-data.tar.gz target-server:/path/

# 3. 在目标服务器恢复
tar -xzf ceph-exporter-data.tar.gz
sudo chown -R 65534:65534 data/prometheus
sudo chown -R 472:472 data/grafana

# 4. 启动服务
docker-compose up -d
```

### 安全最佳实践

#### 1. 访问控制

- **修改默认密码**: 首次登录后立即修改 Grafana 管理员密码
- **禁用匿名访问**: 禁用 Grafana 匿名访问
- **配置防火墙**: 只开放必要的端口
- **使用 HTTPS**: 生产环境使用 HTTPS 访问

#### 2. 数据安全

- **加密备份**: 备份数据时使用加密
- **权限控制**: 严格控制数据目录权限
- **定期审计**: 定期审计访问日志

#### 3. 网络安全

- **内网部署**: 监控服务部署在内网
- **VPN 访问**: 外部访问通过 VPN
- **反向代理**: 使用反向代理（Nginx）提供额外的安全层

### 维护最佳实践

#### 1. 定期维护

```bash
# 每周维护任务
# 1. 检查服务状态
./scripts/deploy.sh status

# 2. 查看资源使用
docker stats
du -sh data/*

# 3. 检查日志错误
docker-compose logs | grep -i error

# 4. 备份数据
tar -czf backup-$(date +%Y%m%d).tar.gz data/

# 5. 清理旧备份
find /backup -name "backup-*.tar.gz" -mtime +30 -delete
```

#### 2. 更新升级

```bash
# 更新流程
# 1. 备份数据
tar -czf backup-before-update.tar.gz data/

# 2. 拉取最新镜像
docker-compose pull

# 3. 重新创建容器
docker-compose up -d

# 4. 验证更新
./scripts/deploy.sh verify

# 5. 检查日志
docker-compose logs --tail 100
```

#### 3. 故障预防

- **监控资源使用**: 设置资源使用告警
- **定期检查日志**: 查找潜在问题
- **测试恢复流程**: 定期测试备份恢复
- **文档更新**: 及时更新操作文档

---

## 附录

### A. 快速参考

#### 常用 URL

| 功能 | URL |
|------|-----|
| Grafana（中文） | http://localhost:3000 |
| Prometheus 主页 | http://localhost:9090 |
| Prometheus 采集目标 | http://localhost:9090/targets |
| Prometheus 告警规则 | http://localhost:9090/alerts |
| Alertmanager | http://localhost:9093 |
| Ceph Exporter 指标 | http://localhost:9128/metrics |
| Ceph Exporter 健康检查 | http://localhost:9128/health |
| Kibana | http://localhost:5601 |
| Jaeger | http://localhost:16686 |

#### 常用命令

```bash
# 部署管理
./scripts/deploy.sh full          # 完整部署
./scripts/deploy.sh status         # 查看状态
./scripts/deploy.sh logs           # 查看日志
./scripts/deploy.sh verify         # 验证部署
./scripts/deploy.sh diagnose       # 诊断问题
./scripts/deploy.sh fix            # 修复问题
./scripts/deploy.sh stop           # 停止服务
./scripts/deploy.sh clean          # 清理数据

# Docker 管理
docker-compose up -d               # 启动服务
docker-compose down                # 停止服务
docker-compose restart             # 重启服务
docker-compose ps                  # 查看状态
docker-compose logs -f             # 查看日志

# Ceph 管理
docker exec ceph-demo ceph -s      # 查看集群状态
docker exec ceph-demo ceph health detail  # 查看健康详情
docker exec ceph-demo ceph osd stat       # 查看 OSD 状态
docker exec ceph-demo ceph df             # 查看容量使用
```

### B. 相关文档

- [README.md](README.md) - 项目主文档
- [QUICK_START.md](QUICK_START.md) - 快速开始指南
- [DEPLOYMENT_GUIDE.md](DEPLOYMENT_GUIDE.md) - 完整部署指南
- [Prometheus使用指南.md](Prometheus使用指南.md) - Prometheus 详细说明
- [中文界面配置说明.md](中文界面配置说明.md) - 中文界面配置
- [中文可观测性实现指南.md](中文可观测性实现指南.md) - 可观测性实现
- [ceph-exporter/deployments/README.md](ceph-exporter/deployments/README.md) - 部署配置说明
- [ceph-exporter/deployments/TROUBLESHOOTING.md](ceph-exporter/deployments/TROUBLESHOOTING.md) - 故障排查指南
- [ceph-exporter/deployments/DATA_STORAGE.md](ceph-exporter/deployments/DATA_STORAGE.md) - 数据存储说明

### C. 技术支持

如遇到问题，请按以下步骤操作：

1. **查看文档**: 首先查看相关文档和本指南的故障排查部分
2. **运行诊断**: 使用 `./scripts/deploy.sh diagnose` 收集诊断信息
3. **查看日志**: 使用 `./scripts/deploy.sh logs` 查看服务日志
4. **尝试修复**: 使用 `sudo ./scripts/deploy.sh fix` 尝试自动修复
5. **提交 Issue**: 如果问题仍未解决，请提交 Issue 并附上诊断信息

---

**文档版本**: 1.0
**最后更新**: 2026-03-11
**维护者**: Ceph-Exporter 项目团队

---

**结束语**

本文档提供了 Ceph-Exporter 项目的完整操作指南，涵盖了从部署到维护的所有方面。通过本指南，您应该能够：

- ✅ 快速部署和配置 Ceph-Exporter 监控系统
- ✅ 熟练使用各个服务的界面和功能
- ✅ 掌握常用的监控和管理操作
- ✅ 独立排查和解决常见问题
- ✅ 遵循最佳实践进行系统维护

如有任何问题或建议，欢迎反馈！
