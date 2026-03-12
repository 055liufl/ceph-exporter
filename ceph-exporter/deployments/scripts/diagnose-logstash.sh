#!/bin/bash
# =============================================================================
# Logstash 状态诊断脚本
# =============================================================================

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}╔══════════════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║                  Logstash 状态诊断                                        ║${NC}"
echo -e "${BLUE}╚══════════════════════════════════════════════════════════════════════════╝${NC}"
echo ""

# 1. 检查容器状态
echo -e "${YELLOW}[1/6] 检查容器状态...${NC}"
CONTAINER_STATUS=$(docker ps -a --filter "name=logstash" --format "{{.Status}}")
if [ -z "$CONTAINER_STATUS" ]; then
    echo -e "${RED}✗ Logstash 容器不存在${NC}"
    exit 1
else
    echo -e "${GREEN}✓ 容器状态: $CONTAINER_STATUS${NC}"
fi
echo ""

# 2. 检查容器是否运行
echo -e "${YELLOW}[2/6] 检查容器是否运行...${NC}"
if docker ps --filter "name=logstash" | grep -q logstash; then
    echo -e "${GREEN}✓ Logstash 容器正在运行${NC}"
else
    echo -e "${RED}✗ Logstash 容器未运行${NC}"
    echo "尝试启动容器:"
    echo "  sudo docker-compose -f docker-compose-lightweight-full.yml start logstash"
    exit 1
fi
echo ""

# 3. 检查端口监听
echo -e "${YELLOW}[3/6] 检查端口监听...${NC}"
if netstat -tuln 2>/dev/null | grep -q ":9600"; then
    echo -e "${GREEN}✓ 端口 9600 (API) 正在监听${NC}"
else
    echo -e "${YELLOW}⚠ 端口 9600 未监听（可能还在启动中）${NC}"
fi

if netstat -tuln 2>/dev/null | grep -q ":5044"; then
    echo -e "${GREEN}✓ 端口 5044 (Beats) 正在监听${NC}"
else
    echo -e "${YELLOW}⚠ 端口 5044 未监听（可能还在启动中）${NC}"
fi

if netstat -tuln 2>/dev/null | grep -q ":5000"; then
    echo -e "${GREEN}✓ 端口 5000 (TCP) 正在监听${NC}"
else
    echo -e "${YELLOW}⚠ 端口 5000 未监听（可能还在启动中）${NC}"
fi
echo ""

# 4. 检查最新日志
echo -e "${YELLOW}[4/6] 检查最新日志（最后 20 行）...${NC}"
echo "----------------------------------------"
docker logs logstash 2>&1 | tail -20
echo "----------------------------------------"
echo ""

# 5. 检查是否有错误
echo -e "${YELLOW}[5/6] 检查错误日志...${NC}"
ERROR_COUNT=$(docker logs logstash 2>&1 | grep -i "error\|exception\|failed" | wc -l)
if [ "$ERROR_COUNT" -gt 0 ]; then
    echo -e "${YELLOW}⚠ 发现 $ERROR_COUNT 条错误/异常日志${NC}"
    echo "最近的错误:"
    docker logs logstash 2>&1 | grep -i "error\|exception\|failed" | tail -5
else
    echo -e "${GREEN}✓ 没有发现错误日志${NC}"
fi
echo ""

# 6. 检查启动状态
echo -e "${YELLOW}[6/6] 检查启动状态...${NC}"
if docker logs logstash 2>&1 | grep -q "Successfully started Logstash API endpoint"; then
    echo -e "${GREEN}✓ Logstash API 已成功启动${NC}"

    # 测试 API
    echo ""
    echo "测试 API 连接..."
    if curl -s http://localhost:9600/_node/stats/pipelines > /dev/null 2>&1; then
        echo -e "${GREEN}✓ API 响应正常${NC}"
    else
        echo -e "${YELLOW}⚠ API 暂时无响应（可能还在初始化）${NC}"
    fi
else
    echo -e "${YELLOW}⚠ Logstash 可能还在启动中...${NC}"
    echo ""
    echo "启动过程通常需要 1-2 分钟，请稍候再试。"
    echo ""
    echo "实时查看启动日志:"
    echo "  docker logs -f logstash"
fi
echo ""

# 7. 显示内存配置
echo -e "${YELLOW}内存配置:${NC}"
docker logs logstash 2>&1 | grep "JVM bootstrap flags" | tail -1 | grep -o "\-Xms[^ ]* \-Xmx[^ ]*"
echo ""

# 8. 建议
echo -e "${BLUE}╔══════════════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║  建议操作                                                                 ║${NC}"
echo -e "${BLUE}╚══════════════════════════════════════════════════════════════════════════╝${NC}"
echo ""
echo "1. 实时查看日志:"
echo "   docker logs -f logstash"
echo ""
echo "2. 等待 1-2 分钟后重新检查:"
echo "   curl http://localhost:9600/_node/stats/pipelines"
echo ""
echo "3. 如果长时间未启动，重启容器:"
echo "   sudo docker-compose -f docker-compose-lightweight-full.yml restart logstash"
echo ""
echo "4. 查看完整日志:"
echo "   docker logs logstash 2>&1 | less"
echo ""
