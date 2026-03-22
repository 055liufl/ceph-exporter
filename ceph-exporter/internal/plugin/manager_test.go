// =============================================================================
// 插件管理器单元测试（Phase 5 完整实现）
// =============================================================================
// 测试插件管理器的完整生命周期管理功能，包括:
//   - 管理器创建和初始化
//   - 插件注册（正常注册和重复注册）
//   - 插件注销（正常注销和不存在的插件）
//   - 插件获取和列表
//   - 批量启动和停止（区分启用/未启用插件）
//   - 健康检查（健康/不健康插件）
//   - 管理器关闭和资源释放
//   - HTTP 插件的创建、初始化和生命周期
//
// 测试工具:
//
//	mockPlugin: 模拟插件实现，用于跟踪方法调用状态
//	（initCalled, startCalled, stopCalled 等标志位）
//
// =============================================================================
package plugin

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// mockPlugin 模拟插件，用于单元测试
// 实现 Plugin 接口，通过标志位跟踪各方法的调用状态
//
// 字段说明:
//   - name: 插件名称
//   - version: 插件版本
//   - description: 插件描述
//   - initCalled: Init 方法是否被调用
//   - startCalled: Start 方法是否被调用
//   - stopCalled: Stop 方法是否被调用
//   - healthErr: Health 方法返回的错误（nil 表示健康）
type mockPlugin struct {
	name        string
	version     string
	description string
	initCalled  bool
	startCalled bool
	stopCalled  bool
	healthErr   error
}

// newMockPlugin 创建模拟插件实例
// 参数:
//   - name: 插件名称
//   - version: 插件版本
//   - description: 插件描述
func newMockPlugin(name, version, description string) *mockPlugin {
	return &mockPlugin{
		name:        name,
		version:     version,
		description: description,
	}
}

// ----- mockPlugin 接口方法实现 -----
// 以下方法实现 Plugin 接口，用于在测试中模拟插件行为

func (p *mockPlugin) Name() string        { return p.name }        // 返回插件名称
func (p *mockPlugin) Version() string     { return p.version }     // 返回插件版本
func (p *mockPlugin) Description() string { return p.description } // 返回插件描述

// Init 模拟初始化，记录调用状态
func (p *mockPlugin) Init(config map[string]interface{}) error {
	p.initCalled = true
	return nil
}

// Start 模拟启动，记录调用状态
func (p *mockPlugin) Start(ctx context.Context) error {
	p.startCalled = true
	return nil
}

// Stop 模拟停止，记录调用状态
func (p *mockPlugin) Stop() error {
	p.stopCalled = true
	return nil
}

// Health 模拟健康检查，返回预设的错误（nil 表示健康）
func (p *mockPlugin) Health() error {
	return p.healthErr
}

// TestNewManager 测试创建插件管理器
func TestNewManager(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	registry := prometheus.NewRegistry()

	manager := NewManager(logger, registry)

	if manager == nil {
		t.Fatal("插件管理器为 nil")
	}
	if manager.plugins == nil {
		t.Error("插件映射未初始化")
	}
	if len(manager.plugins) != 0 {
		t.Errorf("新建管理器应该没有插件，实际有 %d 个", len(manager.plugins))
	}
}

// TestManager_Register 测试插件注册
func TestManager_Register(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	registry := prometheus.NewRegistry()
	manager := NewManager(logger, registry)

	plugin := newMockPlugin("test-plugin", "v1.0.0", "Test plugin")
	info := &PluginInfo{
		Name:        "test-plugin",
		Version:     "v1.0.0",
		Description: "Test plugin",
		Type:        PluginTypeGeneric,
		Enabled:     true,
		Config:      make(map[string]interface{}),
	}

	// 测试注册成功
	err := manager.Register(plugin, info)
	if err != nil {
		t.Fatalf("注册插件失败: %v", err)
	}

	if manager.Count() != 1 {
		t.Errorf("期望插件数量为 1，实际为 %d", manager.Count())
	}

	// 测试重复注册失败
	err = manager.Register(plugin, info)
	if err == nil {
		t.Error("期望重复注册返回错误，但返回了 nil")
	}
}

