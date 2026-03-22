# 配置示例

本目录包含各种配置文件的示例，用于参考和学习。

## 📋 文件列表

### Docker Compose 示例

**docker-compose-elk.yaml**
- 用途: 展示两种 ELK 日志方案的部署
- 包含方案:
  - 方案1: ceph-exporter 直接推送到 Logstash
  - 方案2: 使用 Filebeat Sidecar 收集日志
- 参考文档: `ceph-exporter/docs/ELK-LOGGING-GUIDE.md`

### 日志收集配置

**filebeat.yml**
- 用途: Filebeat 日志收集器配置示例
- 使用场景: 容器日志收集方案
- 说明: 这是可选的日志收集方案，项目默认使用 Logstash Hook 直接推送

## 🎯 使用说明

这些文件是**示例配置**，不是项目运行所必需的。

### 实际使用的配置文件

项目实际使用的配置文件位于：

- **主配置**: `ceph-exporter/configs/ceph-exporter.yaml`
- **Docker Compose**: `ceph-exporter/deployments/docker-compose-*.yml`
- **Logstash**: `ceph-exporter/deployments/logstash/logstash.conf`

### 如何使用示例

1. **学习参考**: 查看示例了解不同的配置方案
2. **复制修改**: 复制示例文件并根据需要修改
3. **测试验证**: 在测试环境中验证配置

## 📚 相关文档

- [ELK 日志指南](../../ceph-exporter/docs/ELK-LOGGING-GUIDE.md)
- [完整操作指南](../../Ceph-Exporter项目完整操作指南.md)

## 💡 提示

如果你需要配置 ELK 日志系统，建议：

1. 使用 `docker-compose-lightweight-full.yml`（已包含完整的 ELK 栈）
2. 参考本目录中的示例了解不同方案
3. 根据实际需求选择合适的日志收集方案
