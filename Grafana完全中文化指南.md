# Grafana 完全中文化指南

> **问题**: Grafana 仪表板中图例表格的列标题（Name、Last、Max、Mean）显示为英文
> **目标**: 实现 Grafana 界面的完全中文化

---

## 问题分析

从截图中可以看到：

**已中文化的部分**:
- ✅ 面板标题（集群概览、OSD 状态等）
- ✅ 图例中的数据系列名称（总容量、已用容量、在线、离线等）
- ✅ 左侧导航菜单

**仍为英文的部分**:
- ❌ 图例表格列标题：Name、Last、Max、Mean
- ❌ 这些是 Grafana 系统级别的界面元素

---

## 解决方案

### 方案 1: 升级 Grafana 版本（推荐）

较新版本的 Grafana 对中文的支持更完善。

```bash
# 查看当前 Grafana 版本
docker exec grafana grafana-cli --version

# 更新到最新版本
cd /home/lfl/ceph-exporter/ceph-exporter/deployments

# 编辑 docker-compose 文件，更新 Grafana 镜像版本
vi docker-compose-lightweight-full.yml

# 修改 Grafana 镜像版本为最新
# grafana:
#   image: grafana/grafana:latest  # 或指定版本如 10.2.0

# 重新部署
docker-compose -f docker-compose-lightweight-full.yml pull grafana
docker-compose -f docker-compose-lightweight-full.yml up -d grafana
```

### 方案 2: 清除浏览器缓存

有时浏览器缓存会导致语言设置不生效。

**Chrome/Edge**:
```
1. 按 Ctrl + Shift + Delete
2. 选择"缓存的图片和文件"
3. 时间范围选择"全部时间"
4. 点击"清除数据"
5. 刷新 Grafana 页面（Ctrl + F5 强制刷新）
```

**Firefox**:
```
1. 按 Ctrl + Shift + Delete
2. 选择"缓存"
3. 时间范围选择"全部"
4. 点击"立即清除"
5. 刷新 Grafana 页面（Ctrl + F5 强制刷新）
```

### 方案 3: 检查 Grafana 语言设置

确保 Grafana 的语言设置正确。

```bash
# 检查环境变量
docker exec grafana env | grep GF_DEFAULT_LOCALE

# 应该输出: GF_DEFAULT_LOCALE=zh-CN

# 如果没有，编辑 docker-compose 文件
vi docker-compose-lightweight-full.yml

# 确保 Grafana 服务包含以下环境变量:
# environment:
#   - GF_DEFAULT_LOCALE=zh-CN

# 重启 Grafana
docker-compose -f docker-compose-lightweight-full.yml restart grafana
```

### 方案 4: 在 Grafana 界面中设置语言

```
1. 登录 Grafana (http://localhost:3000)
2. 点击左下角的用户头像
3. 选择 "Preferences"（偏好设置）
4. 在 "UI Language" 下拉菜单中选择 "中文（简体）"
5. 点击 "Save"（保存）
6. 刷新页面
```

### 方案 5: 使用自定义 CSS（高级）

如果以上方法都不行，可以使用自定义 CSS 来替换英文文本。

**创建自定义 CSS 文件**:

```bash
# 创建自定义 CSS 目录
mkdir -p /home/lfl/ceph-exporter/ceph-exporter/deployments/grafana/custom

# 创建 CSS 文件
cat > /home/lfl/ceph-exporter/ceph-exporter/deployments/grafana/custom/chinese.css << 'EOF'
/* Grafana 图例表格列标题中文化 */

/* 替换 "Name" 为 "名称" */
[aria-label="Name"] {
    visibility: hidden;
}
[aria-label="Name"]::after {
    content: "名称";
    visibility: visible;
}

/* 替换 "Last" 为 "最新值" */
[aria-label="Last"] {
    visibility: hidden;
}
[aria-label="Last"]::after {
    content: "最新值";
    visibility: visible;
}

/* 替换 "Max" 为 "最大值" */
[aria-label="Max"] {
    visibility: hidden;
}
[aria-label="Max"]::after {
    content: "最大值";
    visibility: visible;
}

/* 替换 "Mean" 为 "平均值" */
[aria-label="Mean"] {
    visibility: hidden;
}
[aria-label="Mean"]::after {
    content: "平均值";
    visibility: visible;
}

/* 替换 "Min" 为 "最小值" */
[aria-label="Min"] {
    visibility: hidden;
}
[aria-label="Min"]::after {
    content: "最小值";
    visibility: visible;
}

/* 替换 "Total" 为 "总计" */
[aria-label="Total"] {
    visibility: hidden;
}
[aria-label="Total"]::after {
    content: "总计";
    visibility: visible;
}

/* 替换 "Current" 为 "当前值" */
[aria-label="Current"] {
    visibility: hidden;
}
[aria-label="Current"]::after {
    content: "当前值";
    visibility: visible;
}
EOF
```

