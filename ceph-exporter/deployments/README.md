# deployments 目录说明

本目录包含 ceph-exporter 项目的所有部署配置文件和脚本。

---

## 📂 目录结构

```
deployments/
├── docker-compose-*.yml           # Docker Compose 配置文件
├── prometheus/                    # Prometheus 配置
│   ├── prometheus.yml            # 主配置
│   └── alert_rules.yml           # 告警规则
├── alertmanager/                  # Alertmanager 配置
│   └── alertmanager.yml
├── grafana/                       # Grafana 配置
│   └── provisioning/             # 自动配置
├── logstash/                      # Logstash 配置
│   └── logstash.conf
└── scripts/                       # 部署脚本
    ├── deploy.sh                  # 通用部署脚本
    ├── diagnose.sh                # 诊断脚本
    └── verify-deployment.sh       # 验证脚本
```

---

## 🚀 快速开始

### CentOS 7 + Docker 环境

```bash
# 轻量级完整栈（推荐）
docker-compose -f docker-compose-lightweight-full.yml up -d

# 或标准监控栈
docker-compose up -d
```

---

## 📋 Docker Compose 配置文件

| 文件名 | 说明 | 资源需求 | 推荐场景 |
|--------|------|----------|----------|
| `docker-compose-integration-test.yml` | 集成测试配置 | 2-3GB | 开发测试 ⭐ |
| `docker-compose-lightweight-full.yml` | 轻量级完整栈 | 4-6GB | 功能演示 ⭐ |
| `docker-compose.yml` | 标准监控栈 | 1GB | 已有 Ceph |
| `docker-compose-ceph-demo.yml` | Ceph Demo | 1GB | 仅 Ceph |
| `docker-compose-logging.yml` | ELK 日志系统 | 1GB | 日志收集 |
| `docker-compose-tracing.yml` | Jaeger 追踪 | 256MB | 分布式追踪 |
| `docker-compose-full.yml` | 生产级完整 | 8GB+ | 生产环境 |

**注意**: 每个 `.yml` 文件都有对应的 `.zh-CN.yml` 中文备份版本。

---

## 🔧 部署脚本

### deploy.sh

通用部署脚本，适用于 CentOS 7 + Docker 环境。

```bash
# 完整部署
./scripts/deploy.sh full

# 最小部署
./scripts/deploy.sh minimal

# 查看帮助
./scripts/deploy.sh help

# 查看状态
./scripts/deploy.sh status
```

---

## ⚙️ 配置文件说明

### Prometheus 配置

- **prometheus.yml**: Prometheus 主配置，定义抓取目标和规则
- **alert_rules.yml**: 告警规则，包含 Ceph 集群监控告警

### Alertmanager 配置

- **alertmanager.yml**: 告警路由和通知配置

### Grafana 配置

- **provisioning/datasources/**: 数据源自动配置
- **provisioning/dashboards/**: 仪表板自动配置
- **dashboards/**: Grafana 仪表板 JSON 文件

### Logstash 配置

- **logstash.conf**: 日志处理管道配置

---

## 📖 详细文档

如需完整的部署指南，请查看：

- **[统一部署指南](../DEPLOYMENT_GUIDE.md)** - 完整的部署步骤和说明 ⭐
- **[镜像配置指南](../DOCKER_MIRROR_CONFIGURATION.md)** - Docker 镜像加速配置

---

## 🔍 常用命令

```bash
# 查看容器状态
docker-compose ps

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down

# 停止并删除数据
docker-compose down -v

# 重启服务
docker-compose restart
```

---

## 💡 提示

- **不确定用哪个配置？** 查看 [统一部署指南](../DEPLOYMENT_GUIDE.md) 的"部署方式选择"章节
- **CentOS 7 用户**: 使用 localhost 访问所有服务
- **国内用户**: 建议配置 Docker 镜像加速器

---

## 📝 更新日志

- **2026-03-07**: 所有配置文件和文档已同步到最新状态
- **2026-03-07**: 更新为 CentOS 7 + Docker 环境
- **2026-03-02**: 新增轻量级完整部署配置

---

**最后更新**: 2026-03-07
