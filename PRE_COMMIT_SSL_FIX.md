# Pre-commit SSL 问题修复指南

**最后更新**: 2026-03-15

---

## 问题描述

在使用 pre-commit 时可能遇到 SSL 证书验证错误，常见错误信息：

```
SSL: CERTIFICATE_VERIFY_FAILED
URLError: <urlopen error [SSL: CERTIFICATE_VERIFY_FAILED]>
```

---

## 快速修复

### 方法 1: 使用自动修复脚本（推荐）

```bash
cd <project-root>
./scripts/fix-precommit.sh
```

脚本会自动：
1. 配置 pip 使用国内镜像源
2. 清理 pre-commit 缓存
3. 重新安装 hooks
4. 测试运行

### 方法 2: 手动配置 pip 镜像源

#### Linux/Mac:

```bash
# 创建配置目录
mkdir -p ~/.pip

# 创建配置文件
cat > ~/.pip/pip.conf << 'CONF'
[global]
index-url = https://pypi.tuna.tsinghua.edu.cn/simple
trusted-host = pypi.tuna.tsinghua.edu.cn

[install]
trusted-host = pypi.tuna.tsinghua.edu.cn
CONF

# 验证配置
pip config list
```

#### Windows:

```powershell
# 创建配置目录
New-Item -Path "$env:APPDATA\pip" -ItemType Directory -Force

# 创建配置文件
@"
[global]
index-url = https://pypi.tuna.tsinghua.edu.cn/simple
trusted-host = pypi.tuna.tsinghua.edu.cn

[install]
trusted-host = pypi.tuna.tsinghua.edu.cn
"@ | Out-File -FilePath "$env:APPDATA\pip\pip.ini" -Encoding ASCII

# 验证配置
pip config list
```

### 方法 3: 使用简化配置

如果完整配置仍有问题，使用简化配置（仅 Go 检查）：

```bash
# 安装简化配置
pre-commit install --config .pre-commit-config-simple.yaml

# 运行检查
pre-commit run --config .pre-commit-config-simple.yaml --all-files
```

### 方法 4: 跳过 pre-commit，使用 Makefile

```bash
# 格式化代码
make fmt

# 静态检查
make lint

# 运行测试
make test

# 或一次性运行所有检查
make pre-commit
```

---

## 可用的 pip 镜像源

如果清华源不可用，可以尝试其他镜像源：

### 1. 清华大学镜像（推荐）
```ini
[global]
index-url = https://pypi.tuna.tsinghua.edu.cn/simple
trusted-host = pypi.tuna.tsinghua.edu.cn
```

### 2. 阿里云镜像
```ini
[global]
index-url = https://mirrors.aliyun.com/pypi/simple/
trusted-host = mirrors.aliyun.com
```

### 3. 中科大镜像
```ini
[global]
index-url = https://pypi.mirrors.ustc.edu.cn/simple/
trusted-host = pypi.mirrors.ustc.edu.cn
```

### 4. 豆瓣镜像
```ini
[global]
index-url = https://pypi.douban.com/simple/
trusted-host = pypi.douban.com
```

---

## 常见问题

### Q1: 配置后仍然报 SSL 错误

**解决方案**:

```bash
# 1. 清理 pre-commit 缓存
pre-commit clean

# 2. 清理 pip 缓存
pip cache purge

# 3. 重新安装 pre-commit
pip uninstall pre-commit
pip install pre-commit

# 4. 重新安装 hooks
pre-commit install --install-hooks
```

### Q2: pre-commit 下载速度很慢

**解决方案**:

使用简化配置或直接使用 Makefile：

```bash
# 使用 Makefile（不需要下载额外工具）
make fmt lint test
```

### Q3: 某些 hook 总是失败

**解决方案**:

跳过特定的 hook：

```bash
# 跳过特定 hook
SKIP=golangci-lint git commit -m "message"

# 跳过所有 hooks
git commit --no-verify -m "message"
```

### Q4: Windows 环境下配置不生效

**解决方案**:

```powershell
# 检查配置文件位置
pip config list

# 手动指定配置文件
$env:PIP_CONFIG_FILE = "$env:APPDATA\pip\pip.ini"

# 验证
pip config list
```

---

## 验证修复

### 1. 验证 pip 配置

```bash
pip config list
```

预期输出：
```
global.index-url='https://pypi.tuna.tsinghua.edu.cn/simple'
global.trusted-host='pypi.tuna.tsinghua.edu.cn'
install.trusted-host='pypi.tuna.tsinghua.edu.cn'
```

### 2. 测试 pre-commit

```bash
# 运行单个 hook
pre-commit run trailing-whitespace --all-files

# 运行所有 hooks
pre-commit run --all-files
```

### 3. 测试 git commit

```bash
# 创建测试提交
echo "test" >> test.txt
git add test.txt
git commit -m "test: pre-commit"
```

---

## 配置文件说明

### 完整配置 (.pre-commit-config.yaml)

包含所有检查：
- 通用检查（trailing-whitespace, end-of-file-fixer, check-yaml）
- Go 检查（gofmt, goimports, go vet, golangci-lint）

优点：全面的代码质量检查
缺点：需要下载多个工具，可能遇到网络问题

### 简化配置 (.pre-commit-config-simple.yaml)

仅包含 Go 检查：
- gofmt
- goimports
- go vet

优点：下载快，问题少
缺点：检查不够全面

### Makefile

不依赖 pre-commit，直接使用本地工具：
- `make fmt`: 格式化代码
- `make lint`: 静态检查
- `make test`: 运行测试

优点：不需要网络，速度快
缺点：需要手动运行

---

## 推荐方案

### 开发环境（本地）

使用 Makefile（最简单）：
```bash
make fmt lint test
```

### CI/CD 环境

使用完整配置：
```bash
pre-commit run --all-files
```

### 网络受限环境

使用简化配置：
```bash
pre-commit run --config .pre-commit-config-simple.yaml --all-files
```

---

## 相关文档

- **PRECOMMIT_SETUP.md** - Pre-commit 完整设置指南
- **DEVELOPMENT.md** - 开发环境设置
- **ceph-exporter/DEVELOPMENT.md** - 项目开发指南

---

## 获取帮助

如果以上方法都无法解决问题：

1. 检查网络连接
   ```bash
   ping pypi.tuna.tsinghua.edu.cn
   ```

2. 检查 Python 和 pip 版本
   ```bash
   python --version
   pip --version
   ```

3. 查看详细错误信息
   ```bash
   pre-commit run --all-files --verbose
   ```

4. 使用 Makefile 替代
   ```bash
   make pre-commit
   ```

---

**文档版本**: 1.0
**最后更新**: 2026-03-15
