#!/bin/bash
# =============================================================================
# 彻底清理 Ceph 数据卷内容
# =============================================================================

set -e

echo "=========================================="
echo "彻底清理 Ceph 数据卷"
echo "=========================================="
echo ""
echo "此脚本将:"
echo "  1. 启动一个临时容器"
echo "  2. 挂载 Ceph 数据卷"
echo "  3. 手动删除卷内的所有文件"
echo "  4. 然后重新启动 Ceph Demo"
echo ""
read -p "确认继续? (输入 yes): " confirm

if [ "$confirm" != "yes" ]; then
    echo "操作已取消"
    exit 0
fi

cd /home/lfl/ceph-exporter/ceph-exporter/deployments

echo ""
echo "步骤 1: 停止所有服务..."
docker-compose -f docker-compose-ceph-demo.yml down 2>/dev/null || true
docker-compose -f docker-compose-lightweight-full.yml down 2>/dev/null || true

echo ""
echo "步骤 2: 使用临时容器清理 ceph-demo-data 卷..."
docker run --rm -v deployments_ceph-demo-data:/data alpine sh -c "rm -rf /data/* /data/.*" 2>/dev/null || echo "数据卷已清理"

echo ""
echo "步骤 3: 使用临时容器清理 ceph-demo-config 卷..."
docker run --rm -v deployments_ceph-demo-config:/config alpine sh -c "rm -rf /config/* /config/.*" 2>/dev/null || echo "配置卷已清理"

echo ""
echo "步骤 4: 验证卷已清空..."
echo "ceph-demo-data 内容:"
docker run --rm -v deployments_ceph-demo-data:/data alpine ls -la /data || echo "卷为空"
echo ""
echo "ceph-demo-config 内容:"
docker run --rm -v deployments_ceph-demo-config:/config alpine ls -la /config || echo "卷为空"

echo ""
echo "步骤 5: 启动 Ceph Demo..."
docker-compose -f docker-compose-ceph-demo.yml up -d

echo ""
echo "步骤 6: 等待 5 分钟..."
for i in {1..300}; do
    echo -n "."
    sleep 1
    if [ $((i % 30)) -eq 0 ]; then
        echo " ${i}秒"
    fi
done
echo ""

echo ""
echo "步骤 7: 检查状态..."
docker ps --filter name=ceph-demo

echo ""
echo "步骤 8: 查看日志..."
docker logs ceph-demo --tail 50

echo ""
echo "步骤 9: 测试 Ceph..."
timeout 30 docker exec ceph-demo ceph -s 2>&1 || echo "⚠️  Ceph 命令超时"

echo ""
echo "=========================================="
echo "清理完成"
echo "=========================================="
