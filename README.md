# ceph-exporter

基于 Go 语言开发的 Ceph 集群 Prometheus 指标导出器。

**环境要求**: Ubuntu 20.04 + Docker

---

## 🚀 快速开始（5 分钟）

```bash
cd ceph-exporter/deployments
./scripts/deploy.sh full
```

部署完成后访问：
- **ceph-exporter**: http://localhost:9128/metrics
- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000 (admin/admin)

详细步骤请查看 [完整操作指南](Ceph-Exporter项目完整操作指南.md)

---

## 📖 文档导航

### 部署相关
- 📘 [完整操作指南](Ceph-Exporter项目完整操作指南.md) - 部署、配置、使用一站式指南
- 📙 [Docker 镜像加速配置](DOCKER_MIRROR_CONFIGURATION.md) - 国内用户必读
- 📕 [部署配置说明](ceph-exporter/deployments/README.md) - Docker Compose 配置详解
- 📓 [数据存储说明](ceph-exporter/deployments/DATA_STORAGE.md) - 数据目录结构和管理
- 🕐 [时区配置说明](ceph-exporter/deployments/TIMEZONE_CONFIGURATION.md) - 容器时区配置详解

### 开发相关
- 📔 [项目详细文档](ceph-exporter/README.md) - 架构设计和验收清单
- 📒 [开发指南](ceph-exporter/DEVELOPMENT.md) - 开发环境配置
- 📑 [集成测试文档](ceph-exporter/test/integration/README.md) - 集成测试说明

### 故障排查
- 🔧 [故障排查指南](ceph-exporter/deployments/TROUBLESHOOTING.md) - 常见问题解决方案
- 🛠️ [部署脚本说明](ceph-exporter/deployments/README.md) - 部署脚本详细说明

---

## 📋 部署方式

| 方式 | 命令 | 包含组件 | 资源需求 | 适用场景 |
|------|------|----------|----------|----------|
| **完整监控栈** | `./scripts/deploy.sh full` | Ceph Demo + 监控 + ELK + Jaeger | 4-6GB | 演示、功能测试、开发 ⭐ |
| **集成测试** | `./scripts/deploy.sh integration` | Ceph Demo + 监控 | 2-3GB | 开发测试、CI/CD |
| **最小监控栈** | `./scripts/deploy.sh minimal` | 监控组件 | 1GB | 生产环境（已有 Ceph） |

**完整监控栈包含**:
- **存储层**: Ceph Demo (单节点 All-in-One 集群)
- **监控层**: ceph-exporter、Prometheus、Grafana、Alertmanager
- **日志层**: Elasticsearch、Logstash、Kibana、Filebeat (ELK Stack)
- **追踪层**: Jaeger (分布式追踪)

详见 [部署配置说明](ceph-exporter/deployments/README.md)

---

## 🔧 常用命令

```bash
cd ceph-exporter/deployments

# 查看服务状态
./scripts/deploy.sh status

# 查看日志
./scripts/deploy.sh logs [service-name]

# 验证部署
./scripts/deploy.sh verify

# 诊断问题
./scripts/deploy.sh diagnose

# 修复部署问题
sudo ./scripts/deploy.sh fix

# 停止服务
./scripts/deploy.sh stop

# 清理数据
./scripts/deploy.sh clean
```

---

## 📁 项目结构

```
ceph-exporter/
├── cmd/ceph-exporter/          # 程序入口
├── internal/                   # 核心代码
│   ├── collector/              # 7 个 Prometheus 采集器
│   ├── ceph/                   # Ceph 客户端封装
│   ├── config/                 # 配置管理
│   ├── logger/                 # 日志系统
│   ├── server/                 # HTTP 服务器
│   ├── tracer/                 # OpenTelemetry 追踪
│   └── plugin/                 # 插件系统
├── configs/                    # 配置文件模板
├── deployments/                # 部署配置
│   ├── data/                   # 数据存储目录（自动创建）
│   ├── scripts/                # 部署和管理脚本
│   └── *.yml                   # Docker Compose 配置
└── test/integration/           # 集成测试
```

---

## 📚 核心特性

