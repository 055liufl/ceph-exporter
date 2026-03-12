#!/bin/bash
# =============================================================================
# 快速启用 Jaeger 分布式追踪
# =============================================================================

set -e

echo "╔══════════════════════════════════════════════════════════════════════════╗"
echo "║              快速启用 Jaeger 分布式追踪                                   ║"
echo "╚══════════════════════════════════════════════════════════════════════════╝"
echo ""

# 1. 启用追踪配置
echo "[1/4] 启用追踪配置..."
if grep -q "enabled: false" ../configs/ceph-exporter.yaml; then
    sed -i 's/enabled: false/enabled: true/' ../configs/ceph-exporter.yaml
    echo "✅ 已启用追踪功能"
else
    echo "✅ 追踪功能已经启用"
fi
echo ""

# 2. 启动 Jaeger
echo "[2/4] 启动 Jaeger..."
docker-compose -f docker-compose-lightweight-full.yml up -d jaeger
echo "✅ Jaeger 已启动"
echo ""

# 3. 重新构建并启动 ceph-exporter
echo "[3/4] 重新构建并启动 ceph-exporter..."
cd ..
docker build -t ceph-exporter:dev -f deployments/Dockerfile .
cd deployments
docker-compose -f docker-compose-lightweight-full.yml up -d ceph-exporter
echo "✅ ceph-exporter 已重启"
echo ""

# 4. 等待服务就绪
echo "[4/4] 等待服务就绪..."
sleep 5

# 检查服务状态
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo " 服务状态"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

if docker ps | grep -q jaeger; then
    echo "✅ Jaeger 运行中"
else
    echo "❌ Jaeger 未运行"
fi

if docker ps | grep -q ceph-exporter; then
    echo "✅ ceph-exporter 运行中"
else
    echo "❌ ceph-exporter 未运行"
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo " 访问地址"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "Jaeger UI:       http://localhost:16686"
echo "ceph-exporter:   http://localhost:9128/metrics"
echo ""

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo " 下一步"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "1. 生成追踪数据:"
echo "   curl http://localhost:9128/metrics"
echo ""
echo "2. 查看追踪数据:"
echo "   访问 http://localhost:16686"
echo "   选择 Service: ceph-exporter"
echo "   点击 'Find Traces'"
echo ""
echo "3. 运行测试脚本:"
echo "   ./scripts/test-jaeger-tracing.sh"
echo ""

echo "╔══════════════════════════════════════════════════════════════════════════╗"
echo "║                    启用完成                                               ║"
echo "╚══════════════════════════════════════════════════════════════════════════╝"
