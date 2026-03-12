#!/bin/bash
# =============================================================================
# Logstash OOM 快速修复脚本
# =============================================================================

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${YELLOW}Logstash 内存不足问题修复脚本${NC}"
echo ""

# 检查是否在 deployments 目录
if [ ! -f "docker-compose-lightweight-full.yml" ]; then
    echo -e "${RED}错误: 请在 deployments 目录下运行此脚本${NC}"
    exit 1
fi

# 使用正确的 compose 文件
COMPOSE_FILE="docker-compose-lightweight-full.yml"

echo -e "${YELLOW}[1/4] 停止 Logstash 服务...${NC}"
docker-compose -f "$COMPOSE_FILE" stop logstash || true

echo -e "${YELLOW}[2/4] 删除旧容器...${NC}"
docker-compose -f "$COMPOSE_FILE" rm -f logstash || true

echo -e "${YELLOW}[3/4] 重新创建 Logstash 服务（新配置：512MB 堆内存，768MB 容器限制）...${NC}"
docker-compose -f "$COMPOSE_FILE" up -d logstash

echo -e "${YELLOW}[4/4] 等待 Logstash 启动...${NC}"
sleep 10

echo ""
echo -e "${GREEN}✓ Logstash 已重启${NC}"
echo ""
echo "查看日志:"
echo "  docker-compose -f $COMPOSE_FILE logs -f logstash"
echo ""
echo "检查状态:"
echo "  curl http://localhost:9600/_node/stats/pipelines"
echo ""
echo "详细文档: LOGSTASH-OOM-FIX.md"