### 指标采集
- ✅ **7 个 Prometheus 采集器**: Cluster、Pool、OSD、Monitor、Health、MDS、RGW
- ✅ **CGO 集成**: 使用 go-ceph 库直接与 Ceph 通信
- ✅ **50+ 指标**: 涵盖容量、性能、健康状态等

### 可观测性
- ✅ **结构化日志**: 基于 logrus，支持 JSON/Text 格式
- ✅ **日志轮转**: 自动轮转、压缩、清理
- ✅ **ELK 集成**:
  - Logstash Hook 直接推送日志
  - 支持 TCP/UDP 协议
  - 异步推送，缓冲队列
  - 自动重连机制
- ✅ **分布式追踪**:
  - OpenTelemetry + Jaeger 集成
  - HTTP 请求追踪
  - 追踪 ID 与日志关联
  - 可配置采样率

### 部署和扩展
- ✅ **容器化部署**: Docker Compose 一键部署
- ✅ **多种部署模式**: minimal、integration、full
- ✅ **插件系统**: 支持自定义插件扩展
- ✅ **完整测试**: 90 个单元测试，100% 通过率，覆盖率 68.1%

---

## 🔍 可观测性功能

### 日志系统

**特性**:
- 结构化日志（JSON/Text 格式）
- 多级别日志（trace、debug、info、warn、error、fatal、panic）
- 日志文件轮转（自动压缩、清理）
- 组件标签和追踪 ID 关联

**ELK 集成**:
```yaml
logger:
  enable_elk: true
  logstash_url: "logstash:5000"
  logstash_protocol: "tcp"
```

**日志推送方案**:
1. **直接推送**: 通过 Logstash Hook 直接推送到 Logstash (TCP/UDP)
2. **容器日志**: 输出到 stdout，使用 Filebeat 采集
3. **文件日志**: 写入文件，使用 Filebeat 监控

详见配置文件中的注释说明。

### 分布式追踪

**技术栈**: OpenTelemetry + Jaeger

**特性**:
- HTTP 请求自动追踪
- 追踪 ID 与日志关联
- 可配置采样率
- OTLP HTTP 导出

**配置**:
```yaml
tracer:
  enabled: true
  jaeger_url: "jaeger:4318"
  sample_rate: 1.0
```

**快速启用**:
```bash
cd deployments
./scripts/enable-jaeger-tracing.sh
```

访问 Jaeger UI: http://localhost:16686

### 监控指标

**50+ Prometheus 指标**，包括：
- 集群容量和 IO 吞吐量
- 存储池使用率和性能
- OSD 状态、容量、延迟
- Monitor 和 MDS 状态
- 健康检查详情
- RGW 对象网关状态

访问指标: http://localhost:9128/metrics

---

## 🐛 故障排查

遇到问题？按以下顺序排查：

1. **查看日志**: `./scripts/deploy.sh logs [service-name]`
2. **运行诊断**: `./scripts/deploy.sh diagnose`
3. **修复部署**: `sudo ./scripts/deploy.sh fix`
4. **查看文档**: [故障排查指南](ceph-exporter/deployments/TROUBLESHOOTING.md)

常见问题：
- 权限问题 → 运行 `sudo ./scripts/deploy.sh fix`
- 内存不足 → 使用 `./scripts/deploy.sh minimal`
- 镜像拉取慢 → 配置 [Docker 镜像加速](DOCKER_MIRROR_CONFIGURATION.md)

---

## 🧪 开发和测试

```bash
# 编译项目
cd ceph-exporter
CGO_ENABLED=1 go build -tags octopus -o build/ceph-exporter ./cmd/ceph-exporter

# 运行测试
CGO_ENABLED=1 go test -tags octopus -v ./internal/...

# 测试覆盖率
go test -coverprofile=coverage.out ./internal/...
go tool cover -html=coverage.out

# 代码检查
go vet ./...
golangci-lint run
```

详见 [开发指南](ceph-exporter/DEVELOPMENT.md)

---

## 📊 项目状态

- **代码**: ✅ 完整，所有核心模块已实现
- **测试**: ✅ 单元测试 100% 通过，覆盖率 68.1%
- **部署**: ✅ 支持多种部署方式，自动化脚本完善
- **文档**: ✅ 完整的部署、开发和故障排查文档

---

**版本**: 1.0
**最后更新**: 2026-03-15
**许可证**: MIT
