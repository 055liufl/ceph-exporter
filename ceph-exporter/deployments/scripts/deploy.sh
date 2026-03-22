#!/bin/bash
# =============================================================================
# ceph-exporter 部署脚本
# =============================================================================
# 自动检查环境、配置镜像加速、分阶段部署所有组件
# 适用环境: Ubuntu 20.04 + Docker
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
    echo -e "${GREEN}[信息]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[警告]${NC} $1"
}

log_error() {
    echo -e "${RED}[错误]${NC} $1"
}

log_step() {
    echo -e "${BLUE}[步骤]${NC} $1"
}

# 打印横幅
print_banner() {
    echo -e "${GREEN}"
    echo "=============================================="
    echo "  Ceph Exporter 中文监控系统"
    echo "=============================================="
    echo -e "${NC}"
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

    if [ -f /etc/os-release ]; then
        source /etc/os-release
        log_info "检测到 $PRETTY_NAME"
        if [[ "$ID" != "ubuntu" ]] || [[ "$VERSION_ID" != "20.04" ]]; then
            log_warn "此脚本针对 Ubuntu 20.04 优化，当前系统: $PRETTY_NAME"
        fi
    else
        log_warn "未检测到系统版本信息"
    fi
}

# 检查 Docker 安装
check_docker() {
    log_step "检查 Docker 安装..."

    if ! command -v docker &> /dev/null; then
        log_error "Docker 未安装"
        log_info "请执行以下命令安装 Docker:"
        echo "  sudo apt-get update"
        echo "  sudo apt-get install -y ca-certificates curl gnupg"
        echo "  sudo install -m 0755 -d /etc/apt/keyrings"
        echo "  curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg"
        echo "  sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin"
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

    if docker compose version &> /dev/null; then
        COMPOSE_CMD="docker compose"
        log_info "Docker Compose 插件已安装: $(docker compose version)"
    elif command -v docker-compose &> /dev/null; then
        COMPOSE_CMD="docker-compose"
        log_warn "检测到旧版 docker-compose，建议升级到 Docker Compose 插件"
    else
        log_error "Docker Compose 未安装"
        log_info "请执行: sudo apt-get install -y docker-compose-plugin"
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

    local ports=(9128 9090 3000 9093 9200 5601 16686 8080)

    if command -v ufw &> /dev/null; then
        log_info "配置 UFW 防火墙规则..."
        for port in "${ports[@]}"; do
            sudo ufw allow "$port"/tcp comment "ceph-exporter" 2>/dev/null || true
        done
        log_info "防火墙规则配置完成"
    else
        log_info "未检测到 UFW，跳过防火墙配置"
    fi
}

