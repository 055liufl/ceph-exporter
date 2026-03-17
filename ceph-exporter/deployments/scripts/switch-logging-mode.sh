#!/bin/bash
# =============================================================================
# 日志方案快速切换脚本
# =============================================================================
# 用法:
#   ./switch-logging-mode.sh [mode]
#
# 模式:
#   direct    - 方案1: 直接推送到 Logstash (TCP)
#   direct-udp - 方案1: 直接推送到 Logstash (UDP)
#   container - 方案2: 容器日志收集（推荐）
#   file      - 方案3: 文件日志 + Filebeat
#   dev       - 开发模式: stdout + text 格式
#   show      - 显示当前配置
# =============================================================================

set -e

# 获取脚本所在目录的父目录（deployments）
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DEPLOY_DIR="$(dirname "$SCRIPT_DIR")"

CONFIG_FILE="$DEPLOY_DIR/configs/ceph-exporter.yaml"
BACKUP_FILE="$DEPLOY_DIR/configs/ceph-exporter.yaml.bak"
COMPOSE_FILE="$DEPLOY_DIR/docker-compose-lightweight-full.yml"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 检测 docker-compose 命令
detect_compose_cmd() {
    if command -v docker-compose &> /dev/null; then
        COMPOSE_CMD="docker-compose"
    elif docker compose version &> /dev/null 2>&1; then
        COMPOSE_CMD="docker compose"
    else
        COMPOSE_CMD=""
    fi
}

# 启动 filebeat-sidecar
start_filebeat_sidecar() {
    detect_compose_cmd
    if [ -z "$COMPOSE_CMD" ]; then
        echo -e "${YELLOW}!${NC} 未检测到 docker-compose，请手动启动 filebeat-sidecar"
        return
    fi

    if [ ! -f "$COMPOSE_FILE" ]; then
        echo -e "${YELLOW}!${NC} 未找到 $COMPOSE_FILE，请手动启动 filebeat-sidecar"
        return
    fi

    echo -e "${GREEN}✓${NC} 启动 filebeat-sidecar..."
    cd "$DEPLOY_DIR"
    ${COMPOSE_CMD} -f docker-compose-lightweight-full.yml up -d filebeat-sidecar 2>/dev/null || {
        echo -e "${YELLOW}!${NC} filebeat-sidecar 启动失败，请手动启动"
    }
}

# 停止 filebeat-sidecar
stop_filebeat_sidecar() {
    detect_compose_cmd
    if [ -z "$COMPOSE_CMD" ]; then
        return
    fi

    if [ ! -f "$COMPOSE_FILE" ]; then
        return
    fi

    # 检查 filebeat-sidecar 是否在运行
    if docker ps --format '{{.Names}}' 2>/dev/null | grep -q "filebeat-sidecar"; then
        echo -e "${GREEN}✓${NC} 停止 filebeat-sidecar（直接推送模式不需要）..."
        cd "$DEPLOY_DIR"
        ${COMPOSE_CMD} -f docker-compose-lightweight-full.yml stop filebeat-sidecar 2>/dev/null || true
        ${COMPOSE_CMD} -f docker-compose-lightweight-full.yml rm -f filebeat-sidecar 2>/dev/null || true
    fi
}

# 显示使用说明
usage() {
    echo "用法: $0 [mode]"
    echo ""
    echo "可用模式:"
    echo "  direct      - 方案1: 直接推送到 Logstash (TCP)"
    echo "  direct-udp  - 方案1: 直接推送到 Logstash (UDP)"
    echo "  container   - 方案2: 容器日志收集（推荐）"
    echo "  file        - 方案3: 文件日志 + Filebeat"
    echo "  dev         - 开发模式: stdout + text 格式"
    echo "  show        - 显示当前配置"
    echo ""
    echo "示例:"
    echo "  $0 direct      # 切换到直接推送模式"
    echo "  $0 container   # 切换到容器日志收集模式"
    echo "  $0 show        # 显示当前配置"
    exit 1
}

# 备份配置文件
backup_config() {
    if [ -f "$CONFIG_FILE" ]; then
        cp "$CONFIG_FILE" "$BACKUP_FILE"
        echo -e "${GREEN}✓${NC} 配置文件已备份到 $BACKUP_FILE"
    fi
}

# 显示当前配置
show_config() {
    echo -e "${YELLOW}当前日志配置:${NC}"
    echo ""
    grep -A 15 "^logger:" "$CONFIG_FILE" | grep -E "(level|format|output|file_path|enable_elk|logstash_url|logstash_protocol|service_name):" | sed 's/^/  /'
    echo ""
}

