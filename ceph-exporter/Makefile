# =============================================================================
# ceph-exporter Makefile
# =============================================================================
# 提供项目的构建、测试、Docker 镜像构建等常用操作
#
# 常用目标:
#   make build      - 编译二进制文件
#   make test       - 运行单元测试
#   make lint       - 代码静态检查
#   make docker     - 构建 Docker 镜像
#   make clean      - 清理构建产物
#   make help       - 显示帮助信息
# =============================================================================

# 项目信息
APP_NAME    := ceph-exporter
VERSION     := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME  := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Go 编译参数
# CGO_ENABLED=1: 启用 CGO 以支持 Ceph C 库绑定
# BUILD_TAGS: Ceph 版本构建标签（octopus = Ceph 15.x）
GO          := go
GOFLAGS     := -v
CGO_ENABLED := 1
BUILD_TAGS  := octopus
LDFLAGS     := -ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.gitCommit=$(GIT_COMMIT)"

# 目录
CMD_DIR     := ./cmd/ceph-exporter
BUILD_DIR   := ./build
CONFIGS_DIR := ./configs

# Docker 参数
DOCKER_IMAGE := ceph-exporter
DOCKER_TAG   := $(VERSION)

# =============================================================================
# 默认目标
# =============================================================================
.DEFAULT_GOAL := help

# =============================================================================
# 构建目标
# =============================================================================

## build: 编译 ceph-exporter 二进制文件
.PHONY: build
build:
	@echo ">>> 编译 $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=$(CGO_ENABLED) $(GO) build $(GOFLAGS) -tags $(BUILD_TAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME) $(CMD_DIR)
	@echo ">>> 编译完成: $(BUILD_DIR)/$(APP_NAME)"

## build-linux: 交叉编译 Linux amd64 二进制文件
.PHONY: build-linux
build-linux:
	@echo ">>> 交叉编译 Linux amd64..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=$(CGO_ENABLED) GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) -tags $(BUILD_TAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 $(CMD_DIR)
	@echo ">>> 编译完成: $(BUILD_DIR)/$(APP_NAME)-linux-amd64"

# =============================================================================
# 测试目标
# =============================================================================

## test: 运行所有单元测试
.PHONY: test
test:
	@echo ">>> 运行单元测试..."
	CGO_ENABLED=$(CGO_ENABLED) $(GO) test -tags $(BUILD_TAGS) -v -count=1 ./internal/...
	@echo ">>> 测试完成"

## test-cover: 运行测试并生成覆盖率报告
.PHONY: test-cover
test-cover:
	@echo ">>> 运行测试（含覆盖率）..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=$(CGO_ENABLED) $(GO) test -tags $(BUILD_TAGS) -v -coverprofile=$(BUILD_DIR)/coverage.out -covermode=atomic ./internal/...
	$(GO) tool cover -html=$(BUILD_DIR)/coverage.out -o $(BUILD_DIR)/coverage.html
	@echo ">>> 覆盖率报告: $(BUILD_DIR)/coverage.html"

## test-short: 运行快速测试（跳过集成测试）
.PHONY: test-short
test-short:
	@echo ">>> 运行快速测试..."
	CGO_ENABLED=$(CGO_ENABLED) $(GO) test -tags $(BUILD_TAGS) -v -short -count=1 ./internal/...

## test-integration: 运行集成测试
.PHONY: test-integration
test-integration:
	@echo ">>> 运行集成测试..."
	@echo ">>> 注意: 集成测试需要 Docker 和 Docker Compose"
	cd test/integration && $(GO) test -tags $(BUILD_TAGS) -v -timeout 30m

## test-all: 运行所有测试（单元测试 + 集成测试）
.PHONY: test-all
test-all: test test-integration
	@echo ">>> 所有测试完成"

# =============================================================================
# 代码质量
# =============================================================================

## lint: 运行代码静态检查
.PHONY: lint
lint:
	@echo ">>> 运行代码检查..."
	$(GO) vet -tags $(BUILD_TAGS) ./...
	@echo ">>> 代码检查完成"

## fmt: 格式化代码
.PHONY: fmt
fmt:
	@echo ">>> 格式化代码..."
	$(GO) fmt ./...
	@echo ">>> 格式化完成"

## tidy: 整理 Go 模块依赖
.PHONY: tidy
tidy:
	@echo ">>> 整理依赖..."
	$(GO) mod tidy
	@echo ">>> 依赖整理完成"

## pre-commit: 安装 pre-commit 并运行所有 hooks
.PHONY: pre-commit
pre-commit:
	@echo ">>> 运行 pre-commit..."
	pre-commit run --all-files
	@echo ">>> pre-commit 完成"

## pre-commit-install: 安装 pre-commit git hooks
.PHONY: pre-commit-install
pre-commit-install:
	@echo ">>> 安装 pre-commit hooks..."
	pre-commit install
	@echo ">>> 运行 pre-commit install --install-hooks 以安装 hook 环境..."
	pre-commit install --install-hooks
	@echo ">>> pre-commit 安装完成"

# =============================================================================
# Docker 目标
# =============================================================================

## docker: 构建 Docker 镜像
.PHONY: docker
docker:
	@echo ">>> 构建 Docker 镜像 $(DOCKER_IMAGE):$(DOCKER_TAG)..."
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	docker tag $(DOCKER_IMAGE):$(DOCKER_TAG) $(DOCKER_IMAGE):latest
	@echo ">>> Docker 镜像构建完成"

## docker-push: 推送 Docker 镜像到仓库
.PHONY: docker-push
docker-push: docker
	@echo ">>> 推送 Docker 镜像..."
	docker push $(DOCKER_IMAGE):$(DOCKER_TAG)
	docker push $(DOCKER_IMAGE):latest

# =============================================================================
# 运行目标
# =============================================================================

## run: 本地运行 ceph-exporter
.PHONY: run
run: build
	@echo ">>> 启动 $(APP_NAME)..."
	$(BUILD_DIR)/$(APP_NAME) -config $(CONFIGS_DIR)/ceph-exporter.yaml

# =============================================================================
# 清理目标
# =============================================================================

## clean: 清理构建产物
.PHONY: clean
clean:
	@echo ">>> 清理构建产物..."
	rm -rf $(BUILD_DIR)
	@echo ">>> 清理完成"

# =============================================================================
# 帮助
# =============================================================================

## help: 显示帮助信息
.PHONY: help
help:
	@echo "ceph-exporter Makefile"
	@echo ""
	@echo "用法: make [目标]"
	@echo ""
	@echo "可用目标:"
	@sed -n 's/^## //p' $(MAKEFILE_LIST) | column -t -s ':' | sed 's/^/  /'