# 检查 SELinux
check_selinux() {
    # Ubuntu 20.04 使用 AppArmor，无需 SELinux 配置
    log_info "跳过 SELinux 检查 (Ubuntu 使用 AppArmor)"
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
        "quay.io/ceph/daemon:latest-octopus"
        "prom/prometheus:v2.51.0"
        "grafana/grafana:10.4.0"
        "prom/alertmanager:v0.25.0"
        "docker.elastic.co/elasticsearch/elasticsearch:7.17.0"
        "docker.elastic.co/logstash/logstash:7.17.0"
        "docker.elastic.co/kibana/kibana:7.17.0"
        "docker.elastic.co/beats/filebeat:7.17.0"
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

# 初始化数据目录
init_data_dirs() {
    log_step "初始化数据目录..."

    cd "$DEPLOY_DIR"

    # 创建数据目录
    log_info "创建数据目录结构..."
    mkdir -p data/{ceph-demo/{data,config},prometheus,grafana,alertmanager,elasticsearch}
    mkdir -p data/test/{ceph-demo/{data,config},prometheus,grafana}

    # 设置权限
    log_info "设置目录权限..."

    # Grafana 需要 472 用户权限
    if command -v sudo &> /dev/null; then
        sudo chown -R 472:472 data/grafana data/test/grafana 2>/dev/null || {
            log_warn "无法设置 Grafana 目录权限，可能需要手动设置"
        }

        # Elasticsearch 需要 1000 用户权限
        sudo chown -R 1000:1000 data/elasticsearch 2>/dev/null || {
            log_warn "无法设置 Elasticsearch 目录权限，可能需要手动设置"
        }

        # Prometheus 需要 nobody 用户权限 (UID 65534)
        sudo chown -R 65534:65534 data/prometheus data/test/prometheus 2>/dev/null || {
            log_warn "无法设置 Prometheus 目录权限，可能需要手动设置"
        }

        # Alertmanager 使用当前用户权限
        sudo chown -R $USER:$USER data/alertmanager 2>/dev/null || true
    else
        log_warn "未检测到 sudo，跳过权限设置"
    fi

    # 创建 configs 目录软链接（如果不存在）
    if [ ! -e configs ]; then
        log_info "创建 configs 目录软链接..."
        ln -s ../configs configs 2>/dev/null || {
            log_warn "无法创建 configs 软链接，请确保 ../configs 目录存在"
        }
    fi

    log_info "数据目录初始化完成"
    log_info "数据存储位置: $DEPLOY_DIR/data/"
    log_info "时区配置: 所有容器已自动挂载宿主机时区 (/etc/localtime, /etc/timezone)"
}

# 部署最小监控栈
deploy_minimal() {
    log_step "部署最小监控栈..."

    init_data_dirs

    cd "$DEPLOY_DIR"
    ${COMPOSE_CMD} -f docker-compose.yml up -d

    log_info "等待服务启动..."
    sleep 30

    show_access_info_minimal
}

# 部署集成测试环境
deploy_integration() {
    log_step "部署集成测试环境..."

    init_data_dirs

    cd "$DEPLOY_DIR"
    ${COMPOSE_CMD} -f docker-compose-integration-test.yml up -d

    log_info "等待服务启动..."
    sleep 90

    # 修复 Ceph keyring 文件权限
    # 说明：集成测试环境也包含 Ceph Demo，需要修复 keyring 权限
    log_info "修复 Ceph keyring 文件权限..."
    if [ -f "data/ceph-demo/config/ceph.client.admin.keyring" ]; then
        chmod 644 data/ceph-demo/config/ceph.client.admin.keyring
        log_info "✓ ceph.client.admin.keyring 权限已修复"
    fi
    if [ -f "data/ceph-demo/config/ceph.mon.keyring" ]; then
        chmod 644 data/ceph-demo/config/ceph.mon.keyring
        log_info "✓ ceph.mon.keyring 权限已修复"
    fi

    # 重启 ceph-exporter 以应用权限修复
    log_info "重启 ceph-exporter 服务..."
    ${COMPOSE_CMD} -f docker-compose-integration-test.yml restart ceph-exporter

    # 等待 ceph-exporter 重启完成
    log_info "等待 ceph-exporter 启动..."
    sleep 10

    show_access_info_integration
}

# 部署完整轻量级栈
deploy_full() {
    log_step "部署完整轻量级监控栈..."

    init_data_dirs

    # 设置系统参数（Elasticsearch 需要）
    log_info "设置系统参数..."
    sudo sysctl -w vm.max_map_count=262144 2>/dev/null || log_warn "无法设置 vm.max_map_count，可能需要 root 权限"

    # 修复数据目录权限（在启动前）
    log_info "修复数据目录权限..."
    [ -d "data/elasticsearch" ] && sudo chown -R 1000:1000 data/elasticsearch 2>/dev/null || true
    [ -d "data/prometheus" ] && sudo chown -R 65534:65534 data/prometheus 2>/dev/null || true
    [ -d "data/grafana" ] && sudo chown -R 472:472 data/grafana 2>/dev/null || true
    [ -d "data/ceph-demo/config" ] && sudo chmod -R 755 data/ceph-demo/config 2>/dev/null || true

    cd "$DEPLOY_DIR"

    # 检查 ceph-exporter:dev 镜像是否存在，不存在则构建
    if docker image inspect ceph-exporter:dev >/dev/null 2>&1; then
        log_info "镜像 ceph-exporter:dev 已存在，跳过构建"
    else
        log_info "镜像 ceph-exporter:dev 不存在，开始构建..."
        cd "$PROJECT_DIR"
        docker build -t ceph-exporter:dev -f Dockerfile . || {
            log_error "构建 ceph-exporter:dev 镜像失败"
            exit 1
        }
    fi

    cd "$DEPLOY_DIR"

    # 选择日志方案
    local logging_mode="${LOGGING_MODE:-}"
    if [ -z "$logging_mode" ]; then
        echo ""
        echo -e "${YELLOW}请选择日志收集方案:${NC}"
        echo "  1) container  - 容器日志收集（推荐，通过 Filebeat sidecar 采集）"
        echo "  2) direct     - 直接推送到 Logstash (TCP，无需 Filebeat)"
        echo "  3) direct-udp - 直接推送到 Logstash (UDP，高性能)"
        echo "  4) file       - 文件日志 + Filebeat（日志持久化）"
        echo "  5) dev        - 开发模式（stdout + text，方便调试）"
        echo ""
        read -p "请输入选项 [1]: " logging_choice
        case "${logging_choice:-1}" in
            1|container)
                logging_mode="container"
                ;;
            2|direct)
                logging_mode="direct"
                ;;
            3|direct-udp)
                logging_mode="direct-udp"
                ;;
            4|file)
                logging_mode="file"
                ;;
            5|dev)
                logging_mode="dev"
                ;;
            *)
                logging_mode="container"
                ;;
        esac
    fi

    # 应用日志方案配置
    log_info "应用日志方案: $logging_mode"
    "$SCRIPT_DIR/switch-logging-mode.sh" "$logging_mode"

    # 启动核心服务（不包含 filebeat-sidecar）
    ${COMPOSE_CMD} -f docker-compose-lightweight-full.yml up -d \
        ceph-demo ceph-exporter prometheus grafana alertmanager \
        elasticsearch logstash kibana jaeger

    log_info "等待 Elasticsearch 启动..."
    sleep 30

    log_info "等待 ceph-demo 生成配置..."
    sleep 30

    # 修复 Ceph keyring 文件权限
    log_info "修复 Ceph keyring 文件权限..."
    if [ -f "data/ceph-demo/config/ceph.client.admin.keyring" ]; then
        chmod 644 data/ceph-demo/config/ceph.client.admin.keyring
        log_info "✓ ceph.client.admin.keyring 权限已修复"
    fi
    if [ -f "data/ceph-demo/config/ceph.mon.keyring" ]; then
        chmod 644 data/ceph-demo/config/ceph.mon.keyring
        log_info "✓ ceph.mon.keyring 权限已修复"
    fi

    # 重启 ceph-exporter 以应用权限修复和网络配置
    log_info "重启 ceph-exporter 服务..."
    ${COMPOSE_CMD} -f docker-compose-lightweight-full.yml up -d ceph-exporter

    # 根据日志方案决定是否启动 filebeat-sidecar
    if [ "$logging_mode" = "container" ]; then
        log_info "启动 Filebeat sidecar（容器日志收集模式）..."
        ${COMPOSE_CMD} -f docker-compose-lightweight-full.yml up -d filebeat-sidecar
    else
        log_info "当前日志方案: $logging_mode，不需要 Filebeat sidecar"
        # 确保 filebeat-sidecar 未运行
        ${COMPOSE_CMD} -f docker-compose-lightweight-full.yml stop filebeat-sidecar 2>/dev/null || true
        ${COMPOSE_CMD} -f docker-compose-lightweight-full.yml rm -f filebeat-sidecar 2>/dev/null || true
    fi

    log_info "等待所有服务就绪..."
    sleep 20

    # 生成测试数据
    log_info "生成测试数据..."
    for i in {1..10}; do
        curl -s http://localhost:9128/metrics > /dev/null 2>&1 || true
        sleep 0.5
    done

    log_info "等待数据推送到 ELK 和 Jaeger..."
    sleep 10

    show_access_info_full

    # 显示额外的使用提示
    echo ""
    log_info "在 Kibana 中创建索引模式:"
    echo "  1. 访问 http://localhost:5601"
    echo "  2. Stack Management → 索引模式"
    echo "  3. 创建索引模式: ceph-exporter-*"
    echo "  4. 选择时间字段: @timestamp"
    echo ""
    log_info "在 Jaeger UI 中查看追踪:"
    echo "  1. 访问 http://localhost:16686"
    echo "  2. Service 选择: ceph-exporter"
    echo "  3. 点击 'Find Traces'"
    echo "  4. 查看追踪数据（包含响应状态码和 Span 状态）"
    echo ""
}