// TestManager_Unregister 测试插件注销
func TestManager_Unregister(t *testing.T) {
	logger := logrus.New()
	registry := prometheus.NewRegistry()
	manager := NewManager(logger, registry)

	plugin := newMockPlugin("test-plugin", "v1.0.0", "Test plugin")
	info := &PluginInfo{
		Name:    "test-plugin",
		Type:    PluginTypeGeneric,
		Enabled: true,
		Config:  make(map[string]interface{}),
	}

	// 注册插件
	err := manager.Register(plugin, info)
	if err != nil {
		t.Fatalf("注册插件失败: %v", err)
	}

	// 测试注销成功
	err = manager.Unregister("test-plugin")
	if err != nil {
		t.Fatalf("注销插件失败: %v", err)
	}

	if manager.Count() != 0 {
		t.Errorf("期望插件数量为 0，实际为 %d", manager.Count())
	}

	if !plugin.stopCalled {
		t.Error("插件的 Stop 方法未被调用")
	}

	// 测试注销不存在的插件
	err = manager.Unregister("non-existent")
	if err == nil {
		t.Error("期望注销不存在的插件返回错误，但返回了 nil")
	}
}

// TestManager_Get 测试获取插件
func TestManager_Get(t *testing.T) {
	logger := logrus.New()
	registry := prometheus.NewRegistry()
	manager := NewManager(logger, registry)

	plugin := newMockPlugin("test-plugin", "v1.0.0", "Test plugin")
	info := &PluginInfo{
		Name:    "test-plugin",
		Type:    PluginTypeGeneric,
		Enabled: true,
		Config:  make(map[string]interface{}),
	}

	// 注册插件
	err := manager.Register(plugin, info)
	if err != nil {
		t.Fatalf("注册插件失败: %v", err)
	}

	// 测试获取存在的插件
	p, exists := manager.Get("test-plugin")
	if !exists {
		t.Error("插件应该存在")
	}
	if p != plugin {
		t.Error("返回的插件不匹配")
	}

	// 测试获取不存在的插件
	_, exists = manager.Get("non-existent")
	if exists {
		t.Error("不存在的插件不应该返回 true")
	}
}

// TestManager_List 测试列出所有插件
func TestManager_List(t *testing.T) {
	logger := logrus.New()
	registry := prometheus.NewRegistry()
	manager := NewManager(logger, registry)

	// 注册多个插件
	for i := 1; i <= 3; i++ {
		plugin := newMockPlugin(
			fmt.Sprintf("plugin-%d", i),
			"v1.0.0",
			fmt.Sprintf("Plugin %d", i),
		)
		info := &PluginInfo{
			Name:    plugin.Name(),
			Type:    PluginTypeGeneric,
			Enabled: true,
			Config:  make(map[string]interface{}),
		}
		err := manager.Register(plugin, info)
		if err != nil {
			t.Fatalf("注册插件失败: %v", err)
		}
	}

	// 测试列出所有插件
	infos := manager.List()
	if len(infos) != 3 {
		t.Errorf("期望 3 个插件，实际 %d 个", len(infos))
	}
}

