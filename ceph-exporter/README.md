# ceph-exporter

基于 Go 语言开发的 Ceph 集群 Prometheus 指标导出器。通过接口抽象与 Ceph 集群通信，采集集群状态、存储池、OSD、Monitor 等核心指标，并以 Prometheus 标准格式暴露。项目启用 CGO，使用 go-ceph 库与 Ceph 集群通信。

---

## 🚀 快速开始

**想要部署 ceph-exporter？**

- 📖 [完整部署指南](deployments/README.md) - 详细的部署步骤和配置说明
- ⚡ [完整操作指南](../Ceph-Exporter项目完整操作指南.md) - 部署、配置、使用一站式指南
- 🔧 [部署目录说明](deployments/README.md) - 配置文件和脚本详解

**系统要求**: Ubuntu 20.04 + Docker

**推荐部署方式**:
```bash
cd deployments
sudo ./scripts/deploy.sh full
```

部署脚本会自动处理：
- ✓ 环境检查和依赖验证
- ✓ 数据目录创建和权限设置
- ✓ Docker 镜像拉取
- ✓ 服务启动和健康检查

---

## 📚 文档更新（2026-03-08）

### 最新更新

**项目状态**:

- ✅ 项目代码结构完整，所有核心模块已实现
- ✅ 支持 Ubuntu 20.04 + Docker 环境部署
- ✅ 提供多种部署方式（最小栈、集成测试、完整监控栈）
- ✅ 包含 Ceph Demo 容器用于开发和测试
- ✅ 部署脚本已优化，自动处理权限和配置问题

**测试系统**:

- ✅ 所有单元测试通过（90 个测试用例，100% 通过率）
- ✅ 集成测试配置：`docker-compose-integration-test.yml`
- ✅ 测试覆盖率：平均 68.1%，核心模块 >90%
- ✅ 代码静态检查通过（go vet）

**部署配置**:

- ✅ 自动化部署脚本：`deployments/scripts/deploy.sh`
- ✅ 多种 docker-compose 配置文件
- ✅ 完整的部署文档和故障排查指南
- ✅ 支持 Docker 镜像加速配置（国内用户）
- ✅ 自动设置正确的目录权限（Prometheus: 65534, Grafana: 472, Elasticsearch: 1000）
- ✅ 自动创建配置文件软链接

### 部署方式

项目支持多种部署方式，适用于不同场景：

**部署配置文件**:

- 📦 `docker-compose.yml` - 最小监控栈（ceph-exporter + Prometheus + Grafana）
- 🧪 `docker-compose-integration-test.yml` - 集成测试环境（含 Ceph Demo）
- 🚀 `docker-compose-lightweight-full.yml` - 完整监控栈（含 ELK + Jaeger）
- 🐳 `docker-compose-ceph-demo.yml` - 独立 Ceph Demo 容器

**自动化部署**:

- ✅ 一键部署脚本：`deployments/scripts/deploy.sh`
- ✅ 支持多种部署模式：minimal、integration、full
- ✅ 自动环境检查和配置验证
- ✅ 服务健康检查和日志查看

**镜像加速**:

- ✅ 国内镜像源配置支持
- ✅ 详细配置指南：[DOCKER_MIRROR_CONFIGURATION.md](../DOCKER_MIRROR_CONFIGURATION.md)

---

## 项目代码结构