# 显示访问信息（最小栈）
show_access_info_minimal() {
    log_step "服务访问信息:"
    echo ""
    echo -e "${GREEN}访问地址：${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo -e "${BLUE}📊 Grafana 监控仪表盘（中文）：${NC}"
    echo "   http://localhost:3000"
    echo "   账号：admin / admin"
    echo ""
    echo -e "${BLUE}📈 Prometheus 指标查询：${NC}"
    echo "   http://localhost:9090"
    echo ""
    echo -e "${BLUE}🔔 Alertmanager 告警管理（中文）：${NC}"
    echo "   http://localhost:9093"
    echo ""
    echo -e "${BLUE}🔌 Ceph Exporter：${NC}"
    echo "   http://localhost:9128/metrics"
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    log_info "提示：首次启动可能需要 1-2 分钟初始化"
    echo ""
}

# 显示访问信息（集成测试）
show_access_info_integration() {
    log_step "服务访问信息:"
    echo ""
    echo -e "${GREEN}访问地址：${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo -e "${BLUE}🗄️  Ceph RGW (S3 对象存储 API)：${NC}"
    echo "   http://localhost:8080"
    echo ""
    echo -e "${BLUE}📊 Grafana 监控仪表盘（中文）：${NC}"
    echo "   http://localhost:3000"
    echo "   账号：admin / admin"
    echo ""
    echo -e "${BLUE}📈 Prometheus 指标查询：${NC}"
    echo "   http://localhost:9090"
    echo ""
    echo -e "${BLUE}🔌 Ceph Exporter：${NC}"
    echo "   http://localhost:9128/metrics"
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    log_info "提示：首次启动可能需要 1-2 分钟初始化"
    echo ""
}

