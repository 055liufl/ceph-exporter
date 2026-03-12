# Jaeger 分布式追踪 - 状态更新

## ✅ 实现完成！

**Phase 3 分布式追踪功能已完成！** 🎉

之前 Jaeger 中没有追踪数据是因为只有 Phase 1 占位实现。现在已经实现了完整的 OpenTelemetry + Jaeger 集成。

## 当前状态

### Phase 1（已完成）✅
- ✅ 配置结构定义
- ✅ 追踪模块框架
- ✅ 空操作（no-op）函数

### Phase 2（已完成）✅
- ✅ 日志系统完整实现
- ✅ ELK 集成成功
- ✅ 可以在 Kibana 中查看日志

### Phase 3（已完成）✅
- ✅ OpenTelemetry SDK 集成
- ✅ Jaeger OTLP HTTP Exporter
- ✅ HTTP 请求自动追踪
- ✅ Trace ID 和 Span ID 生成
- ✅ 追踪属性记录
- ✅ Docker Compose 集成

## 快速开始

### 1. 启用追踪功能

运行快速启用脚本：

```bash
cd deployments
./scripts/enable-jaeger-tracing.sh
```

或者手动启用：

编辑 `configs/ceph-exporter.yaml`:

```yaml
tracer:
  enabled: true                     # 启用追踪
  jaeger_url: "jaeger:4318"
  service_name: "ceph-exporter"
  sample_rate: 1.0
```

### 2. 重启服务

```bash
cd deployments
docker-compose -f docker-compose-lightweight-full.yml restart ceph-exporter
```

### 3. 生成追踪数据

```bash
# 发送请求
curl http://localhost:9128/metrics
```

### 4. 查看追踪数据

访问 Jaeger UI: http://localhost:16686

1. Service 下拉框选择: `ceph-exporter`
2. 点击 `Find Traces` 按钮
3. 查看追踪详情

### 5. 运行测试

```bash
cd deployments
./scripts/test-jaeger-tracing.sh
```

## 实现详情

查看完整实现文档：

- **实现文档**: `docs/JAEGER-TRACING-IMPLEMENTATION.md`
- **测试脚本**: `deployments/scripts/test-jaeger-tracing.sh`
- **启用脚本**: `deployments/scripts/enable-jaeger-tracing.sh`

## 追踪数据示例

启用后，Jaeger UI 中将显示：

```
Service: ceph-exporter
  └─ Trace: 3f2a1b4c5d6e7f8g
      └─ Span: /metrics (Duration: 125ms)
          ├─ http.method: GET
          ├─ http.url: http://localhost:9128/metrics
          ├─ http.host: localhost:9128
          └─ http.user_agent: curl/7.29.0
```

## 技术栈

- **OpenTelemetry SDK**: v1.24.0
- **OTLP HTTP Exporter**: v1.24.0
- **Jaeger**: v1.35 (支持 OTLP)
- **协议**: OTLP HTTP (端口 4318)

## 故障排查

如果 Jaeger 中仍然没有追踪数据：

1. ✅ 检查追踪是否启用
   ```bash
   grep "enabled:" configs/ceph-exporter.yaml
   ```

2. ✅ 检查 Jaeger 是否运行
   ```bash
   docker ps | grep jaeger
   ```

3. ✅ 检查 ceph-exporter 日志
   ```bash
   docker logs ceph-exporter | grep tracer
   ```

4. ✅ 运行测试脚本
   ```bash
   ./scripts/test-jaeger-tracing.sh
   ```

## 总结

**从 Phase 1 占位实现到 Phase 3 完整实现，分布式追踪功能现已可用！**

- ✅ 完整的 OpenTelemetry 集成
- ✅ Jaeger OTLP HTTP 支持
- ✅ 自动 HTTP 请求追踪
- ✅ 配置管理和测试工具

**现在可以在 Jaeger UI 中看到完整的追踪数据了！** 🎉

## 当前实现状态

### Phase 1（当前状态）- 占位实现

✅ **已完成**：
- 配置结构定义（`internal/config/config.go`）
- 追踪模块框架（`internal/tracer/tracer.go`）
- 空操作（no-op）的追踪函数
- 配置文件结构（`configs/ceph-exporter.yaml`）

❌ **未实现**：
- OpenTelemetry SDK 集成
- Jaeger Exporter 配置
- 实际的追踪数据生成
- HTTP 请求追踪
- Ceph 命令执行追踪
- Trace Context 传播

### 当前配置

```yaml
tracer:
  enabled: false                                    # 默认禁用
  jaeger_url: "http://jaeger:14268/api/traces"     # Jaeger Collector URL
  service_name: "ceph-exporter"                     # 服务名称
  sample_rate: 1.0                                  # 采样率
```

### 当前代码实现

```go
// Phase 1 占位实现 - 所有函数都是 no-op
func StartSpan(ctx context.Context, name string) (context.Context, *Span) {
    // 返回原始上下文和空操作 Span
    return ctx, &Span{name: name}
}

func (s *Span) End() {
    // no-op - 不产生任何追踪数据
}
```

## Phase 3 计划实现内容

### 1. OpenTelemetry SDK 集成