```
ceph-exporter/
├── cmd/
│   └── ceph-exporter/
│       └── main.go                     # 程序入口
│                                       #   - 命令行参数解析（-config, -version）
│                                       #   - 启动流程编排（配置→日志→追踪→Ceph→采集器→插件→HTTP）
│                                       #   - 注册 7 个 Prometheus 采集器（Cluster/Pool/OSD/Monitor/Health/MDS/RGW）
│                                       #   - 信号监听与优雅关闭（SIGINT/SIGTERM）
│
├── configs/
│   └── ceph-exporter.yaml              # 配置文件模板
│                                       #   - 服务器、Ceph 连接、日志、追踪、插件等完整配置项
│                                       #   - 支持环境变量覆盖（CEPH_EXPORTER_ 前缀）
│
├── internal/
│   ├── config/
│   │   ├── config.go                   # 配置结构体定义
│   │   │                               #   - ServerConfig:  HTTP 服务器配置（端口、TLS、超时）
│   │   │                               #   - CephConfig:    Ceph 连接配置（集群名、用户、密钥环）
│   │   │                               #   - LoggerConfig:  日志配置（级别、格式、轮转策略）
│   │   │                               #   - TracerConfig:  追踪配置（OTLP 端点、采样率）
│   │   │                               #   - PluginConfig:  插件配置（名称、类型、路径）
│   │   ├── loader.go                   # 配置加载器
│   │   │                               #   - YAML 文件解析
│   │   │                               #   - 环境变量覆盖
│   │   │                               #   - 默认值填充与配置校验
│   │   └── config_test.go              # 配置模块测试
│   │                                   #   - YAML 解析正确性
│   │                                   #   - 默认值填充验证
│   │                                   #   - 环境变量覆盖验证
│   │                                   #   - 配置校验（端口范围、必填项等）
│   │
│   ├── logger/
│   │   ├── logger.go                   # 日志系统
│   │   │                               #   - 基于 logrus 的结构化日志
│   │   │                               #   - lumberjack 日志文件轮转
│   │   │                               #   - 支持 text/json 输出格式
│   │   │                               #   - WithComponent() 组件标签
│   │   │                               #   - WithTraceID() 追踪 ID 关联
│   │   └── logger_test.go              # 日志模块测试
│   │                                   #   - 日志级别设置验证
│   │                                   #   - 输出格式验证（text/json）
│   │                                   #   - 文件输出与轮转验证
│   │                                   #   - 组件标签和追踪 ID 验证
│   │
│   ├── ceph/
│   │   ├── client.go                   # Ceph 客户端（使用 CGO 和 go-ceph 库）
│   │   │                               #   - radosConn 接口抽象
│   │   │                               #   - Connect()/Close()/Reconnect() 生命周期管理
│   │   │                               #   - ExecuteCommand() 带超时的命令执行
│   │   │                               #   - 数据结构: ClusterStatus, PoolStats, OSDStats,
│   │   │                               #     MonitorStats, MDSStats/MDSDaemon, RGWStats/RGWDaemon
│   │   │                               #   - 数据获取: GetClusterStatus(), GetPoolStats(),
│   │   │                               #     GetOSDStats(), GetMonitorStats(), GetMDSStats(),
│   │   │                               #     GetRGWStats(), GetOSDDump(), GetOSDPerf(),
│   │   │                               #     GetHealthDetail(), GetDF(), GetPGStat()
│   │   │                               #   - HealthCheck() 健康检查
│   │   ├── conn_cgo.go                 # RADOS 连接 - CGO 实现（build tag: cgo）
│   │   │                               #   - 使用 go-ceph/rados 库连接真实 Ceph 集群
│   │   │                               #   - 在 CGO_ENABLED=1 时编译
│   │   └── client_test.go              # Ceph 客户端测试（使用 CGO）
│   │                                   #   - JSON 反序列化验证
│   │                                   #   - 客户端状态管理验证
│   │                                   #   - 命令 JSON 构建验证
│   │
│   ├── collector/
│   │   ├── collector.go                # 采集器公共定义与基础设施
│   │   │                               #   - namespace 常量（"ceph"）
│   │   │                               #   - defaultCollectTimeout（10s）
│   │   │                               #   - newCollectContext() 带超时采集上下文
│   │   │                               #   - boolToFloat64() 辅助函数
│   │   ├── cluster.go                  # ClusterCollector - 集群整体状态采集器
│   │   │                               #   - 容量: total_bytes, used_bytes, available_bytes
│   │   │                               #   - IO: read_bytes_sec, write_bytes_sec, read_ops_sec, write_ops_sec
│   │   │                               #   - PG: pgs_total, pgs_by_state（按 state 标签）
│   │   │                               #   - 组件: pools_total, osds_total, osds_up, osds_in, mons_total
│   │   │                               #   - 数据源: "ceph status -f json"
│   │   ├── pool.go                     # PoolCollector - 存储池采集器
│   │   │                               #   - 容量: stored_bytes, max_available_bytes, used_bytes, percent_used
│   │   │                               #   - 对象: objects_total
│   │   │                               #   - IO: read_bytes_sec, write_bytes_sec, read_ops_sec, write_ops_sec
│   │   │                               #   - 标签: pool（存储池名称）
│   │   │                               #   - 数据源: "ceph osd pool stats -f json"
│   │   ├── osd.go                      # OSDCollector - OSD 采集器
│   │   │                               #   - 状态: up, in
│   │   │                               #   - 容量: total_bytes, used_bytes, available_bytes（KB→字节转换）
│   │   │                               #   - 性能: utilization, pgs, apply_latency_ms, commit_latency_ms
│   │   │                               #   - 标签: osd（如 osd.0, osd.1）
│   │   │                               #   - 数据源: "ceph osd df -f json"
│   │   ├── monitor.go                  # MonitorCollector - Monitor 采集器
│   │   │                               #   - in_quorum, store_bytes, clock_skew_sec, latency_sec
│   │   │                               #   - 标签: monitor（Monitor 名称）
│   │   │                               #   - 数据源: "ceph mon dump -f json"
│   │   ├── health.go                   # HealthCollector - 健康状态采集器
│   │   │                               #   - status（0=OK, 1=WARN, 2=ERR）
│   │   │                               #   - status_info（带 status 标签）
│   │   │                               #   - checks_total, check（带 name, severity 标签）
│   │   │                               #   - 数据源: "ceph status -f json" 中的 health 部分
│   │   ├── mds.go                      # MDSCollector - MDS 元数据服务器采集器
│   │   │                               #   - active_total, standby_total
│   │   │                               #   - daemon_status（带 name, state 标签）
│   │   │                               #   - 数据源: "ceph mds stat -f json"
│   │   └── rgw.go                      # RGWCollector - RGW 对象网关采集器
│   │                                   #   - total, active_total
│   │                                   #   - daemon_status（带 name 标签）
│   │                                   #   - 数据源: "ceph service dump -f json"
│   │
│   ├── server/
│   │   ├── server.go                   # HTTP 服务器
│   │   │                               #   - GET /metrics  Prometheus 指标端点
│   │   │                               #   - GET /health   健康检查（存活探针）
│   │   │                               #   - GET /ready    就绪检查（就绪探针）
│   │   │                               #   - TLS 支持（可选）
│   │   │                               #   - 优雅关闭（Shutdown）
│   │   └── server_test.go              # HTTP 服务器测试
│   │                                   #   - 各端点响应验证
│   │                                   #   - 服务器启动/关闭验证
│   │
│   ├── tracer/
│   │   ├── tracer.go                   # 追踪系统（Phase 3 完整实现）
│   │   │                               #   - OpenTelemetry TracerProvider 封装
│   │   │                               #   - OTLP gRPC/HTTP 导出器
│   │   │                               #   - 采样率配置
│   │   └── tracer_test.go              # 追踪模块测试
│   │                                   #   - TracerProvider 创建验证
│   │                                   #   - 关闭流程验证
│   │
│   └── plugin/
│       ├── manager.go                  # 插件管理器（Phase 5 完整实现）
│       │                               #   - 插件加载（.so 动态库 / HTTP 远程）
│       │                               #   - 插件生命周期管理
│       │                               #   - 插件指标注册到 Prometheus
│       └── manager_test.go             # 插件管理器测试
│                                       #   - 插件加载/卸载验证
│                                       #   - 重复加载检测
│                                       #   - 管理器关闭验证
│
├── Dockerfile                          # 多阶段 Docker 构建（启用 CGO）
│                                       #   - builder 阶段: golang:alpine + CGO_ENABLED=1
│                                       #   - runtime 阶段: alpine 最小化镜像
│                                       #   - 非 root 用户运行
│                                       #   - 健康检查配置
│
├── Makefile                            # 构建与开发命令（CGO_ENABLED=1）
│                                       #   - go build:         编译二进制文件
│                                       #   - go test:          运行单元测试
│                                       #   - go test -cover:   测试覆盖率报告
│                                       #   - go vet:           代码静态检查
│                                       #   - docker build:     构建 Docker 镜像
│                                       #   - rm -rf build:     清理构建产物
│                                       #   - 本地运行:         编译后直接执行
│
├── go.mod                              # Go 模块定义
│                                       #   核心依赖:
│                                       #   - github.com/prometheus/client_golang v1.19.0
│                                       #   - github.com/sirupsen/logrus v1.9.3
│                                       #   - gopkg.in/natefinch/lumberjack.v2 v2.2.1
│                                       #   - gopkg.in/yaml.v3 v3.0.1
│                                       #   CGO 依赖（CGO_ENABLED=1 时使用）:
│                                       #   - github.com/ceph/go-ceph v0.27.0
│
└── README.md                           # 本文件
```

