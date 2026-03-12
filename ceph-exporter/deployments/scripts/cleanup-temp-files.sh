#!/bin/bash
# =============================================================================
# 清理 ELK 集成过程中的临时文件（可选）
# =============================================================================

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# 检查是否有 --confirm 参数
if [ "$1" != "--confirm" ]; then
    echo -e "${RED}警告: 此脚本将删除临时文件${NC}"
    echo ""
    echo "请先运行以下命令查看将要删除的文件:"
    echo "  ./scripts/list-temp-files.sh"
    echo ""
    echo "确认后，使用 --confirm 参数运行:"
    echo "  ./scripts/cleanup-temp-files.sh --confirm"
    echo ""
    exit 1
fi

echo -e "${YELLOW}开始清理临时文件...${NC}"
echo ""

# 1. 清理 /tmp 临时文件
echo "[1/3] 清理 /tmp 临时文件..."
rm -f /tmp/elk-*.txt /tmp/logstash-*.txt /tmp/network-*.txt 2>/dev/null
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ /tmp 临时文件已清理${NC}"
else
    echo -e "${YELLOW}⚠ 没有找到 /tmp 临时文件${NC}"
fi
echo ""

# 2. 清理临时文档
echo "[2/3] 清理临时文档..."
if [ -f "deployments/LOGSTASH-VERIFICATION-SUCCESS.md" ]; then
    rm -f deployments/LOGSTASH-VERIFICATION-SUCCESS.md
    echo -e "${GREEN}✓ 临时文档已删除${NC}"
else
    echo -e "${YELLOW}⚠ 临时文档不存在${NC}"
fi
echo ""

# 3. 询问是否删除备份文件
echo "[3/3] 备份文件处理..."
BAK_FILES=$(find . -name "*.bak" 2>/dev/null)
if [ -n "$BAK_FILES" ]; then
    echo "找到以下备份文件:"
    echo "$BAK_FILES"
    echo ""
    read -p "是否删除备份文件? (y/N): " -n 1 -r
    echo ""
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        find . -name "*.bak" -delete
        echo -e "${GREEN}✓ 备份文件已删除${NC}"
    else
        echo -e "${YELLOW}⚠ 保留备份文件${NC}"
    fi
else
    echo -e "${YELLOW}⚠ 没有找到备份文件${NC}"
fi
echo ""

echo -e "${GREEN}清理完成！${NC}"
