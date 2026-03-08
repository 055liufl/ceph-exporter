// =============================================================================
// 插件管理器（Phase 5 完整实现）
// =============================================================================
// 本文件实现完整的插件管理系统，支持：
//   - HTTP 远程插件调用
//   - 插件生命周期管理（加载、启动、停止、卸载）
//   - 插件健康检查
//   - 插件指标注册到 Prometheus
//   - 并发安全的插件管理
//
// =============================================================================
package plugin

import (
	"context"
	"fmt"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// Manager 插件管理器
// 负责插件的加载、启动、停止和生命周期管理
type Manager struct {
	// plugins 存储所有已加载的插件
	plugins map[string]Plugin

	// pluginInfos 存储插件信息
	pluginInfos map[string]*PluginInfo

	// collectors 存储采集器插件的 Prometheus Collector
	collectors map[string]prometheus.Collector

	// mu 保护并发访问
	mu sync.RWMutex

	// logger 日志记录器
	logger *logrus.Entry

	// ctx 上下文
	ctx context.Context

	// cancel 取消函数
	cancel context.CancelFunc

	// registry Prometheus 注册表
	registry *prometheus.Registry
}

// NewManager 创建插件管理器
//
// 参数:
//   - logger: 日志记录器
//   - registry: Prometheus 注册表（用于注册插件采集器）
//
// 返回:
//   - *Manager: 插件管理器实例
func NewManager(logger *logrus.Logger, registry *prometheus.Registry) *Manager {
	ctx, cancel := context.WithCancel(context.Background())

	return &Manager{
		plugins:     make(map[string]Plugin),
		pluginInfos: make(map[string]*PluginInfo),
		collectors:  make(map[string]prometheus.Collector),
		logger:      logger.WithField("component", "plugin-manager"),
		ctx:         ctx,
		cancel:      cancel,
		registry:    registry,
	}
}

// Register 注册插件
//
// 参数:
//   - plugin: 要注册的插件实例
//   - info: 插件信息
//
// 返回:
//   - error: 注册失败时返回错误（如插件名称冲突）
func (m *Manager) Register(plugin Plugin, info *PluginInfo) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	name := plugin.Name()

	// 检查插件是否已注册
	if _, exists := m.plugins[name]; exists {
		return fmt.Errorf("plugin %s already registered", name)
	}

	// 验证插件信息
	if info.Name != name {
		return fmt.Errorf("plugin name mismatch: info.Name=%s, plugin.Name()=%s", info.Name, name)
	}

	// 注册插件
	m.plugins[name] = plugin
	m.pluginInfos[name] = info

	// 如果是采集器插件，注册到 Prometheus
	if collectorPlugin, ok := plugin.(CollectorPlugin); ok && m.registry != nil {
		collector := collectorPlugin.Collector()
		if err := m.registry.Register(collector); err != nil {
			// 如果注册失败，清理已注册的插件
			delete(m.plugins, name)
			delete(m.pluginInfos, name)
			return fmt.Errorf("register collector for plugin %s: %w", name, err)
		}
		m.collectors[name] = collector
	}

	m.logger.WithFields(logrus.Fields{
		"plugin":  name,
		"version": plugin.Version(),
		"type":    info.Type,
	}).Info("插件已注册")

	return nil
}

// Unregister 注销插件
//
// 参数:
//   - name: 插件名称
//
// 返回:
//   - error: 注销失败时返回错误（如插件不存在）
func (m *Manager) Unregister(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	plugin, exists := m.plugins[name]
	if !exists {
		return fmt.Errorf("plugin %s not found", name)
	}

	// 停止插件
	if err := plugin.Stop(); err != nil {
		m.logger.WithError(err).WithField("plugin", name).Warn("停止插件失败")
	}

	// 如果是采集器插件，从 Prometheus 注销
	if collector, ok := m.collectors[name]; ok && m.registry != nil {
		m.registry.Unregister(collector)
		delete(m.collectors, name)
	}

	// 删除插件
	delete(m.plugins, name)
	delete(m.pluginInfos, name)

	m.logger.WithField("plugin", name).Info("插件已注销")

	return nil
}

// Get 获取插件
//
// 参数:
//   - name: 插件名称
//
// 返回:
//   - Plugin: 插件实例
//   - bool: 插件是否存在
func (m *Manager) Get(name string) (Plugin, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	plugin, exists := m.plugins[name]
	return plugin, exists
}

