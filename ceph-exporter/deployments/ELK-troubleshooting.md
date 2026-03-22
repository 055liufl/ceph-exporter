# ELK 系统 Kibana 无日志来源 — 排查与解决记录

## 问题描述

在 ELK（Elasticsearch + Logstash + Kibana）系统中，Kibana 的 Discover 页面没有任何日志数据来源，无法查看日志。

**环境信息：**

| 组件 | 版本 | 容器名 |
|------|------|--------|
| Elasticsearch | 7.17.0 | elasticsearch |
| Logstash | 7.17.0 | logstash |
| Kibana | 7.17.0 | kibana |
| Filebeat | 7.17.0 | filebeat-sidecar |
| Ceph Exporter | dev | ceph-exporter |

**部署方式：** Docker Compose（`docker-compose-lightweight-full.yml`）

---

## 排查过程

### 第一步：检查 Elasticsearch 集群状态

```bash
curl -X GET "localhost:9200/_cluster/health?pretty"
```

**结果：** 集群状态为 `green`，运行正常。

```bash
curl -X GET "localhost:9200/_cat/indices?v"
```

**结果：** 没有 `logstash-*`、`filebeat-*` 或 `ceph-exporter-*` 索引，说明没有日志数据进入 ES。

---

### 第二步：检查各容器运行状态

```bash
docker ps -a | grep -E 'logstash|filebeat|elastic|kibana'
```

**发现问题 1：** `filebeat-sidecar` 容器状态为 `Restarting`，不断崩溃重启。

---

### 第三步：查看 Filebeat 崩溃原因

```bash
docker logs filebeat-sidecar --tail 50
```

**错误信息：**

```
Exiting: error loading config file: config file ("filebeat.yml") must be owned by the user identifier (uid=0) or root
```

**原因：** Filebeat 要求配置文件必须由 root 用户拥有（安全机制），但宿主机上的 `filebeat.yml` 不属于 root。

**解决方案：**

```bash
# 修改文件所有者为 root
sudo chown root:root filebeat/filebeat.yml

# 设置权限为 644
sudo chmod 644 filebeat/filebeat.yml

# 重启 Filebeat
sudo docker restart filebeat-sidecar
```

**结果：** Filebeat 成功启动，不再崩溃。

---

### 第四步：Filebeat 启动但仍无数据

Filebeat 启动后，监控指标显示：

```
harvester:  {"open_files": 0, "running": 0}   ← 没有在读取任何文件
events:     {"active": 0}                      ← 没有事件产生
registrar:  {"states": {"current": 0}}         ← 没有追踪到任何文件
```

进一步检查容器内的日志目录：

```bash
sudo docker exec filebeat-sidecar ls /var/lib/docker/containers/
```

**发现问题 2：** 容器内 `/var/lib/docker/containers/` 目录为空！

**根本原因：** Docker 的数据目录不在默认的 `/var/lib/docker/`，而是被自定义到了 `/home/docker/`。

验证：

```bash
sudo ls /home/docker/
# 输出: buildkit containers engine-id image network overlay2 plugins runtimes swarm tmp volumes
```

docker-compose 中的挂载配置指向了错误的宿主机路径：

```yaml
# 错误 ❌
- /var/lib/docker/containers:/var/lib/docker/containers:ro

# 正确 ✅
- /home/docker/containers:/var/lib/docker/containers:ro
```

**解决方案：**

修改 `docker-compose-lightweight-full.yml` 第 177 行：

```yaml
# 修改前
volumes:
  - ./filebeat/filebeat.yml:/usr/share/filebeat/filebeat.yml:ro
  - /var/lib/docker/containers:/var/lib/docker/containers:ro   # ← 路径错误

# 修改后
volumes:
  - ./filebeat/filebeat.yml:/usr/share/filebeat/filebeat.yml:ro
  - /home/docker/containers:/var/lib/docker/containers:ro      # ← 修正为实际路径
```

重新创建容器：

```bash
cd ceph-exporter/deployments
sudo docker compose -f docker-compose-lightweight-full.yml up -d filebeat-sidecar
```

---

## 修复后验证

```bash
# 1. 确认 Filebeat 容器内能看到日志文件
sudo docker exec filebeat-sidecar ls /var/lib/docker/containers/ | head -5

# 2. 确认 Filebeat 正常运行
sudo docker ps | grep filebeat
sudo docker logs filebeat-sidecar --tail 20

# 3. 等待 30 秒后检查 ES 是否有日志索引
curl -X GET "localhost:9200/_cat/indices/ceph-exporter-*?v"

# 4. 在 Kibana 中创建 Index Pattern
#    Management → Stack Management → Index Patterns → Create index pattern
#    输入: ceph-exporter-*
#    时间字段选择: @timestamp
```

---

## 问题总结

| 序号 | 问题 | 原因 | 解决方案 |
|------|------|------|----------|
| 1 | Filebeat 容器不断重启 | `filebeat.yml` 文件所有者不是 root | `chown root:root` + `chmod 644` |
| 2 | Filebeat 启动但无数据 | Docker 数据目录为 `/home/docker/`，但挂载路径写的是 `/var/lib/docker/` | 修改 docker-compose 中的挂载源路径 |

---

## 数据链路参考

```
ceph-exporter (容器日志)
    ↓
/home/docker/containers/<container-id>/<container-id>-json.log
    ↓  (bind mount)
Filebeat (采集 + 过滤 container.name == "ceph-exporter")
    ↓  (port 5044, Beats 协议)
Logstash (解析 + 处理)
    ↓  (port 9200)
Elasticsearch (索引: ceph-exporter-YYYY.MM.dd)
    ↓
Kibana (Index Pattern: ceph-exporter-*)
```

---

## 相关文件路径

| 文件 | 路径 |
|------|------|
| Docker Compose | `docker-compose-lightweight-full.yml` |
| Filebeat 配置 | `filebeat/filebeat.yml` |
| Logstash 管道配置 | 容器内 `/usr/share/logstash/pipeline/logstash.conf` |
| Docker 数据目录 | `/home/docker/` （非默认路径） |
