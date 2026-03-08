# 集成测试文档

## 概述

本目录包含 ceph-exporter 项目的集成测试套件，用于验证容器化部署的完整性和正确性。

**重要更新 (2026-03-02)**: 集成测试已更新为使用 `docker-compose-integration-test.yml`，包含 Ceph Demo 容器，提供完整的测试环境。

## 测试内容

### 1. Docker 容器测试 (`docker_test.go`)

- 验证所有容器能正常启动
- 检查容器运行状态
- 使用 `docker-compose-integration-test.yml` 配置
- 包含 Ceph Demo 容器

### 2. 网络通信测试 (`network_test.go`)

- 测试容器间网络通信
- 验证所有服务的健康检查端点
- 验证服务间的 DNS 解析
- 支持 CentOS 7 + Docker 环境

### 3. 服务功能测试 (`services_test.go`)

- **数据持久化测试**: 验证 Prometheus、Grafana 的数据卷
- **Prometheus 集成测试**: 验证 Prometheus 能采集 ceph-exporter 指标
- **Grafana 集成测试**: 验证 Grafana 数据源和 Dashboard 配置
- **容器健康检查测试**: 验证所有容器的健康状态
- **Web UI 访问测试**: 验证所有 Web 界面可访问
- **指标采集测试**: 验证指标格式和采集功能

### 4. 资源约束测试 (`resource_test.go`)

- 验证所有容器都设置了内存限制
- 检查容器实际内存使用情况
- 验证资源约束配置正确性

## 前置条件

1. 安装 Docker 和 Docker Compose
2. 确保以下端口未被占用:
   - 8080 (Ceph Demo Dashboard)
   - 9128 (ceph-exporter)
   - 9090 (Prometheus)
   - 3000 (Grafana)

3. **系统资源要求**:
   - 内存: 至少 2-3GB 可用
   - CPU: 至少 2 核
   - 磁盘: 至少 5GB 可用空间

### CentOS 7 环境说明

本项目运行在 CentOS 7 + Docker 环境中：

1. **访问服务**: 使用 localhost 访问所有服务
   - Prometheus: http://localhost:9090
   - Grafana: http://localhost:3000
   - Alertmanager: http://localhost:9093
   - ceph-exporter: http://localhost:9128/metrics

2. **防火墙配置**: 如需远程访问，请开放相应端口
   ```bash
   sudo firewall-cmd --permanent --add-port=9128/tcp
   sudo firewall-cmd --permanent --add-port=9090/tcp
   sudo firewall-cmd --permanent --add-port=3000/tcp
   sudo firewall-cmd --reload
   ```

## 运行测试

### 运行所有集成测试

```bash
# 从项目根目录运行
cd test/integration
go test -v -timeout 30m
```

### 运行特定测试

```bash
# 只运行网络通信测试
go test -v -run TestContainerNetworkCommunication

# 只运行数据持久化测试
go test -v -run TestDataPersistence

# 只运行 Prometheus 集成测试
go test -v -run TestPrometheusTargets
```

### 跳过集成测试

```bash
# 使用 -short 标志跳过集成测试
go test -short
```

## 测试流程

1. **环境准备** (`TestMain`)
   - 清理旧的测试环境
   - 启动 docker-compose 服务
   - 等待服务完全启动（45秒）

2. **执行测试**
   - 按顺序执行各个测试用例
   - 每个测试都有重试机制和超时控制

3. **环境清理** (`TestMain`)
   - 停止所有容器
   - 清理测试数据（可选）

## 测试超时设置

- 单个测试超时: 10分钟
- 整体测试超时: 30分钟
- HTTP 请求超时: 10秒
- 服务启动等待: 45秒

## 故障排查

### 容器启动失败

```bash
# 查看容器日志
docker-compose -f ../../deployments/docker-compose.yml logs

# 查看特定服务日志
docker-compose -f ../../deployments/docker-compose.yml logs ceph-exporter
```

### 网络连接失败

```bash
# 检查容器网络
docker network ls
docker network inspect ceph-monitor-net

# 检查端口占用
netstat -an | grep 9128
netstat -an | grep 9090
netstat -an | grep 3000
netstat -an | grep 9093
```

### 健康检查失败

```bash
# 检查容器健康状态
docker ps
docker inspect --format='{{.State.Health.Status}}' ceph-exporter

# 手动测试健康检查端点
curl http://localhost:9128/health
curl http://localhost:9090/-/healthy
curl http://localhost:3000/api/health
curl http://localhost:9093/-/healthy
```

### 数据持久化问题

```bash
# 查看数据目录
ls -lh ../../deployments/data/

# 查看数据占用
du -sh ../../deployments/data/*

# 清理数据（谨慎使用）
cd ../../deployments
./scripts/deploy.sh clean
```

## 测试覆盖的验收标准

根据 `docs/ceph-exporter-plan.md` Phase 4 的验收标准：

- ✅ 所有容器能正常启动
- ✅ 容器间网络通信正常
- ✅ 数据持久化正常
- ✅ 能通过 Web UI 访问所有服务
- ✅ 集成测试全部通过

## CI/CD 集成

### GitHub Actions 示例

```yaml
name: Integration Tests

on: [push, pull_request]

jobs:
  integration-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - name: Run integration tests
        run: |
          cd test/integration
          go test -v -timeout 30m
```

### GitLab CI 示例

```yaml
integration-test:
  stage: test
  image: golang:1.21
  services:
    - docker:dind
  script:
    - cd test/integration
    - go test -v -timeout 30m
```

## 注意事项

1. **测试环境隔离**: 每次测试都会清理旧环境，确保测试的独立性
2. **资源消耗**: 集成测试会启动多个容器，需要足够的系统资源
3. **测试时间**: 完整的集成测试需要 5-10 分钟
4. **并发限制**: 不建议并发运行多个集成测试实例
5. **数据清理**: 测试结束后会自动清理容器，但不会删除数据卷（除非使用 `-v` 参数）

## 扩展测试

如果需要添加新的集成测试：

1. 在 `test/integration` 目录创建新的测试文件
2. 测试函数命名遵循 `Test*` 格式
3. 使用 `testing.Short()` 支持跳过集成测试
4. 添加适当的重试机制和超时控制
5. 更新本 README 文档

## 参考资料

- [Docker Compose 文档](https://docs.docker.com/compose/)
- [Go 测试文档](https://golang.org/pkg/testing/)
- [Prometheus API 文档](https://prometheus.io/docs/prometheus/latest/querying/api/)
- [Grafana API 文档](https://grafana.com/docs/grafana/latest/http_api/)