// GetInfo 获取插件信息
//
// 参数:
//   - name: 插件名称
//
// 返回:
//   - *PluginInfo: 插件信息
//   - bool: 插件是否存在
func (m *Manager) GetInfo(name string) (*PluginInfo, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	info, exists := m.pluginInfos[name]
	return info, exists
}

// List 列出所有插件
//
// 返回:
//   - []*PluginInfo: 插件信息列表
func (m *Manager) List() []*PluginInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	infos := make([]*PluginInfo, 0, len(m.pluginInfos))
	for _, info := range m.pluginInfos {
		// 创建副本以避免外部修改
		infoCopy := *info
		infos = append(infos, &infoCopy)
	}

	return infos
}

// StartAll 启动所有已启用的插件
//
// 返回:
//   - error: 有插件启动失败时返回错误
func (m *Manager) StartAll() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	m.logger.Info("开始启动所有插件")

	var errors []error

	for name, plugin := range m.plugins {
		info := m.pluginInfos[name]

		// 跳过未启用的插件
		if !info.Enabled {
			m.logger.WithField("plugin", name).Debug("插件未启用，跳过启动")
			continue
		}

		// 初始化插件
		m.logger.WithField("plugin", name).Debug("初始化插件")
		if err := plugin.Init(info.Config); err != nil {
			m.logger.WithError(err).WithField("plugin", name).Error("初始化插件失败")
			errors = append(errors, fmt.Errorf("init plugin %s: %w", name, err))
			continue
		}

		// 启动插件
		m.logger.WithField("plugin", name).Debug("启动插件")
		if err := plugin.Start(m.ctx); err != nil {
			m.logger.WithError(err).WithField("plugin", name).Error("启动插件失败")
			errors = append(errors, fmt.Errorf("start plugin %s: %w", name, err))
			continue
		}

		m.logger.WithField("plugin", name).Info("插件已启动")
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to start %d plugin(s)", len(errors))
	}

	m.logger.Info("所有插件启动完成")
	return nil
}

// StopAll 停止所有插件
//
// 返回:
//   - error: 有插件停止失败时返回错误
func (m *Manager) StopAll() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	m.logger.Info("开始停止所有插件")

	// 取消上下文
	m.cancel()

	var errors []error

	for name, plugin := range m.plugins {
		m.logger.WithField("plugin", name).Debug("停止插件")
		if err := plugin.Stop(); err != nil {
			m.logger.WithError(err).WithField("plugin", name).Error("停止插件失败")
			errors = append(errors, fmt.Errorf("stop plugin %s: %w", name, err))
			continue
		}

		m.logger.WithField("plugin", name).Info("插件已停止")
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to stop %d plugin(s)", len(errors))
	}

	m.logger.Info("所有插件停止完成")
	return nil
}

// HealthCheck 检查所有插件健康状态
//
// 返回:
//   - map[string]error: 不健康的插件列表（插件名称 -> 错误信息）
func (m *Manager) HealthCheck() map[string]error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	unhealthy := make(map[string]error)

	for name, plugin := range m.plugins {
		info := m.pluginInfos[name]

		// 跳过未启用的插件
		if !info.Enabled {
			continue
		}

		if err := plugin.Health(); err != nil {
			unhealthy[name] = err
			m.logger.WithError(err).WithField("plugin", name).Warn("插件健康检查失败")
		}
	}

	return unhealthy
}

// Count 返回插件总数
//
// 返回:
//   - int: 插件数量
func (m *Manager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.plugins)
}

// EnabledCount 返回已启用的插件数量
//
// 返回:
//   - int: 已启用的插件数量
func (m *Manager) EnabledCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	count := 0
	for _, info := range m.pluginInfos {
		if info.Enabled {
			count++
		}
	}

	return count
}

// Close 关闭插件管理器，释放所有资源
//
// 返回:
//   - error: 关闭过程中的错误
func (m *Manager) Close() error {
	m.logger.Info("关闭插件管理器")

	// 停止所有插件
	if err := m.StopAll(); err != nil {
		m.logger.WithError(err).Warn("停止插件时出现错误")
	}

	// 注销所有采集器
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.registry != nil {
		for name, collector := range m.collectors {
			m.registry.Unregister(collector)
			m.logger.WithField("plugin", name).Debug("采集器已注销")
		}
	}

	// 清空所有映射
	m.plugins = make(map[string]Plugin)
	m.pluginInfos = make(map[string]*PluginInfo)
	m.collectors = make(map[string]prometheus.Collector)

	m.logger.Info("插件管理器已关闭")
	return nil
}
