# ceph-exporter 日志推送到 ELK 配置指南

本文档说明如何配置 ceph-exporter 将日志推送到 ELK（Elasticsearch + Logstash + Kibana）。

## 方案概览

### 方案 1: 直接推送到 Logstash (Direct Push)

**架构:**
```
ceph-exporter --> Logstash --> Elasticsearch --> Kibana
```

**优点:**
- 实时推送，延迟低
- 配置简单，无需额外组件
- 支持 TCP（可靠）和 UDP（快速）两种模式

**缺点:**
- 需要网络连接
- Logstash 故障会影响日志推送（但不影响应用运行）
- 日志缓冲区满时可能丢失日志

**适用场景:**
- 小规模部署
- 对日志实时性要求高
- 网络稳定的环境

### 方案 2: 容器日志收集 (Container Log Collection) - **推荐**

**架构:**
```
ceph-exporter (stdout) --> Filebeat --> Logstash --> Elasticsearch --> Kibana
```

**优点:**
- 解耦，Logstash 故障不影响应用
- Filebeat 提供缓冲和重试机制
- 符合云原生最佳实践
- 可以统一采集多个容器的日志

**缺点:**
- 需要部署 Filebeat
- 配置相对复杂

**适用场景:**
- 生产环境（推荐）
- Kubernetes/Docker 容器化部署
- 多服务统一日志收集

---

## 快速开始

### 使用 deploy.sh 部署（推荐）

`deploy.sh full` 部署时会交互式选择日志方案，也支持通过环境变量指定：

```bash
cd ceph-exporter/deployments

# 交互式选择日志方案
./scripts/deploy.sh full

# 或通过环境变量指定
LOGGING_MODE=container ./scripts/deploy.sh full     # 容器日志收集（推荐）
LOGGING_MODE=direct ./scripts/deploy.sh full        # 直接推送 TCP
LOGGING_MODE=direct-udp ./scripts/deploy.sh full    # 直接推送 UDP
LOGGING_MODE=file ./scripts/deploy.sh full          # 文件日志
LOGGING_MODE=dev ./scripts/deploy.sh full           # 开发模式
```

`container` 模式会自动启动 `filebeat-sidecar` 服务，其他模式不启动。

### 运行中切换日志方案

```bash
# 切换到方案2（容器日志收集）- 自动启动 filebeat-sidecar
./deployments/scripts/switch-logging-mode.sh container

# 切换到方案1（直接推送 TCP）- 自动停止 filebeat-sidecar
./deployments/scripts/switch-logging-mode.sh direct

# 切换到方案1（直接推送 UDP）- 自动停止 filebeat-sidecar
./deployments/scripts/switch-logging-mode.sh direct-udp

# 切换到文件日志模式
./deployments/scripts/switch-logging-mode.sh file

# 切换到开发模式
./deployments/scripts/switch-logging-mode.sh dev

# 查看当前配置
./deployments/scripts/switch-logging-mode.sh show
```

### 方案 1: 直接推送到 Logstash

#### 1. 修改配置文件

编辑 `configs/ceph-exporter.yaml`:

```yaml
logger:
  level: "info"
  format: "json"
  output: "stdout"              # 同时输出到 stdout（可选）
  enable_elk: true              # 启用 ELK 集成
  logstash_url: "logstash:5044" # Logstash 地址
  logstash_protocol: "tcp"      # tcp（可靠）或 udp（快速）
  service_name: "ceph-exporter"
```

#### 2. 启动服务

```bash
# 使用 Docker Compose 启动（包含 ELK Stack）
docker-compose -f configs/docker-compose-elk.yaml up -d ceph-exporter-direct logstash elasticsearch kibana

# 或直接运行
./ceph-exporter -config configs/ceph-exporter.yaml
```

#### 3. 验证日志

```bash
# 查看 Logstash 是否接收到日志
curl http://localhost:9600/_node/stats/pipelines

# 在 Kibana 中查看日志
# 访问 http://localhost:5601
# 创建索引模式: ceph-exporter-*
```

---

### 方案 2: 容器日志收集（推荐）

#### 1. 修改配置文件

编辑 `configs/ceph-exporter.yaml`:

```yaml
logger:
  level: "info"
  format: "json"
  output: "stdout"              # 输出到标准输出
  enable_elk: false             # 不直接推送
  service_name: "ceph-exporter"
```

#### 2. 配置 Filebeat

使用提供的 `configs/filebeat.yml` 配置文件，或根据需要修改。

#### 3. 启动服务

```bash
# 使用 Docker Compose 启动（包含 Filebeat sidecar）
docker-compose -f configs/docker-compose-elk.yaml up -d ceph-exporter-sidecar filebeat-sidecar logstash elasticsearch kibana
```

#### 4. Kubernetes 部署示例

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: ceph-exporter
spec:
  containers:
  # 主容器
  - name: ceph-exporter
    image: ceph-exporter:latest
    ports:
    - containerPort: 9128
    volumeMounts:
    - name: config
      mountPath: /app/configs
    - name: ceph-config
      mountPath: /etc/ceph
      readOnly: true

  # Filebeat sidecar
  - name: filebeat
    image: elastic/filebeat:8.11.0
    volumeMounts:
    - name: filebeat-config
      mountPath: /usr/share/filebeat/filebeat.yml
      subPath: filebeat.yml
    - name: varlog
      mountPath: /var/log/pods
      readOnly: true

  volumes:
  - name: config
    configMap:
      name: ceph-exporter-config
  - name: ceph-config
    hostPath:
      path: /etc/ceph
  - name: filebeat-config
    configMap:
      name: filebeat-config
  - name: varlog
    hostPath:
      path: /var/log/pods
