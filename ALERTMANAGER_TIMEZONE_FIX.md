# Alertmanager 时区问题解决方案

## 问题描述

Alertmanager 告警事件的时间戳与本地时区（CST，UTC+8）相差 8 小时，显示的是 UTC 时间。

**原因**: Alertmanager 内部使用 UTC 时间存储和显示所有时间戳，这是设计决定，无法通过配置更改。

## 解决方案

### 方案 1: 使用 Grafana 查看告警（推荐）

Grafana 可以自动将 UTC 时间转换为本地时区。

**步骤**:

1. 访问 Grafana: http://localhost:3000
2. 登录（默认: admin/admin）
3. 进入 **Alerting** → **Alert rules**
4. Grafana 会自动显示本地时区的时间

**优点**:
- 自动时区转换
- 更好的可视化
- 统一的告警管理界面

### 方案 2: 浏览器扩展自动转换时区

安装浏览器扩展来自动转换 Alertmanager Web UI 中的时间。

**Chrome/Edge 扩展**:
- **Timezone Converter**: 自动检测并转换页面中的时间戳
- **Time Zone Converter**: 可以配置时区转换规则

**Firefox 扩展**:
- **Timezone Converter**: 类似功能

### 方案 3: 自定义 Alertmanager 模板

通过自定义通知模板，在告警消息中显示本地时间。

**步骤**:

1. 创建模板文件 `alertmanager/templates/default.tmpl`:

```go
{{ define "custom.title" }}
[{{ .Status | toUpper }}{{ if eq .Status "firing" }}:{{ .Alerts.Firing | len }}{{ end }}] {{ .GroupLabels.alertname }}
{{ end }}

{{ define "custom.text" }}
{{ range .Alerts }}
告警名称: {{ .Labels.alertname }}
严重程度: {{ .Labels.severity }}
组件: {{ .Labels.component }}
描述: {{ .Annotations.description }}
触发时间: {{ .StartsAt.Add 28800000000000 | date "2006-01-02 15:04:05" }} (CST)
{{ if .EndsAt }}恢复时间: {{ .EndsAt.Add 28800000000000 | date "2006-01-02 15:04:05" }} (CST){{ end }}
{{ end }}
{{ end }}
```

**说明**: `.Add 28800000000000` 添加 8 小时（28800 秒 = 8 小时，单位是纳秒）

2. 修改 `alertmanager/alertmanager.yml`:

```yaml
global:
  resolve_timeout: 5m

# 添加模板配置
templates:
  - "/etc/alertmanager/templates/*.tmpl"

receivers:
  - name: "default-webhook"
    webhook_configs:
      - url: "http://localhost:5001/webhook/alerts"
        send_resolved: true
```

3. 更新 docker-compose.yml，挂载模板目录:

```yaml
  alertmanager:
    volumes:
      - ./alertmanager/alertmanager.yml:/etc/alertmanager/alertmanager.yml:ro
      - ./alertmanager/templates:/etc/alertmanager/templates:ro  # 添加这行
      - ./data/alertmanager:/alertmanager
      - /etc/localtime:/etc/localtime:ro
      - /etc/timezone:/etc/timezone:ro
```

4. 重启 Alertmanager:

```bash
cd /home/lfl/ceph-exporter/ceph-exporter/deployments
docker-compose restart alertmanager
```

### 方案 4: 使用反向代理修改响应

通过 Nginx 反向代理，使用 JavaScript 在客户端转换时间。

**Nginx 配置示例**:

```nginx
server {
    listen 9094;
    server_name localhost;

    location / {
        proxy_pass http://localhost:9093;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;

        # 注入 JavaScript 转换时间
        sub_filter '</head>' '<script>
            document.addEventListener("DOMContentLoaded", function() {
                // 查找所有时间戳并转换为本地时间
                const timeElements = document.querySelectorAll("time");
                timeElements.forEach(el => {
                    const utcTime = new Date(el.getAttribute("datetime"));
                    el.textContent = utcTime.toLocaleString("zh-CN", {
                        timeZone: "Asia/Shanghai",
                        year: "numeric",
                        month: "2-digit",
                        day: "2-digit",
                        hour: "2-digit",
                        minute: "2-digit",
                        second: "2-digit"
                    });
                });
            });
        </script></head>';
        sub_filter_once off;
    }
}
```

### 方案 5: 使用 Prometheus Alertmanager Webhook 接收器

