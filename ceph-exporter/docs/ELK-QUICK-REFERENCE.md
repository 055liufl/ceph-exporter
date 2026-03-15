┌─────────────────────────────────────────────────────────────────────┐
│                  ceph-exporter ELK 日志集成                         │
│                        快速参考卡片                                  │
└─────────────────────────────────────────────────────────────────────┘

╔═══════════════════════════════════════════════════════════════════╗
║  方案选择                                                          ║
╠═══════════════════════════════════════════════════════════════════╣
║                                                                   ║
║  方案1: 直接推送 (Direct Push)                                     ║
║  ├─ 适用: 小规模、实时性要求高                                     ║
║  ├─ 配置: enable_elk: true                                        ║
║  └─ 命令: ./deployments/scripts/switch-logging-mode.sh direct                ║
║                                                                   ║
║  方案2: 容器日志收集 (Container Log) - 推荐                        ║
║  ├─ 适用: 生产环境、Kubernetes                                     ║
║  ├─ 配置: enable_elk: false, output: stdout                       ║
║  └─ 命令: ./deployments/scripts/switch-logging-mode.sh container             ║
║                                                                   ║
╚═══════════════════════════════════════════════════════════════════╝

╔═══════════════════════════════════════════════════════════════════╗
║  快速切换命令                                                      ║
╠═══════════════════════════════════════════════════════════════════╣
║                                                                   ║
║  # 切换到方案1（TCP）                                              ║
║  $ ./deployments/scripts/switch-logging-mode.sh direct                       ║
║                                                                   ║
║  # 切换到方案1（UDP）                                              ║
║  $ ./deployments/scripts/switch-logging-mode.sh direct-udp                   ║
║                                                                   ║
║  # 切换到方案2（推荐）                                             ║
║  $ ./deployments/scripts/switch-logging-mode.sh container                    ║
║                                                                   ║
║  # 开发模式                                                        ║
║  $ ./deployments/scripts/switch-logging-mode.sh dev                          ║
║                                                                   ║
║  # 查看当前配置                                                    ║
║  $ ./deployments/scripts/switch-logging-mode.sh show                         ║
║                                                                   ║
╚═══════════════════════════════════════════════════════════════════╝

╔═══════════════════════════════════════════════════════════════════╗
║  配置文件位置                                                      ║
╠═══════════════════════════════════════════════════════════════════╣
║                                                                   ║
║  主配置:     configs/ceph-exporter.yaml                           ║
║  示例配置:   configs/logger-examples.yaml                         ║
║  Filebeat:   configs/filebeat.yml                                ║
║  Logstash:   configs/logstash.conf                               ║
║  Docker:     configs/docker-compose-elk.yaml                     ║
║                                                                   ║
╚═══════════════════════════════════════════════════════════════════╝

╔═══════════════════════════════════════════════════════════════════╗
║  关键配置参数                                                      ║
╠═══════════════════════════════════════════════════════════════════╣
║                                                                   ║
║  logger:                                                          ║
║    level: "info"                    # 日志级别                    ║
║    format: "json"                   # json 或 text               ║
║    output: "stdout"                 # stdout, stderr, file       ║
║    enable_elk: false                # 启用直接推送                ║
║    logstash_url: "logstash:5044"    # Logstash 地址             ║
║    logstash_protocol: "tcp"         # tcp 或 udp                ║
║    service_name: "ceph-exporter"    # 服务名称                   ║
║                                                                   ║
╚═══════════════════════════════════════════════════════════════════╝

╔═══════════════════════════════════════════════════════════════════╗
║  启动服务                                                          ║
╠═══════════════════════════════════════════════════════════════════╣
║                                                                   ║
║  # 方案1: 直接推送                                                 ║
║  $ docker-compose -f configs/docker-compose-elk.yaml up -d \     ║
║      ceph-exporter-direct logstash elasticsearch kibana          ║
║                                                                   ║
║  # 方案2: 容器日志收集                                             ║
║  $ docker-compose -f configs/docker-compose-elk.yaml up -d \     ║
║      ceph-exporter-sidecar filebeat-sidecar \                    ║
║      logstash elasticsearch kibana                               ║
║                                                                   ║
╚═══════════════════════════════════════════════════════════════════╝

╔═══════════════════════════════════════════════════════════════════╗
║  验证日志                                                          ║
╠═══════════════════════════════════════════════════════════════════╣
║                                                                   ║
║  # 查看 ceph-exporter 日志                                        ║
║  $ docker logs ceph-exporter | grep -i elk                       ║
║                                                                   ║
║  # 查看 Logstash 统计                                             ║
║  $ curl http://localhost:9600/_node/stats/pipelines             ║
║                                                                   ║
║  # 访问 Kibana                                                    ║
║  $ open http://localhost:5601                                    ║
║  创建索引模式: ceph-exporter-*                                     ║
║                                                                   ║
╚═══════════════════════════════════════════════════════════════════╝

╔═══════════════════════════════════════════════════════════════════╗
║  故障排查                                                          ║
╠═══════════════════════════════════════════════════════════════════╣
║                                                                   ║
║  # 检查配置                                                        ║
║  $ grep -A 5 "enable_elk" configs/ceph-exporter.yaml             ║
║                                                                   ║
║  # 测试网络连接                                                    ║
║  $ telnet logstash 5044                                          ║
║                                                                   ║
║  # 查看 Filebeat 状态                                             ║
║  $ docker logs filebeat-sidecar                                  ║
║                                                                   ║
║  # 测试 Filebeat 配置                                             ║
║  $ docker exec filebeat-sidecar filebeat test config            ║
║                                                                   ║
╚═══════════════════════════════════════════════════════════════════╝

╔═══════════════════════════════════════════════════════════════════╗
║  文档                                                              ║
╠═══════════════════════════════════════════════════════════════════╣
║                                                                   ║
║  完整指南:   docs/ELK-LOGGING-GUIDE.md                            ║
║  实现总结:   docs/ELK-IMPLEMENTATION-SUMMARY.md                   ║
║                                                                   ║
╚═══════════════════════════════════════════════════════════════════╝

╔═══════════════════════════════════════════════════════════════════╗
║  推荐配置                                                          ║
╠═══════════════════════════════════════════════════════════════════╣
║                                                                   ║
║  开发环境:   output: stdout, format: text, enable_elk: false     ║
║  生产环境:   output: stdout, format: json, enable_elk: false     ║
║              + Filebeat sidecar (方案2)                           ║
║                                                                   ║
╚═══════════════════════════════════════════════════════════════════╝
