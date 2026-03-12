#!/bin/bash
# =============================================================================
# 测试 Jaeger 分布式追踪功能
# =============================================================================

set -e

echo "╔══════════════════════════════════════════════════════════════════════════╗"
echo "║              测试 Jaeger 分布式追踪功能                                   ║"
echo "╚══════════════════════════════════════════════════════════════════════════╝"
echo ""

# 检查 Jaeger 是否运行
echo "[1/5] 检查 Jaeger 服务状态..."
if ! docker ps | grep -q jaeger; then
    echo "❌ Jaeger 未运行"
    echo "请先启动: docker-compose -f docker-compose-lightweight-full.yml up -d jaeger"
    exit 1
fi
echo "✅ Jaeger 正在运行"
echo ""

# 检查 Jaeger OTLP 端口
echo "[2/5] 检查 Jaeger OTLP HTTP 端口 (4318)..."
if ! nc -z localhost 4318 2>/dev/null; then
    echo "❌ Jaeger OTLP HTTP 端口 4318 未开放"
    echo "请检查 docker-compose 配置"
    exit 1
fi
echo "✅ OTLP HTTP 端口 4318 已开放"
echo ""

# 检查配置文件
echo "[3/5] 检查追踪配置..."
if grep -q "enabled: false" ../configs/ceph-exporter.yaml; then
    echo "⚠️  追踪功能未启用"
    echo ""
    echo "要启用追踪，请修改 configs/ceph-exporter.yaml:"
    echo ""
    echo "  tracer:"
    echo "    enabled: true"
    echo "    jaeger_url: \"jaeger:4318\""
    echo "    service_name: \"ceph-exporter\""
    echo "    sample_rate: 1.0"
    echo ""
    read -p "是否现在启用追踪功能? (y/n) " -n 1 -r
    echo ""
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        sed -i 's/enabled: false/enabled: true/' ../configs/ceph-exporter.yaml
        echo "✅ 已启用追踪功能"
        echo "⚠️  需要重启 ceph-exporter: docker-compose restart ceph-exporter"
    else
        echo "跳过启用追踪"
        exit 0
    fi
else
    echo "✅ 追踪功能已启用"
fi
echo ""

# 检查 ceph-exporter 是否运行
echo "[4/5] 检查 ceph-exporter 服务状态..."
if ! docker ps | grep -q ceph-exporter; then
    echo "❌ ceph-exporter 未运行"
    echo "请先启动: docker-compose -f docker-compose-lightweight-full.yml up -d ceph-exporter"
    exit 1
fi
echo "✅ ceph-exporter 正在运行"
echo ""

# 生成追踪数据
echo "[5/5] 生成追踪数据..."
echo "发送 10 个请求到 /metrics 端点..."
for i in {1..10}; do
    curl -s http://localhost:9128/metrics > /dev/null
    echo -n "."
    sleep 0.5
done
echo ""
echo "✅ 已发送 10 个请求"
echo ""

# 等待追踪数据上传
echo "等待追踪数据上传到 Jaeger (5秒)..."
sleep 5
echo ""

# 检查 Jaeger 中的追踪数据
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo " 查询 Jaeger 追踪数据"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# 查询服务列表
echo "查询 Jaeger 服务列表..."
SERVICES=$(curl -s "http://localhost:16686/api/services" | jq -r '.data[]' 2>/dev/null || echo "")

if [ -z "$SERVICES" ]; then
    echo "❌ 未找到任何服务"
    echo ""
    echo "可能的原因:"
    echo "  1. 追踪功能未启用 (检查配置文件)"
    echo "  2. ceph-exporter 未重启 (需要重启以加载新配置)"
    echo "  3. 追踪数据尚未上传 (等待更长时间)"
    echo ""
    exit 1
fi

echo "✅ 找到以下服务:"
echo "$SERVICES" | while read service; do
    echo "  - $service"
done
echo ""

# 检查是否有 ceph-exporter 服务
if echo "$SERVICES" | grep -q "ceph-exporter"; then
    echo "✅ 找到 ceph-exporter 服务"
    echo ""

    # 查询最近的追踪
    echo "查询最近的追踪记录..."
    TRACES=$(curl -s "http://localhost:16686/api/traces?service=ceph-exporter&limit=5" | jq -r '.data[].traceID' 2>/dev/null || echo "")

    if [ -z "$TRACES" ]; then
        echo "❌ 未找到追踪记录"
    else
        TRACE_COUNT=$(echo "$TRACES" | wc -l)
        echo "✅ 找到 $TRACE_COUNT 条追踪记录"
        echo ""
        echo "Trace IDs:"
        echo "$TRACES" | head -5 | while read trace_id; do
            echo "  - $trace_id"
        done
    fi
else
    echo "❌ 未找到 ceph-exporter 服务"
    echo ""
    echo "请检查:"
    echo "  1. 追踪功能是否已启用"
    echo "  2. ceph-exporter 是否已重启"
    echo "  3. 配置中的 jaeger_url 是否正确"
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo " 访问 Jaeger UI"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "Jaeger UI: http://localhost:16686"
echo ""
echo "在 Jaeger UI 中:"
echo "  1. Service 下拉框选择: ceph-exporter"
echo "  2. 点击 'Find Traces' 按钮"
echo "  3. 查看追踪详情"
echo ""
echo "╔══════════════════════════════════════════════════════════════════════════╗"
echo "║                    测试完成                                               ║"
echo "╚══════════════════════════════════════════════════════════════════════════╝"