```

---

## 配置切换

### 使用脚本切换（推荐）

脚本会自动修改配置文件并管理 `filebeat-sidecar` 服务：

```bash
# 切换到方案1（TCP）- 自动停止 filebeat-sidecar
./deployments/scripts/switch-logging-mode.sh direct

# 切换到方案1（UDP）- 自动停止 filebeat-sidecar
./deployments/scripts/switch-logging-mode.sh direct-udp

# 切换到方案2（推荐）- 自动启动 filebeat-sidecar
./deployments/scripts/switch-logging-mode.sh container

# 切换后重启 ceph-exporter 以应用配置
docker-compose -f docker-compose-lightweight-full.yml restart ceph-exporter
```

### 手动切换：从方案 2 切换到方案 1

```bash
# 1. 修改配置
sed -i 's/enable_elk: false/enable_elk: true/' configs/ceph-exporter.yaml

# 2. 重启服务
docker-compose -f docker-compose-lightweight-full.yml restart ceph-exporter

# 3. 停止 Filebeat（可选）
docker-compose -f docker-compose-lightweight-full.yml stop filebeat-sidecar
```

### 手动切换：从方案 1 切换到方案 2

```bash
# 1. 修改配置
sed -i 's/enable_elk: true/enable_elk: false/' configs/ceph-exporter.yaml

# 2. 启动 Filebeat
docker-compose -f docker-compose-lightweight-full.yml up -d filebeat-sidecar

# 3. 重启服务
docker-compose -f docker-compose-lightweight-full.yml restart ceph-exporter
```

---

## 配置参数说明

### logger 配置项

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `level` | string | `info` | 日志级别: trace, debug, info, warn, error, fatal, panic |
| `format` | string | `json` | 日志格式: json（结构化）, text（文本） |
| `output` | string | `stdout` | 输出目标: stdout, stderr, file |
| `file_path` | string | - | 日志文件路径（output=file 时生效） |
| `enable_elk` | bool | `false` | 是否启用直接推送到 Logstash |
| `logstash_url` | string | - | Logstash 地址（host:port） |
| `logstash_protocol` | string | `tcp` | 协议: tcp（可靠）, udp（快速） |
| `service_name` | string | `ceph-exporter` | 服务名称，用于在 ELK 中标识 |

---

## 故障排查

### 方案 1 故障排查

#### 日志未推送到 Logstash

1. 检查配置:
```bash
grep -A 5 "enable_elk" configs/ceph-exporter.yaml
```

2. 检查网络连接:
```bash
telnet logstash 5044
```

3. 查看 ceph-exporter 日志:
```bash
docker logs ceph-exporter | grep -i elk
# 应该看到: "ELK 集成已启用，日志将推送到 tcp://logstash:5044"
```

4. 查看 Logstash 日志:
```bash
docker logs logstash | grep -i error
```

#### 日志推送缓慢

- TCP 模式下，如果 Logstash 响应慢，可能导致缓冲区满
- 解决方案: 切换到 UDP 模式或增加缓冲区大小

### 方案 2 故障排查

#### Filebeat 未采集到日志

1. 检查 Filebeat 状态:
```bash
docker logs filebeat-sidecar
```

2. 检查日志路径:
```bash
# 容器日志路径
ls -la /var/lib/docker/containers/

# 文件日志路径
ls -la /var/log/ceph-exporter/
```

3. 测试 Filebeat 配置:
```bash
docker exec filebeat-sidecar filebeat test config
docker exec filebeat-sidecar filebeat test output
```

---

## 性能优化

### 方案 1 优化

1. **使用 UDP 协议**（如果可以接受少量日志丢失）:
```yaml
logstash_protocol: "udp"
```

2. **调整日志级别**（减少日志量）:
```yaml
level: "warn"  # 只记录警告和错误
```

### 方案 2 优化

1. **Filebeat 批量发送**:
```yaml
output.logstash:
  bulk_max_size: 2048
  compression_level: 3
```

2. **日志采样**（高流量场景）:
```yaml
processors:
  - drop_event:
      when:
        regexp:
          message: "^DEBUG"
```

---

## 监控指标

### 方案 1 监控

ceph-exporter 内部会记录日志推送的统计信息（通过日志输出）:
- 推送成功数
- 推送失败数
- 缓冲区使用率

### 方案 2 监控

Filebeat 提供监控 API:
```bash
curl http://localhost:5066/stats
```

---

## 最佳实践

1. **生产环境推荐方案 2**（容器日志收集）
   - 更稳定，解耦
   - 符合云原生架构

2. **开发环境使用 stdout + text 格式**
   ```yaml
   output: "stdout"
   format: "text"
   enable_elk: false
   ```

3. **启用日志压缩**（节省存储空间）
   ```yaml
   compress: true
   ```

4. **设置合理的日志保留策略**
   ```yaml
   max_age: 7  # 保留 7 天
   ```

5. **使用 JSON 格式**（便于 ELK 解析）
   ```yaml
   format: "json"
   ```

---

## 参考资料

- [Logstash 官方文档](https://www.elastic.co/guide/en/logstash/current/index.html)
- [Filebeat 官方文档](https://www.elastic.co/guide/en/beats/filebeat/current/index.html)
- [ELK Stack 最佳实践](https://www.elastic.co/guide/en/elastic-stack/current/index.html)
