# 脚本文件详细说明 - 中文注释版本

**创建日期**: 2026-03-07
**最后更新**: 2026-03-09
**用途**: 帮助理解所有部署和维护脚本的功能和用法

---

## 📋 脚本文件清单

### 部署脚本目录: `ceph-exporter/deployments/scripts/`

| 脚本文件 | 中文注释版本 | 说明 |
|---------|-------------|------|
| deploy.sh | deploy.sh.zh-CN | 主部署脚本（最重要） |
| fix-deployment.sh | fix-deployment.sh.zh-CN | 部署问题修复脚本 ⭐ |
| diagnose.sh | diagnose.sh.zh-CN | 诊断脚本 |
| verify-deployment.sh | verify-deployment.sh.zh-CN | 部署验证脚本 |
| clean-volumes.sh | clean-volumes.sh.zh-CN | 数据卷清理脚本 |
| deploy-full-stack.sh | deploy-full-stack.sh.zh-CN | 完整栈部署脚本 |
| test-ceph-demo.sh | test-ceph-demo.sh.zh-CN | Ceph Demo 测试脚本 |

### 其他脚本

| 脚本文件 | 中文注释版本 | 说明 |
|---------|-------------|------|
| scripts/fix-precommit.sh | scripts/fix-precommit.sh.zh-CN | Pre-commit 修复脚本 |
| ceph-exporter/test/integration/run-integration-tests.sh | run-integration-tests.sh.zh-CN | 集成测试脚本 |
| install_go.sh | install_go.sh.zh-CN | Go 环境安装脚本 |
| install_ceph.sh | install_ceph.sh.zh-CN | Ceph 开发库安装脚本 |

---

## 🔍 脚本详细说明

### 1. deploy.sh - 主部署脚本 ⭐⭐⭐

**重要程度**: ⭐⭐⭐ 最重要

**功能**:
- 完整的自动化部署解决方案
- 环境检查和配置
- 支持多种部署模式

**主要功能模块**:

#### 1.1 环境检查模块
```bash
# 检查项目:
- 操作系统版本（CentOS 7）
- Docker 安装和版本
- Docker Compose 安装和版本
- 系统资源（内存、CPU、磁盘）
- 防火墙状态
- SELinux 状态
```

#### 1.2 环境配置模块
```bash
# 配置项目:
- Docker 镜像加速器（国内镜像源）
- 防火墙规则（开放必需端口）
- SELinux 设置（临时禁用）
```

#### 1.3 镜像管理模块
```bash
# 管理的镜像:
- ceph/demo:latest-nautilus          # Ceph Demo 集群
- prom/prometheus:latest             # Prometheus 监控
- grafana/grafana:latest             # Grafana 可视化
- prom/alertmanager:latest           # Alertmanager 告警
- elasticsearch:7.17.0               # Elasticsearch 日志
- logstash:7.17.0                    # Logstash 日志处理
- kibana:7.17.0                      # Kibana 日志可视化
- jaegertracing/all-in-one:1.35      # Jaeger 追踪
```

#### 1.4 部署模式
```bash
# 模式 1: minimal - 最小监控栈
资源需求: 1GB 内存, 1-2 CPU
包含组件: ceph-exporter + Prometheus + Grafana + Alertmanager
适用场景: 生产环境（连接真实 Ceph 集群）

# 模式 2: integration - 集成测试环境
资源需求: 2-3GB 内存, 2 CPU
包含组件: Ceph Demo + ceph-exporter + Prometheus + Grafana
适用场景: 开发和测试

# 模式 3: full - 完整监控栈（推荐）
资源需求: 4-6GB 内存, 2-4 CPU
包含组件: Ceph Demo + 监控 + ELK + Jaeger
适用场景: 演示和功能测试
```

#### 1.5 服务管理功能
```bash
# 可用命令:
./deploy.sh check           # 检查环境
./deploy.sh mirror          # 配置镜像加速
./deploy.sh pull            # 预拉取镜像
./deploy.sh init            # 初始化数据目录（包含时区配置说明）
./deploy.sh minimal         # 部署最小栈
./deploy.sh integration     # 部署集成测试
./deploy.sh full            # 部署完整栈
./deploy.sh status          # 查看状态
./deploy.sh logs [service]  # 查看日志
./deploy.sh verify          # 验证部署（包含时区验证）
./deploy.sh stop            # 停止服务
./deploy.sh clean           # 清理数据
```

