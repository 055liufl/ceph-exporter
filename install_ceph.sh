#!/bin/bash
# Ceph Nautilus 开发库安装脚本
# 适用于 CentOS 7 系统
set -e

echo "=== 安装 Ceph Nautilus 开发库 ==="

# 步骤 1: 安装 Ceph 仓库
echo "1. 安装 Ceph 仓库配置..."
sudo yum install -y centos-release-ceph-nautilus

# 步骤 2: 安装开发库
echo "2. 安装 Ceph 开发库..."
sudo yum install -y librados-devel librbd-devel

# 步骤 3: 验证安装
echo "3. 验证安装..."
if [ -f /usr/include/rados/librados.h ]; then
    echo "✓ librados.h 已安装"
else
    echo "✗ librados.h 未找到"
    exit 1
fi

if [ -f /usr/lib64/librados.so ]; then
    echo "✓ librados.so 已安装"
else
    echo "✗ librados.so 未找到"
    exit 1
fi

echo ""
echo "=== 安装完成 ==="
echo "已安装的包:"
rpm -qa | grep -E "librados-devel|librbd-devel"

echo ""
echo "现在可以运行 pre-commit 检查:"
echo "  cd /home/lfl/ceph-exporter"
echo "  source ~/.bashrc"
echo "  pre-commit run --all-files"
