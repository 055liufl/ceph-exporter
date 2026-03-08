# ceph-exporter

基于 Go 语言开发的 Ceph 集群 Prometheus 指标导出器。

**环境要求**: CentOS 7 + Docker

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

详细步骤请查看 [快速开始指南](QUICK_START.md)

---

## 📖 文档导航

### 部署相关
- 📘 [快速开始指南](QUICK_START.md) - 5 分钟快速部署
- 📗 [完整部署指南](DEPLOYMENT_GUIDE.md) - 详细的环境准备和故障排查
- 📙 [Docker 镜像加速配置](DOCKER_MIRROR_CONFIGURATION.md) - 国内用户必读
- 📕 [部署配置说明](ceph-exporter/deployments/README.md) - Docker Compose 配置详解
- 📓 [数据存储说明](ceph-exporter/deployments/DATA_STORAGE.md) - 数据目录结构和管理

### 开发相关
- 📔 [项目详细文档](ceph-exporter/README.md) - 架构设计和验收清单
- 📒 [开发指南](ceph-exporter/DEVELOPMENT.md) - 开发环境配置
- 📑 [集成测试文档](ceph-exporter/test/integration/README.md) - 集成测试说明

### 故障排查
- 🔧 [故障排查指南](ceph-exporter/deployments/TROUBLESHOOTING.md) - 常见问题解决方案
- 🛠️ [脚本使用指南](SCRIPTS_GUIDE.zh-CN.md) - 部署脚本详细说明

---

## 📋 部署方式

| 方式 | 命令 | 包含组件 | 资源需求 | 适用场景 |
|------|------|----------|----------|----------|
| **完整监控栈** | `./scripts/deploy.sh full` | Ceph Demo + 监控 + ELK + Jaeger | 4-6GB | 演示、功能测试 ⭐ |
| **集成测试** | `./scripts/deploy.sh integration` | Ceph Demo + 监控 | 2-3GB | 开发测试 |
| **最小监控栈** | `./scripts/deploy.sh minimal` | 监控组件 | 1GB | 生产环境 |

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

- ✅ **7 个 Prometheus 采集器**: Cluster、Pool、OSD、Monitor、Health、MDS、RGW
- ✅ **CGO 集成**: 使用 go-ceph 库直接与 Ceph 通信
- ✅ **完整测试**: 81 个单元测试，100% 通过率，覆盖率 68.1%
- ✅ **容器化部署**: Docker Compose 一键部署
- ✅ **可观测性**: 支持 OpenTelemetry 分布式追踪
- ✅ **插件系统**: 支持自定义插件扩展

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
CGO_ENABLED=1 go build -o build/ceph-exporter ./cmd/ceph-exporter

# 运行测试
CGO_ENABLED=1 go test -v ./internal/...

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
**最后更新**: 2026-03-08
**许可证**: MIT
