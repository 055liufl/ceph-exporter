#!/bin/bash
# =============================================================================
# 列出 ELK 集成过程中的临时文件
# =============================================================================

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}╔══════════════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║              ELK 集成 - 临时文件清单                                      ║${NC}"
echo -e "${BLUE}╚══════════════════════════════════════════════════════════════════════════╝${NC}"
echo ""

# 1. 备份文件
echo -e "${YELLOW}[1] 备份文件 (.bak, .backup)${NC}"
echo "这些是配置切换时自动创建的备份文件"
echo ""
find . -name "*.bak" -o -name "*.backup" 2>/dev/null | while read file; do
    if [ -f "$file" ]; then
        SIZE=$(ls -lh "$file" | awk '{print $5}')
        DATE=$(ls -l "$file" | awk '{print $6, $7, $8}')
        echo -e "  ${YELLOW}[备份]${NC} $file"
        echo "         大小: $SIZE, 日期: $DATE"
    fi
done
echo ""

# 2. 临时文档
echo -e "${YELLOW}[2] 临时文档（内容重复）${NC}"
echo "这些文档的内容已包含在其他文档中"
echo ""
TEMP_DOCS=(
    "deployments/LOGSTASH-VERIFICATION-SUCCESS.md"
)
for doc in "${TEMP_DOCS[@]}"; do
    if [ -f "$doc" ]; then
        SIZE=$(ls -lh "$doc" | awk '{print $5}')
        echo -e "  ${YELLOW}[文档]${NC} $doc (大小: $SIZE)"
    fi
done
echo ""

# 3. /tmp 临时文件
echo -e "${YELLOW}[3] /tmp 临时文件${NC}"
echo "这些是脚本生成的临时输出文件"
echo ""
ls -lh /tmp/elk-*.txt /tmp/logstash-*.txt /tmp/network-*.txt 2>/dev/null | while read line; do
    echo "  ${YELLOW}[临时]${NC} $line"
done
if [ $? -ne 0 ]; then
    echo "  (没有找到 /tmp 临时文件)"
fi
echo ""

# 4. 统计
echo -e "${BLUE}╔══════════════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║  统计信息                                                                 ║${NC}"
echo -e "${BLUE}╚══════════════════════════════════════════════════════════════════════════╝${NC}"
echo ""

BAK_COUNT=$(find . -name "*.bak" -o -name "*.backup" 2>/dev/null | wc -l)
TMP_COUNT=$(ls /tmp/elk-*.txt /tmp/logstash-*.txt /tmp/network-*.txt 2>/dev/null | wc -l)

echo "备份文件数量: $BAK_COUNT"
echo "临时文档数量: 1 (如果存在)"
echo "/tmp 临时文件: $TMP_COUNT"
echo ""

# 5. 删除建议
echo -e "${BLUE}╔══════════════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║  删除建议                                                                 ║${NC}"
echo -e "${BLUE}╚══════════════════════════════════════════════════════════════════════════╝${NC}"
echo ""

echo -e "${GREEN}安全删除（推荐）:${NC}"
echo ""
echo "1. 删除 /tmp 临时文件（完全安全）:"
echo "   rm -f /tmp/elk-*.txt /tmp/logstash-*.txt /tmp/network-*.txt"
echo ""

echo "2. 删除临时文档（确认当前系统正常后）:"
echo "   rm -f deployments/LOGSTASH-VERIFICATION-SUCCESS.md"
echo ""

echo -e "${YELLOW}谨慎删除（请先确认）:${NC}"
echo ""
echo "3. 删除备份文件（确认当前配置正常后）:"
echo "   find . -name '*.bak' -delete"
echo ""

echo -e "${RED}注意事项:${NC}"
echo "- 删除前请确认当前系统运行正常"
echo "- 备份文件可以帮助恢复之前的配置"
echo "- 如果磁盘空间充足，建议保留备份文件"
echo ""

# 6. 一键清理脚本
echo -e "${BLUE}╔══════════════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║  一键清理（可选）                                                         ║${NC}"
echo -e "${BLUE}╚══════════════════════════════════════════════════════════════════════════╝${NC}"
echo ""
echo "如果要一键清理所有临时文件，运行:"
echo "  ./scripts/cleanup-temp-files.sh --confirm"
echo ""
echo "详细说明请查看: docs/ELK-FILES-CLEANUP-GUIDE.md"
echo ""
