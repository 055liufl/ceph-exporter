#!/bin/bash
# =============================================================================
# 彻底清理 Ceph 数据目录
# =============================================================================
# 注意: 现在使用绑定挂载，数据存储在 ./data/ 目录
# 建议使用 ./deploy.sh clean 命令代替此脚本
# =============================================================================

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DEPLOY_DIR="$(dirname "$SCRIPT_DIR")"

echo "=========================================="
echo "彻底清理 Ceph 数据目录"
echo "=========================================="
echo ""
echo "此脚本将:"
echo "  1. 停止所有服务"
echo "  2. 删除 ./data/ 目录中的所有数据"
echo "  3. 重新初始化数据目录"
echo "  4. 重新启动 Ceph Demo"
echo ""
echo "⚠️  警告: 这将删除所有持久化数据！"
echo ""
read -p "确认继续? (输入 yes): " confirm

if [ "$confirm" != "yes" ]; then
    echo "操作已取消"
    exit 0
fi

cd "$DEPLOY_DIR"

echo ""
echo "步骤 1: 停止所有服务..."
docker-compose -f docker-compose-ceph-demo.yml down 2>/dev/null || true
docker-compose -f docker-compose-lightweight-full.yml down 2>/dev/null || true
docker-compose -f docker-compose-integration-test.yml down 2>/dev/null || true
docker-compose down 2>/dev/null || true

echo ""
echo "步骤 2: 删除数据目录..."
if [ -d "data" ]; then
    echo "删除 data/ 目录..."
    rm -rf data/
    echo "✓ 数据目录已删除"
else
    echo "数据目录不存在，跳过"
fi

echo ""
echo "步骤 3: 重新初始化数据目录..."
./scripts/deploy.sh init

echo ""
echo "步骤 4: 启动 Ceph Demo..."
docker-compose -f docker-compose-ceph-demo.yml up -d

echo ""
echo "步骤 5: 等待 Ceph 初始化（5 分钟）..."
for i in {1..300}; do
    echo -n "."
    sleep 1
    if [ $((i % 30)) -eq 0 ]; then
        echo " ${i}秒"
    fi
done
echo ""

echo ""
echo "步骤 6: 检查状态..."
docker ps --filter name=ceph-demo

echo ""
echo "步骤 7: 查看日志..."
docker logs ceph-demo --tail 50

echo ""
echo "步骤 8: 测试 Ceph..."
timeout 30 docker exec ceph-demo ceph -s 2>&1 || echo "⚠️  Ceph 命令超时"

echo ""
echo "=========================================="
echo "清理完成"
echo "=========================================="
echo ""
echo "数据存储位置: $DEPLOY_DIR/data/"
echo "查看数据: ls -lh data/"
