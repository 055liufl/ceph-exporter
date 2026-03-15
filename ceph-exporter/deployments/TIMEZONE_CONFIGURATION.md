# 容器时区配置说明

**创建日期**: 2026-03-15
**用途**: 说明 Docker 容器时区配置的实现方法和验证步骤

---

## 📋 概述

所有 Docker Compose 配置文件已自动配置宿主机时区挂载，确保容器内的时间与宿主机保持一致。这对于日志记录、监控数据时间戳和告警时间的准确性至关重要。

---

## 🔧 实现方法

### 配置方式

在所有服务的 `volumes` 配置中添加以下两行：

```yaml
volumes:
  - /etc/localtime:/etc/localtime:ro
  - /etc/timezone:/etc/timezone:ro
```

### 配置说明

- **`/etc/localtime`**: 包含系统的本地时区信息（二进制格式）
- **`/etc/timezone`**: 包含时区名称（文本格式，如 `Asia/Shanghai`）
- **`:ro`**: 只读挂载，防止容器修改宿主机时区配置

---

## 📦 已配置的服务

以下所有服务都已配置时区挂载：

### docker-compose.yml（标准监控栈）
- ceph-exporter
- prometheus
- grafana
- alertmanager

### docker-compose-integration-test.yml（集成测试）
- ceph-demo
- ceph-exporter
- prometheus
- grafana

### docker-compose-ceph-demo.yml（Ceph Demo）
- ceph-demo

### docker-compose-lightweight-full.yml（完整栈）
- ceph-demo
- ceph-exporter
- prometheus
- grafana
- alertmanager
- elasticsearch
- logstash
- kibana
- jaeger

---

## ✅ 优点

1. **时间一致性**: 容器时间与宿主机完全同步
2. **日志准确性**: 日志时间戳反映真实时间
3. **监控精确性**: Prometheus 采集的时间序列数据时间准确
4. **告警及时性**: Alertmanager 告警时间正确
5. **无需配置**: 开箱即用，无需设置 TZ 环境变量
6. **自动更新**: 宿主机时区变更后容器自动同步

---

## 🔍 验证方法

### 方法 1: 使用部署脚本验证

```bash
cd ceph-exporter/deployments

# 验证部署（包含时区检查）
./scripts/deploy.sh verify
```

输出示例：
```
[INFO] 验证容器时区配置...
宿主机时区: CST +0800
✓ 容器时区与宿主机一致: CST +0800
```

### 方法 2: 手动验证

```bash
# 检查宿主机时区
date
timedatectl

# 检查容器时区
docker exec ceph-exporter date
docker exec prometheus date
docker exec grafana date

# 比较时区信息
echo "宿主机时区: $(date +"%Z %z")"
echo "ceph-exporter: $(docker exec ceph-exporter date +"%Z %z")"
echo "prometheus: $(docker exec prometheus date +"%Z %z")"
```

### 方法 3: 验证时区文件

```bash
# 检查容器内的时区文件
docker exec ceph-exporter cat /etc/timezone
docker exec ceph-exporter ls -l /etc/localtime
```

---

## 🛠️ 故障排查

### 问题 1: 容器时区与宿主机不一致

**症状**: 容器显示的时间与宿主机不同

**可能原因**:
- 时区文件挂载失败
- 宿主机时区文件不存在
- 容器使用了 TZ 环境变量覆盖

**解决方案**:
```bash
# 1. 检查宿主机时区文件
ls -l /etc/localtime /etc/timezone

# 2. 如果文件不存在，设置时区
sudo timedatectl set-timezone Asia/Shanghai

# 3. 重启容器
docker-compose restart
```

### 问题 2: 时区文件不存在

**症状**: 容器启动失败或时区为 UTC

**解决方案**:
```bash
# CentOS 7 设置时区
sudo timedatectl set-timezone Asia/Shanghai

# 验证时区文件
ls -l /etc/localtime /etc/timezone

# 重新部署
./scripts/deploy.sh full
```

### 问题 3: 容器使用 UTC 时间

**症状**: 尽管挂载了时区文件，容器仍显示 UTC

**可能原因**: 某些容器镜像会优先使用 TZ 环境变量

**解决方案**:
```yaml
# 在 docker-compose.yml 中添加 TZ 环境变量
environment:
  - TZ=Asia/Shanghai
volumes:
  - /etc/localtime:/etc/localtime:ro
  - /etc/timezone:/etc/timezone:ro
```

---

## 📝 常见时区设置

### 中国时区
```bash
sudo timedatectl set-timezone Asia/Shanghai
```

### 美国时区
```bash
# 东部时间
sudo timedatectl set-timezone America/New_York

# 太平洋时间
sudo timedatectl set-timezone America/Los_Angeles
```

### 欧洲时区
```bash
# 伦敦
sudo timedatectl set-timezone Europe/London

# 巴黎
sudo timedatectl set-timezone Europe/Paris
```

### 查看可用时区
```bash
timedatectl list-timezones
```

---

## 🔄 时区变更流程

如果需要更改系统时区：

```bash
# 1. 更改宿主机时区
sudo timedatectl set-timezone Asia/Shanghai

# 2. 验证时区变更
date
timedatectl

# 3. 重启容器以应用新时区
cd ceph-exporter/deployments
docker-compose restart

# 4. 验证容器时区
./scripts/deploy.sh verify
```

---

## 📊 时区对监控的影响

### Prometheus
- **数据采集时间**: 使用容器时区记录采集时间戳
- **查询结果**: 时间范围查询基于容器时区
- **告警评估**: 告警规则的时间判断使用容器时区

### Grafana
- **仪表板时间**: 默认使用浏览器时区，但数据时间戳来自 Prometheus
- **时间选择器**: 可以在 Grafana 设置中配置时区
- **告警通知**: 告警时间使用容器时区

### Elasticsearch
- **日志时间戳**: 使用容器时区解析和存储日志时间
- **索引管理**: 基于时间的索引使用容器时区
- **查询结果**: 时间范围查询基于容器时区

---

## 🔐 安全考虑

1. **只读挂载**: 使用 `:ro` 标志防止容器修改宿主机时区
2. **文件权限**: 时区文件通常是系统文件，权限已正确设置
3. **容器隔离**: 时区挂载不会影响容器的其他隔离特性

---

## 📚 相关文档

- **DEPLOYMENT_GUIDE.md** - 完整部署指南
- **README.md** - 部署目录说明
- **TROUBLESHOOTING.md** - 故障排查指南

---

## 💡 最佳实践

1. **统一时区**: 建议所有服务器使用相同的时区（如 UTC 或 Asia/Shanghai）
2. **NTP 同步**: 确保宿主机启用 NTP 时间同步
3. **定期验证**: 部署后验证容器时区配置
4. **文档记录**: 在部署文档中记录使用的时区
5. **监控告警**: 监控时间偏差，及时发现时区配置问题

---

**版本**: 1.0
**最后更新**: 2026-03-15