// TestManager_StartAll 测试批量启动插件
// 验证:
//   - 已启用的插件会被初始化和启动（initCalled=true, startCalled=true）
//   - 未启用的插件不会被初始化和启动（initCalled=false, startCalled=false）
func TestManager_StartAll(t *testing.T) {
	logger := logrus.New()
	registry := prometheus.NewRegistry()
	manager := NewManager(logger, registry)

	// 注册已启用的插件
	enabledPlugin := newMockPlugin("enabled-plugin", "v1.0.0", "Enabled plugin")
	enabledInfo := &PluginInfo{
		Name:    "enabled-plugin",
		Type:    PluginTypeGeneric,
		Enabled: true,
		Config:  make(map[string]interface{}),
	}
	err := manager.Register(enabledPlugin, enabledInfo)
	if err != nil {
		t.Fatalf("注册插件失败: %v", err)
	}

	// 注册未启用的插件
	disabledPlugin := newMockPlugin("disabled-plugin", "v1.0.0", "Disabled plugin")
	disabledInfo := &PluginInfo{
		Name:    "disabled-plugin",
		Type:    PluginTypeGeneric,
		Enabled: false,
		Config:  make(map[string]interface{}),
	}
	err = manager.Register(disabledPlugin, disabledInfo)
	if err != nil {
		t.Fatalf("注册插件失败: %v", err)
	}

	// 启动所有插件
	err = manager.StartAll()
	if err != nil {
		t.Fatalf("启动插件失败: %v", err)
	}

	// 验证已启用的插件被初始化和启动
	if !enabledPlugin.initCalled {
		t.Error("已启用插件的 Init 方法未被调用")
	}
	if !enabledPlugin.startCalled {
		t.Error("已启用插件的 Start 方法未被调用")
	}

	// 验证未启用的插件未被初始化和启动
	if disabledPlugin.initCalled {
		t.Error("未启用插件的 Init 方法不应该被调用")
	}
	if disabledPlugin.startCalled {
		t.Error("未启用插件的 Start 方法不应该被调用")
	}
}

// TestManager_StopAll 测试停止所有插件
func TestManager_StopAll(t *testing.T) {
	logger := logrus.New()
	registry := prometheus.NewRegistry()
	manager := NewManager(logger, registry)

	// 注册插件
	plugin := newMockPlugin("test-plugin", "v1.0.0", "Test plugin")
	info := &PluginInfo{
		Name:    "test-plugin",
		Type:    PluginTypeGeneric,
		Enabled: true,
		Config:  make(map[string]interface{}),
	}
	err := manager.Register(plugin, info)
	if err != nil {
		t.Fatalf("注册插件失败: %v", err)
	}

	// 启动插件
	err = manager.StartAll()
	if err != nil {
		t.Fatalf("启动插件失败: %v", err)
	}

	// 停止所有插件
	err = manager.StopAll()
	if err != nil {
		t.Fatalf("停止插件失败: %v", err)
	}

	// 验证插件被停止
	if !plugin.stopCalled {
		t.Error("插件的 Stop 方法未被调用")
	}
}

// TestManager_HealthCheck 测试插件健康检查
// 验证:
//   - 健康的插件不会出现在不健康列表中
//   - 不健康的插件会出现在不健康列表中
//   - 返回的 map 中 key 为插件名称，value 为错误信息
func TestManager_HealthCheck(t *testing.T) {
	logger := logrus.New()
	registry := prometheus.NewRegistry()
	manager := NewManager(logger, registry)

	// 注册健康的插件
	healthyPlugin := newMockPlugin("healthy-plugin", "v1.0.0", "Healthy plugin")
	healthyInfo := &PluginInfo{
		Name:    "healthy-plugin",
		Type:    PluginTypeGeneric,
		Enabled: true,
		Config:  make(map[string]interface{}),
	}
	err := manager.Register(healthyPlugin, healthyInfo)
	if err != nil {
		t.Fatalf("注册插件失败: %v", err)
	}

	// 注册不健康的插件
	unhealthyPlugin := newMockPlugin("unhealthy-plugin", "v1.0.0", "Unhealthy plugin")
	unhealthyPlugin.healthErr = fmt.Errorf("plugin is unhealthy")
	unhealthyInfo := &PluginInfo{
		Name:    "unhealthy-plugin",
		Type:    PluginTypeGeneric,
		Enabled: true,
		Config:  make(map[string]interface{}),
	}
	err = manager.Register(unhealthyPlugin, unhealthyInfo)
	if err != nil {
		t.Fatalf("注册插件失败: %v", err)
	}

	// 执行健康检查
	unhealthy := manager.HealthCheck()

	// 验证结果
	if len(unhealthy) != 1 {
		t.Errorf("期望 1 个不健康的插件，实际 %d 个", len(unhealthy))
	}

	if _, exists := unhealthy["unhealthy-plugin"]; !exists {
		t.Error("不健康的插件未在结果中")
	}

	if _, exists := unhealthy["healthy-plugin"]; exists {
		t.Error("健康的插件不应该在结果中")
	}
}