**时区配置**: 所有部署模式都会自动挂载宿主机时区（`/etc/localtime` 和 `/etc/timezone`），确保容器时间与宿主机一致。

#### 1.6 关键配置项说明

**Docker 镜像加速配置** (`/etc/docker/daemon.json`):
```json
{
  "registry-mirrors": [
    "https://docker.mirrors.ustc.edu.cn",    // 中科大镜像（推荐）
    "https://hub-mirror.c.163.com",          // 网易镜像
    "https://mirror.ccs.tencentyun.com"      // 腾讯云镜像
  ],
  "log-driver": "json-file",                 // 日志驱动类型
  "log-opts": {
    "max-size": "100m",                      // 单个日志文件最大 100MB
    "max-file": "3"                          // 最多保留 3 个日志文件
  },
  "storage-driver": "overlay2"               // 存储驱动（推荐）
}
```

**防火墙端口配置**:
```bash
9128  # ceph-exporter 指标端口
9090  # Prometheus Web UI
3000  # Grafana Web UI
9093  # Alertmanager Web UI
9200  # Elasticsearch API
5601  # Kibana Web UI
16686 # Jaeger UI
8080  # Ceph Dashboard
```

---

### 2. fix-deployment.sh - 部署问题修复脚本 ⭐⭐⭐

**重要程度**: ⭐⭐⭐ 非常重要

**功能**:
- 自动修复常见的部署问题
- 修复目录权限问题
- 修复系统配置问题
- 重启失败的服务

**主要修复内容**:

#### 2.1 权限修复
```bash
# Prometheus 数据目录权限
- 目标权限: 65534:65534 (nobody 用户)
- 修复目录: data/prometheus
- 常见错误: "permission denied" 导致 Prometheus 不断重启

# Grafana 数据目录权限
- 目标权限: 472:472 (Grafana 容器用户)
- 修复目录: data/grafana
- 常见错误: 无法保存仪表板和配置

# Elasticsearch 数据目录权限
- 目标权限: 1000:1000 (Elasticsearch 容器用户)
- 修复目录: data/elasticsearch
- 常见错误: 无法启动，数据目录访问被拒绝
```

#### 2.2 配置修复
```bash
# configs 目录软链接
- 检查并创建 configs -> ../configs 软链接
- 解决配置文件找不到的问题

# Ceph keyring 文件权限
- 修改为 644 权限，允许容器读取
- 修复文件:
  - ceph.client.admin.keyring
  - ceph.mon.keyring
```

#### 2.3 系统参数修复
```bash
# vm.max_map_count 设置
- 当前值检查: sysctl -n vm.max_map_count
- 需要值: 262144
- 作用: Elasticsearch 需要较大的虚拟内存映射区域
- 修复方式:
  - 临时: sysctl -w vm.max_map_count=262144
  - 永久: 写入 /etc/sysctl.conf
```

#### 2.4 服务重启
```bash
# 检测并重启失败的服务
- 检查 Prometheus 状态（Restarting）
- 检查 ceph-exporter 状态（Restarting）
- 自动重启失败的服务
```

**使用方法**:
```bash
# 必须使用 root 权限运行
sudo ./scripts/fix-deployment.sh

# 或通过 deploy.sh 调用
./scripts/deploy.sh fix
```

**使用场景**:
1. **首次部署失败** - 权限问题导致服务无法启动
2. **Prometheus 不断重启** - 数据目录权限不正确
3. **Grafana 无法保存配置** - 目录权限问题
4. **Elasticsearch 启动失败** - vm.max_map_count 太小
5. **ceph-exporter 无法连接** - keyring 文件权限问题
6. **配置文件找不到** - configs 软链接缺失

**执行流程**:
```bash
步骤 1: 检查 root 权限
步骤 2: 修复 Prometheus 权限 (65534:65534)
步骤 3: 修复 Grafana 权限 (472:472)
步骤 4: 修复 Elasticsearch 权限 (1000:1000)
步骤 5: 检查/创建 configs 软链接
步骤 6: 修复 Ceph keyring 权限 (644)
步骤 7: 检查/设置 vm.max_map_count (262144)
步骤 8: 重启失败的服务
步骤 9: 等待 30 秒让服务启动
步骤 10: 显示服务状态
```

