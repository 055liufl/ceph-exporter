# Docker 镜像加速配置指南

**适用环境**: CentOS 7 + Docker

本文档介绍如何在 CentOS 7 环境下配置 Docker 镜像加速器，提高镜像拉取速度。

---

## 📋 推荐镜像源

国内可用的 Docker 镜像加速器（按推荐顺序）：

1. **中科大镜像** - https://docker.mirrors.ustc.edu.cn
2. **网易镜像** - https://hub-mirror.c.163.com
3. **腾讯云镜像** - https://mirror.ccs.tencentyun.com
4. **阿里云镜像** - https://registry.cn-hangzhou.aliyuncs.com

---

## 🔧 配置方法

### 方法 1: 使用部署脚本自动配置（推荐）

```bash
cd ceph-exporter/deployments
./scripts/deploy.sh mirror
```

脚本会自动配置镜像加速器并重启 Docker 服务。

### 方法 2: 手动配置

#### 步骤 1: 创建或编辑 Docker 配置文件

```bash
sudo mkdir -p /etc/docker
sudo vi /etc/docker/daemon.json
```

#### 步骤 2: 添加镜像源配置

```json
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
```

#### 步骤 3: 重启 Docker 服务

```bash
sudo systemctl daemon-reload
sudo systemctl restart docker
```

---

## ✅ 验证配置

### 查看已配置的镜像源

```bash
docker info | grep -A 5 "Registry Mirrors"
```

**预期输出**:

```
Registry Mirrors:
 https://docker.mirrors.ustc.edu.cn/
 https://hub-mirror.c.163.com/
 https://mirror.ccs.tencentyun.com/
```

### 测试镜像拉取速度

```bash
# 清除本地镜像缓存
docker rmi alpine:latest 2>/dev/null || true

# 测试拉取速度
time docker pull alpine:latest
```

---

## 🚀 常用镜像预拉取

配置完成后，可以预拉取项目所需的镜像：

```bash
# 拉取 Ceph Demo 镜像
docker pull ceph/demo:latest-nautilus

# 拉取监控组件镜像
docker pull prom/prometheus:latest
docker pull grafana/grafana:latest
docker pull prom/alertmanager:latest

# 拉取 ELK 镜像
docker pull docker.elastic.co/elasticsearch/elasticsearch:7.17.0
docker pull docker.elastic.co/logstash/logstash:7.17.0
docker pull docker.elastic.co/kibana/kibana:7.17.0

# 拉取 Jaeger 镜像
docker pull jaegertracing/all-in-one:1.35
```

或使用部署脚本自动拉取：

```bash
cd ceph-exporter/deployments
./scripts/deploy.sh pull
```

---

## 🔍 故障排查

### 问题 1: 配置后仍然很慢

**解决方案**:

1. 检查网络连接
2. 尝试更换镜像源顺序
3. 测试各个镜像源的连通性

```bash
# 测试镜像源连通性
curl -I https://docker.mirrors.ustc.edu.cn
curl -I https://hub-mirror.c.163.com
curl -I https://mirror.ccs.tencentyun.com
```

### 问题 2: Docker 服务重启失败

**解决方案**:

```bash
# 检查配置文件语法
cat /etc/docker/daemon.json | python -m json.tool

# 查看 Docker 服务日志
sudo journalctl -u docker -n 50

# 如果配置有误，删除配置文件重新配置
sudo rm /etc/docker/daemon.json
sudo systemctl restart docker
```

### 问题 3: 镜像源不可用

**解决方案**:

如果某个镜像源不可用，从配置中移除该源：

```bash
sudo vi /etc/docker/daemon.json
# 删除不可用的镜像源
sudo systemctl restart docker
```

---

## 📝 配置说明

### registry-mirrors

指定 Docker 镜像加速器地址，Docker 会按顺序尝试从这些镜像源拉取镜像。

### log-driver 和 log-opts

配置 Docker 日志驱动和日志轮转策略，防止日志文件过大：

- `max-size`: 单个日志文件最大大小
- `max-file`: 保留的日志文件数量

### storage-driver

指定 Docker 存储驱动，`overlay2` 是推荐的存储驱动，性能更好。

---

## 💡 最佳实践

1. **配置多个镜像源**: 提高可用性，一个源不可用时自动切换
2. **定期测试**: 镜像源可能会变更，定期测试确保可用
3. **日志管理**: 配置日志轮转，防止磁盘空间耗尽
4. **使用 overlay2**: 确保使用推荐的存储驱动

---

## 🔗 相关文档

- [完整部署指南](./DEPLOYMENT_GUIDE.md)
- [快速开始](./QUICK_START.md)

---

**文档版本**: 2.0
**最后更新**: 2026-03-15
**维护者**: ceph-exporter 项目团队
