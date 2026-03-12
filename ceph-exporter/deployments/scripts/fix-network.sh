#!/bin/bash
# =============================================================================
# 修复 ceph-exporter 和 logstash 网络连接问题
# =============================================================================

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}╔══════════════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║          修复 ceph-exporter 和 logstash 网络连接                         ║${NC}"
echo -e "${BLUE}╚══════════════════════════════════════════════════════════════════════════╝${NC}"
echo ""

echo -e "${YELLOW}问题: ceph-exporter 无法解析 logstash 主机名${NC}"
echo "错误: dial tcp: lookup logstash on 127.0.0.11:53: no such host"
echo ""
echo "原因: ceph-exporter 和 logstash 不在同一个 Docker 网络中"
echo ""

# 1. 检查当前网络
echo -e "${YELLOW}[1/5] 检查容器网络...${NC}"
echo "ceph-exporter 网络:"
docker inspect ceph-exporter --format '{{range $key, $value := .NetworkSettings.Networks}}  - {{$key}}{{end}}'

echo ""
echo "logstash 网络:"
docker inspect logstash --format '{{range $key, $value := .NetworkSettings.Networks}}  - {{$key}}{{end}}'
echo ""

# 2. 获取 logstash 所在的网络
LOGSTASH_NETWORK=$(docker inspect logstash --format '{{range $key, $value := .NetworkSettings.Networks}}{{$key}}{{end}}' | awk '{print $1}')
echo -e "${GREEN}✓ logstash 在网络: $LOGSTASH_NETWORK${NC}"
echo ""

# 3. 将 ceph-exporter 连接到 logstash 的网络
echo -e "${YELLOW}[2/5] 将 ceph-exporter 连接到 $LOGSTASH_NETWORK 网络...${NC}"
if docker network connect $LOGSTASH_NETWORK ceph-exporter 2>/dev/null; then
    echo -e "${GREEN}✓ 成功连接到网络${NC}"
else
    echo -e "${YELLOW}⚠ 容器可能已经在该网络中${NC}"
fi
echo ""

# 4. 验证网络连接
echo -e "${YELLOW}[3/5] 验证网络连接...${NC}"
if docker exec ceph-exporter ping -c 2 logstash >/dev/null 2>&1; then
    echo -e "${GREEN}✓ ceph-exporter 可以 ping 通 logstash${NC}"
else
    echo -e "${RED}✗ 仍然无法 ping 通，可能需要重启容器${NC}"
fi
echo ""

# 5. 重启 ceph-exporter
echo -e "${YELLOW}[4/5] 重启 ceph-exporter...${NC}"
docker-compose restart ceph-exporter
echo -e "${GREEN}✓ ceph-exporter 已重启${NC}"
echo ""

# 6. 等待并验证
echo -e "${YELLOW}[5/5] 等待 10 秒并验证...${NC}"
sleep 10

# 检查日志
if docker logs ceph-exporter 2>&1 | tail -5 | grep -q "no such host"; then
    echo -e "${RED}✗ 仍然有 DNS 解析错误${NC}"
    echo ""
    echo "可能需要重新创建容器。运行:"
    echo "  docker-compose stop ceph-exporter"
    echo "  docker-compose rm -f ceph-exporter"
    echo "  docker-compose up -d ceph-exporter"
else
    echo -e "${GREEN}✓ 没有发现 DNS 错误${NC}"

    # 检查是否成功连接
    if docker logs ceph-exporter 2>&1 | tail -10 | grep -q "ELK 集成已启用"; then
        echo -e "${GREEN}✓ ELK 集成已成功启用${NC}"
    fi
fi
echo ""

# 发送测试日志
echo -e "${YELLOW}发送测试日志...${NC}"
echo '{"message":"Test after network fix","level":"info","service":"ceph-exporter"}' | nc -w 2 localhost 5000
sleep 3

# 检查索引
echo -e "${YELLOW}检查 Elasticsearch 索引...${NC}"
INDICES=$(curl -s http://localhost:9200/_cat/indices?v 2>/dev/null | grep ceph)
if [ -n "$INDICES" ]; then
    echo -e "${GREEN}✓ 找到 ceph 索引:${NC}"
    echo "$INDICES"
else
    echo -e "${YELLOW}⚠ 还没有索引，请等待几秒后再检查${NC}"
    echo "  curl http://localhost:9200/_cat/indices?v | grep ceph"
fi
echo ""

echo -e "${BLUE}╔══════════════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║  下一步                                                                   ║${NC}"
echo -e "${BLUE}╚══════════════════════════════════════════════════════════════════════════╝${NC}"
echo ""
echo "1. 查看 ceph-exporter 日志:"
echo "   docker logs ceph-exporter | tail -20"
echo ""
echo "2. 检查 Elasticsearch 索引:"
echo "   curl http://localhost:9200/_cat/indices?v"
echo ""
echo "3. 在 Kibana 中刷新索引管理页面:"
echo "   http://localhost:5601"
echo ""
