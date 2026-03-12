# ELK 日志集成 - 文件整理清单

## 问题解决过程回顾

### 遇到的问题
1. **Logstash 内存不足** (OutOfMemoryError)
   - 原因: JVM 堆内存只有 128MB
   - 解决: 增加到 512MB

2. **配置端口错误**
   - 原因: logstash_url 配置为 5044 (Beats端口)，应该是 5000 (TCP端口)
   - 解决: 修改为 5000

3. **网络连接问题**
   - 原因: ceph-exporter 和 logstash 不在同一 Docker 网络
   - 解决: 将 ceph-exporter 连接到 logstash 的网络

---

## 核心修复内容（已合并到主文件）

### 1. docker-compose-lightweight-full.yml
**修改内容**:
```yaml
logstash:
  environment:
    - "LS_JAVA_OPTS=-Xms512m -Xmx512m"  # 从 128m 增加到 512m
  volumes:
    - ./logstash/logstash.conf:/usr/share/logstash/pipeline/logstash.conf:ro  # 添加配置挂载
  mem_limit: 768m  # 从 256m 增加到 768m
```

### 2. deployments/logstash/logstash.conf
**修改内容**: 更新为支持 ceph-exporter 的完整配置
- 添加 TCP 输入 (端口 5000)
- 添加 JSON 日志解析
- 添加字段提取和清理

### 3. deployments/configs/ceph-exporter.yaml
**修改内容**:
```yaml
logger:
  enable_elk: true
  logstash_url: "logstash:5000"  # 修正端口从 5044 到 5000
  logstash_protocol: "tcp"
  service_name: "ceph-exporter"
```

---

## 文件分类

### A. 核心功能文件（保留）

#### 代码实现
- `internal/logger/logstash_hook.go` - Logstash Hook 实现
- `internal/logger/logstash_hook_test.go` - 单元测试
- `internal/logger/logger.go` - 已修改，集成 Hook
- `internal/config/config.go` - 已修改，新增配置项

#### 配置文件
- `configs/ceph-exporter.yaml` - 主配置文件（已修改）
- `configs/logger-examples.yaml` - 6种场景配置示例
- `configs/filebeat.yml` - Filebeat 配置
- `configs/logstash.conf` - Logstash 配置示例
- `configs/docker-compose-elk.yaml` - ELK Stack 部署配置
- `deployments/logstash/logstash.conf` - Logstash 实际配置（已修改）
- `deployments/docker-compose-lightweight-full.yml` - 主部署文件（已修改）

#### 文档
- `docs/ELK-LOGGING-GUIDE.md` - 完整使用指南
- `docs/ELK-IMPLEMENTATION-SUMMARY.md` - 实现总结
- `docs/ELK-QUICK-REFERENCE.txt` - 快速参考

#### 脚本
- `deployments/scripts/switch-logging-mode.sh` - 日志模式切换（核心功能）
- `deployments/scripts/README-switch-logging.md` - 脚本使用说明

---

### B. 问题诊断和修复文件（可选保留）

这些文件是在解决问题过程中创建的，用于诊断和修复特定问题。

#### 修复脚本
1. **deployments/scripts/fix-logstash-oom.sh**
   - 用途: 修复 Logstash 内存不足问题
   - 功能: 重启 Logstash 应用新配置
   - 建议: **可以保留**，作为快速修复工具

2. **deployments/scripts/fix-network.sh**
   - 用途: 修复网络连接问题
   - 功能: 将 ceph-exporter 连接到 logstash 网络
   - 建议: **可以保留**，用于类似问题的快速修复

3. **deployments/scripts/test-elk-e2e.sh**
   - 用途: 端到端测试脚本
   - 功能: 测试整个日志链路
   - 建议: **建议保留**，用于验证部署

#### 诊断脚本
4. **deployments/scripts/diagnose-logstash.sh**
   - 用途: 诊断 Logstash 状态
   - 功能: 检查 Logstash 启动状态、端口、日志
   - 建议: **建议保留**，用于日常运维

5. **deployments/scripts/diagnose-elk-full.sh**
   - 用途: 完整的 ELK 链路诊断
   - 功能: 检查所有环节（ceph-exporter、Logstash、Elasticsearch）
   - 建议: **建议保留**，用于故障排查

#### 文档
6. **deployments/LOGSTASH-OOM-FIX.md**
   - 用途: Logstash 内存不足问题修复文档
   - 内容: 问题分析、修复步骤、验证方法
   - 建议: **建议保留**，作为故障排查参考

7. **deployments/LOGSTASH-VERIFICATION-SUCCESS.md**
   - 用途: 修复验证成功报告
   - 内容: 验证步骤、下一步操作
   - 建议: **可以删除**，内容已包含在其他文档中