# 显示访问信息（完整栈）
show_access_info_full() {
    log_step "服务启动成功！"
    echo ""
    echo -e "${GREEN}访问地址：${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo -e "${BLUE}🗄️  Ceph RGW (S3 对象存储 API)：${NC}"
    echo "   http://localhost:8080"
    echo ""
    echo -e "${BLUE}📊 Grafana 监控仪表盘（中文）：${NC}"
    echo "   http://localhost:3000"
    echo "   账号：admin / admin"
    echo ""
    echo -e "${BLUE}📈 Prometheus 指标查询：${NC}"
    echo "   http://localhost:9090"
    echo ""
    echo -e "${BLUE}🔔 Alertmanager 告警管理（中文）：${NC}"
    echo "   http://localhost:9093"
    echo ""
    echo -e "${BLUE}📋 Kibana 日志分析（中文）：${NC}"
    echo "   http://localhost:5601"
    echo ""
    echo -e "${BLUE}🔍 Jaeger 链路追踪：${NC}"
    echo "   http://localhost:16686"
    echo ""
    echo -e "${BLUE}🔌 Elasticsearch：${NC}"
    echo "   http://localhost:9200"
    echo ""
    echo -e "${BLUE}🔌 Ceph Exporter：${NC}"
    echo "   http://localhost:9128/metrics"
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    log_info "提示：首次启动可能需要 1-2 分钟初始化"
    log_info "所有服务已配置中文界面，包括 Grafana、Prometheus 告警、Alertmanager"
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
        "alertmanager:9093/-/healthy"
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

        # 特殊处理 ceph-demo：RGW 根路径返回 404 是正常的
        if [ "$name" = "ceph-demo" ]; then
            if curl -s "http://localhost:${endpoint}" 2>&1 | grep -q "NoSuchBucket\|InvalidBucketName\|404"; then
                echo -e "${GREEN}✓${NC} ${name} 运行正常"
            else
                echo -e "${YELLOW}✗${NC} ${name} 无法访问或尚未就绪"
            fi
        else
            if curl -sf "http://localhost:${endpoint}" > /dev/null 2>&1; then
                echo -e "${GREEN}✓${NC} ${name} 运行正常"
            else
                echo -e "${YELLOW}✗${NC} ${name} 无法访问或尚未就绪"
            fi
        fi
    done
    echo ""

    # 验证时区配置
    log_info "验证容器时区配置..."
    if docker ps --format '{{.Names}}' | grep -q "ceph-exporter"; then
        local host_tz=$(date +"%Z %z")
        local container_tz=$(docker exec ceph-exporter date +"%Z %z" 2>/dev/null || echo "无法获取")
        if [ "$host_tz" = "$container_tz" ]; then
            echo -e "${GREEN}✓${NC} 容器时区与宿主机一致: $host_tz"
        else
            echo -e "${YELLOW}!${NC} 宿主机时区: $host_tz, 容器时区: $container_tz"
        fi
    fi
    echo ""
}

