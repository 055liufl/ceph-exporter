#!/bin/bash
# =============================================================================
# ceph-exporter 统一诊断脚本
# =============================================================================
# 功能: 全面诊断所有服务状态，收集日志和配置信息
# 用途: 快速定位部署问题
# 使用: sudo ./scripts/diagnose.sh [service-name]
#       service-name 可选: all(默认), ceph-exporter, ceph-demo, prometheus,
#                          grafana, kibana, elasticsearch, jaeger
# =============================================================================

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DEPLOY_DIR="$(dirname "$SCRIPT_DIR")"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
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

log_section() {
    echo ""
    echo -e "${CYAN}═══════════════════════════════════════════════════════════════════${NC}"
    echo -e "${CYAN}  $1${NC}"
    echo -e "${CYAN}═══════════════════════════════════════════════════════════════════${NC}"
    echo ""
}

# 显示标题
show_header() {
    echo ""
    echo "╔══════════════════════════════════════════════════════════════════════╗"
    echo "║              ceph-exporter 部署诊断报告                             ║"
    echo "╚══════════════════════════════════════════════════════════════════════╝"
    echo ""
    echo "诊断时间: $(date '+%Y-%m-%d %H:%M:%S')"
    echo "诊断模式: ${1:-完整诊断}"
    echo ""
}

# 1. 检查所有容器状态
check_all_containers() {
    log_section "1. 容器状态总览"

    docker ps -a --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" 2>/dev/null || {
        log_error "无法访问 Docker，请确保 Docker 服务正在运行"
        exit 1
    }
}

# 2. 诊断 ceph-exporter
diagnose_ceph_exporter() {
    log_section "2. Ceph-Exporter 详细诊断"

    # 容器状态
    log_step "容器状态"
    local status=$(docker inspect ceph-exporter --format='{{.State.Status}}' 2>/dev/null || echo "not found")
    local restarts=$(docker inspect ceph-exporter --format='{{.RestartCount}}' 2>/dev/null || echo "0")
    echo "状态: $status"
    echo "重启次数: $restarts"
    echo ""

    # 最近日志
    log_step "最近日志 (最后 30 行)"
    docker logs ceph-exporter --tail 30 2>&1
    echo ""

    # 配置检查
    log_step "配置文件检查"
    echo "[configs 软链接]"
    ls -la "$DEPLOY_DIR/configs" 2>&1 || log_warn "configs 软链接不存在"
    echo ""

    echo "[ceph-exporter.yaml]"
    ls -la "$DEPLOY_DIR/configs/ceph-exporter.yaml" 2>&1 || log_warn "配置文件不存在"
    echo ""

    echo "[Ceph 配置目录]"
    ls -la "$DEPLOY_DIR/data/ceph-demo/config/" 2>&1 || log_warn "Ceph 配置目录不存在"
    echo ""

    # 挂载检查
    log_step "数据卷挂载"
    docker inspect ceph-exporter --format='{{range .Mounts}}{{.Type}}: {{.Source}} -> {{.Destination}}{{"\n"}}{{end}}' 2>&1
    echo ""

    # 网络检查
    log_step "网络连接测试"
    docker exec ceph-exporter ping -c 2 ceph-demo 2>&1 || log_warn "无法连接到 ceph-demo"
    echo ""

    # 诊断建议
    if [ "$restarts" -gt 5 ]; then
        log_error "重启次数过多 ($restarts 次)"
        echo ""
        echo "可能原因:"
        echo "  1. Ceph keyring 权限问题 (应为 644)"
        echo "  2. configs 软链接不存在"
        echo "  3. Ceph 集群尚未完全启动"
        echo ""
        echo "修复命令:"
        echo "  sudo chmod 644 $DEPLOY_DIR/data/ceph-demo/config/*.keyring"
        echo "  cd $DEPLOY_DIR && ln -s ../configs configs"
        echo "  docker compose restart ceph-exporter"
        echo ""
    fi
}

