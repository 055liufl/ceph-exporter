#!/bin/bash
# =============================================================================
# 使用已运行的 Ceph Demo 启动完整监控栈
# =============================================================================

set -e

# 获取脚本所在目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DEPLOY_DIR="$(dirname "$SCRIPT_DIR")"

echo "=========================================="
echo "启动完整监控栈"
echo "=========================================="
echo ""
echo "Ceph Demo 已经在运行，现在启动其他服务"
echo ""

cd "$DEPLOY_DIR"

echo "步骤 1: 检查 Ceph Demo 状态..."
docker ps --filter name=ceph-demo

echo ""
echo "步骤 2: 测试 Ceph 集群..."
docker exec ceph-demo ceph -s

echo ""
echo "步骤 3: 停止 ceph-demo 独立配置..."
docker-compose -f docker-compose-ceph-demo.yml down

echo ""
echo "步骤 4: 启动完整监控栈（包括 Ceph Demo）..."
docker-compose -f docker-compose-lightweight-full.yml up -d

echo ""
echo "步骤 5: 等待 1 分钟..."
sleep 60

echo ""
echo "步骤 6: 检查所有容器..."
docker ps --format "table {{.Names}}\t{{.Status}}"

echo ""
echo "步骤 7: 检查 ceph-exporter..."
docker logs ceph-exporter --tail 30

echo ""
echo "步骤 8: 测试健康端点..."
curl -s http://localhost:9128/health && echo "" || echo "健康端点尚未就绪"

echo ""
echo "步骤 9: 测试指标端点..."
curl -s http://localhost:9128/metrics | head -20 || echo "指标端点尚未就绪"

echo ""
echo "=========================================="
echo "部署完成"
echo "=========================================="
echo ""
echo "访问地址:"
echo "  Ceph Dashboard:  http://localhost:8080"
echo "  Ceph Exporter:   http://localhost:9128/metrics"
echo "  Prometheus:      http://localhost:9090"
echo "  Grafana:         http://localhost:3000 (admin/admin)"
echo "  Elasticsearch:   http://localhost:9200"
echo "  Kibana:          http://localhost:5601"
echo "  Jaeger:          http://localhost:16686"
echo ""
