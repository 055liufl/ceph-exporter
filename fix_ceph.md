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

**步骤 1: 禁用有问题的仓库**
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

#### 方案 1: 降级 go-ceph 版本 (已采用)

修改 `ceph-exporter/go.mod`:
```go
require (
    github.com/ceph/go-ceph v0.20.0  // 从 v0.27.0 降级到 v0.20.0
    // ... 其他依赖
)
```

更新依赖:
```bash
cd ceph-exporter
go mod tidy
```

#### 方案 2: 使用构建标签

即使降级到 v0.20.0，`ioctx_octopus.go` 文件仍然存在。需要使用构建标签来排除 Octopus 特定代码。

修改 `.pre-commit-config.yaml`:
```yaml
- id: go-vet
  name: go vet
  entry: bash -c 'cd ceph-exporter && go vet -tags nautilus ./...'
  language: system
  files: ^ceph-exporter/.*\.go$
  pass_filenames: false

- id: golangci-lint
  name: golangci-lint
  entry: bash -c 'cd ceph-exporter && golangci-lint run --build-tags nautilus --config ../.golangci.yml --timeout 5m'
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
- **操作系统**: CentOS Linux 7 (Core)
- **内核版本**: 3.10.0-1160.118.1.el7.x86_64
- **Python 版本**: 3.8.6
- **Go 版本**: 1.21.6
- **Ceph 版本**: 14.2.20 (Nautilus)

### 安装的工具和库
| 工具/库 | 版本 | 用途 |
|---------|------|------|
| pre-commit | 3.5.0 | Git hooks 管理 |
| pre-commit-hooks | v4.6.0 | 通用代码检查 |
| Go | 1.21.6 | Go 编译器 |
| goimports | latest | Go import 管理 |
| golangci-lint | v1.55.2 | Go 代码检查 |
| librados-devel | 14.2.20 | Ceph RADOS 开发库 |
| librbd-devel | 14.2.20 | Ceph RBD 开发库 |
| go-ceph | v0.20.0 | Go Ceph 客户端库 |

### 关键配置修改

#### 1. `.pre-commit-config.yaml`
```yaml
repos:
  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v4.6.0  # 支持 Python 3.8

  - repo: local
    hooks:
      - id: go-vet
        entry: bash -c 'cd ceph-exporter && go vet -tags nautilus ./...'

      - id: golangci-lint
        entry: bash -c 'cd ceph-exporter && golangci-lint run --build-tags nautilus --config ../.golangci.yml --timeout 5m'
```

#### 2. `ceph-exporter/go.mod`
```go
require (
    github.com/ceph/go-ceph v0.20.0  // 兼容 Ceph 14.x
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
go build -tags nautilus -o ceph-exporter ./cmd/ceph-exporter

# 运行测试
go test -tags nautilus ./...
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
go vet -tags nautilus ./...
golangci-lint run --build-tags nautilus
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
