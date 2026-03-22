# 数据存储说明

本项目使用绑定挂载（bind mount）方式存储数据，所有数据存储在 `./data/` 目录下。

---

## 📂 数据目录结构

```
deployments/
└── data/
    ├── ceph-demo/
    │   ├── data/              # Ceph 集群数据（OSD、Mon 数据）
    │   └── config/            # Ceph 配置文件（ceph.conf、密钥环）
    ├── prometheus/            # Prometheus 时序数据库
    ├── grafana/               # Grafana 仪表板和配置
    ├── alertmanager/          # Alertmanager 告警状态
    ├── elasticsearch/         # Elasticsearch 索引数据
    └── test/                  # 集成测试数据
        ├── ceph-demo/
        │   ├── data/
        │   └── config/
        ├── prometheus/
        └── grafana/
```

---

## 🎯 优势

使用绑定挂载的优势：

1. **可见性**: 数据存储在项目目录下，易于查看和管理
2. **备份方便**: 直接备份 `data/` 目录即可
3. **权限控制**: 可以使用文件系统权限控制访问
4. **迁移简单**: 复制 `data/` 目录即可迁移数据
5. **调试友好**: 可以直接查看和修改数据文件

---

## 📍 数据存储位置

| 服务 | 数据路径 | 用途 |
|------|---------|------|
| Ceph Demo | `./data/ceph-demo/data/` | Ceph 集群数据 |
| Ceph Demo | `./data/ceph-demo/config/` | Ceph 配置文件 |
| Prometheus | `./data/prometheus/` | 时序数据库 |
| Grafana | `./data/grafana/` | 仪表板和配置 |
| Alertmanager | `./data/alertmanager/` | 告警状态 |
| Elasticsearch | `./data/elasticsearch/` | 索引数据 |

---

## 🔧 初始化数据目录

首次启动前，需要创建数据目录并设置权限：

```bash
# 创建数据目录
mkdir -p data/{ceph-demo/{data,config},prometheus,grafana,alertmanager,elasticsearch}
mkdir -p data/test/{ceph-demo/{data,config},prometheus,grafana}

# 设置权限（根据需要调整）
# Grafana 需要 472 用户权限
sudo chown -R 472:472 data/grafana data/test/grafana

# Elasticsearch 需要 1000 用户权限
sudo chown -R 1000:1000 data/elasticsearch

# Prometheus 需要 nobody 用户权限 (UID 65534)
sudo chown -R 65534:65534 data/prometheus data/alertmanager
sudo chown -R 65534:65534 data/test/prometheus
```

或使用部署脚本自动初始化：

```bash
./scripts/deploy.sh init
```

---

## 💾 备份和恢复

### 备份数据

```bash
# 备份所有数据
tar -czf ceph-exporter-data-$(date +%Y%m%d).tar.gz data/

# 备份特定服务数据
tar -czf prometheus-data-$(date +%Y%m%d).tar.gz data/prometheus/
tar -czf grafana-data-$(date +%Y%m%d).tar.gz data/grafana/
```

### 恢复数据

```bash
# 停止服务
docker compose down

# 恢复数据
tar -xzf ceph-exporter-data-20260308.tar.gz

# 启动服务
docker compose up -d
```

---

## 🧹 清理数据

### 清理所有数据

```bash
# 停止并删除容器
docker compose down

# 删除数据目录
rm -rf data/
```

### 清理特定服务数据

```bash
# 停止服务
docker compose stop prometheus

# 删除数据
rm -rf data/prometheus/*

# 重启服务
docker compose start prometheus
```

---

## 📊 查看数据占用

```bash
# 查看总体占用
du -sh data/

# 查看各服务占用
du -sh data/*

# 详细查看
du -h --max-depth=2 data/
```

---

## ⚠️ 注意事项

1. **权限问题**: 某些容器需要特定的用户权限，启动前需要正确设置
2. **磁盘空间**: 确保有足够的磁盘空间，特别是 Prometheus 和 Elasticsearch
3. **数据持久化**: `data/` 目录已添加到 `.gitignore`，不会提交到 Git
4. **SELinux**: 如果启用了 SELinux，可能需要设置 SELinux 上下文：
   ```bash
   sudo chcon -Rt svirt_sandbox_file_t data/
   ```

---

## 🔗 相关文档

- [完整操作指南](../../Ceph-Exporter项目完整操作指南.md)
- [Docker Compose 配置](./README.md)
- [故障排查](./TROUBLESHOOTING.md)

---

**最后更新**: 2026-03-15
