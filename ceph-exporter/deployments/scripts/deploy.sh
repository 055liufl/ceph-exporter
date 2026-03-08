#!/bin/bash
# =============================================================================
# ceph-exporter 部署脚本
# =============================================================================
# 自动检查环境、配置镜像加速、分阶段部署所有组件
# 适用环境: CentOS 7 + Docker
# =============================================================================

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DEPLOY_DIR="$(dirname "$SCRIPT_DIR")"
PROJECT_DIR="$(dirname "$DEPLOY_DIR")"

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

# 检查是否为 root 用户
check_root() {
    if [ "$EUID" -eq 0 ]; then
        log_warn "检测到以 root 用户运行"
        log_warn "建议使用普通用户并将其添加到 docker 组"
    fi
}

# 检查操作系统
check_os() {
    log_step "检查操作系统..."

    if [ -f /etc/centos-release ]; then
        local version=$(cat /etc/centos-release | grep -oP '(?<=release )\d+')
        log_info "检测到 CentOS $version"

        if [ "$version" != "7" ]; then
            log_warn "此脚本针对 CentOS 7 优化，当前版本: $version"
        fi
    else
        log_warn "未检测到 CentOS 系统"
    fi
}

# 检查 Docker 安装
check_docker() {
    log_step "检查 Docker 安装..."

    if ! command -v docker &> /dev/null; then
        log_error "Docker 未安装"
        log_info "请运行以下命令安装 Docker:"
        echo ""
        echo "  sudo yum install -y yum-utils"
        echo "  sudo yum-config-manager --add-repo https://mirrors.aliyun.com/docker-ce/linux/centos/docker-ce.repo"
        echo "  sudo yum install -y docker-ce docker-ce-cli containerd.io"
        echo "  sudo systemctl start docker"
        echo "  sudo systemctl enable docker"
        echo ""
        exit 1
    fi

    local docker_version=$(docker --version | grep -oP '\d+\.\d+\.\d+' | head -1)
    log_info "Docker 版本: $docker_version"

    # 检查 Docker 服务状态
    if ! sudo systemctl is-active --quiet docker; then
        log_warn "Docker 服务未运行"
        log_info "启动 Docker 服务..."
        sudo systemctl start docker
    fi
}

# 检查 Docker Compose 安装
check_docker_compose() {
    log_step "检查 Docker Compose 安装..."

    if command -v docker-compose &> /dev/null; then
        COMPOSE_CMD="docker-compose"
        local compose_version=$(docker-compose --version | grep -oP '\d+\.\d+\.\d+' | head -1)
        log_info "Docker Compose 版本: $compose_version"
    elif docker compose version &> /dev/null; then
        COMPOSE_CMD="docker compose"
        local compose_version=$(docker compose version | grep -oP '\d+\.\d+\.\d+' | head -1)
        log_info "Docker Compose 版本: $compose_version (v2)"
    else
        log_error "Docker Compose 未安装"
        log_info "请运行以下命令安装 Docker Compose:"
        echo ""
        echo "  sudo curl -L \"https://get.daocloud.io/docker/compose/releases/download/v2.24.0/docker-compose-\$(uname -s)-\$(uname -m)\" -o /usr/local/bin/docker-compose"
        echo "  sudo chmod +x /usr/local/bin/docker-compose"
        echo "  sudo ln -sf /usr/local/bin/docker-compose /usr/bin/docker-compose"
        echo ""
        exit 1
    fi
}

# 检查系统资源
check_resources() {
    log_step "检查系统资源..."

    # 检查内存
    local total_mem=$(free -m | awk '/^Mem:/{print $2}')
    local avail_mem=$(free -m | awk '/^Mem:/{print $7}')

    log_info "总内存: ${total_mem}MB"
    log_info "可用内存: ${avail_mem}MB"

    if [ "$avail_mem" -lt 2048 ]; then
        log_warn "可用内存不足 2GB，可能影响服务运行"
    fi

    # 检查 CPU
    local cpu_cores=$(nproc)
    log_info "CPU 核心数: $cpu_cores"

    if [ "$cpu_cores" -lt 2 ]; then
        log_warn "CPU 核心数少于 2，建议增加"
    fi

    # 检查磁盘空间
    local disk_avail=$(df -BG . | awk 'NR==2 {print $4}' | sed 's/G//')
    log_info "可用磁盘空间: ${disk_avail}GB"

    if [ "$disk_avail" -lt 20 ]; then
        log_warn "可用磁盘空间不足 20GB"
    fi
}

