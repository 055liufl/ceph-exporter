# 部署故障排查指南

本文档列出了 ceph-exporter 部署过程中常见的问题及其解决方案。

---

## 📋 目录

- [服务启动问题](#服务启动问题)
- [权限问题](#权限问题)
- [配置问题](#配置问题)
- [网络问题](#网络问题)
- [资源问题](#资源问题)
- [验证问题](#验证问题)

---

## 服务启动问题

### 1. Prometheus 不断重启

**症状**:
```bash
$ docker ps
CONTAINER ID   NAME         STATUS
abc123         prometheus   Restarting
```

**日志信息**:
```
open /prometheus/queries.active: permission denied
panic: Unable to create mmap-ed active query log
```

**原因**: Prometheus 容器以 UID 65534 (nobody) 运行，但数据目录权限不正确

**解决方案**:

```bash
# 方法 1: 使用部署脚本自动修复（推荐）
cd deployments
sudo ./scripts/deploy.sh init

# 方法 2: 手动修复权限
sudo chown -R 65534:65534 data/prometheus
docker-compose restart prometheus

# 验证修复
docker logs prometheus --tail 20
```

---

### 2. Ceph-Exporter 连接失败

**症状**:
```bash
$ docker logs ceph-exporter
{"level":"error","message":"连接 Ceph 集群失败: rados: ret=-13, Permission denied"}
```

**原因**:
1. configs 目录软链接不存在
2. Ceph keyring 文件权限不正确
3. Ceph 集群尚未完全启动

**解决方案**:

```bash
# 1. 检查 configs 软链接
cd deployments
ls -la configs
# 如果不存在，创建软链接
ln -s ../configs configs

# 2. 等待 ceph-demo 完全启动（约 1-2 分钟）
docker logs ceph-demo --tail 50

# 3. 修复 keyring 权限
sudo chmod 644 data/ceph-demo/config/ceph.client.admin.keyring

# 4. 重启 ceph-exporter
docker-compose restart ceph-exporter

# 5. 验证连接
docker logs ceph-exporter --tail 20
curl http://localhost:9128/metrics
```

---

### 3. Grafana 无法启动

**症状**:
```bash
$ docker logs grafana
mkdir: cannot create directory '/var/lib/grafana/plugins': Permission denied
```

**原因**: Grafana 数据目录权限不正确（需要 UID 472）

**解决方案**:

```bash
# 修复权限
sudo chown -R 472:472 data/grafana
docker-compose restart grafana

# 验证
curl http://localhost:3000/api/health
```

---

### 4. Elasticsearch 启动失败

**症状**:
```bash
$ docker logs elasticsearch
max virtual memory areas vm.max_map_count [65530] is too low
```

**原因**: Linux 系统 vm.max_map_count 设置过低

**解决方案**:

```bash
# 临时设置
sudo sysctl -w vm.max_map_count=262144

# 永久生效
echo "vm.max_map_count=262144" | sudo tee -a /etc/sysctl.conf
sudo sysctl -p

# 重启 Elasticsearch
docker-compose restart elasticsearch
```

---

## 权限问题

### 权限要求总结

| 服务 | 运行用户 | UID | 数据目录权限 |
|------|---------|-----|-------------|
| Prometheus | nobody | 65534 | 65534:65534 |
| Grafana | grafana | 472 | 472:472 |
| Elasticsearch | elasticsearch | 1000 | 1000:1000 |
| Alertmanager | 当前用户 | $UID | $USER:$USER |

### 一键修复所有权限

```bash
cd deployments
sudo ./scripts/deploy.sh init
```

### 手动修复权限

```bash
cd deployments

# Prometheus
sudo chown -R 65534:65534 data/prometheus

# Grafana
sudo chown -R 472:472 data/grafana

# Elasticsearch
sudo chown -R 1000:1000 data/elasticsearch

# Alertmanager
sudo chown -R $USER:$USER data/alertmanager

# 重启所有服务
docker-compose restart
```

---

## 配置问题

### 1. 配置文件找不到

**症状**:
```bash
$ docker logs ceph-exporter
open /etc/ceph-exporter/ceph-exporter.yaml: no such file or directory
```

**原因**: docker-compose.yml 中引用了 `../configs/ceph-exporter.yaml`，但 configs 软链接不存在

**解决方案**:

```bash
cd deployments

# 检查软链接
ls -la configs

# 创建软链接
ln -s ../configs configs

# 验证
ls -la configs/ceph-exporter.yaml

# 重启服务
docker-compose restart ceph-exporter
```

---

### 2. Ceph 配置文件缺失

**症状**:
```bash
$ docker logs ceph-exporter
open /etc/ceph/ceph.conf: no such file or directory
```

**原因**: ceph-demo 尚未生成配置文件

**解决方案**:

```bash
# 1. 检查 ceph-demo 状态
docker ps | grep ceph-demo

# 2. 等待 ceph-demo 完全启动
docker logs ceph-demo --tail 50

# 3. 验证配置文件已生成
ls -la data/ceph-demo/config/

# 4. 重启 ceph-exporter
docker-compose restart ceph-exporter
```

---

## 网络问题

### 1. 端口被占用

**症状**:
```bash
Error starting userland proxy: listen tcp 0.0.0.0:9090: bind: address already in use
```

**解决方案**:

```bash
# 查找占用端口的进程
sudo netstat -tlnp | grep 9090
# 或
sudo ss -tlnp | grep 9090

# 停止占用端口的进程
sudo kill <PID>

# 或修改 docker-compose.yml 中的端口映射
# 例如: "9091:9090" 改为使用 9091 端口
```

---

### 2. 容器间无法通信

**症状**: ceph-exporter 无法连接到 ceph-demo

**解决方案**:

```bash
# 检查网络配置
docker network ls
docker network inspect deployments_ceph-network

# 测试容器间连接
docker exec ceph-exporter ping ceph-demo

# 重建网络
docker-compose down
docker-compose up -d
```

---

## 资源问题

### 1. 内存不足

**症状**:
```bash
$ docker inspect <container> | grep OOMKilled
"OOMKilled": true
```

**解决方案**:

```bash
# 查看内存使用
docker stats

# 检查系统可用内存
free -h

# 解决方法:
# 1. 增加系统内存
# 2. 减少运行的服务数量
# 3. 调整 docker-compose.yml 中的 mem_limit
```

---

### 2. 磁盘空间不足

**症状**:
```bash
no space left on device
```

**解决方案**:

```bash
# 检查磁盘使用
df -h
du -sh data/*

# 清理 Docker 资源
docker system prune -a

# 清理旧数据
./scripts/deploy.sh clean

# 调整数据保留策略
# 编辑 prometheus.yml 中的 retention.time
```

---

## 验证问题

### 1. Ceph-Demo 验证失败

**症状**:
```bash
$ ./scripts/deploy.sh verify
✗ ceph-demo 无法访问或尚未就绪
```

**说明**: 这可能不是错误！RGW 根路径返回 HTTP 404 是正常行为

**验证方法**:

```bash
# 1. 检查容器状态
docker ps | grep ceph-demo
# 应该显示 "Up X minutes"

# 2. 检查 Ceph 集群状态
docker exec ceph-demo ceph -s
# 应该显示 "HEALTH_OK"

# 3. 测试 RGW 端口（返回 404 表示正常）
curl -v http://localhost:8080
# 应该返回 HTTP/1.1 404 Not Found

# 4. 检查端口监听
docker exec ceph-demo netstat -tlnp | grep 8080
# 应该显示 radosgw 进程在监听
```

**如果使用旧版本脚本**: 更新到最新版本的 deploy.sh，已修复验证逻辑

---

### 2. 服务健康检查失败

**症状**: 某个服务验证失败

**诊断步骤**:

```bash
# 1. 检查容器状态
docker ps -a

# 2. 查看容器日志
docker logs <container-name> --tail 100

# 3. 测试服务端点
curl -v http://localhost:<port>

# 4. 检查端口监听
docker exec <container-name> netstat -tlnp

# 5. 检查资源使用
docker stats <container-name>
```

---

## 完整诊断流程

如果遇到部署问题，按以下顺序排查：

### 1. 检查环境

```bash
# 检查 Docker
docker --version
sudo systemctl status docker

# 检查 Docker Compose
docker-compose --version

# 检查系统资源
free -h
df -h
```

### 2. 检查容器状态

```bash
# 查看所有容器
docker ps -a

# 查看失败容器的日志
docker logs <container-name> --tail 100
```

### 3. 检查权限

```bash
# 检查数据目录权限
ls -la data/

# 修复权限
sudo ./scripts/deploy.sh init
```

### 4. 检查配置

```bash
# 检查 configs 软链接
ls -la configs

# 检查 Ceph 配置
ls -la data/ceph-demo/config/
```

### 5. 重启服务

```bash
# 重启单个服务
docker-compose restart <service-name>

# 重启所有服务
docker-compose restart

# 完全重新部署
docker-compose down
sudo ./scripts/deploy.sh full
```

---

## 获取帮助

如果以上方法都无法解决问题：

1. **收集诊断信息**:
```bash
# 运行统一诊断脚本
cd /home/lfl/ceph-exporter/ceph-exporter/deployments
sudo ./scripts/diagnose.sh > diagnosis.log 2>&1

# 或通过 deploy.sh
sudo ./scripts/deploy.sh diagnose > diagnosis.log 2>&1

# 或手动收集
docker ps -a > docker-ps.log
docker logs prometheus > prometheus.log 2>&1
docker logs ceph-exporter > ceph-exporter.log 2>&1
docker logs ceph-demo > ceph-demo.log 2>&1
```

2. **查看文档**:
   - [部署目录 README](README.md)
   - [Docker Compose 配置说明](README-zh-CN.md)
   - [数据存储说明](DATA_STORAGE.md)
   - [诊断脚本整合说明](docs/DIAGNOSE_INTEGRATION.md)

3. **提交 Issue**: 附上诊断日志和错误信息

---

## 预防措施

### 首次部署最佳实践

```bash
# 1. 使用部署脚本（推荐）
cd deployments
sudo ./scripts/deploy.sh full

# 2. 等待服务完全启动
sleep 120

# 3. 验证部署
sudo ./scripts/deploy.sh verify

# 4. 检查所有服务状态
docker ps
docker-compose ps
```

### 定期维护

```bash
# 检查磁盘使用
du -sh data/*

# 清理旧日志
docker-compose logs --tail 0

# 备份数据
tar -czf backup-$(date +%Y%m%d).tar.gz data/

# 更新镜像
docker-compose pull
docker-compose up -d
```

---

**最后更新**: 2026-03-08
