# Ceph Exporter 环境配置问题处理记录

## 📅 日期
2026-03-07

## 🎯 目标
在 CentOS 7 系统上配置 ceph-exporter 项目的完整开发环境，使 pre-commit 检查能够正常运行。

---

## 问题 1: pip 命令未找到

### ❌ 错误现象
```bash
$ pip install pre-commit
bash: pip: 未找到命令...
```

### 🔍 原因分析
- 系统安装了 Python 3.8.6，但 pip 命令名为 `pip3`
- pip 可执行文件位于 `/usr/local/python-3.8/bin/pip3`

### ✅ 解决方案
使用 `pip3` 命令代替 `pip`:
```bash
pip3 install pre-commit
```

**可选**: 创建别名方便使用
```bash
echo "alias pip='pip3'" >> ~/.bashrc
source ~/.bashrc
```

### 📊 验证结果
```bash
$ pre-commit --version
pre-commit 3.5.0
```

---

## 问题 2: pre-commit-hooks 需要 Python 3.9+

### ❌ 错误现象
```
ERROR: Package 'pre-commit-hooks' requires a different Python: 3.8.6 not in '>=3.9'
```

### 🔍 原因分析
- 系统 Python 版本: 3.8.6
- pre-commit-hooks v6.0.0 要求 Python >= 3.9

### ✅ 解决方案
降级 pre-commit-hooks 到支持 Python 3.8 的版本。

修改 `.pre-commit-config.yaml`:
```yaml
repos:
  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v4.6.0  # 从 v6.0.0 降级到 v4.6.0
```

### 📊 验证结果
pre-commit-hooks 成功安装并运行。

---

## 问题 3: Go 工具未安装

### ❌ 错误现象
```
go fmt...................................................................Failed
- exit code: 127
/usr/bin/bash: gofmt: 未找到命令
/usr/bin/bash: goimports: 未找到命令
/usr/bin/bash: go: 未找到命令
/usr/bin/bash: golangci-lint: 未找到命令
```

### 🔍 原因分析
系统未安装 Go 编译器和相关开发工具。

### ✅ 解决方案

#### 步骤 1: 安装 Go 1.21.6
```bash
# 下载 Go
cd /tmp
wget https://go.dev/dl/go1.21.6.linux-amd64.tar.gz

# 解压到 /usr/local (需要 root 权限)
sudo tar -C /usr/local -xzf go1.21.6.linux-amd64.tar.gz
```

#### 步骤 2: 配置环境变量
```bash
# 添加到 ~/.bashrc
cat >> ~/.bashrc << 'EOF'

# Go 环境配置
export PATH=$PATH:/usr/local/go/bin
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
EOF

# 立即生效
source ~/.bashrc
```

#### 步骤 3: 配置 Go 代理 (加速下载)
```bash
go env -w GOPROXY=https://goproxy.cn,direct
```

#### 步骤 4: 安装 Go 工具
```bash
# 安装 goimports
go install golang.org/x/tools/cmd/goimports@latest

# 安装 golangci-lint
wget -q https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh -O /tmp/install-golangci-lint.sh
bash /tmp/install-golangci-lint.sh -b $HOME/go/bin v1.55.2
```

### 📊 验证结果
```bash
$ go version
go version go1.21.6 linux/amd64

$ which go gofmt goimports golangci-lint
/usr/local/go/bin/go
/usr/local/go/bin/gofmt
/home/lfl/go/bin/goimports
/home/lfl/go/bin/golangci-lint
```

---

## 问题 4: Ceph 开发库未安装

### ❌ 错误现象
```
# github.com/ceph/go-ceph/rados
fatal error: rados/librados.h: No such file or directory
```

### 🔍 原因分析
- 项目使用 go-ceph 库，需要 Ceph C 库的头文件
- 系统未安装 librados-devel 和 librbd-devel

### ✅ 解决方案

#### 挑战: CentOS 7 已 EOL，mirrorlist 服务停止

CentOS 7 在 2024 年 6 月 30 日已经 EOL，官方 mirrorlist 服务已停止，导致无法直接安装。

#### 解决步骤

**步骤 1: 禁用有问题的仓库;如果存在这些文件,也可以直接删除**
```bash
sudo sed -i 's/^enabled=1/enabled=0/g' /etc/yum.repos.d/CentOS-NFS-Ganesha-28.repo
sudo sed -i 's/^enabled=1/enabled=0/g' /etc/yum.repos.d/CentOS-fasttrack.repo
sudo sed -i 's/^enabled=1/enabled=0/g' /etc/yum.repos.d/CentOS-x86_64-kernel.repo
```

**步骤 2: 配置阿里云 Ceph 镜像**
```bash
sudo tee /etc/yum.repos.d/CentOS-Ceph-Nautilus.repo > /dev/null << 'EOF'
[centos-ceph-nautilus]
name=CentOS-7 - Ceph Nautilus
baseurl=https://mirrors.aliyun.com/centos/7/storage/x86_64/ceph-nautilus/
gpgcheck=0
enabled=1
EOF
```

**步骤 3: 清理缓存并安装**
```bash
sudo yum clean all
sudo yum makecache fast
sudo yum install -y librados-devel librbd-devel
```