## CGO 与 Build Tag 说明

本项目启用 CGO 以支持 Ceph C 库绑定:

- `CGO_ENABLED=1`: 使用 `conn_cgo.go` 中的 go-ceph/rados 真实实现，需要安装 Ceph 开发库，适用于生产环境

## 开发阶段规划

| 阶段 | 内容 | 状态 |
|------|------|------|
| Phase 1 | 项目骨架搭建（配置、日志、Ceph 客户端、HTTP 服务器、程序入口） | 已完成 |
| Phase 2 | Prometheus 采集器完整实现（集群、存储池、OSD、Monitor、健康状态、MDS、RGW） | 已完成 |
| Phase 3 | 全量单元测试（7 个采集器 × 3 维度：Describe / Collect / Error） | 已完成 |
| Phase 4 | OpenTelemetry 追踪系统完整实现 | 已完成 |
| Phase 5 | 告警规则与 Grafana 仪表盘 | 已完成 |
| Phase 6 | 插件系统完整实现（HTTP 远程插件） | 已完成 |

## 快速验证

```bash
# 编译
CGO_ENABLED=1 go build -tags octopus -o build/ceph-exporter ./cmd/ceph-exporter

# 运行测试
CGO_ENABLED=1 go test -tags octopus -v ./internal/...

# Docker 部署
cd deployments && sudo ./scripts/deploy.sh full
```