# 停止服务
stop_services() {
    log_step "停止服务..."

    cd "$DEPLOY_DIR"

    # 检测并停止所有可能运行的 compose 配置
    for compose_file in docker-compose.yml docker-compose-integration-test.yml docker-compose-lightweight-full.yml docker-compose-ceph-demo.yml; do
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

    log_step "停止服务并清除数据..."

    cd "$DEPLOY_DIR"

    # 停止所有服务
    for compose_file in docker-compose.yml docker-compose-integration-test.yml docker-compose-lightweight-full.yml docker-compose-ceph-demo.yml; do
        if [ -f "$compose_file" ]; then
            ${COMPOSE_CMD} -f "$compose_file" down 2>/dev/null || true
        fi
    done

    # 删除数据目录
    if [ -d "data" ]; then
        log_info "删除数据目录..."
        rm -rf data/
        log_info "数据目录已删除"
    fi

    log_info "所有服务已停止，数据卷已清除"
}

# 显示帮助
show_help() {
    print_banner
    cat << 'EOF'

用法:
  ./deploy.sh <命令>

命令:
  check           检查系统环境（Docker、资源、防火墙等）
  mirror          配置 Docker 镜像加速器
  pull            预拉取所有镜像
  init            初始化数据目录（创建目录并设置权限）
  minimal         部署最小监控栈（标准部署，需要现有 Ceph 集群）
  integration     部署集成测试环境（包含 Ceph Demo）
  full            部署完整轻量级栈（包含 Ceph Demo + ELK + Jaeger，推荐）
  status          查看服务状态
  logs [service]  查看日志（可指定服务名）
  verify          验证部署状态
  diagnose [svc]  诊断服务问题（可指定服务名或 all）
  fix             修复常见部署问题（权限、配置等）
  stop            停止所有服务
  clean           停止服务并清除数据
  help            显示此帮助信息

别名（为了兼容性）:
  standard        等同于 minimal（标准部署）
  test            等同于 integration（集成测试）

示例:
  # 完整部署（推荐，包含中文界面）
  ./deploy.sh full

  # 完整部署 - 指定日志方案（跳过交互选择）
  LOGGING_MODE=container ./deploy.sh full     # 容器日志收集（推荐）
  LOGGING_MODE=direct ./deploy.sh full        # 直接推送到 Logstash (TCP)
  LOGGING_MODE=direct-udp ./deploy.sh full    # 直接推送到 Logstash (UDP)
  LOGGING_MODE=file ./deploy.sh full          # 文件日志 + Filebeat
  LOGGING_MODE=dev ./deploy.sh full           # 开发模式

  # 标准部署（连接现有 Ceph 集群）
  ./deploy.sh minimal
  # 或
  ./deploy.sh standard

  # 集成测试环境
  ./deploy.sh integration
  # 或
  ./deploy.sh test

  # 检查环境
  ./deploy.sh check

  # 查看服务状态
  ./deploy.sh status

  # 查看特定服务日志
  ./deploy.sh logs ceph-exporter

  # 完整诊断
  ./deploy.sh diagnose

  # 诊断特定服务
  ./deploy.sh diagnose ceph-exporter

  # 验证部署
  ./deploy.sh verify

  # 修复部署问题
  ./deploy.sh fix

中文界面支持:
  所有部署模式都已配置中文界面，包括：
  - Grafana 监控仪表盘（中文界面和 Dashboard）
  - Prometheus 告警规则（中文描述）
  - Alertmanager 告警管理（中文配置）
  - Kibana 日志分析（中文界面支持）

故障排查:
  如果遇到部署问题，请查看:
  - 故障排查指南: cat TROUBLESHOOTING.md
  - 运行修复脚本: ./deploy.sh fix
  - 查看服务日志: ./deploy.sh logs <service-name>
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
        init)
            init_data_dirs
            ;;
        minimal|standard)
            full_check
            configure_mirror
            deploy_minimal
            verify_deployment
            ;;
        integration|test)
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
        diagnose)
            shift
            log_info "运行诊断脚本..."
            exec "$SCRIPT_DIR/diagnose.sh" "$@"
            ;;
        fix)
            log_info "运行修复脚本..."
            exec "$SCRIPT_DIR/fix-deployment.sh"
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
