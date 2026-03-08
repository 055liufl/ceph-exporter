#!/bin/bash
# =============================================================================
# 快速修复部署问题脚本
# =============================================================================
# 用途: 修复常见的部署问题（权限、配置路径等）
# 使用: sudo ./scripts/fix-deployment.sh
# =============================================================================

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DEPLOY_DIR="$(dirname "$SCRIPT_DIR")"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

# 检查是否为 root 或有 sudo 权限
check_sudo() {
    if [ "$EUID" -ne 0 ]; then
        log_error "此脚本需要 root 权限"
        log_info "请使用: sudo $0"
        exit 1
    fi
}

# 修复 Prometheus 权限
fix_prometheus_permissions() {
    log_step "修复 Prometheus 数据目录权限..."

    local prom_dir="$DEPLOY_DIR/data/prometheus"

    if [ -d "$prom_dir" ]; then
        chown -R 65534:65534 "$prom_dir"
        log_info "✓ Prometheus 权限已修复 (65534:65534)"
    else
        log_warn "Prometheus 数据目录不存在: $prom_dir"
    fi
}

# 修复 Grafana 权限
fix_grafana_permissions() {
    log_step "修复 Grafana 数据目录权限..."

    local grafana_dir="$DEPLOY_DIR/data/grafana"

    if [ -d "$grafana_dir" ]; then
        chown -R 472:472 "$grafana_dir"
        log_info "✓ Grafana 权限已修复 (472:472)"
    else
        log_warn "Grafana 数据目录不存在: $grafana_dir"
    fi
}

# 修复 Elasticsearch 权限
fix_elasticsearch_permissions() {
    log_step "修复 Elasticsearch 数据目录权限..."

    local es_dir="$DEPLOY_DIR/data/elasticsearch"

    if [ -d "$es_dir" ]; then
        chown -R 1000:1000 "$es_dir"
        log_info "✓ Elasticsearch 权限已修复 (1000:1000)"
    else
        log_warn "Elasticsearch 数据目录不存在: $es_dir"
    fi
}

# 创建 configs 软链接
fix_configs_symlink() {
    log_step "检查 configs 目录软链接..."

    cd "$DEPLOY_DIR"

    if [ -L "configs" ]; then
        log_info "✓ configs 软链接已存在"
    elif [ -d "configs" ]; then
        log_warn "configs 是一个目录，不是软链接"
        log_info "如果遇到配置文件问题，请手动检查"
    else
        if [ -d "../configs" ]; then
            ln -s ../configs configs
            log_info "✓ 已创建 configs -> ../configs 软链接"
        else
            log_error "上级目录中不存在 configs 目录"
            log_info "请确保项目结构正确"
        fi
    fi
}

# 修复 Ceph keyring 权限
fix_ceph_keyring() {
    log_step "修复 Ceph keyring 文件权限..."

    local keyring="$DEPLOY_DIR/data/ceph-demo/config/ceph.client.admin.keyring"

    if [ -f "$keyring" ]; then
        # 修改为 644 权限，允许容器读取
        chmod 644 "$keyring"
        log_info "✓ Keyring 权限已修复 (644)"

        # 同时修复 mon keyring
        local mon_keyring="$DEPLOY_DIR/data/ceph-demo/config/ceph.mon.keyring"
        if [ -f "$mon_keyring" ]; then
            chmod 644 "$mon_keyring"
            log_info "✓ Mon Keyring 权限已修复 (644)"
        fi
    else
        log_warn "Keyring 文件不存在（等待 ceph-demo 生成）"
        log_info "如果 ceph-demo 已启动，请稍后重试"
    fi
}

# 检查并修复 vm.max_map_count（Elasticsearch 需要）
fix_vm_max_map_count() {
    log_step "检查 vm.max_map_count 设置..."

    local current=$(sysctl -n vm.max_map_count 2>/dev/null || echo 0)
    local required=262144

    if [ "$current" -lt "$required" ]; then
        log_warn "vm.max_map_count 当前值: $current (需要: $required)"
        sysctl -w vm.max_map_count=$required
        log_info "✓ vm.max_map_count 已设置为 $required"

        # 永久生效
        if ! grep -q "vm.max_map_count" /etc/sysctl.conf; then
            echo "vm.max_map_count=$required" >> /etc/sysctl.conf
            log_info "✓ 已添加到 /etc/sysctl.conf（永久生效）"
        fi
    else
        log_info "✓ vm.max_map_count 已正确设置: $current"
    fi
}

# 重启失败的服务
restart_failed_services() {
    log_step "检查并重启失败的服务..."

    cd "$DEPLOY_DIR"

    # 检测使用的 compose 文件
    local compose_file="docker-compose.yml"
    if docker ps --format '{{.Names}}' 2>/dev/null | grep -q "kibana"; then
        compose_file="docker-compose-lightweight-full.yml"
    elif docker ps --format '{{.Names}}' 2>/dev/null | grep -q "ceph-demo"; then
        compose_file="docker-compose-integration-test.yml"
    fi

    log_info "使用配置文件: $compose_file"

    # 检查 Prometheus
    if docker ps -a --format '{{.Names}}\t{{.Status}}' 2>/dev/null | grep "prometheus" | grep -q "Restarting"; then
        log_info "重启 Prometheus..."
        docker-compose -f "$compose_file" restart prometheus
    fi

    # 检查 ceph-exporter
    if docker ps -a --format '{{.Names}}\t{{.Status}}' 2>/dev/null | grep "ceph-exporter" | grep -q "Restarting"; then
        log_info "重启 ceph-exporter..."
        docker-compose -f "$compose_file" restart ceph-exporter
    fi

    log_info "✓ 服务重启完成"
}

# 显示服务状态
show_status() {
    log_step "当前服务状态..."
    echo ""

    cd "$DEPLOY_DIR"

    # 检测使用的 compose 文件
    if docker ps --format '{{.Names}}' 2>/dev/null | grep -q "kibana"; then
        docker-compose -f docker-compose-lightweight-full.yml ps
    elif docker ps --format '{{.Names}}' 2>/dev/null | grep -q "ceph-demo"; then
        docker-compose -f docker-compose-integration-test.yml ps
    else
        docker-compose ps
    fi
}

# 主函数
main() {
    echo "========================================"
    echo "  ceph-exporter 部署问题修复脚本"
    echo "========================================"
    echo ""

    check_sudo

    # 执行所有修复步骤
    fix_prometheus_permissions
    echo ""

    fix_grafana_permissions
    echo ""

    fix_elasticsearch_permissions
    echo ""

    fix_configs_symlink
    echo ""

    fix_ceph_keyring
    echo ""

    fix_vm_max_map_count
    echo ""

    restart_failed_services
    echo ""

    log_info "等待 30 秒让服务启动..."
    sleep 30
    echo ""

    show_status
    echo ""

    log_step "修复完成！"
    echo ""
    log_info "如果服务仍然失败，请查看日志:"
    echo "  docker logs prometheus"
    echo "  docker logs ceph-exporter"
    echo "  docker logs ceph-demo"
    echo ""
    log_info "或运行验证脚本:"
    echo "  sudo ./scripts/deploy.sh verify"
    echo ""
    log_info "详细故障排查指南:"
    echo "  cat ../TROUBLESHOOTING.md"
}

main "$@"
