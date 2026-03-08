# ceph-exporter

基于 Go 语言开发的 Ceph 集群 Prometheus 指标导出器。

**环境要求**: CentOS 7 + Docker

---

## 🚀 快速开始（5 分钟）

### 1. 安装 Docker

```bash
sudo yum install -y yum-utils
sudo yum-config-manager --add-repo https://mirrors.aliyun.com/docker-ce/linux/centos/docker-ce.repo
sudo yum install -y docker-ce docker-ce-cli containerd.io
sudo systemctl start docker
sudo systemctl enable docker
```

### 2. 安装 Docker Compose

```bash
sudo curl -L "https://get.daocloud.io/docker/compose/releases/download/v2.24.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose
```

### 3. 配置镜像加速（国内必需）

```bash
sudo mkdir -p /etc/docker
sudo tee /etc/docker/daemon.json <<EOF
{
  "registry-mirrors": [
    "https://docker.mirrors.ustc.edu.cn",
    "https://hub-mirror.c.163.com"
  ]
}
EOF
sudo systemctl restart docker
```

### 4. 部署服务

```bash
cd /home/lfl/ceph-exporter/ceph-exporter/deployments
./scripts/deploy.sh full
```

### 5. 访问服务

- ceph-exporter: http://localhost:9128/metrics
- Prometheus: http://localhost:9090
- Grafana: http://localhost:3000 (admin/admin)

---

## 📋 部署方式

### 方式 1: 集成测试环境（推荐用于开发）

```bash
cd ceph-exporter/deployments
docker-compose -f docker-compose-integration-test.yml up -d
```

**包含**: Ceph Demo + ceph-exporter + Prometheus + Grafana
**资源**: 2-3GB 内存，2 CPU

### 方式 2: 完整监控栈（推荐用于演示）

```bash
cd ceph-exporter/deployments
./scripts/deploy.sh full
```

**包含**: Ceph Demo + 监控 + ELK + Jaeger
**资源**: 4-6GB 内存，2-4 CPU

### 方式 3: 最小监控栈（生产环境）

```bash
cd ceph-exporter/deployments
docker-compose up -d
```

**包含**: ceph-exporter + Prometheus + Grafana
**资源**: 1GB 内存，1-2 CPU

---

## 🔧 常用命令

```bash
# 查看服务状态
./scripts/deploy.sh status

# 查看日志
./scripts/deploy.sh logs

# 验证部署
./scripts/deploy.sh verify

# 停止服务
./scripts/deploy.sh stop

# 清理数据
./scripts/clean-volumes.sh
```

---

## 📁 项目结构

```
ceph-exporter/
├── cmd/ceph-exporter/          # 程序入口
├── internal/                   # 核心代码
│   ├── config/                 # 配置管理
│   ├── logger/                 # 日志系统
│   ├── ceph/                   # Ceph 客户端
│   ├── collector/              # 7 个采集器
│   ├── server/                 # HTTP 服务器
│   ├── tracer/                 # 追踪系统
│   └── plugin/                 # 插件系统
├── configs/                    # 配置文件
├── deployments/                # 部署配置
│   ├── scripts/                # 部署脚本
│   └── *.yml                   # Docker Compose 配置
└── test/integration/           # 集成测试
```

---

## 🧪 开发和测试

### 编译项目

```bash
cd ceph-exporter
CGO_ENABLED=1 go build -o build/ceph-exporter ./cmd/ceph-exporter
```

### 运行测试

```bash
# 单元测试
CGO_ENABLED=1 go test -v ./internal/...

# 测试覆盖率
go test -coverprofile=coverage.out ./internal/...
go tool cover -html=coverage.out

# 代码检查
go vet ./...
```

### Pre-commit 设置

```bash
# 安装 pre-commit
pip install pre-commit

# 安装 hooks
pre-commit install

# 手动运行
pre-commit run --all-files
```

---

## 🐛 故障排查

### 容器无法启动

```bash
# 查看日志
docker logs <container-name>

# 检查资源
docker stats
df -h

# 清理资源
docker system prune -a
```

### 内存不足

```bash
# 检查内存
free -h

# 使用最小部署
./scripts/deploy.sh minimal
```

### 镜像拉取失败

```bash
# 检查镜像加速器
docker info | grep -A 5 "Registry Mirrors"

# 手动拉取
docker pull ceph/demo:latest-nautilus
docker pull prom/prometheus:latest
```

### 防火墙问题

```bash
# 临时关闭
sudo systemctl stop firewalld

# 或开放端口
sudo firewall-cmd --permanent --add-port=9128/tcp
sudo firewall-cmd --permanent --add-port=9090/tcp
sudo firewall-cmd --permanent --add-port=3000/tcp
sudo firewall-cmd --reload
```

---

## 📚 核心特性

- ✅ 7 个 Prometheus 采集器（Cluster、Pool、OSD、Monitor、Health、MDS、RGW）
- ✅ 使用 CGO 和 go-ceph 库
- ✅ 完整的单元测试（81 个测试用例，100% 通过率）
- ✅ 测试覆盖率 68.1%，核心模块 >90%
- ✅ Docker 容器化部署
- ✅ 支持 OpenTelemetry 追踪
- ✅ 插件系统支持

---

## 📖 详细文档

项目包含以下详细文档：

- **QUICK_START.md** - 快速开始指南
- **DEPLOYMENT_GUIDE.md** - 完整部署指南（环境准备、故障排查）
- **DOCKER_MIRROR_CONFIGURATION.md** - Docker 镜像加速配置
- **ceph-exporter/README.md** - 项目详细文档（架构、验收清单）
- **ceph-exporter/deployments/README.md** - 部署配置说明
- **ceph-exporter/test/integration/README.md** - 集成测试文档

---

## 🆘 获取帮助

```bash
# 查看脚本帮助
./scripts/deploy.sh help

# 运行诊断
./scripts/diagnose.sh

# 查看容器日志
docker logs ceph-exporter
docker logs prometheus
docker logs grafana
```

---

## 📊 项目状态

- **代码**: 完整，所有核心模块已实现
- **测试**: 单元测试 100% 通过，覆盖率 68.1%
- **部署**: 支持多种部署方式，配置完整
- **文档**: 完整的部署和开发文档

---

**版本**: 1.0
**最后更新**: 2026-03-07
**许可证**: MIT