# 检查防火墙
check_firewall() {
    log_step "检查防火墙状态..."

    if sudo systemctl is-active --quiet firewalld; then
        log_warn "防火墙已启用"
        log_info "需要开放以下端口: 9128, 9090, 3000, 9093, 9200, 5601, 16686"

        read -p "是否自动配置防火墙规则? (y/N): " configure_fw
        if [[ "$configure_fw" == "y" || "$configure_fw" == "Y" ]]; then
            configure_firewall
        else
            log_info "请手动配置防火墙或临时关闭: sudo systemctl stop firewalld"
        fi
    else
        log_info "防火墙未启用"
    fi
}

# 配置防火墙
configure_firewall() {
    log_step "配置防火墙规则..."

    local ports=(9128 9090 3000 9093 9200 5601 16686 8080)

    for port in "${ports[@]}"; do
        sudo firewall-cmd --permanent --add-port=${port}/tcp
        log_info "已开放端口: $port"
    done

    sudo firewall-cmd --reload
    log_info "防火墙规则已重新加载"
}

# 检查 SELinux
check_selinux() {
    log_step "检查 SELinux 状态..."

    local selinux_status=$(getenforce 2>/dev/null || echo "Disabled")
    log_info "SELinux 状态: $selinux_status"

    if [ "$selinux_status" == "Enforcing" ]; then
        log_warn "SELinux 处于强制模式，可能影响 Docker 容器运行"

        read -p "是否临时禁用 SELinux? (y/N): " disable_selinux
        if [[ "$disable_selinux" == "y" || "$disable_selinux" == "Y" ]]; then
            sudo setenforce 0
            log_info "SELinux 已临时禁用"
            log_warn "重启后将恢复，如需永久禁用请编辑 /etc/selinux/config"
        fi
    fi
}

# 配置 Docker 镜像加速
configure_mirror() {
    log_step "配置 Docker 镜像加速器..."

    if [ -f /etc/docker/daemon.json ]; then
        log_info "检测到现有 Docker 配置"

        if grep -q "registry-mirrors" /etc/docker/daemon.json; then
            log_info "镜像加速器已配置"
            return
        fi
    fi

    log_info "添加国内镜像源..."
    sudo mkdir -p /etc/docker

    sudo tee /etc/docker/daemon.json > /dev/null <<EOF
{
  "registry-mirrors": [
    "https://docker.mirrors.ustc.edu.cn",
    "https://hub-mirror.c.163.com",
    "https://mirror.ccs.tencentyun.com"
  ],
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "100m",
    "max-file": "3"
  },
  "storage-driver": "overlay2"
}
EOF

    log_info "重启 Docker 服务..."
    sudo systemctl daemon-reload
    sudo systemctl restart docker

    sleep 3
    log_info "镜像加速器配置完成"
}

# 预拉取镜像
pull_images() {
    log_step "预拉取所需镜像..."

    local images=(
        "ceph/demo:latest-nautilus"
        "prom/prometheus:latest"
        "grafana/grafana:latest"
        "prom/alertmanager:latest"
        "docker.elastic.co/elasticsearch/elasticsearch:7.17.0"
        "docker.elastic.co/logstash/logstash:7.17.0"
        "docker.elastic.co/kibana/kibana:7.17.0"
        "jaegertracing/all-in-one:1.35"
    )

    for image in "${images[@]}"; do
        log_info "拉取镜像: $image"
        if docker pull "$image" 2>&1 | grep -q "error\|failed"; then
            log_warn "拉取 $image 失败，将在部署时重试"
        else
            log_info "✓ $image 拉取成功"
        fi
    done
}

# 部署最小监控栈
deploy_minimal() {
    log_step "部署最小监控栈..."

    cd "$DEPLOY_DIR"
    ${COMPOSE_CMD} -f docker-compose.yml up -d

    log_info "等待服务启动..."
    sleep 30

    show_access_info_minimal
}

