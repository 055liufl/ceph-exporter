# 日志切换脚本使用说明

## switch-logging-mode.sh

快速切换 ceph-exporter 日志推送方案的脚本。

### 使用方法

```bash
cd ceph-exporter
./deployments/scripts/switch-logging-mode.sh [mode]
```

### 可用模式

| 模式 | 说明 | 适用场景 |
|------|------|----------|
| `direct` | 方案1: 直接推送到 Logstash (TCP) | 小规模部署、实时性要求高 |
| `direct-udp` | 方案1: 直接推送到 Logstash (UDP) | 对性能要求高、可接受少量日志丢失 |
| `container` | 方案2: 容器日志收集（推荐） | 生产环境、Kubernetes 部署 |
| `file` | 方案3: 文件日志 + Filebeat | 需要日志持久化 |
| `dev` | 开发模式 | 本地开发调试 |
| `show` | 显示当前配置 | 查看当前日志配置 |

### 示例

```bash
# 查看当前配置
./deployments/scripts/switch-logging-mode.sh show

# 切换到方案1（直接推送 TCP）
./deployments/scripts/switch-logging-mode.sh direct

# 切换到方案1（直接推送 UDP）
./deployments/scripts/switch-logging-mode.sh direct-udp

# 切换到方案2（容器日志收集，推荐）
./deployments/scripts/switch-logging-mode.sh container

# 切换到文件日志模式
./deployments/scripts/switch-logging-mode.sh file

# 切换到开发模式
./deployments/scripts/switch-logging-mode.sh dev
```

### 与 deploy.sh 集成

`deploy.sh full` 部署时会交互式选择日志方案，也可以通过 `LOGGING_MODE` 环境变量跳过交互：

```bash
# 交互式选择（部署时提示选择日志方案）
./scripts/deploy.sh full

# 通过环境变量指定日志方案（跳过交互）
LOGGING_MODE=container ./scripts/deploy.sh full     # 容器日志收集（推荐）
LOGGING_MODE=direct ./scripts/deploy.sh full        # 直接推送到 Logstash (TCP)
LOGGING_MODE=direct-udp ./scripts/deploy.sh full    # 直接推送到 Logstash (UDP)
LOGGING_MODE=file ./scripts/deploy.sh full          # 文件日志 + Filebeat
LOGGING_MODE=dev ./scripts/deploy.sh full           # 开发模式
```

### filebeat-sidecar 自动管理

脚本会根据日志方案自动管理 `filebeat-sidecar` 服务：

- **`container` 模式**: 自动启动 `filebeat-sidecar`，通过 Docker socket 采集 ceph-exporter 容器日志
- **`direct` / `direct-udp` 模式**: 自动停止 `filebeat-sidecar`（不需要）
- **`file` / `dev` 模式**: 不涉及 `filebeat-sidecar`

`filebeat-sidecar` 服务定义在 `docker-compose-lightweight-full.yml` 中，使用 `docker.elastic.co/beats/filebeat:7.17.0` 镜像，配置文件位于 `deployments/filebeat/filebeat.yml`。

### 注意事项

1. **自动备份**: 脚本会自动备份配置文件到 `configs/ceph-exporter.yaml.bak`
2. **重启服务**: 切换配置后需要重启 ceph-exporter 服务才能生效
3. **Filebeat**: 使用 `container` 模式时，脚本会自动启动 filebeat-sidecar

### 切换后操作

#### 方案1（直接推送）

```bash
# 重启 ceph-exporter
docker compose -f docker-compose-lightweight-full.yml restart ceph-exporter

# 验证日志推送
docker logs ceph-exporter | grep -i elk
# 应该看到: "ELK 集成已启用，日志将推送到 tcp://logstash:5000"
```

#### 方案2（容器日志收集）

```bash
# filebeat-sidecar 会由脚本自动启动，也可以手动启动
docker compose -f docker-compose-lightweight-full.yml up -d filebeat-sidecar

# 重启 ceph-exporter
docker compose -f docker-compose-lightweight-full.yml restart ceph-exporter

# 验证 Filebeat 状态
docker logs filebeat-sidecar
```

### 配置文件位置

- 主配置: `configs/ceph-exporter.yaml`
- 备份文件: `configs/ceph-exporter.yaml.bak`
- Filebeat 配置: `deployments/filebeat/filebeat.yml`

### 相关文档

- 完整指南: [docs/ELK-LOGGING-GUIDE.md](../../docs/ELK-LOGGING-GUIDE.md)
