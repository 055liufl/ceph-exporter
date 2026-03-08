#!/bin/bash
# Pre-commit 快速修复脚本
# 用于解决 SSL 连接问题

set -e

echo "=========================================="
echo "Pre-commit SSL 问题快速修复"
echo "=========================================="
echo ""

# 检查操作系统
if [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "win32" ]]; then
    IS_WINDOWS=true
    PIP_CONFIG_DIR="$APPDATA/pip"
    PIP_CONFIG_FILE="$PIP_CONFIG_DIR/pip.ini"
else
    IS_WINDOWS=false
    PIP_CONFIG_DIR="$HOME/.pip"
    PIP_CONFIG_FILE="$PIP_CONFIG_DIR/pip.conf"
fi

echo "步骤 1: 配置 pip 使用国内镜像源"
echo "----------------------------------------"

# 创建配置目录
mkdir -p "$PIP_CONFIG_DIR"

# 写入配置
cat > "$PIP_CONFIG_FILE" << 'EOF'
[global]
index-url = https://pypi.tuna.tsinghua.edu.cn/simple
trusted-host = pypi.tuna.tsinghua.edu.cn

[install]
trusted-host = pypi.tuna.tsinghua.edu.cn
EOF

echo "✓ pip 配置已更新: $PIP_CONFIG_FILE"
echo ""

echo "步骤 2: 验证 pip 配置"
echo "----------------------------------------"
pip config list
echo ""

echo "步骤 3: 清理 pre-commit 缓存"
echo "----------------------------------------"
pre-commit clean || echo "警告: pre-commit clean 失败，继续..."
echo "✓ 缓存已清理"
echo ""

echo "步骤 4: 选择配置文件"
echo "----------------------------------------"
echo "请选择要使用的配置:"
echo "  1) 完整配置 (.pre-commit-config.yaml) - 包含所有检查"
echo "  2) 简化配置 (.pre-commit-config-simple.yaml) - 仅 Go 检查"
echo "  3) 跳过 pre-commit，使用 Makefile"
echo ""
read -p "请输入选项 (1/2/3) [默认: 2]: " choice
choice=${choice:-2}

case $choice in
    1)
        echo "使用完整配置..."
        CONFIG_FILE=".pre-commit-config.yaml"
        ;;
    2)
        echo "使用简化配置..."
        CONFIG_FILE=".pre-commit-config-simple.yaml"
        ;;
    3)
        echo "跳过 pre-commit，使用 Makefile..."
        echo ""
        echo "运行以下命令进行代码检查:"
        echo "  make fmt    # 格式化代码"
        echo "  make lint   # 静态检查"
        echo "  make test   # 运行测试"
        echo ""
        exit 0
        ;;
    *)
        echo "无效选项，使用简化配置"
        CONFIG_FILE=".pre-commit-config-simple.yaml"
        ;;
esac

echo ""
echo "步骤 5: 安装 pre-commit hooks"
echo "----------------------------------------"
if [ "$CONFIG_FILE" = ".pre-commit-config-simple.yaml" ]; then
    pre-commit install --config "$CONFIG_FILE"
    echo "✓ Hooks 已安装（使用简化配置）"
else
    pre-commit install --install-hooks
    echo "✓ Hooks 已安装（使用完整配置）"
fi
echo ""

echo "步骤 6: 测试运行"
echo "----------------------------------------"
echo "运行 trailing-whitespace 检查作为测试..."
if [ "$CONFIG_FILE" = ".pre-commit-config-simple.yaml" ]; then
    pre-commit run --config "$CONFIG_FILE" go-fmt --all-files || true
else
    pre-commit run trailing-whitespace --all-files || true
fi
echo ""

echo "=========================================="
echo "修复完成！"
echo "=========================================="
echo ""
echo "下一步:"
if [ "$CONFIG_FILE" = ".pre-commit-config-simple.yaml" ]; then
    echo "  运行: pre-commit run --config .pre-commit-config-simple.yaml --all-files"
else
    echo "  运行: pre-commit run --all-files"
fi
echo ""
echo "或使用 Makefile:"
echo "  make pre-commit"
echo ""
echo "如果仍有问题，请查看: PRE_COMMIT_SSL_FIX.md"
echo ""