# 切换到方案1: 直接推送 (TCP)
switch_to_direct_tcp() {
    echo -e "${YELLOW}切换到方案1: 直接推送到 Logstash (TCP)${NC}"
    backup_config

    sed -i 's/enable_elk: false/enable_elk: true/' "$CONFIG_FILE"
    sed -i 's/logstash_protocol: "udp"/logstash_protocol: "tcp"/' "$CONFIG_FILE"

    echo -e "${GREEN}✓${NC} 已切换到直接推送模式 (TCP)"
    echo ""
    echo "配置详情:"
    echo "  - enable_elk: true"
    echo "  - logstash_protocol: tcp"
    echo "  - output: stdout"
    echo ""

    # 停止不需要的 filebeat-sidecar
    stop_filebeat_sidecar

    echo ""
    echo "重启服务以应用配置:"
    echo "  docker-compose -f docker-compose-lightweight-full.yml restart ceph-exporter"
}

# 切换到方案1: 直接推送 (UDP)
switch_to_direct_udp() {
    echo -e "${YELLOW}切换到方案1: 直接推送到 Logstash (UDP)${NC}"
    backup_config

    sed -i 's/enable_elk: false/enable_elk: true/' "$CONFIG_FILE"
    sed -i 's/logstash_protocol: "tcp"/logstash_protocol: "udp"/' "$CONFIG_FILE"

    echo -e "${GREEN}✓${NC} 已切换到直接推送模式 (UDP)"
    echo ""
    echo "配置详情:"
    echo "  - enable_elk: true"
    echo "  - logstash_protocol: udp"
    echo "  - output: stdout"
    echo ""

    # 停止不需要的 filebeat-sidecar
    stop_filebeat_sidecar

    echo ""
    echo "重启服务以应用配置:"
    echo "  docker-compose -f docker-compose-lightweight-full.yml restart ceph-exporter"
}

# 切换到方案2: 容器日志收集
switch_to_container() {
    echo -e "${YELLOW}切换到方案2: 容器日志收集（推荐）${NC}"
    backup_config

    sed -i 's/enable_elk: true/enable_elk: false/' "$CONFIG_FILE"
    sed -i 's/output: "file"/output: "stdout"/' "$CONFIG_FILE"
    sed -i 's/format: "text"/format: "json"/' "$CONFIG_FILE"

    echo -e "${GREEN}✓${NC} 已切换到容器日志收集模式"
    echo ""
    echo "配置详情:"
    echo "  - enable_elk: false"
    echo "  - output: stdout"
    echo "  - format: json"
    echo ""

    # 自动启动 filebeat-sidecar
    start_filebeat_sidecar

    echo ""
    echo "重启服务以应用配置:"
    echo "  docker-compose -f docker-compose-lightweight-full.yml restart ceph-exporter"
}

# 切换到方案3: 文件日志
switch_to_file() {
    echo -e "${YELLOW}切换到方案3: 文件日志 + Filebeat${NC}"
    backup_config

    sed -i 's/enable_elk: true/enable_elk: false/' "$CONFIG_FILE"
    sed -i 's/output: "stdout"/output: "file"/' "$CONFIG_FILE"
    sed -i 's/format: "text"/format: "json"/' "$CONFIG_FILE"

    echo -e "${GREEN}✓${NC} 已切换到文件日志模式"
    echo ""
    echo "配置详情:"
    echo "  - enable_elk: false"
    echo "  - output: file"
    echo "  - format: json"
    echo ""
    echo "确保日志目录存在:"
    echo "  mkdir -p /var/log/ceph-exporter"
    echo ""
    echo "配置 Filebeat 监控日志文件:"
    echo "  参考 configs/filebeat.yml"
    echo ""
    echo "重启服务以应用配置:"
    echo "  docker-compose restart ceph-exporter"
}

# 切换到开发模式
switch_to_dev() {
    echo -e "${YELLOW}切换到开发模式${NC}"
    backup_config

    sed -i 's/enable_elk: true/enable_elk: false/' "$CONFIG_FILE"
    sed -i 's/output: "file"/output: "stdout"/' "$CONFIG_FILE"
    sed -i 's/format: "json"/format: "text"/' "$CONFIG_FILE"
    sed -i 's/level: "info"/level: "debug"/' "$CONFIG_FILE"

    echo -e "${GREEN}✓${NC} 已切换到开发模式"
    echo ""
    echo "配置详情:"
    echo "  - enable_elk: false"
    echo "  - output: stdout"
    echo "  - format: text"
    echo "  - level: debug"
    echo ""
    echo "重启服务以应用配置:"
    echo "  docker-compose restart ceph-exporter"
}

# 主逻辑
main() {
    if [ $# -eq 0 ]; then
        usage
    fi

    if [ ! -f "$CONFIG_FILE" ]; then
        echo -e "${RED}错误: 配置文件不存在: $CONFIG_FILE${NC}"
        exit 1
    fi

    case "$1" in
        direct)
            switch_to_direct_tcp
            ;;
        direct-udp)
            switch_to_direct_udp
            ;;
        container)
            switch_to_container
            ;;
        file)
            switch_to_file
            ;;
        dev)
            switch_to_dev
            ;;
        show)
            show_config
            ;;
        *)
            echo -e "${RED}错误: 未知模式 '$1'${NC}"
            echo ""
            usage
            ;;
    esac
}

main "$@"