**注意事项**:
- ⚠️ 必须使用 `sudo` 运行
- ⚠️ 会修改系统参数（vm.max_map_count）
- ⚠️ 会重启失败的服务
- ✅ 不会删除任何数据
- ✅ 可以重复运行，不会造成问题

**常见问题解决**:

**问题 1: Prometheus 不断重启**
```bash
# 症状
docker ps  # 显示 prometheus 状态为 Restarting

# 原因
数据目录权限不正确，Prometheus 以 UID 65534 运行

# 解决
sudo ./scripts/fix-deployment.sh
# 或
sudo chown -R 65534:65534 data/prometheus
docker-compose restart prometheus
```

**问题 2: Elasticsearch 无法启动**
```bash
# 症状
docker logs elasticsearch  # 显示 vm.max_map_count 错误

# 原因
系统默认值 65530 太小，Elasticsearch 需要至少 262144

# 解决
sudo ./scripts/fix-deployment.sh
# 或
sudo sysctl -w vm.max_map_count=262144
```

**问题 3: ceph-exporter 无法读取 keyring**
```bash
# 症状
docker logs ceph-exporter  # 显示 permission denied

# 原因
keyring 文件权限过于严格（600），容器无法读取

# 解决
sudo ./scripts/fix-deployment.sh
# 或
sudo chmod 644 data/ceph-demo/config/*.keyring
```

---

### 3. clean-volumes.sh - 数据卷清理脚本 ⭐⭐

**重要程度**: ⭐⭐ 重要

**功能**:
- 彻底清理 Ceph Demo 的数据卷
- 解决数据损坏或配置错误问题
- 重新初始化 Ceph 环境

**使用场景**:
1. Ceph Demo 无法正常启动
2. Ceph 集群状态异常（HEALTH_ERR）
3. 需要完全重置 Ceph 环境
4. 数据卷中有损坏的文件

**工作流程**:
```bash
步骤 1: 停止所有服务
  - docker-compose down

步骤 2: 清理 ceph-demo-data 数据卷
  - 使用临时 Alpine 容器挂载卷
  - 删除所有文件（包括隐藏文件）

步骤 3: 清理 ceph-demo-config 配置卷
  - 删除所有配置文件

步骤 4: 验证卷已清空
  - 列出卷内容确认为空

步骤 5: 启动 Ceph Demo
  - docker-compose up -d

步骤 6: 等待 5 分钟
  - Ceph 需要时间初始化

步骤 7-9: 检查和验证
  - 查看容器状态
  - 查看日志
  - 测试 ceph -s 命令
```

**关键技术点**:
```bash
# 清理数据目录
cd deployments
rm -rf data/ceph-demo/*

# 为什么不使用 Docker 卷？
# - 本项目使用绑定挂载，数据存储在 ./data/ 目录
# - 直接删除目录内容即可
# - 更简单、更直观
```

---

### 4. diagnose.sh - 诊断脚本 ⭐⭐

**重要程度**: ⭐⭐ 重要

**功能**:
- 收集所有服务的状态信息
- 检查配置文件
- 测试网络连接
- 生成诊断报告

**检查项目**:
```bash
1. 容器状态
   - 所有容器的运行状态
   - 端口映射情况

2. ceph-exporter 状态
   - 运行状态
   - 重启次数
   - 最近日志

3. Ceph 集群状态
   - ceph-demo 容器状态
   - Ceph 集群健康状态
   - Ceph 版本信息

4. 配置文件检查
   - /etc/ceph/ 目录内容
   - ceph.conf 配置
   - keyring 文件

5. 网络连接测试
   - ceph-exporter 到 ceph-demo 的连接
   - 各服务的健康检查端点

6. 资源使用情况
   - 容器内存使用
   - 容器 CPU 使用
   - 磁盘空间
```

**使用方法**:
```bash
# 运行诊断
./diagnose.sh

# 保存诊断报告
./diagnose.sh > diagnose-report.txt

# 查看特定部分
./diagnose.sh | grep -A 10 "Ceph 集群状态"
```

---

### 5. verify-deployment.sh - 部署验证脚本 ⭐

**重要程度**: ⭐ 常用

**功能**:
- 验证所有服务是否正常运行
- 测试健康检查端点
- 检查服务可访问性

