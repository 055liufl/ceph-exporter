# Prometheus 状态页面中文化指南

> **问题**: Prometheus 自带的 `/stats` 页面全部显示为英文
> **解决方案**: 在 Grafana 中创建完全中文的 Prometheus 状态监控仪表板

---

## 问题分析

### 当前情况

你看到的页面是 **Prometheus 内置的状态页面**（访问地址：`http://localhost:9090/stats`），这个页面的特点是：

1. **硬编码的英文界面**：所有文本都是硬编码在 Prometheus 的 Go 源代码中
2. **不支持国际化**：Prometheus 官方不提供多语言支持
3. **无法通过配置修改**：即使修改 Prometheus 配置文件也无法改变界面语言

### 页面内容对照

| 英文原文 | 中文翻译 | 说明 |
|---------|---------|------|
| Samples Appended | 追加的样本数 | 每秒追加到 TSDB 的样本数量 |
| Scrape Duration | 采集持续时间 | 从目标采集指标所需的时间 |
| Memory Profile | 内存概况 | Prometheus 进程的内存使用情况 |
| WAL Corruptions | WAL 损坏 | 预写日志的损坏次数 |
| Active Appenders | 活跃追加器 | 当前活跃的数据追加器数量 |
| Blocks Loaded | 已加载的块 | TSDB 中已加载的数据块数量 |
| Head Chunks | 头部块 | 内存中的数据块数量 |
| Head Block GC Activity | 头部块 GC 活动 | 垃圾回收活动统计 |
| Compaction Activity | 压缩活动 | 数据压缩操作统计 |
| Reload Count | 重载次数 | 配置重载次数 |
| Query Durations | 查询持续时间 | 查询执行时间统计 |
| Rule Group Eval Duration | 规则组评估持续时间 | 告警规则评估时间 |
| Rule Group Eval Activity | 规则组评估活动 | 规则评估活动统计 |

---

## 解决方案

### 方案 1: 使用 Grafana 中文仪表板（推荐）✅

我已经为你创建了一个**完全中文的 Prometheus 状态监控仪表板**，它提供与 Prometheus `/stats` 页面相同的功能，但界面完全中文化。

#### 文件位置

```
/home/lfl/ceph-exporter/ceph-exporter/deployments/grafana/dashboards/prometheus-stats-zh.json
```

#### 部署步骤

**步骤 1: 重启 Grafana 加载新仪表板**

```bash
cd /home/lfl/ceph-exporter/ceph-exporter/deployments

# 重启 Grafana
docker-compose restart grafana

# 等待 Grafana 启动
sleep 10
```

**步骤 2: 访问中文仪表板**

```
1. 打开浏览器访问: http://localhost:3000
2. 登录 Grafana (admin/admin)
3. 点击左侧菜单 "仪表盘" → "浏览"
4. 找到 "Prometheus 2.0 运行状态（中文版）"
5. 点击打开
```

#### 仪表板内容

新创建的中文仪表板包含以下面板：

**1. 采集统计**
- ✅ 追加的样本数（替代 Samples Appended）
- ✅ 采集持续时间（替代 Scrape Duration）
- ✅ 内存概况（替代 Memory Profile）

**2. TSDB 状态**
- ✅ 活跃追加器（替代 Active Appenders）
- ✅ 已加载的块（替代 Blocks Loaded）
- ✅ 头部块（替代 Head Chunks）

**3. 压缩和重载**
- ✅ 压缩活动（替代 Compaction Activity）
- ✅ 重载次数（替代 Reload Count）

**4. 查询性能**
- ✅ 查询持续时间（替代 Query Durations）
- ✅ 规则组评估持续时间（替代 Rule Group Eval Duration）
- ✅ 规则组评估活动（替代 Rule Group Eval Activity）

**5. WAL 和检查点**
- ✅ WAL 损坏（替代 WAL Corruptions）
- ✅ 头部块 GC 活动（替代 Head Block GC Activity）

---

### 方案 2: 修改 Prometheus 源代码（不推荐）

