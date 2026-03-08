#!/bin/bash
set -e

echo "=== 安装 Go 环境 ==="

# 检查是否有 sudo 权限
if sudo -n true 2>/dev/null; then
    echo "检测到 sudo 权限,安装到 /usr/local"
    sudo tar -C /usr/local -xzf /tmp/go1.21.6.linux-amd64.tar.gz
    GO_ROOT="/usr/local/go"
else
    echo "无 sudo 权限,安装到用户目录"
    mkdir -p ~/go-sdk
    tar -C ~/go-sdk -xzf /tmp/go1.21.6.linux-amd64.tar.gz
    GO_ROOT="$HOME/go-sdk/go"
fi

# 配置环境变量
echo "配置环境变量..."
cat >> ~/.bashrc << EOF

# Go 环境配置
export PATH=\$PATH:$GO_ROOT/bin
export GOPATH=\$HOME/go
export PATH=\$PATH:\$GOPATH/bin
EOF

# 立即生效
export PATH=$PATH:$GO_ROOT/bin
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin

# 验证安装
echo "验证 Go 安装..."
go version

# 配置 Go 代理 (加速下载)
go env -w GOPROXY=https://goproxy.cn,direct

# 安装 Go 工具
echo "安装 goimports..."
go install golang.org/x/tools/cmd/goimports@latest

echo "安装 golangci-lint..."
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

# 验证工具安装
echo "验证工具安装..."
which go gofmt goimports golangci-lint
go version
goimports -h > /dev/null && echo "✓ goimports 安装成功"
golangci-lint version

echo "=== Go 环境安装完成 ==="
echo "请运行: source ~/.bashrc"
echo "然后执行: pre-commit run --all-files"