**验证项目**:
```bash
1. 容器状态检查
   - 所有容器是否运行
   - 容器健康状态

2. 端点可访问性测试
   - ceph-exporter: http://localhost:9128/health
   - Prometheus: http://localhost:9090/-/healthy
   - Grafana: http://localhost:3000/api/health
   - Elasticsearch: http://localhost:9200
   - Kibana: http://localhost:5601/api/status
   - Jaeger: http://localhost:16686

3. Ceph 集群验证
   - ceph -s 命令测试
   - 集群健康状态
```

---

### 6. deploy-full-stack.sh - 完整栈部署脚本 ⭐

**重要程度**: ⭐ 辅助脚本

**功能**:
- 快速部署完整监控栈
- 简化的部署流程

**与 deploy.sh full 的区别**:
```bash
deploy.sh full:
  - 完整的环境检查
  - 配置镜像加速
  - 预拉取镜像
  - 部署服务
  - 验证部署

deploy-full-stack.sh:
  - 直接部署
  - 适合已配置好环境的情况
  - 更快但检查较少
```

---

### 7. test-ceph-demo.sh - Ceph Demo 测试脚本 ⭐

**重要程度**: ⭐ 测试工具

**功能**:
- 测试 Ceph Demo 独立运行
- 验证 Ceph 功能
- 调试 Ceph 配置

---

## 💡 使用建议

### 新用户部署流程
```bash
# 1. 检查环境
./deploy.sh check

# 2. 配置镜像加速（国内用户必需）
./deploy.sh mirror

# 3. 部署完整栈（推荐）
./deploy.sh full

# 4. 验证部署
./deploy.sh verify
```

### 遇到问题时的诊断流程
```bash
# 1. 查看服务状态
./deploy.sh status

# 2. 运行诊断脚本
./diagnose.sh

# 3. 查看特定服务日志
./deploy.sh logs ceph-exporter
./deploy.sh logs ceph-demo

# 4. 如果 Ceph 有问题，清理并重新初始化
./clean-volumes.sh
```

### 日常维护
```bash
# 查看状态
./deploy.sh status

# 查看日志
./deploy.sh logs

# 停止服务
./deploy.sh stop

# 清理数据（谨慎使用）
./deploy.sh clean
```

---

## 🔧 常见问题和解决方案

### 问题 1: 镜像拉取失败
```bash
# 解决方案:
1. 配置镜像加速
   ./deploy.sh mirror

2. 手动拉取镜像
   ./deploy.sh pull

3. 检查网络连接
   ping docker.mirrors.ustc.edu.cn
```

### 问题 2: Ceph Demo 无法启动
```bash
# 解决方案:
1. 查看日志
   docker logs ceph-demo

2. 清理数据卷
   ./clean-volumes.sh

3. 检查资源
   free -h
   df -h
```

### 问题 3: 服务无法访问
```bash
# 解决方案:
1. 检查防火墙
   sudo systemctl status firewalld

2. 检查端口占用
   sudo netstat -tulpn | grep 9128

3. 运行诊断
   ./diagnose.sh
```

---

## 📝 脚本开发说明

### Bash 最佳实践

所有脚本都遵循以下最佳实践:

1. **严格模式**
```bash
set -euo pipefail
# -e: 遇到错误立即退出
# -u: 使用未定义变量时报错
# -o pipefail: 管道中任何命令失败都会失败
```

2. **颜色输出**
```bash
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'
```

3. **日志函数**
```bash
log_info()   # 信息日志（绿色）
log_warn()   # 警告日志（黄色）
log_error()  # 错误日志（红色）
log_step()   # 步骤提示（蓝色）
```

4. **错误处理**
```bash
command || {
    log_error "命令失败"
    exit 1
}
```

---

## 📚 相关文档

- **README.md** - 项目主文档
- **DEPLOYMENT_GUIDE.md** - 完整部署指南
- **QUICK_START.md** - 快速开始
- **ceph-exporter/README.md** - 详细架构文档

---

**文档版本**: 1.1
**最后更新**: 2026-03-09
**维护者**: ceph-exporter 项目团队

**更新日志**:
- **2026-03-09**: 添加时区配置说明，所有容器自动挂载宿主机时区
- **2026-03-07**: 初始版本创建
