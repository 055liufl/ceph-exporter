// =============================================================================
// ceph-exporter 程序入口
// =============================================================================
// 本文件是 ceph-exporter 的主入口，负责:
//  1. 解析命令行参数
//  2. 加载配置文件
//  3. 初始化日志系统
//  4. 初始化追踪系统（Phase 1 占位）
//  5. 建立 Ceph 集群连接
//  6. 创建 Prometheus 采集器并注册
//  7. 加载插件（Phase 1 占位）
//  8. 启动 HTTP 服务器
//  9. 监听退出信号，执行优雅关闭
//
// 启动流程:
//
//	配置加载 -> 日志初始化 -> 追踪初始化 -> Ceph 连接 ->
//	采集器注册 -> 插件加载 -> HTTP 服务启动 -> 等待退出信号
//
// 退出流程（优雅关闭）:
//
//	收到 SIGINT/SIGTERM -> 关闭 HTTP 服务器 -> 关闭 Ceph 连接 ->
//	关闭追踪系统 -> 关闭插件管理器 -> 关闭日志系统
//
// 命令行参数:
//
//	-config string  配置文件路径（默认: configs/ceph-exporter.yaml）
//	-version        显示版本信息
//
// =============================================================================
package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ceph-exporter/internal/ceph"
	"ceph-exporter/internal/collector"
	"ceph-exporter/internal/config"
	"ceph-exporter/internal/logger"
	"ceph-exporter/internal/plugin"
	"ceph-exporter/internal/server"
	"ceph-exporter/internal/tracer"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

// 版本信息（通过编译时 -ldflags 注入）
// 示例: go build -ldflags "-X main.version=1.0.0 -X main.buildTime=2024-01-01"
var (
	version   = "dev"     // 版本号
	buildTime = "unknown" // 构建时间
	gitCommit = "unknown" // Git 提交哈希
)

func main() {
	// =========================================================================
	// 第一步: 解析命令行参数
	// =========================================================================
	configPath := flag.String("config", "configs/ceph-exporter.yaml", "配置文件路径")
	showVersion := flag.Bool("version", false, "显示版本信息")
	flag.Parse()

	// 显示版本信息并退出
	if *showVersion {
		fmt.Printf("ceph-exporter %s\n", version)
		fmt.Printf("  构建时间: %s\n", buildTime)
		fmt.Printf("  Git 提交: %s\n", gitCommit)
		os.Exit(0)
	}

	// =========================================================================
	// 第二步: 加载配置文件
	// =========================================================================
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// =========================================================================
	// 第三步: 初始化日志系统
	// =========================================================================
	log, err := logger.NewLogger(&cfg.Logger)
	if err != nil {
		fmt.Fprintf(os.Stderr, "初始化日志失败: %v\n", err)
		os.Exit(1)
	}
	defer log.Close()

	log.WithComponent("main").Infof("ceph-exporter %s 启动中...", version)
	log.WithComponent("main").Infof("配置文件: %s", *configPath)

	// =========================================================================
	// 第四步: 初始化追踪系统（Phase 3 完整实现）
	// =========================================================================
	var tp *tracer.TracerProvider
	if cfg.Tracer.Enabled {
		tp, err = tracer.NewTracerProvider(&cfg.Tracer, log)
		if err != nil {
			log.WithComponent("main").Errorf("初始化追踪系统失败: %v", err)
			os.Exit(1)
		}
		defer func() {
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if shutdownErr := tp.Shutdown(shutdownCtx); shutdownErr != nil {
				log.WithComponent("main").Warnf("Failed to shutdown tracer: %v", shutdownErr)
			}
		}()
	}

	// =========================================================================
	// 第五步: 建立 Ceph 集群连接
	// =========================================================================
	cephClient, err := ceph.NewClient(&cfg.Ceph, log)
	if err != nil {
		log.WithComponent("main").Errorf("创建 Ceph 客户端失败: %v", err)
		os.Exit(1)
	}

	if err := cephClient.Connect(); err != nil {
		log.WithComponent("main").Errorf("连接 Ceph 集群失败: %v", err)
		os.Exit(1)
	}
	defer cephClient.Close()

	// =========================================================================
	// 第六步: 创建 Prometheus 采集器并注册
	// =========================================================================
	// 创建自定义 Registry（不包含 Go 运行时默认指标）
	registry := prometheus.NewRegistry()

	// 注册 Go 运行时指标和进程指标
	registry.MustRegister(collectors.NewGoCollector())
	registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

	// 创建并注册 Ceph 采集器
	registry.MustRegister(collector.NewClusterCollector(cephClient, log))
	registry.MustRegister(collector.NewPoolCollector(cephClient, log))
	registry.MustRegister(collector.NewOSDCollector(cephClient, log))
	registry.MustRegister(collector.NewMonitorCollector(cephClient, log))
	registry.MustRegister(collector.NewHealthCollector(cephClient, log))
	registry.MustRegister(collector.NewMDSCollector(cephClient, log))
	registry.MustRegister(collector.NewRGWCollector(cephClient, log))

	log.WithComponent("main").Info("Prometheus 采集器注册完成（7 个采集器）")

	// =========================================================================
	// 第七步: 加载插件（Phase 1 占位实现）
	// =========================================================================
	pluginMgr := plugin.NewManager(log.Logger, registry)
	defer pluginMgr.Close()

	for _, pluginCfg := range cfg.Plugins {
		if pluginCfg.Enabled {
			if err := pluginMgr.Register(nil, &plugin.PluginInfo{
				Name:        pluginCfg.Name,
				Version:     "1.0.0",
				Description: "Plugin",
				Type:        plugin.PluginTypeCollector,
			}); err != nil {
				log.WithComponent("main").Warnf("加载插件 '%s' 失败: %v", pluginCfg.Name, err)
			}
		}
	}

	// =========================================================================
	// 第八步: 启动 HTTP 服务器
	// =========================================================================
	srv := server.NewServer(&cfg.Server, registry, cephClient, log, tp)

	// 在 goroutine 中启动 HTTP 服务器（非阻塞）
	go func() {
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			log.WithComponent("main").Errorf("HTTP 服务器异常退出: %v", err)
			os.Exit(1)
		}
	}()

	log.WithComponent("main").Infof("ceph-exporter 启动完成，监听 %s:%d",
		cfg.Server.Host, cfg.Server.Port)

	// =========================================================================
	// 第九步: 监听退出信号，执行优雅关闭
	// =========================================================================
	// 创建信号通道，监听 SIGINT（Ctrl+C）和 SIGTERM（kill）
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 阻塞等待退出信号
	sig := <-quit
	log.WithComponent("main").Infof("收到退出信号: %v，开始优雅关闭...", sig)

	// 创建关闭超时上下文（最多等待 30 秒）
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 优雅关闭 HTTP 服务器
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.WithComponent("main").Errorf("HTTP 服务器关闭失败: %v", err)
	}

	log.WithComponent("main").Info("ceph-exporter 已安全退出")
}