// TestManager_Close 测试关闭管理器
func TestManager_Close(t *testing.T) {
	logger := logrus.New()
	registry := prometheus.NewRegistry()
	manager := NewManager(logger, registry)

	// 注册插件
	plugin := newMockPlugin("test-plugin", "v1.0.0", "Test plugin")
	info := &PluginInfo{
		Name:    "test-plugin",
		Type:    PluginTypeGeneric,
		Enabled: true,
		Config:  make(map[string]interface{}),
	}
	err := manager.Register(plugin, info)
	if err != nil {
		t.Fatalf("注册插件失败: %v", err)
	}

	// 启动插件
	err = manager.StartAll()
	if err != nil {
		t.Fatalf("启动插件失败: %v", err)
	}

	// 关闭管理器
	err = manager.Close()
	if err != nil {
		t.Fatalf("关闭管理器失败: %v", err)
	}

	// 验证插件被停止
	if !plugin.stopCalled {
		t.Error("插件的 Stop 方法未被调用")
	}

	// 验证所有插件被清空
	if manager.Count() != 0 {
		t.Errorf("关闭后应该没有插件，实际有 %d 个", manager.Count())
	}
}

// TestHTTPPlugin 测试 HTTP 远程插件的完整生命周期
// 验证:
//   - 基本属性（Name, Version, Description）
//   - Init 方法正确解析 endpoint、timeout、headers 配置
//   - Start 和 Stop 方法正常工作
func TestHTTPPlugin(t *testing.T) {
	plugin := NewHTTPPlugin("http-storage", "v1.0.0", "HTTP storage plugin")

	// 测试基本属性
	if plugin.Name() != "http-storage" {
		t.Errorf("期望名称为 'http-storage'，实际为 '%s'", plugin.Name())
	}
	if plugin.Version() != "v1.0.0" {
		t.Errorf("期望版本为 'v1.0.0'，实际为 '%s'", plugin.Version())
	}
	if plugin.Description() != "HTTP storage plugin" {
		t.Errorf("期望描述为 'HTTP storage plugin'，实际为 '%s'", plugin.Description())
	}

	// 测试初始化
	config := map[string]interface{}{
		"endpoint": "http://example.com",
		"timeout":  5.0,
		"headers": map[string]interface{}{
			"Authorization": "Bearer token123",
		},
	}

	err := plugin.Init(config)
	if err != nil {
		t.Fatalf("初始化插件失败: %v", err)
	}

	if plugin.endpoint != "http://example.com" {
		t.Errorf("期望 endpoint 为 'http://example.com'，实际为 '%s'", plugin.endpoint)
	}
	if plugin.timeout != 5*time.Second {
		t.Errorf("期望 timeout 为 5s，实际为 %v", plugin.timeout)
	}
	if plugin.headers["Authorization"] != "Bearer token123" {
		t.Errorf("期望 Authorization 为 'Bearer token123'，实际为 '%s'", plugin.headers["Authorization"])
	}

	// 测试启动
	ctx := context.Background()
	err = plugin.Start(ctx)
	if err != nil {
		t.Fatalf("启动插件失败: %v", err)
	}

	// 测试停止
	err = plugin.Stop()
	if err != nil {
		t.Fatalf("停止插件失败: %v", err)
	}
}
