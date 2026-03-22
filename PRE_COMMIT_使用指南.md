# Pre-commit 使用指南

**最后更新**: 2026-03-21
**适用环境**: Ubuntu 20.04 + Python

---

## 📋 什么是 Pre-commit

Pre-commit 是一个 Git hooks 管理工具，可以在提交代码前自动运行代码检查、格式化等任务。

**优点**:
- 自动化代码质量检查
- 统一团队代码风格
- 在提交前发现问题
- 支持多种编程语言

---

## 🚀 快速开始

### 1. 安装 Pre-commit

```bash
# 使用 pip 安装
pip install pre-commit

# 或使用 pip3
pip3 install pre-commit

# 验证安装
pre-commit --version
```

### 2. 配置 pip 镜像源（国内用户推荐）

```bash
# 使用自动脚本（推荐）
./scripts/fix-precommit.sh

# 或手动配置
mkdir -p ~/.pip
cat > ~/.pip/pip.conf << 'CONF'
[global]
index-url = https://pypi.tuna.tsinghua.edu.cn/simple
trusted-host = pypi.tuna.tsinghua.edu.cn
CONF
```

### 3. 安装 Git Hooks

```bash
cd <project-root>

# 方式 1: 使用完整配置（推荐）
pre-commit install

# 方式 2: 使用简化配置（仅 Go 检查）
pre-commit install --config .pre-commit-config-simple.yaml
```

### 4. 运行检查

```bash
# 检查所有文件
pre-commit run --all-files

# 或使用简化配置
pre-commit run --config .pre-commit-config-simple.yaml --all-files

# 检查暂存的文件（git add 后的文件）
pre-commit run
```

---

## 📁 配置文件说明

### 完整配置 (.pre-commit-config.yaml)

**包含的检查**:
- 通用检查: trailing-whitespace, end-of-file-fixer, check-yaml, check-json
- Go 检查: gofmt, goimports, go vet, golangci-lint

**优点**: 全面的代码质量检查
**缺点**: 首次运行需要下载多个工具

**使用**:
```bash
pre-commit install
pre-commit run --all-files
```

### 简化配置 (.pre-commit-config-simple.yaml)

**包含的检查**:
- Go 检查: gofmt, goimports, go vet

**优点**: 下载快，问题少
**缺点**: 检查不够全面

**使用**:
```bash
pre-commit install --config .pre-commit-config-simple.yaml
pre-commit run --config .pre-commit-config-simple.yaml --all-files
```

---

## 🔧 常用命令

### 基本命令

```bash
# 安装 hooks
pre-commit install

# 卸载 hooks
pre-commit uninstall

# 运行所有 hooks
pre-commit run --all-files

# 运行特定 hook
pre-commit run gofmt --all-files
pre-commit run go-vet --all-files

# 更新 hooks 到最新版本
pre-commit autoupdate

# 清理缓存
pre-commit clean
```

### 跳过检查

```bash
# 跳过特定 hook
SKIP=golangci-lint git commit -m "message"

# 跳过所有 hooks（不推荐）
git commit --no-verify -m "message"
```

---

## 🐛 常见问题

### Q1: SSL 证书验证失败

**错误信息**:
```
SSL: CERTIFICATE_VERIFY_FAILED
URLError: <urlopen error [SSL: CERTIFICATE_VERIFY_FAILED]>
```

**解决方案**:

```bash
# 方法 1: 使用自动修复脚本（推荐）
./scripts/fix-precommit.sh

# 方法 2: 手动配置 pip 镜像源
mkdir -p ~/.pip
cat > ~/.pip/pip.conf << 'CONF'
[global]
index-url = https://pypi.tuna.tsinghua.edu.cn/simple
trusted-host = pypi.tuna.tsinghua.edu.cn

[install]
trusted-host = pypi.tuna.tsinghua.edu.cn
CONF

# 方法 3: 清理缓存后重试
pre-commit clean
pip cache purge
pip uninstall pre-commit
pip install pre-commit
pre-commit install --install-hooks
```

**可用的 pip 镜像源**（如果清华源不可用）:

| 镜像源 | index-url |
|--------|-----------|
| 清华大学（推荐） | `https://pypi.tuna.tsinghua.edu.cn/simple` |
| 阿里云 | `https://mirrors.aliyun.com/pypi/simple/` |
| 中科大 | `https://pypi.mirrors.ustc.edu.cn/simple/` |
| 豆瓣 | `https://pypi.douban.com/simple/` |

**Windows 用户**:
```powershell
New-Item -Path "$env:APPDATA\pip" -ItemType Directory -Force
@"
[global]
index-url = https://pypi.tuna.tsinghua.edu.cn/simple
trusted-host = pypi.tuna.tsinghua.edu.cn
"@ | Out-File -FilePath "$env:APPDATA\pip\pip.ini" -Encoding ASCII
pip config list
```

### Q2: 下载速度很慢

**解决方案**:
```bash
# 1. 配置 pip 镜像源
./scripts/fix-precommit.sh

# 2. 使用简化配置
pre-commit install --config .pre-commit-config-simple.yaml

# 3. 或直接使用 Makefile
make fmt lint test
```

### Q3: end-of-file-fixer 因文件权限报错 (PermissionError)