# 3. 诊断 ceph-demo
diagnose_ceph_demo() {
    log_section "3. Ceph-Demo 详细诊断"

    # 容器状态
    log_step "容器状态"
    docker ps -a --filter "name=ceph-demo" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
    echo ""

    # Ceph 集群状态
    log_step "Ceph 集群状态"
    docker exec ceph-demo ceph -s 2>&1 || log_warn "Ceph 集群尚未就绪"
    echo ""

    # 端口监听
    log_step "端口监听检查"
    docker exec ceph-demo netstat -tlnp 2>/dev/null | grep -E ':(8080|5000)' || \
    docker exec ceph-demo ss -tlnp 2>/dev/null | grep -E ':(8080|5000)' || \
    log_warn "无法获取端口信息"
    echo ""

    # RGW 进程
    log_step "RGW 进程检查"
    docker exec ceph-demo ps aux | grep radosgw | grep -v grep || log_warn "未找到 RGW 进程"
    echo ""

    # 配置文件
    log_step "配置文件"
    docker exec ceph-demo ls -la /etc/ceph/ 2>&1
    echo ""

    # 测试端口
    log_step "端口连接测试"
    timeout 3 curl -v http://localhost:8080 2>&1 | head -15
    echo ""
}

# 4. 诊断 Prometheus
diagnose_prometheus() {
    log_section "4. Prometheus 详细诊断"

    # 容器状态
    log_step "容器状态"
    docker ps -a --filter "name=prometheus" --format "table {{.Names}}\t{{.Status}}"
    echo ""

    # 最近日志
    log_step "最近日志 (最后 30 行)"
    docker logs prometheus --tail 30 2>&1
    echo ""

    # 数据目录权限
    log_step "数据目录权限"
    ls -la "$DEPLOY_DIR/data/prometheus" 2>&1
    echo ""
    echo "注意: Prometheus 需要 UID 65534 (nobody) 权限"
    echo ""

    # 健康检查
    log_step "健康检查"
    curl -sf http://localhost:9090/-/healthy && echo "✓ Prometheus 健康" || echo "✗ Prometheus 不健康"
    echo ""
}

# 5. 诊断 Grafana
diagnose_grafana() {
    log_section "5. Grafana 详细诊断"

    # 容器状态
    log_step "容器状态"
    docker ps -a --filter "name=grafana" --format "table {{.Names}}\t{{.Status}}"
    echo ""

    # 最近日志
    log_step "最近日志 (最后 20 行)"
    docker logs grafana --tail 20 2>&1
    echo ""

    # 数据目录权限
    log_step "数据目录权限"
    ls -la "$DEPLOY_DIR/data/grafana" 2>&1
    echo ""
    echo "注意: Grafana 需要 UID 472 权限"
    echo ""

    # 健康检查
    log_step "健康检查"
    curl -sf http://localhost:3000/api/health && echo "✓ Grafana 健康" || echo "✗ Grafana 不健康"
    echo ""
}

# 6. 诊断 Kibana
diagnose_kibana() {
    log_section "6. Kibana 详细诊断"

    # 容器状态
    log_step "容器状态"
    docker ps -a --filter "name=kibana" --format "table {{.Names}}\t{{.Status}}"
    echo ""

    # 最近日志
    log_step "最近日志 (最后 50 行)"
    docker logs kibana --tail 50 2>&1
    echo ""

    # Elasticsearch 连接
    log_step "Elasticsearch 连接测试"
    curl -s http://localhost:9200/_cluster/health?pretty
    echo ""

    # Kibana 状态
    log_step "Kibana 状态"
    timeout 3 curl -v http://localhost:5601/api/status 2>&1 | head -20
    echo ""

    # 内存检查
    log_step "内存使用"
    docker stats kibana --no-stream --format "table {{.Name}}\t{{.MemUsage}}\t{{.MemPerc}}"
    echo ""
    echo "注意: Kibana 建议至少 1GB 内存"
    echo ""
}

# 7. 诊断 Elasticsearch
diagnose_elasticsearch() {
    log_section "7. Elasticsearch 详细诊断"

    # 容器状态
    log_step "容器状态"
    docker ps -a --filter "name=elasticsearch" --format "table {{.Names}}\t{{.Status}}"
    echo ""

    # 集群健康
    log_step "集群健康状态"
    curl -s http://localhost:9200/_cluster/health?pretty
    echo ""

    # 数据目录权限
    log_step "数据目录权限"
    ls -la "$DEPLOY_DIR/data/elasticsearch" 2>&1
    echo ""
    echo "注意: Elasticsearch 需要 UID 1000 权限"
    echo ""
}

# 8. 诊断 Jaeger
diagnose_jaeger() {
    log_section "8. Jaeger 详细诊断"

    # 容器状态
    log_step "容器状态"
    docker ps -a --filter "name=jaeger" --format "table {{.Names}}\t{{.Status}}"
    echo ""

    # 健康检查
    log_step "健康检查"
    timeout 3 curl -v http://localhost:16686 2>&1 | head -15
    echo ""
}

