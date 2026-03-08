// =============================================================================
// 日志模块单元测试
// =============================================================================
// 测试覆盖:
//   - 日志实例创建（各种级别和格式）
//   - 输出目标设置（stdout/stderr/file）
//   - 文件输出和日志轮转
//   - 追踪 ID / Span ID / 组件名称字段注入
//   - 关闭和资源释放
//
// =============================================================================
package logger

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ceph-exporter/internal/config"
)

// TestNewLogger_JSONFormat 测试 JSON 格式日志创建
func TestNewLogger_JSONFormat(t *testing.T) {
	cfg := &config.LoggerConfig{
		Level:  "info",
		Format: "json",
		Output: "stdout",
	}

	log, err := NewLogger(cfg)
	if err != nil {
		t.Fatalf("创建 JSON 格式日志失败: %v", err)
	}
	defer log.Close()

	// 验证日志实例不为 nil
	if log == nil {
		t.Fatal("日志实例为 nil")
	}

	// 捕获日志输出并验证 JSON 格式
	var buf bytes.Buffer
	log.SetOutput(&buf)
	log.Info("测试消息")

	// 验证输出是合法的 JSON
	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Errorf("日志输出不是合法的 JSON: %v, 输出内容: %s", err, buf.String())
	}

	// 验证 JSON 字段映射
	if _, ok := logEntry["timestamp"]; !ok {
		t.Error("JSON 日志缺少 'timestamp' 字段")
	}
	if _, ok := logEntry["level"]; !ok {
		t.Error("JSON 日志缺少 'level' 字段")
	}
	if _, ok := logEntry["message"]; !ok {
		t.Error("JSON 日志缺少 'message' 字段")
	}
}

// TestNewLogger_TextFormat 测试 Text 格式日志创建
func TestNewLogger_TextFormat(t *testing.T) {
	cfg := &config.LoggerConfig{
		Level:  "debug",
		Format: "text",
		Output: "stdout",
	}

	log, err := NewLogger(cfg)
	if err != nil {
		t.Fatalf("创建 Text 格式日志失败: %v", err)
	}
	defer log.Close()

	// 捕获日志输出
	var buf bytes.Buffer
	log.SetOutput(&buf)
	log.Debug("调试消息")

	// 验证输出包含关键信息
	output := buf.String()
	if !strings.Contains(output, "调试消息") {
		t.Errorf("Text 日志输出不包含消息内容: %s", output)
	}
}

// TestNewLogger_AllLevels 测试所有日志级别
func TestNewLogger_AllLevels(t *testing.T) {
	levels := []string{"trace", "debug", "info", "warn", "error"}

	for _, level := range levels {
		t.Run("级别_"+level, func(t *testing.T) {
			cfg := &config.LoggerConfig{
				Level:  level,
				Format: "json",
				Output: "stdout",
			}

			log, err := NewLogger(cfg)
			if err != nil {
				t.Fatalf("创建 %s 级别日志失败: %v", level, err)
			}
			defer log.Close()
		})
	}
}

// TestNewLogger_InvalidLevel 测试无效日志级别
func TestNewLogger_InvalidLevel(t *testing.T) {
	cfg := &config.LoggerConfig{
		Level:  "invalid",
		Format: "json",
		Output: "stdout",
	}

	_, err := NewLogger(cfg)
	if err == nil {
		t.Fatal("期望无效日志级别返回错误，但返回了 nil")
	}
}

// TestNewLogger_FileOutput 测试文件输出
func TestNewLogger_FileOutput(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	cfg := &config.LoggerConfig{
		Level:      "info",
		Format:     "json",
		Output:     "file",
		FilePath:   logFile,
		MaxSize:    1,
		MaxBackups: 1,
		MaxAge:     1,
		Compress:   false,
	}

	log, err := NewLogger(cfg)
	if err != nil {
		t.Fatalf("创建文件输出日志失败: %v", err)
	}

	// 写入日志
	log.Info("文件输出测试消息")

	// 关闭日志以确保数据写入磁盘
	if closeErr := log.Close(); closeErr != nil {
		t.Fatalf("关闭日志失败: %v", closeErr)
	}

	// 验证日志文件已创建
	if _, statErr := os.Stat(logFile); os.IsNotExist(statErr) {
		t.Fatal("日志文件未创建")
	}

	// 读取并验证文件内容
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("读取日志文件失败: %v", err)
	}

	if !strings.Contains(string(content), "文件输出测试消息") {
		t.Errorf("日志文件不包含期望的消息内容: %s", string(content))
	}
}