---

### C. 临时文件（建议删除）

这些是在调试过程中生成的临时文件，不影响系统运行。

#### 配置备份
1. **deployments/configs/ceph-exporter.yaml.bak**
   - 类型: 自动备份文件
   - 用途: 切换日志模式时的备份
   - 建议: **可以删除**（如果确认当前配置正常）

#### 临时测试文件
2. **/tmp/elk-*.txt** (如果存在)
   - 类型: 临时输出文件
   - 用途: 脚本生成的临时报告
   - 建议: **可以删除**

---

## 建议的文件组织结构

```
ceph-exporter/
├── internal/
│   ├── logger/
│   │   ├── logstash_hook.go          [核心] Logstash Hook 实现
│   │   ├── logstash_hook_test.go     [核心] 单元测试
│   │   └── logger.go                 [核心] 日志模块
│   └── config/
│       └── config.go                 [核心] 配置结构
│
├── configs/
│   ├── ceph-exporter.yaml            [核心] 主配置
│   ├── logger-examples.yaml          [参考] 配置示例
│   ├── filebeat.yml                  [参考] Filebeat 配置
│   ├── logstash.conf                 [参考] Logstash 配置示例
│   └── docker-compose-elk.yaml       [参考] ELK 部署示例
│
├── deployments/
│   ├── docker-compose-lightweight-full.yml  [核心] 主部署文件
│   ├── logstash/
│   │   └── logstash.conf             [核心] Logstash 配置
│   ├── scripts/
│   │   ├── switch-logging-mode.sh    [核心] 日志模式切换
│   │   ├── diagnose-elk-full.sh      [运维] 完整诊断
│   │   ├── diagnose-logstash.sh      [运维] Logstash 诊断
│   │   ├── fix-logstash-oom.sh       [运维] 内存修复
│   │   ├── fix-network.sh            [运维] 网络修复
│   │   ├── test-elk-e2e.sh           [测试] 端到端测试
│   │   └── README-switch-logging.md  [文档] 脚本说明
│   ├── LOGSTASH-OOM-FIX.md           [文档] 故障排查
│   └── LOGSTASH-VERIFICATION-SUCCESS.md  [临时] 可删除
│
└── docs/
    ├── ELK-LOGGING-GUIDE.md          [文档] 完整指南
    ├── ELK-IMPLEMENTATION-SUMMARY.md [文档] 实现总结
    └── ELK-QUICK-REFERENCE.txt       [文档] 快速参考
```

---

## 可以删除的文件清单

### 1. 临时备份文件
```bash
# 查找备份文件
find . -name "*.bak" -o -name "*.backup"

# 示例
deployments/configs/ceph-exporter.yaml.bak
```

### 2. 临时文档（内容重复）
```bash
deployments/LOGSTASH-VERIFICATION-SUCCESS.md  # 内容已包含在其他文档中
```

### 3. 临时测试输出
```bash
/tmp/elk-*.txt
/tmp/logstash-*.txt
/tmp/network-*.txt
```

---

## 保留建议总结

### 必须保留（核心功能）
- 所有代码文件 (internal/*)
- 主配置文件 (configs/ceph-exporter.yaml, deployments/docker-compose-lightweight-full.yml)
- Logstash 配置 (deployments/logstash/logstash.conf)
- 核心脚本 (switch-logging-mode.sh)
- 主要文档 (docs/ELK-*.md)

### 建议保留（运维工具）
- 诊断脚本 (diagnose-*.sh)
- 修复脚本 (fix-*.sh)
- 测试脚本 (test-*.sh)
- 故障排查文档 (LOGSTASH-OOM-FIX.md)

### 可以删除（临时文件）
- *.bak 备份文件
- LOGSTASH-VERIFICATION-SUCCESS.md
- /tmp/ 下的临时文件

---

## 删除命令（请谨慎执行）

```bash
cd /home/lfl/ceph-exporter/ceph-exporter

# 删除备份文件（请先确认）
# find deployments/configs -name "*.bak" -delete

# 删除临时文档
# rm -f deployments/LOGSTASH-VERIFICATION-SUCCESS.md

# 清理 /tmp 临时文件
# rm -f /tmp/elk-*.txt /tmp/logstash-*.txt /tmp/network-*.txt
```

**注意**: 以上删除命令已注释，请根据实际情况决定是否执行。

---

## 最终建议

1. **核心文件**: 全部保留，这些是系统正常运行必需的
2. **运维工具**: 建议保留，用于日常运维和故障排查
3. **临时文件**: 可以删除，不影响系统运行
4. **备份文件**: 确认当前配置正常后可以删除

如果磁盘空间充足，建议保留所有文件，因为这些工具和文档在未来可能有用。