# 部署集成测试环境
deploy_integration() {
    log_step "部署集成测试环境..."

    cd "$DEPLOY_DIR"
    ${COMPOSE_CMD} -f docker-compose-integration-test.yml up -d

    log_info "等待服务启动..."
    sleep 90

    show_access_info_integration
}

# 部署完整轻量级栈
deploy_full() {
    log_step "部署完整轻量级监控栈..."

    cd "$DEPLOY_DIR"
    ${COMPOSE_CMD} -f docker-compose-lightweight-full.yml up -d

    log_info "等待所有服务启动（这可能需要几分钟）..."
    sleep 120

    show_access_info_full
}

# 显示访问信息（最小栈）
show_access_info_minimal() {
    log_step "服务访问信息:"
    echo ""
    echo "监控服务:"
    echo "  Ceph Exporter:   http://localhost:9128/metrics"
    echo "  Prometheus:      http://localhost:9090"
    echo "  Grafana:         http://localhost:3000 (admin/admin)"
    echo "  Alertmanager:    http://localhost:9093"
    echo ""
}

# 显示访问信息（集成测试）
show_access_info_integration() {
    log_step "服务访问信息:"
    echo ""
    echo "Ceph 服务:"
    echo "  Ceph Dashboard:  http://localhost:8080"
    echo ""
    echo "监控服务:"
    echo "  Ceph Exporter:   http://localhost:9128/metrics"
    echo "  Prometheus:      http://localhost:9090"
    echo "  Grafana:         http://localhost:3000 (admin/admin)"
    echo ""
}

# 显示访问信息（完整栈）
show_access_info_full() {
    log_step "服务访问信息:"
    echo ""
    echo "Ceph 服务:"
    echo "  Ceph Dashboard:  http://localhost:8080"
    echo ""
    echo "监控服务:"
    echo "  Ceph Exporter:   http://localhost:9128/metrics"
    echo "  Prometheus:      http://localhost:9090"
    echo "  Grafana:         http://localhost:3000 (admin/admin)"
    echo "  Alertmanager:    http://localhost:9093"
    echo ""
    echo "日志服务:"
    echo "  Elasticsearch:   http://localhost:9200"
    echo "  Kibana:          http://localhost:5601"
    echo ""
    echo "追踪服务:"
    echo "  Jaeger UI:       http://localhost:16686"
    echo ""
}

# 查看服务状态
show_status() {
    log_step "服务状态:"

    cd "$DEPLOY_DIR"

    # 尝试查找正在运行的 compose 文件
    if docker ps --format '{{.Label "com.docker.compose.project"}}' | grep -q "deployments"; then
        # 检测使用的 compose 文件
        if docker ps --format '{{.Names}}' | grep -q "kibana"; then
            ${COMPOSE_CMD} -f docker-compose-lightweight-full.yml ps
        elif docker ps --format '{{.Names}}' | grep -q "ceph-demo"; then
            ${COMPOSE_CMD} -f docker-compose-integration-test.yml ps
        else
            ${COMPOSE_CMD} -f docker-compose.yml ps
        fi
    else
        log_warn "未检测到运行中的服务"
        docker ps
    fi
}

# 查看日志
show_logs() {
    cd "$DEPLOY_DIR"

    local service="${1:-}"

    # 检测使用的 compose 文件
    local compose_file="docker-compose.yml"
    if docker ps --format '{{.Names}}' | grep -q "kibana"; then
        compose_file="docker-compose-lightweight-full.yml"
    elif docker ps --format '{{.Names}}' | grep -q "ceph-demo"; then
        compose_file="docker-compose-integration-test.yml"
    fi

    if [ -n "$service" ]; then
        ${COMPOSE_CMD} -f "$compose_file" logs -f "$service"
    else
        ${COMPOSE_CMD} -f "$compose_file" logs -f
    fi
}

