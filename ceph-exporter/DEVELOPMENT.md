# Ceph Exporter Development Guide

本指南提供 ceph-exporter 项目的开发环境配置、代码规范、测试方法和最佳实践。

---

## 📋 目录

- [环境要求](#环境要求)
- [开发环境配置](#开发环境配置)
- [代码规范](#代码规范)
- [构建和测试](#构建和测试)
- [调试技巧](#调试技巧)
- [CI/CD](#cicd)

---

## 环境要求

### 必需工具

- **Go**: 1.21+ (项目使用 Go 1.21 特性)
- **Docker**: 用于容器化部署和集成测试
- **Docker Compose**: 用于多容器编排
- **Git**: 版本控制
- **Make**: 构建自动化 (可选)

### 可选工具

- **pre-commit**: 代码质量检查 (强烈推荐)
- **golangci-lint**: Go 代码静态分析
- **goimports**: Go 代码格式化和导入管理

### Ceph 开发库

项目使用 CGO 与 Ceph 集群通信，需要安装 Ceph 15.x (Octopus) 开发库：

```bash
# Ubuntu 20.04 (Focal) - 默认仓库已包含 Ceph 15.x (Octopus)
sudo apt-get install -y librados-dev librbd-dev

# CentOS 7 (需要配置 Ceph Octopus 仓库，不再是主要开发环境)
sudo yum install -y librados-devel librbd-devel
```

---

## 开发环境配置

### 1. 克隆项目

```bash
git clone <repository-url>
cd ceph-exporter
```

### 2. 安装 Go 依赖

```bash
cd ceph-exporter
go mod download
go mod tidy
```

### 3. 配置 Pre-commit (推荐)

Pre-commit 会在提交前自动运行代码检查和格式化：

```bash
# 安装 pre-commit
pip install pre-commit

# 安装 git hooks
pre-commit install
pre-commit install --install-hooks

# 手动运行所有 hooks
pre-commit run --all-files
```

### 4. 安装开发工具

```bash
# goimports (代码格式化和导入管理)
go install golang.org/x/tools/cmd/goimports@latest

# golangci-lint (代码静态分析)
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# 确保工具在 PATH 中
export PATH=$PATH:$(go env GOPATH)/bin
```

---

## 代码规范

### Pre-commit Hooks

项目配置了以下 pre-commit hooks：

**通用检查**:
- `trailing-whitespace`: 删除行尾空格
- `end-of-file-fixer`: 确保文件以换行符结尾
- `check-yaml`: 验证 YAML 文件语法
- `check-json`: 验证 JSON 文件语法
- `check-added-large-files`: 防止提交大文件

**Go 代码检查**:
- `gofmt`: Go 代码格式化
- `goimports`: 导入语句管理
- `go vet`: Go 代码静态检查
- `golangci-lint`: 综合代码质量检查

### 代码风格

遵循 Go 官方代码规范：

- 使用 `gofmt` 格式化代码
- 使用 `goimports` 管理导入
- 导出的函数和类型必须有注释
- 错误处理要明确，不要忽略错误
- 使用有意义的变量名和函数名

### 提交规范

使用语义化提交信息：

```
<type>(<scope>): <subject>

<body>

<footer>
```

**类型 (type)**:
- `feat`: 新功能
- `fix`: 修复 bug
- `docs`: 文档更新
- `style`: 代码格式调整
- `refactor`: 代码重构
- `test`: 测试相关
- `chore`: 构建/工具相关

**示例**:
```
feat(collector): 添加 RGW 采集器

实现 RADOS Gateway 指标采集，包括：
- RGW 守护进程状态
- 对象存储操作统计
- 请求延迟指标

Closes #123
```

---

## 构建和测试

### 编译项目

```bash
# 基本编译 (启用 CGO，需要 -tags octopus)
cd ceph-exporter
CGO_ENABLED=1 go build -tags octopus -o build/ceph-exporter ./cmd/ceph-exporter

# 带版本信息编译
CGO_ENABLED=1 go build -tags octopus -v \
  -ldflags "-X main.version=dev -X main.buildTime=$(date -u '+%Y-%m-%d_%H:%M:%S') -X main.gitCommit=$(git rev-parse --short HEAD)" \
  -o build/ceph-exporter ./cmd/ceph-exporter

# 使用 Makefile
make build
```

### 运行单元测试

```bash
# 运行所有测试
CGO_ENABLED=1 go test -tags octopus -v ./internal/...

# 运行特定包的测试
CGO_ENABLED=1 go test -tags octopus -v ./internal/collector/...

# 运行单个测试
CGO_ENABLED=1 go test -tags octopus -v -run TestClusterCollector ./internal/collector/

# 使用 Makefile
make test
```

### 测试覆盖率

```bash
# 生成覆盖率报告
mkdir -p build
CGO_ENABLED=1 go test -tags octopus -v -coverprofile=build/coverage.out -covermode=atomic ./internal/...

# 查看覆盖率统计
go tool cover -func=build/coverage.out

# 生成 HTML 报告
go tool cover -html=build/coverage.out -o build/coverage.html

# 使用 Makefile
make coverage
```

### 代码静态检查

```bash
# go vet
go vet ./...

# golangci-lint
golangci-lint run --config ../.golangci.yml

# 使用 Makefile
make lint
```

### 集成测试

```bash
# 进入集成测试目录
cd test/integration

# 运行集成测试
CGO_ENABLED=1 go test -tags octopus -v -timeout 30m

# 或使用脚本
./run-integration-tests.sh

# 使用 Makefile (从项目根目录)
make test-integration
```

---

## 调试技巧

### 本地运行

```bash
# 编译
CGO_ENABLED=1 go build -tags octopus -o build/ceph-exporter ./cmd/ceph-exporter

# 运行 (需要 Ceph 集群)
./build/ceph-exporter -config configs/ceph-exporter.yaml

# 查看版本信息
./build/ceph-exporter -version
```

### 使用 Docker Compose 调试

```bash
# 启动完整栈 (包含 Ceph Demo)
cd deployments
./scripts/deploy.sh full

# 查看日志
docker compose logs -f ceph-exporter

# 进入容器调试
docker exec -it ceph-exporter sh

# 测试指标端点
curl http://localhost:9128/metrics
curl http://localhost:9128/health
curl http://localhost:9128/ready
```

### 调试日志配置

修改 `configs/ceph-exporter.yaml` 中的日志级别：

```yaml
logger:
  level: "debug"    # 设置为 debug 查看详细日志
  format: "json"    # 或 "text" 查看纯文本日志
  output: "stdout"  # 输出到标准输出
```

### 追踪调试

启用 Jaeger 追踪：

```bash
# 使用脚本快速启用
cd deployments
./scripts/enable-jaeger-tracing.sh

# 访问 Jaeger UI
open http://localhost:16686
```

### 性能分析

```bash
# 启用 pprof (在 main.go 中添加)
import _ "net/http/pprof"

# 访问性能分析端点
go tool pprof http://localhost:9128/debug/pprof/profile
go tool pprof http://localhost:9128/debug/pprof/heap
```

---

## CI/CD

### GitHub Actions 工作流

项目包含以下 CI 工作流：

#### 1. Pre-commit (`pre-commit.yml`)

- **触发**: 每次 push 和 PR 到 main/develop 分支
- **功能**: 运行所有 pre-commit hooks
- **检查**: 代码格式、静态分析、YAML/JSON 验证

#### 2. CI (`ci.yml`)

- **触发**: 每次 push 和 PR
- **功能**:
  - 在 Go 1.21 和 1.22 上构建和测试
  - 运行 golangci-lint
  - 上传测试覆盖率到 Codecov
- **矩阵测试**: 多个 Go 版本并行测试

#### 3. Integration Tests (`integration-test.yml`)

- **触发**: 每次 push 和 PR
- **功能**: 使用 Docker Compose 运行集成测试
- **环境**: 完整的 Ceph Demo + 监控栈

### 本地运行 CI 检查

```bash
# 运行 pre-commit 检查
pre-commit run --all-files

# 运行单元测试
make test

# 运行代码检查
make lint

# 运行集成测试
make test-integration
```

---

## 项目结构

```
ceph-exporter/
├── cmd/ceph-exporter/          # 程序入口
├── internal/                   # 核心代码
│   ├── collector/              # 7 个 Prometheus 采集器
│   ├── ceph/                   # Ceph 客户端封装
│   ├── config/                 # 配置管理
│   ├── logger/                 # 日志系统 (含 Logstash Hook)
│   ├── server/                 # HTTP 服务器
│   ├── tracer/                 # OpenTelemetry 追踪
│   └── plugin/                 # 插件系统
├── configs/                    # 配置文件模板
├── deployments/                # 部署配置
│   ├── scripts/                # 部署脚本
│   └── *.yml                   # Docker Compose 配置
├── test/integration/           # 集成测试
├── Dockerfile                  # Docker 镜像构建
├── Makefile                    # 构建自动化
└── go.mod                      # Go 模块定义
```

---

## 常见问题

### 1. CGO 编译错误

**问题**: `fatal error: rados/librados.h: No such file or directory`

**解决**:
```bash
# 安装 Ceph 开发库
sudo apt-get install -y librados-dev librbd-dev
```

### 2. 测试失败

**问题**: 单元测试失败

**解决**:
```bash
# 确保启用 CGO 和 octopus 构建标签
CGO_ENABLED=1 go test -tags octopus -v ./internal/...

# 清理缓存
go clean -testcache
```

### 3. Docker 构建失败

**问题**: Docker 镜像构建失败

**解决**:
```bash
# 检查 Dockerfile
docker build -t ceph-exporter:dev .

# 查看构建日志
docker build --progress=plain -t ceph-exporter:dev .
```

---

## 最佳实践

### 开发流程

1. 创建功能分支: `git checkout -b feature/xxx`
2. 编写代码和测试
3. 运行 pre-commit: `pre-commit run --all-files`
4. 提交代码: `git commit -m "feat: xxx"`
5. 推送分支: `git push origin feature/xxx`
6. 创建 Pull Request

### 测试策略

- 每个新功能必须有单元测试
- 测试覆盖率目标: >80%
- 关键路径必须有集成测试
- 使用 mock 隔离外部依赖

### 代码审查

- 所有代码必须经过 PR 审查
- 至少一个 reviewer 批准
- CI 检查必须通过
- 代码覆盖率不能降低

---

**最后更新**: 2026-03-15
