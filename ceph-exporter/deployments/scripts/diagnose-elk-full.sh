#!/bin/bash
# =============================================================================
# 完整的 ELK 日志链路诊断脚本
# =============================================================================

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}╔══════════════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║              ELK 日志链路完整诊断                                         ║${NC}"
echo -e "${BLUE}╚══════════════════════════════════════════════════════════════════════════╝${NC}"
echo ""

# 1. 检查 ceph-exporter 容器
echo -e "${YELLOW}[1/10] 检查 ceph-exporter 容器状态...${NC}"
if docker ps | grep -q ceph-exporter; then
    echo -e "${GREEN}✓ ceph-exporter 容器正在运行${NC}"
    docker ps --filter "name=ceph-exporter" --format "  状态: {{.Status}}"
else
    echo -e "${RED}✗ ceph-exporter 容器未运行${NC}"
    echo "请先启动 ceph-exporter"
    exit 1
fi
echo ""

# 2. 检查配置文件
echo -e "${YELLOW}[2/10] 检查 ceph-exporter 配置...${NC}"
if [ -f "configs/ceph-exporter.yaml" ]; then
    echo "ELK 配置:"
    grep -A 4 "enable_elk:" configs/ceph-exporter.yaml | grep -E "enable_elk|logstash_url|logstash_protocol"

    if grep -q "enable_elk: true" configs/ceph-exporter.yaml; then
        echo -e "${GREEN}✓ ELK 集成已启用${NC}"
    else
        echo -e "${RED}✗ ELK 集成未启用${NC}"
        echo "运行: ./scripts/switch-logging-mode.sh direct"
        exit 1
    fi
else
    echo -e "${RED}✗ 配置文件不存在${NC}"
    exit 1
fi
echo ""

# 3. 检查 ceph-exporter 日志
echo -e "${YELLOW}[3/10] 检查 ceph-exporter 启动日志...${NC}"
if docker logs ceph-exporter 2>&1 | grep -q "ELK 集成已启用\|ELK.*enabled"; then
    echo -e "${GREEN}✓ ceph-exporter 已启用 ELK 集成${NC}"
    docker logs ceph-exporter 2>&1 | grep -i "elk\|logstash" | tail -3
else
    echo -e "${RED}✗ ceph-exporter 未启用 ELK 或未重启${NC}"
    echo "需要重启 ceph-exporter:"
    echo "  docker-compose restart ceph-exporter"
fi
echo ""

# 4. 检查 Logstash 容器
echo -e "${YELLOW}[4/10] 检查 Logstash 容器状态...${NC}"
if docker ps | grep -q logstash; then
    echo -e "${GREEN}✓ Logstash 容器正在运行${NC}"

    if docker logs logstash 2>&1 | grep -q "Successfully started Logstash API endpoint"; then
        echo -e "${GREEN}✓ Logstash 已完全启动${NC}"
    else
        echo -e "${YELLOW}⚠ Logstash 可能还在启动中${NC}"
    fi
else
    echo -e "${RED}✗ Logstash 容器未运行${NC}"
    exit 1
fi
echo ""

# 5. 检查端口监听
echo -e "${YELLOW}[5/10] 检查端口监听...${NC}"
if netstat -tuln 2>/dev/null | grep -q ":5000"; then
    echo -e "${GREEN}✓ 端口 5000 (TCP input) 正在监听${NC}"
else
    echo -e "${RED}✗ 端口 5000 未监听${NC}"
    echo "Logstash TCP 输入端口未就绪"
fi

if netstat -tuln 2>/dev/null | grep -q ":9200"; then
    echo -e "${GREEN}✓ 端口 9200 (Elasticsearch) 正在监听${NC}"
else
    echo -e "${RED}✗ 端口 9200 未监听${NC}"
fi
echo ""

# 6. 测试 Logstash TCP 输入
echo -e "${YELLOW}[6/10] 测试 Logstash TCP 输入...${NC}"
TEST_LOG='{"message":"Diagnostic test log","level":"info","service":"ceph-exporter","timestamp":"'$(date -Iseconds)'"}'
echo "$TEST_LOG" | nc -w 2 localhost 5000 2>/dev/null
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ 成功发送测试日志到 Logstash${NC}"
    echo "等待 3 秒让日志处理..."
    sleep 3
else
    echo -e "${RED}✗ 无法连接到 Logstash TCP 端口 5000${NC}"
fi
echo ""

