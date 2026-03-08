#!/bin/bash
# =============================================================================
# CentOS 7 部署验证脚本
# =============================================================================
# 验证所有服务是否正常运行并可访问
# =============================================================================

set -euo pipefail

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

# 检查容器状态
check_containers() {
    log_step "检查容器状态..."
    echo ""

    local containers=(
        "ceph-exporter"
        "prometheus"
        "grafana"
        "alertmanager"
    )

    # 检查是否部署了完整栈
    if docker ps --format '{{.Names}}' | grep -q "elasticsearch"; then
        containers+=("elasticsearch" "logstash" "kibana" "jaeger")
    fi

    # 检查是否部署了 Ceph Demo
    if docker ps --format '{{.Names}}' | grep -q "ceph-demo"; then
        containers+=("ceph-demo")
    fi

    local all_running=true

    for container in "${containers[@]}"; do
        if docker ps --format '{{.Names}}' | grep -q "^${container}$"; then
            local status=$(docker inspect --format='{{.State.Status}}' "$container" 2>/dev/null || echo "not found")
            local health=$(docker inspect --format='{{.State.Health.Status}}' "$container" 2>/dev/null || echo "no healthcheck")

            if [ "$status" == "running" ]; then
                if [ "$health" == "healthy" ] || [ "$health" == "no healthcheck" ]; then
                    echo -e "${GREEN}✓${NC} $container: 运行中"
                else
                    echo -e "${YELLOW}⚠${NC} $container: 运行中 (健康检查: $health)"
                    all_running=false
                fi
            else
                echo -e "${RED}✗${NC} $container: $status"
                all_running=false
            fi
        else
            echo -e "${RED}✗${NC} $container: 未找到"
            all_running=false
        fi
    done

    echo ""
    if [ "$all_running" = true ]; then
        log_info "所有容器运行正常"
    else
        log_warn "部分容器状态异常"
    fi

    return 0
}

# 检查服务端点
check_endpoints() {
    log_step "检查服务端点..."
    echo ""

    local endpoints=(
        "ceph-exporter:http://localhost:9128/health"
        "ceph-exporter-metrics:http://localhost:9128/metrics"
        "prometheus:http://localhost:9090/-/healthy"
        "grafana:http://localhost:3000/api/health"
        "alertmanager:http://localhost:9093/-/healthy"
    )

    # 检查是否部署了完整栈
    if docker ps --format '{{.Names}}' | grep -q "elasticsearch"; then
        endpoints+=(
            "elasticsearch:http://localhost:9200"
            "kibana:http://localhost:5601/api/status"
            "jaeger:http://localhost:16686"
        )
    fi

    # 检查是否部署了 Ceph Demo
    if docker ps --format '{{.Names}}' | grep -q "ceph-demo"; then
        endpoints+=("ceph-dashboard:http://localhost:8080")
    fi

    local all_accessible=true

    for endpoint in "${endpoints[@]}"; do
        local name="${endpoint%%:*}"
        local url="${endpoint#*:}"

        if curl -sf --max-time 5 "$url" > /dev/null 2>&1; then
            echo -e "${GREEN}✓${NC} $name: 可访问"
        else
            echo -e "${RED}✗${NC} $name: 无法访问 ($url)"
            all_accessible=false
        fi
    done

    echo ""
    if [ "$all_accessible" = true ]; then
        log_info "所有服务端点可访问"
    else
        log_warn "部分服务端点无法访问"
    fi

    return 0
}

