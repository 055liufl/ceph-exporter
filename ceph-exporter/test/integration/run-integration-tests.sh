#!/bin/bash
# =============================================================================
# 集成测试运行脚本
# =============================================================================
# 提供便捷的集成测试执行和环境管理功能
#
# 使用方式:
#   ./run-integration-tests.sh          运行所有集成测试
#   ./run-integration-tests.sh setup    仅设置测试环境
#   ./run-integration-tests.sh test     仅运行测试（假设环境已启动）
#   ./run-integration-tests.sh cleanup  仅清理测试环境
#   ./run-integration-tests.sh help     显示帮助信息
# =============================================================================

set -euo pipefail

# 脚本所在目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$(dirname "$SCRIPT_DIR")")"
DEPLOY_DIR="${PROJECT_DIR}/deployments"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# ---------------------------------------------------------------------------
# 辅助函数
# ---------------------------------------------------------------------------

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

# 检查前置条件
check_prerequisites() {
    log_step "检查前置条件..."

    # 检查 Docker
    if ! command -v docker &> /dev/null; then
        log_error "Docker 未安装"
        exit 1
    fi

    # 检查 Docker Compose
    if docker compose version &> /dev/null; then
        COMPOSE_CMD="docker compose"
    elif command -v docker-compose &> /dev/null; then
        COMPOSE_CMD="docker-compose"
    else
        echo "Error: Docker Compose not found"
        exit 1
    fi

    # 检查 Go
    if ! command -v go &> /dev/null; then
        log_error "Go 未安装"
        exit 1
    fi

    log_info "前置条件检查通过"
    log_info "Docker: $(docker --version)"
    log_info "Docker Compose: $(${COMPOSE_CMD} version --short 2>/dev/null || echo 'unknown')"
    log_info "Go: $(go version)"
}

# 设置测试环境
setup_environment() {
    log_step "设置测试环境..."

    # 清理旧环境
    log_info "清理旧的测试环境..."
    ${COMPOSE_CMD} -f "${DEPLOY_DIR}/docker-compose-integration-test.yml" down -v 2>/dev/null || true

    # 启动服务
    log_info "启动 Docker Compose 服务..."
    ${COMPOSE_CMD} -f "${DEPLOY_DIR}/docker-compose-integration-test.yml" up -d --build

    # 等待服务启动
    log_info "等待服务启动（45秒）..."
    sleep 45

    # 检查服务状态
    log_info "检查服务状态..."
    ${COMPOSE_CMD} -f "${DEPLOY_DIR}/docker-compose-integration-test.yml" ps

    log_info "测试环境设置完成"
}

# 运行集成测试
run_tests() {
    log_step "运行集成测试..."

    cd "${SCRIPT_DIR}"

    # 运行测试
    if go test -v -timeout 30m; then
        log_info "集成测试通过 ✓"
        return 0
    else
        log_error "集成测试失败 ✗"
        return 1
    fi
}

# 清理测试环境
cleanup_environment() {
    log_step "清理测试环境..."

    ${COMPOSE_CMD} -f "${DEPLOY_DIR}/docker-compose-integration-test.yml" down

    log_info "测试环境清理完成"
}

# 收集日志
collect_logs() {
    log_step "收集容器日志..."

    local log_dir="${SCRIPT_DIR}/logs"
    mkdir -p "${log_dir}"

    local timestamp=$(date +%Y%m%d_%H%M%S)
    local log_file="${log_dir}/integration-test-${timestamp}.log"

    ${COMPOSE_CMD} -f "${DEPLOY_DIR}/docker-compose-integration-test.yml" logs > "${log_file}"

    log_info "日志已保存到: ${log_file}"
}

# 显示帮助信息
show_help() {
    cat << 'EOF'
集成测试运行脚本

使用方式:
  ./run-integration-tests.sh [command]

命令:
  (无参数)    运行完整的集成测试流程（设置 -> 测试 -> 清理）
  setup       仅设置测试环境
  test        仅运行测试（假设环境已启动）
  cleanup     仅清理测试环境
  logs        收集容器日志
  help        显示此帮助信息

示例:
  # 运行完整测试
  ./run-integration-tests.sh

  # 手动控制测试流程
  ./run-integration-tests.sh setup
  ./run-integration-tests.sh test
  ./run-integration-tests.sh cleanup

  # 收集日志
  ./run-integration-tests.sh logs

环境变量:
  SKIP_CLEANUP=1    测试后不清理环境（用于调试）

注意事项:
  - 需要 Docker、Docker Compose 和 Go 环境
  - 确保端口 9128、9090、3000、9093 未被占用
  - 测试需要 5-10 分钟完成
EOF
}

# ---------------------------------------------------------------------------
# 主流程
# ---------------------------------------------------------------------------

main() {
    local command="${1:-all}"

    case "$command" in
        setup)
            check_prerequisites
            setup_environment
            ;;
        test)
            check_prerequisites
            run_tests
            ;;
        cleanup)
            check_prerequisites
            cleanup_environment
            ;;
        logs)
            check_prerequisites
            collect_logs
            ;;
        help)
            show_help
            ;;
        all|*)
            check_prerequisites

            # 设置环境
            setup_environment

            # 运行测试
            local test_result=0
            run_tests || test_result=$?

            # 收集日志（如果测试失败）
            if [ $test_result -ne 0 ]; then
                log_warn "测试失败，收集日志..."
                collect_logs
            fi

            # 清理环境（除非设置了 SKIP_CLEANUP）
            if [ "${SKIP_CLEANUP:-0}" != "1" ]; then
                cleanup_environment
            else
                log_warn "跳过环境清理（SKIP_CLEANUP=1）"
            fi

            # 返回测试结果
            exit $test_result
            ;;
    esac
}

# 执行主流程
main "$@"