# 7. 检查 Elasticsearch 状态
echo -e "${YELLOW}[7/10] 检查 Elasticsearch 状态...${NC}"
ES_HEALTH=$(curl -s http://localhost:9200/_cluster/health 2>/dev/null)
if [ -n "$ES_HEALTH" ]; then
    echo -e "${GREEN}✓ Elasticsearch 正常响应${NC}"
    echo "$ES_HEALTH" | grep -o '"status":"[^"]*"'
else
    echo -e "${RED}✗ Elasticsearch 未响应${NC}"
    exit 1
fi
echo ""

# 8. 检查 Elasticsearch 索引
echo -e "${YELLOW}[8/10] 检查 Elasticsearch 索引...${NC}"
INDICES=$(curl -s http://localhost:9200/_cat/indices?v 2>/dev/null)
if echo "$INDICES" | grep -q "ceph"; then
    echo -e "${GREEN}✓ 找到 ceph 相关索引:${NC}"
    echo "$INDICES" | grep ceph

    # 查询文档数量
    DOC_COUNT=$(curl -s http://localhost:9200/ceph-*/_count 2>/dev/null | grep -o '"count":[0-9]*' | cut -d: -f2)
    if [ -n "$DOC_COUNT" ]; then
        echo -e "${GREEN}✓ 索引中有 $DOC_COUNT 条文档${NC}"
    fi
else
    echo -e "${RED}✗ 未找到 ceph 相关索引${NC}"
    echo ""
    echo "所有索引:"
    echo "$INDICES"
fi
echo ""

# 9. 检查 Logstash 日志
echo -e "${YELLOW}[9/10] 检查 Logstash 最近日志...${NC}"
echo "最近 10 行日志:"
docker logs logstash 2>&1 | tail -10
echo ""

# 10. 检查网络连通性
echo -e "${YELLOW}[10/10] 检查容器网络连通性...${NC}"
if docker exec ceph-exporter ping -c 1 logstash >/dev/null 2>&1; then
    echo -e "${GREEN}✓ ceph-exporter 可以 ping 通 logstash${NC}"
else
    echo -e "${RED}✗ ceph-exporter 无法 ping 通 logstash${NC}"
    echo "可能的网络问题"
fi
echo ""

# 总结和建议
echo -e "${BLUE}╔══════════════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║  诊断总结和建议                                                           ║${NC}"
echo -e "${BLUE}╚══════════════════════════════════════════════════════════════════════════╝${NC}"
echo ""

# 判断问题
if ! docker logs ceph-exporter 2>&1 | grep -q "ELK 集成已启用\|ELK.*enabled"; then
    echo -e "${YELLOW}问题: ceph-exporter 未启用 ELK 或未重启${NC}"
    echo ""
    echo "解决方案:"
    echo "  1. 确认配置已修改: cat configs/ceph-exporter.yaml | grep enable_elk"
    echo "  2. 重启 ceph-exporter: docker-compose restart ceph-exporter"
    echo "  3. 查看启动日志: docker logs ceph-exporter | grep -i elk"
    echo ""
elif ! netstat -tuln 2>/dev/null | grep -q ":5000"; then
    echo -e "${YELLOW}问题: Logstash TCP 端口 5000 未监听${NC}"
    echo ""
    echo "解决方案:"
    echo "  1. 检查 Logstash 配置: cat logstash/logstash.conf"
    echo "  2. 重启 Logstash: docker-compose -f docker-compose-lightweight-full.yml restart logstash"
    echo "  3. 查看 Logstash 日志: docker logs logstash | grep -i \"tcp\|5000\""
    echo ""
elif ! curl -s http://localhost:9200/_cat/indices?v 2>/dev/null | grep -q "ceph"; then
    echo -e "${YELLOW}问题: Elasticsearch 中没有 ceph 索引${NC}"
    echo ""
    echo "可能原因:"
    echo "  1. ceph-exporter 还没有生成日志"
    echo "  2. 日志没有成功发送到 Logstash"
    echo "  3. Logstash 处理日志时出错"
    echo ""
    echo "调试步骤:"
    echo "  1. 查看 ceph-exporter 日志:"
    echo "     docker logs ceph-exporter | tail -50"
    echo ""
    echo "  2. 手动发送测试日志:"
    echo "     echo '{\"message\":\"test\",\"level\":\"info\"}' | nc localhost 5000"
    echo ""
    echo "  3. 立即检查 Elasticsearch:"
    echo "     curl http://localhost:9200/_cat/indices?v"
    echo ""
    echo "  4. 查看 Logstash 处理日志:"
    echo "     docker logs logstash | tail -50"
    echo ""
    echo "  5. 检查 Logstash 配置语法:"
    echo "     docker exec logstash /usr/share/logstash/bin/logstash \\"
    echo "       --config.test_and_exit \\"
    echo "       -f /usr/share/logstash/pipeline/logstash.conf"
    echo ""
else
    echo -e "${GREEN}✓ 所有检查通过！${NC}"
    echo ""
    echo "在 Kibana 中查看日志:"
    echo "  1. 访问 http://localhost:5601"
    echo "  2. Stack Management > 索引模式 > 创建索引模式"
    echo "  3. 输入: ceph-exporter-* 或 ceph-*"
    echo "  4. 选择时间字段: @timestamp"
    echo "  5. Discover > 查看日志"
    echo ""
fi

echo "详细文档: deployments/LOGSTASH-OOM-FIX.md"
echo ""