# 检查 Prometheus 目标
check_prometheus_targets() {
    log_step "检查 Prometheus 目标..."
    echo ""

    if ! docker ps --format '{{.Names}}' | grep -q "^prometheus$"; then
        log_warn "Prometheus 未运行，跳过目标检查"
        return 0
    fi

    local targets_json=$(curl -sf http://localhost:9090/api/v1/targets 2>/dev/null || echo "{}")

    if [ "$targets_json" == "{}" ]; then
        log_error "无法获取 Prometheus 目标信息"
        return 0
    fi

    # 简单检查是否有 ceph-exporter 目标
    if echo "$targets_json" | grep -q "ceph-exporter"; then
        echo -e "${GREEN}✓${NC} ceph-exporter 目标已配置"
    else
        echo -e "${YELLOW}⚠${NC} 未找到 ceph-exporter 目标"
    fi

    echo ""
}

# 检查 Grafana 数据源
check_grafana_datasources() {
    log_step "检查 Grafana 数据源..."
    echo ""

    if ! docker ps --format '{{.Names}}' | grep -q "^grafana$"; then
        log_warn "Grafana 未运行，跳过数据源检查"
        return 0
    fi

    # 等待 Grafana 完全启动
    sleep 2

    local datasources=$(curl -sf -u admin:admin http://localhost:3000/api/datasources 2>/dev/null || echo "[]")

    if [ "$datasources" == "[]" ]; then
        log_warn "无法获取 Grafana 数据源信息或未配置数据源"
        return 0
    fi

    if echo "$datasources" | grep -q "Prometheus"; then
        echo -e "${GREEN}✓${NC} Prometheus 数据源已配置"
    else
        echo -e "${YELLOW}⚠${NC} 未找到 Prometheus 数据源"
    fi

    echo ""
}

# 检查资源使用
check_resource_usage() {
    log_step "检查资源使用情况..."
    echo ""

    # 检查内存使用
    local total_mem=$(free -m | awk '/^Mem:/{print $2}')
    local used_mem=$(free -m | awk '/^Mem:/{print $3}')
    local mem_percent=$((used_mem * 100 / total_mem))

    echo "系统内存: ${used_mem}MB / ${total_mem}MB (${mem_percent}%)"

    if [ "$mem_percent" -gt 90 ]; then
        log_warn "内存使用率超过 90%"
    fi

    # 检查磁盘使用
    local disk_usage=$(df -h . | awk 'NR==2 {print $5}' | sed 's/%//')
    echo "磁盘使用: ${disk_usage}%"

    if [ "$disk_usage" -gt 90 ]; then
        log_warn "磁盘使用率超过 90%"
    fi

    echo ""
    echo "Docker 容器资源使用:"
    docker stats --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}" | head -10

    echo ""
}

# 生成验证报告
generate_report() {
    log_step "生成验证报告..."
    echo ""

    local report_file="/tmp/ceph-exporter-verification-$(date +%Y%m%d-%H%M%S).txt"

    {
        echo "=========================================="
        echo "ceph-exporter 部署验证报告"
        echo "=========================================="
        echo "验证时间: $(date '+%Y-%m-%d %H:%M:%S')"
        echo "主机名: $(hostname)"
        echo "操作系统: $(cat /etc/centos-release 2>/dev/null || echo 'Unknown')"
        echo ""
        echo "=========================================="
        echo "容器状态"
        echo "=========================================="
        docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
        echo ""
        echo "=========================================="
        echo "资源使用"
        echo "=========================================="
        echo "内存:"
        free -h
        echo ""
        echo "磁盘:"
        df -h
        echo ""
        echo "Docker 容器资源:"
        docker stats --no-stream
        echo ""
        echo "=========================================="
        echo "服务日志（最后 20 行）"
        echo "=========================================="
        for container in ceph-exporter prometheus grafana; do
            if docker ps --format '{{.Names}}' | grep -q "^${container}$"; then
                echo ""
                echo "--- $container ---"
                docker logs --tail 20 "$container" 2>&1
            fi
        done
    } > "$report_file"

    log_info "验证报告已保存到: $report_file"
    echo ""
}

# 显示访问信息
show_access_info() {
    log_step "服务访问信息:"
    echo ""

    # 获取服务器 IP
    local server_ip=$(hostname -I | awk '{print $1}')

    echo "本地访问:"
    echo "  Ceph Exporter:   http://localhost:9128/metrics"
    echo "  Prometheus:      http://localhost:9090"
    echo "  Grafana:         http://localhost:3000 (admin/admin)"
    echo "  Alertmanager:    http://localhost:9093"

    if docker ps --format '{{.Names}}' | grep -q "elasticsearch"; then
        echo "  Elasticsearch:   http://localhost:9200"
        echo "  Kibana:          http://localhost:5601"
        echo "  Jaeger UI:       http://localhost:16686"
    fi

    if docker ps --format '{{.Names}}' | grep -q "ceph-demo"; then
        echo "  Ceph Dashboard:  http://localhost:8080"
    fi

    echo ""
    echo "远程访问 (服务器 IP: $server_ip):"
    echo "  Ceph Exporter:   http://${server_ip}:9128/metrics"
    echo "  Prometheus:      http://${server_ip}:9090"
    echo "  Grafana:         http://${server_ip}:3000 (admin/admin)"
    echo "  Alertmanager:    http://${server_ip}:9093"

    if docker ps --format '{{.Names}}' | grep -q "elasticsearch"; then
        echo "  Elasticsearch:   http://${server_ip}:9200"
        echo "  Kibana:          http://${server_ip}:5601"
        echo "  Jaeger UI:       http://${server_ip}:16686"
    fi

    if docker ps --format '{{.Names}}' | grep -q "ceph-demo"; then
        echo "  Ceph Dashboard:  http://${server_ip}:8080"
    fi

    echo ""
}

# 主函数
main() {
    echo ""
    echo "=========================================="
    echo "ceph-exporter 部署验证"
    echo "=========================================="
    echo ""

    check_containers
    check_endpoints
    check_prometheus_targets
    check_grafana_datasources
    check_resource_usage
    show_access_info

    # 询问是否生成详细报告
    read -p "是否生成详细验证报告? (y/N): " generate
    if [[ "$generate" == "y" || "$generate" == "Y" ]]; then
        generate_report
    fi

    echo ""
    log_info "验证完成"
    echo ""
}

main "$@"