```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/jaeger"
    "go.opentelemetry.io/otel/sdk/trace"
)

func NewTracerProvider(cfg *config.TracerConfig, log *logger.Logger) (*TracerProvider, error) {
    // 创建 Jaeger Exporter
    exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(
        jaeger.WithEndpoint(cfg.JaegerURL),
    ))

    // 创建 TracerProvider
    tp := trace.NewTracerProvider(
        trace.WithBatcher(exporter),
        trace.WithSampler(trace.TraceIDRatioBased(cfg.SampleRate)),
        trace.WithResource(resource.NewWithAttributes(
            semconv.ServiceNameKey.String(cfg.ServiceName),
        )),
    )

    otel.SetTracerProvider(tp)
    return &TracerProvider{tp: tp}, nil
}
```

### 2. HTTP 请求追踪

```go
// HTTP Handler 中间件
func TracingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ctx, span := otel.Tracer("http").Start(r.Context(), r.URL.Path)
        defer span.End()

        span.SetAttributes(
            attribute.String("http.method", r.Method),
            attribute.String("http.url", r.URL.String()),
        )

        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

### 3. Ceph 命令追踪

```go
func (c *CephClient) ExecuteCommand(ctx context.Context, cmd string) error {
    ctx, span := otel.Tracer("ceph").Start(ctx, "ceph.command")
    defer span.End()

    span.SetAttributes(
        attribute.String("ceph.command", cmd),
        attribute.String("ceph.cluster", c.cluster),
    )

    // 执行 Ceph 命令
    err := c.execute(cmd)
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
    }

    return err
}
```

### 4. 追踪数据示例

实现后，Jaeger 中将显示：

```
Service: ceph-exporter
  └─ Trace: GET /metrics
      ├─ Span: http.request (100ms)
      │   ├─ http.method: GET
      │   ├─ http.url: /metrics
      │   └─ http.status_code: 200
      │
      ├─ Span: ceph.health (50ms)
      │   ├─ ceph.command: ceph health
      │   └─ ceph.cluster: ceph
      │
      └─ Span: ceph.df (30ms)
          ├─ ceph.command: ceph df
          └─ ceph.cluster: ceph
```

## 为什么 Phase 1 只做占位实现？

1. **分阶段开发策略**
   - Phase 1: 核心功能（指标采集）
   - Phase 2: 日志系统（已完成 ✅）
   - Phase 3: 分布式追踪（计划中）

2. **依赖管理**
   - OpenTelemetry SDK 会增加依赖和二进制大小
   - 先确保核心功能稳定

3. **可选功能**
   - 追踪是可选的高级功能
   - 不是所有用户都需要

## 如何启用追踪功能（Phase 3 实现后）

### 1. 修改配置

```yaml
tracer:
  enabled: true                                     # 启用追踪
  jaeger_url: "http://jaeger:14268/api/traces"     # Jaeger Collector URL
  service_name: "ceph-exporter"                     # 服务名称
  sample_rate: 1.0                                  # 100% 采样
```

### 2. 确保 Jaeger 运行

```bash
# 检查 Jaeger 状态
docker ps | grep jaeger

# 访问 Jaeger UI
http://localhost:16686
```

### 3. 重启 ceph-exporter

```bash
docker-compose restart ceph-exporter
```

### 4. 生成追踪数据

```bash
# 访问 metrics 端点生成追踪
curl http://localhost:9128/metrics
```

### 5. 在 Jaeger UI 中查看

- Service: 选择 `ceph-exporter`
- Operation: 选择 `GET /metrics`
- 点击 "Find Traces"

## 当前解决方案

### 临时方案：使用日志关联

虽然没有分布式追踪，但可以使用日志中的 `trace_id` 和 `span_id` 字段来关联请求：

```json
{
  "level": "info",
  "message": "处理 metrics 请求",
  "trace_id": "abc123",
  "span_id": "def456",
  "component": "http"
}
```

在 Kibana 中搜索相同的 `trace_id` 可以看到同一个请求的所有日志。

### 长期方案：实现 Phase 3

如果需要完整的分布式追踪功能，需要：

1. **添加 OpenTelemetry 依赖**
   ```bash
   go get go.opentelemetry.io/otel
   go get go.opentelemetry.io/otel/exporters/jaeger
   go get go.opentelemetry.io/otel/sdk/trace
   ```

2. **实现真正的追踪功能**
   - 替换 `internal/tracer/tracer.go` 中的占位实现
   - 添加 HTTP 中间件
   - 在关键路径添加 Span

3. **测试和验证**
   - 确保追踪数据正确发送到 Jaeger
   - 验证 Span 关系正确

## 总结

**当前状态**：
- ❌ Jaeger 中没有追踪数据是正常的
- ✅ 追踪模块只是 Phase 1 占位实现
- ✅ 配置结构已就绪，等待 Phase 3 实现

**替代方案**：
- ✅ 使用 ELK 日志系统（已完成）
- ✅ 通过日志中的 trace_id 关联请求
- ✅ 使用 Prometheus 指标监控

**未来计划**：
- 📅 Phase 3 将实现完整的 OpenTelemetry + Jaeger 集成
- 📅 届时可以在 Jaeger UI 中看到完整的追踪数据

## 参考文档

- OpenTelemetry Go SDK: https://opentelemetry.io/docs/instrumentation/go/
- Jaeger Documentation: https://www.jaegertracing.io/docs/
- 当前实现: `internal/tracer/tracer.go`
- 配置文件: `configs/ceph-exporter.yaml`
