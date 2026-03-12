# 日志切换脚本使用说明

## switch-logging-mode.sh

快速切换 ceph-exporter 日志推送方案的脚本。

### 使用方法

```bash
cd /home/lfl/ceph-exporter/ceph-exporter
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

# 切换到方案2（容器日志收集，推荐）
./deployments/scripts/switch-logging-mode.sh container

# 切换到开发模式
./deployments/scripts/switch-logging-mode.sh dev
```

### 注意事项

1. **自动备份**: 脚本会自动备份配置文件到 `configs/ceph-exporter.yaml.bak`
2. **重启服务**: 切换配置后需要重启 ceph-exporter 服务才能生效
3. **Filebeat**: 使用方案2时，需要确保 Filebeat 正在运行

### 切换后操作

#### 方案1（直接推送）

```bash
# 重启 ceph-exporter
docker-compose restart ceph-exporter

# 验证日志推送
docker logs ceph-exporter | grep -i elk
# 应该看到: "ELK 集成已启用，日志将推送到 tcp://logstash:5044"
```

#### 方案2（容器日志收集）

```bash
# 启动 Filebeat（如果未运行）
docker-compose up -d filebeat-sidecar

# 重启 ceph-exporter
docker-compose restart ceph-exporter

# 验证 Filebeat 状态
docker logs filebeat-sidecar
```

### 配置文件位置

- 主配置: `deployments/configs/ceph-exporter.yaml`
- 备份文件: `deployments/configs/ceph-exporter.yaml.bak`

### 相关文档

- 完整指南: [docs/ELK-LOGGING-GUIDE.md](../../docs/ELK-LOGGING-GUIDE.md)
- 实现总结: [docs/ELK-IMPLEMENTATION-SUMMARY.md](../../docs/ELK-IMPLEMENTATION-SUMMARY.md)
- 快速参考: [docs/ELK-QUICK-REFERENCE.txt](../../docs/ELK-QUICK-REFERENCE.txt)