如果你确实需要修改 Prometheus 自带的 `/stats` 页面，需要：

1. 下载 Prometheus 源代码
2. 修改所有界面文本
3. 重新编译 Prometheus
4. 替换现有的 Prometheus 二进制文件

**缺点**:
- ❌ 工作量大，需要修改大量代码
- ❌ 每次 Prometheus 更新都需要重新修改和编译
- ❌ 维护成本高
- ❌ 可能引入新的 bug

**因此不推荐这种方法。**

---

## 使用指南

### 访问中文仪表板

```bash
# 1. 确保 Grafana 正在运行
docker ps | grep grafana

# 2. 访问 Grafana
# 浏览器打开: http://localhost:3000

# 3. 导航到仪表板
# 左侧菜单 → 仪表盘 → 浏览 → Prometheus 2.0 运行状态（中文版）
```

### 对比原始页面

| 功能 | Prometheus /stats | Grafana 中文仪表板 |
|------|-------------------|-------------------|
| 界面语言 | ❌ 英文 | ✅ 中文 |
| 数据来源 | Prometheus 内部 | Prometheus 指标 |
| 可定制性 | ❌ 不可定制 | ✅ 完全可定制 |
| 告警支持 | ❌ 无 | ✅ 支持 |
| 数据导出 | ❌ 不支持 | ✅ 支持 |
| 时间范围 | 固定 | ✅ 可调整 |
| 刷新间隔 | 固定 | ✅ 可调整 |

### 自定义仪表板

如果需要修改仪表板，可以：

```bash
# 1. 在 Grafana 中打开仪表板
# 2. 点击右上角的 "仪表盘设置"（齿轮图标）
# 3. 选择 "JSON 模型"
# 4. 编辑 JSON 内容
# 5. 点击 "保存更改"

# 或者直接编辑文件
vi /home/lfl/ceph-exporter/ceph-exporter/deployments/grafana/dashboards/prometheus-stats-zh.json

# 重启 Grafana 加载更改
docker-compose restart grafana
```

---

## 常见问题

### Q1: 为什么不能直接修改 Prometheus 的界面？

**A**: Prometheus 是用 Go 语言编写的，界面文本是硬编码在源代码中的。要修改界面，必须：
1. 修改源代码
2. 重新编译
3. 替换二进制文件

这个过程复杂且不易维护。使用 Grafana 创建中文仪表板是更好的选择。

### Q2: Grafana 仪表板的数据和 Prometheus /stats 页面一样吗？

**A**: 是的。两者都是从 Prometheus 的内部指标获取数据，数据来源完全相同。Grafana 仪表板只是提供了更好的可视化和中文界面。

### Q3: 我可以同时使用两个页面吗？

**A**: 可以。Prometheus 的 `/stats` 页面仍然可以访问，你可以根据需要选择使用：
- 英文原始页面: http://localhost:9090/stats
- 中文 Grafana 仪表板: http://localhost:3000（导航到对应仪表板）

### Q4: 如何添加更多指标到仪表板？

**A**:
1. 在 Grafana 中打开仪表板
2. 点击 "添加面板"
3. 选择 "添加新面板"
4. 配置查询（使用 Prometheus 指标）
5. 设置面板标题为中文
6. 保存仪表板

### Q5: 仪表板会自动更新吗？

**A**: 是的。仪表板默认每 30 秒自动刷新一次。你可以在右上角调整刷新间隔。

---

## Prometheus 指标参考

### 常用指标

```promql
# 样本追加速率
rate(prometheus_tsdb_head_samples_appended_total[5m])

# 采集持续时间
scrape_duration_seconds

# 内存使用
process_resident_memory_bytes
go_memstats_heap_alloc_bytes

# TSDB 状态
prometheus_tsdb_head_active_appenders
prometheus_tsdb_blocks_loaded
prometheus_tsdb_head_chunks

# 压缩活动
rate(prometheus_tsdb_compactions_total[5m])
rate(prometheus_tsdb_compactions_failed_total[5m])

# 查询性能
prometheus_engine_query_duration_seconds
prometheus_rule_group_duration_seconds

# WAL 状态
prometheus_tsdb_wal_corruptions_total
```

