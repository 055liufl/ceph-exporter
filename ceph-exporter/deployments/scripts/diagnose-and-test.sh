#!/bin/bash
# =============================================================================
# 完整的服务诊断和测试脚本
# =============================================================================

set -e

cd /home/lfl/ceph-exporter/ceph-exporter/deployments

echo "╔══════════════════════════════════════════════════════════════════════════╗"
echo "║              服务诊断和测试                                               ║"
echo "╚══════════════════════════════════════════════════════════════════════════╝"
echo ""

echo "[1/10] 检查容器状态..."
sudo docker ps --format "table {{.Names}}\t{{.Status}}"
echo ""

echo "[2/10] 检查 Elasticsearch 健康状态..."
if curl -s http://localhost:9200/_cluster/health?pretty; then
    echo "✅ Elasticsearch 正常"
else
    echo "❌ Elasticsearch 未响应"
    echo "查看日志:"
    sudo docker logs elasticsearch --tail 20
fi
echo ""

echo "[3/10] 检查 Elasticsearch 索引..."
curl -s http://localhost:9200/_cat/indices?v || echo "❌ 无法获取索引列表"
echo ""

echo "[4/10] 检查 ceph-exporter 状态..."
if curl -s http://localhost:9128/health > /dev/null 2>&1; then
    echo "✅ ceph-exporter 正常"
else
    echo "❌ ceph-exporter 未响应"
    echo "查看日志:"
    sudo docker logs ceph-exporter --tail 20
fi
echo ""

echo "[5/10] 检查 Logstash 状态..."
sudo docker logs logstash --tail 10 | grep -E "Pipeline started|error" || echo "查看完整日志"
echo ""

echo "[6/10] 检查 Jaeger 状态..."
if curl -s http://localhost:16686 > /dev/null 2>&1; then
    echo "✅ Jaeger 正常"
else
    echo "❌ Jaeger 未响应"
fi
echo ""

echo "[7/10] 生成测试数据..."
echo "发送 10 个请求到 ceph-exporter..."
for i in {1..10}; do
    if curl -s http://localhost:9128/metrics > /dev/null 2>&1; then
        echo "  请求 $i - 成功"
    else
        echo "  请求 $i - 失败"
    fi
    sleep 1
done
echo ""

echo "[8/10] 等待日志推送到 Elasticsearch (10秒)..."
sleep 10
echo ""

echo "[9/10] 再次检查 Elasticsearch 索引..."
curl -s http://localhost:9200/_cat/indices?v || echo "❌ 无法获取索引列表"
echo ""

echo "[10/10] 查询日志数据..."
if curl -s "http://localhost:9200/ceph-exporter-*/_search?size=1&pretty" | head -30; then
    echo "✅ 找到日志数据"
else
    echo "❌ 未找到日志数据"
fi
echo ""

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo " 诊断总结"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# 检查各个服务
ES_STATUS=$(curl -s http://localhost:9200/_cluster/health 2>&1 | grep -q "status" && echo "✅" || echo "❌")
CEPH_STATUS=$(curl -s http://localhost:9128/health 2>&1 > /dev/null && echo "✅" || echo "❌")
JAEGER_STATUS=$(curl -s http://localhost:16686 2>&1 > /dev/null && echo "✅" || echo "❌")
KIBANA_STATUS=$(curl -s http://localhost:5601 2>&1 > /dev/null && echo "✅" || echo "❌")

echo "服务状态:"
echo "  Elasticsearch: $ES_STATUS"
echo "  ceph-exporter: $CEPH_STATUS"
echo "  Jaeger:        $JAEGER_STATUS"
echo "  Kibana:        $KIBANA_STATUS"
echo ""

echo "访问地址:"
echo "  Kibana:  http://localhost:5601"
echo "  Jaeger:  http://localhost:16686"
echo "  ceph-exporter: http://localhost:9128"
echo ""

if [ "$ES_STATUS" = "❌" ]; then
    echo "⚠️  Elasticsearch 未运行，请检查:"
    echo "  1. sudo docker logs elasticsearch"
    echo "  2. 确认权限已修复: sudo chown -R 1000:1000 data/elasticsearch"
    echo "  3. 重启: sudo docker-compose -f docker-compose-lightweight-full.yml restart elasticsearch"
fi

if [ "$CEPH_STATUS" = "❌" ]; then
    echo "⚠️  ceph-exporter 未运行，请检查:"
    echo "  1. sudo docker logs ceph-exporter"
    echo "  2. 确认 Ceph 配置存在: ls -la data/ceph-demo/config/"
    echo "  3. 重启: sudo docker-compose -f docker-compose-lightweight-full.yml restart ceph-exporter"
fi

echo ""
echo "╔══════════════════════════════════════════════════════════════════════════╗"
echo "║                    诊断完成                                               ║"
echo "╚══════════════════════════════════════════════════════════════════════════╝"