**错误信息**:
```
PermissionError: [Errno 13] Permission denied: 'ceph-exporter/deployments/filebeat/filebeat.yml'
```

**原因**: `filebeat.yml` 文件所有者为 `root`（部署需要），当前用户没有写权限，导致 `end-of-file-fixer` 无法修改该文件。

**解决方案**: 在 `.pre-commit-config.yaml` 中为 `end-of-file-fixer` 添加 `exclude` 规则，跳过该文件：

```yaml
- id: end-of-file-fixer
  exclude: 'ceph-exporter/deployments/filebeat/filebeat\.yml$'
```

> **注意**: 不要通过 `chown` 修改该文件权限，因为 `filebeat.yml` 必须保持 `root` 所有权，否则部署后服务会不断重启。如果其他部署文件也有类似权限要求，可以在 `exclude` 中用 `|` 添加多个路径，或排除整个目录：
> ```yaml
> exclude: 'ceph-exporter/deployments/(filebeat|其他目录)/.*\.yml$'
> ```

### Q4: 某些 hook 总是失败

**解决方案**:
```bash
# 1. 查看详细错误
pre-commit run --verbose --all-files

# 2. 跳过失败的 hook
SKIP=failing-hook git commit -m "message"

# 3. 或修复代码问题
make fmt  # 格式化代码
make lint # 查看具体问题
```

### Q5: Pre-commit 未安装

**解决方案**:
```bash
# 检查 Python 和 pip
python --version
pip --version

# 安装 pre-commit
pip install pre-commit

# 或使用 pip3
pip3 install pre-commit
```

---

## 💡 最佳实践

### 开发工作流

```bash
# 1. 修改代码
vim internal/config/config.go

# 2. 格式化代码（可选，pre-commit 会自动做）
make fmt

# 3. 添加到暂存区
git add internal/config/config.go

# 4. 提交（pre-commit 自动运行）
git commit -m "feat: add new config option"

# 5. 如果 pre-commit 失败，修复问题后重新提交
make fmt
git add .
git commit -m "feat: add new config option"
```

### 首次设置

```bash
# 1. 安装 pre-commit
pip install pre-commit

# 2. 配置镜像源（国内用户）
./scripts/fix-precommit.sh

# 3. 安装 hooks
pre-commit install

# 4. 运行一次检查所有文件
pre-commit run --all-files

# 5. 修复发现的问题
make fmt
```

### 团队协作

```bash
# 1. 确保所有成员安装 pre-commit
pip install pre-commit

# 2. 统一使用相同的配置
pre-commit install

# 3. 定期更新 hooks
pre-commit autoupdate

# 4. 在 CI/CD 中也运行 pre-commit
# 见 .github/workflows/pre-commit.yml
```

---

## 🔄 Pre-commit vs Makefile

### Pre-commit

**优点**:
- 自动运行（git commit 时）
- 统一的配置格式
- 支持多种语言
- 社区维护的 hooks

**缺点**:
- 需要安装 Python 和 pip
- 首次运行需要下载工具
- 可能遇到网络问题

**适用场景**:
- 团队协作
- 需要自动化检查
- 多语言项目

### Makefile

**优点**:
- 不需要额外依赖
- 速度快
- 灵活可控

**缺点**:
- 需要手动运行
- 不会自动在提交时运行

**适用场景**:
- 个人开发
- 网络受限环境
- 快速检查

### 推荐方案

**同时使用两者**:
```bash
# 日常开发：使用 pre-commit 自动检查
git commit -m "message"

# 快速检查：使用 Makefile
make fmt lint test

# CI/CD：两者都用
pre-commit run --all-files
make test
```

---

## 📚 相关文档

- **scripts/fix-precommit.sh** - 自动修复脚本
- **scripts/fix-precommit.sh.zh-CN** - 脚本详细注释版本
- **README.md** - 项目主文档
- **ceph-exporter/DEVELOPMENT.md** - 开发指南

---

## 🆘 获取帮助

### 遇到问题时

```bash
# 1. 查看详细错误
pre-commit run --verbose --all-files

# 4. 清理缓存重试
pre-commit clean
pre-commit install
pre-commit run --all-files

# 5. 使用 Makefile 替代
make fmt lint test
```

### 常用资源

- Pre-commit 官方文档: https://pre-commit.com/
- Pre-commit hooks 仓库: https://github.com/pre-commit/pre-commit-hooks
- Golangci-lint 文档: https://golangci-lint.run/

---

## ✅ 快速检查清单

安装和配置:
- [ ] 已安装 Python 和 pip
- [ ] 已安装 pre-commit (`pip install pre-commit`)
- [ ] 已配置 pip 镜像源（国内用户）
- [ ] 已安装 git hooks (`pre-commit install`)
- [ ] 已运行首次检查 (`pre-commit run --all-files`)

日常使用:
- [ ] 提交前代码已格式化 (`make fmt`)
- [ ] Pre-commit 检查通过
- [ ] 如遇问题，已查看错误信息并修复
- [ ] 必要时使用 SKIP 跳过特定检查

---

**文档版本**: 1.0
**最后更新**: 2026-03-21
**建议**: 首次使用请运行 `./scripts/fix-precommit.sh`