# 9. 系统资源检查
check_system_resources() {
    log_section "9. 系统资源使用"

    log_step "容器资源使用"
    docker stats --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.MemPerc}}"
    echo ""

    log_step "系统资源"
    echo "[内存]"
    free -h
    echo ""
    echo "[磁盘]"
    df -h "$DEPLOY_DIR"
    echo ""
}

# 10. 服务健康检查
check_services_health() {
    log_section "10. 服务健康检查"

    local services=(
        "ceph-exporter:9128/health"
        "prometheus:9090/-/healthy"
        "grafana:3000/api/health"
        "elasticsearch:9200"
        "kibana:5601/api/status"
        "jaeger:16686"
        "ceph-demo:8080"
    )

    for service in "${services[@]}"; do
        local name="${service%%:*}"
        local endpoint="${service#*:}"

        # 特殊处理 ceph-demo
        if [ "$name" = "ceph-demo" ]; then
            if curl -s --max-time 3 "http://localhost:${endpoint}" 2>&1 | grep -q "NoSuchBucket\|InvalidBucketName\|404"; then
                echo -e "${GREEN}✓${NC} $name: 运行正常 (RGW)"
            else
                echo -e "${YELLOW}⚠${NC} $name: 无法访问或尚未就绪"
            fi
        else
            if curl -sf --max-time 3 "http://localhost:${endpoint}" > /dev/null 2>&1; then
                echo -e "${GREEN}✓${NC} $name: 运行正常"
            else
                echo -e "${YELLOW}⚠${NC} $name: 无法访问或尚未就绪"
            fi
        fi
    done
    echo ""
}

# 11. 快速修复建议
show_fix_suggestions() {
    log_section "11. 快速修复建议"

    echo "如果遇到问题，尝试以下修复:"
    echo ""
    echo "1. 运行自动修复脚本:"
    echo "   sudo ./scripts/fix-deployment.sh"
    echo ""
    echo "2. 修复权限问题:"
    echo "   sudo ./scripts/deploy.sh init"
    echo ""
    echo "3. 重启失败的服务:"
    echo "   docker compose restart <service-name>"
    echo ""
    echo "4. 查看详细故障排查指南:"
    echo "   cat TROUBLESHOOTING.md"
    echo ""
    echo "5. 完全重新部署:"
    echo "   sudo ./scripts/deploy.sh clean"
    echo "   sudo ./scripts/deploy.sh full"
    echo ""
}

# 主函数
main() {
    local target="${1:-all}"

    cd "$DEPLOY_DIR"

    case "$target" in
        all)
            show_header "完整诊断"
            check_all_containers
            diagnose_ceph_exporter
            diagnose_ceph_demo
            diagnose_prometheus
            diagnose_grafana
            diagnose_kibana
            diagnose_elasticsearch
            diagnose_jaeger
            check_system_resources
            check_services_health
            show_fix_suggestions
            ;;
        ceph-exporter)
            show_header "Ceph-Exporter 诊断"
            diagnose_ceph_exporter
            ;;
        ceph-demo)
            show_header "Ceph-Demo 诊断"
            diagnose_ceph_demo
            ;;
        prometheus)
            show_header "Prometheus 诊断"
            diagnose_prometheus
            ;;
        grafana)
            show_header "Grafana 诊断"
            diagnose_grafana
            ;;
        kibana)
            show_header "Kibana 诊断"
            diagnose_kibana
            ;;
        elasticsearch)
            show_header "Elasticsearch 诊断"
            diagnose_elasticsearch
            ;;
        jaeger)
            show_header "Jaeger 诊断"
            diagnose_jaeger
            ;;
        *)
            echo "用法: $0 [service-name]"
            echo ""
            echo "可用的服务名:"
            echo "  all (默认)      - 完整诊断所有服务"
            echo "  ceph-exporter   - 仅诊断 ceph-exporter"
            echo "  ceph-demo       - 仅诊断 ceph-demo"
            echo "  prometheus      - 仅诊断 prometheus"
            echo "  grafana         - 仅诊断 grafana"
            echo "  kibana          - 仅诊断 kibana"
            echo "  elasticsearch   - 仅诊断 elasticsearch"
            echo "  jaeger          - 仅诊断 jaeger"
            echo ""
            exit 1
            ;;
    esac

    echo ""
    echo "╚══════════════════════════════════════════════════════════════════════╝"
    echo ""
    log_info "诊断完成！"
    echo ""
}

main "$@"
