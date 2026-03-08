// =============================================================================
// 插件管理器单元测试（Phase 5 完整实现）
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

// mockPlugin 模拟插件用于测试
type mockPlugin struct {
	name        string
	version     string
	description string
	initCalled  bool
	startCalled bool
	stopCalled  bool
	healthErr   error
}

func newMockPlugin(name, version, description string) *mockPlugin {
	return &mockPlugin{
		name:        name,
		version:     version,
		description: description,
	}
}

func (p *mockPlugin) Name() string        { return p.name }
func (p *mockPlugin) Version() string     { return p.version }
func (p *mockPlugin) Description() string { return p.description }

func (p *mockPlugin) Init(config map[string]interface{}) error {
	p.initCalled = true
	return nil
}

func (p *mockPlugin) Start(ctx context.Context) error {
	p.startCalled = true
	return nil
}

func (p *mockPlugin) Stop() error {
	p.stopCalled = true
	return nil
}

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

// TestManager_StartAll 测试启动所有插件
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

// TestManager_HealthCheck 测试健康检查
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

// TestHTTPPlugin 测试 HTTP 插件
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