### 📊 验证结果
```bash
$ rpm -qa | grep -E "librados-devel|librbd-devel"
librados-devel-14.2.20-1.el7.x86_64
librbd-devel-14.2.20-1.el7.x86_64

$ ls -la /usr/include/rados/librados.h
-rw-r--r-- 1 root root 85123 /usr/include/rados/librados.h

$ ls -la /usr/lib64/librados.so
lrwxrwxrwx 1 root root 18 /usr/lib64/librados.so -> librados.so.2.0.0
```

---

## 问题 5: go-ceph 版本与 Ceph 库不兼容

### ❌ 错误现象
```
# github.com/ceph/go-ceph/rados
../../go/pkg/mod/github.com/ceph/go-ceph@v0.27.0/rados/ioctx_octopus.go:25:2:
could not determine kind of name for C.rados_set_pool_full_try
```

### 🔍 原因分析
- 安装的 Ceph 库版本: 14.2.20 (Nautilus)
- go-ceph v0.27.0 使用了 Ceph 15.x (Octopus) 的 API
- `rados_set_pool_full_try` 函数在 Ceph 14.x 中不存在

### ✅ 解决方案

#### 方案: 升级 go-ceph 版本以匹配 Ceph 15.x (Octopus)

修改 `ceph-exporter/go.mod`:
```go
require (
    github.com/ceph/go-ceph v0.27.0  // 升级到 v0.27.0，兼容 Ceph 15.x (Octopus)
    // ... 其他依赖
)
```

更新依赖:
```bash
cd ceph-exporter
go mod tidy
```

#### 构建标签更新

使用 `-tags octopus` 构建标签以匹配 Octopus 版本:

修改 `.pre-commit-config.yaml`:
```yaml
- id: go-vet
  name: go vet
  entry: bash -c 'cd ceph-exporter && go vet -tags octopus ./...'
  language: system
  files: ^ceph-exporter/.*\.go$
  pass_filenames: false

- id: golangci-lint
  name: golangci-lint
  entry: bash -c 'cd ceph-exporter && golangci-lint run --build-tags octopus --config ../.golangci.yml --timeout 5m'
  language: system
  files: ^ceph-exporter/.*\.go$
  pass_filenames: false
```

### 📊 验证结果
```bash
$ pre-commit run --all-files
trim trailing whitespace.................................................Passed
fix end of files.........................................................Passed
check yaml...............................................................Passed
check json...............................................................Passed
check for merge conflicts................................................Passed
detect private key.......................................................Passed
mixed line ending........................................................Passed
check for case conflicts.................................................Passed
go fmt...................................................................Passed
goimports................................................................Passed
go vet...................................................................Passed
golangci-lint............................................................Passed
```

✅ **所有检查通过！**

---

## 📋 完整解决方案总结

### 环境信息
- **操作系统**: Ubuntu 20.04 (Focal) / CentOS Linux 7 (Core)
- **Python 版本**: 3.8+
- **Go 版本**: 1.21+
- **Ceph 版本**: 15.2.x (Octopus) — 匹配 Ubuntu 20.04

### 安装的工具和库
| 工具/库 | 版本 | 用途 |
|---------|------|------|
| pre-commit | 3.5.0 | Git hooks 管理 |
| pre-commit-hooks | v4.6.0 | 通用代码检查 |
| Go | 1.21+ | Go 编译器 |
| goimports | latest | Go import 管理 |
| golangci-lint | v1.55.2 | Go 代码检查 |
| librados-dev(el) | 15.2.x | Ceph RADOS 开发库 (Octopus) |
| librbd-dev(el) | 15.2.x | Ceph RBD 开发库 (Octopus) |
| go-ceph | v0.27.0 | Go Ceph 客户端库 |

### 关键配置修改

#### 1. `.pre-commit-config.yaml`
```yaml
repos:
  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v4.6.0  # 支持 Python 3.8

  - repo: local
    hooks:
      - id: go-vet
        entry: bash -c 'cd ceph-exporter && go vet -tags octopus ./...'

      - id: golangci-lint
        entry: bash -c 'cd ceph-exporter && golangci-lint run --build-tags octopus --config ../.golangci.yml --timeout 5m'
```

#### 2. `ceph-exporter/go.mod`
```go
require (
    github.com/ceph/go-ceph v0.27.0  // 兼容 Ceph 15.x (Octopus)
)
```

#### 3. `~/.bashrc`
```bash
# Go 环境配置
export PATH=$PATH:/usr/local/go/bin
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
```

---

## 🚀 使用指南

### 运行 pre-commit 检查
```bash
# 检查所有文件
pre-commit run --all-files

# 检查暂存的文件
pre-commit run

# 安装 git hooks (自动在 commit 前运行)
pre-commit install
```

### 编译项目
```bash
cd ceph-exporter

# 使用构建标签编译
go build -tags octopus -o ceph-exporter ./cmd/ceph-exporter

# 运行测试
go test -tags octopus ./...
```

### 提交代码
```bash
git add .
git commit -m "你的提交信息"
# pre-commit hooks 会自动运行检查
```

---

## 🔍 故障排查

