# deployments 目录说明

本目录包含 ceph-exporter 项目的所有部署配置文件和脚本。

---

## 📂 目录结构

```
deployments/
├── docker-compose-*.yml           # Docker Compose 配置文件
├── data/                          # 数据存储目录（自动创建）
│   ├── ceph-demo/                # Ceph 集群数据
│   ├── prometheus/               # Prometheus 时序数据
│   ├── grafana/                  # Grafana 仪表板
│   ├── alertmanager/             # Alertmanager 告警
│   ├── elasticsearch/            # Elasticsearch 索引
│   └── test/                     # 测试环境数据
├── prometheus/                    # Prometheus 配置
│   ├── prometheus.yml            # 主配置
│   └── alert_rules.yml           # 告警规则
├── alertmanager/                  # Alertmanager 配置
│   └── alertmanager.yml
├── grafana/                       # Grafana 配置
│   └── provisioning/             # 自动配置
├── logstash/                      # Logstash 配置
│   └── logstash.conf
└── scripts/                       # 部署和管理脚本
    ├── deploy.sh                  # 主部署脚本（支持 full/minimal/integration 等）
    ├── diagnose.sh                # 诊断脚本（检查服务状态和配置）
    ├── fix-deployment.sh          # 修复脚本（权限、配置等问题）
    ├── verify-deployment.sh       # 验证脚本（健康检查）
    ├── test-ceph-demo.sh          # Ceph Demo 测试脚本
    ├── deploy-full-stack.sh       # 完整栈部署脚本
    └── clean-volumes.sh           # 数据清理脚本
```

---

## 🚀 快速开始

### CentOS 7 + Docker 环境

```bash
# 轻量级完整栈（推荐）
# 会自动初始化 ./data/ 目录并设置正确的权限
./scripts/deploy.sh full

# 或手动部署
# 1. 初始化数据目录（会自动设置权限和创建软链接）
./scripts/deploy.sh init

# 2. 启动服务
docker-compose -f docker-compose-lightweight-full.yml up -d

# 3. 验证部署
./scripts/deploy.sh verify

# 或标准监控栈
docker-compose up -d
```

**数据存储**: 所有服务数据存储在 `./data/` 目录，详见 [DATA_STORAGE.md](DATA_STORAGE.md)。

**重要提示**:
- 部署脚本会自动设置正确的目录权限（Prometheus: 65534, Grafana: 472, Elasticsearch: 1000）
- 会自动创建 `configs` 软链接指向 `../configs` 目录
- 首次部署建议使用 `./scripts/deploy.sh full` 以确保所有配置正确

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

### deploy.sh - 主部署脚本

通用部署脚本，适用于 CentOS 7 + Docker 环境。

```bash
# 环境检查
./scripts/deploy.sh check

# 初始化数据目录
./scripts/deploy.sh init

# 完整部署（推荐）
./scripts/deploy.sh full

# 集成测试环境
./scripts/deploy.sh integration

# 最小部署
./scripts/deploy.sh minimal

# 查看状态
./scripts/deploy.sh status

# 查看日志
./scripts/deploy.sh logs [service-name]

# 验证部署
./scripts/deploy.sh verify

# 诊断问题
./scripts/deploy.sh diagnose [service-name]

# 修复部署问题
./scripts/deploy.sh fix

# 停止服务
./scripts/deploy.sh stop

# 清理数据
./scripts/deploy.sh clean

# 查看帮助
./scripts/deploy.sh help
```

### fix-deployment.sh - 修复脚本

修复常见的部署问题（权限、配置等）。

```bash
# 修复所有问题（需要 root 权限）
sudo ./scripts/fix-deployment.sh
```

**修复内容**：
- ✅ Prometheus 数据目录权限（65534:65534）
- ✅ Grafana 数据目录权限（472:472）
- ✅ Elasticsearch 数据目录权限（1000:1000）
- ✅ Ceph keyring 文件权限（644）
- ✅ configs 目录软链接
- ✅ vm.max_map_count 系统参数
- ✅ 重启失败的服务

