#!/bin/bash
# =============================================================================
# ceph-exporter 诊断脚本
# =============================================================================
# 收集所有服务的状态和日志信息，帮助诊断问题
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

echo "╔══════════════════════════════════════════════════════════════════════╗"
echo "║              ceph-exporter 部署诊断报告                             ║"
echo "╚══════════════════════════════════════════════════════════════════════╝"
echo ""
echo "诊断时间: $(date '+%Y-%m-%d %H:%M:%S')"
echo ""

# 1. 检查容器状态
log_step "1. 容器状态"
echo ""
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" 2>/dev/null || {
    log_error "无法访问 Docker，请确保 Docker 服务正在运行"
    exit 1
}
echo ""

# 2. 检查 ceph-exporter 状态
log_step "2. ceph-exporter 详细状态"
echo ""
CEPH_EXPORTER_STATUS=$(docker inspect ceph-exporter --format='{{.State.Status}}' 2>/dev/null || echo "not found")
CEPH_EXPORTER_RESTARTS=$(docker inspect ceph-exporter --format='{{.RestartCount}}' 2>/dev/null || echo "0")

echo "状态: $CEPH_EXPORTER_STATUS"
echo "重启次数: $CEPH_EXPORTER_RESTARTS"
echo ""

if [ "$CEPH_EXPORTER_STATUS" != "running" ]; then
    log_warn "ceph-exporter 未正常运行"
    echo ""
    log_step "ceph-exporter 最近日志:"
    echo ""
    docker logs ceph-exporter --tail 30 2>&1
    echo ""
fi

# 3. 检查 Ceph 集群状态
log_step "3. Ceph 集群状态"
echo ""
CEPH_DEMO_STATUS=$(docker inspect ceph-demo --format='{{.State.Status}}' 2>/dev/null || echo "not found")
echo "ceph-demo 容器状态: $CEPH_DEMO_STATUS"
echo ""

if [ "$CEPH_DEMO_STATUS" = "running" ]; then
    log_info "检查 Ceph 集群健康状态..."
    echo ""
    docker exec ceph-demo ceph -s 2>&1 || {
        log_warn "Ceph 集群尚未就绪"
        echo ""
        log_step "ceph-demo 最近日志:"
        echo ""
        docker logs ceph-demo --tail 20 2>&1
    }
else
    log_error "ceph-demo 容器未运行"
fi
echo ""

# 4. 检查 Ceph 配置文件
log_step "4. Ceph 配置文件检查"
echo ""
log_info "ceph-demo 配置文件:"
docker exec ceph-demo ls -la /etc/ceph/ 2>&1 || log_error "无法访问 ceph-demo 配置目录"
echo ""

log_info "ceph-exporter 配置文件:"
docker exec ceph-exporter ls -la /etc/ceph/ 2>&1 || log_warn "ceph-exporter 无法访问配置目录（容器可能在重启）"
echo ""

# 5. 检查网络连接
log_step "5. 网络连接检查"
echo ""
log_info "检查 ceph-exporter 到 ceph-demo 的网络连接..."
docker exec ceph-exporter ping -c 3 ceph-demo 2>&1 || log_warn "网络连接测试失败（容器可能在重启）"
echo ""

# 6. 检查数据卷挂载
log_step "6. 数据卷挂载检查"
echo ""
log_info "ceph-exporter 挂载信息:"
docker inspect ceph-exporter --format='{{range .Mounts}}{{.Type}}: {{.Source}} -> {{.Destination}}{{"\n"}}{{end}}' 2>&1
echo ""

log_info "ceph-demo 挂载信息:"
docker inspect ceph-demo --format='{{range .Mounts}}{{.Type}}: {{.Source}} -> {{.Destination}}{{"\n"}}{{end}}' 2>&1
echo ""

# 7. 检查其他服务状态
log_step "7. 其他服务健康检查"
echo ""

services=(
    "prometheus:9090/-/healthy"
    "grafana:3000/api/health"
    "elasticsearch:9200"
    "kibana:5601/api/status"
    "jaeger:16686"
)

for service in "${services[@]}"; do
    name="${service%%:*}"
    endpoint="${service#*:}"

    if curl -sf --max-time 3 "http://localhost:${endpoint}" > /dev/null 2>&1; then
        echo -e "${GREEN}✓${NC} $name: 运行正常"
    else
        echo -e "${YELLOW}⚠${NC} $name: 无法访问或尚未就绪"
    fi
done
echo ""

# 8. 资源使用情况
log_step "8. 资源使用情况"
echo ""
docker stats --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.MemPerc}}" 2>&1
echo ""

# 9. 诊断建议
log_step "9. 诊断建议"
echo ""

if [ "$CEPH_EXPORTER_RESTARTS" -gt 5 ]; then
    log_error "ceph-exporter 重启次数过多 ($CEPH_EXPORTER_RESTARTS 次)"
    echo ""
    echo "可能的原因:"
    echo "  1. Ceph 集群尚未完全启动"
    echo "  2. 配置文件未正确挂载"
    echo "  3. 网络连接问题"
    echo ""
    echo "建议操作:"
    echo "  1. 等待 2-3 分钟让 Ceph 集群完全启动"
    echo "  2. 检查 ceph-exporter 日志: docker logs ceph-exporter"
    echo "  3. 重启 ceph-exporter: docker-compose restart ceph-exporter"
    echo ""
fi

# 10. 快速修复命令
log_step "10. 快速修复命令"
echo ""
echo "如果 Ceph 集群已就绪但 ceph-exporter 仍在重启，执行:"
echo ""
echo "  cd /home/lfl/ceph-exporter/ceph-exporter/deployments"
echo "  docker-compose -f docker-compose-lightweight-full.yml restart ceph-exporter"
echo "  docker logs -f ceph-exporter"
echo ""
echo "如果需要完全重新部署:"
echo ""
echo "  cd /home/lfl/ceph-exporter/ceph-exporter/deployments"
echo "  docker-compose -f docker-compose-lightweight-full.yml down"
echo "  docker-compose -f docker-compose-lightweight-full.yml up -d"
echo ""

echo "╚══════════════════════════════════════════════════════════════════════╝"