### 指标说明

| 指标名称 | 中文说明 | 单位 |
|---------|---------|------|
| `prometheus_tsdb_head_samples_appended_total` | 追加到 TSDB 的样本总数 | 计数 |
| `scrape_duration_seconds` | 采集目标所需时间 | 秒 |
| `process_resident_memory_bytes` | 进程常驻内存 | 字节 |
| `prometheus_tsdb_head_active_appenders` | 活跃的追加器数量 | 计数 |
| `prometheus_tsdb_blocks_loaded` | 已加载的数据块数量 | 计数 |
| `prometheus_tsdb_head_chunks` | 内存中的数据块数量 | 计数 |
| `prometheus_tsdb_compactions_total` | 压缩操作总次数 | 计数 |
| `prometheus_engine_query_duration_seconds` | 查询执行时间 | 秒 |
| `prometheus_tsdb_wal_corruptions_total` | WAL 损坏次数 | 计数 |

---

## 快速操作

### 一键部署中文仪表板

```bash
# 进入部署目录
cd /home/lfl/ceph-exporter/ceph-exporter/deployments

# 确认文件已创建
ls -la grafana/dashboards/prometheus-stats-zh.json

# 重启 Grafana
docker-compose restart grafana

# 等待启动
sleep 10

# 访问 Grafana
echo "请访问: http://localhost:3000"
echo "仪表盘名称: Prometheus 2.0 运行状态（中文版）"
```

### 验证部署

```bash
# 检查 Grafana 日志
docker logs grafana --tail 50

# 检查仪表板文件
cat grafana/dashboards/prometheus-stats-zh.json | jq '.title'

# 应该输出: "Prometheus 2.0 运行状态（中文版）"
```

---

## 术语对照表

### 界面元素

| 英文 | 中文 |
|------|------|
| Dashboard | 仪表盘 |
| Panel | 面板 |
| Query | 查询 |
| Legend | 图例 |
| Time Range | 时间范围 |
| Refresh | 刷新 |
| Settings | 设置 |
| Variables | 变量 |
| Annotations | 注释 |
| Alerts | 告警 |

### 统计术语

| 英文 | 中文 |
|------|------|
| Samples | 样本 |
| Appended | 追加的 |
| Scrape | 采集 |
| Duration | 持续时间 |
| Memory | 内存 |
| Profile | 概况 |
| Corruptions | 损坏 |
| Active | 活跃的 |
| Appenders | 追加器 |
| Blocks | 块 |
| Loaded | 已加载 |
| Chunks | 数据块 |
| Compaction | 压缩 |
| Activity | 活动 |
| Reload | 重载 |
| Count | 次数 |
| Query | 查询 |
| Rule Group | 规则组 |
| Evaluation | 评估 |
| WAL | 预写日志 |
| Checkpoint | 检查点 |
| GC | 垃圾回收 |

---

## 相关资源

- Prometheus 官方文档: https://prometheus.io/docs/
- Grafana 官方文档: https://grafana.com/docs/
- Prometheus 指标列表: https://prometheus.io/docs/prometheus/latest/querying/basics/
- Grafana Dashboard 开发: https://grafana.com/docs/grafana/latest/dashboards/

---

**最后更新**: 2026-03-12
**创建者**: Ceph-Exporter 项目团队

---

## 总结

✅ **已完成**:
- 创建了完全中文的 Prometheus 状态监控仪表板
- 所有面板标题、图例、说明都是中文
- 提供与原始 `/stats` 页面相同的功能
- 支持自定义和扩展

✅ **优势**:
- 界面完全中文化
- 可定制性强
- 支持告警
- 数据可导出
- 易于维护

🎯 **推荐使用**: Grafana 中文仪表板替代 Prometheus 原始 `/stats` 页面

如有任何问题或需要添加更多指标，欢迎反馈！