### diagnose.sh - 诊断脚本

诊断服务状态和配置问题。

```bash
# 诊断所有服务
./scripts/diagnose.sh

# 诊断特定服务
./scripts/diagnose.sh ceph-exporter
./scripts/diagnose.sh prometheus
```

### verify-deployment.sh - 验证脚本

验证服务健康状态。

```bash
./scripts/verify-deployment.sh
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
- **[数据存储说明](DATA_STORAGE.md)** - 数据目录结构和管理 ⭐
- **[故障排查指南](TROUBLESHOOTING.md)** - 常见问题和解决方案 ⭐

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
./scripts/deploy.sh clean

# 重启服务
docker-compose restart

# 备份数据
tar -czf backup-$(date +%Y%m%d).tar.gz data/

# 查看数据占用
du -sh data/*
```

---

## 💡 提示

- **不确定用哪个配置？** 查看 [统一部署指南](../DEPLOYMENT_GUIDE.md) 的"部署方式选择"章节
- **CentOS 7 用户**: 使用 localhost 访问所有服务
- **国内用户**: 建议配置 Docker 镜像加速器

---

## 🐛 常见问题

### 1. Prometheus 不断重启

**症状**: `docker ps` 显示 prometheus 状态为 `Restarting`

**原因**: Prometheus 数据目录权限不正确

**解决方案**:
```bash
# 方法 1: 使用部署脚本自动修复
sudo ./scripts/deploy.sh init

# 方法 2: 手动修复权限
sudo chown -R 65534:65534 data/prometheus
docker-compose restart prometheus
```

### 2. Ceph-Exporter 连接失败

**症状**: 日志显示 `rados: ret=-13, Permission denied`

**原因**:
- configs 目录软链接不存在
- Ceph keyring 文件权限不正确

**解决方案**:
```bash
# 创建 configs 软链接
cd deployments
ln -s ../configs configs

# 修复 keyring 权限（等待 ceph-demo 启动后）
sudo chmod 644 data/ceph-demo/config/ceph.client.admin.keyring

# 重启 ceph-exporter
docker-compose restart ceph-exporter
```

### 3. Ceph-Demo 验证失败

**症状**: `./scripts/deploy.sh verify` 显示 ceph-demo 无法访问

**原因**: RGW 根路径返回 HTTP 404 是正常行为

**解决方案**:
- 这不是错误！RGW 服务正常运行时根路径会返回 404
- 使用最新版本的 deploy.sh 脚本，已修复验证逻辑
- 手动验证: `curl -v http://localhost:8080` 返回 404 表示服务正常

### 4. 首次部署建议

为避免上述问题，首次部署请按以下步骤操作：

```bash
# 1. 进入部署目录
cd ceph-exporter/deployments

# 2. 使用部署脚本（推荐）
sudo ./scripts/deploy.sh full

# 3. 等待服务启动（约 2-3 分钟）
sleep 120

# 4. 验证部署
sudo ./scripts/deploy.sh verify
```

部署脚本会自动处理：
- ✓ 创建数据目录
- ✓ 设置正确的权限（Prometheus: 65534, Grafana: 472, Elasticsearch: 1000）
- ✓ 创建 configs 软链接
- ✓ 拉取镜像
- ✓ 启动所有服务

---

## 📝 更新日志

- **2026-03-08**:
  - 改用绑定挂载，数据存储在 ./data/ 目录
  - 修复 Prometheus 权限问题（需要 UID 65534）
  - 修复 ceph-exporter 配置路径问题（自动创建 configs 软链接）
  - 修复 ceph-demo 验证逻辑（RGW 返回 404 是正常的）
  - 更新部署脚本，自动处理所有权限和配置问题
- **2026-03-07**: 所有配置文件和文档已同步到最新状态
- **2026-03-07**: 更新为 CentOS 7 + Docker 环境
- **2026-03-02**: 新增轻量级完整部署配置

---

**最后更新**: 2026-03-08
