# 快速开始

5 分钟快速部署 ceph-exporter。

---

## 前提条件

- CentOS 7.x 系统
- 4GB 内存、2 核 CPU
- 30GB 磁盘空间
- root 或 sudo 权限

---

## 步骤 1: 安装 Docker

```bash
sudo yum install -y yum-utils
sudo yum-config-manager --add-repo https://mirrors.aliyun.com/docker-ce/linux/centos/docker-ce.repo
sudo yum install -y docker-ce docker-ce-cli containerd.io
sudo systemctl start docker
sudo systemctl enable docker
```

## 步骤 2: 安装 Docker Compose

```bash
sudo curl -L "https://get.daocloud.io/docker/compose/releases/download/v2.24.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose
```

## 步骤 3: 配置镜像加速

```bash
sudo mkdir -p /etc/docker
sudo tee /etc/docker/daemon.json <<EOF
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
sudo systemctl restart docker
```

## 步骤 4: 部署服务

```bash
cd ceph-exporter/deployments
chmod +x scripts/deploy.sh

# 自动部署（会自动初始化数据目录）
./scripts/deploy.sh full
```

**注意**: 部署脚本会自动创建 `./data/` 目录用于存储所有服务数据，包括：
- Ceph 集群数据和配置
- Prometheus 时序数据
- Grafana 仪表板
- Alertmanager 告警状态
- Elasticsearch 索引数据

数据存储在项目目录下，方便备份和管理。详见 [数据存储说明](ceph-exporter/deployments/DATA_STORAGE.md)。

**时区配置**: 所有容器已自动挂载宿主机时区（`/etc/localtime` 和 `/etc/timezone`），确保容器时间与宿主机一致。

## 步骤 5: 验证部署

```bash
# 查看状态
./scripts/deploy.sh status

# 访问服务
curl http://localhost:9128/metrics
curl http://localhost:9090
curl http://localhost:3000
```

---

## 服务访问

### 基础监控服务

| 服务 | 地址 | 凭据 |
|------|------|------|
| Ceph Exporter | http://localhost:9128/metrics | - |
| Prometheus | http://localhost:9090 | - |
| Grafana | http://localhost:3000 | admin/admin |
| Alertmanager | http://localhost:9093 | - |

### 完整监控栈 (full 模式)

| 服务 | 地址 | 凭据 | 说明 |
|------|------|------|------|
| Ceph Dashboard | http://localhost:8080 | - | Ceph 管理界面 |
| Elasticsearch | http://localhost:9200 | - | 日志存储 |
| Kibana | http://localhost:5601 | - | 日志查询和可视化 |
| Jaeger UI | http://localhost:16686 | - | 分布式追踪界面 |

---

## 常用命令

```bash
# 查看状态
./scripts/deploy.sh status

# 查看日志
./scripts/deploy.sh logs [service-name]

# 验证部署
./scripts/deploy.sh verify

# 诊断问题
./scripts/deploy.sh diagnose

# 修复部署问题
sudo ./scripts/deploy.sh fix

# 停止服务
./scripts/deploy.sh stop

# 清理数据（会删除 ./data/ 目录）
./scripts/deploy.sh clean

# 备份数据
tar -czf backup-$(date +%Y%m%d).tar.gz data/
```

---

## 常见问题

### 防火墙阻止访问

```bash
sudo systemctl stop firewalld
```

### SELinux 权限问题

```bash
sudo setenforce 0
```

### 内存不足

```bash
# 使用最小部署
./scripts/deploy.sh minimal
```

### 镜像拉取失败

```bash
# 检查镜像加速器
docker info | grep -A 5 "Registry Mirrors"
```

---

## 更多文档

- **README.md** - 项目主文档
- **DEPLOYMENT_GUIDE.md** - 完整部署指南
- **DOCKER_MIRROR_CONFIGURATION.md** - 镜像配置

---

**版本**: 2.0
**最后更新**: 2026-03-15