**配置 Grafana 加载自定义 CSS**:

```bash
# 编辑 docker-compose 文件
vi docker-compose-lightweight-full.yml

# 在 Grafana 服务中添加:
# grafana:
#   volumes:
#     - ./grafana/custom/chinese.css:/usr/share/grafana/public/css/custom.css:ro
#   environment:
#     - GF_DEFAULT_LOCALE=zh-CN
#     - GF_SERVER_CUSTOM_CSS_PATH=/usr/share/grafana/public/css/custom.css

# 重启 Grafana
docker-compose -f docker-compose-lightweight-full.yml restart grafana
```

**注意**: 这个方法可能在 Grafana 更新后失效，因为 HTML 结构可能会改变。

---

## 验证中文化

完成上述步骤后，验证中文化是否成功：

```bash
# 1. 清除浏览器缓存
# 2. 访问 Grafana: http://localhost:3000
# 3. 打开 "Ceph 集群监控" 仪表板
# 4. 检查图例表格的列标题是否已变为中文
```

**预期结果**:
- Name → 名称
- Last → 最新值
- Max → 最大值
- Mean → 平均值

---

## 术语对照表

| 英文 | 中文 | 说明 |
|------|------|------|
| Name | 名称 | 数据系列名称 |
| Last | 最新值 | 最后一个数据点的值 |
| Max | 最大值 | 时间范围内的最大值 |
| Mean | 平均值 | 时间范围内的平均值 |
| Min | 最小值 | 时间范围内的最小值 |
| Total | 总计 | 所有值的总和 |
| Current | 当前值 | 当前时间点的值 |
| First | 首个值 | 第一个数据点的值 |
| Delta | 变化量 | 首尾值的差值 |
| Diff | 差值 | 值的差异 |
| Range | 范围 | 最大值和最小值的差 |
| Count | 计数 | 数据点的数量 |

---

## 常见问题

### Q1: 为什么有些地方是中文，有些是英文？

**A**: Grafana 的中文化分为几个层次：
1. **用户界面**：菜单、按钮等（由 GF_DEFAULT_LOCALE 控制）
2. **Dashboard 内容**：面板标题、图例名称（由 Dashboard JSON 定义）
3. **系统级元素**：图例表格列标题（由 Grafana 版本的翻译完整度决定）

### Q2: 我已经设置了 GF_DEFAULT_LOCALE=zh-CN，为什么还是英文？

**A**: 可能的原因：
1. 浏览器缓存未清除
2. Grafana 版本较旧，中文翻译不完整
3. 需要在用户偏好设置中手动选择中文

### Q3: 升级 Grafana 会丢失数据吗？

**A**: 不会。Grafana 的数据存储在 `/var/lib/grafana` 目录中，只要数据卷挂载正确，升级不会丢失数据。但建议升级前备份：

```bash
# 备份 Grafana 数据
tar -czf grafana-backup-$(date +%Y%m%d).tar.gz \
  /home/lfl/ceph-exporter/ceph-exporter/deployments/data/grafana/
```

### Q4: 自定义 CSS 方法安全吗？

**A**: 自定义 CSS 只是改变显示效果，不会影响数据或功能。但缺点是：
- 可能在 Grafana 更新后失效
- 需要维护 CSS 文件
- 不是官方支持的方法

建议优先使用方案 1（升级 Grafana）或方案 2（清除缓存）。

---

## 推荐方案

根据实际情况选择：

**如果是测试环境**:
1. 先尝试清除浏览器缓存（方案 2）
2. 如果不行，升级 Grafana 到最新版本（方案 1）

**如果是生产环境**:
1. 先在测试环境验证升级 Grafana 的兼容性
2. 备份数据后再升级
3. 如果不能升级，使用自定义 CSS（方案 5）

---

## 快速操作步骤

**最简单的方法（推荐先尝试）**:

```bash
# 1. 清除浏览器缓存
# 在浏览器中按 Ctrl + Shift + Delete，清除缓存

# 2. 强制刷新 Grafana 页面
# 按 Ctrl + F5

# 3. 如果还是英文，重启 Grafana
cd /home/lfl/ceph-exporter/ceph-exporter/deployments
docker-compose -f docker-compose-lightweight-full.yml restart grafana

# 4. 等待 10 秒后再次访问
sleep 10
# 访问 http://localhost:3000
```

---

## 相关资源

- Grafana 官方文档: https://grafana.com/docs/
- Grafana 国际化: https://grafana.com/docs/grafana/latest/setup-grafana/configure-grafana/#default_language
- Grafana 中文社区: https://grafana.com/grafana/dashboards?language=zh

---

**最后更新**: 2026-03-11
**适用版本**: Grafana 8.0+
