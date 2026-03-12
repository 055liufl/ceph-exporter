#!/bin/bash
# =============================================================================
# ceph-exporter 日志推送到 ELK 端到端测试脚本
# =============================================================================

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}╔══════════════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║          ceph-exporter 日志推送到 ELK - 端到端测试                       ║${NC}"
echo -e "${BLUE}╚══════════════════════════════════════════════════════════════════════════╝${NC}"
echo ""

# 检查是否在 deployments 目录
if [ ! -f "docker-compose.yml" ]; then
    echo -e "${RED}错误: 请在 deployments 目录下运行此脚本${NC}"
    exit 1
fi

# 1. 检查 Logstash 状态
echo -e "${YELLOW}[1/8] 检查 Logstash 状态...${NC}"
if docker ps | grep -q logstash; then
    echo -e "${GREEN}✓ Logstash 容器正在运行${NC}"

    if docker logs logstash 2>&1 | grep -q "Successfully started Logstash API endpoint"; then
        echo -e "${GREEN}✓ Logstash 已完全启动${NC}"
    else
        echo -e "${YELLOW}⚠ Logstash 可能还在启动中，请等待...${NC}"
        sleep 5
    fi
else
    echo -e "${RED}✗ Logstash 未运行，请先启动 Logstash${NC}"
    exit 1
fi
echo ""

# 2. 检查 Elasticsearch 状态
echo -e "${YELLOW}[2/8] 检查 Elasticsearch 状态...${NC}"
if curl -s http://localhost:9200/_cluster/health > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Elasticsearch 正常运行${NC}"
else
    echo -e "${RED}✗ Elasticsearch 未响应${NC}"
    exit 1
fi
echo ""

# 3. 检查配置
echo -e "${YELLOW}[3/8] 检查 ceph-exporter 配置...${NC}"
if grep -q "enable_elk: true" configs/ceph-exporter.yaml; then
    echo -e "${GREEN}✓ ELK 集成已启用${NC}"
    LOGSTASH_URL=$(grep "logstash_url:" configs/ceph-exporter.yaml | awk '{print $2}' | tr -d '"')
    echo -e "  Logstash URL: $LOGSTASH_URL"
else
    echo -e "${RED}✗ ELK 集成未启用${NC}"
    exit 1
fi
echo ""

# 4. 重启 ceph-exporter
echo -e "${YELLOW}[4/8] 重启 ceph-exporter...${NC}"
docker-compose restart ceph-exporter
echo -e "${GREEN}✓ ceph-exporter 已重启${NC}"
echo ""

# 5. 等待服务启动
echo -e "${YELLOW}[5/8] 等待 ceph-exporter 启动（10秒）...${NC}"
sleep 10
echo -e "${GREEN}✓ 等待完成${NC}"
echo ""

# 6. 发送测试日志
echo -e "${YELLOW}[6/8] 发送测试日志到 Logstash...${NC}"
TEST_LOG='{"message":"Test log from end-to-end script","level":"info","service":"ceph-exporter","component":"test","timestamp":"'$(date -Iseconds)'"}'
echo "$TEST_LOG" | nc localhost 5000
echo -e "${GREEN}✓ 测试日志已发送${NC}"
echo ""

# 7. 等待日志处理
echo -e "${YELLOW}[7/8] 等待日志处理（5秒）...${NC}"
sleep 5
echo ""

# 8. 检查 Elasticsearch 索引
echo -e "${YELLOW}[8/8] 检查 Elasticsearch 索引...${NC}"
INDICES=$(curl -s http://localhost:9200/_cat/indices?v | grep ceph)
if [ -n "$INDICES" ]; then
    echo -e "${GREEN}✓ 找到 ceph-exporter 索引:${NC}"
    echo "$INDICES"
    echo ""

    # 查询文档数量
    DOC_COUNT=$(curl -s http://localhost:9200/ceph-exporter-*/_count | grep -o '"count":[0-9]*' | cut -d: -f2)
    echo -e "${GREEN}✓ 索引中有 $DOC_COUNT 条日志${NC}"
    echo ""

    # 显示最新的一条日志
    echo -e "${BLUE}最新日志示例:${NC}"
    curl -s http://localhost:9200/ceph-exporter-*/_search?size=1&sort=@timestamp:desc | jq '.hits.hits[0]._source' 2>/dev/null || echo "需要安装 jq 才能格式化显示"
else
    echo -e "${YELLOW}⚠ 未找到 ceph-exporter 索引${NC}"
    echo ""
    echo "可能的原因:"
    echo "  1. 日志还在处理中，请等待几秒后重试"
    echo "  2. ceph-exporter 未成功连接到 Logstash"
    echo "  3. Logstash 配置有问题"
    echo ""
    echo "诊断命令:"
    echo "  查看 ceph-exporter 日志: docker logs ceph-exporter | tail -20"
    echo "  查看 Logstash 日志: docker logs logstash | tail -20"
    echo "  手动测试: echo '{\"test\":\"log\"}' | nc localhost 5000"
fi
echo ""

# 总结
echo -e "${BLUE}╔══════════════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║  下一步操作                                                               ║${NC}"
echo -e "${BLUE}╚══════════════════════════════════════════════════════════════════════════╝${NC}"
echo ""
echo "1. 在 Kibana 中创建索引模式:"
echo "   访问: http://localhost:5601"
echo "   Stack Management > 索引模式 > 创建索引模式"
echo "   索引模式: ceph-exporter-*"
echo "   时间字段: @timestamp"
echo ""
echo "2. 查看日志:"
echo "   Kibana > Discover > 选择 ceph-exporter-* 索引"
echo ""
echo "3. 实时监控日志:"
echo "   docker logs -f ceph-exporter"
echo ""
echo "4. 查看 Elasticsearch 中的日志:"
echo "   curl http://localhost:9200/ceph-exporter-*/_search?pretty"
echo ""
