#!/bin/bash
# =============================================================================
# 使用 Ceph Demo 独立配置测试
# =============================================================================

set -e

echo "=========================================="
echo "使用 Ceph Demo 独立配置"
echo "=========================================="
echo ""
echo "此配置使用固定 IP 地址 172.20.0.10"
echo "这应该能解决 MON_IP 的问题"
echo ""
read -p "确认继续? (输入 yes): " confirm

if [ "$confirm" != "yes" ]; then
    echo "操作已取消"
    exit 0
fi

cd /home/lfl/ceph-exporter/ceph-exporter/deployments

echo ""
echo "步骤 1: 停止所有服务..."
docker-compose -f docker-compose-lightweight-full.yml down 2>/dev/null || true
docker-compose -f docker-compose-ceph-demo.yml down 2>/dev/null || true

echo ""
echo "步骤 2: 强制删除所有数据卷..."
docker volume prune -f

echo ""
echo "步骤 3: 启动 Ceph Demo（使用固定 IP）..."
docker-compose -f docker-compose-ceph-demo.yml up -d

echo ""
echo "步骤 4: 等待 5 分钟让 Ceph 完全初始化..."
for i in {1..300}; do
    echo -n "."
    sleep 1
    if [ $((i % 30)) -eq 0 ]; then
        echo " ${i}秒"
    fi
done
echo ""

echo ""
echo "步骤 5: 检查 ceph-demo 状态..."
docker ps --filter name=ceph-demo

echo ""
echo "步骤 6: 查看日志..."
docker logs ceph-demo --tail 50

echo ""
echo "步骤 7: 检查配置文件..."
docker exec ceph-demo cat /etc/ceph/ceph.conf 2>&1 || echo "⚠️  无法读取"

echo ""
echo "步骤 8: 测试 Ceph 集群..."
timeout 30 docker exec ceph-demo ceph -s 2>&1 || echo "⚠️  Ceph 命令超时"

echo ""
echo "=========================================="
echo "Ceph Demo 测试完成"
echo "=========================================="
echo ""
echo "如果 Ceph 正常，请执行:"
echo "  cd /home/lfl/ceph-exporter/ceph-exporter/deployments"
echo "  sudo docker-compose -f docker-compose-lightweight-full.yml up -d"
echo ""
