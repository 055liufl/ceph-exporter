# 部署指南

**环境**: CentOS 7 + Docker

---

## 📋 环境要求

| 项目 | 要求 |
|------|------|
| 操作系统 | CentOS 7.x |
| Docker | 19.03+ |
| Docker Compose | 1.25+ |
| 内存 | 4GB（推荐 8GB） |
| CPU | 2 核（推荐 4 核） |
| 磁盘 | 30GB |

---

## 🔧 环境准备

### 1. 安装 Docker

```bash
sudo yum install -y yum-utils
sudo yum-config-manager --add-repo https://mirrors.aliyun.com/docker-ce/linux/centos/docker-ce.repo
sudo yum install -y docker-ce docker-ce-cli containerd.io
sudo systemctl start docker
sudo systemctl enable docker
```

### 2. 安装 Docker Compose

```bash
sudo curl -L "https://get.daocloud.io/docker/compose/releases/download/v2.24.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose
```

### 3. 配置镜像加速（国内必需）

```bash
sudo mkdir -p /etc/docker
sudo tee /etc/docker/daemon.json <<EOF
{
  "registry-mirrors": [
    "https://docker.mirrors.ustc.edu.cn",
    "https://hub-mirror.c.163.com"
  ]
}
EOF
sudo systemctl restart docker
```

### 4. 配置防火墙（可选）

```bash
# 开放端口
sudo firewall-cmd --permanent --add-port=9128/tcp
sudo firewall-cmd --permanent --add-port=9090/tcp
sudo firewall-cmd --permanent --add-port=3000/tcp
sudo firewall-cmd --reload

# 或临时关闭
sudo systemctl stop firewalld
```

---

## 🚀 部署步骤

### 方式 1: 自动部署（推荐）

```bash
cd /home/lfl/ceph-exporter/ceph-exporter/deployments
chmod +x scripts/deploy.sh
./scripts/deploy.sh full
```

### 方式 2: 手动部署

```bash
cd /home/lfl/ceph-exporter/ceph-exporter/deployments

# 集成测试环境
docker-compose -f docker-compose-integration-test.yml up -d

# 或完整监控栈
docker-compose -f docker-compose-lightweight-full.yml up -d

# 或最小监控栈
docker-compose up -d
```

---

## 🔍 验证部署

### 检查服务状态

```bash
docker ps
```

### 访问服务

- ceph-exporter: http://localhost:9128/metrics
- Prometheus: http://localhost:9090
- Grafana: http://localhost:3000 (admin/admin)

### 健康检查

```bash
curl http://localhost:9128/health
curl http://localhost:9090/-/healthy
curl http://localhost:3000/api/health
```

---

## 🐛 故障排查

### 1. 容器无法启动

```bash
# 查看日志
docker logs <container-name>

# 检查资源
docker stats
df -h

# 清理资源
docker system prune -a
```

### 2. 内存不足

```bash
# 检查内存
free -h

# 使用最小部署
./scripts/deploy.sh minimal
```

### 3. 镜像拉取失败

```bash
# 检查镜像加速器
docker info | grep -A 5 "Registry Mirrors"

# 手动拉取镜像
docker pull ceph/demo:latest-nautilus
docker pull prom/prometheus:latest
docker pull grafana/grafana:latest
```

### 4. 端口冲突

```bash
# 检查端口占用
sudo netstat -tulpn | grep 9128
sudo netstat -tulpn | grep 9090
sudo netstat -tulpn | grep 3000

# 查找占用进程
sudo lsof -i :9128
```

### 5. SELinux 问题

```bash
# 临时禁用
sudo setenforce 0

# 永久禁用（需重启）
sudo sed -i 's/^SELINUX=enforcing/SELINUX=disabled/' /etc/selinux/config
```

---

## 🛑 停止和清理

### 停止服务

```bash
cd /home/lfl/ceph-exporter/ceph-exporter/deployments

# 停止服务（保留数据）
docker-compose down

# 停止并删除数据
docker-compose down -v
```

### 完全清理

```bash
# 清理所有资源
docker system prune -a
docker volume prune
docker network prune
```

---

## 📚 相关文档

- **README.md** - 项目主文档
- **QUICK_START.md** - 快速开始
- **DOCKER_MIRROR_CONFIGURATION.md** - 镜像配置
- **ceph-exporter/README.md** - 详细架构文档

---

**版本**: 2.0
**最后更新**: 2026-03-07