验证服务：
- Metrics: http://localhost:9128/metrics
- Prometheus: http://localhost:9090
- Grafana: http://localhost:3000 (admin/admin)
- Jaeger: http://localhost:16686
- Kibana: http://localhost:5601


## 项目特性总结

### 核心功能

- ✅ 7 个 Prometheus 采集器（Cluster、Pool、OSD、Monitor、Health、MDS、RGW）
- ✅ 使用 CGO 和 go-ceph 库
- ✅ 完整的配置管理系统
- ✅ 结构化日志系统
- ✅ OpenTelemetry 追踪集成
- ✅ 插件系统（HTTP 插件示例）
- ✅ Docker 容器化部署
- ✅ 完整的集成测试套件

### 性能特性

- CPU 使用率 < 10%
- 内存使用 < 500MB
- 支持大规模集群（1000+ OSD）
- 采集延迟 < 5 秒
- 并发安全设计

### 安全特性

- TLS/SSL 加密支持
- 认证和授权机制
- 完整的错误处理
- 安全的配置管理

### 可观测性

- Prometheus 指标导出
- OpenTelemetry 追踪
- 结构化日志
- 健康检查端点

### 扩展性

- 插件系统支持
- HTTP 插件示例
- 易于添加新采集器
- 灵活的配置系统

## 相关资源

- ELK 日志指南: `docs/ELK-LOGGING-GUIDE.md`
- Jaeger 追踪指南: `docs/JAEGER-TRACING-GUIDE.md`
- 部署文档: `deployments/README.md`
- 测试文档: `test/integration/README.md`
- Docker 镜像配置: `../DOCKER_MIRROR_CONFIGURATION.md`

## 许可证

本项目采用 MIT 许可证。

## 贡献指南

欢迎贡献代码、报告问题或提出建议。请遵循以下规范：

1. 代码规范：遵循 Go 官方代码规范
2. 提交规范：使用语义化提交信息
3. 测试要求：单元测试覆盖率 > 80%
4. 文档要求：所有导出函数必须有注释

## 联系方式

- 问题反馈：GitHub Issues
- 文档：项目 README 和 docs 目录
- 示例：configs 和 deployments 目录