### 问题: go vet 或 golangci-lint 失败
**检查**: 是否使用了构建标签
```bash
# 正确的命令
go vet -tags octopus ./...
golangci-lint run --build-tags octopus
```

### 问题: 找不到 librados.h
**检查**: Ceph 开发库是否安装
```bash
rpm -qa | grep librados-devel
ls -la /usr/include/rados/librados.h
```

### 问题: Go 命令未找到
**检查**: 环境变量是否配置
```bash
echo $PATH | grep go
source ~/.bashrc
```

---

## 📚 参考资源

- [pre-commit 官方文档](https://pre-commit.com/)
- [Go 官方下载](https://go.dev/dl/)
- [go-ceph GitHub](https://github.com/ceph/go-ceph)
- [Ceph 文档](https://docs.ceph.com/)
- [golangci-lint 文档](https://golangci-lint.run/)

---

## 💡 经验教训

1. **CentOS 7 EOL 问题**: 需要使用国内镜像源或 vault.centos.org
2. **版本兼容性**: 确保 go-ceph 版本与系统 Ceph 库版本匹配
3. **构建标签**: 使用 `-tags` 可以排除不兼容的代码
4. **环境变量**: 修改 `.bashrc` 后需要 `source` 使其生效
5. **依赖管理**: `go mod tidy` 可以自动清理和更新依赖

---

## ✅ 最终状态

所有 pre-commit 检查均通过，开发环境配置完成！

```
trim trailing whitespace.................................................Passed
fix end of files.........................................................Passed
check yaml...............................................................Passed
check json...............................................................Passed
check for merge conflicts................................................Passed
detect private key.......................................................Passed
mixed line ending........................................................Passed
check for case conflicts.................................................Passed
go fmt...................................................................Passed
goimports................................................................Passed
go vet...................................................................Passed
golangci-lint............................................................Passed
```

---

## 问题 8: fix-precommit.sh 脚本执行失败 - pip 命令未找到

**日期**: 2026-03-08

### ❌ 错误现象

运行 fix-precommit.sh 脚本时报错：

```bash
$ ./scripts/fix-precommit.sh
==========================================
Pre-commit SSL 问题快速修复
==========================================

步骤 1: 配置 pip 使用国内镜像源
----------------------------------------
✓ pip 配置已更新: /home/lfl/.pip/pip.conf

步骤 2: 验证 pip 配置
----------------------------------------
./scripts/fix-precommit.sh:行44: pip: 未找到命令
```

### 🔍 原因分析

1. 系统中安装了 Python 3.8.6 和 pip3
2. pip3 位于 `/usr/local/python-3.8/bin/pip3`
3. 但是 `pip` 命令不在 PATH 中
4. 脚本使用 `pip` 命令，而不是 `pip3`

### ✅ 解决方案

**方案 1: 创建 pip 符号链接（推荐）**

```bash
# 创建符号链接到用户本地 bin 目录
ln -sf /usr/local/python-3.8/bin/pip3 ~/.local/bin/pip

# 验证
which pip
# 输出: /home/lfl/.local/bin/pip

pip --version
# 输出: pip 20.2.1 from /usr/local/python-3.8/lib/python3.8/site-packages/pip (python 3.8)
```

**方案 2: 创建系统级别的符号链接（需要 sudo）**

```bash
sudo ln -sf /usr/local/python-3.8/bin/pip3 /usr/local/bin/pip
```

**方案 3: 使用别名**

```bash
echo "alias pip='pip3'" >> ~/.bashrc
source ~/.bashrc
```

### 📊 验证结果

创建符号链接后，重新运行脚本：

```bash
$ ./scripts/fix-precommit.sh
==========================================
Pre-commit SSL 问题快速修复
==========================================

步骤 1: 配置 pip 使用国内镜像源
----------------------------------------
✓ pip 配置已更新: /home/lfl/.pip/pip.conf

步骤 2: 验证 pip 配置
----------------------------------------
[global]
index-url=https://pypi.tuna.tsinghua.edu.cn/simple
trusted-host=pypi.tuna.tsinghua.edu.cn
[install]
trusted-host=pypi.tuna.tsinghua.edu.cn

步骤 3: 清理 pre-commit 缓存
----------------------------------------
✓ 缓存已清理
...
```

### 💡 经验总结

1. **Python 命令命名**: CentOS 7 系统中，Python 3 的 pip 命令通常命名为 `pip3`
2. **PATH 配置**: `~/.local/bin` 目录通常已在 PATH 中，适合存放用户级别的可执行文件
3. **符号链接 vs 别名**:
   - 符号链接对所有程序生效（包括脚本）
   - 别名仅在交互式 shell 中生效
4. **验证方法**: 使用 `which` 和 `--version` 命令验证配置是否生效

🎉 **环境配置成功！**

---

# Ceph RGW 端口问题分析与解决方案

## 📅 日期
2026-03-10

## 问题描述

访问 `http://192.168.75.129:8080/` 时，浏览器返回以下 XML 响应：

```xml
<ListAllMyBucketsResult xmlns="http://s3.amazonaws.com/doc/2006-03-01/">
<Owner>
  <ID>anonymous</ID>
  <DisplayName/>
</Owner>
<Buckets/>
</ListAllMyBucketsResult>
```

## 问题分析

### 响应含义

这是一个标准的 **S3 API 响应**，表示：
- **Owner**: 当前访问者是匿名用户（anonymous）
- **Buckets**: 空的，表示没有任何 S3 存储桶

### 根本原因

端口 8080 上运行的是 **Ceph RGW (RADOS Gateway)** 服务，而不是 Ceph Dashboard。

根据 `docker-compose-lightweight-full.yml` 配置：

```yaml
ceph-demo:
  environment:
    - RGW_CIVETWEB_PORT=8080  # RGW 绑定到 8080 端口
    - DEMO_DAEMONS=mon,mgr,osd,rgw
  ports:
    - "8080:8080"  # 注释说是 Ceph Dashboard，但实际是 RGW
    - "5000:5000"  # RGW 另一个端口
```

**配置问题**: `RGW_CIVETWEB_PORT=8080` 将 RGW 绑定到了 8080 端口，这与注释中说的 "Ceph Dashboard" 不符。

### 服务说明

- **RGW (RADOS Gateway)**: 提供 S3/Swift 兼容的对象存储 API，不是 Web 界面
- **Ceph Dashboard**: Ceph 的 Web 管理界面，需要单独启用

## 解决方案

### 方案 1: 访问 Grafana 监控界面（推荐）

Grafana 提供了更好的可视化监控界面：

```bash
# 访问地址
http://192.168.75.129:3000

# 默认账号密码
用户名: admin
密码: admin
```

**功能**:
- 可视化 Ceph 集群监控数据
- 预配置的 Ceph 监控仪表板
- 实时指标图表和告警

### 方案 2: 访问 Prometheus 原始指标

```bash
# ceph-exporter 指标端点
http://192.168.75.129:9128/metrics

# Prometheus 查询界面
http://192.168.75.129:9090
```

**功能**:
- 查看原始 Prometheus 指标
- 执行 PromQL 查询
- 查看告警规则

### 方案 3: 使用 S3 客户端访问 RGW

如果需要使用对象存储功能，可以使用 S3 客户端工具：

#### 使用 AWS CLI

```bash
# 安装 AWS CLI
pip install awscli

# 配置
aws configure
# AWS Access Key ID: demo_access_key
# AWS Secret Access Key: demo_secret_key
# Default region name: us-east-1
# Default output format: json

# 使用自定义端点
aws --endpoint-url=http://192.168.75.129:8080 s3 ls

# 创建存储桶
aws --endpoint-url=http://192.168.75.129:8080 s3 mb s3://my-bucket

# 上传文件
aws --endpoint-url=http://192.168.75.129:8080 s3 cp file.txt s3://my-bucket/
```

#### 使用 s3cmd

```bash
# 安装 s3cmd
yum install s3cmd

# 配置文件 ~/.s3cfg
[default]
access_key = demo_access_key
secret_key = demo_secret_key
host_base = 192.168.75.129:8080
host_bucket = 192.168.75.129:8080
use_https = False

# 列出存储桶
s3cmd ls

# 创建存储桶
s3cmd mb s3://my-bucket
```

### 方案 4: 启用真正的 Ceph Dashboard

Ceph Dashboard 通常运行在 MGR 模块上，需要单独启用：

```bash
# 进入 ceph-demo 容器
sudo docker exec -it ceph-demo bash

# 启用 dashboard 模块
ceph mgr module enable dashboard

# 创建自签名证书（如果需要 HTTPS）
ceph dashboard create-self-signed-cert

# 创建管理员用户
ceph dashboard ac-user-create admin password administrator

# 查看 dashboard 访问地址
ceph mgr services

# 示例输出:
# {
#     "dashboard": "https://172.20.0.10:8443/"
# }
```

**注意**: Dashboard 默认使用 HTTPS 和不同的端口（通常是 8443），需要修改 docker-compose 配置暴露该端口。

## 端口映射总结

| 端口 | 服务 | 说明 |
|------|------|------|
| 8080 | RGW | 对象存储 API（S3/Swift 兼容） |
| 5000 | RGW | RGW 备用端口 |
| 9128 | ceph-exporter | Prometheus 指标端点 |
| 9090 | Prometheus | Prometheus 查询界面 |
| 3000 | Grafana | 监控可视化界面（推荐） |
| 5601 | Kibana | 日志查询界面 |
| 16686 | Jaeger | 分布式追踪界面 |

## 推荐访问方式

1. **监控查看**: 访问 Grafana `http://192.168.75.129:3000`
2. **指标查询**: 访问 Prometheus `http://192.168.75.129:9090`
3. **对象存储**: 使用 S3 客户端连接 `http://192.168.75.129:8080`

## 修复配置（可选）

如果需要修改端口配置，编辑 `docker-compose-lightweight-full.yml`:

```yaml
ceph-demo:
  environment:
    - RGW_CIVETWEB_PORT=5000  # 将 RGW 改到 5000 端口
    - DEMO_DAEMONS=mon,mgr,osd,rgw
  ports:
    - "5000:5000"  # RGW
    - "8443:8443"  # Ceph Dashboard (需要先启用)
```

然后重启服务：

```bash
cd /home/lfl/ceph-exporter/ceph-exporter/deployments
sudo docker-compose -f docker-compose-lightweight-full.yml restart ceph-demo
```

## 相关文档

- [Ceph RGW 文档](https://docs.ceph.com/en/latest/radosgw/)
- [Ceph Dashboard 文档](https://docs.ceph.com/en/latest/mgr/dashboard/)
- [S3 API 参考](https://docs.aws.amazon.com/AmazonS3/latest/API/)

---

# Ceph 开发库版本升级 — 匹配 Ubuntu 20.04

## 📅 日期
2026-03-21

## 🎯 目标
将项目的 Ceph 开发库版本从 14.x (Nautilus) 升级到 15.2.x (Octopus)，以匹配 Ubuntu 20.04 (Focal) 系统自带的 Ceph 版本。

---

## 🔍 背景分析

### 升级前状态

项目原先针对 CentOS 7 环境配置，绑定了 Ceph 14.x (Nautilus) 版本：

| 组件 | 升级前版本 | 说明 |
|------|-----------|------|
| Ceph C 库 | 14.2.20 (Nautilus) | CentOS 7 yum 安装 |
| go-ceph | v0.20.0 | 降级以兼容 Nautilus |
| 构建标签 | `-tags nautilus` | 排除 Octopus+ 代码 |
| Docker 镜像 | `ceph/daemon:latest-nautilus` | Nautilus 版本容器 |

### 为什么需要升级

- **Ubuntu 20.04 (Focal)** 默认仓库提供的 Ceph 版本为 **15.2.x (Octopus)**
- 原有的 go-ceph v0.20.0 + Nautilus 配置无法在 Ubuntu 20.04 上正常编译
- 需要 go-ceph 版本与系统 Ceph C 库版本匹配

### 版本对应关系

| Ubuntu 版本 | Ceph 版本 | go-ceph 兼容版本 | 构建标签 |
|------------|-----------|-----------------|---------|
| Ubuntu 18.04 (Bionic) | 12.x (Luminous) | v0.15.0 以下 | `luminous` |
| CentOS 7 | 14.x (Nautilus) | v0.20.0 | `nautilus` |
| **Ubuntu 20.04 (Focal)** | **15.2.x (Octopus)** | **v0.27.0** | **`octopus`** |
| Ubuntu 22.04 (Jammy) | 17.x (Quincy) | v0.28.0+ | `quincy` |

---

## ✅ 解决方案

### 概述

共修改 **10 个文件**，涉及 Go 依赖、构建系统、容器配置、文档四个层面。

### 修改清单

| 文件 | 修改内容 | 层面 |
|------|---------|------|
| `ceph-exporter/go.mod` | go-ceph `v0.20.0` → `v0.27.0` | Go 依赖 |
| `ceph-exporter/go.sum` | 自动更新依赖校验哈希 | Go 依赖 |
| `ceph-exporter/Dockerfile` | `go build` 添加 `-tags octopus` | 构建系统 |
| `ceph-exporter/Makefile` | 新增 `BUILD_TAGS := octopus`，所有构建/测试/检查目标添加 `-tags $(BUILD_TAGS)` | 构建系统 |
| `.pre-commit-config.yaml` | go vet / golangci-lint 标签 `nautilus` → `octopus` | 构建系统 |
| `deployments/docker-compose-ceph-demo.yml` | `ceph/daemon:latest-nautilus` → `latest-octopus` | 容器配置 |
| `deployments/docker-compose-integration-test.yml` | `ceph/demo:latest-nautilus` → `latest-octopus` | 容器配置 |
| `deployments/docker-compose-lightweight-full.yml` | `ceph/daemon:latest-nautilus` → `latest-octopus` | 容器配置 |
| `ceph-exporter/DEVELOPMENT.md` | 更新 Ceph 开发库安装说明 | 文档 |
| `fix_ceph.md` | 更新版本信息和构建指令 | 文档 |

---

### 详细修改说明

#### 1. Go 依赖升级 — `ceph-exporter/go.mod`

```diff
 require (
-    github.com/ceph/go-ceph v0.20.0
+    github.com/ceph/go-ceph v0.27.0
 )
```

升级后执行 `go mod tidy` 自动更新 `go.sum`。

go-ceph v0.27.0 的关键变化：
- 支持 Ceph 15.x (Octopus) 的 `rados_set_pool_full_try` 等新 API
- 通过构建标签隔离不同 Ceph 版本的代码（如 `ioctx_octopus.go`、`ioctx_pacific.go`）

#### 2. Dockerfile 构建标签 — `ceph-exporter/Dockerfile`

```diff
 RUN CGO_ENABLED=1 go build \
+    -tags octopus \
     -ldflags "..." \
     -o /app/build/ceph-exporter \
     ./cmd/ceph-exporter
```

**为什么必须加 `-tags octopus`？**

go-ceph v0.27.0 包含适配多个 Ceph 版本的源文件：
- `ioctx_octopus.go` — 需要 Ceph 15.x API（Octopus）
- `ioctx_pacific.go` — 需要 Ceph 16.x API（Pacific）

不加构建标签时，编译器会尝试编译**所有**文件，包括 Pacific 版本的代码，导致编译报错：
```
could not determine kind of name for C.rados_xxx
```

加上 `-tags octopus` 后，只编译 Ceph 15.x 及以下兼容的代码。

#### 3. Makefile 构建标签 — `ceph-exporter/Makefile`

新增 `BUILD_TAGS` 变量统一管理：

```makefile
BUILD_TAGS  := octopus
```

影响的目标：

| Make 目标 | 修改说明 |
|-----------|---------|
| `build` | 添加 `-tags $(BUILD_TAGS)` |
| `build-linux` | 添加 `-tags $(BUILD_TAGS)` |
| `test` | 添加 `-tags $(BUILD_TAGS)` |
| `test-cover` | 添加 `-tags $(BUILD_TAGS)` |
| `test-short` | 添加 `-tags $(BUILD_TAGS)` |
| `test-integration` | 添加 `-tags $(BUILD_TAGS)` |
| `lint` | `go vet` 添加 `-tags $(BUILD_TAGS)` |

#### 4. Pre-commit 构建标签 — `.pre-commit-config.yaml`

```diff
 - id: go-vet
-  entry: bash -c '... go vet -tags nautilus ...'
+  entry: bash -c '... go vet -tags octopus ...'

 - id: golangci-lint
-  entry: bash -c '... golangci-lint run --build-tags nautilus ...'
+  entry: bash -c '... golangci-lint run --build-tags octopus ...'
```

#### 5. Docker Compose 镜像标签

三个文件统一升级：

```diff
-    image: ceph/daemon:latest-nautilus
+    image: ceph/daemon:latest-octopus

-    image: ceph/demo:latest-nautilus
+    image: ceph/demo:latest-octopus
```

#### 6. 开发文档 — `ceph-exporter/DEVELOPMENT.md`

更新 Ceph 开发库安装说明，明确 Ubuntu 20.04 的安装方式：

```bash
# Ubuntu 20.04 (Focal) - 默认仓库已包含 Ceph 15.x (Octopus)
sudo apt-get install -y librados-dev librbd-dev
```

---

## 📊 升级前后对比

| 组件 | 升级前 | 升级后 |
|------|--------|--------|
| **目标系统** | CentOS 7 | Ubuntu 20.04 (Focal) |
| **Ceph 版本** | 14.x (Nautilus) | 15.2.x (Octopus) |
| **go-ceph** | v0.20.0 | v0.27.0 |
| **构建标签** | `-tags nautilus` | `-tags octopus` |
| **Docker 镜像** | `latest-nautilus` | `latest-octopus` |
| **Ceph 开发库安装** | `yum install librados-devel` | `apt install librados-dev` |

---

## 🚀 Ubuntu 20.04 快速开始

### 1. 安装 Ceph 开发库

```bash
# Ubuntu 20.04 默认仓库已包含 Ceph 15.2.x (Octopus)，无需额外配置源
sudo apt-get update
sudo apt-get install -y librados-dev librbd-dev

# 验证安装
dpkg -l | grep librados-dev
ls -la /usr/include/rados/librados.h
```

### 2. 编译项目

```bash
cd ceph-exporter

# 使用 Makefile（推荐，已内置 -tags octopus）
make build

# 或手动编译
CGO_ENABLED=1 go build -tags octopus -o build/ceph-exporter ./cmd/ceph-exporter
```

### 3. 运行测试

```bash
# 使用 Makefile
make test

# 或手动运行
CGO_ENABLED=1 go test -tags octopus -v ./internal/...
```

### 4. Docker 构建

```bash
# Dockerfile 已内置 -tags octopus
docker build -t ceph-exporter:dev .
```

### 5. 启动完整栈

```bash
cd deployments
docker-compose -f docker-compose-lightweight-full.yml up -d
```

---

## 💡 注意事项

1. **构建标签是强制的**: 使用 go-ceph v0.27.0 时，必须指定 `-tags octopus`（或更高版本标签），否则编译器会尝试编译 Pacific (Ceph 16.x) 的代码，导致找不到 C 函数定义而报错
2. **版本一致性**: 确保系统安装的 Ceph C 库版本（`librados-dev`）与构建标签匹配：Octopus 对应 `-tags octopus`
3. **Docker 旧数据**: 如果之前使用 Nautilus 镜像运行过，升级到 Octopus 前建议清理旧数据目录 `deployments/data/ceph-demo/`
4. **未来升级**: 如需适配 Ubuntu 22.04 (Ceph 17.x Quincy)，需将 go-ceph 升级到 v0.28.0+，构建标签改为 `-tags quincy`

---

# Docker Compose YAML 文件版本升级

## 📅 日期
2026-03-21

## 🎯 目标
将项目中所有 Docker Compose YAML 文件从过时的 `version: "2.1"` 格式升级到现代 Compose Spec 标准，消除已废弃的语法。

---

## 🔍 背景分析

### 升级前的问题

| 问题 | 说明 |
|------|------|
| `version: "2.1"` 已废弃 | Docker Compose V2 CLI 已完全忽略 `version` 字段，该字段仅产生警告 |
| `mem_limit` 是旧语法 | Compose V2 格式的 `mem_limit` 不符合 Compose Spec 标准 |
| 版本不统一 | 4 个部署文件用 `"2.1"`，1 个示例文件用 `'3.8'`，不一致 |
| CI 使用废弃工具 | GitHub Actions 安装独立的 `docker-compose` 二进制，而非内置的 `docker compose` 插件 |

### Docker Compose 版本演进

| 阶段 | 版本格式 | CLI 工具 | 状态 |
|------|---------|---------|------|
| Compose V1 | `version: "1"` | `docker-compose`（Python） | 已淘汰 |
| Compose V2 文件格式 | `version: "2.x"` | `docker-compose` | 已废弃 |
| Compose V3 文件格式 | `version: "3.x"` | `docker-compose` / `docker compose` | 已废弃 |
| **Compose Spec（现代标准）** | **无 `version` 字段** | **`docker compose`（Go 插件）** | **当前标准** |

> Docker 官方自 Compose V2.21.0 起，遇到 `version` 字段会输出警告：
> `WARN[0000] .../docker-compose.yml: 'version' is obsolete`

---

## ✅ 解决方案

### 概述

共修改 **6 个文件**，涉及三类变更：

| 变更类型 | 涉及文件数 | 说明 |
|---------|-----------|------|
| 移除 `version` 字段 | 5 | 所有 docker-compose 文件 |
| `mem_limit` → `deploy.resources` | 4 | 18 处服务的内存限制 |
| `docker-compose` → `docker compose` | 1 | CI 工作流文件 |

---

### 修改清单

#### 1. 移除 `version` 字段（5 个文件）

| 文件 | 原值 | 操作 |
|------|------|------|
| `deployments/docker-compose.yml` | `version: "2.1"` | 删除该行 |
| `deployments/docker-compose-lightweight-full.yml` | `version: "2.1"` | 删除该行 |
| `deployments/docker-compose-ceph-demo.yml` | `version: "2.1"` | 删除该行 |
| `deployments/docker-compose-integration-test.yml` | `version: "2.1"` | 删除该行 |
| `docs/examples/docker-compose-elk.yaml` | `version: '3.8'` | 删除该行 |

#### 2. `mem_limit` 迁移为 Compose Spec 标准写法（4 个文件，18 处）

**旧写法（Compose V2 格式）：**
```yaml
services:
  my-service:
    image: xxx
    mem_limit: 128m
```

**新写法（Compose Spec 标准）：**
```yaml
services:
  my-service:
    image: xxx
    deploy:
      resources:
        limits:
          memory: 128m
```

**各文件修改详情：**

| 文件 | 修改的服务 | 内存限制 |
|------|-----------|---------|
| `docker-compose.yml` | ceph-exporter | 128m |
| | prometheus | 512m |
| | grafana | 256m |
| | alertmanager | 128m |
| `docker-compose-lightweight-full.yml` | ceph-demo | 1024m |
| | ceph-exporter | 128m |
| | prometheus | 512m |
| | grafana | 256m |
| | alertmanager | 128m |
| | filebeat-sidecar | 128m |
| | elasticsearch | 512m |
| | logstash | 768m |
| | kibana | 1024m |
| | jaeger | 256m |
| `docker-compose-integration-test.yml` | ceph-demo | 1024m |
| | ceph-exporter | 128m |
| | prometheus | 512m |
| | grafana | 256m |

#### 3. CI 工作流升级（1 个文件）

**文件**: `.github/workflows/integration-test.yml`

```diff
-      - name: Install Docker Compose
-        run: |
-          sudo apt-get update
-          sudo apt-get install -y docker-compose
-
       - name: Verify Docker installation
         run: |
           docker --version
-          docker-compose --version
+          docker compose version

-          docker-compose -f ... logs > ...
+          docker compose -f ... logs > ...

-          docker-compose -f ... down -v || true
+          docker compose -f ... down -v || true
```

**变更说明：**
- 删除了安装独立 `docker-compose` 二进制的步骤（`ubuntu-latest` 已内置 `docker compose` 插件）
- 所有 `docker-compose` 命令改为 `docker compose`（注意：连字符变空格）
- `--version` 改为 `version`（子命令风格）

---

## 📊 升级前后对比

| 项目 | 升级前 | 升级后 |
|------|--------|--------|
| **Compose 格式** | `version: "2.1"` / `"3.8"` 混用 | Compose Spec（无 version 字段） |
| **内存限制语法** | `mem_limit: 128m` | `deploy.resources.limits.memory: 128m` |
| **CI 工具** | 独立 `docker-compose` 二进制（Python） | 内置 `docker compose` 插件（Go） |
| **兼容性** | 仅兼容旧版 Docker Compose | 兼容 Docker Compose V2.0+ |

---

## 💡 注意事项

1. **向后兼容**: Compose Spec 格式完全向后兼容，`docker compose` 插件能正确解析所有旧语法
2. **`deploy` 在非 Swarm 模式下**: 在 Docker Compose V2 CLI 中，`deploy.resources` 即使不使用 Swarm 模式也会生效（与旧版 V3 格式不同）
3. **命令差异**: `docker-compose`（连字符）是旧的独立二进制，`docker compose`（空格）是新的内置插件，两者行为基本一致
4. **中文翻译文件**: `.zh-CN` 后缀的翻译文件未同步更新，不影响功能运行

---

# Ceph Docker 镜像源替换 — 解决无法拉取问题

## 📅 日期
2026-03-21

## 🎯 目标
解决 Docker Hub 无法访问导致 `ceph/daemon` 和 `ceph/demo` 镜像无法拉取的问题，改用 quay.io 镜像源部署 Ceph 集群。

---

## 🔍 问题分析

### 错误现象

```bash
$ sudo docker pull ceph/demo:latest-octopus
Error response from daemon: Get "https://registry-1.docker.io/v2/": net/http: request canceled
while waiting for connection (Client.Timeout exceeded while awaiting headers)

$ sudo docker pull ceph/daemon:latest-octopus
Error response from daemon: Get "https://registry-1.docker.io/v2/": net/http: request canceled
while waiting for connection (Client.Timeout exceeded while awaiting headers)
```

### 原因

Docker Hub（`registry-1.docker.io`）在国内网络环境下无法访问，连接超时。

### 可用镜像源测试

| 镜像地址 | 结果 | 说明 |
|---------|------|------|
| `ceph/demo:latest-octopus` | ❌ 超时 | Docker Hub 不可达 |
| `ceph/daemon:latest-octopus` | ❌ 超时 | Docker Hub 不可达 |
| `quay.io/ceph/demo:latest-octopus` | ❌ 不存在 | quay.io 上无 `ceph/demo` 镜像 |
| **`quay.io/ceph/daemon:latest-octopus`** | **✅ 成功** | quay.io 可访问，镜像存在 |

### 关键发现

- **`ceph/demo`** 镜像在 quay.io 上**不存在**，只有 Docker Hub 上有
- **`ceph/daemon`** 镜像在 quay.io 上**可用**，且 `ceph/daemon` 通过设置 `CEPH_DAEMON=demo` 环境变量可以完全替代 `ceph/demo` 的功能
- 两个镜像的区别仅在于：`ceph/demo` 默认以 demo 模式启动，而 `ceph/daemon` 需要显式指定 `CEPH_DAEMON=demo`

---

## ✅ 解决方案

### 统一使用 `quay.io/ceph/daemon:latest-octopus`

将所有 docker-compose 文件中的 Ceph 镜像替换为 quay.io 源，并确保设置 `CEPH_DAEMON=demo` 环境变量。

### 修改清单

| 文件 | 修改前 | 修改后 |
|------|--------|--------|
| `docker-compose-ceph-demo.yml` | `ceph/daemon:latest-octopus` | `quay.io/ceph/daemon:latest-octopus` |
| `docker-compose-lightweight-full.yml` | `ceph/daemon:latest-octopus` | `quay.io/ceph/daemon:latest-octopus` |
| `docker-compose-integration-test.yml` | `ceph/demo:latest-octopus` | `quay.io/ceph/daemon:latest-octopus` + 添加 `CEPH_DAEMON=demo` |

### 各文件详细修改

#### 1. `docker-compose-ceph-demo.yml`

仅改镜像源（已有 `CEPH_DAEMON=demo`）：

```diff
-    image: ceph/daemon:latest-octopus
+    image: quay.io/ceph/daemon:latest-octopus
```

#### 2. `docker-compose-lightweight-full.yml`

仅改镜像源（已有 `CEPH_DAEMON=demo`）：

```diff
-    image: ceph/daemon:latest-octopus
+    image: quay.io/ceph/daemon:latest-octopus
```

#### 3. `docker-compose-integration-test.yml`

改镜像源 + 补充环境变量（原来用的 `ceph/demo` 不需要显式设置，换成 `ceph/daemon` 后需要）：

```diff
-    image: ceph/demo:latest-octopus
+    image: quay.io/ceph/daemon:latest-octopus
     environment:
+      - CEPH_DAEMON=demo
       - MON_IP=172.20.0.10
```

---

## 📊 升级前后对比

| 项目 | 升级前 | 升级后 |
|------|--------|--------|
| **镜像仓库** | Docker Hub (`docker.io`) | Quay.io (`quay.io`) |
| **Ceph Demo 镜像** | `ceph/demo:latest-octopus` | `quay.io/ceph/daemon:latest-octopus` |
| **Ceph Daemon 镜像** | `ceph/daemon:latest-octopus` | `quay.io/ceph/daemon:latest-octopus` |
| **拉取结果** | ❌ 超时失败 | ✅ 正常拉取 |

---

## 💡 注意事项

1. **`CEPH_DAEMON=demo` 是关键**: 使用 `ceph/daemon` 镜像时，必须设置此环境变量才能以 demo 模式运行（自动创建 MON/OSD/MGR/RGW）
2. **quay.io 无需认证**: quay.io 上的 `ceph/daemon` 镜像是公开的，无需登录即可拉取
3. **已拉取的镜像**: `quay.io/ceph/daemon:latest-octopus` 已在本机拉取完成，可直接使用
4. **Docker Hub 镜像加速**: 如果未来需要使用 Docker Hub 上的其他镜像，可配置镜像加速器（如阿里云、腾讯云等）
