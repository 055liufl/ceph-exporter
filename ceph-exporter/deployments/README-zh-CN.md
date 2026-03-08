# Docker Compose 配置文件说明

本目录包含 ceph-exporter 项目的所有 Docker Compose 配置文件及其详细中文注释版本。

## 📁 文件列表

### 原始配置文件

| 文件名 | 用途 | 适用场景 |
|--------|------|----------|
| `docker-compose.yml` | 标准监控栈配置 | 已有 Ceph 集群的生产环境 |
| `docker-compose-ceph-demo.yml` | Ceph Demo 独立部署 | 快速测试 Ceph 功能 |
| `docker-compose-integration-test.yml` | 集成测试环境 | CI/CD 自动化测试 |
| `docker-compose-lightweight-full.yml` | 轻量级完整栈 | Docker Toolbox / 资源受限环境 |

**时区配置**: 所有配置文件已自动配置宿主机时区挂载（`/etc/localtime` 和 `/etc/timezone`），确保容器时间与宿主机一致。

### 中文注释备份文件

| 文件名 | 大小 | 说明 |
|--------|------|------|
| `docker-compose.yml.zh-CN` | 31KB | 标准监控栈的详细中文注释版本 |
| `docker-compose-ceph-demo.yml.zh-CN` | 23KB | Ceph Demo 的详细中文注释版本 |
| `docker-compose-integration-test.yml.zh-CN` | 21KB | 集成测试的详细中文注释版本 |
| `docker-compose-lightweight-full.yml.zh-CN` | 38KB | 轻量级完整栈的详细中文注释版本 |

## 📖 中文注释文件特点

所有 `.zh-CN` 后缀的文件都包含以下详细注释:

### 1. 文件头部说明
- 📋 文件用途和目标
- 🔧 包含的组件列表
- 💻 资源需求说明
- 🎯 使用场景分析
- 📝 使用方法示例
- 🌐 访问地址列表
- ⚠️ 注意事项提醒

### 2. 配置项详细注释
- **每个配置项都有详细说明**
- **关键配置项有特别标注**
- **包含配置原理和影响**
- **提供最佳实践建议**
- **说明安全注意事项**

### 3. 实用参考信息
- 🚀 快速启动命令
- 🔍 故障排查步骤
- 📊 性能优化建议
- 💾 备份恢复方法
- 🧪 测试用例示例

## 🎯 使用建议

### 学习和理解
如果你想深入理解 Docker Compose 配置:
```bash
# 阅读中文注释版本
cat docker-compose-ceph-demo.yml.zh-CN
```

### 实际部署
实际部署时使用原始配置文件:
```bash
# 推荐：使用部署脚本（自动处理权限和配置）
./scripts/deploy.sh full

# 或手动使用原始配置文件
docker-compose -f docker-compose-ceph-demo.yml up -d
```

### 配置修改
修改配置时参考中文注释:
1. 打开 `.zh-CN` 文件查看详细说明
2. 理解配置项的作用和影响
3. 在原始文件中进行修改
4. 测试修改后的配置

---

## 🐛 常见部署问题

### 1. Prometheus 权限错误

**症状**: Prometheus 容器不断重启，日志显示 `permission denied`

**原因**: Prometheus 以 UID 65534 (nobody) 运行，数据目录权限不正确

**解决方案**:
```bash
# 使用部署脚本自动修复
sudo ./scripts/deploy.sh init

# 或手动修复
sudo chown -R 65534:65534 data/prometheus
docker-compose restart prometheus
```

### 2. Ceph-Exporter 配置文件找不到

**症状**: ceph-exporter 日志显示配置文件不存在

**原因**: deployments 目录下缺少 configs 软链接

**解决方案**:
```bash
cd deployments
ln -s ../configs configs
docker-compose restart ceph-exporter
```

### 3. Ceph-Demo 验证失败

**症状**: 验证脚本显示 ceph-demo 无法访问

**说明**: RGW 根路径返回 HTTP 404 是正常的，不是错误

**验证方法**:
```bash
# 检查容器状态
docker ps | grep ceph-demo

# 检查 Ceph 集群状态
docker exec ceph-demo ceph -s

# 测试 RGW 端口（返回 404 表示正常）
curl -v http://localhost:8080
```

### 4. 首次部署最佳实践

```bash
# 1. 进入部署目录
cd ceph-exporter/deployments

# 2. 使用部署脚本（推荐，自动处理所有配置）
sudo ./scripts/deploy.sh full

# 3. 等待服务启动
sleep 120

# 4. 验证部署
sudo ./scripts/deploy.sh verify
```

## 📚 配置文件详细说明

### 1. docker-compose.yml (标准监控栈)

**包含组件:**
- ceph-exporter: Ceph 指标导出器
- prometheus: 时序数据库
- grafana: 可视化仪表板
- alertmanager: 告警管理器

**资源需求:** 1-2GB 内存, 1-2 CPU

**适用场景:**
- ✅ 已有 Ceph 集群
- ✅ 生产环境监控
- ✅ 长期数据存储
- ❌ 不包含 Ceph 集群本身

**启动命令:**
```bash
docker-compose up -d
```

---

### 2. docker-compose-ceph-demo.yml (Ceph Demo)

**包含组件:**
- ceph-demo: 单节点 Ceph 集群 (All-in-One)

**资源需求:** 2GB 内存, 1-2 CPU

**适用场景:**
- ✅ 快速测试 Ceph 功能
- ✅ 学习 Ceph 架构
- ✅ 开发环境
- ❌ 不适合生产环境

**启动命令:**
```bash
docker-compose -f docker-compose-ceph-demo.yml up -d
```