// TestNewLogger_FileOutput_EmptyPath 测试文件输出但路径为空
func TestNewLogger_FileOutput_EmptyPath(t *testing.T) {
	cfg := &config.LoggerConfig{
		Level:  "info",
		Format: "json",
		Output: "file",
		// FilePath 为空
	}

	_, err := NewLogger(cfg)
	if err == nil {
		t.Fatal("期望文件路径为空时返回错误，但返回了 nil")
	}
}

// TestNewLogger_StderrOutput 测试 stderr 输出
func TestNewLogger_StderrOutput(t *testing.T) {
	cfg := &config.LoggerConfig{
		Level:  "info",
		Format: "json",
		Output: "stderr",
	}

	log, err := NewLogger(cfg)
	if err != nil {
		t.Fatalf("创建 stderr 输出日志失败: %v", err)
	}
	defer log.Close()
}

// TestLogger_WithTraceID 测试追踪 ID 字段注入
func TestLogger_WithTraceID(t *testing.T) {
	cfg := &config.LoggerConfig{
		Level:  "info",
		Format: "json",
		Output: "stdout",
	}

	log, err := NewLogger(cfg)
	if err != nil {
		t.Fatalf("创建日志失败: %v", err)
	}
	defer log.Close()

	// 捕获输出
	var buf bytes.Buffer
	log.SetOutput(&buf)

	// 使用 WithTraceID 写入日志
	log.WithTraceID("abc123def456").Info("带追踪ID的消息")

	// 验证输出包含 trace_id 字段
	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("解析日志 JSON 失败: %v", err)
	}

	if traceID, ok := logEntry["trace_id"]; !ok {
		t.Error("日志缺少 trace_id 字段")
	} else if traceID != "abc123def456" {
		t.Errorf("trace_id 期望 'abc123def456'，实际 '%v'", traceID)
	}
}

// TestLogger_WithSpanID 测试 Span ID 字段注入
func TestLogger_WithSpanID(t *testing.T) {
	cfg := &config.LoggerConfig{
		Level:  "info",
		Format: "json",
		Output: "stdout",
	}

	log, err := NewLogger(cfg)
	if err != nil {
		t.Fatalf("创建日志失败: %v", err)
	}
	defer log.Close()

	var buf bytes.Buffer
	log.SetOutput(&buf)

	log.WithSpanID("span789").Info("带SpanID的消息")

	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("解析日志 JSON 失败: %v", err)
	}

	if spanID, ok := logEntry["span_id"]; !ok {
		t.Error("日志缺少 span_id 字段")
	} else if spanID != "span789" {
		t.Errorf("span_id 期望 'span789'，实际 '%v'", spanID)
	}
}

// TestLogger_WithComponent 测试组件名称字段注入
func TestLogger_WithComponent(t *testing.T) {
	cfg := &config.LoggerConfig{
		Level:  "info",
		Format: "json",
		Output: "stdout",
	}

	log, err := NewLogger(cfg)
	if err != nil {
		t.Fatalf("创建日志失败: %v", err)
	}
	defer log.Close()

	var buf bytes.Buffer
	log.SetOutput(&buf)

	log.WithComponent("ceph-client").Info("组件日志消息")

	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("解析日志 JSON 失败: %v", err)
	}

	if component, ok := logEntry["component"]; !ok {
		t.Error("日志缺少 component 字段")
	} else if component != "ceph-client" {
		t.Errorf("component 期望 'ceph-client'，实际 '%v'", component)
	}
}

// TestLogger_Close_NoFileWriter 测试关闭没有文件写入器的日志
func TestLogger_Close_NoFileWriter(t *testing.T) {
	cfg := &config.LoggerConfig{
		Level:  "info",
		Format: "json",
		Output: "stdout",
	}

	log, err := NewLogger(cfg)
	if err != nil {
		t.Fatalf("创建日志失败: %v", err)
	}

	// 关闭没有文件写入器的日志应该不报错
	if err := log.Close(); err != nil {
		t.Errorf("关闭日志失败: %v", err)
	}
}