# 验证部署
verify_deployment() {
    log_step "验证部署状态..."

    local services=(
        "ceph-exporter:9128/health"
        "prometheus:9090/-/healthy"
        "grafana:3000/api/health"
    )

    # 检查是否部署了完整栈
    if docker ps --format '{{.Names}}' | grep -q "elasticsearch"; then
        services+=(
            "elasticsearch:9200"
            "kibana:5601/api/status"
            "jaeger:16686"
        )
    fi

    # 检查是否部署了 Ceph Demo
    if docker ps --format '{{.Names}}' | grep -q "ceph-demo"; then
        services+=("ceph-demo:8080")
    fi

    echo ""
    for service in "${services[@]}"; do
        local name="${service%%:*}"
        local endpoint="${service#*:}"

        log_info "检查 ${name}..."
        if curl -sf "http://localhost:${endpoint}" > /dev/null 2>&1; then
            echo -e "${GREEN}✓${NC} ${name} 运行正常"
        else
            echo -e "${YELLOW}✗${NC} ${name} 无法访问或尚未就绪"
        fi
    done
    echo ""
}

# 停止服务
stop_services() {
    log_step "停止服务..."

    cd "$DEPLOY_DIR"

    # 检测并停止所有可能运行的 compose 配置
    for compose_file in docker-compose.yml docker-compose-integration-test.yml docker-compose-lightweight-full.yml; do
        if [ -f "$compose_file" ]; then
            ${COMPOSE_CMD} -f "$compose_file" down 2>/dev/null || true
        fi
    done

    log_info "所有服务已停止"
}

# 清理数据
clean_data() {
    log_warn "此操作将停止所有服务并删除持久化数据"
    read -p "确认继续? (y/N): " confirm
    if [[ "$confirm" != "y" && "$confirm" != "Y" ]]; then
        log_info "操作已取消"
        exit 0
    fi

    log_step "停止服务并清除数据卷..."

    cd "$DEPLOY_DIR"

    # 停止并删除所有配置的数据卷
    for compose_file in docker-compose.yml docker-compose-integration-test.yml docker-compose-lightweight-full.yml; do
        if [ -f "$compose_file" ]; then
            ${COMPOSE_CMD} -f "$compose_file" down -v 2>/dev/null || true
        fi
    done

    log_info "所有服务已停止，数据卷已清除"
}

# 显示帮助
show_help() {
    cat << 'EOF'
ceph-exporter 部署脚本

用法:
  ./deploy.sh <command>

命令:
  check           检查系统环境（Docker、资源、防火墙等）
  mirror          配置 Docker 镜像加速器
  pull            预拉取所有镜像
  minimal         部署最小监控栈
  integration     部署集成测试环境
  full            部署完整轻量级栈（推荐）
  status          查看服务状态
  logs [service]  查看日志（可指定服务名）
  verify          验证部署状态
  stop            停止所有服务
  clean           停止服务并清除数据卷
  help            显示此帮助信息

示例:
  # 完整部署（推荐）
  ./deploy.sh full

  # 检查环境
  ./deploy.sh check

  # 查看服务状态
  ./deploy.sh status

  # 查看特定服务日志
  ./deploy.sh logs ceph-exporter

  # 验证部署
  ./deploy.sh verify
EOF
}

# 完整环境检查
full_check() {
    check_root
    check_os
    check_docker
    check_docker_compose
    check_resources
    check_firewall
    check_selinux

    log_info "环境检查完成"
}

# 主函数
main() {
    case "${1:-help}" in
        check)
            full_check
            ;;
        mirror)
            check_docker
            configure_mirror
            ;;
        pull)
            check_docker
            check_docker_compose
            pull_images
            ;;
        minimal)
            full_check
            configure_mirror
            deploy_minimal
            verify_deployment
            ;;
        integration)
            full_check
            configure_mirror
            pull_images
            deploy_integration
            verify_deployment
            ;;
        full)
            full_check
            configure_mirror
            pull_images
            deploy_full
            verify_deployment
            ;;
        status)
            check_docker_compose
            show_status
            ;;
        logs)
            check_docker_compose
            shift
            show_logs "$@"
            ;;
        verify)
            verify_deployment
            ;;
        stop)
            check_docker_compose
            stop_services
            ;;
        clean)
            check_docker_compose
            clean_data
            ;;
        help|*)
            show_help
            ;;
    esac
}

main "$@"