**访问地址:**
- Ceph Dashboard: http://localhost:8080
- RGW API: http://localhost:5000

---

### 3. docker-compose-integration-test.yml (集成测试)

**包含组件:**
- ceph-demo: 测试用 Ceph 集群
- ceph-exporter: 被测试的导出器
- prometheus: 时序数据库
- grafana: 可视化平台

**资源需求:** 2-3GB 内存, 2 CPU

**适用场景:**
- ✅ 自动化集成测试
- ✅ CI/CD 流水线
- ✅ 功能验证
- ❌ 不适合生产环境

**启动命令:**
```bash
docker-compose -f docker-compose-integration-test.yml up -d
```

**测试完成后清理:**
```bash
docker-compose -f docker-compose-integration-test.yml down -v
```

---

### 4. docker-compose-lightweight-full.yml (轻量级完整栈) ⭐ 推荐

**包含组件:**
- **存储层:** ceph-demo
- **监控层:** ceph-exporter, prometheus, grafana, alertmanager
- **日志层:** elasticsearch, logstash, kibana
- **追踪层:** jaeger

**资源需求:** 4-6GB 内存, 2-4 CPU

**适用场景:**
- ✅ 开发和测试
- ✅ 完整功能演示
- ✅ 培训和学习
- ❌ 不适合生产环境

**启动命令:**
```bash
# 方法 1: 直接使用 docker-compose
docker-compose -f docker-compose-lightweight-full.yml up -d

# 方法 2: 使用部署脚本 (推荐)
./scripts/deploy.sh full
```

**访问地址:**
- Ceph Dashboard: http://localhost:8080
- Prometheus: http://localhost:9090
- Grafana: http://localhost:3000 (admin/admin)
- Kibana: http://localhost:5601
- Jaeger: http://localhost:16686

---

## 🔧 常用命令

### 启动服务
```bash
# 启动所有服务
docker-compose -f [配置文件] up -d

# 启动特定服务
docker-compose -f [配置文件] up -d [服务名]
```

### 查看状态
```bash
# 查看服务状态
docker-compose -f [配置文件] ps

# 查看资源使用
docker stats

# 查看服务日志
docker-compose -f [配置文件] logs -f [服务名]
```

### 停止和清理
```bash
# 停止所有服务
docker-compose -f [配置文件] down

# 停止并删除数据卷
docker-compose -f [配置文件] down -v

# 重启特定服务
docker-compose -f [配置文件] restart [服务名]
```

### 进入容器
```bash
# 进入容器 shell
docker exec -it [容器名] sh

# 执行单个命令
docker exec [容器名] [命令]
```

## 🐛 故障排查

### 1. 服务无法启动
```bash
# 查看容器日志
docker logs [容器名] --tail 100

# 查看容器状态
docker inspect [容器名]

# 检查端口占用
netstat -tlnp | grep [端口号]
```

### 2. 内存不足
```bash
# 查看内存使用
docker stats

# 检查是否有 OOM
docker inspect [容器名] | grep OOMKilled

# 解决方法:
# - 增加系统内存
# - 增加 Docker 内存限制
# - 减少运行的服务数量
```

### 3. 网络连接问题
```bash
# 测试容器间连接
docker exec [容器名] ping [目标容器名]

# 查看网络配置
docker network inspect [网络名]

# 检查端口监听
docker exec [容器名] netstat -tlnp
```

### 4. Elasticsearch 启动失败
```bash
# Linux 系统需要设置 vm.max_map_count
sudo sysctl -w vm.max_map_count=262144

# 永久生效
echo "vm.max_map_count=262144" | sudo tee -a /etc/sysctl.conf
```

## 📊 性能优化

### 1. 内存优化
- 根据实际需求调整各服务的 `mem_limit`
- 监控内存使用情况,避免 OOM
- 考虑使用 swap (但会影响性能)

### 2. 存储优化
- 定期清理旧数据
- 调整数据保留时间
- 使用 SSD 提高 I/O 性能

### 3. 网络优化
- 使用 host 网络模式 (如果安全)
- 减少不必要的端口映射
- 优化服务间通信

## 🔒 安全建议

### 1. 密码安全
- ⚠️ 修改默认密码 (admin/admin)
- ⚠️ 使用强密码
- ⚠️ 不要在代码仓库中硬编码密码

### 2. 网络安全
- 限制端口暴露范围
- 使用防火墙规则
- 启用 HTTPS (生产环境)

### 3. 访问控制
- 配置用户权限
- 启用认证和授权
- 定期审计访问日志

## 📞 获取帮助

如果遇到问题:
1. 查看对应的 `.zh-CN` 文件中的详细注释
2. 查看本 README 的故障排查章节
3. 查看 [时区配置说明](TIMEZONE_CONFIGURATION.md) 了解时区相关问题
4. 查看容器日志: `docker logs [容器名]`
5. 查看项目文档和 Issue

## 📝 更新日志

- **2026-03-09**:
  - 所有服务添加宿主机时区挂载配置
  - 自动挂载 `/etc/localtime` 和 `/etc/timezone`
  - 确保容器时间与宿主机保持一致
- **2026-03-08**:
  - 创建所有配置文件的详细中文注释版本
  - 添加了 4 个 `.zh-CN` 文件
  - 每个配置项都有详细说明
  - 包含使用场景和最佳实践
  - 提供故障排查和优化建议
  - 新增常见部署问题解决方案
  - 修复 Prometheus 权限问题
  - 修复 ceph-exporter 配置路径问题
  - 修复 ceph-demo 验证逻辑
- **2026-03-07**: 初始版本创建

---

**注意:** 中文注释文件 (`.zh-CN`) 仅用于学习和参考,实际部署请使用原始配置文件或部署脚本。
