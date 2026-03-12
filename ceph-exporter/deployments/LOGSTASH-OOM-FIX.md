# Logstash 内存不足问题修复

## 问题描述

Logstash 启动时出现 `java.lang.OutOfMemoryError: Java heap space` 错误，导致服务无法正常运行。

## 原因分析

原配置中 Logstash 的 JVM 堆内存设置过小：
- JVM 堆内存: `-Xms128m -Xmx128m` (128MB)
- 容器内存限制: `256m`

这对于 Logstash 7.17.0 来说远远不够，特别是在处理日志解析和转发时。

## 已修复内容

### 1. 增加 JVM 堆内存

**文件**: `deployments/docker-compose-lightweight-full.yml`

**修改前**:
```yaml
environment:
  - "LS_JAVA_OPTS=-Xms128m -Xmx128m"
mem_limit: 256m
```

**修改后**:
```yaml
environment:
  - "LS_JAVA_OPTS=-Xms512m -Xmx512m"
mem_limit: 768m
```

### 2. 添加 Logstash 配置文件挂载

```yaml
volumes:
  - ./logstash/logstash.conf:/usr/share/logstash/pipeline/logstash.conf:ro
  - /etc/localtime:/etc/localtime:ro
  - /etc/timezone:/etc/timezone:ro
```

### 3. 更新 Logstash 配置

**文件**: `deployments/logstash/logstash.conf`

更新为支持 ceph-exporter 日志的完整配置：
- 支持 Filebeat 输入（方案2）
- 支持 TCP 直接推送（方案1）
- JSON 日志解析
- 字段提取和清理
- 输出到 Elasticsearch

## 应用修复

**重要**: Logstash 服务定义在 `docker-compose-lightweight-full.yml` 文件中，不是默认的 `docker-compose.yml`。

### 方法1: 重启 Logstash 服务（推荐）

```bash
cd /home/lfl/ceph-exporter/ceph-exporter/deployments
sudo ./scripts/fix-logstash-oom.sh
```

### 方法2: 手动重新创建服务

```bash
cd /home/lfl/ceph-exporter/ceph-exporter/deployments

# 停止并删除 Logstash 容器
sudo docker-compose -f docker-compose-lightweight-full.yml stop logstash
sudo docker-compose -f docker-compose-lightweight-full.yml rm -f logstash

# 重新创建并启动
sudo docker-compose -f docker-compose-lightweight-full.yml up -d logstash
```

### 方法3: 使用部署脚本

```bash
cd /home/lfl/ceph-exporter/ceph-exporter/deployments
sudo ./scripts/deploy.sh -f docker-compose-lightweight-full.yml restart logstash
```

## 验证修复

### 1. 查看 Logstash 日志

```bash
sudo docker-compose -f docker-compose-lightweight-full.yml logs -f logstash

# 或使用部署脚本
sudo ./scripts/deploy.sh -f docker-compose-lightweight-full.yml logs logstash
```

应该看到类似以下的成功启动日志：
```
[INFO ][logstash.javapipeline][main] Starting pipeline
[INFO ][logstash.agent] Successfully started Logstash API endpoint
```

### 2. 检查 Logstash 状态

```bash
# 检查容器状态
sudo docker ps | grep logstash

# 检查 Logstash API
curl http://localhost:9600/_node/stats/pipelines
```

### 3. 测试日志接收

```bash
# 测试 TCP 输入（方案1）
echo '{"message":"test","level":"info"}' | nc localhost 5000

# 查看 Elasticsearch 中的日志
curl http://localhost:9200/ceph-exporter-*/_search?pretty
```

## 资源要求

修复后的 Logstash 资源要求：

| 资源 | 最小值 | 推荐值 |
|------|--------|--------|
| 内存 | 768MB | 1GB |
| CPU | 0.5 核 | 1 核 |
| 磁盘 | 100MB | 500MB |

## 性能优化建议

### 1. 根据负载调整内存

如果日志量很大，可以进一步增加内存：

```yaml
environment:
  - "LS_JAVA_OPTS=-Xms1g -Xmx1g"
mem_limit: 1536m
```

### 2. 调整 Worker 数量

在 `logstash.conf` 中添加：

```ruby
# 在 input 部分之前添加
pipeline {
  workers => 2
  batch.size => 125
}
```

### 3. 禁用调试输出

生产环境中注释掉 stdout 输出：

```ruby
output {
  elasticsearch {
    hosts => ["elasticsearch:9200"]
    index => "ceph-exporter-%{+YYYY.MM.dd}"
  }

  # 生产环境注释掉
  # stdout {
  #   codec => rubydebug
  # }
}
```

## 故障排查

### 问题1: 仍然出现 OOM

**解决方案**: 进一步增加内存

```yaml
environment:
  - "LS_JAVA_OPTS=-Xms1g -Xmx1g"
mem_limit: 1536m
```

### 问题2: Logstash 启动慢

**原因**: JVM 初始化和插件加载需要时间

**解决方案**: 等待 1-2 分钟，或查看日志确认启动进度

### 问题3: 配置文件语法错误

**验证配置**:
```bash
sudo docker exec logstash /usr/share/logstash/bin/logstash --config.test_and_exit -f /usr/share/logstash/pipeline/logstash.conf
```

## 相关文档

- [ELK 日志集成指南](../../docs/ELK-LOGGING-GUIDE.md)
- [Logstash 配置示例](../../configs/logstash.conf)
- [日志切换脚本](./scripts/README-switch-logging.md)

## 总结

本次修复将 Logstash 的内存从 128MB 增加到 512MB，容器限制从 256MB 增加到 768MB，解决了内存不足导致的启动失败问题。同时更新了 Logstash 配置以支持 ceph-exporter 的日志格式。

修复后，Logstash 应该能够正常启动并处理日志。