创建自定义 Webhook 接收器，在接收到告警时转换时间并发送到其他通知渠道。

**Python 示例**:

```python
from flask import Flask, request
from datetime import datetime, timedelta
import json

app = Flask(__name__)

@app.route('/webhook/alerts', methods=['POST'])
def receive_alert():
    data = request.json

    # 转换时间为 CST
    for alert in data.get('alerts', []):
        if 'startsAt' in alert:
            utc_time = datetime.fromisoformat(alert['startsAt'].replace('Z', '+00:00'))
            cst_time = utc_time + timedelta(hours=8)
            alert['startsAt_CST'] = cst_time.strftime('%Y-%m-%d %H:%M:%S CST')

        if 'endsAt' in alert:
            utc_time = datetime.fromisoformat(alert['endsAt'].replace('Z', '+00:00'))
            cst_time = utc_time + timedelta(hours=8)
            alert['endsAt_CST'] = cst_time.strftime('%Y-%m-%d %H:%M:%S CST')

    # 发送到其他通知渠道（企业微信、钉钉等）
    print(json.dumps(data, indent=2, ensure_ascii=False))

    return {'status': 'success'}

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5001)
```

## 推荐方案

**生产环境推荐**: 方案 1（使用 Grafana）+ 方案 3（自定义模板）

**理由**:
1. Grafana 提供更好的告警可视化和管理
2. 自定义模板确保通知消息中的时间是本地时区
3. 不需要修改 Alertmanager 核心配置
4. 易于维护和扩展

## 验证方法

### 1. 触发测试告警

```bash
# 创建测试告警
curl -X POST http://localhost:9093/api/v1/alerts -H "Content-Type: application/json" -d '[
  {
    "labels": {
      "alertname": "TestAlert",
      "severity": "warning",
      "component": "test"
    },
    "annotations": {
      "description": "这是一个测试告警，用于验证时区配置"
    },
    "startsAt": "'$(date -u +%Y-%m-%dT%H:%M:%S.000Z)'",
    "endsAt": "'$(date -u -d '+5 minutes' +%Y-%m-%dT%H:%M:%S.000Z)'"
  }
]'
```

### 2. 检查时间显示

- **Alertmanager UI**: http://localhost:9093
- **Grafana**: http://localhost:3000/alerting/list
- **Webhook 日志**: 检查接收到的告警消息

### 3. 对比时间

```bash
# 当前 UTC 时间
date -u

# 当前本地时间（CST）
date

# 时差应该是 8 小时
```

## 常见问题

### Q1: 为什么 Alertmanager 不支持时区配置？

**A**: Alertmanager 设计为使用 UTC 时间，这是为了：
- 避免时区转换的复杂性和错误
- 确保分布式系统中时间的一致性
- 简化日志和调试

### Q2: 容器已经挂载了 /etc/localtime，为什么还是 UTC？

**A**: 挂载 `/etc/localtime` 只影响容器内部命令（如 `date`）的输出，不影响 Alertmanager 应用程序本身的时间显示逻辑。

### Q3: 如何在告警规则中使用本地时间？

**A**: Prometheus 告警规则中的时间函数（如 `time()`）返回的是 Unix 时间戳，与时区无关。时区转换应该在展示层（Grafana、通知模板）进行。

### Q4: Webhook 接收到的时间是什么格式？

**A**: Webhook 接收到的时间是 RFC3339 格式的 UTC 时间，例如：
```
2024-01-15T10:30:00.000Z
```

转换为 CST 需要加 8 小时：
```
2024-01-15T18:30:00.000+08:00
```

## 相关文档

- [Alertmanager 官方文档](https://prometheus.io/docs/alerting/latest/alertmanager/)
- [Alertmanager 配置说明](https://prometheus.io/docs/alerting/latest/configuration/)
- [Grafana Alerting](https://grafana.com/docs/grafana/latest/alerting/)
- [时区配置说明](ceph-exporter/deployments/TIMEZONE_CONFIGURATION.md)

## 总结

Alertmanager 的时区问题是设计决定，不是 bug。推荐的解决方案是：

1. **日常查看**: 使用 Grafana（自动时区转换）
2. **告警通知**: 使用自定义模板或 Webhook 接收器转换时间
3. **Web UI**: 如果必须使用，可以通过浏览器扩展或反向代理转换

**最佳实践**: 在团队中统一使用 UTC 时间进行沟通和记录，避免时区混淆。
