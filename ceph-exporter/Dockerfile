# =============================================================================
# ceph-exporter Dockerfile
# =============================================================================
# 多阶段构建（启用 CGO 以支持 Ceph 库）:
#   1. builder 阶段: 使用 golang:alpine 编译，安装 Ceph 开发库
#   2. runtime 阶段: 使用 alpine 运行时镜像，安装 Ceph 运行时库
#
# 构建命令:
#   docker build -t ceph-exporter:dev .
#
# 运行命令:
#   docker run -d \
#     -p 9128:9128 \
#     -v /etc/ceph:/etc/ceph:ro \
#     ceph-exporter:dev
# =============================================================================

# ---------------------------------------------------------------------------
# 第一阶段: 编译（启用 CGO，安装 Ceph 开发库）
# ---------------------------------------------------------------------------
FROM golang:1.21-alpine AS builder

# 配置 Alpine 镜像源（国内加速）
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

# 配置 Go 代理（国内加速，多个备选）
# 可通过 --build-arg GOPROXY=xxx 覆盖
ARG GOPROXY=https://goproxy.cn,https://mirrors.aliyun.com/goproxy/,https://goproxy.io,direct
ENV GOPROXY=${GOPROXY}
ENV GO111MODULE=on
ENV GOSUMDB=off

# 安装编译依赖：gcc、musl-dev（C 编译器）、Ceph 开发库
RUN apk add --no-cache \
    gcc \
    musl-dev \
    ceph-dev \
    linux-headers

WORKDIR /app

# 先复制依赖文件，利用 Docker 层缓存
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 编译二进制文件
# CGO_ENABLED=1: 启用 CGO 以支持 Ceph C 库绑定
# -tags octopus: 匹配 Ceph 15.x (Octopus) 版本的构建标签
# -ldflags: 注入版本信息
ARG VERSION=dev
ARG BUILD_TIME=unknown
ARG GIT_COMMIT=unknown
RUN CGO_ENABLED=1 go build \
    -tags octopus \
    -ldflags "-X main.version=${VERSION} -X main.buildTime=${BUILD_TIME} -X main.gitCommit=${GIT_COMMIT}" \
    -o /app/build/ceph-exporter \
    ./cmd/ceph-exporter

# ---------------------------------------------------------------------------
# 第二阶段: 运行时（alpine 镜像 + Ceph 运行时库）
# ---------------------------------------------------------------------------
FROM alpine:3.19

# 配置 Alpine 镜像源（国内加速）
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

# 安装运行时依赖：Ceph 客户端库
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    ceph-common

# 创建非 root 用户运行服务
RUN addgroup -S ceph-exporter && adduser -S ceph-exporter -G ceph-exporter

# 创建必要的目录
RUN mkdir -p /etc/ceph-exporter /var/log/ceph-exporter \
    && chown -R ceph-exporter:ceph-exporter /var/log/ceph-exporter

# 从 builder 阶段复制二进制文件和配置
COPY --from=builder /app/build/ceph-exporter /usr/local/bin/ceph-exporter
COPY --from=builder /app/configs/ceph-exporter.yaml /etc/ceph-exporter/ceph-exporter.yaml

# 切换到非 root 用户
USER ceph-exporter

# 暴露 Prometheus 指标端口
EXPOSE 9128

# 健康检查
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget -qO- http://localhost:9128/health || exit 1

# 启动命令
ENTRYPOINT ["ceph-exporter"]
CMD ["-config", "/etc/ceph-exporter/ceph-exporter.yaml"]
